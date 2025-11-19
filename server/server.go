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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	gwrt "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/yaml"

	pb "github.com/olive-io/gflow/api/rpc"
	"github.com/olive-io/gflow/pkg/casbin"
	"github.com/olive-io/gflow/pkg/dbutil"
	traceutil "github.com/olive-io/gflow/pkg/trace"
	"github.com/olive-io/gflow/plugins"
	"github.com/olive-io/gflow/server/config"
	"github.com/olive-io/gflow/server/dao"
	"github.com/olive-io/gflow/server/dispatch"
	"github.com/olive-io/gflow/server/docs"
	"github.com/olive-io/gflow/server/plugin/receive"
	"github.com/olive-io/gflow/server/plugin/send"
	"github.com/olive-io/gflow/server/plugin/service"
	"github.com/olive-io/gflow/server/scheduler"
	"github.com/olive-io/gflow/third-party/swagger"
)

var (
	DefaultMaxHeaderBytes = 1024 * 1024 * 30
)

type Server struct {
	name string

	cfg *config.Config

	tracerProvider *sdktrace.TracerProvider
}

func NewServer(name string, cfg *config.Config) (*Server, error) {
	err := InitMetrics(cfg.Server.ID)
	if err != nil {
		return nil, fmt.Errorf("initialize metrics: %w", err)
	}

	traceProvider := traceutil.DefaultProvider()
	if cfg.Trace != nil {
		ctx := context.Background()
		traceProvider, err = traceutil.NewJaegerTraceProvider(ctx, cfg.Trace)
		if err != nil {
			return nil, fmt.Errorf("build jaeger trace provider: %w", err)
		}
	}

	server := &Server{
		name: name,
		cfg:  cfg,

		tracerProvider: traceProvider,
	}

	return server, nil
}

func (s *Server) Start(ctx context.Context) error {
	cfg := s.cfg
	lg := s.cfg.Logger()

	address := cfg.Server.Listen
	lg.Info("listening on " + address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen tcp on %s: %w", address, err)
	}

	handler, err := s.buildHandler(ctx)
	if err != nil {
		return fmt.Errorf("build handler: %w", err)
	}

	hs := &http.Server{
		Handler:        handler,
		MaxHeaderBytes: DefaultMaxHeaderBytes,
	}

	ech := make(chan error, 1)
	lg.Sugar().Infof("starting %s", s.name)
	go func() {
		err = hs.Serve(listener)
		if err != nil {
			ech <- err
		}
	}()

	select {
	case err = <-ech:
		return fmt.Errorf("start %s: %w", s.name, err)
	case <-ctx.Done():
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
	}

	lg.Sugar().Infof("shutting down %s", s.name)
	if err = hs.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown %s: %w", s.name, err)
	}
	if err = s.tracerProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown trace provider: %w", err)
	}

	return nil
}

