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
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	gwrt "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
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

	interfaces []*types.Interface
	enforcer   *casbin.Enforcer

	roleDao *dao.RoleDao
	userDao *dao.UserDao
}

func newAuthServer(ctx context.Context, lg *otelzap.Logger, enforcer *casbin.Enforcer, roleDao *dao.RoleDao, userDao *dao.UserDao) (*authGRPCServer, grpc.UnaryServerInterceptor, error) {

	server := &authGRPCServer{
		ctx:      ctx,
		lg:       lg,
		enforcer: enforcer,
		roleDao:  roleDao,
		userDao:  userDao,
	}

	return server, server.unaryInterceptor, nil
}

func (s *authGRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.userDao.GetByName(ctx, req.Username)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	if user.IsLocked() {
		return nil, status.Error(codes.FailedPrecondition, "user is locked")
	}

	loginSpec := user.Login
	if loginSpec == nil {
		loginSpec = &types.UserLoginSpec{}
	}

	matched := user.VerifyPassword(req.Password)
	if !matched {
		loginSpec.FailedRetry += 1
		if loginSpec.FailedRetry > types.DefaultMaxFailedLoginRetry {
			loginSpec.FailedRetry = 0
			loginSpec.IsLocked = 1
			loginSpec.LockTimestamp = time.Now().Unix()
		}
		user.Login = loginSpec
		if err = s.userDao.Update(ctx, user.Id, user); err != nil {
			s.lg.Error("failed to update user", zap.Error(err))
		}
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	now := time.Now().Unix()
	if loginSpec.FirstLogon == 0 {
		loginSpec.FirstLogon = now
	}
	loginSpec.LastLogon = now
	loginSpec.FailedRetry = 0
	loginSpec.IsLocked = 0
	loginSpec.LockTimestamp = 0
	user.Login = loginSpec

	if err = s.userDao.Update(ctx, user.Id, user); err != nil {
		return nil, toGRPCErr(err)
	}

	token, err := types.GenerateToken(user.Id, user.RoleId)
	if err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.LoginResponse{
		Token: token,
	}
	return resp, nil
}

func (s *authGRPCServer) GetSelf(ctx context.Context, req *pb.GetSelfRequest) (*pb.GetSelfResponse, error) {
	userInfo, found := types.GetUserInfo(ctx)
	if !found {
		return nil, status.Error(codes.Unauthenticated, "unknown user")
	}

	userPolicies, err := s.enforcer.GetFilteredNamedPolicy("p", 0, userInfo.User.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	rolePolicies, err := s.enforcer.GetFilteredNamedPolicy("p", 0, userInfo.Role.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	policies := make([]*types.Policy, 0)
	for _, value := range append(userPolicies, rolePolicies...) {
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

	resp := &pb.GetSelfResponse{
		User:     userInfo.User,
		Role:     userInfo.Role,
		Policies: policies,
	}

	return resp, nil
}

func (s *authGRPCServer) UpdateSelf(ctx context.Context, req *pb.UpdateSelfRequest) (*pb.UpdateSelfResponse, error) {
	userInfo, found := types.GetUserInfo(ctx)
	if !found {
		return nil, status.Error(codes.Unauthenticated, "unknown user")
	}

	current := userInfo.User
	if req.OldPassword != "" && req.Password != "" {
		if matched := current.VerifyPassword(req.OldPassword); !matched {
			return nil, status.Error(codes.InvalidArgument, "old password not matched")
		}
		current.SetPassword(req.Password)
	}

	if req.Email != "" {
		current.Email = req.Email
	}
	if req.Description != "" {
		current.Description = req.Description
	}
	if len(req.Metadata) != 0 {
		current.Metadata = req.Metadata
	}

	if err := s.userDao.Update(ctx, current.Id, current); err != nil {
		return nil, toGRPCErr(err)
	}

	resp := &pb.UpdateSelfResponse{User: current}
	return resp, nil
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
		return nil, status.Errorf(codes.InvalidArgument, "empty authorization")
	}

	authorization := md.Get("Authorization")
	if len(authorization) < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "empty authorization")
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	if token == "" {
		return nil, status.Errorf(codes.Unauthenticated, "token is required")
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
	httpMethod := strings.ToUpper(strings.Join(md.Get(httpMethodKey), ""))
	if httpPath == "" || httpMethod == "" {
		return handler(ctx, req)
	}

	if userInfo.IsAdmin() {
		return handler(ctx, req)
	}

	matched, err := s.enforcer.Enforce(userInfo.Role.Name, httpPath, httpMethod)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "casbin enforce failed: %v", err)
	}
	if !matched {
		matched, err = s.enforcer.Enforce(userInfo.User.Username, httpMethod, httpPath)
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
