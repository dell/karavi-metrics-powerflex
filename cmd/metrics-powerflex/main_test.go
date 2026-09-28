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
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/dell/csmlog"
	"github.com/dell/goscaleio"
	"github.com/dell/karavi-metrics-powerflex/internal/entrypoint"
	"github.com/dell/karavi-metrics-powerflex/internal/k8s"
	"github.com/dell/karavi-metrics-powerflex/internal/service"
	otlexporters "github.com/dell/karavi-metrics-powerflex/opentelemetry/exporters"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func assertFatalPath(t *testing.T, scenario, want string) {
	t.Helper()
	cmd := exec.Command("/proc/self/exe", "-test.run=TestFatalPathHelper")
	cmd.Env = append(os.Environ(), "TEST_FATAL_PATH=1", "TEST_FATAL_CASE="+scenario)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected fatal exit for %s", scenario)
	}
	if !strings.Contains(string(output), want) {
		t.Fatalf("expected output for %s to contain %q, got %s", scenario, want, string(output))
	}
}

func TestFatalPathHelper(t *testing.T) {
	if os.Getenv("TEST_FATAL_PATH") != "1" {
		return
	}
	scenario := os.Getenv("TEST_FATAL_CASE")
	switch scenario {
	case "onChangeUpdate/Empty Address":
		viper.Reset()
		svc := &service.PowerFlexService{}
		sdcFinder := &k8s.SDCFinder{API: &k8s.API{}}
		storageClassFinder := &k8s.StorageClassFinder{API: &k8s.API{}}
		volumeFinder := &k8s.VolumeFinder{API: &k8s.API{}}
		config := &entrypoint.Config{}
		exporter := &otlexporters.OtlCollectorExporter{}
		onChangeUpdate(svc, config, sdcFinder, exporter, storageClassFinder, volumeFinder)
	case "updateCollectorAddress/Empty Address":
		viper.Reset()
		viper.Set("COLLECTOR_ADDR", "")
		updateCollectorAddress(&entrypoint.Config{}, &otlexporters.OtlCollectorExporter{})
	case "updateMetricsEnabled/sdcMetricsEnabled error":
		viper.Reset()
		viper.Set("POWERFLEX_SDC_METRICS_ENABLED", "test")
		viper.Set("POWERFLEX_VOLUME_METRICS_ENABLED", "true")
		viper.Set("POWERFLEX_STORAGE_POOL_METRICS_ENABLED", "true")
		viper.Set("POWERFLEX_TOPOLOGY_METRICS_ENABLED", "true")
		updateMetricsEnabled(&entrypoint.Config{})
	case "updateMetricsEnabled/volumeMetricsEnabled error":
		viper.Reset()
		viper.Set("POWERFLEX_SDC_METRICS_ENABLED", "true")
		viper.Set("POWERFLEX_VOLUME_METRICS_ENABLED", "test")
		viper.Set("POWERFLEX_STORAGE_POOL_METRICS_ENABLED", "true")
		viper.Set("POWERFLEX_TOPOLOGY_METRICS_ENABLED", "true")
		updateMetricsEnabled(&entrypoint.Config{})
	case "updateMetricsEnabled/storagePoolMetricsEnabled error":
		viper.Reset()
		viper.Set("POWERFLEX_SDC_METRICS_ENABLED", "true")
		viper.Set("POWERFLEX_VOLUME_METRICS_ENABLED", "true")
		viper.Set("POWERFLEX_STORAGE_POOL_METRICS_ENABLED", "test")
		viper.Set("POWERFLEX_TOPOLOGY_METRICS_ENABLED", "true")
		updateMetricsEnabled(&entrypoint.Config{})
	case "updateProvisionerNames/Empty Provisioners":
		viper.Reset()
		viper.Set("provisioner_names", "")
		sdcFinder := &k8s.SDCFinder{StorageSystemID: []k8s.StorageSystemID{{ID: "system-id"}}}
		storageClassFinder := &k8s.StorageClassFinder{StorageSystemID: []k8s.StorageSystemID{{ID: "system-id"}}}
		volumeFinder := &k8s.VolumeFinder{StorageSystemID: []k8s.StorageSystemID{{ID: "system-id"}}}
		updateProvisionerNames(sdcFinder, storageClassFinder, volumeFinder)
	case "updateTickIntervals/Invalid SDC IO":
		viper.Reset()
		viper.Set("POWERFLEX_SDC_IO_POLL_FREQUENCY", "invalid")
		viper.Set("POWERFLEX_VOLUME_IO_POLL_FREQUENCY", "25")
		viper.Set("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY", "15")
		viper.Set("POWERFLEX_TOPOLOGY_METRICS_POLL_FREQUENCY", "invalid")
		updateTickIntervals(&entrypoint.Config{})
	case "updateTickIntervals/Invalid Volume IO":
		viper.Reset()
		viper.Set("POWERFLEX_SDC_IO_POLL_FREQUENCY", "30")
		viper.Set("POWERFLEX_VOLUME_IO_POLL_FREQUENCY", "invalid")
		viper.Set("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY", "15")
		viper.Set("POWERFLEX_TOPOLOGY_METRICS_POLL_FREQUENCY", "invalid")
		updateTickIntervals(&entrypoint.Config{})
	case "updateTickIntervals/Invalid Storage Pool":
		viper.Reset()
		viper.Set("POWERFLEX_SDC_IO_POLL_FREQUENCY", "30")
		viper.Set("POWERFLEX_VOLUME_IO_POLL_FREQUENCY", "10")
		viper.Set("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY", "invalid")
		viper.Set("POWERFLEX_TOPOLOGY_METRICS_POLL_FREQUENCY", "invalid")
		updateTickIntervals(&entrypoint.Config{})
	case "updateTickIntervals/Negative SDC IO":
		viper.Reset()
		viper.Set("POWERFLEX_SDC_IO_POLL_FREQUENCY", "-1")
		viper.Set("POWERFLEX_VOLUME_IO_POLL_FREQUENCY", "25")
		viper.Set("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY", "15")
		viper.Set("POWERFLEX_TOPOLOGY_METRICS_POLL_FREQUENCY", "invalid")
		updateTickIntervals(&entrypoint.Config{})
	case "updateTickIntervals/Negative Volume IO":
		viper.Reset()
		viper.Set("POWERFLEX_SDC_IO_POLL_FREQUENCY", "30")
		viper.Set("POWERFLEX_VOLUME_IO_POLL_FREQUENCY", "-1")
		viper.Set("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY", "15")
		updateTickIntervals(&entrypoint.Config{})
	case "updateTickIntervals/Negative Storage Pool":
		viper.Reset()
		viper.Set("POWERFLEX_SDC_IO_POLL_FREQUENCY", "30")
		viper.Set("POWERFLEX_VOLUME_IO_POLL_FREQUENCY", "25")
		viper.Set("POWERFLEX_STORAGE_POOL_POLL_FREQUENCY", "-1")
		updateTickIntervals(&entrypoint.Config{})
	case "updateService/Invalid Value":
		viper.Reset()
		viper.Set("POWERFLEX_MAX_CONCURRENT_QUERIES", "invalid")
		updateService(&service.PowerFlexService{})
	case "updateService/Null Value":
		viper.Reset()
		viper.Set("POWERFLEX_MAX_CONCURRENT_QUERIES", "0")
		updateService(&service.PowerFlexService{})
	case "updatePowerFlexConnection/Config Reader Error":
		viper.Reset()
		updatePowerFlexConnection("testdata/not-exist.yaml", &entrypoint.Config{}, &k8s.SDCFinder{}, &k8s.StorageClassFinder{}, &k8s.VolumeFinder{})
	case "updatePowerFlexConnection/Empty Endpoint Error":
		viper.Reset()
		updatePowerFlexConnection("testdata/invalid-endpoint-config.yaml", &entrypoint.Config{}, &k8s.SDCFinder{}, &k8s.StorageClassFinder{}, &k8s.VolumeFinder{})
	case "updatePowerFlexConnection/Empty Password Error":
		viper.Reset()
		updatePowerFlexConnection("testdata/invalid-password-config.yaml", &entrypoint.Config{}, &k8s.SDCFinder{}, &k8s.StorageClassFinder{}, &k8s.VolumeFinder{})
	case "updatePowerFlexConnection/Empty System ID Error":
		viper.Reset()
		updatePowerFlexConnection("testdata/invalid-systemid-config.yaml", &entrypoint.Config{}, &k8s.SDCFinder{}, &k8s.StorageClassFinder{}, &k8s.VolumeFinder{})
	case "updatePowerFlexConnection/Empty Username Error":
		viper.Reset()
		updatePowerFlexConnection("testdata/invalid-username-config.yaml", &entrypoint.Config{}, &k8s.SDCFinder{}, &k8s.StorageClassFinder{}, &k8s.VolumeFinder{})
	case "updatePowerFlexConnection/Authentication Error":
		viper.Reset()
		viper.Set("provisioner_names", "csi-vxflexos.dellemc.com")
		updatePowerFlexConnection("testdata/config.yaml", &entrypoint.Config{}, &k8s.SDCFinder{}, &k8s.StorageClassFinder{}, &k8s.VolumeFinder{})
	case "updatePowerFlexConnection/Client Error":
		tmpFile, err := os.CreateTemp("", "powerflex-config-*.yaml")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		_, err = tmpFile.WriteString("- username: admin\n  password: password\n  systemID: test-system\n  endpoint: http://127.0.0.1\n")
		if err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}
		_ = tmpFile.Close()
		origClient := goscaleioClient
		goscaleioClient = func(string, string, int64, bool, bool, string) (*goscaleio.Client, error) {
			return nil, fmt.Errorf("mock client creation error")
		}
		defer func() { goscaleioClient = origClient }()
		viper.Reset()
		viper.Set("provisioner_names", "csi-vxflexos.dellemc.com")
		updatePowerFlexConnection(tmpFile.Name(), &entrypoint.Config{}, &k8s.SDCFinder{}, &k8s.StorageClassFinder{}, &k8s.VolumeFinder{})
	default:
		t.Fatalf("unknown fatal scenario %q", scenario)
	}
	t.Fatalf("scenario %q did not exit fatally", scenario)
}

