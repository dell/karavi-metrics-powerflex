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
	"errors"
	"testing"

	"github.com/dell/karavi-metrics-powerflex/internal/service"
	"github.com/dell/karavi-metrics-powerflex/internal/service/mocks"
	"github.com/golang/mock/gomock"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
)

func Test_Metrics_Record(t *testing.T) {
	type checkFn func(*testing.T, error)
	checkFns := func(checkFns ...checkFn) []checkFn { return checkFns }

	verifyError := func(t *testing.T, err error) {
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
	}

	verifyNoError := func(t *testing.T, err error) {
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	}

	metas := []interface{}{
		&service.VolumeMeta{
			Name:                      "newVolume",
			ID:                        "123",
			PersistentVolumeName:      "pvol0",
			PersistentVolumeClaimName: "pvc0",
			Namespace:                 "namespace0",
			MappedSDCs: []service.MappedSDC{
				{
					SdcID: "111",
					SdcIP: "1.2.3.4",
				},
			},
		},
		&service.SDCMeta{
			Name:    "newSDC",
			ID:      "123",
			IP:      "1.2.3.5",
			SdcGUID: "321",
		},
	}

	tests := map[string]func(t *testing.T) ([]*service.MetricsWrapper, []checkFn){
		"success": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)

			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				provider := mocks.NewMockInstrumentProvider(ctrl)
				otMeter := global.Meter(prefix + "_test")
				readBW, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readLatency, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				writeLatency, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "write_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().AsyncFloat64().Return(provider).Times(6)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readLatency, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeLatency, nil)

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerflex_volume_"),
				getMeter("powerflex_export_node_"),
			}

			return mws, checkFns(verifyNoError)
		},
		"error creating read_bw_megabytes_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
			provider := mocks.NewMockInstrumentProvider(ctrl)
			otMeter := global.Meter("")
			empty, err := otMeter.AsyncFloat64().UpDownCounter("")
			if err != nil {
				t.Fatal(err)
			}

			meter.EXPECT().AsyncFloat64().Return(provider).Times(2)
			provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error")).Times(2)

			mws := []*service.MetricsWrapper{{Meter: meter}, {Meter: meter}}

			return mws, checkFns(verifyError)
		},
		"error creating write_bw_megabytes_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				provider := mocks.NewMockInstrumentProvider(ctrl)
				otMeter := global.Meter(prefix + "_test")
				readBW, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}
				empty, err := otMeter.AsyncFloat64().UpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}
				meter.EXPECT().AsyncFloat64().Return(provider).Times(2)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerflex_volume_"),
				getMeter("powerflex_export_node_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating read_iops_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				provider := mocks.NewMockInstrumentProvider(ctrl)
				otMeter := global.Meter(prefix + "_test")

				readBW, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				empty, err := otMeter.AsyncFloat64().UpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().AsyncFloat64().Return(provider).Times(3)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerflex_volume_"),
				getMeter("powerflex_export_node_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating write_iops_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				provider := mocks.NewMockInstrumentProvider(ctrl)
				otMeter := global.Meter(prefix + "_test")

				readBW, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				empty, err := otMeter.AsyncFloat64().UpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().AsyncFloat64().Return(provider).Times(4)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerflex_volume_"),
				getMeter("powerflex_export_node_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating read_latency_milliseconds": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				provider := mocks.NewMockInstrumentProvider(ctrl)
				otMeter := global.Meter(prefix + "_test")

				readBW, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				empty, err := otMeter.AsyncFloat64().UpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}
				meter.EXPECT().AsyncFloat64().Return(provider).Times(5)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerflex_volume_"),
				getMeter("powerflex_export_node_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating write_latency_milliseconds": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				provider := mocks.NewMockInstrumentProvider(ctrl)
				otMeter := global.Meter(prefix + "_test")

				readBW, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readLatency, err := otMeter.AsyncFloat64().UpDownCounter(prefix + "read_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				empty, err := otMeter.AsyncFloat64().UpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().AsyncFloat64().Return(provider).Times(6)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readLatency, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerflex_volume_"),
				getMeter("powerflex_export_node_"),
			}

			return mws, checkFns(verifyError)
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mws, checks := tc(t)
			for i := range mws {
				err := mws[i].Record(context.Background(), metas[i], 1, 2, 3, 4, 5, 6)
				for _, check := range checks {
					check(t, err)
				}
			}
		})
	}
}

