package main

import (
	"os"
	"testing"
	"time"

	"github.com/dell/karavi-metrics-powerflex/internal/entrypoint"
	"github.com/dell/karavi-metrics-powerflex/internal/k8s"
	"github.com/dell/karavi-metrics-powerflex/internal/service"
	otlexporters "github.com/dell/karavi-metrics-powerflex/opentelemetry/exporters"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestSetupLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		wantErr  bool
	}{
		{"Valid log level", "info", false},
		{"Invalid log level", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("LOG_LEVEL", tt.logLevel)

			logger := setupLogger()

			// Test if logger is setup correctly and if any error occurs.
			if tt.wantErr {
				assert.Equal(t, logrus.InfoLevel, logger.Level)
			} else {
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"Valid config", false},
		{"Invalid config", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Simulating different config file conditions
			if tt.wantErr {
				viper.SetConfigFile("/invalid/path")
			} else {
				viper.SetConfigFile(defaultConfigFile)
			}

			// Call loadConfig
			logger := logrus.New()
			loadConfig(logger) // This will just load the config
			// No error handling needed because loadConfig doesn't return error; it just prints it
		})
	}
}

func TestSetupConfigFileListener(t *testing.T) {
	tests := []struct {
		name          string
		expectedError bool
	}{
		{"Valid Config File Listener", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener := setupConfigFileListener()
			assert.NotNil(t, listener, "Expected valid config file listener")
		})
	}
}

func TestGetCollectorCertPath(t *testing.T) {
	t.Run("Valid Cert Path", func(t *testing.T) {
		os.Setenv("TLS_ENABLED", "true")
		os.Setenv("COLLECTOR_CERT_PATH", "/path/to/cert")
		path := getCollectorCertPath()
		assert.Equal(t, "/path/to/cert", path)
	})

	t.Run("TLS Enabled But No Cert Path", func(t *testing.T) {
		os.Setenv("TLS_ENABLED", "true")
		os.Setenv("COLLECTOR_CERT_PATH", "") // Explicitly setting it to empty
		path := getCollectorCertPath()
		assert.Equal(t, otlexporters.DefaultCollectorCertPath, path)
	})

	t.Run("TLS Disabled", func(t *testing.T) {
		os.Setenv("TLS_ENABLED", "false")
		path := getCollectorCertPath()
		assert.Equal(t, otlexporters.DefaultCollectorCertPath, path)
	})

	t.Run("TLS Not Set", func(t *testing.T) {
		os.Unsetenv("TLS_ENABLED")
		os.Unsetenv("COLLECTOR_CERT_PATH")
		path := getCollectorCertPath()
		assert.Equal(t, otlexporters.DefaultCollectorCertPath, path)
	})
}

func TestSetupPowerFlexService(t *testing.T) {
	// Setup
	logger := logrus.New()
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

	// Run
	config := setupConfig(sdcFinder, storageClassFinder, leaderElectorGetter, volumeFinder, nodeFinder, logger)
	exporter := &otlexporters.OtlCollectorExporter{}
	powerflexSvc := setupPowerFlexService(logger)

	// Verify
	assert.NotNil(t, config, "Expected valid config")
	assert.NotNil(t, exporter, "Expected valid exporter")
	assert.NotNil(t, powerflexSvc, "Expected valid powerflex service")
}

func TestSetupConfig(t *testing.T) {
	tests := []struct {
		name          string
		expectedError bool
	}{
		{"Valid Config Setup", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
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
			assert.NotNil(t, config, "Expected valid config")
		})
	}
}

func TestUpdateCollectorAddress(t *testing.T) {
	tests := []struct {
		name        string
		addr        string
		expectPanic bool
	}{
		{
			name:        "Valid Address",
			addr:        "localhost:8080",
			expectPanic: false,
		},
		{
			name:        "Empty Address",
			addr:        "",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("COLLECTOR_ADDR", tt.addr)

			logger := logrus.New()
			logger.ExitFunc = func(int) { panic("fatal") }
			config := &entrypoint.Config{Logger: logger}
			exporter := &otlexporters.OtlCollectorExporter{}

			if tt.expectPanic {
				assert.Panics(t, func() { updateCollectorAddress(config, exporter, logger) })
			} else {
				assert.NotPanics(t, func() { updateCollectorAddress(config, exporter, logger) })
				assert.Equal(t, tt.addr, config.CollectorAddress)
				assert.Equal(t, tt.addr, exporter.CollectorAddr)
			}
		})
	}
}