func TestInitializeComponents(t *testing.T) {
	tests := []struct {
		name         string
		provisioners string
		expected     []string
	}{
		{
			name:         "Single Provisioner",
			provisioners: "csi-vxflexos.dellemc.com",
			expected:     []string{"csi-vxflexos.dellemc.com"},
		},
		{
			name:         "Empty Provisioners",
			provisioners: "",
			expected:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("provisioner_names", tt.provisioners)
			sdcFinder, storageClassFinder, _, volumeFinder, _, _ := initializeComponents()
			for _, StorageSystemID := range sdcFinder.StorageSystemID {
				assert.Equal(t, tt.expected, StorageSystemID.DriverNames)
			}
			for _, StorageSystemID := range volumeFinder.StorageSystemID {
				assert.Equal(t, tt.expected, StorageSystemID.DriverNames)
			}
			for _, StorageSystemID := range storageClassFinder.StorageSystemID {
				assert.Equal(t, tt.expected, StorageSystemID.DriverNames)
			}
		})
	}
}

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

			assert.NotPanics(t, func() {
				loadConfig()
				setLoggingSettings()
			})
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
			loadConfig() // This will just load the config
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
		_ = os.Setenv("TLS_ENABLED", "true")
		_ = os.Setenv("COLLECTOR_CERT_PATH", "/path/to/cert")
		path := getCollectorCertPath()
		assert.Equal(t, "/path/to/cert", path)
	})

	t.Run("TLS Enabled But No Cert Path", func(t *testing.T) {
		_ = os.Setenv("TLS_ENABLED", "true")
		_ = os.Setenv("COLLECTOR_CERT_PATH", "") // Explicitly setting it to empty
		path := getCollectorCertPath()
		assert.Equal(t, otlexporters.DefaultCollectorCertPath, path)
	})

	t.Run("TLS Disabled", func(t *testing.T) {
		_ = os.Setenv("TLS_ENABLED", "false")
		path := getCollectorCertPath()
		assert.Equal(t, otlexporters.DefaultCollectorCertPath, path)
	})

	t.Run("TLS Not Set", func(t *testing.T) {
		_ = os.Unsetenv("TLS_ENABLED")
		_ = os.Unsetenv("COLLECTOR_CERT_PATH")
		path := getCollectorCertPath()
		assert.Equal(t, otlexporters.DefaultCollectorCertPath, path)
	})
}

