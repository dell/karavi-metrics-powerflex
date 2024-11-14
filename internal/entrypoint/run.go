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

package entrypoint

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	pflexServices "github.com/dell/karavi-metrics-powerflex/internal/service"
	otlexporters "github.com/dell/karavi-metrics-powerflex/opentelemetry/exporters"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"google.golang.org/grpc/credentials"
)

const (
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

// PowerflexClient wraps a real powerflex client and an authentication token getter
type PowerflexClient struct {
	Client      pflexServices.PowerFlexClient
	TokenGetter pflexServices.TokenGetter
}

// Config holds data that will be used by the service
type Config struct {
	SDCTickInterval           time.Duration
	VolumeTickInterval        time.Duration
	StoragePoolTickInterval   time.Duration
	PowerFlexClient           map[string]*PowerflexClient
	SDCFinder                 pflexServices.SDCFinder
	StorageClassFinder        pflexServices.StorageClassFinder
	LeaderElector             pflexServices.LeaderElector
	VolumeFinder              pflexServices.VolumeFinder
	NodeFinder                pflexServices.NodeFinder
	SDCMetricsEnabled         bool
	VolumeMetricsEnabled      bool
	StoragePoolMetricsEnabled bool
	CollectorAddress          string
	CollectorCertPath         string
	Logger                    *logrus.Logger
	Lock                      *sync.Mutex
}

// Run is the entry point for starting the service
func Run(ctx context.Context, config *Config, exporter otlexporters.Otlexporter, pflexSvc pflexServices.Service) error {
	err := ConfigValidatorFunc(config)
	if err != nil {
		return err
	}
	logger := config.Logger

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

		errCh <- exporter.InitExporter(options...)
	}()

	defer exporter.StopExporter()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// set initial tick intervals
	SDCTickInterval := config.SDCTickInterval
	VolumeTickInterval := config.VolumeTickInterval
	StoragePoolTickInterval := config.StoragePoolTickInterval
	sdcTicker := time.NewTicker(SDCTickInterval)
	volumeTicker := time.NewTicker(VolumeTickInterval)
	storagePoolTicker := time.NewTicker(StoragePoolTickInterval)
	for {
		select {
		case <-sdcTicker.C:
			if !config.LeaderElector.IsLeader() {
				logger.Info("not leader pod to collect metrics")
				continue
			}
			if !config.SDCMetricsEnabled {
				logger.Info("powerflex SDC metrics collection is disabled")
				continue
			}

			logger.WithField("number of PowerFlexClient", len(config.PowerFlexClient)).Debug("PowerFlexClient")

			config.Lock.Lock()
			for key, sioConfig := range config.PowerFlexClient {
				logger.WithField("storage system id", key).Debug("storage system id")

				endpoint := sioConfig.Client.GetConfigConnect().Endpoint

				token, err := sioConfig.TokenGetter.GetToken(ctx)
				if err != nil {
					logger.WithError(err).Error("getting authentication token")
					continue
				}
				sioConfig.Client.SetToken(token)

				sdcs, err := pflexSvc.GetSDCs(ctx, sioConfig.Client, config.SDCFinder)
				if err != nil {
					logger.WithError(err).WithField("endpoint", endpoint).Error("getting SDCs")
					continue
				}

				nodes, err := config.NodeFinder.GetNodes()
				if err != nil {
					logger.WithError(err).Error("getting kubernetes nodes")
					continue
				}

				pflexSvc.GetSDCStatistics(ctx, nodes, sdcs)
			}
			config.Lock.Unlock()

		case <-volumeTicker.C:
			if !config.LeaderElector.IsLeader() {
				logger.Info("not leader pod to collect metrics")
				continue
			}
			if !config.VolumeMetricsEnabled {
				logger.Info("powerflex volume metrics collection is disabled")
				continue
			}

			logger.WithField("number of PowerFlexClient", len(config.PowerFlexClient)).Debug("PowerFlexClient")

			config.Lock.Lock()
			for key, sioConfig := range config.PowerFlexClient {
				logger.WithField("storage system id", key).Debug("storage system id")

				endpoint := sioConfig.Client.GetConfigConnect().Endpoint

				token, err := sioConfig.TokenGetter.GetToken(ctx)
				if err != nil {
					logger.WithError(err).Error("getting authentication token")
					continue
				}
				sioConfig.Client.SetToken(token)

				sdcs, err := pflexSvc.GetSDCs(ctx, sioConfig.Client, config.SDCFinder)
				if err != nil {
					logger.WithError(err).WithField("endpoint", endpoint).Error("getting SDCs")
					continue
				}

				volumes, err := pflexSvc.GetVolumes(ctx, sdcs)
				if err != nil {
					logger.WithError(err).Error("getting volumes")
					continue
				}
				pflexSvc.ExportVolumeStatistics(ctx, volumes, config.VolumeFinder)
			}
			config.Lock.Unlock()

		case <-storagePoolTicker.C:
			if !config.LeaderElector.IsLeader() {
				logger.Info("not leader pod to collect metrics")
				continue
			}
			if !config.StoragePoolMetricsEnabled {
				logger.Info("powerflex storage pool metrics collection is disabled")
				continue
			}

			logger.WithField("number of PowerFlexClient", len(config.PowerFlexClient)).Debug("PowerFlexClient")

			config.Lock.Lock()
			for key, sioConfig := range config.PowerFlexClient {
				logger.WithField("storage system id", key).Debug("storage system id")

				endpoint := sioConfig.Client.GetConfigConnect().Endpoint

				token, err := sioConfig.TokenGetter.GetToken(ctx)
				if err != nil {
					logger.WithError(err).Error("getting authentication token")
					continue
				}
				sioConfig.Client.SetToken(token)

				storageClassMetas, err := pflexSvc.GetStorageClasses(ctx, sioConfig.Client, config.StorageClassFinder)
				if err != nil {
					logger.WithError(err).WithField("endpoint", endpoint).Error("getting storage class and storage pool information")
					continue
				}

				logger.WithField("storageClassMetas", storageClassMetas).Debug("storageClassMetas")
				pflexSvc.GetStoragePoolStatistics(ctx, storageClassMetas)
			}
			config.Lock.Unlock()

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
		return fmt.Errorf("Volume polling frequency not within allowed range of %v and %v", MinimumVolTickInterval.String(), MaximumVolTickInterval.String())
	}

	return nil
}