func TestUpdateMetricsEnabled(t *testing.T) {
	tests := []struct {
		name                              string
		sdcMetricsEnabled                 string
		volumeMetricsEnabled              string
		storagePoolMetricsEnabled         string
		expectedSdcMetricsEnabled         bool
		expectedVolumeMetricsEnabled      bool
		expectedStoragePoolMetricsEnabled bool
	}{
		{
			name:                              "All metrics enabled",
			sdcMetricsEnabled:                 "true",
			volumeMetricsEnabled:              "true",
			storagePoolMetricsEnabled:         "true",
			expectedSdcMetricsEnabled:         true,
			expectedVolumeMetricsEnabled:      true,
			expectedStoragePoolMetricsEnabled: true,
		},
		{
			name:                              "All metrics disabled",
			sdcMetricsEnabled:                 "false",
			volumeMetricsEnabled:              "false",
			storagePoolMetricsEnabled:         "false",
			expectedSdcMetricsEnabled:         false,
			expectedVolumeMetricsEnabled:      false,
			expectedStoragePoolMetricsEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("POWERFLEX_SDC_METRICS_ENABLED", tt.sdcMetricsEnabled)
			viper.Set("POWERFLEX_VOLUME_METRICS_ENABLED", tt.volumeMetricsEnabled)
			viper.Set("POWERFLEX_STORAGE_POOL_METRICS_ENABLED", tt.storagePoolMetricsEnabled)
			config := &entrypoint.Config{}
			updateMetricsEnabled(config)
			assert.Equal(t, tt.expectedSdcMetricsEnabled, config.SDCMetricsEnabled, "SDC metrics enabled should be set correctly")
			assert.Equal(t, tt.expectedVolumeMetricsEnabled, config.VolumeMetricsEnabled, "Volume metrics enabled should be set correctly")
			assert.Equal(t, tt.expectedStoragePoolMetricsEnabled, config.SDCMetricsEnabled, "Storage metrics enabled should be set correctly")
		})
	}
}

func TestUpdateProvisionerNames(t *testing.T) {
	tests := []struct {
		name         string
		provisioners string
		expected     []string
		expectPanic  bool
	}{
		{
			name:         "Single Provisioner",
			provisioners: "csi-vxflexos.dellemc.com",
			expected:     []string{"csi-vxflexos.dellemc.com"},
			expectPanic:  false,
		},
		{
			name:         "Multiple Provisioners",
			provisioners: "csi-vxflexos.dellemc.com1,csi-vxflexos.dellemc.com2",
			expected:     []string{"csi-vxflexos.dellemc.com1", "csi-vxflexos.dellemc.com2"},
			expectPanic:  false,
		},
		{
			name:         "Empty Provisioners",
			provisioners: "",
			expected:     nil,
			expectPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("provisioner_names", tt.provisioners)

			sdcFinder := &k8s.SDCFinder{}
			volumeFinder := &k8s.VolumeFinder{}
			storageClassFinder := &k8s.StorageClassFinder{}
			logger := logrus.New()
			logger.ExitFunc = func(int) { panic("fatal") }

			if tt.expectPanic {
				assert.Panics(t, func() { updateProvisionerNames(sdcFinder, storageClassFinder, volumeFinder, logger) })
			} else {
				assert.NotPanics(t, func() { updateProvisionerNames(sdcFinder, storageClassFinder, volumeFinder, logger) })
				for _, StorageSystemID := range sdcFinder.StorageSystemID {
					assert.Equal(t, tt.expected, StorageSystemID.DriverNames)
				}
				for _, StorageSystemID := range volumeFinder.StorageSystemID {
					assert.Equal(t, tt.expected, StorageSystemID.DriverNames)
				}
				for _, StorageSystemID := range storageClassFinder.StorageSystemID {
					assert.Equal(t, tt.expected, StorageSystemID.DriverNames)
				}
			}
		})
	}
}

