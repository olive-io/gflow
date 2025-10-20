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

package runner

import (
	"context"
	"testing"
)

type Data1 struct {
	Name string
}

type TestRequest struct {
	ContentType string   `gflow:"hr:content_type"`
	D1          *Data1   `gflow:"dt:d1"`
	P1          int32    `json:"p1"`
	P2          string   `json:"p2"`
	S1          []string `json:"s1"`
}

type TestResponse struct {
	R1 string `json:"r1"`
	R2 int32  `json:"r2"`
}

var _ Task = (*SimpleTask)(nil)

type SimpleTask struct{}

func (s SimpleTask) Commit(ctx context.Context, request any) (any, error) {
	return request, nil
}

func (s SimpleTask) Rollback(ctx context.Context) error {
	return nil
}

func (s SimpleTask) Destroy(ctx context.Context) error {
	return nil
}

func (s SimpleTask) String() string {
	return "SimpleTask"
}

func TestExtractTask(t *testing.T) {
	st := &SimpleTask{}

	options := NewOptions(WithName("test"), WithRequest(new(TestRequest)), WithResponse(new(TestResponse)))
	endpoint, _, err := extractTask(st, options)
	if err != nil {
		t.Fatalf("failed to extract task: %v", err)
	}

	t.Logf("%+v", endpoint)
}

func Fn1(ctx context.Context, name string) (int, error) { return 0, nil }

func Fn2(name string) error { return nil }

func Fn3() error { return nil }

func Fn4(request *TestRequest) (*TestRequest, error) { return nil, nil }

func TestExtractFunc(t *testing.T) {
	endpoint, _, err := extractFunc(Fn1, NewOptions())
	if err != nil {
		t.Fatalf("failed to extract function: %v", err)
	}

	t.Logf("%+v", endpoint)

	endpoint, _, err = extractFunc(Fn2, NewOptions())
	if err != nil {
		t.Fatalf("failed to extract function: %v", err)
	}

	t.Logf("%+v", endpoint)

	endpoint, _, err = extractFunc(Fn3, NewOptions())
	if err != nil {
		t.Fatalf("failed to extract function: %v", err)
	}

	t.Logf("%+v", endpoint)

	endpoint, _, err = extractFunc(Fn4, NewOptions())
	if err != nil {
		t.Fatalf("failed to extract function: %v", err)
	}

	t.Logf("%+v", endpoint)
}
