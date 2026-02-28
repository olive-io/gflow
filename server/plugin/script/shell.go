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
	"strings"
	"time"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
)

const (
	ShellPlugin = "sh"
	BashPlugin  = "bash"
)

var _ plugins.Plugin = (*shellPlugin)(nil)

type shellPlugin struct {
	shell    string
	timeout  time.Duration
}

func NewShellPlugin(shell string, timeout time.Duration) plugins.Plugin {
	return &shellPlugin{
		shell:    shell,
		timeout:  timeout,
	}
}

func (sp *shellPlugin) Name() string { return sp.shell }

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

	timeout := sp.timeout
	if options.Timeout > 0 {
		timeout = time.Duration(options.Timeout) * time.Millisecond
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, sp.shell, "-c", request.Script)

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
		ExitCode: 0,
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
	timeout  time.Duration
}

func NewPythonPlugin(timeout time.Duration) plugins.Plugin {
	return &pythonPlugin{
		timeout:  timeout,
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

	timeout := pp.timeout
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
		ExitCode: 0,
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

func parseScriptFormat(script string, format string) (string, []string, error) {
	lines := strings.Split(script, "\n")
	if len(lines) == 0 {
		return "", nil, fmt.Errorf("empty script")
	}

	firstLine := strings.TrimSpace(lines[0])
	if strings.HasPrefix(firstLine, "#!") {
		shebang := strings.TrimPrefix(firstLine, "#!")
		shebang = strings.TrimSpace(shebang)
		parts := strings.Fields(shebang)
		if len(parts) > 0 {
			return parts[0], parts[1:], nil
		}
	}

	switch format {
	case "sh", "bash":
		return "/bin/sh", nil, nil
	case "python", "python3":
		return "python3", nil, nil
	default:
		return "", nil, fmt.Errorf("unsupported script format: %s", format)
	}
}
