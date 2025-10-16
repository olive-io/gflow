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

package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/olive-io/gflow/api/types"
)

type TestData1 struct {
	Name string
}

type TestData struct {
	ContentType string         `gflow:"hr:content_type"`
	D1          *TestData1     `gflow:"dt:d1"`
	P1          int32          `json:"p1"`
	P2          string         `json:"p2"`
	S1          []string       `json:"s1"`
	M1          map[string]any `json:"m1"`
	B1          []byte         `json:"b1"`
}

func TestInjectValue(t *testing.T) {
	target := &TestData{}
	req := &Request{
		Headers: map[string]string{
			"Content_type": "application/json",
		},
		Properties: map[string]*types.Value{
			"p1": types.NewValue(1),
			"p2": types.NewValue("a"),
			"s1": types.NewValue([]string{"a"}),
			"m1": types.NewValue(map[string]interface{}{"a": "b"}),
			"b1": types.NewValue([]byte("a")),
		},
		DataObjects: map[string]*types.Value{
			"d1": types.NewValue(&TestData1{Name: "test1"}),
		},
	}

	if err := req.ApplyTo(target); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, target.D1, &TestData1{Name: "test1"})
	assert.Equal(t, target.P1, int32(1))
	assert.Equal(t, target.P2, "a")
	assert.Equal(t, target.S1, []string{"a"})
	assert.Equal(t, target.M1, map[string]any{"a": "b"})
	assert.Equal(t, target.B1, []byte("a"))
}
