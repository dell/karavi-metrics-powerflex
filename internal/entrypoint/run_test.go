package entrypoint_test

// Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/dell/karavi-powerflex-metrics/internal/entrypoint"
	pflexServices "github.com/dell/karavi-powerflex-metrics/internal/service"
	"github.com/dell/karavi-powerflex-metrics/internal/service/mocks"
	metrics "github.com/dell/karavi-powerflex-metrics/internal/service/mocks"
	otlexporters "github.com/dell/karavi-powerflex-metrics/opentelemetry/exporters"
	exportermocks "github.com/dell/karavi-powerflex-metrics/opentelemetry/exporters/mocks"

	sio "github.com/dell/goscaleio"
	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Run(t *testing.T) {
	validSDCTickInterval := entrypoint.MinimumSDCTickInterval
	validVolumeTickInterval := entrypoint.MinimumSDCTickInterval

	tests := map[string]func(t *testing.T) (expectError bool, config *entrypoint.Config, exporter otlexporters.Otlexporter, pflexSvc pflexServices.Service, prevConfigValidationFunc func(*entrypoint.Config) error, ctrl *gomock.Controller, validatingConfig bool){

		"success": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			pfClient.EXPECT().Authenticate(gomock.Any()).AnyTimes().Return(sio.Cluster{}, nil)

			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			nodeFinder := mocks.NewMockNodeFinder(ctrl)
			nodeFinder.EXPECT().GetNodes().AnyTimes().
				Return([]corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Address: "1.2.3.6",
								},
							},
						},
					},
				}, nil)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:   pfClient,
				PowerFlexConfig:   sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:         sdcFinder,
				NodeFinder:        nodeFinder,
				LeaderElector:     leaderElector,
				SDCMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.StatisticsGetter{},
				nil,
			)
			svc.EXPECT().GetSDCStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},

		"success even if error during call to GetSDCs": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			pfClient.EXPECT().Authenticate(gomock.Any()).AnyTimes().Return(sio.Cluster{}, nil)

			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			nodeFinder := mocks.NewMockNodeFinder(ctrl)
			nodeFinder.EXPECT().GetNodes().AnyTimes().
				Return([]corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Address: "1.2.3.6",
								},
							},
						},
					},
				}, nil)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:   pfClient,
				PowerFlexConfig:   sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:         sdcFinder,
				NodeFinder:        nodeFinder,
				LeaderElector:     leaderElector,
				SDCMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				nil,
				errors.New("error"),
			)
			svc.EXPECT().GetSDCStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success even if error during call to NodeFinder": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			pfClient.EXPECT().Authenticate(gomock.Any()).AnyTimes().Return(sio.Cluster{}, nil)

			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			nodeFinder := mocks.NewMockNodeFinder(ctrl)
			nodeFinder.EXPECT().GetNodes().AnyTimes().
				Return([]corev1.Node{}, errors.New("error"))

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:   pfClient,
				PowerFlexConfig:   sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:         sdcFinder,
				NodeFinder:        nodeFinder,
				LeaderElector:     leaderElector,
				SDCMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.StatisticsGetter{},
				nil,
			)
			svc.EXPECT().GetSDCStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success even if SDC metrics collection is disabled": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			// authenticate should not be called because SDC metrics collection is disabled
			pfClient.EXPECT().Authenticate(gomock.Any()).Times(0).Return(sio.Cluster{}, nil)

			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			// GetSDCGuids should not be called because SDC metrics collection is disabled
			sdcFinder.EXPECT().GetSDCGuids().Times(0).Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			nodeFinder := mocks.NewMockNodeFinder(ctrl)
			nodeFinder.EXPECT().GetNodes().AnyTimes().
				Return([]corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Address: "1.2.3.6",
								},
							},
						},
					},
				}, nil)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:   pfClient,
				PowerFlexConfig:   sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:         sdcFinder,
				NodeFinder:        nodeFinder,
				LeaderElector:     leaderElector,
				SDCMetricsEnabled: false,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			// GetSDCs should not be called because SDC metrics collection is disabled
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).Times(0).Return(
				nil,
				errors.New("error"),
			)
			svc.EXPECT().GetSDCStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success even if PowerFlex authentication fails": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			pfClient.EXPECT().Authenticate(gomock.Any()).Return(sio.Cluster{}, fmt.Errorf("An error occurred while authenticating with the PowerFlex server")).AnyTimes()
			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			nodeFinder := mocks.NewMockNodeFinder(ctrl)
			nodeFinder.EXPECT().GetNodes().AnyTimes().
				Return([]corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Address: "1.2.3.6",
								},
							},
						},
					},
				}, nil)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:   pfClient,
				PowerFlexConfig:   sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:         sdcFinder,
				NodeFinder:        nodeFinder,
				LeaderElector:     leaderElector,
				SDCMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.StatisticsGetter{},
				nil,
			)

			svc.EXPECT().GetSDCStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"error no PowerFlex client": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			nodeFinder := mocks.NewMockNodeFinder(ctrl)
			leaderElector := mocks.NewMockLeaderElector(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:    nil,
				PowerFlexConfig:    sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:          sdcFinder,
				NodeFinder:         nodeFinder,
				LeaderElector:      leaderElector,
				SDCMetricsEnabled:  true,
				SDCTickInterval:    validSDCTickInterval,
				VolumeTickInterval: validVolumeTickInterval,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error no SDC Finder": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			leaderElector := mocks.NewMockLeaderElector(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:    pfClient,
				PowerFlexConfig:    sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:          nil,
				LeaderElector:      leaderElector,
				SDCMetricsEnabled:  true,
				SDCTickInterval:    validSDCTickInterval,
				VolumeTickInterval: validVolumeTickInterval,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error no Node Finder": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			leaderElector := mocks.NewMockLeaderElector(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:    pfClient,
				PowerFlexConfig:    sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:          sdcFinder,
				NodeFinder:         nil,
				LeaderElector:      leaderElector,
				SDCMetricsEnabled:  true,
				SDCTickInterval:    validSDCTickInterval,
				VolumeTickInterval: validVolumeTickInterval,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error invalid SDC poll time": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			nodeFinder := mocks.NewMockNodeFinder(ctrl)
			leaderElector := mocks.NewMockLeaderElector(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:    pfClient,
				PowerFlexConfig:    sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:          sdcFinder,
				NodeFinder:         nodeFinder,
				LeaderElector:      leaderElector,
				SDCMetricsEnabled:  true,
				SDCTickInterval:    entrypoint.MinimumSDCTickInterval - time.Second,
				VolumeTickInterval: validVolumeTickInterval,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error invalid Volume poll time (too low)": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			nodeFinder := mocks.NewMockNodeFinder(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:    pfClient,
				PowerFlexConfig:    sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:          sdcFinder,
				NodeFinder:         nodeFinder,
				SDCMetricsEnabled:  true,
				SDCTickInterval:    validSDCTickInterval,
				VolumeTickInterval: entrypoint.MinimumVolTickInterval - time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error invalid Volume poll time (too high)": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			nodeFinder := mocks.NewMockNodeFinder(ctrl)
			leaderElector := mocks.NewMockLeaderElector(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:    pfClient,
				PowerFlexConfig:    sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:          sdcFinder,
				NodeFinder:         nodeFinder,
				LeaderElector:      leaderElector,
				SDCMetricsEnabled:  true,
				SDCTickInterval:    validSDCTickInterval,
				VolumeTickInterval: entrypoint.MaximumVolTickInterval + time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error nil config": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			e := exportermocks.NewMockOtlexporter(ctrl)

			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			svc := metrics.NewMockService(ctrl)

			return true, nil, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error initializing exporter": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			sdcFinder := mocks.NewMockSDCFinder(ctrl)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:   pfClient,
				PowerFlexConfig:   sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:         sdcFinder,
				LeaderElector:     leaderElector,
				SDCMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(fmt.Errorf("An error occurred while initializing the exporter"))
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success for volume metrics": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			pfClient.EXPECT().Authenticate(gomock.Any()).AnyTimes().Return(sio.Cluster{}, nil)

			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:      pfClient,
				PowerFlexConfig:      sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:            sdcFinder,
				LeaderElector:        leaderElector,
				VolumeMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.StatisticsGetter{},
				nil,
			)
			svc.EXPECT().GetVolumes(gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.VolumeStatisticsGetter{},
				nil,
			)
			svc.EXPECT().ExportVolumeStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success even with failed Powerflex authentication": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			pfClient.EXPECT().Authenticate(gomock.Any()).Return(sio.Cluster{}, fmt.Errorf("An error occurred while authenticating with the PowerFlex server")).AnyTimes()
			sdcFinder := mocks.NewMockSDCFinder(ctrl)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:      pfClient,
				PowerFlexConfig:      sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:            sdcFinder,
				LeaderElector:        leaderElector,
				VolumeMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"error getting volumes": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			pfClient.EXPECT().Authenticate(gomock.Any()).AnyTimes().Return(sio.Cluster{}, nil)

			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:      pfClient,
				PowerFlexConfig:      sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:            sdcFinder,
				LeaderElector:        leaderElector,
				VolumeMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.StatisticsGetter{},
				nil,
			)
			svc.EXPECT().GetVolumes(gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.VolumeStatisticsGetter{},
				errors.New("error"),
			)

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"volume success even if error during call to GetSDCs": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			pfClient.EXPECT().Authenticate(gomock.Any()).AnyTimes().Return(sio.Cluster{}, nil)

			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:      pfClient,
				PowerFlexConfig:      sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:            sdcFinder,
				LeaderElector:        leaderElector,
				VolumeMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				nil,
				errors.New("error"),
			)
			svc.EXPECT().GetVolumes(gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.VolumeStatisticsGetter{},
				errors.New("error"),
			)
			svc.EXPECT().ExportVolumeStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success for storage class/pool": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			pfClient.EXPECT().Authenticate(gomock.Any()).AnyTimes().Return(sio.Cluster{}, nil)

			sc1 := v1.StorageClass{}
			sc1.Provisioner = "csi-vxflexos.dellemc.com"
			sc1.ObjectMeta = metav1.ObjectMeta{
				UID:  "123",
				Name: "class-1",
			}
			sc1.Parameters = map[string]string{
				"storagepool": "pool-1",
			}

			storageClassFinder := mocks.NewMockStorageClassFinder(ctrl)
			storageClassFinder.EXPECT().GetStorageClasses().AnyTimes().
				Return([]v1.StorageClass{sc1}, nil)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:           pfClient,
				PowerFlexConfig:           sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				StorageClassFinder:        storageClassFinder,
				LeaderElector:             leaderElector,
				StoragePoolMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			svc.EXPECT().GetStorageClasses(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]pflexServices.StorageClassMeta{
					{
						ID:           "123",
						Name:         "class-1",
						Driver:       "csi-vxflexos.dellemc.com",
						StoragePools: map[string]pflexServices.StoragePoolStatisticsGetter{},
					},
				}, nil).AnyTimes()

			svc.EXPECT().GetStoragePoolStatistics(gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"error no LeaderElector": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			sdcFinder := mocks.NewMockSDCFinder(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:    pfClient,
				PowerFlexConfig:    sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:          sdcFinder,
				LeaderElector:      nil,
				SDCMetricsEnabled:  true,
				SDCTickInterval:    validSDCTickInterval,
				VolumeTickInterval: validVolumeTickInterval,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success even if is leader is false": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			// authenticate should not be called because SDC metrics collection is disabled
			pfClient.EXPECT().Authenticate(gomock.Any()).Times(0).Return(sio.Cluster{}, nil)

			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			// GetSDCGuids should not be called because SDC metrics collection is disabled
			sdcFinder.EXPECT().GetSDCGuids().Times(0).Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(false)

			config := &entrypoint.Config{
				PowerFlexClient:   pfClient,
				PowerFlexConfig:   sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:         sdcFinder,
				LeaderElector:     leaderElector,
				SDCMetricsEnabled: false,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			// GetSDCs should not be called because SDC metrics collection is disabled
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).Times(0).Return(
				nil,
				errors.New("error"),
			)
			svc.EXPECT().GetSDCStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success for storage class/pool with GetStorageClasses err": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			pfClient.EXPECT().Authenticate(gomock.Any()).AnyTimes().Return(sio.Cluster{}, nil)

			sc1 := v1.StorageClass{}
			sc1.Provisioner = "csi-vxflexos.dellemc.com"
			sc1.ObjectMeta = metav1.ObjectMeta{
				UID:  "123",
				Name: "class-1",
			}
			sc1.Parameters = map[string]string{
				"storagepool": "pool-1",
			}

			storageClassFinder := mocks.NewMockStorageClassFinder(ctrl)
			storageClassFinder.EXPECT().GetStorageClasses().AnyTimes().
				Return([]v1.StorageClass{sc1}, nil)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:           pfClient,
				PowerFlexConfig:           sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				StorageClassFinder:        storageClassFinder,
				LeaderElector:             leaderElector,
				StoragePoolMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			svc.EXPECT().GetStorageClasses(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, fmt.Errorf("there was error getting the StorageClass")).AnyTimes()

			svc.EXPECT().GetStoragePoolStatistics(gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success even if PowerFlex authentication fails for StoragePoolMetrics": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metrics.NewMockPowerFlexClient(ctrl)
			pfClient.EXPECT().Authenticate(gomock.Any()).Return(sio.Cluster{}, fmt.Errorf("An error occurred while authenticating with the PowerFlex server")).AnyTimes()
			sdcFinder := mocks.NewMockSDCFinder(ctrl)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-powerflex-metrics", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:           pfClient,
				PowerFlexConfig:           sio.ConfigConnect{Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"},
				SDCFinder:                 sdcFinder,
				LeaderElector:             leaderElector,
				StoragePoolMetricsEnabled: true,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			// GetStorageClasses should not be called if authentication fails
			svc.EXPECT().GetStorageClasses(gomock.Any(), gomock.Any(), gomock.Any()).Times(0).
				Return(nil, fmt.Errorf("there was error getting the StorageClass"))

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expectError, config, exporter, svc, prevConfValidation, ctrl, validateConfig := test(t)
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			if config != nil && !validateConfig {
				// The configuration is not nil and the test is not attempting to validate the configuration.
				// In this case, we can use smaller intervals for testing purposes.
				config.SDCTickInterval = 100 * time.Millisecond
				config.VolumeTickInterval = 100 * time.Millisecond
				config.StoragePoolTickInterval = 100 * time.Millisecond
			}
			err := entrypoint.Run(ctx, config, exporter, svc)
			errorOccurred := err != nil
			if expectError != errorOccurred {
				t.Errorf("Unexpected result from test \"%v\": wanted error (%v), but got (%v)", name, expectError, errorOccurred)
			}
			entrypoint.ConfigValidatorFunc = prevConfValidation
			ctrl.Finish()
		})
	}
}

func noCheckConfig(_ *entrypoint.Config) error {
	return nil
}
