/*
 Copyright (c) 2025 Dell Inc. or its subsidiaries. All Rights Reserved.

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

package entrypoint

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/dell/csmlog"
	pflexServices "github.com/dell/karavi-metrics-powerflex/internal/service"
	otlexporters "github.com/dell/karavi-metrics-powerflex/opentelemetry/exporters"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"google.golang.org/grpc/credentials"

	sio "github.com/dell/goscaleio"
)

const (
	// MaximumTickInterval is the maximum allowed interval when querying metrics
	MaximumTickInterval = 10 * time.Minute
	// MinimumTickInterval is the minimum allowed interval when querying metrics
	MinimumTickInterval = 5 * time.Second
	// MaximumSDCTickInterval is the maximum allowed interval when querying SDC metrics
	MaximumSDCTickInterval = 10 * time.Minute
	// MinimumSDCTickInterval is the minimum allowed interval when querying SDC metrics
	MinimumSDCTickInterval = 5 * time.Second
	// MaximumVolTickInterval is the maximum allowed interval when querying volume metrics
	MaximumVolTickInterval = 10 * time.Minute
	// MinimumVolTickInterval is the minimum allowed interval when querying volume metrics
	MinimumVolTickInterval = 5 * time.Second
	// DefaultEndPoint for leader election path
	DefaultEndPoint = "karavi-metrics-powerflex" // #nosec G101
	// DefaultNameSpace for powerflex pod running metrics collection
	DefaultNameSpace = "karavi"
)

// ConfigValidatorFunc is used to override config validation in testing
var ConfigValidatorFunc = ValidateConfig

// Config holds data that will be used by the service
type Config struct {
	SDCTickInterval             time.Duration
	VolumeTickInterval          time.Duration
	StoragePoolTickInterval     time.Duration
	TopologyMetricsTickInterval time.Duration
	PowerFlexClient             map[string]pflexServices.PowerFlexClient
	PowerFlexConfig             map[string]sio.ConfigConnect
	SDCFinder                   pflexServices.SDCFinder
	StorageClassFinder          pflexServices.StorageClassFinder
	LeaderElector               pflexServices.LeaderElector
	VolumeFinder                pflexServices.VolumeFinder
	NodeFinder                  pflexServices.NodeFinder
	SDCMetricsEnabled           bool
	VolumeMetricsEnabled        bool
	StoragePoolMetricsEnabled   bool
	CollectorAddress            string
	CollectorCertPath           string
	TopologyMetricsEnabled      bool
}

var errOTELExportFailed = errors.New("OTEL export failed")

func snapshotPowerFlexSystemIDs(config map[string]sio.ConfigConnect) []string {
	systemIDs := make([]string, 0, len(config))
	for systemID := range config {
		systemIDs = append(systemIDs, systemID)
	}
	return systemIDs
}

// Run is the entry point for starting the service
func Run(ctx context.Context, config *Config, exporter otlexporters.Otlexporter, pflexSvc pflexServices.Service) error {
	err := ConfigValidatorFunc(config)
	if err != nil {
		return err
	}
	errCh := make(chan error, 1)
	go func() {
		powerflexEndpoint := os.Getenv("POWERFLEX_METRICS_ENDPOINT")
		if powerflexEndpoint == "" {
			powerflexEndpoint = DefaultEndPoint
		}
		powerflexNamespace := os.Getenv("POWERFLEX_METRICS_NAMESPACE")
		if powerflexNamespace == "" {
			powerflexNamespace = DefaultNameSpace
		}
		errCh <- config.LeaderElector.InitLeaderElection(powerflexEndpoint, powerflexNamespace)
	}()

	go func() {
		options := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(config.CollectorAddress),
		}

		if config.CollectorCertPath != "" {
			transportCreds, err := credentials.NewClientTLSFromFile(config.CollectorCertPath, "")
			if err != nil {
				errCh <- err
			}
			options = append(options, otlpmetricgrpc.WithTLSCredentials(transportCreds))
		} else {
			options = append(options, otlpmetricgrpc.WithInsecure())
		}

		// Set up export failure callback to track OTEL export failures
		if otlExporter, ok := exporter.(*otlexporters.OtlCollectorExporter); ok {
			systemIDs := snapshotPowerFlexSystemIDs(config.PowerFlexConfig)
			otlExporter.SetExportFailureRecorder(func() {
				// Record export failure in observability metrics
				for _, systemID := range systemIDs {
					pflexSvc.RecordObsMetrics(systemID, 0, true, errOTELExportFailed)
				}
			})
		}

		errCh <- exporter.InitExporter(options...)
	}()

	defer func() {
		if err := exporter.StopExporter(); err != nil {
			csmlog.WithContext(ctx).Errorf("Failed to stop exporter: %v", err)
		}
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// set initial tick intervals
	SDCTickInterval := config.SDCTickInterval
	VolumeTickInterval := config.VolumeTickInterval
	StoragePoolTickInterval := config.StoragePoolTickInterval
	sdcTicker := time.NewTicker(SDCTickInterval)
	volumeTicker := time.NewTicker(VolumeTickInterval)
	storagePoolTicker := time.NewTicker(StoragePoolTickInterval)
	TopologyMetricsTickInterval := config.TopologyMetricsTickInterval
	topologyMetricsTicker := time.NewTicker(TopologyMetricsTickInterval)
	for {
		select {
		case <-sdcTicker.C:
			if !config.LeaderElector.IsLeader() {
				csmlog.WithContext(ctx).Info("Not a leader. Only the leader pod can collect metrics")
				continue
			}
			if !config.SDCMetricsEnabled {
				csmlog.WithContext(ctx).Info("PowerFlex SDC metrics collection is disabled")
				continue
			}

			csmlog.WithContext(ctx).WithFields(csmlog.Fields{"powerflex_client_count": len(config.PowerFlexClient)}).Debug("Starting PowerFlex SDC metrics collection")

			for key, client := range config.PowerFlexClient {
				start := time.Now()
				csmlog.WithContext(ctx).WithFields(csmlog.Fields{"storage_system_id": key}).Debug("Collecting SDC metrics for storage system")
				sioConfig, ok := config.PowerFlexConfig[key]
				if !ok {
					csmlog.WithContext(ctx).WithFields(csmlog.Fields{"storage_system_id": key}).Error("No configuration found for storage system")
					pflexSvc.RecordObsMetrics(key, time.Since(start), false, nil)
					continue
				}

				sdcs, err := pflexSvc.GetSDCs(ctx, client, config.SDCFinder)
				if err != nil {
					csmlog.WithContext(ctx).WithFields(csmlog.Fields{"error": err, "endpoint": sioConfig.Endpoint}).Error("Failed to get SDCs")
					pflexSvc.RecordObsMetrics(key, time.Since(start), false, nil)
					continue
				}

				nodes, err := config.NodeFinder.GetNodes()
				if err != nil {
					csmlog.WithContext(ctx).WithFields(csmlog.Fields{"error": err}).Error("Failed to get Kubernetes nodes")
					pflexSvc.RecordObsMetrics(key, time.Since(start), false, nil)
					continue
				}

				pflexSvc.GetSDCStatistics(ctx, nodes, sdcs)
				pflexSvc.RecordObsMetrics(key, time.Since(start), true, nil)
			}

		case <-volumeTicker.C:
			if !config.LeaderElector.IsLeader() {
				csmlog.WithContext(ctx).Info("Not a leader. Only the leader pod can collect metrics")
				continue
			}
			if !config.VolumeMetricsEnabled {
				csmlog.WithContext(ctx).Info("PowerFlex volume metrics collection is disabled")
				continue
			}

			csmlog.WithContext(ctx).WithFields(csmlog.Fields{"powerflex_client_count": len(config.PowerFlexClient)}).Debug("Starting PowerFlex volume metrics collection")

			for key, client := range config.PowerFlexClient {
				start := time.Now()
				csmlog.WithContext(ctx).WithFields(csmlog.Fields{"storage_system_id": key}).Debug("Collecting volume metrics for storage system")
				sioConfig, ok := config.PowerFlexConfig[key]
				if !ok {
					csmlog.WithContext(ctx).WithFields(csmlog.Fields{"storage_system_id": key}).Error("No configuration found for storage system")
					pflexSvc.RecordObsMetrics(key, time.Since(start), false, nil)
					continue
				}
				sdcs, err := pflexSvc.GetSDCs(ctx, client, config.SDCFinder)
				if err != nil {
					csmlog.WithContext(ctx).WithFields(csmlog.Fields{"error": err, "endpoint": sioConfig.Endpoint}).Error("Failed to get SDCs")
					pflexSvc.RecordObsMetrics(key, time.Since(start), false, nil)
					continue
				}

				volumes, err := pflexSvc.GetVolumes(ctx, client, sdcs)
				if err != nil {
					csmlog.WithContext(ctx).WithFields(csmlog.Fields{"error": err}).Error("Failed to get volumes")
					pflexSvc.RecordObsMetrics(key, time.Since(start), false, nil)
					continue
				}
				exportErr := pflexSvc.ExportVolumeStatistics(ctx, volumes, config.VolumeFinder)
				pflexSvc.RecordObsMetrics(key, time.Since(start), true, exportErr)
			}

		case <-storagePoolTicker.C:
			if !config.LeaderElector.IsLeader() {
				csmlog.WithContext(ctx).Info("Not a leader. Only the leader pod can collect metrics")
				continue
			}
			if !config.StoragePoolMetricsEnabled {
				csmlog.WithContext(ctx).Info("PowerFlex storage pool metrics collection is disabled")
				continue
			}

			csmlog.WithContext(ctx).WithFields(csmlog.Fields{"powerflex_client_count": len(config.PowerFlexClient)}).Debug("Starting PowerFlex storage pool metrics collection")

			for key, client := range config.PowerFlexClient {
				start := time.Now()
				csmlog.WithContext(ctx).WithFields(csmlog.Fields{"storage_system_id": key}).Debug("Collecting storage pool metrics for storage system")

				sioConfig, ok := config.PowerFlexConfig[key]
				if !ok {
					csmlog.WithContext(ctx).WithFields(csmlog.Fields{"storage_system_id": key}).Error("No configuration found for storage system")
					pflexSvc.RecordObsMetrics(key, time.Since(start), false, nil)
					continue
				}

				storageClassMetas, err := pflexSvc.GetStorageClasses(ctx, client, config.StorageClassFinder)
				if err != nil {
					csmlog.WithContext(ctx).WithFields(csmlog.Fields{"error": err, "endpoint": sioConfig.Endpoint}).Error("Failed to get storage class and storage pool information")
					pflexSvc.RecordObsMetrics(key, time.Since(start), false, nil)
					continue
				}

				csmlog.WithContext(ctx).WithFields(csmlog.Fields{"storage_class_meta_count": len(storageClassMetas)}).Debug("Resolved storage classes for storage pool metrics collection")
				pflexSvc.GetStoragePoolStatistics(ctx, storageClassMetas)
				pflexSvc.RecordObsMetrics(key, time.Since(start), true, nil)
			}

		case <-topologyMetricsTicker.C:
			if !config.LeaderElector.IsLeader() {
				csmlog.WithContext(ctx).Info("Not a leader. Only the leader pod can collect metrics")
				continue
			}
			if !config.TopologyMetricsEnabled {
				csmlog.WithContext(ctx).Info("PowerFlex topology metrics collection is disabled")
				continue
			}
			start := time.Now()
			pflexSvc.ExportTopologyMetrics(ctx)
			for key := range config.PowerFlexClient {
				pflexSvc.RecordObsMetrics(key, time.Since(start), true, nil)
			}

		case err := <-errCh:
			if err == nil {
				continue
			}
			return err
		case <-ctx.Done():
			return nil
		}

		// check if tick interval config settings have changed
		if SDCTickInterval != config.SDCTickInterval {
			SDCTickInterval = config.SDCTickInterval
			sdcTicker = time.NewTicker(SDCTickInterval)
		}
		if VolumeTickInterval != config.VolumeTickInterval {
			VolumeTickInterval = config.VolumeTickInterval
			volumeTicker = time.NewTicker(VolumeTickInterval)
		}
		if StoragePoolTickInterval != config.StoragePoolTickInterval {
			StoragePoolTickInterval = config.StoragePoolTickInterval
			storagePoolTicker = time.NewTicker(StoragePoolTickInterval)
		}
		if TopologyMetricsTickInterval != config.TopologyMetricsTickInterval {
			TopologyMetricsTickInterval = config.TopologyMetricsTickInterval
			topologyMetricsTicker = time.NewTicker(TopologyMetricsTickInterval)
		}
	}
}

// ValidateConfig will validate the configuration and return any errors
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("no config provided")
	}

	if config.PowerFlexClient == nil {
		return fmt.Errorf("no PowerFlexClient provided in config")
	}

	if config.SDCFinder == nil {
		return fmt.Errorf("no SDCFinder provided in config")
	}

	if config.NodeFinder == nil {
		return fmt.Errorf("no NodeFinder provided in config")
	}

	if config.SDCTickInterval > MaximumSDCTickInterval || config.SDCTickInterval < MinimumSDCTickInterval {
		return fmt.Errorf("SDC polling frequency not within allowed range of %v and %v", MinimumSDCTickInterval.String(), MaximumSDCTickInterval.String())
	}

	if config.VolumeTickInterval > MaximumVolTickInterval || config.VolumeTickInterval < MinimumVolTickInterval {
		return fmt.Errorf("volume polling frequency not within allowed range of %v and %v", MinimumVolTickInterval.String(), MaximumVolTickInterval.String())
	}

	if config.TopologyMetricsTickInterval > MaximumTickInterval || config.TopologyMetricsTickInterval < MinimumTickInterval {
		return fmt.Errorf("topology metrics polling frequency not within allowed range of %v and %v", MinimumTickInterval.String(), MaximumTickInterval.String())
	}
	return nil
}