func TestSetupPowerFlexService(t *testing.T) {
	// Setup
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

	// Run
	config := setupConfig(sdcFinder, storageClassFinder, leaderElectorGetter, volumeFinder, nodeFinder)
	exporter := &otlexporters.OtlCollectorExporter{}
	powerflexSvc := setupPowerFlexService(volumeFinder)

	// Verify
	assert.NotNil(t, config, "Expected valid config")
	assert.NotNil(t, exporter, "Expected valid exporter")
	assert.NotNil(t, powerflexSvc, "Expected valid powerflex service")
}

func TestOnChangeUpdate(t *testing.T) {
	tests := []struct {
		name        string
		expectPanic bool
	}{
		{
			name:        "Empty Address",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			svc := &service.PowerFlexService{}
			sdcFinder := &k8s.SDCFinder{
				API: &k8s.API{},
			}
			storageClassFinder := &k8s.StorageClassFinder{
				API: &k8s.API{},
			}
			volumeFinder := &k8s.VolumeFinder{
				API: &k8s.API{},
			}
			config := &entrypoint.Config{}
			exporter := &otlexporters.OtlCollectorExporter{}
			_ = svc
			_ = sdcFinder
			_ = storageClassFinder
			_ = volumeFinder
			_ = config
			_ = exporter
			if tt.expectPanic {
				assertFatalPath(t, "onChangeUpdate/"+tt.name, "COLLECTOR_ADDR is required")
			}
		})
	}
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
			config := setupConfig(sdcFinder, storageClassFinder, leaderElectorGetter, volumeFinder, nodeFinder)
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

			config := &entrypoint.Config{}
			exporter := &otlexporters.OtlCollectorExporter{}

			if tt.expectPanic {
				assertFatalPath(t, "updateCollectorAddress/"+tt.name, "COLLECTOR_ADDR is required")
			} else {
				assert.NotPanics(t, func() { updateCollectorAddress(config, exporter) })
				assert.Equal(t, tt.addr, config.CollectorAddress)
				assert.Equal(t, tt.addr, exporter.CollectorAddr)
			}
		})
	}
}

