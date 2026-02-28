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
	"context"
	"testing"
	"time"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
)

func TestShellPlugin_Name(t *testing.T) {
	plugin := NewShellPlugin("sh", 30*time.Second)
	if plugin.Name() != "sh" {
		t.Errorf("expected name 'sh', got '%s'", plugin.Name())
	}
}

func TestShellPlugin_Do_Success(t *testing.T) {
	plugin := NewShellPlugin("sh", 30*time.Second)

	req := &plugins.Request{
		Properties: map[string]*types.Value{
			"script": types.NewValue("echo 'hello world'"),
		},
	}

	resp, err := plugin.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	stdout := resp.Results["stdout"]
	if stdout == nil {
		t.Fatal("expected stdout in results")
	}
	if stdout.Value != "hello world\n" && stdout.Value != "hello world\r\n" {
		t.Errorf("expected stdout to contain 'hello world', got '%s'", stdout.Value)
	}

	exitCode := resp.Results["exitCode"]
	if exitCode == nil || exitCode.Value != "0" {
		t.Errorf("expected exitCode 0, got '%s'", exitCode.Value)
	}

	success := resp.Results["success"]
	if success == nil || success.Value != "true" {
		t.Errorf("expected success true, got '%s'", success.Value)
	}
}

func TestShellPlugin_Do_WithEnv(t *testing.T) {
	plugin := NewShellPlugin("sh", 30*time.Second)

	req := &plugins.Request{
		Properties: map[string]*types.Value{
			"script": types.NewValue("echo $MY_VAR"),
			"env":    types.NewValue(map[string]any{"MY_VAR": "test_value"}),
		},
	}

	resp, err := plugin.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	stdout := resp.Results["stdout"]
	if stdout == nil {
		t.Fatal("expected stdout in results")
	}
	if stdout.Value != "test_value\n" && stdout.Value != "test_value\r\n" {
		t.Errorf("expected stdout to contain 'test_value', got '%s'", stdout.Value)
	}
}

func TestShellPlugin_Do_WithDir(t *testing.T) {
	plugin := NewShellPlugin("sh", 30*time.Second)

	req := &plugins.Request{
		Properties: map[string]*types.Value{
			"script": types.NewValue("pwd"),
			"dir":    types.NewValue("/tmp"),
		},
	}

	resp, err := plugin.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	stdout := resp.Results["stdout"]
	if stdout == nil {
		t.Fatal("expected stdout in results")
	}
	if stdout.Value != "/tmp\n" && stdout.Value != "/private/tmp\n" && stdout.Value != "/private/tmp\r\n" {
		t.Errorf("expected stdout to contain '/tmp', got '%s'", stdout.Value)
	}
}

func TestShellPlugin_Do_Error(t *testing.T) {
	plugin := NewShellPlugin("sh", 30*time.Second)

	req := &plugins.Request{
		Properties: map[string]*types.Value{
			"script": types.NewValue("exit 1"),
		},
	}

	resp, err := plugin.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	exitCode := resp.Results["exitCode"]
	if exitCode == nil || exitCode.Value != "1" {
		t.Errorf("expected exitCode 1, got '%s'", exitCode.Value)
	}

	success := resp.Results["success"]
	if success == nil || success.Value != "false" {
		t.Errorf("expected success false, got '%s'", success.Value)
	}
}

func TestShellPlugin_Do_Timeout(t *testing.T) {
	plugin := NewShellPlugin("sh", 100*time.Millisecond)

	req := &plugins.Request{
		Properties: map[string]*types.Value{
			"script": types.NewValue("sleep 5"),
		},
	}

	_, err := plugin.Do(context.Background(), req)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if err.Error() != "script execution timeout after 100ms" && err.Error() != "signal: killed" {
		t.Logf("got error: %v", err)
	}
}

func TestShellPlugin_Do_Disabled(t *testing.T) {
	plugin := NewShellPlugin("sh", 30*time.Second)

	req := &plugins.Request{
		Properties: map[string]*types.Value{
			"script": types.NewValue("echo 'hello'"),
		},
	}

	_, err := plugin.Do(context.Background(), req)
	if err == nil {
		t.Fatal("expected disabled error")
	}
}

func TestShellPlugin_Do_EmptyScript(t *testing.T) {
	plugin := NewShellPlugin("sh", 30*time.Second)

	req := &plugins.Request{
		Properties: map[string]*types.Value{},
	}

	_, err := plugin.Do(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty script")
	}
}

func TestShellPlugin_Do_StageSkip(t *testing.T) {
	plugin := NewShellPlugin("sh", 30*time.Second)

	req := &plugins.Request{
		Properties: map[string]*types.Value{
			"script": types.NewValue("echo 'hello'"),
		},
	}

	resp, err := plugin.Do(context.Background(), req, plugins.DoWithTaskStage(types.CallTaskStage_Rollback))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected empty results for rollback stage, got %d", len(resp.Results))
	}
}

func TestPythonPlugin_Name(t *testing.T) {
	plugin := NewPythonPlugin(30*time.Second)
	if plugin.Name() != "python" {
		t.Errorf("expected name 'python', got '%s'", plugin.Name())
	}
}
