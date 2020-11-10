package otlexporters

// Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

import (
	"time"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

// OtlCollectorExporter is the exporter for the OpenTelemetry Collector
type OtlCollectorExporter struct {
	CollectorAddr string
	exporter      *otlp.Exporter
	pusher        *push.Controller
}

// InitExporter is the initialization method for the OpenTelemetry Collector exporter
func (c *OtlCollectorExporter) InitExporter(opts ...otlp.ExporterOption) error {
	exporter, pusher, err := c.initOTLPExporter()
	if err != nil {
		return err
	}
	c.exporter = exporter
	c.pusher = pusher

	return err
}

// StopExporter stops the activity of the Otl Collector's required services
func (c *OtlCollectorExporter) StopExporter() error {
	err := c.exporter.Stop()
	if err != nil {
		return err
	}
	c.pusher.Stop()
	return nil
}

func (c *OtlCollectorExporter) initOTLPExporter(opts ...otlp.ExporterOption) (*otlp.Exporter, *push.Controller, error) {
	exporter, err := otlp.NewExporter(opts...)
	if err != nil {
		return nil, nil, err
	}

	pusher := push.New(
		simple.NewWithExactDistribution(),
		exporter,
		push.WithPeriod(5*time.Second),
	)

	pusher.Start()

	global.SetMeterProvider(pusher.Provider())

	return exporter, pusher, nil
}
