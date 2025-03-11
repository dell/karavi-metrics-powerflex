/*
 Copyright (c) 2025 Dell Inc. or its subsidiaries. All Rights Reserved.

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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
)

func TestInitExporter(t *testing.T) {
	tests := []struct {
		name          string
		collector     *OtlCollectorExporter
		opts          []otlpmetricgrpc.Option
		ExpectedError error
	}{
		{
			name: "Successful Exporter Initialization",
			collector: &OtlCollectorExporter{
				CollectorAddr: "localhost:8080",
			},
			opts: []otlpmetricgrpc.Option{
				otlpmetricgrpc.WithInsecure(),
			},
			ExpectedError: nil,
		},
		{
			name: "Invalid Service Config",
			collector: &OtlCollectorExporter{
				CollectorAddr: "localhost:8080",
			},
			opts: []otlpmetricgrpc.Option{
				otlpmetricgrpc.WithServiceConfig("invalid config"),
			},
			ExpectedError: errors.New("grpc: the provided default service config is invalid: invalid character 'i' looking for beginning of value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.collector.InitExporter(tt.opts...)
			assert.Equal(t, err, tt.ExpectedError)
		})
	}
}

func TestOtlCollectorExporter_StopExporter(t *testing.T) {
	tests := []struct {
		name          string
		collector     *OtlCollectorExporter
		opts          []otlpmetricgrpc.Option
		preShutdown   bool
		ExpectedError error
	}{
		{
			name: "Error: gRPC exporter is shutdown",
			collector: &OtlCollectorExporter{
				CollectorAddr: "localhost:8080",
			},
			opts: []otlpmetricgrpc.Option{
				otlpmetricgrpc.WithInsecure(),
			},
			preShutdown:   true,
			ExpectedError: errors.New("gRPC exporter is shutdown"),
		},
		{
			name: "Shutdown Exporter",
			collector: &OtlCollectorExporter{
				CollectorAddr: "localhost:8080",
			},
			opts: []otlpmetricgrpc.Option{
				otlpmetricgrpc.WithInsecure(),
				otlpmetricgrpc.WithEndpoint("localhost:8080"),
			},
			preShutdown:   false,
			ExpectedError: errors.New("gRPC exporter is shutdown"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.collector.InitExporter(tt.opts...)
			if err != nil {
				t.Fatal(err)
			}

			if tt.preShutdown {
				_ = tt.collector.exporter.Shutdown(context.Background())
			}

			err = tt.collector.StopExporter()
			if err != nil && tt.ExpectedError == nil {
				t.Fatal(err)
			}
		})
	}
}