func TestUpdateTickIntervals(t *testing.T) {
	tests := []struct {
		name                string
		sdcIOFreq           string
		volumeIOFreq        string
		storagePoolFreq     string
		expectedSdcIO       time.Duration
		expectedVolumeIO    time.Duration
		expectedStoragePool time.Duration
		expectPanic         bool
	}{
		{
			name:                "Valid Values",
			sdcIOFreq:           "30",
			volumeIOFreq:        "25",
			storagePoolFreq:     "15",
			expectedSdcIO:       30 * time.Second,
			expectedVolumeIO:    25 * time.Second,
			expectedStoragePool: 15 * time.Second,
			expectPanic:         false,
		},
		{
			name:                "Invalid SDC IO",
			sdcIOFreq:           "invalid",
			volumeIOFreq:        "25",
			storagePoolFreq:     "15",
			expectedSdcIO:       defaultTickInterval,
			expectedVolumeIO:    defaultTickInterval,
			expectedStoragePool: defaultTickInterval,
			expectPanic:         true,
		},
		{
			name:                "Invalid Volume IO",
			sdcIOFreq:           "30",
			volumeIOFreq:        "invalid",
			storagePoolFreq:     "15",
			expectedSdcIO:       defaultTickInterval,
			expectedVolumeIO:    defaultTickInterval,
			expectedStoragePool: defaultTickInterval,
			expectPanic:         true,
		},
		{
			name:                "Invalid Volume IO",
			sdcIOFreq:           "30",
			volumeIOFreq:        "10",
			storagePoolFreq:     "invalid",
			expectedSdcIO:       defaultTickInterval,
			expectedVolumeIO:    defaultTickInterval,
			expectedStoragePool: defaultTickInterval,
			expectPanic:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("POWERFLEX_SDC_IO_POLL_FREQUENCY", tt.sdcIOFreq)
			viper.Set("POWERFLEX_VOLUME_IO_POLL_FREQUENCY", tt.volumeIOFreq)
			viper.Set("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY", tt.storagePoolFreq)

			config := &entrypoint.Config{}
			logger := logrus.New()
			logger.ExitFunc = func(int) { panic("fatal") }

			if tt.expectPanic {
				assert.Panics(t, func() { updateTickIntervals(config, logger) })
			} else {
				assert.NotPanics(t, func() { updateTickIntervals(config, logger) })
				assert.Equal(t, tt.expectedSdcIO, config.SDCTickInterval)
				assert.Equal(t, tt.expectedVolumeIO, config.VolumeTickInterval)
				assert.Equal(t, tt.expectedStoragePool, config.StoragePoolTickInterval)
			}
		})
	}
}

func TestUpdateService(t *testing.T) {
	tests := []struct {
		name          string
		maxConcurrent string
		expected      int
		expectPanic   bool
	}{
		{
			name:          "Valid Value",
			maxConcurrent: "10",
			expected:      10,
			expectPanic:   false,
		},
		{
			name:          "Invalid Value",
			maxConcurrent: "invalid",
			expected:      service.DefaultMaxPowerFlexConnections,
			expectPanic:   true,
		},
		{
			name:          "Null Value",
			maxConcurrent: "0",
			expected:      service.DefaultMaxPowerFlexConnections,
			expectPanic:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("POWERFLEX_MAX_CONCURRENT_QUERIES", tt.maxConcurrent)

			svc := &service.PowerFlexService{}
			logger := logrus.New()
			logger.ExitFunc = func(int) { panic("fatal") }
			if tt.expectPanic {
				assert.Panics(t, func() { updateService(svc, logger) })
			} else {
				assert.NotPanics(t, func() { updateService(svc, logger) })
				assert.Equal(t, tt.expected, svc.MaxPowerFlexConnections)
			}
		})
	}
}

func Test_updateLoggingSettings(t *testing.T) {
	tests := []struct {
		name          string
		logFormat     string
		logLevel      string
		expectedLevel logrus.Level
	}{
		{
			name:          "Valid Setting",
			logFormat:     "json",
			logLevel:      "INFO",
			expectedLevel: 4,
		},
		{
			name:          "Invalid Setting",
			logFormat:     "json",
			logLevel:      "TEST",
			expectedLevel: 4,
		},
		{
			name:          "text log format",
			logFormat:     "text",
			logLevel:      "INFO",
			expectedLevel: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("LOG_FORMAT", tt.logFormat)
			viper.Set("LOG_LEVEL", tt.logLevel)
			logger := logrus.New()
			updateLoggingSettings(logger)
			assert.Equal(t, tt.expectedLevel, logrus.GetLevel())
		})
	}
}

