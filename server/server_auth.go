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

	"github.com/casbin/casbin/v2"
	gwrt "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/olive-io/gflow/api/rpc"
	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/pkg/authutil"
	"github.com/olive-io/gflow/server/dao"
)

var (
	httpMethodKey = gwrt.MetadataPrefix + "Http-Method"
	httpPathKey   = gwrt.MetadataPrefix + "Http-Path"
	userInfoKey   = "User-Info"
)

type UserInfo struct {
	authutil.Claims

	User *types.User
	Role *types.Role
}

func (u *UserInfo) IsAdmin() bool {
	role := u.Role
	if role == nil {
		return false
	}
	return role.Type == types.Role_Root
}

type authGRPCServer struct {
	pb.UnimplementedAuthRPCServer

	ctx context.Context
	lg  *otelzap.Logger

	enforcer *casbin.Enforcer

	roleDao  *dao.RoleDao
	userDao  *dao.UserDao
	tokenDao *dao.TokenDao
	routeDao *dao.RouteDao
}

func newAuthServer(ctx context.Context, lg *otelzap.Logger, enforcer *casbin.Enforcer, roleDao *dao.RoleDao, userDao *dao.UserDao, tokenDao *dao.TokenDao, routeDao *dao.RouteDao) (*authGRPCServer, grpc.UnaryServerInterceptor) {
	server := &authGRPCServer{
		ctx:      ctx,
		lg:       lg,
		enforcer: enforcer,
		roleDao:  roleDao,
		userDao:  userDao,
		tokenDao: tokenDao,
		routeDao: routeDao,
	}

	return server, server.unaryInterceptor
}

func (s *authGRPCServer) unaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

	fullMethod := info.FullMethod
	switch fullMethod {

	case pb.SystemRPC_Ping_FullMethodName,
		pb.AuthRPC_Login_FullMethodName:
		return handler(ctx, req)
	default:
	}

	if authutil.IsGflowAgent(ctx) {
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

	claims, err := authutil.ParseToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	userInfo := &UserInfo{
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
	ctx = context.WithValue(ctx, userInfoKey, userInfo)

	resp, err := handler(ctx, req)
	return resp, err
}
