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
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/olive-io/gflow/api/types"
)

const (
	DefaultTimeout = time.Second * 30
)

type ConfigTLS struct {
	CertFile   string
	KeyFile    string
	CaFile     string
	ServerName string
}

type Config struct {
	Target         string
	DialTimeout    time.Duration
	RequestTimeout time.Duration

	TLS *ConfigTLS

	IsRunner bool

	Token string
}

func NewConfig(target string) *Config {
	opts := &Config{
		Target:         target,
		DialTimeout:    DefaultTimeout,
		RequestTimeout: DefaultTimeout,
	}
	return opts
}

func (cfg *Config) Validate() error {
	if cfg.Target == "" {
		return errors.New("target is required")
	}
	if cfg.DialTimeout < 0 {
		cfg.DialTimeout = DefaultTimeout
	}
	if cfg.RequestTimeout < 0 {
		cfg.RequestTimeout = DefaultTimeout
	}

	return nil
}

func (cfg *Config) buildUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if cfg.Token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", "Bearer "+cfg.Token)
		}
		if cfg.IsRunner {
			ctx = types.AppendGflowAgent(ctx)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
