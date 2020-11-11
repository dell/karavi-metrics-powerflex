package service

// Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/metric"
)

// MetricsRecorder supports recording I/O metrics
//go:generate mockgen -destination=mocks/metrics_mocks.go -package=mocks github.com/dell/karavi-powerflex-metrics/internal/service MetricsRecorder,Float64UpDownCounterCreater
type MetricsRecorder interface {
	Record(ctx context.Context, meta interface{},
		readBW, writeBW,
		readIOPS, writeIOPS,
		readLatency, writeLatency float64) error
	RecordCapacity(ctx context.Context, meta interface{},
		totalLogicalCapacity, logicalCapacityAvailable, logicalCapacityInUse, logicalProvisioned float64) error
}

// Float64UpDownCounterCreater creates a Float64UpDownCounter metric
type Float64UpDownCounterCreater interface {
	NewFloat64UpDownCounter(name string, options ...metric.InstrumentOption) (metric.Float64UpDownCounter, error)
}

// MetricsWrapper contains data used for pushing metrics data
type MetricsWrapper struct {
	Meter           Float64UpDownCounterCreater
	Metrics         sync.Map
	Labels          sync.Map
	CapacityMetrics sync.Map
}

// Metrics contains the list of metrics data that is collected
type Metrics struct {
	ReadBW       metric.BoundFloat64UpDownCounter
	WriteBW      metric.BoundFloat64UpDownCounter
	ReadIOPS     metric.BoundFloat64UpDownCounter
	WriteIOPS    metric.BoundFloat64UpDownCounter
	ReadLatency  metric.BoundFloat64UpDownCounter
	WriteLatency metric.BoundFloat64UpDownCounter
}

// CapacityMetrics contains the metrics related to a capacity
type CapacityMetrics struct {
	TotalLogicalCapacity     metric.BoundFloat64UpDownCounter
	LogicalCapacityAvailable metric.BoundFloat64UpDownCounter
	LogicalCapacityInUse     metric.BoundFloat64UpDownCounter
	LogicalProvisioned       metric.BoundFloat64UpDownCounter
}

func (mw *MetricsWrapper) initMetrics(prefix, metaID string, labels []kv.KeyValue) (*Metrics, error) {
	unboundReadBW, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_bw")
	if err != nil {
		return nil, err
	}
	readBW := unboundReadBW.Bind(labels...)

	unboundWriteBW, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_bw")
	if err != nil {
		return nil, err
	}
	writeBW := unboundWriteBW.Bind(labels...)

	unboundReadIOPS, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_iops")
	if err != nil {
		return nil, err
	}
	readIOPS := unboundReadIOPS.Bind(labels...)

	unboundWriteIOPS, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_iops")
	if err != nil {
		return nil, err
	}
	writeIOPS := unboundWriteIOPS.Bind(labels...)

	unboundReadLatency, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_latency")
	if err != nil {
		return nil, err
	}
	readLatency := unboundReadLatency.Bind(labels...)

	unboundWriteLatency, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_latency")
	if err != nil {
		return nil, err
	}
	writeLatency := unboundWriteLatency.Bind(labels...)

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

func (mw *MetricsWrapper) initCapacityMetrics(prefix, metaID string, labels []kv.KeyValue) (*CapacityMetrics, error) {
	unboundTotalLogicalCapacity, err := mw.Meter.NewFloat64UpDownCounter(prefix + "total_logical_capacity")
	if err != nil {
		return nil, err
	}
	totalLogicalCapacity := unboundTotalLogicalCapacity.Bind(labels...)

	unboundLogicalCapacityAvailable, err := mw.Meter.NewFloat64UpDownCounter(prefix + "logical_capacity_available")
	if err != nil {
		return nil, err
	}
	logicalCapacityAvailable := unboundLogicalCapacityAvailable.Bind(labels...)

	unboundLogicalCapacityInUse, err := mw.Meter.NewFloat64UpDownCounter(prefix + "logical_capacity_in_use")
	if err != nil {
		return nil, err
	}
	logicalCapacityInUse := unboundLogicalCapacityInUse.Bind(labels...)

	unboundLogicalProvisioned, err := mw.Meter.NewFloat64UpDownCounter(prefix + "logical_provisioned")
	if err != nil {
		return nil, err
	}
	logicalProvisioned := unboundLogicalProvisioned.Bind(labels...)

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
func (mw *MetricsWrapper) Record(ctx context.Context, meta interface{},
	readBW, writeBW,
	readIOPS, writeIOPS,
	readLatency, writeLatency float64,
) error {

	var prefix string
	var metaID string
	var labels []kv.KeyValue
	switch v := meta.(type) {
	case *VolumeMeta:
		prefix, metaID = "powerflex_volume_", v.ID
		mappedSDCIDs := "__"
		mappedSDCIPs := "__"
		for _, ip := range v.MappedSDCs {
			mappedSDCIDs += (ip.SdcID + "__")
			mappedSDCIPs += (ip.SdcIP + "__")
		}
		labels = []kv.KeyValue{
			kv.String("volume_id", v.ID),
			kv.String("volume_name", v.Name),
			kv.String("persistent_volume_name", v.PersistentVolumeName),
			kv.String("mapped_node_ids", mappedSDCIDs),
			kv.String("mapped_node_ips", mappedSDCIPs),
			kv.String("plot_with_mean", "No"),
		}
	case *SDCMeta:
		prefix, metaID = "powerflex_export_node_", v.ID
		labels = []kv.KeyValue{
			kv.String("id", v.ID),
			kv.String("name", v.Name),
			kv.String("ip", v.IP),
			kv.String("node_guid", v.SdcGUID),
			kv.String("plot_with_mean", "No"),
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
		// If Metrics for this MetricsWrapper exist, then check if any labels have changed and update them
		currentLabels, ok := mw.Labels.Load(metaID)
		if !ok {
			newMetrics, err := mw.initMetrics(prefix, metaID, labels)
			if err != nil {
				return err
			}
			metricsMapValue = newMetrics
		} else {
			currentLabels := currentLabels.([]kv.KeyValue)
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

	metrics.ReadBW.Add(ctx, readBW)
	metrics.WriteBW.Add(ctx, writeBW)
	metrics.ReadIOPS.Add(ctx, readIOPS)
	metrics.WriteIOPS.Add(ctx, writeIOPS)
	metrics.ReadLatency.Add(ctx, readLatency)
	metrics.WriteLatency.Add(ctx, writeLatency)

	return nil
}

// RecordCapacity will publish capacity metrics for a given instance
func (mw *MetricsWrapper) RecordCapacity(ctx context.Context, meta interface{},
	totalLogicalCapacity, logicalCapacityAvailable, logicalCapacityInUse, logicalProvisioned float64) error {

	switch v := meta.(type) {
	case StorageClassMeta:
		switch v.Driver {
		case "csi-vxflexos.dellemc.com":
			prefix, metaID := "storage_pool_", v.ID
			for pool := range v.StoragePools {
				labels := []kv.KeyValue{
					kv.String("storage_class", v.Name),
					kv.String("driver", v.Driver),
					kv.String("storage_pool", pool),
					kv.String("storage_system_name", v.StorageSystemName),
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

				metrics.TotalLogicalCapacity.Add(ctx, totalLogicalCapacity)
				metrics.LogicalCapacityAvailable.Add(ctx, logicalCapacityAvailable)
				metrics.LogicalCapacityInUse.Add(ctx, logicalCapacityInUse)
				metrics.LogicalProvisioned.Add(ctx, logicalProvisioned)
			}
		}
	default:
		return errors.New("unknown MetaData type")
	}
	return nil
}
