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

// func TestMetricsWrapper_Record_Label_Update(t *testing.T) {
// 	mw := &service.MetricsWrapper{
// 		Meter: otel.Meter("powerflex-test"),
// 	}
// 	metaFirst := &service.VolumeMeta{
// 		Name:                      "newVolume",
// 		ID:                        "123",
// 		PersistentVolumeName:      "pvol0",
// 		PersistentVolumeClaimName: "pvc0",
// 		Namespace:                 "namespace0",
// 		MappedSDCs: []service.MappedSDC{
// 			{
// 				SdcID: "111",
// 				SdcIP: "1.2.3.4",
// 			},
// 		},
// 	}

// 	metaSecond := &service.VolumeMeta{
// 		Name:                      "newVolume",
// 		ID:                        "123",
// 		PersistentVolumeName:      "pvol0",
// 		PersistentVolumeClaimName: "pvc0",
// 		Namespace:                 "namespace0",
// 		MappedSDCs: []service.MappedSDC{
// 			{
// 				SdcID: "111",
// 				SdcIP: "1.2.3.4",
// 			},
// 		},
// 	}

// 	metaThird := &service.VolumeMeta{
// 		Name:                      "newVolume",
// 		ID:                        "123",
// 		PersistentVolumeName:      "pvol1",
// 		PersistentVolumeClaimName: "pvc1",
// 		Namespace:                 "namespace0",
// 		MappedSDCs: []service.MappedSDC{
// 			{
// 				SdcID: "111",
// 				SdcIP: "1.2.3.4",
// 			},
// 		},
// 	}

// 	expectedLables := []attribute.KeyValue{
// 		attribute.String("VolumeID", metaSecond.ID),
// 		attribute.String("PlotWithMean", "No"),
// 		attribute.String("PersistentVolumeName", metaSecond.PersistentVolumeName),
// 		attribute.String("PersistentVolumeClaimName", metaSecond.PersistentVolumeClaimName),
// 		attribute.String("Namespace", metaSecond.Namespace),
// 	}
// 	expectedLablesUpdate := []attribute.KeyValue{
// 		attribute.String("VolumeID", metaThird.ID),
// 		attribute.String("PlotWithMean", "No"),
// 		attribute.String("PersistentVolumeName", metaThird.PersistentVolumeName),
// 		attribute.String("PersistentVolumeClaimName", metaThird.PersistentVolumeClaimName),
// 		attribute.String("Namespace", metaThird.Namespace),
// 	}

// 	exporter := &otlexporters.OtlCollectorExporter{}
// 	err := exporter.InitExporter()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	t.Run("success: volume metric labels updated", func(t *testing.T) {
// 		err := mw.Record(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #1), got %v", err)
// 		}
// 		err = mw.Record(context.Background(), metaSecond, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #2), got %v", err)
// 		}

// 		newLabels, ok := mw.Labels.Load(metaFirst.ID)
// 		if !ok {
// 			t.Errorf("expected labels to exist for %v, but did not find them", metaFirst.ID)
// 		}
// 		labels := newLabels.([]attribute.KeyValue)
// 		for _, l := range labels {
// 			for _, e := range expectedLables {
// 				if l.Key == e.Key {
// 					if l.Value.AsString() != e.Value.AsString() {
// 						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
// 					}
// 				}
// 			}
// 		}
// 	})

// 	t.Run("success: volume metric labels updated with PV Name and PVC Update", func(t *testing.T) {
// 		err := mw.Record(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #1), got %v", err)
// 		}
// 		err = mw.Record(context.Background(), metaThird, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #2), got %v", err)
// 		}

// 		newLabels, ok := mw.Labels.Load(metaThird.ID)
// 		if !ok {
// 			t.Errorf("expected labels to exist for %v, but did not find them", metaThird.ID)
// 		}
// 		labels := newLabels.([]attribute.KeyValue)
// 		for _, l := range labels {
// 			for _, e := range expectedLablesUpdate {
// 				if l.Key == e.Key {
// 					if l.Value.AsString() != e.Value.AsString() {
// 						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
// 					}
// 				}
// 			}
// 		}
// 	})
// }

