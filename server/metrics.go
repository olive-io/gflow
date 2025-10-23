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
	"fmt"
	goruntime "runtime"

	"github.com/prometheus/common/version"
	"go.opentelemetry.io/otel/attribute"
	otelMetric "go.opentelemetry.io/otel/metric"

	"github.com/olive-io/gflow/pkg/metrics"
)

var (
	currentVersion   otelMetric.Int64Gauge
	currentGoVersion otelMetric.Int64Gauge
)

func InitMetrics(name string) error {
	if err := metrics.InitMeter(name); err != nil {
		return fmt.Errorf("initialize metrics: %w", err)
	}

	var err error

	ctx := context.Background()
	currentVersion, err = metrics.NewInt64Gauge("gflow_version", otelMetric.WithDescription("Which version is running. 1 for 'runner_version' label with current version."))
	if err != nil {
		return err
	}

	opts := otelMetric.WithAttributes(
		attribute.Key("namespace").String(metrics.DefaultNamespace),
		attribute.Key("subsystem").String(metrics.DefaultSubsystem),
		attribute.Key("runner_version").String(version.Version),
	)
	currentVersion.Record(ctx, 1, opts)

	currentGoVersion, err = metrics.NewInt64Gauge("gflow_go_version", otelMetric.WithDescription("Which Go version runner is running with. 1 for 'runner_go_version' label with current version."))
	if err != nil {
		return err
	}
	opts = otelMetric.WithAttributes(
		attribute.Key("namespace").String(metrics.DefaultNamespace),
		attribute.Key("subsystem").String(metrics.DefaultSubsystem),
		attribute.Key("runner_go_version").String(goruntime.Version()),
	)
	currentGoVersion.Record(ctx, 1, opts)

	return err
}
