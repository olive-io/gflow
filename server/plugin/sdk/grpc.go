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

package sdk

import (
	"bytes"
	"context"
	"fmt"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GRPCConfig struct {
	Host    string
	Timeout time.Duration
}

type GRPCClient struct {
	cfg    *GRPCConfig
	lg     *otelzap.Logger
	source grpcurl.DescriptorSource
	cc     *grpc.ClientConn
}

func NewGRPCClient(lg *otelzap.Logger, cfg *GRPCConfig) (*GRPCClient, error) {
	network := "tcp"
	var creds credentials.TransportCredentials
	var options []grpc.DialOption

	ctx := context.Background()
	host := cfg.Host
	cc, err := grpcurl.BlockingDial(ctx, network, host, creds, options...)
	if err != nil {
		return nil, err
	}

	refCtx := metadata.NewIncomingContext(ctx, metadata.Pairs())
	refClient := grpcreflect.NewClientAuto(refCtx, cc)
	refClient.AllowMissingFileDescriptors()
	rs := grpcurl.DescriptorSourceFromServer(ctx, refClient)

	client := &GRPCClient{
		cfg:    cfg,
		lg:     lg,
		source: rs,
		cc:     cc,
	}
	return client, nil
}

func (c *GRPCClient) Call(ctx context.Context, name string, headers map[string]string, req, resp any) (*status.Status, error) {
	lg := c.lg
	options := grpcurl.FormatOptions{
		EmitJSONDefaultFields: true,
		IncludeTextSeparator:  true,
		AllowUnknownFields:    true,
	}
	format := grpcurl.FormatJSON

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	in := bytes.NewBuffer(data)
	rf, formatter, err := grpcurl.RequestParserAndFormatter(format, c.source, in, options)
	if err != nil {
		return nil, fmt.Errorf("parse request: %w", err)
	}
	verbosityLevel := 0

	lg.Debug("grpc curl",
		zap.String("format", string(format)),
		zap.String("name", name))

	out := bytes.NewBufferString("")
	h := &grpcurl.DefaultEventHandler{
		Out:            out,
		Formatter:      formatter,
		VerbosityLevel: verbosityLevel,
	}

	var headerSlice []string
	for key, value := range headers {
		headerSlice = append(headerSlice, key+":"+value)
	}

	err = grpcurl.InvokeRPC(ctx, c.source, c.cc, name, headerSlice, h, rf.Next)
	if err != nil {
		return nil, fmt.Errorf("invoke request: %w", err)
	}

	if err = json.Unmarshal(out.Bytes(), resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return h.Status, nil
}
