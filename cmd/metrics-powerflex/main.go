// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package main

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
	"github.com/sirupsen/logrus"

	sio "github.com/dell/goscaleio"
	"go.opentelemetry.io/otel/api/global"

	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	defaultTickInterval            = 5 * time.Second
	defaultConfigFile              = "/etc/config/karavi-metrics-powerflex.yaml"
	defaultStorageSystemConfigFile = "/vxflexos-config/config"
)

func main() {

	logger := logrus.New()

	viper.SetConfigFile(defaultConfigFile)

	err := viper.ReadInConfig()
	// if unable to read configuration file, proceed in case we use environment variables
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read Config file: %v", err)
	}

	configFileListener := viper.New()
	configFileListener.SetConfigFile(defaultStorageSystemConfigFile)

	sdcFinder := &k8s.SDCFinder{
		API: &k8s.API{},
	}

	storageClassFinder := &k8s.StorageClassFinder{
		API: &k8s.API{},
	}

	leaderElectorGetter := &k8s.LeaderElector{
		API: &k8s.LeaderElector{},
	}

	volumeFinder := &k8s.VolumeFinder{
		API: &k8s.API{},
	}

	nodeFinder := &k8s.NodeFinder{
		API: &k8s.API{},
	}

	updateLoggingSettings := func(logger *logrus.Logger) {
		logFormat := viper.GetString("LOG_FORMAT")
		if strings.EqualFold(logFormat, "json") {
			logger.SetFormatter(&logrus.JSONFormatter{})
		} else {
			// use text formatter by default
			logger.SetFormatter(&logrus.TextFormatter{})
		}
		logLevel := viper.GetString("LOG_LEVEL")
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			// use INFO level by default
			level = logrus.InfoLevel
		}
		logger.SetLevel(level)
	}

	updateLoggingSettings(logger)
	updateProvisionerNames(sdcFinder, storageClassFinder, volumeFinder, logger)

	var collectorCertPath string
	if tls := os.Getenv("TLS_ENABLED"); tls == "true" {
		collectorCertPath = os.Getenv("COLLECTOR_CERT_PATH")
		if len(strings.TrimSpace(collectorCertPath)) < 1 {
			collectorCertPath = otlexporters.DefaultCollectorCertPath
		}
	}

	config := &entrypoint.Config{
		SDCFinder:          sdcFinder,
		StorageClassFinder: storageClassFinder,
		LeaderElector:      leaderElectorGetter,
		VolumeFinder:       volumeFinder,
		NodeFinder:         nodeFinder,
		CollectorCertPath:  collectorCertPath,
		Logger:             logger,
	}

	exporter := &otlexporters.OtlCollectorExporter{}

	pflexSvc := &service.PowerFlexService{
		MetricsWrapper: &service.MetricsWrapper{
			Meter: global.Meter("powerflex/sdc"),
		},
		Logger: logger,
	}

	updatePowerFlexConnection(config, sdcFinder, storageClassFinder, volumeFinder, logger)
	updateCollectorAddress(config, exporter, logger)
	updateMetricsEnabled(config)
	updateTickIntervals(config, logger)
	updateService(pflexSvc, logger)

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		updateLoggingSettings(logger)
		updateCollectorAddress(config, exporter, logger)
		updateProvisionerNames(sdcFinder, storageClassFinder, volumeFinder, logger)
		updateMetricsEnabled(config)
		updateTickIntervals(config, logger)
		updateService(pflexSvc, logger)
	})

	configFileListener.WatchConfig()
	configFileListener.OnConfigChange(func(e fsnotify.Event) {
		updatePowerFlexConnection(config, sdcFinder, storageClassFinder, volumeFinder, logger)
	})

	if err := entrypoint.Run(context.Background(), config, exporter, pflexSvc); err != nil {
		logger.WithError(err).Fatal("running service")
	}
}

func updatePowerFlexConnection(config *entrypoint.Config, sdcFinder *k8s.SDCFinder, storageClassFinder *k8s.StorageClassFinder, volumeFinder *k8s.VolumeFinder, logger *logrus.Logger) {
	configReader := service.ConfigurationReader{}

	storageSystem, err := configReader.GetStorageSystemConfiguration(defaultStorageSystemConfigFile)
	if err != nil {
		logger.WithError(err).Fatal("getting storage system configuration")
	}

	powerFlexEndpoint := storageSystem.Endpoint
	if powerFlexEndpoint == "" {
		logger.WithError(err).Fatal("powerflex endpoint was empty")
	}

	powerFlexGatewayUser := storageSystem.Username
	if powerFlexGatewayUser == "" {
		logger.WithError(err).Fatal("powerflex username was empty")
	}

	powerFlexGatewayPassword := storageSystem.Password
	if powerFlexGatewayPassword == "" {
		logger.WithError(err).Fatal("powerflex password was empty")
	}

	powerFlexSystemID := storageSystem.SystemID
	if powerFlexSystemID == "" {
		logger.WithError(err).Fatal("powerflex system ID was empty")
	}

	sdcFinder.StorageSystemID = powerFlexSystemID
	storageClassFinder.StorageSystemID = powerFlexSystemID
	storageClassFinder.IsDefaultStorageSystem = storageSystem.IsDefault
	volumeFinder.StorageSystemID = powerFlexSystemID

	client, err := sio.NewClientWithArgs(powerFlexEndpoint, "", storageSystem.Insecure, true)
	if err != nil {
		logger.WithError(err).Fatal("creating powerflex client")
	}
	config.PowerFlexClient = client

	config.PowerFlexConfig = sio.ConfigConnect{Username: powerFlexGatewayUser, Password: powerFlexGatewayPassword}

	logger.WithField("storage_system_id", powerFlexSystemID).Info("set powerflex system ID")
}

