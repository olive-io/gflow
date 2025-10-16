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

package clientgo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	pb "github.com/olive-io/gflow/api/rpc"
	"github.com/olive-io/gflow/api/types"
)

type ListRunnersRequest struct {
	Page, Size int32
}

type ListDefinitionsRequest struct {
	Page, Size int32
}

type ListProcessesRequest struct {
	Page, Size         int32
	DefinitionUID      string
	DefinitionsVersion uint64
	ProcessStatus      types.Process_ProcessStatus
	ProcessStage       types.Process_ProcessStage
}

type ExecuteProcessRequest struct {
	Name               string
	DefinitionsUid     string
	DefinitionsVersion uint64
	Priority           int64
	OnTransaction      bool
	Headers            map[string]string
	Properties         map[string]*types.Value
	DataObjects        map[string]*types.Value
}

type Client struct {
	cfg *Config

	conn *grpc.ClientConn

	bpmnClient   pb.BpmnRPCClient
	systemClient pb.SystemRPCClient

	done chan struct{}
}

func NewClient(cfg *Config) (*Client, error) {
	target := cfg.Target

	var tlsConfig *tls.Config
	if cfgTLS := cfg.TLS; cfgTLS != nil {
		cert, err := tls.LoadX509KeyPair(cfgTLS.CertFile, cfgTLS.KeyFile)
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

		if cfgTLS.CaFile != "" {
			caCert, err := os.ReadFile(cfgTLS.CaFile)
			if err != nil {
				return nil, fmt.Errorf("load certificate CA: %w", err)
			}
			caPool := x509.NewCertPool()
			caPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caPool
		}
	}

	var creds credentials.TransportCredentials
	if tlsConfig != nil {
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}

	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             5 * time.Second,  // wait 5 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}

	DialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithIdleTimeout(DefaultTimeout),
	}

	conn, err := grpc.NewClient(target, DialOpts...)
	if err != nil {
		return nil, fmt.Errorf("initialize grpc client: %w", err)
	}

	bpmnClient := pb.NewBpmnRPCClient(conn)
	systemClient := pb.NewSystemRPCClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	callOpts := make([]grpc.CallOption, 0)
	_, err = systemClient.Ping(ctx, &pb.PingRequest{}, callOpts...)
	if err != nil {
		return nil, parseErr(err)
	}

	client := &Client{
		cfg:          cfg,
		conn:         conn,
		bpmnClient:   bpmnClient,
		systemClient: systemClient,
		done:         make(chan struct{}, 1),
	}

	return client, nil
}

func (c *Client) GetConn() *grpc.ClientConn {
	return c.conn
}

func (c *Client) Ping(ctx context.Context) error {
	opts := c.buildCallOptions()

	in := &pb.PingRequest{}
	_, err := c.systemClient.Ping(ctx, in, opts...)
	if err != nil {
		return parseErr(err)
	}
	return nil
}

func (c *Client) Register(ctx context.Context, runner *types.Runner) (*types.Runner, error) {
	opts := c.buildCallOptions()

	in := &pb.RegisterRequest{
		Runner: runner,
	}
	resp, err := c.systemClient.Register(ctx, in, opts...)
	if err != nil {
		return nil, parseErr(err)
	}
	return resp.Runner, nil
}

func (c *Client) Disregister(ctx context.Context, uid string) (*types.Runner, error) {
	opts := c.buildCallOptions()

	in := &pb.DisregisterRequest{
		Id: uid,
	}
	resp, err := c.systemClient.Disregister(ctx, in, opts...)
	if err != nil {
		return nil, parseErr(err)
	}
	return resp.Runner, nil
}

func (c *Client) ListRunners(ctx context.Context, req *ListRunnersRequest) ([]*types.Runner, int64, error) {
	opts := c.buildCallOptions()

	in := &pb.ListRunnersRequest{
		Page: req.Page,
		Size: req.Size,
	}

	resp, err := c.systemClient.ListRunners(ctx, in, opts...)
	if err != nil {
		return nil, 0, parseErr(err)
	}
	return resp.Runners, resp.Total, nil
}

