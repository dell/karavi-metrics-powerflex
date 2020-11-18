package entrypoint

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
	"log"
	"os"
	"runtime"
	"time"

	"github.com/dell/karavi-metrics-powerflex/internal/service"
	pflexServices "github.com/dell/karavi-metrics-powerflex/internal/service"
	otlexporters "github.com/dell/karavi-metrics-powerflex/opentelemetry/exporters"
	"go.opentelemetry.io/otel/exporters/otlp"
	"google.golang.org/grpc/credentials"

	sio "github.com/dell/goscaleio"
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
	DefaultEndPoint = "karavi-metrics-powerflex"
	// DefaultNameSpace for powerflex pod running metrics collection
	DefaultNameSpace = "karavi"
)

var (
	// ConfigValidatorFunc is used to override config validation in testing
	ConfigValidatorFunc func(*Config) error = ValidateConfig
)

// Config holds data that will be used by the service
type Config struct {
	SDCTickInterval           time.Duration
	VolumeTickInterval        time.Duration
	StoragePoolTickInterval   time.Duration
	PowerFlexClient           pflexServices.PowerFlexClient
	PowerFlexConfig           sio.ConfigConnect
	SDCFinder                 service.SDCFinder
	StorageClassFinder        service.StorageClassFinder
	LeaderElector             service.LeaderElector
	VolumeFinder              service.VolumeFinder
	NodeFinder                service.NodeFinder
	SDCMetricsEnabled         bool
	VolumeMetricsEnabled      bool
	StoragePoolMetricsEnabled bool
	CollectorAddress          string
	CollectorCertPath         string
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
		options := []otlp.ExporterOption{
			otlp.WithAddress(config.CollectorAddress),
		}

		if config.CollectorCertPath != "" {
			transportCreds, err := credentials.NewClientTLSFromFile(config.CollectorCertPath, "")
			if err != nil {
				errCh <- err
			}
			options = append(options, otlp.WithTLSCredentials(transportCreds))
		} else {
			options = append(options, otlp.WithInsecure())
		}

		errCh <- exporter.InitExporter(options...)
	}()

	defer exporter.StopExporter()

	runtime.GOMAXPROCS(runtime.NumCPU())

	sdcTicker := time.NewTicker(config.SDCTickInterval)
	volumeTicker := time.NewTicker(config.VolumeTickInterval)
	storagePoolTicker := time.NewTicker(config.StoragePoolTickInterval)
	for {
		select {
		case <-sdcTicker.C:
			if !config.LeaderElector.IsLeader() {
				log.Printf("Not leader pod to collect metrics")
				continue
			}
			if !config.SDCMetricsEnabled {
				log.Printf("PowerFlex SDC metrics collection is disabled")
				continue
			}
			_, err := config.PowerFlexClient.Authenticate(&config.PowerFlexConfig)
			if err != nil {
				log.Printf("Failed to authenticate with PowerFlex: %v. Retrying on next tick...", err)
				continue
			}
			sdcs, err := pflexSvc.GetSDCs(ctx, config.PowerFlexClient, config.SDCFinder)
			if err != nil {
				log.Printf("error getting SDCs: %v", err)
				continue
			}

			nodes, err := config.NodeFinder.GetNodes()
			if err != nil {
				log.Printf("error getting Kubernetes nodes: %v", err)
				continue
			}

			pflexSvc.GetSDCStatistics(ctx, nodes, sdcs)

		case <-volumeTicker.C:
			if !config.LeaderElector.IsLeader() {
				log.Printf("Not leader pod to collect metrics")
				continue
			}
			if !config.VolumeMetricsEnabled {
				log.Printf("PowerFlex SDC metrics collection is disabled")
				continue
			}
			_, err := config.PowerFlexClient.Authenticate(&config.PowerFlexConfig)
			if err != nil {
				log.Printf("Failed to authenticate with PowerFlex: %v. Retrying on next tick...", err)
				continue
			}
			sdcs, err := pflexSvc.GetSDCs(ctx, config.PowerFlexClient, config.SDCFinder)
			if err != nil {
				log.Printf("error getting SDCs: %v", err)
				continue
			}

			volumes, err := pflexSvc.GetVolumes(ctx, sdcs)
			if err != nil {
				log.Printf("error getting Volumes: %v", err)
				continue
			}
			pflexSvc.ExportVolumeStatistics(ctx, volumes, config.VolumeFinder)

		case <-storagePoolTicker.C:
			if !config.LeaderElector.IsLeader() {
				log.Printf("Not leader pod to collect metrics")
				continue
			}
			if !config.StoragePoolMetricsEnabled {
				log.Printf("PowerFlex Storage Pool metrics collection is disabled")
				continue
			}

			_, err := config.PowerFlexClient.Authenticate(&config.PowerFlexConfig)
			if err != nil {
				log.Printf("Failed to authenticate with PowerFlex: %v. Retrying on next tick...", err)
				continue
			}

			storageClassMetas, err := pflexSvc.GetStorageClasses(ctx, config.PowerFlexClient, config.StorageClassFinder)
			if err != nil {
				log.Printf("error getting storage class and storage pool information: %v", err)
				continue
			}

			pflexSvc.GetStoragePoolStatistics(ctx, storageClassMetas)

		case err := <-errCh:
			if err == nil {
				continue
			}
			return err
		case <-ctx.Done():
			return nil
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
