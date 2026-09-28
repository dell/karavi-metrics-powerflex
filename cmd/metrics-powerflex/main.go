/*
 Copyright (c) 2025  Dell Inc. or its subsidiaries. All Rights Reserved.

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

	csmserver "github.com/dell/csm-metrics-common/pkg/server"
	"github.com/dell/csmlog"
	"github.com/dell/goscaleio"
	"github.com/dell/karavi-metrics-powerflex/internal/entrypoint"
	"github.com/dell/karavi-metrics-powerflex/internal/k8s"
	"github.com/dell/karavi-metrics-powerflex/internal/service"
	otlexporters "github.com/dell/karavi-metrics-powerflex/opentelemetry/exporters"
	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

const (
	defaultTickInterval            = 5 * time.Second
	defaultConfigFile              = "/etc/config/karavi-metrics-powerflex.yaml"
	defaultStorageSystemConfigFile = "/vxflexos-config/config"
	defaultObsMetricsPort          = "8443"
)

const (
	csiObsMetricsEnabledKey = "X_CSI_METRICS_ENABLED"
	csiObsMetricsPortKey    = "X_CSI_METRICS_PORT"
	csiObsMetricsCertKey    = "X_CSI_METRICS_TLS_CERT_FILE"
	csiObsMetricsKeyKey     = "X_CSI_METRICS_TLS_KEY_FILE"
)

var goscaleioClient = goscaleio.NewClientWithArgs

func main() {
	config, exporter, powerflexSvc := configure()
	if shouldStartObservabilityMetricsServer() {
		startMetricsServer(powerflexSvc)
	}
	if err := entrypoint.Run(context.Background(), config, exporter, powerflexSvc); err != nil {
		csmlog.Fatalf("Failed to run service: %v", err)
	}
}

func shouldStartObservabilityMetricsServer() bool {
	enabled := strings.EqualFold(strings.TrimSpace(viper.GetString(csiObsMetricsEnabledKey)), "true")
	if enabled {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv(csiObsMetricsEnabledKey)), "true")
}

func configure() (*entrypoint.Config, otlexporters.Otlexporter, *service.PowerFlexService) {
	loadConfig()
	setLoggingSettings()
	configFileListener := setupConfigFileListener()
	sdcFinder, storageClassFinder, leaderElectorGetter, volumeFinder, nodeFinder, exporter := initializeComponents()
	config := setupConfig(sdcFinder, storageClassFinder, leaderElectorGetter, volumeFinder, nodeFinder)
	powerflexSvc := setupPowerFlexService(volumeFinder)
	onChangeUpdate(powerflexSvc, config, sdcFinder, exporter, storageClassFinder, volumeFinder)
	updatePowerFlexConnection(defaultStorageSystemConfigFile, config, sdcFinder, storageClassFinder, volumeFinder)
	setupConfigWatchers(configFileListener, powerflexSvc, config, sdcFinder, storageClassFinder, volumeFinder, exporter)
	return config, exporter, powerflexSvc
}

func initializeComponents() (*k8s.SDCFinder, *k8s.StorageClassFinder, *k8s.LeaderElector, *k8s.VolumeFinder, *k8s.NodeFinder, *otlexporters.OtlCollectorExporter) {
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
	exporter := &otlexporters.OtlCollectorExporter{}
	return sdcFinder, storageClassFinder, leaderElectorGetter, volumeFinder, nodeFinder, exporter
}

// loadConfig loads the primary configuration file.
func loadConfig() {
	viper.SetConfigFile(defaultConfigFile)
	viper.AutomaticEnv()
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
	if tls := os.Getenv("TLS_ENABLED"); tls == "true" {
		if certPath := strings.TrimSpace(os.Getenv("COLLECTOR_CERT_PATH")); certPath != "" {
			return certPath
		}
	}
	return otlexporters.DefaultCollectorCertPath
}

// setupConfig creates the main configuration structure.
func setupConfig(
	sdcFinder *k8s.SDCFinder,
	storageClassFinder *k8s.StorageClassFinder,
	leaderElectorGetter *k8s.LeaderElector,
	volumeFinder *k8s.VolumeFinder,
	nodeFinder *k8s.NodeFinder,
) *entrypoint.Config {
	return &entrypoint.Config{
		SDCFinder:          sdcFinder,
		StorageClassFinder: storageClassFinder,
		LeaderElector:      leaderElectorGetter,
		VolumeFinder:       volumeFinder,
		NodeFinder:         nodeFinder,
		CollectorCertPath:  getCollectorCertPath(),
	}
}

func setupPowerFlexService(volumeFinder *k8s.VolumeFinder) *service.PowerFlexService {
	return &service.PowerFlexService{
		MetricsWrapper: &service.MetricsWrapper{
			Meter: otel.Meter("powerflex/sdc"),
		},
		VolumeFinder: volumeFinder,
	}
}

func onChangeUpdate(
	powerflexSvc *service.PowerFlexService,
	config *entrypoint.Config,
	sdcFinder *k8s.SDCFinder,
	exporter *otlexporters.OtlCollectorExporter,
	storageClassFinder *k8s.StorageClassFinder,
	volumeFinder *k8s.VolumeFinder,
) {
	updateCollectorAddress(config, exporter)
	updateProvisionerNames(sdcFinder, storageClassFinder, volumeFinder)
	updateMetricsEnabled(config)
	updateTickIntervals(config)
	updateService(powerflexSvc)
}

func setLoggingSettings() {
	logFormat := viper.GetString("LOG_FORMAT")
	if strings.EqualFold(logFormat, "json") {
		csmlog.SetFormat("json")
	} else {
		csmlog.SetFormat("text")
	}

	logLevel := viper.GetString("LOG_LEVEL")
	level, err := csmlog.ParseLevel(logLevel)
	if err != nil {
		csmlog.WithFields(csmlog.Fields{"configured_log_level": logLevel}).Infof("Invalid log level configured, defaulting to INFO: %v", err)
		level = csmlog.InfoLevel
	}
	csmlog.SetLevel(level)
}

// setupConfigWatchers sets up dynamic updates when config files change.
func setupConfigWatchers(configFileListener *viper.Viper, powerflexSvc *service.PowerFlexService, config *entrypoint.Config, sdcFinder *k8s.SDCFinder, storageClassFinder *k8s.StorageClassFinder, volumeFinder *k8s.VolumeFinder, exporter *otlexporters.OtlCollectorExporter) {
	viper.WatchConfig()
	viper.OnConfigChange(func(_ fsnotify.Event) {
		setLoggingSettings()
	})

	configFileListener.WatchConfig()
	configFileListener.OnConfigChange(func(_ fsnotify.Event) {
		onChangeUpdate(powerflexSvc, config, sdcFinder, exporter, storageClassFinder, volumeFinder)
		updatePowerFlexConnection(defaultStorageSystemConfigFile, config, sdcFinder, storageClassFinder, volumeFinder)
	})
}

func updatePowerFlexConnection(
	storageSystemConfigFile string,
	config *entrypoint.Config,
	sdcFinder *k8s.SDCFinder,
	storageClassFinder *k8s.StorageClassFinder,
	volumeFinder *k8s.VolumeFinder,
) {
	configReader := service.ConfigurationReader{}
	storageSystemArray, err := configReader.GetStorageSystemConfiguration(storageSystemConfigFile)
	if err != nil {
		csmlog.Fatalf("Failed to get storage system configuration: %v", err)
	}

	volumeFinder.StorageSystemID = make([]k8s.StorageSystemID, len(storageSystemArray))
	sdcFinder.StorageSystemID = make([]k8s.StorageSystemID, len(storageSystemArray))
	storageClassFinder.StorageSystemID = make([]k8s.StorageSystemID, len(storageSystemArray))

	config.PowerFlexClient = make(map[string]service.PowerFlexClient)
	config.PowerFlexConfig = make(map[string]goscaleio.ConfigConnect)
	for i, storageSystem := range storageSystemArray {
		powerFlexEndpoint := storageSystem.Endpoint
		powerFlexGatewayUser := storageSystem.Username
		powerFlexGatewayPassword := storageSystem.Password
		powerFlexSystemID := storageSystem.SystemID
		storageID := k8s.StorageSystemID{
			ID:               powerFlexSystemID,
			AvailabilityZone: storageSystem.AvailabilityZone,
			IsDefault:        storageSystem.IsDefault,
		}
		sdcFinder.StorageSystemID[i] = storageID
		storageClassFinder.StorageSystemID[i] = storageID
		volumeFinder.StorageSystemID[i] = storageID

		// backwards compatible with previous 'Insecure' flag
		insecure := storageSystem.Insecure || storageSystem.SkipCertificateValidation
		client, err := goscaleioClient(powerFlexEndpoint, "", math.MaxInt64, insecure, true, "")
		if err != nil {
			csmlog.Fatalf("Failed to create PowerFlex client: %v", err)
		}

		_, err = client.Authenticate(&goscaleio.ConfigConnect{Username: powerFlexGatewayUser, Password: powerFlexGatewayPassword})
		if err != nil {
			csmlog.Fatalf("Failed to authenticate with PowerFlex %s: %v", powerFlexSystemID, err)
		}

		config.PowerFlexClient[powerFlexSystemID] = client
		config.PowerFlexConfig[powerFlexSystemID] = goscaleio.ConfigConnect{Username: powerFlexGatewayUser, Password: powerFlexGatewayPassword}
		csmlog.WithFields(csmlog.Fields{"storage_system_id": powerFlexSystemID}).Info("set powerflex system ID")
	}

	// we need to add DriverNames explicitly here because if onConfigChange is called DriverNames would be empty
	updateProvisionerNames(sdcFinder, storageClassFinder, volumeFinder)
}

func updateCollectorAddress(
	config *entrypoint.Config,
	exporter *otlexporters.OtlCollectorExporter,
) {
	collectorAddress := viper.GetString("COLLECTOR_ADDR")
	if collectorAddress == "" {
		csmlog.Fatal("COLLECTOR_ADDR is required")
	}
	config.CollectorAddress = collectorAddress
	exporter.CollectorAddr = collectorAddress
}

func updateProvisionerNames(
	sdcFinder *k8s.SDCFinder,
	storageClassFinder *k8s.StorageClassFinder,
	volumeFinder *k8s.VolumeFinder,
) {
	provisionerNamesValue := viper.GetString("provisioner_names")
	if provisionerNamesValue == "" {
		csmlog.Fatal("PROVISIONER_NAMES is required")
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
	if powerflexSdcMetricsEnabledValue != "true" && powerflexSdcMetricsEnabledValue != "false" {
		csmlog.Fatal("Invalid POWERFLEX_SDC_METRICS_ENABLED value. Valid values are true or false")
	}

	powerflexVolumeMetricsEnabled := true
	powerflexVolumeMetricsEnabledValue := viper.GetString("POWERFLEX_VOLUME_METRICS_ENABLED")
	if powerflexVolumeMetricsEnabledValue == "false" {
		powerflexVolumeMetricsEnabled = false
	}
	if powerflexVolumeMetricsEnabledValue != "true" && powerflexVolumeMetricsEnabledValue != "false" {
		csmlog.Fatal("Invalid POWERFLEX_VOLUME_METRICS_ENABLED value. Valid values are true or false")
	}

	storagePoolMetricsEnabled := true
	storagePoolMetricsEnabledValue := viper.GetString("POWERFLEX_STORAGE_POOL_METRICS_ENABLED")
	if storagePoolMetricsEnabledValue == "false" {
		storagePoolMetricsEnabled = false
	}
	if storagePoolMetricsEnabledValue != "true" && storagePoolMetricsEnabledValue != "false" {
		csmlog.Fatal("Invalid POWERFLEX_STORAGE_POOL_METRICS_ENABLED value. Valid values are true or false")
	}
	config.SDCMetricsEnabled = powerflexSdcMetricsEnabled
	config.VolumeMetricsEnabled = powerflexVolumeMetricsEnabled
	config.StoragePoolMetricsEnabled = storagePoolMetricsEnabled

	powerflexTopologyMetricsEnabled := true
	powerflexTopologyMetricsEnabledValue := viper.GetString("POWERFLEX_TOPOLOGY_METRICS_ENABLED")
	if powerflexTopologyMetricsEnabledValue == "false" {
		powerflexTopologyMetricsEnabled = false
	}

	config.TopologyMetricsEnabled = powerflexTopologyMetricsEnabled
}

func updateTickIntervals(config *entrypoint.Config) {
	sdcTickInterval := defaultTickInterval
	sdcIoPollFrequencySeconds := viper.GetString("POWERFLEX_SDC_IO_POLL_FREQUENCY")
	if sdcIoPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(sdcIoPollFrequencySeconds)
		if err != nil {
			csmlog.Fatalf("Invalid POWERFLEX_SDC_IO_POLL_FREQUENCY. Specify a valid number: %v", err)
		}
		if numSeconds <= 0 {
			csmlog.Fatal("Invalid POWERFLEX_SDC_IO_POLL_FREQUENCY value. Must be greater than 0")
		}
		sdcTickInterval = time.Duration(numSeconds) * time.Second
	}

	volumeTickInterval := defaultTickInterval
	volIoPollFrequencySeconds := viper.GetString("POWERFLEX_VOLUME_IO_POLL_FREQUENCY")
	if volIoPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(volIoPollFrequencySeconds)
		if err != nil {
			csmlog.Fatalf("Invalid POWERFLEX_VOLUME_IO_POLL_FREQUENCY. Specify a valid number: %v", err)
		}
		if numSeconds <= 0 {
			csmlog.Fatal("Invalid POWERFLEX_VOLUME_IO_POLL_FREQUENCY value. Must be greater than 0")
		}
		volumeTickInterval = time.Duration(numSeconds) * time.Second
	}

	storagePoolTickInterval := defaultTickInterval
	storagePoolPollFrequencySeconds := viper.GetString("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY")
	if storagePoolPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(storagePoolPollFrequencySeconds)
		if err != nil {
			csmlog.Fatalf("Invalid POWERFLEX_STORAGE_POOL_POLL_FREQUENCY. Specify a valid number: %v", err)
		}
		if numSeconds <= 0 {
			csmlog.Fatal("Invalid POWERFLEX_STORAGE_POOL_POLL_FREQUENCY value. Must be greater than 0")
		}
		storagePoolTickInterval = time.Duration(numSeconds) * time.Second
	}

	config.SDCTickInterval = sdcTickInterval
	config.VolumeTickInterval = volumeTickInterval
	config.StoragePoolTickInterval = storagePoolTickInterval

	topologyMetricsTickInterval := defaultTickInterval
	topologyMetricsPollFrequencySeconds := viper.GetString("POWERFLEX_TOPOLOGY_METRICS_POLL_FREQUENCY")
	if topologyMetricsPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(topologyMetricsPollFrequencySeconds)
		if err != nil {
			csmlog.Fatalf("Invalid POWERFLEX_TOPOLOGY_METRICS_POLL_FREQUENCY. Specify a valid number: %v", err)
		}

		topologyMetricsTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.TopologyMetricsTickInterval = topologyMetricsTickInterval
	csmlog.WithFields(csmlog.Fields{"cluster_performance_tick_interval": fmt.Sprintf("%v", topologyMetricsTickInterval)}).Debug("setting cluster performance tick interval")
}

func updateService(powerflexSvc *service.PowerFlexService) {
	maxPowerFlexConcurrentRequests := service.DefaultMaxPowerFlexConnections
	maxPowerFlexConcurrentRequestsVar := viper.GetString("POWERFLEX_MAX_CONCURRENT_QUERIES")
	if maxPowerFlexConcurrentRequestsVar != "" {
		maxPowerFlexConcurrentRequests, err := strconv.Atoi(maxPowerFlexConcurrentRequestsVar)
		if err != nil {
			csmlog.Fatalf("POWERFLEX_MAX_CONCURRENT_QUERIES was not set to a valid number: %v", err)
		}
		if maxPowerFlexConcurrentRequests <= 0 {
			csmlog.Fatal("POWERFLEX_MAX_CONCURRENT_QUERIES value was invalid (<= 0)")
		}
	}
	powerflexSvc.MaxPowerFlexConnections = maxPowerFlexConcurrentRequests
}

// validateTLSFiles checks that the certificate and key files exist and are readable.
func validateTLSFiles(certFile, keyFile string) error {
	for _, path := range []string{certFile, keyFile} {
		f, err := os.Open(path) // #nosec G304 -- path comes from trusted configuration
		if err != nil {
			return fmt.Errorf("cannot open TLS file %q: %w", path, err)
		}
		f.Close()
	}
	return nil
}

// startMetricsServer creates a Prometheus registry, registers the PFLXObsInstrumenter,
// and starts the observability self-metrics HTTP(S) server in a background goroutine.
// HTTPS is enabled when both X_CSI_METRICS_TLS_CERT_FILE and X_CSI_METRICS_TLS_KEY_FILE
// are configured; otherwise the server serves plain HTTP.
func startMetricsServer(powerflexSvc *service.PowerFlexService) {
	reg := prometheus.NewRegistry()
	powerflexSvc.ObsInstrumenter = service.NewPFLXObsInstrumenter(reg)

	viper.SetDefault(csiObsMetricsPortKey, defaultObsMetricsPort)
	metricsPort := viper.GetString(csiObsMetricsPortKey)

	certFile := viper.GetString(csiObsMetricsCertKey)
	keyFile := viper.GetString(csiObsMetricsKeyKey)

	scheme := "http"
	if certFile != "" || keyFile != "" {
		if certFile == "" || keyFile == "" {
			csmlog.Fatal("both X_CSI_METRICS_TLS_CERT_FILE and X_CSI_METRICS_TLS_KEY_FILE must be set for TLS")
		}
		if err := validateTLSFiles(certFile, keyFile); err != nil {
			csmlog.WithFields(csmlog.Fields{
				"error": err,
				"cert":  certFile,
				"key":   keyFile,
			}).Fatal("observability metrics server failed to start: invalid TLS configuration")
		}
		scheme = "https"
	}

	srv := csmserver.NewMetricsServer(csmserver.Config{
		Port:     fmt.Sprintf(":%s", metricsPort),
		CertFile: certFile,
		KeyFile:  keyFile,
		Registry: reg,
	})

	go func() {
		csmlog.WithFields(csmlog.Fields{
			"port":   metricsPort,
			"scheme": scheme,
		}).Info("starting observability metrics server")
		if err := srv.Start(); err != nil {
			csmlog.WithFields(csmlog.Fields{"error": err}).Error("observability metrics server closed")
		}
	}()
}
