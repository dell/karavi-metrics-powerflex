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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricsRecorder supports recording I/O metrics
//go:generate mockgen -destination=mocks/metrics_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service MetricsRecorder,Float64UpDownCounterCreater
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
	ReadBW       metric.Float64UpDownCounter
	WriteBW      metric.Float64UpDownCounter
	ReadIOPS     metric.Float64UpDownCounter
	WriteIOPS    metric.Float64UpDownCounter
	ReadLatency  metric.Float64UpDownCounter
	WriteLatency metric.Float64UpDownCounter
}

// CapacityMetrics contains the metrics related to a capacity
type CapacityMetrics struct {
	TotalLogicalCapacity     metric.Float64UpDownCounter
	LogicalCapacityAvailable metric.Float64UpDownCounter
	LogicalCapacityInUse     metric.Float64UpDownCounter
	LogicalProvisioned       metric.Float64UpDownCounter
}

func (mw *MetricsWrapper) initMetrics(prefix, metaID string, labels []attribute.KeyValue) (*Metrics, error) {
	readBW, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}

	writeBW, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}

	readIOPS, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_iops_per_second")
	if err != nil {
		return nil, err
	}

	writeIOPS, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_iops_per_second")
	if err != nil {
		return nil, err
	}

	readLatency, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_latency_milliseconds")
	if err != nil {
		return nil, err
	}

	writeLatency, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_latency_milliseconds")
	if err != nil {
		return nil, err
	}

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

func (mw *MetricsWrapper) initCapacityMetrics(prefix, metaID string, labels []attribute.KeyValue) (*CapacityMetrics, error) {
	totalLogicalCapacity, err := mw.Meter.NewFloat64UpDownCounter(prefix + "total_logical_capacity_gigabytes")
	if err != nil {
		return nil, err
	}

	logicalCapacityAvailable, err := mw.Meter.NewFloat64UpDownCounter(prefix + "logical_capacity_available_gigabytes")
	if err != nil {
		return nil, err
	}

	logicalCapacityInUse, err := mw.Meter.NewFloat64UpDownCounter(prefix + "logical_capacity_in_use_gigabytes")
	if err != nil {
		return nil, err
	}

	logicalProvisioned, err := mw.Meter.NewFloat64UpDownCounter(prefix + "logical_provisioned_gigabytes")
	if err != nil {
		return nil, err
	}

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
		// If Metrics for this MetricsWrapper exist, then check if any labels have changed and update them
		currentLabels, ok := mw.Labels.Load(metaID)
		if !ok {
			newMetrics, err := mw.initMetrics(prefix, metaID, labels)
			if err != nil {
				return err
			}
			metricsMapValue = newMetrics
		} else {
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
