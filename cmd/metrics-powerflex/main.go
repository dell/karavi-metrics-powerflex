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

package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dell/goscaleio"
	"github.com/dell/karavi-metrics-powerflex/internal/entrypoint"
	"github.com/dell/karavi-metrics-powerflex/internal/k8s"
	"github.com/dell/karavi-metrics-powerflex/internal/service"
	otlexporters "github.com/dell/karavi-metrics-powerflex/opentelemetry/exporters"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

const (
	defaultTickInterval            = 5 * time.Second
	defaultConfigFile              = "/etc/config/karavi-metrics-powerflex.yaml"
	defaultStorageSystemConfigFile = "/vxflexos-config/config"
)

var logger *logrus.Logger

func main() {
	config, exporter, powerflexSvc := configure()
	if err := entrypoint.Run(context.Background(), config, exporter, powerflexSvc); err != nil {
		logger.WithError(err).Fatal("running service")
	}
}

func configure() (*entrypoint.Config, otlexporters.Otlexporter, *service.PowerFlexService) {
	logger := setupLogger()

	loadConfig(logger)

	configFileListener := setupConfigFileListener()

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
		API:    &k8s.API{},
		Logger: logger,
	}
	nodeFinder := &k8s.NodeFinder{
		API: &k8s.API{},
	}

	config := setupConfig(sdcFinder, storageClassFinder, leaderElectorGetter, volumeFinder, nodeFinder, logger)

	exporter := &otlexporters.OtlCollectorExporter{}

	powerflexSvc := setupPowerFlexService(logger)

	onChangeUpdate(powerflexSvc, config, sdcFinder, exporter, storageClassFinder, volumeFinder, logger)

	setupConfigWatchers(configFileListener, powerflexSvc, config, sdcFinder, storageClassFinder, volumeFinder, exporter, logger)

	return config, exporter, powerflexSvc
}

func setupLogger() *logrus.Logger {
	logger := logrus.New()
	updateLoggingSettings(logger)
	return logger
}

// loadConfig loads the primary configuration file.
func loadConfig(_ *logrus.Logger) {
	viper.SetConfigFile(defaultConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "unable to read Config file: %v", err)
	}
}

// setupConfigFileListener initializes a secondary config watcher for storage system configs.
func setupConfigFileListener() *viper.Viper {
	configFileListener := viper.New()
	configFileListener.SetConfigFile(defaultStorageSystemConfigFile)
	return configFileListener
}

// getCollectorCertPath retrieves the certificate path for the OpenTelemetry collector.
func getCollectorCertPath() string {
	var collectorCertPath string
	if tls := os.Getenv("TLS_ENABLED"); tls == "true" {
		collectorCertPath = os.Getenv("COLLECTOR_CERT_PATH")
		if len(strings.TrimSpace(collectorCertPath)) < 1 {
			collectorCertPath = otlexporters.DefaultCollectorCertPath
		}
	}
	return collectorCertPath
}

// setupConfig creates the main configuration structure.
func setupConfig(
	sdcFinder *k8s.SDCFinder,
	storageClassFinder *k8s.StorageClassFinder,
	leaderElectorGetter *k8s.LeaderElector,
	volumeFinder *k8s.VolumeFinder,
	nodeFinder *k8s.NodeFinder,
	logger *logrus.Logger,
) *entrypoint.Config {
	return &entrypoint.Config{
		SDCFinder:          sdcFinder,
		StorageClassFinder: storageClassFinder,
		LeaderElector:      leaderElectorGetter,
		VolumeFinder:       volumeFinder,
		NodeFinder:         nodeFinder,
		CollectorCertPath:  getCollectorCertPath(),
		Logger:             logger,
	}
}

func setupPowerFlexService(logger *logrus.Logger) *service.PowerFlexService {
	return &service.PowerFlexService{
		MetricsWrapper: &service.MetricsWrapper{
			Meter: otel.Meter("powerflex/sdc"),
		},
		Logger: logger,
	}
}

func onChangeUpdate(
	powerflexSvc *service.PowerFlexService,
	config *entrypoint.Config,
	sdcFinder *k8s.SDCFinder,
	exporter *otlexporters.OtlCollectorExporter,
	storageClassFinder *k8s.StorageClassFinder,
	volumeFinder *k8s.VolumeFinder,
	logger *logrus.Logger,
) {
	updateCollectorAddress(config, exporter, logger)
	updateProvisionerNames(sdcFinder, storageClassFinder, volumeFinder, logger)
	updateMetricsEnabled(config)
	updateTickIntervals(config, logger)
	updateService(powerflexSvc, logger)
	updatePowerFlexConnection(defaultStorageSystemConfigFile, config, sdcFinder, storageClassFinder, volumeFinder, logger)
}

