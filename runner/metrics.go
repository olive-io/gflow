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
	goruntime "runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"

	"github.com/olive-io/gflow/pkg/metrics"
)

var (
	currentVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metrics.DefaultNamespace,
		Subsystem: metrics.DefaultSubsystem,
		Name:      "version",
		Help:      "Which version is running. 1 for 'runner_version' label with current version.",
	},
		[]string{"runner_version"})
	currentGoVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metrics.DefaultNamespace,
		Subsystem: metrics.DefaultSubsystem,
		Name:      "go_version",
		Help:      "Which Go version runner is running with. 1 for 'runner_go_version' label with current version.",
	},
		[]string{"runner_go_version"})

	taskCounter = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: metrics.DefaultNamespace,
		Subsystem: metrics.DefaultSubsystem,
		Name:      "task_count",
		Help:      "the number of Task",
	})

	taskCommitCounter = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: metrics.DefaultNamespace,
		Subsystem: metrics.DefaultSubsystem,
		Name:      "task_commit_count",
		Help:      "the number of Task on commit",
	})

	taskRollbackCounter = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: metrics.DefaultNamespace,
		Subsystem: metrics.DefaultSubsystem,
		Name:      "task_rollback_count",
		Help:      "the number of Task on rollback",
	})

	taskDestroyCounter = metrics.NewGauge(prometheus.GaugeOpts{
		Namespace: metrics.DefaultNamespace,
		Subsystem: metrics.DefaultSubsystem,
		Name:      "task_destroy_count",
		Help:      "the number of Task on destroy",
	})
)

func init() {
	prometheus.MustRegister(currentVersion)
	prometheus.MustRegister(currentGoVersion)
	prometheus.MustRegister(taskCounter)
	prometheus.MustRegister(taskCommitCounter)
	prometheus.MustRegister(taskRollbackCounter)
	prometheus.MustRegister(taskDestroyCounter)

	currentVersion.With(prometheus.Labels{
		"runner_version": version.Version,
	}).Set(1)

	currentGoVersion.With(prometheus.Labels{
		"runner_go_version": goruntime.Version(),
	}).Set(1)
}