func (s *Server) buildHandler(ctx context.Context) (http.Handler, error) {
	lg := s.cfg.Logger()

	lg.Info("initializing database")
	db, err := dbutil.NewDB(lg, s.cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	lg.Info("initializing casbin")
	enforcer, err := casbin.CreateAdapter(db.DB, s.cfg.Server.AuthorityModel)
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}

	lg.Info("initializing dao layer")
	roleDao, err := dao.NewRoleDao(db)
	if err != nil {
		return nil, fmt.Errorf("create role dao: %w", err)
	}
	userDao, err := dao.NewUserDao(db)
	if err != nil {
		return nil, fmt.Errorf("create user dao: %w", err)
	}
	tokenDao, err := dao.NewTokenDao(db)
	if err != nil {
		return nil, fmt.Errorf("create token dao: %w", err)
	}
	routeDao, err := dao.NewRouteDao(db)
	if err != nil {
		return nil, fmt.Errorf("create route dao: %w", err)
	}
	definitionsDao, err := dao.NewDefinitionsDao(db)
	if err != nil {
		return nil, fmt.Errorf("creates definitions dao: %w", err)
	}
	processDao, err := dao.NewProcessDao(db)
	if err != nil {
		return nil, fmt.Errorf("creates process dao: %w", err)
	}
	runnerDao, err := dao.NewRunnerDao(db)
	if err != nil {
		return nil, fmt.Errorf("creates runner dao: %w", err)
	}
	endpointDao, err := dao.NewEndpointDao(db)
	if err != nil {
		return nil, fmt.Errorf("creates endpoint dao: %w", err)
	}

	dispatcher := dispatch.NewDispatcher()
	dispatcher.Start(ctx)

	// registers plugin factories
	serviceFactory, err := service.NewFactory(lg, dispatcher)
	if err != nil {
		return nil, fmt.Errorf("creates gflow factory: %w", err)
	}
	if err = plugins.Setup(serviceFactory); err != nil {
		return nil, fmt.Errorf("registry plugin factory %s: %w", serviceFactory.Name(), err)
	}
	pluginConfig, err := s.cfg.FormatPluginConfig()
	if err != nil {
		return nil, fmt.Errorf("get plugin config %s: %w", s.name, err)
	}
	if sc := pluginConfig.SendTask; sc != nil {
		sendFactory, err := send.NewFactory(lg, sc)
		if err != nil {
			return nil, fmt.Errorf("creates rabbitmq factory %s: %w", s.name, err)
		}
		if err = plugins.Setup(sendFactory); err != nil {
			return nil, fmt.Errorf("registry plugin factory %s: %w", sendFactory.Name(), err)
		}
	}
	if rc := pluginConfig.ReceiveTask; rc != nil {
		receiveFactory, err := receive.NewFactory(lg, rc)
		if err != nil {
			return nil, fmt.Errorf("creates rabbitmq factory %s: %w", s.name, err)
		}
		if err = plugins.Setup(receiveFactory); err != nil {
			return nil, fmt.Errorf("registry plugin factory %s: %w", receiveFactory.Name(), err)
		}
	}

	openapiYAML, err := docs.GetOpenYAML()
	if err != nil {
		return nil, fmt.Errorf("open openapi yaml: %w", err)
	}
	gflowOpenAPI := &openapi3.T{}
	if err = yaml.Unmarshal(openapiYAML, gflowOpenAPI); err != nil {
		return nil, fmt.Errorf("read openapi yaml: %w", err)
	}

	serverID := s.cfg.Server.ID
	tracer := s.tracerProvider.Tracer(serverID)
	schedulerOptions := scheduler.NewOptions(lg, tracer)
	sch, err := scheduler.NewScheduler(ctx, schedulerOptions)
	if err != nil {
		return nil, fmt.Errorf("creates scheduler: %w", err)
	}

	adminRPC, adminInterceptor, err := newAdminServer(ctx, lg, gflowOpenAPI, enforcer, roleDao, userDao, tokenDao, routeDao)
	if err != nil {
		return nil, fmt.Errorf("create auth rpc server: %w", err)
	}
	authRPC, authInterceptor, err := newAuthServer(ctx, lg, enforcer, roleDao, userDao)
	if err != nil {
		return nil, fmt.Errorf("create auth rpc server: %w", err)
	}
	bpmnRPC, err := newBpmnServer(ctx, lg, sch, definitionsDao, processDao)
	if err != nil {
		return nil, fmt.Errorf("create bpmn rpc server: %w", err)
	}
	systemRPC, err := newSystemGRPCServer(ctx, lg, s.cfg, runnerDao, endpointDao, dispatcher)
	if err != nil {
		return nil, fmt.Errorf("create system rpc server: %w", err)
	}

	endpoints := plugins.ListEndpoints()
	targetID := s.cfg.Server.ID
	for _, endpoint := range endpoints {
		if err = systemRPC.RegisterEndpoint(ctx, endpoint, targetID); err != nil {
			return nil, fmt.Errorf("register endpoint %s: %w", endpoint.Name, err)
		}
	}

	var kaep = keepalive.EnforcementPolicy{
		MinTime:             30 * time.Second, // If a client pings more than once every 1 minute, terminate the connection
		PermitWithoutStream: true,             // Allow pings even when there are no active streams
	}

	var kasp = keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
		MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 15 * time.Second, // Allow 15 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  10 * time.Second, // Ping the client if it is idle for 10 seconds to ensure the connection is still active
		Timeout:               5 * time.Second,  // Wait 5 second for the ping ack before assuming the connection is dead
	}

	sopts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(validateInterceptor, authInterceptor, adminInterceptor),
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
	}

	// registers grpc server
	gs := grpc.NewServer(sopts...)
	pb.RegisterAdminRPCServer(gs, adminRPC)
	pb.RegisterAuthRPCServer(gs, authRPC)
	pb.RegisterBpmnRPCServer(gs, bpmnRPC)
	pb.RegisterSystemRPCServer(gs, systemRPC)
	reflection.Register(gs)

	// creates internal grpc client
	grpcConn, err := s.buildGRPCConn()
	if err != nil {
		return nil, fmt.Errorf("create gRPC connection: %w", err)
	}

	// creates grpc gateway server
	muxOpts := []gwrt.ServeMuxOption{
		gwrt.WithMiddlewares(func(handlerFunc gwrt.HandlerFunc) gwrt.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
				rctx := r.Context()
				rctx = context.WithValue(rctx, httpMethodKey, r.Method)
				r = r.WithContext(rctx)
				handlerFunc(w, r, pathParams)
			}
		}),
	}
	gwmux := gwrt.NewServeMux(muxOpts...)

	// registers grpc handler with grpc client
	if err = pb.RegisterAdminRPCHandlerClient(ctx, gwmux, pb.NewAdminRPCClient(grpcConn)); err != nil {
		return nil, fmt.Errorf("register auth handler: %w", err)
	}
	if err = pb.RegisterAuthRPCHandlerClient(ctx, gwmux, pb.NewAuthRPCClient(grpcConn)); err != nil {
		return nil, fmt.Errorf("register auth handler: %w", err)
	}
	if err = pb.RegisterBpmnRPCHandlerClient(ctx, gwmux, pb.NewBpmnRPCClient(grpcConn)); err != nil {
		return nil, fmt.Errorf("register bpmn handler: %w", err)
	}
	if err = pb.RegisterSystemRPCHandlerClient(ctx, gwmux, pb.NewSystemRPCClient(grpcConn)); err != nil {
		return nil, fmt.Errorf("register system handler: %w", err)
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

	serveMux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		openapiYAML, _ = docs.GetOpenYAML()
		w.WriteHeader(http.StatusOK)
		w.Write(openapiYAML)
	})

	pattern := "/swagger-ui/"
	swaggerFs, err := swagger.GetFS()
	if err != nil {
		return nil, fmt.Errorf("load swagger embed filesystem: %w", err)
	}
	serveMux.PathPrefix(pattern).Handler(http.StripPrefix(pattern, http.FileServer(http.FS(swaggerFs))))
	serveMux.PathPrefix("/").Handler(gwmux)

	return grpcWithHttp(gs, serveMux), nil
}

