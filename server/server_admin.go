/*
Copyright 2025 The gflow Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/olive-io/gflow/api/rpc"
	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/server/dao"
)

type adminGRPCServer struct {
	pb.UnimplementedAdminRPCServer

	ctx context.Context
	lg  *otelzap.Logger

	gflowAPI   *openapi3.T
	interfaces []*types.Interface
	enforcer   *casbin.Enforcer

	roleDao  *dao.RoleDao
	userDao  *dao.UserDao
	tokenDao *dao.TokenDao
	routeDao *dao.RouteDao
}

func newAdminServer(ctx context.Context, lg *otelzap.Logger, gflowAPI *openapi3.T, enforcer *casbin.Enforcer, roleDao *dao.RoleDao, userDao *dao.UserDao, tokenDao *dao.TokenDao, routeDao *dao.RouteDao) (*adminGRPCServer, grpc.UnaryServerInterceptor, error) {

	interfaces := convertToInterfaces(gflowAPI)
	server := &adminGRPCServer{
		ctx:        ctx,
		lg:         lg,
		gflowAPI:   gflowAPI,
		interfaces: interfaces,
		enforcer:   enforcer,
		roleDao:    roleDao,
		userDao:    userDao,
		tokenDao:   tokenDao,
		routeDao:   routeDao,
	}

	for id, role := range types.SystemRoles {
		_, err := roleDao.Get(ctx, id)
		if err != nil {
			if _, err = roleDao.Create(ctx, role); err != nil {
				return nil, nil, fmt.Errorf("initializing role: %w", err)
			}
		}
	}

	if _, err := userDao.Get(ctx, types.RootUser.Id); err != nil {
		_, err = userDao.Create(ctx, types.RootUser)
		if err != nil {
			return nil, nil, fmt.Errorf("initializing admin user: %w", err)
		}
	}

	for _, p := range types.ListInitPolicies() {
		if _, err := enforcer.AddNamedPolicy("p", p.Rules()); err != nil {
			return nil, nil, fmt.Errorf("initializing policy: %w", err)
		}
	}

	return server, server.unaryInterceptor, nil
}

func (s *adminGRPCServer) ListInterfaces(ctx context.Context, in *pb.ListInterfacesRequest) (*pb.ListInterfacesResponse, error) {
	resp := &pb.ListInterfacesResponse{Interfaces: s.interfaces}
	return resp, nil
}

func (s *adminGRPCServer) ListRoles(ctx context.Context, in *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	page, size := getPageSize(in)

	cis := make([]*dao.CondItem, 0)
	roles, total, err := s.roleDao.PageList(ctx, page, size, cis)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.ListRolesResponse{
		Roles: roles,
		Total: total,
	}
	return resp, nil
}

func (s *adminGRPCServer) CreateRole(ctx context.Context, in *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	value, _ := s.roleDao.GetByName(ctx, in.Name)
	if value != nil {
		return nil, status.Errorf(codes.AlreadyExists, "role '%s' already exists", in.Name)
	}

	role := &types.Role{
		Type:        in.Type,
		Name:        in.Name,
		DisplayName: in.DisplayName,
		Description: in.Description,
		Metadata:    in.Metadata,
	}

	if _, err := s.roleDao.Create(ctx, role); err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.CreateRoleResponse{Role: role}
	return resp, nil
}

func (s *adminGRPCServer) GetRole(ctx context.Context, in *pb.GetRoleRequest) (*pb.GetRoleResponse, error) {
	role, err := s.roleDao.Get(ctx, in.Id)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.GetRoleResponse{Role: role}
	return resp, nil
}

func (s *adminGRPCServer) UpdateRole(ctx context.Context, in *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	role, err := s.roleDao.Get(ctx, in.Id)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	if in.Name != "" && in.Name != role.Name {
		value, _ := s.roleDao.GetByName(ctx, in.Name)
		if value != nil {
			return nil, status.Errorf(codes.AlreadyExists, "role '%s' already exists", in.Name)
		}

		role.Name = in.Name
	}
	if in.DisplayName != "" {
		role.DisplayName = in.DisplayName
	}
	if in.Description != "" {
		role.Description = in.Description
	}
	if len(in.Metadata) != 0 {
		role.Metadata = in.Metadata
	}

	if err = s.roleDao.Update(ctx, role.Id, role); err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.UpdateRoleResponse{Role: role}
	return resp, nil
}

func (s *adminGRPCServer) DeleteRole(ctx context.Context, in *pb.DeleteRoleRequest) (*pb.DeleteRoleResponse, error) {
	if _, found := types.SystemRoles[in.Id]; found {
		return nil, status.Error(codes.FailedPrecondition, "role not allowed to delete")
	}

	role, err := s.roleDao.Get(ctx, in.Id)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	count, err := s.roleDao.RelationalUserCount(ctx, in.Id, nil)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	if count != 0 {
		return nil, status.Error(codes.FailedPrecondition, "other users related to this role")
	}

	if err = s.roleDao.Delete(ctx, in.Id); err != nil {
		return nil, toGRPCErr(err)
	}

	return &pb.DeleteRoleResponse{Role: role}, nil
}

func (s *adminGRPCServer) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	page, size := getPageSize(in)
	cis := make([]*dao.CondItem, 0)
	users, total, err := s.userDao.PageList(ctx, page, size, cis)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.ListUsersResponse{
		Users: users,
		Total: total,
	}
	return resp, nil
}

func (s *adminGRPCServer) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	role, err := s.roleDao.Get(ctx, in.RoleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "role not found")
	}

	value, _ := s.userDao.GetByName(ctx, in.Username)
	if value != nil {
		return nil, status.Errorf(codes.AlreadyExists, "user '%s' already exists", in.Username)
	}

	user := &types.User{
		Username:    in.Username,
		Email:       in.Email,
		Description: in.Description,
		Metadata:    in.Metadata,
		RoleId:      role.Id,
		Login:       &types.UserLoginSpec{},
	}
	user.SetPassword(in.Password)
	if _, err = s.userDao.Create(ctx, user); err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.CreateUserResponse{User: user}
	return resp, nil
}

func (s *adminGRPCServer) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.userDao.Get(ctx, in.Id)
	if err != nil {
		return nil, toGRPCErr(err)
	}
	resp := &pb.GetUserResponse{User: user}
	return resp, nil
}

func (s *adminGRPCServer) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	user, err := s.userDao.Get(ctx, in.Id)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	if in.Email != "" {
		user.Email = in.Email
	}
	if in.Description != "" {
		user.Description = in.Description
	}
	if len(in.Metadata) != 0 {
		user.Metadata = in.Metadata
	}
	if in.Password != "" {
		user.SetPassword(in.Password)
	}
	if in.Init != 0 {
		loginSpec := user.Login
		if loginSpec == nil {
			loginSpec = &types.UserLoginSpec{}
		}
		loginSpec.IsLocked = 0
		loginSpec.LockTimestamp = 0
		loginSpec.FailedRetry = 0
	}

	if err = s.userDao.Update(ctx, user.Id, user); err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.UpdateUserResponse{User: user}
	return resp, nil
}

func (s *adminGRPCServer) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	user, err := s.userDao.Get(ctx, in.Id)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	if user.Username == types.DefaultAdministratorUser {
		return nil, status.Error(codes.FailedPrecondition, "administrator user not allowed to delete")
	}

	if err = s.userDao.Delete(ctx, in.Id); err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.DeleteUserResponse{User: user}
	return resp, nil
}

func (s *adminGRPCServer) ListTokens(ctx context.Context, in *pb.ListTokensRequest) (*pb.ListTokensResponse, error) {
	page, size := getPageSize(in)
	cis := make([]*dao.CondItem, 0)
	tokens, total, err := s.tokenDao.PageList(ctx, page, size, cis)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.ListTokensResponse{
		Tokens: tokens,
		Total:  total,
	}
	return resp, nil
}

func (s *adminGRPCServer) ListRoutes(ctx context.Context, in *pb.ListRoutesRequest) (*pb.ListRoutesResponse, error) {
	page, size := getPageSize(in)
	cis := make([]*dao.CondItem, 0)
	routes, total, err := s.routeDao.PageList(ctx, page, size, cis)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.ListRoutesResponse{
		Routes: routes,
		Total:  total,
	}
	return resp, nil
}

func (s *adminGRPCServer) ListPolicies(ctx context.Context, req *pb.ListPoliciesRequest) (*pb.ListPoliciesResponse, error) {
	var values [][]string
	var err error
	if req.Subject == "" {
		values, err = s.enforcer.GetNamedPolicy("p")
	} else {
		values, err = s.enforcer.GetFilteredNamedPolicy("p", 0, req.Subject)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	policies := make([]*types.Policy, 0)
	for _, value := range values {
		p := types.ParsePolicy(value)
		actions := matchAction(p.Action)
		for _, action := range actions {
			p = &types.Policy{
				Subject: p.Subject,
				Object:  p.Object,
				Action:  action,
			}
			policies = append(policies, p)
		}
	}

	resp := &pb.ListPoliciesResponse{Policies: policies}
	return resp, nil
}

func (s *adminGRPCServer) AddPolicy(ctx context.Context, req *pb.AddPolicyRequest) (*pb.AddPolicyResponse, error) {
	rules := make([][]string, 0)

	for _, p := range req.Policies {
		if p.Subject == "*" {
			values := []string{types.DefaultAdministratorRole, p.Object, p.Action}
			rules = append(rules, values)
			continue
		}
		rules = append(rules, p.Rules())
	}
	for _, rule := range rules {
		_, err := s.enforcer.AddNamedPolicy("p", rule)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.AddPolicyResponse{}, nil
}

func (s *adminGRPCServer) DeletePolicy(ctx context.Context, req *pb.DeletePolicyRequest) (*pb.DeletePolicyResponse, error) {
	rules := make([][]string, 0)
	for _, p := range req.Policies {
		if p.Subject == "*" {
			values := []string{types.DefaultAdministratorRole, p.Object, p.Action}
			rules = append(rules, values)
			continue
		}
		rules = append(rules, p.Rules())
	}
	for _, rule := range rules {
		_, err := s.enforcer.RemoveNamedPolicy("p", rule)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.DeletePolicyResponse{}, nil
}

func (s *adminGRPCServer) unaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "empty authorization")
	}

	httpPath := strings.Join(md.Get(httpPathKey), "")
	if strings.HasPrefix(httpPath, "/v1/admin") {
		userInfo, found := types.GetUserInfo(ctx)
		if !found {
			return nil, status.Errorf(codes.PermissionDenied, "unknown user")
		}

		if !userInfo.IsAdmin() {
			return nil, status.Errorf(codes.PermissionDenied, "only admin can unary")
		}
	}

	return handler(ctx, req)
}

func convertToInterfaces(document *openapi3.T) []*types.Interface {
	interfaces := make([]*types.Interface, 0)
	for url, path := range document.Paths.Map() {
		var method string
		var operation *openapi3.Operation
		paths := map[string]*openapi3.Operation{}
		if path.Get != nil {
			method = "get"
			operation = path.Get
			paths[method] = operation
		}
		if path.Post != nil {
			method = "post"
			operation = path.Post
			paths[method] = operation
		}
		if path.Put != nil {
			method = "put"
			operation = path.Put
			paths[method] = operation
		}
		if path.Patch != nil {
			method = "patch"
			operation = path.Patch
			paths[method] = operation
		}
		if path.Delete != nil {
			method = "delete"
			operation = path.Delete
			paths[method] = operation
		}

		for method, operation = range paths {
			security := map[string]string{}
			if operation.Security != nil {
				for _, item := range *operation.Security {
					for k, v := range item {
						security[k] = strings.Join(v, ";")
					}
				}
			}

			operationID := operation.OperationID
			fullMethod := strings.ReplaceAll(operationID, "_", "/")
			fullMethod = "/rpc." + fullMethod

			item := &types.Interface{
				OperationId: operationID,
				FullMethod:  fullMethod,
				Summary:     operation.Summary,
				Description: operation.Description,
				Method:      method,
				Url:         url,
				Tags:        operation.Tags,
				Security:    security,
				Metadata:    map[string]string{},
				Deprecated:  operation.Deprecated,
			}
			interfaces = append(interfaces, item)
		}
	}

	sort.Slice(interfaces, func(i, j int) bool { return interfaces[i].OperationId < interfaces[j].OperationId })
	return interfaces
}

func matchAction(text string) []string {
	re := regexp.MustCompile(`\(([^(|^)]+)\)`)
	matches := re.FindAllStringSubmatch(text, -1)

	var results []string
	for _, match := range matches {
		results = append(results, match[1])
	}
	if len(results) == 0 {
		results = append(results, text)
	}
	return results
}
