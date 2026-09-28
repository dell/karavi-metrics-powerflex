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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// OtlCollectorExporter is the exporter for the OpenTelemetry Collector
type OtlCollectorExporter struct {
	CollectorAddr       string
	exporter            *otlpmetricgrpc.Exporter
	controller          *metric.MeterProvider
	recordExportFailure func()
}

const (
	// DefaultCollectorCertPath is the default location to look for the Collector certificate
	DefaultCollectorCertPath = "/etc/ssl/certs/cert.crt"
)

// SetExportFailureRecorder configures a callback that is invoked whenever an OTLP export fails.
func (c *OtlCollectorExporter) SetExportFailureRecorder(recordExportFailure func()) {
	c.recordExportFailure = recordExportFailure
}

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

	err = c.controller.Shutdown(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (c *OtlCollectorExporter) initOTLPExporter(opts ...otlpmetricgrpc.Option) (*otlpmetricgrpc.Exporter, *metric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(context.Background(), opts...)
	if err != nil {
		return nil, nil, err
	}

	recordingExporter := newRecordingExporter(exporter, c)
	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(recordingExporter, metric.WithInterval(5*time.Second))))

	otel.SetMeterProvider(meterProvider)

	return exporter, meterProvider, nil
}

type recordingExporter struct {
	*otlpmetricgrpc.Exporter
	owner *OtlCollectorExporter
}

func newRecordingExporter(base *otlpmetricgrpc.Exporter, owner *OtlCollectorExporter) metric.Exporter {
	return &recordingExporter{Exporter: base, owner: owner}
}

func (r *recordingExporter) Export(ctx context.Context, metrics *metricdata.ResourceMetrics) error {
	err := r.Exporter.Export(ctx, metrics)
	if err != nil && r.owner != nil && r.owner.recordExportFailure != nil {
		r.owner.recordExportFailure()
	}
	return err
}
