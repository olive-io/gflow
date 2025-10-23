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

package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	otelMetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/atomic"
)

var (
	DefaultNamespace = "olive"
	DefaultSubsystem = "gflow"

	meter otelMetric.Meter
	once  sync.Once
)

func InitMeter(name string) error {
	var err error
	once.Do(func() {
		var exporter *prometheus.Exporter
		exporter, err = prometheus.New()
		if err != nil {
			err = fmt.Errorf("initialize prometheus exporter: %w", err)
			return
		}
		provider := metric.NewMeterProvider(metric.WithReader(exporter))
		meter = provider.Meter(name)
	})

	return err
}

func NewInt64Gauge(name string, opts ...otelMetric.Int64GaugeOption) (otelMetric.Int64Gauge, error) {
	gauge, err := meter.Int64Gauge(name, opts...)
	if err != nil {
		return nil, err
	}

	return gauge, nil
}

type ObserveGauge interface {
	otelMetric.Float64ObservableGauge
	Set(v float64)
	Inc()
	Dec()
	Add(v float64)
	Sub(v float64)
	SetToCurrentTime()
	Get() float64
}

type gaugeImpl struct {
	otelMetric.Float64ObservableGauge
	value *atomic.Float64
}

func NewObserveGauge(name string, opts ...otelMetric.Float64ObservableGaugeOption) (ObserveGauge, error) {
	gi := &gaugeImpl{
		value: atomic.NewFloat64(0),
	}
	gauge, err := meter.Float64ObservableGauge(name, opts...)
	if err != nil {
		return nil, err
	}
	_, err = meter.RegisterCallback(func(ctx context.Context, observer otelMetric.Observer) error {
		observer.ObserveFloat64(gauge, gi.value.Load(),
			otelMetric.WithAttributes(
				attribute.String("namespace", DefaultNamespace),
				attribute.String("subsystem", DefaultSubsystem)))
		return nil
	}, gauge)

	gi.Float64ObservableGauge = gauge

	return gi, nil
}

func (g *gaugeImpl) Set(v float64) {
	g.value.Store(v)
}

func (g *gaugeImpl) Inc() {
	g.value.Add(1)
}

func (g *gaugeImpl) Dec() {
	g.value.Sub(1)
}

func (g *gaugeImpl) Add(v float64) {
	g.value.Add(v)
}

func (g *gaugeImpl) Sub(v float64) {
	g.value.Sub(v)
}

func (g *gaugeImpl) SetToCurrentTime() {
	g.Set(float64(time.Now().Unix()))
}

func (g *gaugeImpl) Get() float64 {
	return g.value.Load()
}
