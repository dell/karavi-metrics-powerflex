/*
 Copyright (c) 2020-2022 Dell Inc. or its subsidiaries. All Rights Reserved.

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

package otlexporters

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

// OtlCollectorExporter is the exporter for the OpenTelemetry Collector
type OtlCollectorExporter struct {
	CollectorAddr string
	exporter      *otlpmetric.Exporter
	controller    *controller.Controller
}

const (
	// DefaultCollectorCertPath is the default location to look for the Collector certificate
	DefaultCollectorCertPath = "/etc/ssl/certs/cert.crt"
)

// InitExporter is the initialization method for the OpenTelemetry Collector exporter
func (c *OtlCollectorExporter) InitExporter(opts ...otlpmetricgrpc.Option) error {
	exporter, controller, err := c.initOTLPExporter(opts...)
	if err != nil {
		return err
	}
	c.exporter = exporter
	c.controller = controller

	return err
}

// StopExporter stops the activity of the Otl Collector's required services
func (c *OtlCollectorExporter) StopExporter() error {
	err := c.exporter.Shutdown(context.Background())
	if err != nil {
		return err
	}

	err = c.controller.Stop(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (c *OtlCollectorExporter) initOTLPExporter(opts ...otlpmetricgrpc.Option) (*otlpmetric.Exporter, *controller.Controller, error) {
	exporter, err := otlpmetricgrpc.New(context.Background(), opts...)
	if err != nil {
		return nil, nil, err
	}

	processor := basic.New(
		simple.NewWithHistogramDistribution(),
		exporter,
	)

	factory := basic.NewFactory(
		processor.AggregatorSelector,
		processor.TemporalitySelector,
	)

	ctrl := controller.New(
		factory,
		controller.WithExporter(exporter),
		controller.WithCollectPeriod(5*time.Second),
	)

	err = ctrl.Start(context.Background())
	if err != nil {
		return nil, nil, err
	}

	global.SetMeterProvider(ctrl)

	return exporter, ctrl, nil
}