func TestUpdateMetricsEnabled(t *testing.T) {
	tests := []struct {
		name                                    string
		sdcMetricsEnabled                       string
		volumeMetricsEnabled                    string
		storagePoolMetricsEnabled               string
		powerflexTopologyMetricsEnabled         string
		expectedSdcMetricsEnabled               bool
		expectedVolumeMetricsEnabled            bool
		expectedStoragePoolMetricsEnabled       bool
		expectedPowerflexTopologyMetricsEnabled bool
		expectPanic                             bool
	}{
		{
			name:                                    "All metrics enabled",
			sdcMetricsEnabled:                       "true",
			volumeMetricsEnabled:                    "true",
			storagePoolMetricsEnabled:               "true",
			powerflexTopologyMetricsEnabled:         "true",
			expectedSdcMetricsEnabled:               true,
			expectedVolumeMetricsEnabled:            true,
			expectedStoragePoolMetricsEnabled:       true,
			expectPanic:                             false,
			expectedPowerflexTopologyMetricsEnabled: true,
		},
		{
			name:                                    "All metrics disabled",
			sdcMetricsEnabled:                       "false",
			volumeMetricsEnabled:                    "false",
			storagePoolMetricsEnabled:               "false",
			powerflexTopologyMetricsEnabled:         "true",
			expectedPowerflexTopologyMetricsEnabled: true,
			expectedSdcMetricsEnabled:               false,
			expectedVolumeMetricsEnabled:            false,
			expectedStoragePoolMetricsEnabled:       false,
			expectPanic:                             false,
		},
		{
			name:                                    "sdcMetricsEnabled error",
			sdcMetricsEnabled:                       "test",
			volumeMetricsEnabled:                    "true",
			storagePoolMetricsEnabled:               "true",
			powerflexTopologyMetricsEnabled:         "true",
			expectedPowerflexTopologyMetricsEnabled: true,
			expectedSdcMetricsEnabled:               true,
			expectedVolumeMetricsEnabled:            true,
			expectedStoragePoolMetricsEnabled:       true,
			expectPanic:                             true,
		},
		{
			name:                                    "volumeMetricsEnabled error",
			sdcMetricsEnabled:                       "true",
			volumeMetricsEnabled:                    "test",
			storagePoolMetricsEnabled:               "true",
			powerflexTopologyMetricsEnabled:         "true",
			expectedPowerflexTopologyMetricsEnabled: true,
			expectedSdcMetricsEnabled:               true,
			expectedVolumeMetricsEnabled:            true,
			expectedStoragePoolMetricsEnabled:       true,
			expectPanic:                             true,
		},
		{
			name:                                    "storagePoolMetricsEnabled error",
			sdcMetricsEnabled:                       "true",
			volumeMetricsEnabled:                    "true",
			storagePoolMetricsEnabled:               "test",
			powerflexTopologyMetricsEnabled:         "true",
			expectedPowerflexTopologyMetricsEnabled: true,
			expectedSdcMetricsEnabled:               true,
			expectedVolumeMetricsEnabled:            true,
			expectedStoragePoolMetricsEnabled:       true,
			expectPanic:                             true,
		},
		{
			name:                                    "Topology metrics disabled",
			sdcMetricsEnabled:                       "true",
			volumeMetricsEnabled:                    "true",
			storagePoolMetricsEnabled:               "true",
			powerflexTopologyMetricsEnabled:         "false",
			expectedSdcMetricsEnabled:               true,
			expectedVolumeMetricsEnabled:            true,
			expectedStoragePoolMetricsEnabled:       true,
			expectedPowerflexTopologyMetricsEnabled: false,
			expectPanic:                             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("POWERFLEX_SDC_METRICS_ENABLED", tt.sdcMetricsEnabled)
			viper.Set("POWERFLEX_VOLUME_METRICS_ENABLED", tt.volumeMetricsEnabled)
			viper.Set("POWERFLEX_STORAGE_POOL_METRICS_ENABLED", tt.storagePoolMetricsEnabled)
			viper.Set("POWERFLEX_TOPOLOGY_METRICS_ENABLED", tt.powerflexTopologyMetricsEnabled)
			config := &entrypoint.Config{}
			if tt.expectPanic {
				want := "Invalid POWERFLEX_SDC_METRICS_ENABLED value. Valid values are true or false"
				if tt.name == "volumeMetricsEnabled error" {
					want = "Invalid POWERFLEX_VOLUME_METRICS_ENABLED value. Valid values are true or false"
				}
				if tt.name == "storagePoolMetricsEnabled error" {
					want = "Invalid POWERFLEX_STORAGE_POOL_METRICS_ENABLED value. Valid values are true or false"
				}
				assertFatalPath(t, "updateMetricsEnabled/"+tt.name, want)
			} else {
				assert.NotPanics(t, func() { updateMetricsEnabled(config) })
				assert.Equal(t, tt.expectedSdcMetricsEnabled, config.SDCMetricsEnabled, "SDC metrics enabled should be set correctly")
				assert.Equal(t, tt.expectedVolumeMetricsEnabled, config.VolumeMetricsEnabled, "Volume metrics enabled should be set correctly")
				assert.Equal(t, tt.expectedStoragePoolMetricsEnabled, config.SDCMetricsEnabled, "Storage metrics enabled should be set correctly")
				assert.Equal(t, tt.expectedPowerflexTopologyMetricsEnabled, config.TopologyMetricsEnabled, "Topology metrics enabled should be set correctly")
			}
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

			sdcFinder := &k8s.SDCFinder{
				StorageSystemID: []k8s.StorageSystemID{
					{
						ID: "system-id",
					},
				},
			}
			volumeFinder := &k8s.VolumeFinder{
				StorageSystemID: []k8s.StorageSystemID{
					{
						ID: "system-id",
					},
				},
			}
			storageClassFinder := &k8s.StorageClassFinder{
				StorageSystemID: []k8s.StorageSystemID{
					{
						ID: "system-id",
					},
				},
			}

			if tt.expectPanic {
				assertFatalPath(t, "updateProvisionerNames/"+tt.name, "PROVISIONER_NAMES is required")
			} else {
				assert.NotPanics(t, func() { updateProvisionerNames(sdcFinder, storageClassFinder, volumeFinder) })
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
		topologyMetricFreq  string
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
			topologyMetricFreq:  "5",
			expectedSdcIO:       30 * time.Second,
			expectedVolumeIO:    25 * time.Second,
			expectedStoragePool: 15 * time.Second,
			expectPanic:         false,
		},
		{
			name:                "Default Values When Empty",
			sdcIOFreq:           "",
			volumeIOFreq:        "",
			storagePoolFreq:     "",
			topologyMetricFreq:  "",
			expectedSdcIO:       defaultTickInterval,
			expectedVolumeIO:    defaultTickInterval,
			expectedStoragePool: defaultTickInterval,
			expectPanic:         false,
		},
		{
			name:                "Invalid SDC IO",
			sdcIOFreq:           "invalid",
			volumeIOFreq:        "25",
			storagePoolFreq:     "15",
			topologyMetricFreq:  "invalid",
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
			topologyMetricFreq:  "invalid",
			expectedSdcIO:       defaultTickInterval,
			expectedVolumeIO:    defaultTickInterval,
			expectedStoragePool: defaultTickInterval,
			expectPanic:         true,
		},
		{
			name:                "Invalid Storage Pool",
			sdcIOFreq:           "30",
			volumeIOFreq:        "10",
			storagePoolFreq:     "invalid",
			topologyMetricFreq:  "invalid",
			expectedSdcIO:       defaultTickInterval,
			expectedVolumeIO:    defaultTickInterval,
			expectedStoragePool: defaultTickInterval,
			expectPanic:         true,
		},
		{
			name:                "Negative SDC IO",
			sdcIOFreq:           "-1",
			volumeIOFreq:        "25",
			storagePoolFreq:     "15",
			topologyMetricFreq:  "invalid",
			expectedSdcIO:       defaultTickInterval,
			expectedVolumeIO:    defaultTickInterval,
			expectedStoragePool: defaultTickInterval,
			expectPanic:         true,
		},
		{
			name:                "Negative Volume IO",
			sdcIOFreq:           "30",
			volumeIOFreq:        "-1",
			storagePoolFreq:     "15",
			expectedSdcIO:       defaultTickInterval,
			expectedVolumeIO:    defaultTickInterval,
			expectedStoragePool: defaultTickInterval,
			expectPanic:         true,
		},
		{
			name:                "Negative Storage Pool",
			sdcIOFreq:           "30",
			volumeIOFreq:        "25",
			storagePoolFreq:     "-1",
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
			viper.Set("POWERFLEX_TOPOLOGY_METRICS_POLL_FREQUENCY", tt.topologyMetricFreq)

			config := &entrypoint.Config{}

			if tt.expectPanic {
				want := "Invalid POWERFLEX_SDC_IO_POLL_FREQUENCY. Specify a valid number"
				if tt.name == "Invalid Volume IO" {
					want = "Invalid POWERFLEX_VOLUME_IO_POLL_FREQUENCY. Specify a valid number"
				}
				if tt.name == "Invalid Storage Pool" {
					want = "Invalid POWERFLEX_STORAGE_POOL_POLL_FREQUENCY. Specify a valid number"
				}
				if tt.name == "Negative SDC IO" {
					want = "Invalid POWERFLEX_SDC_IO_POLL_FREQUENCY value. Must be greater than 0"
				}
				if tt.name == "Negative Volume IO" {
					want = "Invalid POWERFLEX_VOLUME_IO_POLL_FREQUENCY value. Must be greater than 0"
				}
				if tt.name == "Negative Storage Pool" {
					want = "Invalid POWERFLEX_STORAGE_POOL_POLL_FREQUENCY value. Must be greater than 0"
				}
				assertFatalPath(t, "updateTickIntervals/"+tt.name, want)
			} else {
				assert.NotPanics(t, func() { updateTickIntervals(config) })
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
			if tt.expectPanic {
				want := "POWERFLEX_MAX_CONCURRENT_QUERIES was not set to a valid number"
				if tt.name == "Null Value" {
					want = "POWERFLEX_MAX_CONCURRENT_QUERIES value was invalid (<= 0)"
				}
				assertFatalPath(t, "updateService/"+tt.name, want)
			} else {
				assert.NotPanics(t, func() { updateService(svc) })
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
		expectedLevel csmlog.Level
	}{
		{
			name:          "Valid Setting",
			logFormat:     "json",
			logLevel:      "INFO",
			expectedLevel: csmlog.InfoLevel,
		},
		{
			name:          "Invalid Setting",
			logFormat:     "json",
			logLevel:      "TEST",
			expectedLevel: csmlog.InfoLevel,
		},
		{
			name:          "text log format",
			logFormat:     "text",
			logLevel:      "INFO",
			expectedLevel: csmlog.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("LOG_FORMAT", tt.logFormat)
			viper.Set("LOG_LEVEL", tt.logLevel)
			setLoggingSettings()
			assert.Equal(t, tt.expectedLevel, csmlog.GetLevel())
		})
	}
}

func TestSetupConfigWatchers(t *testing.T) {
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
		API: &k8s.API{},
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
				setupConfigWatchers(configFileListener, powerflexSvc, config, sdcFinder, storageClassFinder, volumeFinder, exporter)
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
		configContentFile string
		expectPanic       bool
	}{
		{
			name:              "Config Reader Error",
			configContentFile: "testdata/not-exist.yaml",
			expectPanic:       true,
		},
		{
			name:              "Empty Endpoint Error",
			configContentFile: "testdata/invalid-endpoint-config.yaml",
			expectPanic:       true,
		},
		{
			name:              "Empty Password Error",
			configContentFile: "testdata/invalid-password-config.yaml",
			expectPanic:       true,
		},
		{
			name:              "Empty System ID Error",
			configContentFile: "testdata/invalid-systemid-config.yaml",
			expectPanic:       true,
		},
		{
			name:              "Empty Username Error",
			configContentFile: "testdata/invalid-username-config.yaml",
			expectPanic:       true,
		},
		{
			name:              "Authentication Error",
			configContentFile: "testdata/config.yaml",
			expectPanic:       true,
		},
		// Add more test cases here
	}

	// Iterate over the test table and run the test for each case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			if tt.expectPanic {
				want := "Failed to get storage system configuration"
				if tt.name == "Authentication Error" {
					want = "Failed to authenticate with PowerFlex"
				}
				assertFatalPath(t, "updatePowerFlexConnection/"+tt.name, want)
			}
		})
	}
}

