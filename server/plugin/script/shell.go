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

package script

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"time"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
	"github.com/olive-io/gflow/server/config"
)

const (
	ShellPlugin  = "shell"
	PythonPlugin = "python"
)

var _ plugins.Plugin = (*shellPlugin)(nil)

type shellPlugin struct {
	cfg *config.ScriptTaskPluginConfig
}

func NewShellPlugin(cfg *config.ScriptTaskPluginConfig) plugins.Plugin {
	return &shellPlugin{
		cfg: cfg,
	}
}

func (sp *shellPlugin) Name() string { return ShellPlugin }

func (sp *shellPlugin) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
	var options plugins.DoOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.Stage > types.CallTaskStage_Commit {
		return &plugins.Response{}, nil
	}

	request := new(ShellRequest)
	if err := req.ApplyTo(request); err != nil {
		return nil, fmt.Errorf("binding request: %v", err)
	}

	if request.Script == "" {
		return nil, fmt.Errorf("script is required")
	}

	timeout := time.Duration(options.Timeout) * time.Second
	if options.Timeout > 0 {
		timeout = time.Duration(options.Timeout) * time.Millisecond
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	shell := sp.cfg.Shell.Shell
	cmd := exec.CommandContext(execCtx, shell, "-c", request.Script)

	if request.Dir != "" {
		cmd.Dir = request.Dir
	}

	cmd.Env = os.Environ()
	for k, v := range request.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	response := &ShellResponse{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Success:  err == nil,
		ExitCode: cmd.ProcessState.ExitCode(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			response.ExitCode = exitErr.ExitCode()
			if execCtx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("script execution timeout after %v", timeout)
			}
		} else if execCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("script execution timeout after %v", timeout)
		} else {
			return nil, fmt.Errorf("script execution failed: %v", err)
		}
	}

	return plugins.ExtractResponse(reflect.ValueOf(response)), nil
}

type pythonPlugin struct {
	cfg *config.ScriptTaskPluginConfig
}

func NewPythonPlugin(cfg *config.ScriptTaskPluginConfig) plugins.Plugin {
	return &pythonPlugin{
		cfg: cfg,
	}
}

func (pp *pythonPlugin) Name() string { return "python" }

func (pp *pythonPlugin) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
	var options plugins.DoOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.Stage > types.CallTaskStage_Commit {
		return &plugins.Response{}, nil
	}

	request := new(ShellRequest)
	if err := req.ApplyTo(request); err != nil {
		return nil, fmt.Errorf("binding request: %v", err)
	}

	if request.Script == "" {
		return nil, fmt.Errorf("script is required")
	}

	timeout := time.Duration(pp.cfg.Timeout) * time.Second
	if options.Timeout > 0 {
		timeout = time.Duration(options.Timeout) * time.Millisecond
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "python3", "-c", request.Script)

	if request.Dir != "" {
		cmd.Dir = request.Dir
	}

	cmd.Env = os.Environ()
	for k, v := range request.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	response := &ShellResponse{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Success:  err == nil,
		ExitCode: cmd.ProcessState.ExitCode(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			response.ExitCode = exitErr.ExitCode()
		} else if execCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("script execution timeout after %v", timeout)
		} else {
			return nil, fmt.Errorf("script execution failed: %v", err)
		}
	}

	return plugins.ExtractResponse(reflect.ValueOf(response)), nil
}
