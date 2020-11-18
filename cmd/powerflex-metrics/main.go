package main

// Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dell/karavi-metrics-powerflex/internal/entrypoint"
	"github.com/dell/karavi-metrics-powerflex/internal/k8s"
	"github.com/dell/karavi-metrics-powerflex/internal/service"
	otlexporters "github.com/dell/karavi-metrics-powerflex/opentelemetry/exporters"

	sio "github.com/dell/goscaleio"
	"go.opentelemetry.io/otel/api/global"

	"os"
)

const (
	defaultTickInterval = 5 * time.Second
)

func main() {

	powerFlexEndpoint := os.Getenv("POWERFLEX_ENDPOINT")
	if powerFlexEndpoint == "" {
		fmt.Printf("POWERFLEX_ENDPOINT is required")
		os.Exit(1)
	}

	powerFlexGatewayUser := os.Getenv("POWERFLEX_USER")
	if powerFlexGatewayUser == "" {
		fmt.Printf("POWERFLEX_USER is required")
		os.Exit(1)
	}

	powerFlexGatewayPassword := os.Getenv("POWERFLEX_PASSWORD")
	if powerFlexGatewayPassword == "" {
		fmt.Printf("POWERFLEX_PASSWORD is required")
		os.Exit(1)
	}

	collectorAddress := os.Getenv("COLLECTOR_ADDR")
	if collectorAddress == "" {
		fmt.Printf("COLLECTOR_ADDR is required")
		os.Exit(1)
	}

	provisionerNamesValue := os.Getenv("PROVISIONER_NAMES")
	if provisionerNamesValue == "" {
		fmt.Printf("PROVISIONER_NAMES is required")
		os.Exit(1)
	}

	powerflexSdcMetricsEnabled := true
	powerflexSdcMetricsEnabledValue := os.Getenv("POWERFLEX_SDC_METRICS_ENABLED")
	if powerflexSdcMetricsEnabledValue == "false" {
		powerflexSdcMetricsEnabled = false
	}

	powerflexVolumeMetricsEnabled := true
	powerflexVolumeMetricsEnabledValue := os.Getenv("POWERFLEX_VOLUME_METRICS_ENABLED")
	if powerflexVolumeMetricsEnabledValue == "false" {
		powerflexVolumeMetricsEnabled = false
	}

	storagePoolMetricsEnabled := true
	storagePoolMetricsEnabledValue := os.Getenv("POWERFLEX_STORAGE_POOL_METRICS_ENABLED")
	if storagePoolMetricsEnabledValue == "false" {
		storagePoolMetricsEnabled = false
	}

	provisionerNames := strings.Split(provisionerNamesValue, ",")

	sdcFinder := &k8s.SDCFinder{
		API:         &k8s.API{},
		DriverNames: provisionerNames,
	}

	storageClassFinder := &k8s.StorageClassFinder{
		API:         &k8s.API{},
		DriverNames: provisionerNames,
	}

	leaderElectorGetter := &k8s.LeaderElector{
		API: &k8s.LeaderElector{},
	}

	volumeFinder := &k8s.VolumeFinder{
		API:         &k8s.API{},
		DriverNames: provisionerNames,
	}

	nodeFinder := &k8s.NodeFinder{
		API: &k8s.API{},
	}

	client, err := sio.NewClientWithArgs(powerFlexEndpoint, "", true, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	sdcTickInterval := defaultTickInterval
	sdcIoPollFrequencySeconds := os.Getenv("POWERFLEX_SDC_IO_POLL_FREQUENCY")
	if sdcIoPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(sdcIoPollFrequencySeconds)
		if err != nil {
			fmt.Printf("POWERFLEX_SDC_IO_POLL_FREQUENCY was not set to a valid number")
			os.Exit(1)
		}
		sdcTickInterval = time.Duration(numSeconds) * time.Second
	}

	volumeTickInterval := defaultTickInterval
	volIoPollFrequencySeconds := os.Getenv("POWERFLEX_VOLUME_IO_POLL_FREQUENCY")
	if volIoPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(volIoPollFrequencySeconds)
		if err != nil {
			fmt.Printf("POWERFLEX_VOLUME_IO_POLL_FREQUENCY was not set to a valid number")
			os.Exit(1)
		}
		volumeTickInterval = time.Duration(numSeconds) * time.Second
	}

	storagePoolTickInterval := defaultTickInterval
	storagePoolPollFrequencySeconds := os.Getenv("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY")
	if storagePoolPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(storagePoolPollFrequencySeconds)
		if err != nil {
			fmt.Printf("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY was not set to a valid number")
			os.Exit(1)
		}
		storagePoolTickInterval = time.Duration(numSeconds) * time.Second
	}

	maxPowerFlexConcurrentRequests := service.DefaultMaxPowerFlexConnections
	maxPowerFlexConcurrentRequestsVar := os.Getenv("POWERFLEX_MAX_CONCURRENT_QUERIES")
	if maxPowerFlexConcurrentRequestsVar != "" {
		maxPowerFlexConcurrentRequests, err = strconv.Atoi(maxPowerFlexConcurrentRequestsVar)
		if err != nil {
			fmt.Printf("POWERFLEX_MAX_CONCURRENT_QUERIES was not set to a valid number: '%s'", maxPowerFlexConcurrentRequestsVar)
			os.Exit(1)
		}
		if maxPowerFlexConcurrentRequests <= 0 {
			fmt.Printf("POWERFLEX_MAX_CONCURRENT_QUERIES value was invalid (<= 0)")
			os.Exit(1)
		}
	}

	var collectorCertPath string
	if tls := os.Getenv("TLS_ENABLED"); tls == "true" {
		collectorCertPath = os.Getenv("COLLECTOR_CERT_PATH")
		if len(strings.TrimSpace(collectorCertPath)) < 1 {
			collectorCertPath = otlexporters.DefaultCollectorCertPath
		}
	}

	config := &entrypoint.Config{
		SDCTickInterval:           sdcTickInterval,
		VolumeTickInterval:        volumeTickInterval,
		StoragePoolTickInterval:   storagePoolTickInterval,
		PowerFlexClient:           client,
		PowerFlexConfig:           sio.ConfigConnect{Username: powerFlexGatewayUser, Password: powerFlexGatewayPassword},
		SDCFinder:                 sdcFinder,
		StorageClassFinder:        storageClassFinder,
		LeaderElector:             leaderElectorGetter,
		VolumeFinder:              volumeFinder,
		NodeFinder:                nodeFinder,
		SDCMetricsEnabled:         powerflexSdcMetricsEnabled,
		VolumeMetricsEnabled:      powerflexVolumeMetricsEnabled,
		StoragePoolMetricsEnabled: storagePoolMetricsEnabled,
		CollectorAddress:          collectorAddress,
		CollectorCertPath:         collectorCertPath,
	}

	exporter := &otlexporters.OtlCollectorExporter{CollectorAddr: collectorAddress}

	pflexSvc := &service.PowerFlexService{
		MetricsWrapper: &service.MetricsWrapper{
			Meter: global.Meter("powerflex/sdc"),
		},
		MaxPowerFlexConnections: maxPowerFlexConcurrentRequests,
	}

	if err := entrypoint.Run(context.Background(), config, exporter, pflexSvc); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