func TestUpdatePowerFlexConnectionSuccess(t *testing.T) {
	// Create a fake PowerFlex API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login":
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `"fake-token"`)
		case "/api/version":
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `"4.0"`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Write a temp config file pointing to the test server
	tmpFile, err := os.CreateTemp("", "powerflex-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	configContent := fmt.Sprintf("- username: admin\n  password: password\n  systemID: test-system\n  endpoint: %s\n  insecure: true\n", server.URL)
	_, err = tmpFile.WriteString(configContent)
	assert.NoError(t, err)
	_ = tmpFile.Close()

	// Override goscaleioClient to use the real function
	origClient := goscaleioClient
	goscaleioClient = func(endpoint string, version string, timeout int64, insecure, useCerts bool, caFilePath string) (*goscaleio.Client, error) {
		return goscaleio.NewClientWithArgs(endpoint, version, timeout, insecure, useCerts, caFilePath)
	}
	defer func() { goscaleioClient = origClient }()

	viper.Reset()
	viper.Set("provisioner_names", "csi-vxflexos.dellemc.com")

	config := &entrypoint.Config{}
	sdcFinder := &k8s.SDCFinder{}
	storageClassFinder := &k8s.StorageClassFinder{}
	volumeFinder := &k8s.VolumeFinder{}

	assert.NotPanics(t, func() {
		updatePowerFlexConnection(
			tmpFile.Name(),
			config,
			sdcFinder,
			storageClassFinder,
			volumeFinder,
		)
	})

	assert.Contains(t, config.PowerFlexClient, "test-system")
	assert.Contains(t, config.PowerFlexConfig, "test-system")
	assert.Equal(t, 1, len(sdcFinder.StorageSystemID))
	assert.Equal(t, "test-system", sdcFinder.StorageSystemID[0].ID)
}

