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

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"google.golang.org/grpc"

	pb "github.com/olive-io/gflow/api/rpc"
)

var _ pb.BpmnRPCServer = (*bpmnGRPCServer)(nil)

type authGRPCServer struct {
	pb.UnimplementedAuthRPCServer

	ctx context.Context
	lg  *otelzap.Logger
}

func newAuthServer(ctx context.Context, lg *otelzap.Logger) (*authGRPCServer, grpc.UnaryServerInterceptor) {
	server := &authGRPCServer{
		ctx: ctx,
		lg:  lg,
	}

	return server, server.unaryInterceptor
}

func (s *authGRPCServer) unaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(s.ctx, req)
	return resp, err
}
