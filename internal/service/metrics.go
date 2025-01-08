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

package service

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricsRecorder supports recording I/O metrics
//
//go:generate mockgen -destination=mocks/metrics_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service MetricsRecorder,MeterCreater
type MetricsRecorder interface {
	Record(ctx context.Context, meta interface{},
		readBW, writeBW,
		readIOPS, writeIOPS,
		readLatency, writeLatency float64) error
	RecordCapacity(ctx context.Context, meta interface{},
		totalLogicalCapacity, logicalCapacityAvailable, logicalCapacityInUse, logicalProvisioned float64) error
}

// MeterCreater interface is used to create and provide Meter instances, which are used to report measurements.
//
//go:generate mockgen -destination=mocks/meter_mocks.go -package=mocks go.opentelemetry.io/otel/metric Meter
type MeterCreater interface {
	// AsyncFloat64() asyncfloat64.InstrumentProvider
	MeterProvider() metric.Meter
	// metric.Float64ObservableUpDownCounter
}

// MetricsWrapper contains data used for pushing metrics data
type MetricsWrapper struct {
	Meter           metric.Meter
	Metrics         sync.Map
	Labels          sync.Map
	CapacityMetrics sync.Map
}

// Metrics contains the list of metrics data that is collected
type Metrics struct {
	ReadBW       metric.Float64ObservableUpDownCounter
	WriteBW      metric.Float64ObservableUpDownCounter
	ReadIOPS     metric.Float64ObservableUpDownCounter
	WriteIOPS    metric.Float64ObservableUpDownCounter
	ReadLatency  metric.Float64ObservableUpDownCounter
	WriteLatency metric.Float64ObservableUpDownCounter
}

// CapacityMetrics contains the metrics related to a capacity
type CapacityMetrics struct {
	TotalLogicalCapacity     metric.Float64ObservableUpDownCounter
	LogicalCapacityAvailable metric.Float64ObservableUpDownCounter
	LogicalCapacityInUse     metric.Float64ObservableUpDownCounter
	LogicalProvisioned       metric.Float64ObservableUpDownCounter
}

func (mw *MetricsWrapper) initMetrics(prefix, metaID string, labels []attribute.KeyValue) (*Metrics, error) {
	readBW, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")

	writeBW, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")

	readIOPS, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")

	writeIOPS, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "write_iops_per_second")

	readLatency, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "read_latency_milliseconds")

	writeLatency, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "write_latency_milliseconds")

	metrics := &Metrics{
		ReadBW:       readBW,
		WriteBW:      writeBW,
		ReadIOPS:     readIOPS,
		WriteIOPS:    writeIOPS,
		ReadLatency:  readLatency,
		WriteLatency: writeLatency,
	}

	mw.Metrics.Store(metaID, metrics)
	mw.Labels.Store(metaID, labels)

	return metrics, nil
}

func (mw *MetricsWrapper) initCapacityMetrics(prefix, metaID string, _ []attribute.KeyValue) (*CapacityMetrics, error) {
	totalLogicalCapacity, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "total_logical_capacity_gigabytes")

	logicalCapacityAvailable, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "logical_capacity_available_gigabytes")

	logicalCapacityInUse, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "logical_capacity_in_use_gigabytes")

	logicalProvisioned, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "logical_used_gigabytes")

	metrics := &CapacityMetrics{
		TotalLogicalCapacity:     totalLogicalCapacity,
		LogicalCapacityAvailable: logicalCapacityAvailable,
		LogicalCapacityInUse:     logicalCapacityInUse,
		LogicalProvisioned:       logicalProvisioned,
	}

	mw.CapacityMetrics.Store(metaID, metrics)

	return metrics, nil
}