func (c *Client) GetRunner(ctx context.Context, id int64, uid string) (*types.Runner, error) {
	opts := c.buildCallOptions()

	in := &pb.GetRunnerRequest{
		Id:  id,
		Uid: uid,
	}

	resp, err := c.systemClient.GetRunner(ctx, in, opts...)
	if err != nil {
		return nil, parseErr(err)
	}
	return resp.Runner, nil
}

func (c *Client) AddEndpoints(ctx context.Context, endpoints []*types.Endpoint, target string) error {
	opts := c.buildCallOptions()

	in := &pb.AddEndpointsRequest{
		Endpoints: endpoints,
		Target:    target,
	}
	_, err := c.systemClient.AddEndpoints(ctx, in, opts...)
	if err != nil {
		return parseErr(err)
	}
	return nil
}

func (c *Client) DeployDefinitions(ctx context.Context, definitionsXML []byte, desc string, metadata map[string]string) (*types.Definitions, error) {
	req := &pb.DeployDefinitionsRequest{
		Metadata:    metadata,
		Content:     definitionsXML,
		Description: desc,
	}

	opts := c.buildCallOptions()

	resp, err := c.bpmnClient.DeployDefinition(ctx, req, opts...)
	if err != nil {
		return nil, parseErr(err)
	}
	return resp.Definitions, nil
}

func (c *Client) ListDefinitions(ctx context.Context, in *ListDefinitionsRequest) ([]*types.Definitions, int64, error) {
	req := &pb.ListDefinitionsRequest{
		Page: in.Page,
		Size: in.Size,
	}

	opts := c.buildCallOptions()
	resp, err := c.bpmnClient.ListDefinitions(ctx, req, opts...)
	if err != nil {
		return nil, 0, parseErr(err)
	}
	return resp.DefinitionsList, resp.Total, nil
}

func (c *Client) ListProcesses(ctx context.Context, in *ListProcessesRequest) ([]*types.Process, int64, error) {
	opts := c.buildCallOptions()

	req := &pb.ListProcessRequest{
		Page:               in.Page,
		Size:               in.Size,
		DefinitionsUid:     in.DefinitionUID,
		DefinitionsVersion: in.DefinitionsVersion,
		ProcessStatus:      int32(in.ProcessStatus),
		ProcessStage:       int32(in.ProcessStage),
	}

	resp, err := c.bpmnClient.ListProcess(ctx, req, opts...)
	if err != nil {
		return nil, 0, parseErr(err)
	}
	return resp.Processes, resp.Total, nil
}

func (c *Client) GetDefinitions(ctx context.Context, uid string, version uint64) (*types.Definitions, error) {
	req := &pb.GetDefinitionsRequest{
		Uid:     uid,
		Version: version,
	}

	opts := c.buildCallOptions()
	resp, err := c.bpmnClient.GetDefinitions(ctx, req, opts...)
	if err != nil {
		return nil, parseErr(err)
	}
	return resp.Definitions, nil
}

func (c *Client) ExecuteProcess(ctx context.Context, options *ExecuteProcessRequest) (*types.Process, error) {
	req := &pb.ExecuteProcessRequest{
		Name:               options.Name,
		DefinitionsUid:     options.DefinitionsUid,
		DefinitionsVersion: options.DefinitionsVersion,
		OnTransaction:      options.OnTransaction,
		Priority:           options.Priority,
		Headers:            options.Headers,
		Properties:         options.Properties,
		DataObjects:        options.DataObjects,
	}
	opts := c.buildCallOptions()
	resp, err := c.bpmnClient.ExecuteProcess(ctx, req, opts...)
	if err != nil {
		return nil, parseErr(err)
	}
	return resp.Process, nil
}

func (c *Client) Close() error {
	select {
	case <-c.done:
	default:
		close(c.done)
	}

	return c.conn.Close()
}

func (c *Client) buildCallOptions() []grpc.CallOption {
	opts := make([]grpc.CallOption, 0)
	return opts
}
