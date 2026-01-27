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

package service_test

import (
	"context"
	"testing"

	types "github.com/dell/goscaleio/types/v1"
	"github.com/dell/karavi-metrics-powerflex/internal/service"
	otlexporters "github.com/dell/karavi-metrics-powerflex/opentelemetry/exporters"
	"go.opentelemetry.io/otel"
)

func TestMetricsWrapper_Record(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerflex-test"),
	}
	volumeMetas := []interface{}{
		&service.VolumeMeta{
			ID: "123",
		},
		&service.SDCMeta{
			ID: "123",
		},
	}
	storageClassMetas := []interface{}{
		&service.StorageClassMeta{
			ID: "123",
		},
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx          context.Context
		meta         interface{}
		readBW       float64
		writeBW      float64
		readIOPS     float64
		writeIOPS    float64
		readLatency  float64
		writeLatency float64
	}
	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:          context.Background(),
				meta:         volumeMetas[0],
				readBW:       1,
				writeBW:      2,
				readIOPS:     3,
				writeIOPS:    4,
				readLatency:  5,
				writeLatency: 6,
			},
			wantErr: false,
		},
		{
			name: "fail",
			mw:   mw,
			args: args{
				ctx:          context.Background(),
				meta:         storageClassMetas[0],
				readBW:       1,
				writeBW:      2,
				readIOPS:     3,
				writeIOPS:    4,
				readLatency:  5,
				writeLatency: 6,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.Record(tt.args.ctx, tt.args.meta, tt.args.readBW, tt.args.writeBW, tt.args.readIOPS, tt.args.writeIOPS, tt.args.readLatency, tt.args.writeLatency); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.Record() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type MockStoragePoolStatisticsGetter struct{}

func (m *MockStoragePoolStatisticsGetter) GetStatistics() (*types.Statistics, error) {
	return &types.Statistics{}, nil
}

func TestMetricsWrapper_RecordCapacity(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerflex-test"),
	}
	retriever := newRetriever(t, "v1")
	storageClassMeta := service.StorageClassMeta{
		ID:              "test-id",
		Name:            "test-name",
		Driver:          "csi-vxflexos.dellemc.com",
		StorageSystemID: "test-system-id",
		StoragePools: map[string]service.StoragePoolMetricsRetriever{
			"pool1": retriever,
		},
	}
	volumeMeta := &service.VolumeMeta{
		Name: "newVolume",
	}
	totalLogicalCapacity := 100.0
	logicalCapacityAvailable := 50.0
	logicalCapacityInUse := 30.0
	logicalProvisioned := 20.0
	type args struct {
		ctx                      context.Context
		meta                     interface{}
		totalLogicalCapacity     float64
		logicalCapacityAvailable float64
		logicalCapacityInUse     float64
		logicalProvisioned       float64
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:                      context.Background(),
				meta:                     storageClassMeta,
				totalLogicalCapacity:     totalLogicalCapacity,
				logicalCapacityAvailable: logicalCapacityAvailable,
				logicalCapacityInUse:     logicalCapacityInUse,
				logicalProvisioned:       logicalProvisioned,
			},
			wantErr: false,
		},
		{
			name: "fail",
			mw:   mw,
			args: args{
				ctx:                      context.Background(),
				meta:                     volumeMeta,
				totalLogicalCapacity:     totalLogicalCapacity,
				logicalCapacityAvailable: logicalCapacityAvailable,
				logicalCapacityInUse:     logicalCapacityInUse,
				logicalProvisioned:       logicalProvisioned,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.RecordCapacity(tt.args.ctx, tt.args.meta, tt.args.totalLogicalCapacity, tt.args.logicalCapacityAvailable, tt.args.logicalCapacityInUse, tt.args.logicalProvisioned); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.RecordCapacity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsWrapper_RecordTopologyMetrics(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerflex-test"),
	}
	tests := []struct {
		name    string
		meta    interface{}
		metric  *service.TopologyMetricsRecord
		wantErr bool
	}{
		{
			name: "success",
			meta: &service.TopologyMeta{
				PersistentVolume: "test-pv",
			},
			metric:  &service.TopologyMetricsRecord{},
			wantErr: false,
		},
		{
			name:    "unknown meta data type",
			meta:    "unknown",
			metric:  &service.TopologyMetricsRecord{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mw.RecordTopologyMetrics(context.Background(), tt.meta, tt.metric); (err != nil) != tt.wantErr {
				t.Errorf("RecordTopologyMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