func updateCollectorAddress(config *entrypoint.Config, exporter *otlexporters.OtlCollectorExporter, logger *logrus.Logger) {
	collectorAddress := viper.GetString("COLLECTOR_ADDR")
	if collectorAddress == "" {
		logger.Fatal("COLLECTOR_ADDR is required")
	}
	config.CollectorAddress = collectorAddress
	exporter.CollectorAddr = collectorAddress
}

func updateProvisionerNames(sdcFinder *k8s.SDCFinder, storageClassFinder *k8s.StorageClassFinder, volumeFinder *k8s.VolumeFinder, logger *logrus.Logger) {
	provisionerNamesValue := viper.GetString("provisioner_names")
	if provisionerNamesValue == "" {
		logger.Fatal("PROVISIONER_NAMES is required")
	}
	provisionerNames := strings.Split(provisionerNamesValue, ",")
	sdcFinder.DriverNames = provisionerNames
	storageClassFinder.DriverNames = provisionerNames
	volumeFinder.DriverNames = provisionerNames
}

func updateMetricsEnabled(config *entrypoint.Config) {
	powerflexSdcMetricsEnabled := true
	powerflexSdcMetricsEnabledValue := viper.GetString("POWERFLEX_SDC_METRICS_ENABLED")
	if powerflexSdcMetricsEnabledValue == "false" {
		powerflexSdcMetricsEnabled = false
	}

	powerflexVolumeMetricsEnabled := true
	powerflexVolumeMetricsEnabledValue := viper.GetString("POWERFLEX_VOLUME_METRICS_ENABLED")
	if powerflexVolumeMetricsEnabledValue == "false" {
		powerflexVolumeMetricsEnabled = false
	}

	storagePoolMetricsEnabled := true
	storagePoolMetricsEnabledValue := viper.GetString("POWERFLEX_STORAGE_POOL_METRICS_ENABLED")
	if storagePoolMetricsEnabledValue == "false" {
		storagePoolMetricsEnabled = false
	}
	config.SDCMetricsEnabled = powerflexSdcMetricsEnabled
	config.VolumeMetricsEnabled = powerflexVolumeMetricsEnabled
	config.StoragePoolMetricsEnabled = storagePoolMetricsEnabled
}

func updateTickIntervals(config *entrypoint.Config, logger *logrus.Logger) {
	sdcTickInterval := defaultTickInterval
	sdcIoPollFrequencySeconds := viper.GetString("POWERFLEX_SDC_IO_POLL_FREQUENCY")
	if sdcIoPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(sdcIoPollFrequencySeconds)
		if err != nil {
			logger.WithError(err).Fatal("POWERFLEX_SDC_IO_POLL_FREQUENCY was not set to a valid number")
		}
		sdcTickInterval = time.Duration(numSeconds) * time.Second
	}

	volumeTickInterval := defaultTickInterval
	volIoPollFrequencySeconds := viper.GetString("POWERFLEX_VOLUME_IO_POLL_FREQUENCY")
	if volIoPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(volIoPollFrequencySeconds)
		if err != nil {
			logger.WithError(err).Fatal("POWERFLEX_VOLUME_IO_POLL_FREQUENCY was not set to a valid number")
		}
		volumeTickInterval = time.Duration(numSeconds) * time.Second
	}

	storagePoolTickInterval := defaultTickInterval
	storagePoolPollFrequencySeconds := viper.GetString("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY")
	if storagePoolPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(storagePoolPollFrequencySeconds)
		if err != nil {
			logger.WithError(err).Fatal("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY was not set to a valid number")
		}
		storagePoolTickInterval = time.Duration(numSeconds) * time.Second
	}

	config.SDCTickInterval = sdcTickInterval
	config.VolumeTickInterval = volumeTickInterval
	config.StoragePoolTickInterval = storagePoolTickInterval
}

func updateService(pflexSvc *service.PowerFlexService, logger *logrus.Logger) {
	maxPowerFlexConcurrentRequests := service.DefaultMaxPowerFlexConnections
	maxPowerFlexConcurrentRequestsVar := viper.GetString("POWERFLEX_MAX_CONCURRENT_QUERIES")
	if maxPowerFlexConcurrentRequestsVar != "" {
		maxPowerFlexConcurrentRequests, err := strconv.Atoi(maxPowerFlexConcurrentRequestsVar)
		if err != nil {
			logger.WithError(err).Fatal("POWERFLEX_MAX_CONCURRENT_QUERIES was not set to a valid number")
		}
		if maxPowerFlexConcurrentRequests <= 0 {
			logger.WithError(err).Fatal("POWERFLEX_MAX_CONCURRENT_QUERIES value was invalid (<= 0)")
		}
	}
	pflexSvc.MaxPowerFlexConnections = maxPowerFlexConcurrentRequests
}