func updateLoggingSettings(logger *logrus.Logger) {
	logFormat := viper.GetString("LOG_FORMAT")
	if strings.EqualFold(logFormat, "json") {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	logLevel := viper.GetString("LOG_LEVEL")
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
}

// setupConfigWatchers sets up dynamic updates when config files change.
func setupConfigWatchers(configFileListener *viper.Viper, powerflexSvc *service.PowerFlexService, config *entrypoint.Config, sdcFinder *k8s.SDCFinder, storageClassFinder *k8s.StorageClassFinder, volumeFinder *k8s.VolumeFinder, exporter *otlexporters.OtlCollectorExporter, logger *logrus.Logger) {
	viper.WatchConfig()
	viper.OnConfigChange(func(_ fsnotify.Event) {
		updateLoggingSettings(logger)
	})

	configFileListener.WatchConfig()
	configFileListener.OnConfigChange(func(_ fsnotify.Event) {
		onChangeUpdate(powerflexSvc, config, sdcFinder, exporter, storageClassFinder, volumeFinder, logger)
	})
}

func GetStorageSystemArray(storageSystemConfigFile string) ([]service.ArrayConnectionData, error) {
	configReader := service.ConfigurationReader{}
	storageSystemArray, err := configReader.GetStorageSystemConfiguration(storageSystemConfigFile)
	if err != nil {
		logger.WithError(err).Fatal("getting storage system configuration")
	}
	return storageSystemArray, err
}

func updatePowerFlexConnection(
	storageSystemConfigFile string,
	config *entrypoint.Config,
	sdcFinder *k8s.SDCFinder,
	storageClassFinder *k8s.StorageClassFinder,
	volumeFinder *k8s.VolumeFinder,
	logger *logrus.Logger,
) {
	storageSystemArray, err := GetStorageSystemArray(storageSystemConfigFile)

	volumeFinder.StorageSystemID = make([]k8s.StorageSystemID, len(storageSystemArray))
	sdcFinder.StorageSystemID = make([]k8s.StorageSystemID, len(storageSystemArray))
	storageClassFinder.StorageSystemID = make([]k8s.StorageSystemID, len(storageSystemArray))

	config.PowerFlexClient = make(map[string]service.PowerFlexClient)
	config.PowerFlexConfig = make(map[string]goscaleio.ConfigConnect)
	for i, storageSystem := range storageSystemArray {
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
		storageID := k8s.StorageSystemID{
			ID:        powerFlexSystemID,
			IsDefault: storageSystem.IsDefault,
		}
		sdcFinder.StorageSystemID[i] = storageID
		storageClassFinder.StorageSystemID[i] = storageID
		volumeFinder.StorageSystemID[i] = storageID

		// backwards compatible with previous 'Insecure' flag
		insecure := storageSystem.Insecure || storageSystem.SkipCertificateValidation
		client, err := goscaleio.NewClientWithArgs(powerFlexEndpoint, "", math.MaxInt64, insecure, true)
		if err != nil {
			logger.WithError(err).Fatal("creating powerflex client")
		}

		_, err = client.Authenticate(&goscaleio.ConfigConnect{Username: powerFlexGatewayUser, Password: powerFlexGatewayPassword})
		if err != nil {
			logger.WithError(err).Fatalf("authenticating to powerflex %s", powerFlexSystemID)
		}

		config.PowerFlexClient[powerFlexSystemID] = client

		config.PowerFlexConfig[powerFlexSystemID] = goscaleio.ConfigConnect{Username: powerFlexGatewayUser, Password: powerFlexGatewayPassword}

		logger.WithField("storage_system_id", powerFlexSystemID).Info("set powerflex system ID")
	}

	// we need to add DriverNames explicitly here because if onConfigChange is called DriverNames would be empty
	updateProvisionerNames(sdcFinder, storageClassFinder, volumeFinder, logger)
}

func updateCollectorAddress(
	config *entrypoint.Config,
	exporter *otlexporters.OtlCollectorExporter,
	logger *logrus.Logger,
) {
	collectorAddress := viper.GetString("COLLECTOR_ADDR")
	if collectorAddress == "" {
		logger.Fatal("COLLECTOR_ADDR is required")
	}
	config.CollectorAddress = collectorAddress
	exporter.CollectorAddr = collectorAddress
}

func updateProvisionerNames(
	sdcFinder *k8s.SDCFinder,
	storageClassFinder *k8s.StorageClassFinder,
	volumeFinder *k8s.VolumeFinder,
	logger *logrus.Logger,
) {
	provisionerNamesValue := viper.GetString("provisioner_names")
	if provisionerNamesValue == "" {
		logger.Fatal("PROVISIONER_NAMES is required")
	}
	provisionerNames := strings.Split(provisionerNamesValue, ",")

	for i := range sdcFinder.StorageSystemID {
		sdcFinder.StorageSystemID[i].DriverNames = provisionerNames
	}

	for i := range storageClassFinder.StorageSystemID {
		storageClassFinder.StorageSystemID[i].DriverNames = provisionerNames
	}

	for i := range volumeFinder.StorageSystemID {
		volumeFinder.StorageSystemID[i].DriverNames = provisionerNames
	}
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

func updateService(powerflexSvc *service.PowerFlexService, logger *logrus.Logger) {
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
	powerflexSvc.MaxPowerFlexConnections = maxPowerFlexConcurrentRequests
}