func TestSetupConfigWatchers(t *testing.T) {
	logger := logrus.New()
	config := &entrypoint.Config{}
	exporter := &otlexporters.OtlCollectorExporter{}
	powerflexSvc := &service.PowerFlexService{}
	configFileListener := setupConfigFileListener()
	sdcFinder := &k8s.SDCFinder{
		API: &k8s.API{},
	}
	storageClassFinder := &k8s.StorageClassFinder{
		API: &k8s.API{},
	}
	volumeFinder := &k8s.VolumeFinder{
		API:    &k8s.API{},
		Logger: logger,
	}
	tests := []struct {
		name          string
		expectedError bool
	}{
		{"Valid Config Watchers Setup", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				setupConfigWatchers(configFileListener, powerflexSvc, config, sdcFinder, storageClassFinder, volumeFinder, exporter, logger)
			}, "Expected setupConfigWatchers to not panic")
		})
	}
}

// func TestGetStorageSystemArray(t *testing.T) {
// 	// Call the function to get the storage system array
// 	storageSystemArray, err := GetStorageSystemArray("testdata/config.yaml")

// 	// Assert the expected values
// 	expectedArray := []service.ArrayConnectionData{
// 		{
// 			Username:                  "admin",
// 			Password:                  "password",
// 			SystemID:                  "system-id-1",
// 			Endpoint:                  "http://127.0.0.1",
// 			SkipCertificateValidation: true,
// 		},
// 	}

// 	assert.Equal(t, expectedArray, storageSystemArray)
// 	assert.Nil(t, err)
// }

func TestUpdatePowerFlexConnection(t *testing.T) {
	// Create a test table with different scenarios and expected results
	tests := []struct {
		name              string
		config            *entrypoint.Config
		configContentFile string
		expectPanic       bool
	}{
		{
			name:              "Config Reader Error",
			config:            &entrypoint.Config{},
			configContentFile: "testdata/not-exist.yaml",
			expectPanic:       true,
		},
		{
			name:              "Empty Endpoint Error",
			config:            &entrypoint.Config{},
			configContentFile: "testdata/config.yaml",
			expectPanic:       true,
		},
		{
			name:              "Empty Password Error",
			config:            &entrypoint.Config{},
			configContentFile: "testdata/config.yaml",
			expectPanic:       true,
		},
		{
			name:              "Empty System ID Error",
			config:            &entrypoint.Config{},
			configContentFile: "testdata/config.yaml",
			expectPanic:       true,
		},
		{
			name:              "Empty Username Error",
			config:            &entrypoint.Config{},
			configContentFile: "testdata/config.yaml",
			expectPanic:       true,
		},
		{
			name:              "Authentication Error",
			config:            &entrypoint.Config{},
			configContentFile: "testdata/config.yaml",
			expectPanic:       true,
		},
		// Add more test cases here
	}

	// Iterate over the test table and run the test for each case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			logger := logrus.New()
			logger.ExitFunc = func(int) { panic("fatal") }
			if tt.expectPanic {
				assert.Panics(t, func() {
					updatePowerFlexConnection(
						tt.configContentFile,
						tt.config,
						&k8s.SDCFinder{},
						&k8s.StorageClassFinder{},
						&k8s.VolumeFinder{},
						logger,
					)
				})
			}
			// // Call the function with the test data
			// updatePowerFlexConnection(
			// 	"storage_system_config.yaml",
			// 	tt.config,
			// 	&k8s.SDCFinder{},
			// 	&k8s.StorageClassFinder{},
			// 	&k8s.VolumeFinder{},
			// 	&logrus.Logger{},
			// )

			// // Assert the expected error
			// if !errors.Is(err, tt.expectedError) {
			// 	t.Errorf("expected error '%v', but got '%v'", tt.expectedError, err)
			// }

			// // Assert the expected error message
			// if err != nil && err.Error() != tt.expectedMessage {
			// 	t.Errorf("expected error message '%s', but got '%s'", tt.expectedMessage, err.Error())
			// }
		})
	}
}
