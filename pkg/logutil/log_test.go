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

package logutil

import (
	"testing"

	"go.uber.org/zap"
)

func TestLog(t *testing.T) {
	lc := NewLogConfig()
	lc.Format = "console"
	err := lc.SetupLogging()
	if err != nil {
		t.Fatal(err)
	}

	lc.GetLogger().With(zap.String("a", "b")).Debug("info message")
	lc.GetLogger().With(zap.String("a", "b")).Info("info message")
	lc.GetLogger().With(zap.String("a", "b")).Warn("info message")
	lc.GetLogger().With(zap.String("a", "b")).Error("info message")
}