// func Test_Volume_Metrics_Label_Update(t *testing.T) {
// 	mw := &service.MetricsWrapper{
// 		Meter: otel.Meter("powerstore-test"),
// 	}

// 	metaFirst := &service.VolumeMeta{
// 		Name:                      "newVolume",
// 		ID:                        "123",
// 		PersistentVolumeName:      "pvol1",
// 		PersistentVolumeClaimName: "pvc1",
// 		Namespace:                 "namespace1",
// 		MappedSDCs: []service.MappedSDC{
// 			{
// 				SdcID: "111",
// 				SdcIP: "1.2.3.4",
// 			},
// 		},
// 	}

// 	metaSecond := &service.VolumeMeta{
// 		Name:                      "newVolume",
// 		ID:                        "123",
// 		PersistentVolumeName:      "pvol1",
// 		PersistentVolumeClaimName: "pvc1",
// 		Namespace:                 "namespace2",
// 		MappedSDCs: []service.MappedSDC{
// 			{
// 				SdcID: "222",
// 				SdcIP: "20.20.20.20",
// 			},
// 		},
// 	}

// 	metaThird := &service.VolumeMeta{
// 		Name:                      "newVolume",
// 		ID:                        "123",
// 		PersistentVolumeName:      "pvol2",
// 		PersistentVolumeClaimName: "pvc2",
// 		Namespace:                 "namespace2",
// 		MappedSDCs: []service.MappedSDC{
// 			{
// 				SdcID: "222",
// 				SdcIP: "20.20.20.20",
// 			},
// 		},
// 	}

// 	expectedLables := []attribute.KeyValue{
// 		attribute.String("VolumeID", metaSecond.ID),
// 		attribute.String("VolumeName", metaSecond.Name),
// 		attribute.String("PersistentVolumeName", metaSecond.PersistentVolumeName),
// 		attribute.String("PersistentVolumeClaimName", metaSecond.PersistentVolumeClaimName),
// 		attribute.String("Namespace", metaSecond.Namespace),
// 		attribute.String("MappedNodeIDs", "__"+metaSecond.MappedSDCs[0].SdcID+"__"),
// 		attribute.String("MappedNodeIPs", "__"+metaSecond.MappedSDCs[0].SdcIP+"__"),
// 		attribute.String("PlotWithMean", "No"),
// 	}

// 	expectedLablesUpdate := []attribute.KeyValue{
// 		attribute.String("VolumeID", metaThird.ID),
// 		attribute.String("PlotWithMean", "No"),
// 		attribute.String("PersistentVolumeName", metaThird.PersistentVolumeName),
// 		attribute.String("PersistentVolumeClaimName", metaThird.PersistentVolumeClaimName),
// 		attribute.String("Namespace", metaThird.Namespace),
// 	}

// 	t.Run("success: volume metric labels updated", func(t *testing.T) {
// 		err := mw.Record(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #1), got %v", err)
// 		}
// 		err = mw.Record(context.Background(), metaSecond, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #2), got %v", err)
// 		}

// 		newLabels, ok := mw.Labels.Load(metaFirst.ID)
// 		if !ok {
// 			t.Errorf("expected labels to exist for %v, but did not find them", metaFirst.ID)
// 		}
// 		labels := newLabels.([]attribute.KeyValue)
// 		for _, l := range labels {
// 			for _, e := range expectedLables {
// 				if l.Key == e.Key {
// 					if l.Value.AsString() != e.Value.AsString() {
// 						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
// 					}
// 				}
// 			}
// 		}
// 	})

// 	t.Run("success: volume metric labels updated with PV Name and PVC Update", func(t *testing.T) {
// 		err := mw.Record(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #1), got %v", err)
// 		}
// 		err = mw.Record(context.Background(), metaThird, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #2), got %v", err)
// 		}

