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
	"regexp"
	"sort"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/getkin/kin-openapi/openapi3"
	gwrt "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/olive-io/gflow/api/rpc"
	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/server/dao"
)

var (
	httpMethodKey = gwrt.MetadataPrefix + "Http-Method"
	httpPathKey   = gwrt.MetadataPrefix + "Http-Path"
)

type authGRPCServer struct {
	pb.UnimplementedAuthRPCServer

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

func newAuthServer(ctx context.Context, lg *otelzap.Logger, gflowAPI *openapi3.T, enforcer *casbin.Enforcer, roleDao *dao.RoleDao, userDao *dao.UserDao, tokenDao *dao.TokenDao, routeDao *dao.RouteDao) (*authGRPCServer, grpc.UnaryServerInterceptor) {

	interfaces := convertToInterfaces(gflowAPI)
	server := &authGRPCServer{
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

	return server, server.unaryInterceptor
}

func (s *authGRPCServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	resp := &pb.LoginResponse{}
	return resp, nil
}

func (s *authGRPCServer) ListInterfaces(ctx context.Context, in *pb.ListInterfacesRequest) (*pb.ListInterfacesResponse, error) {
	resp := &pb.ListInterfacesResponse{Interfaces: s.interfaces}
	return resp, nil
}

func (s *authGRPCServer) ListPolicies(ctx context.Context, in *pb.ListPoliciesRequest) (*pb.ListPoliciesResponse, error) {
	var values [][]string
	var err error
	if in.Subject == "" {
		values, err = s.enforcer.GetNamedPolicy("p")
	} else {
		values, err = s.enforcer.GetFilteredNamedPolicy("p", 0, in.Subject)
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

func (s *authGRPCServer) AddPolicy(ctx context.Context, in *pb.AddPolicyRequest) (*pb.AddPolicyResponse, error) {
	userInfo, ok := types.GetUserInfo(ctx)
	if !ok || userInfo.IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, "you are not admin")
	}

	rules := make([][]string, 0)

	for _, p := range in.Policies {
		if p.Subject == "*" {
			values := []string{types.DefaultAdministratorRule, p.Object, p.Action}
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

func (s *authGRPCServer) DeletePolicy(ctx context.Context, in *pb.DeletePolicyRequest) (*pb.DeletePolicyResponse, error) {
	rules := make([][]string, 0)
	for _, p := range in.Policies {
		if p.Subject == "*" {
			values := []string{types.DefaultAdministratorRule, p.Object, p.Action}
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

func (s *authGRPCServer) unaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

	fullMethod := info.FullMethod
	switch fullMethod {

	case pb.SystemRPC_Ping_FullMethodName,
		pb.AuthRPC_Login_FullMethodName,
		pb.AdminRPC_ListInterfaces_FullMethodName:
		return handler(ctx, req)
	default:
	}

	if types.IsGflowAgent(ctx) {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "missing Authorization header")
	}

	authorization := md.Get("Authorization")
	if len(authorization) < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "missing Authorization header")
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	if token == "" {
		return nil, status.Errorf(codes.Unauthenticated, "missing token")
	}

	claims, err := types.ParseToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	userInfo := &types.UserInfo{
		Claims: *claims,
	}

	user, err := s.userDao.Get(ctx, claims.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token, user not found")
	}
	userInfo.User = user

	if claims.RoleId != 0 {
		var role *types.Role
		role, err = s.roleDao.Get(ctx, claims.RoleId)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token, role not found")
		}
		userInfo.Role = role
	}
	ctx = types.SetUserInfo(ctx, userInfo)

	httpPath := strings.Join(md.Get(httpPathKey), "")
	httpMethod := strings.ToLower(strings.Join(md.Get(httpMethodKey), ""))
	if httpPath == "" || httpMethod == "" {
		return handler(ctx, req)
	}

	if userInfo.IsAdmin() {
		return handler(ctx, req)
	}

	matched, err := s.enforcer.Enforce(ctx, userInfo.Role.Name, httpPath, httpMethod)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "casbin enforce failed: %v", err)
	}
	if !matched {
		matched, err = s.enforcer.Enforce(ctx, userInfo.User.Username, httpMethod, httpPath)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "casbin enforce failed: %v", err)
		}
	}

	if !matched {
		return nil, status.Errorf(codes.PermissionDenied, "no permission to access this url")
	}

	resp, err := handler(ctx, req)
	return resp, err
}

func convertToInterfaces(document *openapi3.T) []*types.Interface {
	interfaces := make([]*types.Interface, 0)
	for url, path := range document.Paths.Map() {
		var method string
		var operation *openapi3.Operation
		if path.Get != nil {
			method = "get"
			operation = path.Get
		} else if path.Post != nil {
			method = "post"
			operation = path.Post
		} else if path.Put != nil {
			method = "put"
			operation = path.Put
		} else if path.Patch != nil {
			method = "patch"
			operation = path.Patch
		} else if path.Delete != nil {
			method = "delete"
			operation = path.Delete
		}

		if operation == nil {
			continue
		}

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
