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
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/gorilla/mux"
	gwrt "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	pb "github.com/olive-io/gflow/api/rpc"
	"github.com/olive-io/gflow/server/config"
	"github.com/olive-io/gflow/server/dao"
	"github.com/olive-io/gflow/server/dispatch"
	"github.com/olive-io/gflow/server/scheduler"
)

var (
	DefaultMaxHeaderBytes = 1024 * 1024 * 30
)

type Server struct {
	cfg *config.Config
}

func NewServer(cfg *config.Config) (*Server, error) {
	server := &Server{
		cfg: cfg,
	}

	return server, nil
}

func (s *Server) Start(ctx context.Context) error {
	cfg := s.cfg
	lg := s.cfg.Logger()

	listenAddress := cfg.Server.Listen
	lg.Info("listening on " + listenAddress)
	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return fmt.Errorf("listen tcp on %s: %w", listenAddress, err)
	}

	handler, err := s.buildHandler(ctx)
	if err != nil {
		return fmt.Errorf("build handler: %w", err)
	}

	hs := &http.Server{
		Handler:           handler,
		ReadTimeout:       time.Minute,
		ReadHeaderTimeout: time.Second * 30,
		WriteTimeout:      time.Minute,
		IdleTimeout:       time.Second * 30,
		MaxHeaderBytes:    DefaultMaxHeaderBytes,
	}

	ech := make(chan error, 1)
	go func() {
		err = hs.Serve(ln)
		if err != nil {
			ech <- err
		}
	}()

	select {
	case err = <-ech:
		return fmt.Errorf("start server: %w", err)
	case <-ctx.Done():
	}

	if err = hs.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	return nil
}

func (s *Server) buildHandler(ctx context.Context) (http.Handler, error) {
	lg := s.cfg.Logger()

	db, err := openLocalDB(lg, &s.cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	definitionsDao, err := dao.NewDefinitionsDao(db)
	if err != nil {
		return nil, fmt.Errorf("create definitions model: %w", err)
	}
	processDao, err := dao.NewProcessDao(db)
	if err != nil {
		return nil, fmt.Errorf("create process model: %w", err)
	}
	runnerDao, err := dao.NewRunnerDao(db)
	if err != nil {
		return nil, fmt.Errorf("create runner model: %w", err)
	}

	schedulerOptions := scheduler.NewOptions(lg)
	sch, err := scheduler.NewScheduler(ctx, schedulerOptions)
	if err != nil {
		return nil, fmt.Errorf("create scheduler: %w", err)
	}

	dispatcher := dispatch.NewDispatcher()
	dispatcher.Start(ctx)

	bpmnHandler := newBpmnServer(ctx, lg, sch, definitionsDao, processDao)
	systemHandler := newSystemGRPCServer(ctx, lg, runnerDao, dispatcher)

	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}
	kasp := keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second,
		MaxConnectionAge:      30 * time.Second,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  5 * time.Second,
		Timeout:               3 * time.Second,
	}

	sopts := []grpc.ServerOption{
		grpc.UnaryInterceptor(validateInterceptor),
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
	}
	gs := grpc.NewServer(sopts...)

	muxOpts := []gwrt.ServeMuxOption{}
	gwmux := gwrt.NewServeMux(muxOpts...)

	pb.RegisterBpmnRPCServer(gs, bpmnHandler)
	if err = pb.RegisterBpmnRPCHandlerServer(ctx, gwmux, bpmnHandler); err != nil {
		return nil, fmt.Errorf("setup olive bpmn handler: %w", err)
	}

	pb.RegisterSystemRPCServer(gs, systemHandler)
	if err = pb.RegisterSystemRPCHandlerServer(ctx, gwmux, systemHandler); err != nil {
		return nil, fmt.Errorf("setup olive system handler: %w", err)
	}

	serveMux := mux.NewRouter()
	serveMux.Handle("/metrics", promhttp.Handler())

	pprofMux := http.NewServeMux()
	pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	serveMux.PathPrefix("/debug/pprof/").Handler(pprofMux)

	serveMux.Handle("/v1/",
		wsproxy.WebsocketProxy(
			gwmux,
			wsproxy.WithRequestMutator(
				// Default to the POST method for streams
				func(_ *http.Request, outgoing *http.Request) *http.Request {
					outgoing.Method = "POST"
					return outgoing
				},
			),
			wsproxy.WithMaxRespBodyBufferSize(0x7fffffff),
		),
	)

	return grpcWithHttp(gs, serveMux), nil
}

func grpcWithHttp(gh *grpc.Server, hh http.Handler) http.Handler {
	h2s := &http2.Server{}
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			gh.ServeHTTP(w, r)
		} else {
			hh.ServeHTTP(w, r)
		}
	}), h2s)
}

func validateInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if impl, ok := req.(interface{ ValidateAll() error }); ok {
		if err := impl.ValidateAll(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return handler(ctx, req)
}