func (s *Server) buildGRPCConn() (*grpc.ClientConn, error) {
	var tlsConfig *tls.Config
	if gwCfg := s.cfg.Gateway; gwCfg != nil {
		cert, err := tls.LoadX509KeyPair(gwCfg.CertFile, gwCfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load certificate pair: %w", err)
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
			MaxVersion:   tls.VersionTLS13,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			},
		}

		if gwCfg.CaFile != "" {
			caCert, err := os.ReadFile(gwCfg.CaFile)
			if err != nil {
				return nil, fmt.Errorf("load certificate CA: %w", err)
			}
			caPool := x509.NewCertPool()
			caPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caPool
			if gwCfg.ServerName != "" {
				tlsConfig.ServerName = gwCfg.ServerName
			}
		}
	}

	var creds credentials.TransportCredentials
	if tlsConfig != nil {
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}

	defaultTimeout := time.Minute

	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             5 * time.Second,  // wait 5 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithIdleTimeout(defaultTimeout),
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			httpPath, ok := gwrt.HTTPPathPattern(ctx)
			if ok {
				ctx = metadata.AppendToOutgoingContext(ctx, httpPathKey, httpPath)
			}
			httpMethod, ok := ctx.Value(httpMethodKey).(string)
			if ok {
				ctx = metadata.AppendToOutgoingContext(ctx, httpMethodKey, httpMethod)
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	}

	conn, err := grpc.NewClient(s.cfg.Server.Listen, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("initialize grpc client: %w", err)
	}

	return conn, nil
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
	rsp, err := handler(ctx, req)
	return rsp, err
}