// Record will publish metrics data for a given instance
func (mw *MetricsWrapper) Record(_ context.Context, meta interface{},
	readBW, writeBW,
	readIOPS, writeIOPS,
	readLatency, writeLatency float64,
) error {
	var prefix string
	var metaID string
	var labels []attribute.KeyValue
	switch v := meta.(type) {
	case *VolumeMeta:
		prefix, metaID = "powerflex_volume_", v.ID
		mappedSDCIDs := "__"
		mappedSDCIPs := "__"
		for _, ip := range v.MappedSDCs {
			mappedSDCIDs += (ip.SdcID + "__")
			mappedSDCIPs += (ip.SdcIP + "__")
		}
		labels = []attribute.KeyValue{
			attribute.String("VolumeID", v.ID),
			attribute.String("VolumeName", v.Name),
			attribute.String("StorageSystemID", v.StorageSystemID),
			attribute.String("PersistentVolumeName", v.PersistentVolumeName),
			attribute.String("PersistentVolumeClaimName", v.PersistentVolumeClaimName),
			attribute.String("Namespace", v.Namespace),
			attribute.String("MappedNodeIDs", mappedSDCIDs),
			attribute.String("MappedNodeIPs", mappedSDCIPs),
			attribute.String("PlotWithMean", "No"),
		}
	case *SDCMeta:
		prefix, metaID = "powerflex_export_node_", v.ID
		labels = []attribute.KeyValue{
			attribute.String("ID", v.ID),
			attribute.String("Name", v.Name),
			attribute.String("IP", v.IP),
			attribute.String("NodeGUID", v.SdcGUID),
			attribute.String("PlotWithMean", "No"),
		}
	default:
		return errors.New("unknown MetaData type")
	}

	metricsMapValue, ok := mw.Metrics.Load(metaID)
	if !ok {
		newMetrics, err := mw.initMetrics(prefix, metaID, labels)
		if err != nil {
			return err
		}
		metricsMapValue = newMetrics
	} else {
		// If Metrics for this MetricsWrapper exist, then update the labels
		currentLabels, ok := mw.Labels.Load(metaID)
		if ok {
			currentLabels := currentLabels.([]attribute.KeyValue)
			updatedLabels := currentLabels
			haveLabelsChanged := false
			for i, current := range currentLabels {
				for _, new := range labels {
					if current.Key == new.Key {
						if current.Value != new.Value {
							updatedLabels[i].Value = new.Value
							haveLabelsChanged = true
						}
					}
				}
			}
			if haveLabelsChanged {
				newMetrics, err := mw.initMetrics(prefix, metaID, updatedLabels)
				if err != nil {
					return err
				}
				metricsMapValue = newMetrics
			}
		}
	}

	metrics := metricsMapValue.(*Metrics)

	//_, _ = mw.Meter.RegisterCallback(func(_ context.Context, obs metric.Observer) error {
	done := make(chan struct{})

	reg, err := mw.Meter.RegisterCallback(func(_ context.Context, obs metric.Observer) error {
		obs.ObserveFloat64(metrics.ReadBW, float64(readBW), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.WriteBW, float64(writeBW), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.ReadIOPS, float64(readIOPS), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.WriteIOPS, float64(writeIOPS), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.ReadLatency, float64(readLatency), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.WriteLatency, float64(writeLatency), metric.ObserveOption(metric.WithAttributes(labels...)))
		go func() {
			done <- struct{}{}
		}()
		return nil
	},
		metrics.ReadBW,
		metrics.WriteBW,
		metrics.ReadIOPS,
		metrics.WriteIOPS,
		metrics.ReadLatency,
		metrics.WriteLatency,
	)
	if err != nil {
		return err
	}
	<-done
	_ = reg.Unregister()

	return nil
}

// RecordCapacity will publish capacity metrics for a given instance
func (mw *MetricsWrapper) RecordCapacity(_ context.Context, meta interface{},
	totalLogicalCapacity, logicalCapacityAvailable, logicalCapacityInUse, logicalProvisioned float64,
) error {
	switch v := meta.(type) {
	case StorageClassMeta:
		switch v.Driver {
		case "csi-vxflexos.dellemc.com":
			prefix, metaID := "powerflex_storage_pool_", v.ID
			for pool := range v.StoragePools {
				labels := []attribute.KeyValue{
					attribute.String("StorageClass", v.Name),
					attribute.String("Driver", v.Driver),
					attribute.String("StoragePool", pool),
					attribute.String("StorageSystemID", v.StorageSystemID),
				}

				metricsMapValue, ok := mw.CapacityMetrics.Load(metaID)
				if !ok {
					newMetrics, err := mw.initCapacityMetrics(prefix, metaID, labels)
					if err != nil {
						return err
					}
					metricsMapValue = newMetrics
				}

				metrics := metricsMapValue.(*CapacityMetrics)
				done := make(chan struct{})
				reg, err := mw.Meter.RegisterCallback(func(_ context.Context, obs metric.Observer) error {
					obs.ObserveFloat64(metrics.TotalLogicalCapacity, float64(totalLogicalCapacity), metric.ObserveOption(metric.WithAttributes(labels...)))
					obs.ObserveFloat64(metrics.LogicalCapacityAvailable, float64(logicalCapacityAvailable), metric.ObserveOption(metric.WithAttributes(labels...)))
					obs.ObserveFloat64(metrics.LogicalCapacityInUse, float64(logicalCapacityInUse), metric.ObserveOption(metric.WithAttributes(labels...)))
					obs.ObserveFloat64(metrics.LogicalProvisioned, float64(logicalProvisioned), metric.ObserveOption(metric.WithAttributes(labels...)))
					go func() {
						done <- struct{}{}
					}()
					return nil
				},
					metrics.TotalLogicalCapacity,
					metrics.LogicalCapacityAvailable,
					metrics.LogicalCapacityInUse,
					metrics.LogicalProvisioned,
				)
				if err != nil {
					return err
				}
				<-done
				_ = reg.Unregister()
			}
		}
	default:
		return errors.New("unknown MetaData type")
	}
	return nil
}