// 		newLabels, ok := mw.Labels.Load(metaThird.ID)
// 		if !ok {
// 			t.Errorf("expected labels to exist for %v, but did not find them", metaThird.ID)
// 		}
// 		labels := newLabels.([]attribute.KeyValue)
// 		for _, l := range labels {
// 			for _, e := range expectedLablesUpdate {
// 				if l.Key == e.Key {
// 					if l.Value.AsString() != e.Value.AsString() {
// 						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
// 					}
// 				}
// 			}
// 		}
// 	})
// }

// func Test_Sdc_Metrics_Label_Update(t *testing.T) {
// 	mw := &service.MetricsWrapper{
// 		Meter: otel.Meter("powerstore-test"),
// 	}

// 	metaFirst := &service.SDCMeta{
// 		Name:    "newVolume",
// 		ID:      "123",
// 		IP:      "10.20.20.20",
// 		SdcGUID: "sample-guid",
// 	}

// 	metaSecond := &service.SDCMeta{
// 		Name:    "newVolume",
// 		ID:      "123",
// 		IP:      "10.20.20.20",
// 		SdcGUID: "sample-guid",
// 	}

// 	expectedLables := []attribute.KeyValue{
// 		attribute.String("ID", metaFirst.ID),
// 		attribute.String("Name", metaFirst.Name),
// 		attribute.String("IP", metaFirst.IP),
// 		attribute.String("NodeGUID", metaFirst.SdcGUID),
// 		attribute.String("PlotWithMean", "No"),
// 	}

// 	expectedLablesUpdate := []attribute.KeyValue{
// 		attribute.String("ID", metaSecond.ID),
// 		attribute.String("Name", metaSecond.Name),
// 		attribute.String("IP", metaSecond.IP),
// 		attribute.String("NodeGUID", metaSecond.SdcGUID),
// 		attribute.String("PlotWithMean", "No"),
// 	}

// 	t.Run("success: volume metric labels updated", func(t *testing.T) {
// 		err := mw.Record(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #1), got %v", err)
// 		}
// 		err = mw.Record(context.Background(), metaSecond, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #2), got %v", err)
// 		}

// 		newLabels, ok := mw.Labels.Load(metaFirst.ID)
// 		if !ok {
// 			t.Errorf("expected labels to exist for %v, but did not find them", metaFirst.ID)
// 		}
// 		labels := newLabels.([]attribute.KeyValue)
// 		for _, l := range labels {
// 			for _, e := range expectedLables {
// 				if l.Key == e.Key {
// 					if l.Value.AsString() != e.Value.AsString() {
// 						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
// 					}
// 				}
// 			}
// 		}
// 	})

// 	t.Run("success: volume metric labels updated with PV Name and PVC Update", func(t *testing.T) {
// 		err := mw.Record(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #1), got %v", err)
// 		}
// 		err = mw.Record(context.Background(), metaSecond, 1, 2, 3, 4, 5, 6)
// 		if err != nil {
// 			t.Errorf("expected nil error (record #2), got %v", err)
// 		}

// 		newLabels, ok := mw.Labels.Load(metaFirst.ID)
// 		if !ok {
// 			t.Errorf("expected labels to exist for %v, but did not find them", metaFirst.ID)
// 		}
// 		labels := newLabels.([]attribute.KeyValue)
// 		for _, l := range labels {
// 			for _, e := range expectedLablesUpdate {
// 				if l.Key == e.Key {
// 					if l.Value.AsString() != e.Value.AsString() {
// 						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
// 					}
// 				}
// 			}
// 		}
// 	})
// }

type MockStoragePoolStatisticsGetter struct{}

func (m *MockStoragePoolStatisticsGetter) GetStatistics() (*types.Statistics, error) {
	return &types.Statistics{}, nil
}

func TestMetricsWrapper_RecordCapacity(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerflex-test"),
	}

	storageClassMeta := service.StorageClassMeta{
		ID:              "test-id",
		Name:            "test-name",
		Driver:          "csi-vxflexos.dellemc.com",
		StorageSystemID: "test-system-id",
		StoragePools: map[string]service.StoragePoolStatisticsGetter{
			"pool1": &MockStoragePoolStatisticsGetter{},
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
