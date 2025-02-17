package main

import (
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
			name:                "Invalid Quota",
			sdcIOFreq:           "invalid",
			volumeIOFreq:        "",
			storagePoolFreq:     "",
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