func TestUpdatePowerFlexConnectionClientError(t *testing.T) {
	// Write a temp config file with valid data
	tmpFile, err := os.CreateTemp("", "powerflex-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	configContent := "- username: admin\n  password: password\n  systemID: test-system\n  endpoint: http://127.0.0.1\n"
	_, err = tmpFile.WriteString(configContent)
	assert.NoError(t, err)
	_ = tmpFile.Close()

	// Override goscaleioClient to return an error
	origClient := goscaleioClient
	goscaleioClient = func(string, string, int64, bool, bool, string) (*goscaleio.Client, error) {
		return nil, fmt.Errorf("mock client creation error")
	}
	defer func() { goscaleioClient = origClient }()

	viper.Reset()
	viper.Set("provisioner_names", "csi-vxflexos.dellemc.com")

	config := &entrypoint.Config{}
	sdcFinder := &k8s.SDCFinder{}
	storageClassFinder := &k8s.StorageClassFinder{}
	volumeFinder := &k8s.VolumeFinder{}
	_ = config
	_ = sdcFinder
	_ = storageClassFinder
	_ = volumeFinder

	assertFatalPath(t, "updatePowerFlexConnection/Client Error", "Failed to create PowerFlex client")
}

func TestUpdateServiceDefault(t *testing.T) {
	viper.Reset()
	// Don't set POWERFLEX_MAX_CONCURRENT_QUERIES so the default is used
	svc := &service.PowerFlexService{}

	assert.NotPanics(t, func() { updateService(svc) })
	assert.Equal(t, service.DefaultMaxPowerFlexConnections, svc.MaxPowerFlexConnections)
}

func TestUpdatePowerFlexConnectionInsecureFlags(t *testing.T) {
	// Test the SkipCertificateValidation flag path
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login":
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `"fake-token"`)
		case "/api/version":
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `"4.0"`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tmpFile, err := os.CreateTemp("", "powerflex-config-*.yaml")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	configContent := fmt.Sprintf("- username: admin\n  password: password\n  systemID: test-system\n  endpoint: %s\n  skipCertificateValidation: true\n", server.URL)
	_, err = tmpFile.WriteString(configContent)
	assert.NoError(t, err)
	_ = tmpFile.Close()

	origClient := goscaleioClient
	goscaleioClient = func(endpoint string, version string, _ int64, insecure, useCerts bool, caFilePath string) (*goscaleio.Client, error) {
		return goscaleio.NewClientWithArgs(endpoint, version, math.MaxInt64, insecure, useCerts, caFilePath)
	}
	defer func() { goscaleioClient = origClient }()

	viper.Reset()
	viper.Set("provisioner_names", "csi-vxflexos.dellemc.com")

	config := &entrypoint.Config{}
	sdcFinder := &k8s.SDCFinder{}
	storageClassFinder := &k8s.StorageClassFinder{}
	volumeFinder := &k8s.VolumeFinder{}

	assert.NotPanics(t, func() {
		updatePowerFlexConnection(
			tmpFile.Name(),
			config,
			sdcFinder,
			storageClassFinder,
			volumeFinder,
		)
	})

	assert.Contains(t, config.PowerFlexClient, "test-system")
}