func Test_Metrics_RecordCapacity(t *testing.T) {
	type checkFn func(*testing.T, error)
	checkFns := func(checkFns ...checkFn) []checkFn { return checkFns }

	verifyError := func(t *testing.T, err error) {
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
	}

	verifyNoError := func(t *testing.T, err error) {
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	}

	tests := map[string]func(t *testing.T) (*service.MetricsWrapper, service.StorageClassMeta, []checkFn){
		"success": func(t *testing.T) (*service.MetricsWrapper, service.StorageClassMeta, []checkFn) {
			ctrl := gomock.NewController(t)

			meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
			provider := mocks.NewMockInstrumentProvider(ctrl)
			otMeter := global.Meter("test")

			totalLogicalCapacity, err := otMeter.AsyncFloat64().UpDownCounter("TotalLogicalCapacity")
			if err != nil {
				t.Fatal(err)
			}

			logicalCapacityAvailable, err := otMeter.AsyncFloat64().UpDownCounter("LogicalCapacityAvailable")
			if err != nil {
				t.Fatal(err)
			}

			logicalCapacityInUse, err := otMeter.AsyncFloat64().UpDownCounter("LogicalCapacityInUse")
			if err != nil {
				t.Fatal(err)
			}

			logicalProvisioned, err := otMeter.AsyncFloat64().UpDownCounter("LogicalProvisioned")
			if err != nil {
				t.Fatal(err)
			}

			meter.EXPECT().AsyncFloat64().Return(provider).Times(4)
			provider.EXPECT().UpDownCounter(gomock.Any()).Return(totalLogicalCapacity, nil)
			provider.EXPECT().UpDownCounter(gomock.Any()).Return(logicalCapacityAvailable, nil)
			provider.EXPECT().UpDownCounter(gomock.Any()).Return(logicalCapacityInUse, nil)
			provider.EXPECT().UpDownCounter(gomock.Any()).Return(logicalProvisioned, nil)

			mw := &service.MetricsWrapper{
				Meter: meter,
			}

			scMeta := service.StorageClassMeta{
				ID:     "123",
				Name:   "test",
				Driver: "csi-vxflexos.dellemc.com",
				StoragePools: map[string]service.StoragePoolStatisticsGetter{
					"pool-1": nil,
				},
			}

			return mw, scMeta, checkFns(verifyNoError)
		},
		"error creating CapacityInUse": func(t *testing.T) (*service.MetricsWrapper, service.StorageClassMeta, []checkFn) {
			ctrl := gomock.NewController(t)
			meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
			provider := mocks.NewMockInstrumentProvider(ctrl)
			otMeter := global.Meter("")
			empty, err := otMeter.AsyncFloat64().UpDownCounter("")
			if err != nil {
				t.Fatal(err)
			}

			meter.EXPECT().AsyncFloat64().Return(provider).Times(1)
			provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error")).Times(1)

			mw := &service.MetricsWrapper{
				Meter: meter,
			}

			scMeta := service.StorageClassMeta{
				ID:     "123",
				Name:   "test",
				Driver: "csi-vxflexos.dellemc.com",
				StoragePools: map[string]service.StoragePoolStatisticsGetter{
					"pool-1": nil,
				},
			}

			return mw, scMeta, checkFns(verifyError)
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mw, scMeta, checks := tc(t)
			err := mw.RecordCapacity(context.Background(), scMeta, 1, 2, 3, 4)
			for _, check := range checks {
				check(t, err)
			}
		})
	}
}

func Test_Volume_Metrics_Label_Update(t *testing.T) {
	metaFirst := &service.VolumeMeta{
		Name:                      "newVolume",
		ID:                        "123",
		PersistentVolumeName:      "pvol1",
		PersistentVolumeClaimName: "pvc1",
		Namespace:                 "namespace1",
		MappedSDCs: []service.MappedSDC{
			{
				SdcID: "111",
				SdcIP: "1.2.3.4",
			},
		},
	}

	metaSecond := &service.VolumeMeta{
		Name:                      "newVolume",
		ID:                        "123",
		PersistentVolumeName:      "pvol2",
		PersistentVolumeClaimName: "pvc2",
		Namespace:                 "namespace2",
		MappedSDCs: []service.MappedSDC{
			{
				SdcID: "222",
				SdcIP: "20.20.20.20",
			},
		},
	}

	expectedLables := []attribute.KeyValue{
		attribute.String("VolumeID", metaSecond.ID),
		attribute.String("VolumeName", metaSecond.Name),
		attribute.String("PersistentVolumeName", metaSecond.PersistentVolumeName),
		attribute.String("PersistentVolumeClaimName", metaSecond.PersistentVolumeClaimName),
		attribute.String("Namespace", metaSecond.Namespace),
		attribute.String("MappedNodeIDs", "__"+metaSecond.MappedSDCs[0].SdcID+"__"),
		attribute.String("MappedNodeIPs", "__"+metaSecond.MappedSDCs[0].SdcIP+"__"),
		attribute.String("PlotWithMean", "No"),
	}

	ctrl := gomock.NewController(t)

	meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
	provider := mocks.NewMockInstrumentProvider(ctrl)
	otMeter := global.Meter("powerflex_volume__test")
	readBW, err := otMeter.AsyncFloat64().UpDownCounter("powerflex_volume_read_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	writeBW, err := otMeter.AsyncFloat64().UpDownCounter("powerflex_volume_write_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	readIOPS, err := otMeter.AsyncFloat64().UpDownCounter("powerflex_volume_read_iops_per_second")
	if err != nil {
		t.Fatal(err)
	}

	writeIOPS, err := otMeter.AsyncFloat64().UpDownCounter("powerflex_volume_write_iops_per_second")
	if err != nil {
		t.Fatal(err)
	}

	readLatency, err := otMeter.AsyncFloat64().UpDownCounter("powerflex_volume_read_latency_milliseconds")
	if err != nil {
		t.Fatal(err)
	}

	writeLatency, err := otMeter.AsyncFloat64().UpDownCounter("powerflex_volume_write_latency_milliseconds")
	if err != nil {
		t.Fatal(err)
	}

	meter.EXPECT().AsyncFloat64().Return(provider).Times(12)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readLatency, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeLatency, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readLatency, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeLatency, nil)

	mw := &service.MetricsWrapper{
		Meter: meter,
	}

	t.Run("success: volume metric labels updated", func(t *testing.T) {
		err := mw.Record(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.Record(context.Background(), metaSecond, 1, 2, 3, 4, 5, 6)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(metaFirst.ID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", metaFirst.ID)
		}
		labels := newLabels.([]attribute.KeyValue)
		for _, l := range labels {
			for _, e := range expectedLables {
				if l.Key == e.Key {
					if l.Value.AsString() != e.Value.AsString() {
						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
					}
				}
			}
		}
	})
}
