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

package scheduler

import (
	"fmt"

	otelMetric "go.opentelemetry.io/otel/metric"

	"github.com/olive-io/gflow/pkg/metrics"
)

var (
	processCounter      metrics.ObserveGauge
	taskCounter         metrics.ObserveGauge
	taskCommitCounter   metrics.ObserveGauge
	taskRollbackCounter metrics.ObserveGauge
	taskDestroyCounter  metrics.ObserveGauge
)

func InitMetrics() error {
	var err error

	if err = metrics.InitMeter(""); err != nil {
		return fmt.Errorf("initialize meter: %v", err)
	}

	processCounter, err = metrics.NewObserveGauge("process_count", otelMetric.WithDescription("the number of Bpmn Process"))
	if err != nil {
		return err
	}

	taskCounter, err = metrics.NewObserveGauge("task_count", otelMetric.WithDescription("the number of Task"))
	if err != nil {
		return err
	}

	taskCommitCounter, err = metrics.NewObserveGauge("task_commit_count", otelMetric.WithDescription("the number of Task commit"))
	if err != nil {
		return err
	}

	taskRollbackCounter, err = metrics.NewObserveGauge("task_rollback_count", otelMetric.WithDescription("the number of Task rollback"))
	if err != nil {
		return err
	}

	taskDestroyCounter, err = metrics.NewObserveGauge("task_destroy_count", otelMetric.WithDescription("the number of Task destroy"))
	if err != nil {
		return err
	}

	return nil
}