func TestShouldStartObservabilityMetricsServer(t *testing.T) {
	viper.Reset()
	t.Setenv("X_CSI_METRICS_ENABLED", "true")
	assert.True(t, shouldStartObservabilityMetricsServer())
}

func TestStartMetricsServer(t *testing.T) {
	t.Run("HTTP mode", func(t *testing.T) {
		viper.Reset()
		viper.Set("X_CSI_METRICS_PORT", "0")

		svc := &service.PowerFlexService{}
		assert.Nil(t, svc.ObsInstrumenter)

		startMetricsServer(svc)

		assert.NotNil(t, svc.ObsInstrumenter, "ObsInstrumenter must be wired after startMetricsServer")
	})

	t.Run("HTTPS mode with valid TLS files", func(t *testing.T) {
		certFile, err := os.CreateTemp("", "cert-*.pem")
		assert.NoError(t, err)
		defer func() { _ = os.Remove(certFile.Name()) }()
		_ = certFile.Close()

		keyFile, err := os.CreateTemp("", "key-*.pem")
		assert.NoError(t, err)
		defer func() { _ = os.Remove(keyFile.Name()) }()
		_ = keyFile.Close()

		viper.Reset()
		viper.Set("X_CSI_METRICS_PORT", "0")
		viper.Set("X_CSI_METRICS_TLS_CERT_FILE", certFile.Name())
		viper.Set("X_CSI_METRICS_TLS_KEY_FILE", keyFile.Name())

		svc := &service.PowerFlexService{}
		startMetricsServer(svc)

		assert.NotNil(t, svc.ObsInstrumenter, "ObsInstrumenter must be wired for HTTPS mode")
	})
}

func TestValidateTLSFiles(t *testing.T) {
	t.Run("valid files", func(t *testing.T) {
		certFile, err := os.CreateTemp("", "cert-*.pem")
		assert.NoError(t, err)
		defer func() { _ = os.Remove(certFile.Name()) }()
		_ = certFile.Close()

		keyFile, err := os.CreateTemp("", "key-*.pem")
		assert.NoError(t, err)
		defer func() { _ = os.Remove(keyFile.Name()) }()
		_ = keyFile.Close()

		err = validateTLSFiles(certFile.Name(), keyFile.Name())
		assert.NoError(t, err)
	})

	t.Run("missing cert file", func(t *testing.T) {
		keyFile, err := os.CreateTemp("", "key-*.pem")
		assert.NoError(t, err)
		defer func() { _ = os.Remove(keyFile.Name()) }()
		_ = keyFile.Close()

		err = validateTLSFiles("/nonexistent/cert.pem", keyFile.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot open TLS file")
	})

	t.Run("missing key file", func(t *testing.T) {
		certFile, err := os.CreateTemp("", "cert-*.pem")
		assert.NoError(t, err)
		defer func() { _ = os.Remove(certFile.Name()) }()
		_ = certFile.Close()

		err = validateTLSFiles(certFile.Name(), "/nonexistent/key.pem")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot open TLS file")
	})
}
