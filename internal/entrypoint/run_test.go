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

package entrypoint_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/dell/karavi-metrics-powerflex/internal/entrypoint"
	"github.com/dell/karavi-metrics-powerflex/internal/k8s"
	pflexServices "github.com/dell/karavi-metrics-powerflex/internal/service"
	metricsmocks "github.com/dell/karavi-metrics-powerflex/internal/service/mocks"
	otlexporters "github.com/dell/karavi-metrics-powerflex/opentelemetry/exporters"
	exportermocks "github.com/dell/karavi-metrics-powerflex/opentelemetry/exporters/mocks"

	sio "github.com/dell/goscaleio"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Run(t *testing.T) {
	validSDCTickInterval := entrypoint.MinimumSDCTickInterval
	validVolumeTickInterval := entrypoint.MinimumSDCTickInterval

	tests := map[string]func(t *testing.T) (expectError bool, config *entrypoint.Config, exporter otlexporters.Otlexporter, pflexSvc pflexServices.Service, prevConfigValidationFunc func(*entrypoint.Config) error, ctrl *gomock.Controller, validatingConfig bool){
		"success": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)

			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			nodeFinder := metricsmocks.NewMockNodeFinder(ctrl)
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

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				NodeFinder:                  nodeFinder,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           true,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.SdcMetricsRetriever{},
				nil,
			)
			svc.EXPECT().GetSDCStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			svc.EXPECT().ExportTopologyMetrics(gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},

		"success even if error during call to GetSDCs": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)

			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			nodeFinder := metricsmocks.NewMockNodeFinder(ctrl)
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

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				NodeFinder:                  nodeFinder,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           true,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				nil,
				errors.New("error"),
			)
			svc.EXPECT().GetSDCStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success even if error during call to NodeFinder": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)

			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			nodeFinder := metricsmocks.NewMockNodeFinder(ctrl)
			nodeFinder.EXPECT().GetNodes().AnyTimes().
				Return([]corev1.Node{}, errors.New("error"))

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				NodeFinder:                  nodeFinder,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           true,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.SdcMetricsRetriever{},
				nil,
			)
			svc.EXPECT().GetSDCStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success even if SDC metrics collection is disabled": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)

			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			// GetSDCGuids should not be called because SDC metrics collection is disabled
			sdcFinder.EXPECT().GetSDCGuids().Times(0).Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			nodeFinder := metricsmocks.NewMockNodeFinder(ctrl)
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

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				NodeFinder:                  nodeFinder,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           false,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)
			// GetSDCs should not be called because SDC metrics collection is disabled
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).Times(0).Return(
				nil,
				errors.New("error"),
			)
			svc.EXPECT().GetSDCStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"topology metrics not collected if not leader": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection(gomock.Any(), gomock.Any()).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(false)

			// Service should not receive ExportTopologyMetrics call
			svc := metricsmocks.NewMockService(ctrl)
			// no EXPECT() on ExportTopologyMetrics means if it's called test will fail

			exporter := exportermocks.NewMockOtlexporter(ctrl)
			exporter.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			exporter.EXPECT().StopExporter().Return(nil)

			config := &entrypoint.Config{
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           false,
				VolumeMetricsEnabled:        false,
				StoragePoolMetricsEnabled:   false,
				TopologyMetricsEnabled:      true,
				SDCTickInterval:             100 * time.Millisecond,
				VolumeTickInterval:          100 * time.Millisecond,
				StoragePoolTickInterval:     100 * time.Millisecond,
				TopologyMetricsTickInterval: 100 * time.Millisecond,
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": nil},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {}},
				SDCFinder:                   metricsmocks.NewMockSDCFinder(ctrl),
				NodeFinder:                  metricsmocks.NewMockNodeFinder(ctrl),
			}

			prev := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			return false, config, exporter, svc, prev, ctrl, false
		},
		"error no PowerFlex client": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			nodeFinder := metricsmocks.NewMockNodeFinder(ctrl)
			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient: nil,
				PowerFlexConfig: map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}}, SDCFinder: sdcFinder,
				NodeFinder:                  nodeFinder,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           true,
				SDCTickInterval:             validSDCTickInterval,
				VolumeTickInterval:          validVolumeTickInterval,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metricsmocks.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"success with no PowerFlex config": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			nodeFinder := metricsmocks.NewMockNodeFinder(ctrl)
			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)

			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"wrong": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				NodeFinder:                  nodeFinder,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           true,
				VolumeMetricsEnabled:        true,
				StoragePoolMetricsEnabled:   true,
				SDCTickInterval:             validSDCTickInterval,
				VolumeTickInterval:          validVolumeTickInterval,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"error no SDC Finder": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)
			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   nil,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           true,
				SDCTickInterval:             validSDCTickInterval,
				VolumeTickInterval:          validVolumeTickInterval,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metricsmocks.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error no Node Finder": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)
			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				NodeFinder:                  nil,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           true,
				SDCTickInterval:             validSDCTickInterval,
				VolumeTickInterval:          validVolumeTickInterval,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metricsmocks.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error invalid SDC poll time": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)
			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			nodeFinder := metricsmocks.NewMockNodeFinder(ctrl)
			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				NodeFinder:                  nodeFinder,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           true,
				SDCTickInterval:             entrypoint.MinimumSDCTickInterval - time.Second,
				VolumeTickInterval:          validVolumeTickInterval,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metricsmocks.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error invalid Volume poll time (too low)": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)
			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			nodeFinder := metricsmocks.NewMockNodeFinder(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				NodeFinder:                  nodeFinder,
				SDCMetricsEnabled:           true,
				SDCTickInterval:             validSDCTickInterval,
				VolumeTickInterval:          entrypoint.MinimumVolTickInterval - time.Second,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metricsmocks.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error invalid Volume poll time (too high)": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)
			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			nodeFinder := metricsmocks.NewMockNodeFinder(ctrl)
			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				NodeFinder:                  nodeFinder,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           true,
				SDCTickInterval:             validSDCTickInterval,
				VolumeTickInterval:          entrypoint.MaximumVolTickInterval + time.Second,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metricsmocks.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error nil config": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			e := exportermocks.NewMockOtlexporter(ctrl)

			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			svc := metricsmocks.NewMockService(ctrl)

			return true, nil, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error initializing exporter": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)
			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           true,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(fmt.Errorf("An error occurred while initializing the exporter"))
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success for volume metrics": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)

			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				LeaderElector:               leaderElector,
				VolumeMetricsEnabled:        true,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.SdcMetricsRetriever{},
				nil,
			)
			svc.EXPECT().GetVolumes(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]*pflexServices.VolumeMetaMetrics{},
				nil,
			)
			svc.EXPECT().ExportVolumeStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"error getting volumes": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)

			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				LeaderElector:               leaderElector,
				VolumeMetricsEnabled:        true,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]pflexServices.SdcMetricsRetriever{},
				nil,
			)
			svc.EXPECT().GetVolumes(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]*pflexServices.VolumeMetaMetrics{},
				errors.New("error"),
			)

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"volume success even if error during call to GetSDCs": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)

			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().AnyTimes().Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				LeaderElector:               leaderElector,
				VolumeMetricsEnabled:        true,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)
			svc.EXPECT().GetSDCs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				nil,
				errors.New("error"),
			)
			svc.EXPECT().GetVolumes(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(
				[]*pflexServices.VolumeMetaMetrics{},
				errors.New("error"),
			)
			svc.EXPECT().ExportVolumeStatistics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success for storage class/pool": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)

			sc1 := k8s.StorageClass{
				StorageClass: v1.StorageClass{
					ObjectMeta: metav1.ObjectMeta{
						UID:  "123",
						Name: "class-1",
					},
					Provisioner: "csi-vxflexos.dellemc.com",
					Parameters: map[string]string{
						"storagepool": "pool-1",
					},
				},
				SystemID: "123",
			}

			storageClassFinder := metricsmocks.NewMockStorageClassFinder(ctrl)
			storageClassFinder.EXPECT().GetStorageClasses().AnyTimes().
				Return([]k8s.StorageClass{sc1}, nil)

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				StorageClassFinder:          storageClassFinder,
				LeaderElector:               leaderElector,
				StoragePoolMetricsEnabled:   true,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)
			svc.EXPECT().GetStorageClasses(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]pflexServices.StorageClassMeta{
					{
						ID:           "123",
						Name:         "class-1",
						Driver:       "csi-vxflexos.dellemc.com",
						StoragePools: map[string]pflexServices.StoragePoolMetricsRetriever{},
					},
				}, nil).AnyTimes()

			svc.EXPECT().GetStoragePoolStatistics(gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"error no LeaderElector": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)
			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				LeaderElector:               nil,
				SDCMetricsEnabled:           true,
				SDCTickInterval:             validSDCTickInterval,
				VolumeTickInterval:          validVolumeTickInterval,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metricsmocks.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success even if is leader is false": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)

			sdcFinder := metricsmocks.NewMockSDCFinder(ctrl)
			// GetSDCGuids should not be called because SDC metrics collection is disabled
			sdcFinder.EXPECT().GetSDCGuids().Times(0).Return([]string{"1.2.3.4", "1.2.3.5"}, nil)

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(false)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				SDCFinder:                   sdcFinder,
				LeaderElector:               leaderElector,
				SDCMetricsEnabled:           false,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)
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
			pfClient := metricsmocks.NewMockPowerFlexClient(ctrl)

			sc1 := k8s.StorageClass{
				StorageClass: v1.StorageClass{
					ObjectMeta: metav1.ObjectMeta{
						UID:  "123",
						Name: "class-1",
					},
					Provisioner: "csi-vxflexos.dellemc.com",
					Parameters: map[string]string{
						"storagepool": "pool-1",
					},
				},
				SystemID: "123",
			}

			storageClassFinder := metricsmocks.NewMockStorageClassFinder(ctrl)
			storageClassFinder.EXPECT().GetStorageClasses().AnyTimes().
				Return([]k8s.StorageClass{sc1}, nil)

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"key": pfClient},
				PowerFlexConfig:             map[string]sio.ConfigConnect{"key": {Username: "powerFlexGatewayUser", Password: "powerFlexGatewayPassword"}},
				StorageClassFinder:          storageClassFinder,
				LeaderElector:               leaderElector,
				StoragePoolMetricsEnabled:   true,
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)
			svc.EXPECT().GetStorageClasses(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, fmt.Errorf("there was error getting the StorageClass")).AnyTimes()

			svc.EXPECT().GetStoragePoolStatistics(gomock.Any(), gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success using TLS": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				LeaderElector:               leaderElector,
				CollectorCertPath:           "testdata/test-cert.crt",
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"error reading certificate": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pflexServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)

			leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerflex", "karavi").AnyTimes().Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				LeaderElector:               leaderElector,
				CollectorCertPath:           "testdata/bad-cert.crt",
				TopologyMetricsEnabled:      true,
				TopologyMetricsTickInterval: 30 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metricsmocks.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expectError, config, exporter, svc, prevConfValidation, ctrl, validateConfig := test(t)
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			if config != nil {
				config.Logger = logrus.New()
				if !validateConfig {
					// The configuration is not nil and the test is not attempting to validate the configuration.
					// In this case, we can use smaller intervals for testing purposes.
					config.SDCTickInterval = 100 * time.Millisecond
					config.VolumeTickInterval = 100 * time.Millisecond
					config.StoragePoolTickInterval = 100 * time.Millisecond
				}
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

func Test_ValidateConfig_TopologyTickInterval_OutOfRange(t *testing.T) {
	tooSmall := &entrypoint.Config{
		SDCTickInterval:             entrypoint.MinimumSDCTickInterval,
		VolumeTickInterval:          entrypoint.MinimumSDCTickInterval,
		TopologyMetricsTickInterval: entrypoint.MinimumTickInterval - time.Second, // too small
		PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"k": nil},
		SDCFinder:                   metricsmocks.NewMockSDCFinder(gomock.NewController(t)),
		NodeFinder:                  metricsmocks.NewMockNodeFinder(gomock.NewController(t)),
		LeaderElector:               metricsmocks.NewMockLeaderElector(gomock.NewController(t)),
		SDCMetricsEnabled:           true,
	}

	err := entrypoint.ValidateConfig(tooSmall)
	if err == nil {
		t.Fatalf("expected error for topology tick interval too small, got nil")
	}

	tooLarge := &entrypoint.Config{
		SDCTickInterval:             entrypoint.MinimumSDCTickInterval,
		VolumeTickInterval:          entrypoint.MinimumSDCTickInterval,
		TopologyMetricsTickInterval: entrypoint.MaximumTickInterval + time.Second, // too large
		PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"k": nil},
		SDCFinder:                   metricsmocks.NewMockSDCFinder(gomock.NewController(t)),
		NodeFinder:                  metricsmocks.NewMockNodeFinder(gomock.NewController(t)),
		LeaderElector:               metricsmocks.NewMockLeaderElector(gomock.NewController(t)),
		SDCMetricsEnabled:           true,
	}

	err = entrypoint.ValidateConfig(tooLarge)
	if err == nil {
		t.Fatalf("expected error for topology tick interval too large, got nil")
	}
}

func Test_ValidateConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &entrypoint.Config{
		SDCTickInterval:             entrypoint.MinimumSDCTickInterval,
		VolumeTickInterval:          entrypoint.MinimumVolTickInterval,
		TopologyMetricsTickInterval: entrypoint.MinimumTickInterval,
		PowerFlexClient:             map[string]pflexServices.PowerFlexClient{"k": nil},
		SDCFinder:                   metricsmocks.NewMockSDCFinder(ctrl),
		NodeFinder:                  metricsmocks.NewMockNodeFinder(ctrl),
	}

	err := entrypoint.ValidateConfig(config)
	if err != nil {
		t.Fatalf("expected no error for valid config, got %v", err)
	}
}

func Test_Run_TopologyDisabledWhenLeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
	leaderElector.EXPECT().InitLeaderElection(gomock.Any(), gomock.Any()).Return(nil)
	leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

	svc := metricsmocks.NewMockService(ctrl)

	exporter := exportermocks.NewMockOtlexporter(ctrl)
	exporter.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
	exporter.EXPECT().StopExporter().Return(nil)

	config := &entrypoint.Config{
		LeaderElector:               leaderElector,
		SDCMetricsEnabled:           false,
		VolumeMetricsEnabled:        false,
		StoragePoolMetricsEnabled:   false,
		TopologyMetricsEnabled:      false,
		SDCTickInterval:             100 * time.Millisecond,
		VolumeTickInterval:          100 * time.Millisecond,
		StoragePoolTickInterval:     100 * time.Millisecond,
		TopologyMetricsTickInterval: 100 * time.Millisecond,
		PowerFlexClient:             map[string]pflexServices.PowerFlexClient{},
		PowerFlexConfig:             map[string]sio.ConfigConnect{},
		SDCFinder:                   metricsmocks.NewMockSDCFinder(ctrl),
		NodeFinder:                  metricsmocks.NewMockNodeFinder(ctrl),
		Logger:                      logrus.New(),
	}

	prev := entrypoint.ConfigValidatorFunc
	entrypoint.ConfigValidatorFunc = noCheckConfig
	defer func() { entrypoint.ConfigValidatorFunc = prev }()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := entrypoint.Run(ctx, config, exporter, svc)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func Test_Run_StopExporterError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
	leaderElector.EXPECT().InitLeaderElection(gomock.Any(), gomock.Any()).Return(nil)
	leaderElector.EXPECT().IsLeader().AnyTimes().Return(false)

	svc := metricsmocks.NewMockService(ctrl)

	exporter := exportermocks.NewMockOtlexporter(ctrl)
	exporter.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
	exporter.EXPECT().StopExporter().Return(fmt.Errorf("stop exporter error"))

	config := &entrypoint.Config{
		LeaderElector:               leaderElector,
		SDCMetricsEnabled:           false,
		VolumeMetricsEnabled:        false,
		StoragePoolMetricsEnabled:   false,
		TopologyMetricsEnabled:      false,
		SDCTickInterval:             100 * time.Millisecond,
		VolumeTickInterval:          100 * time.Millisecond,
		StoragePoolTickInterval:     100 * time.Millisecond,
		TopologyMetricsTickInterval: 100 * time.Millisecond,
		PowerFlexClient:             map[string]pflexServices.PowerFlexClient{},
		PowerFlexConfig:             map[string]sio.ConfigConnect{},
		SDCFinder:                   metricsmocks.NewMockSDCFinder(ctrl),
		NodeFinder:                  metricsmocks.NewMockNodeFinder(ctrl),
		Logger:                      logrus.New(),
	}

	prev := entrypoint.ConfigValidatorFunc
	entrypoint.ConfigValidatorFunc = noCheckConfig
	defer func() { entrypoint.ConfigValidatorFunc = prev }()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := entrypoint.Run(ctx, config, exporter, svc)
	if err != nil {
		t.Fatalf("expected no error from Run, got %v", err)
	}
}

func Test_Run_TickIntervalChange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
	leaderElector.EXPECT().InitLeaderElection(gomock.Any(), gomock.Any()).Return(nil)
	leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

	config := &entrypoint.Config{
		LeaderElector:               leaderElector,
		SDCMetricsEnabled:           false,
		VolumeMetricsEnabled:        false,
		StoragePoolMetricsEnabled:   false,
		TopologyMetricsEnabled:      true,
		SDCTickInterval:             100 * time.Millisecond,
		VolumeTickInterval:          100 * time.Millisecond,
		StoragePoolTickInterval:     100 * time.Millisecond,
		TopologyMetricsTickInterval: 100 * time.Millisecond,
		PowerFlexClient:             map[string]pflexServices.PowerFlexClient{},
		PowerFlexConfig:             map[string]sio.ConfigConnect{},
		SDCFinder:                   metricsmocks.NewMockSDCFinder(ctrl),
		NodeFinder:                  metricsmocks.NewMockNodeFinder(ctrl),
		Logger:                      logrus.New(),
	}

	// Change tick intervals inside the mock callback so the mutation
	// happens in the same goroutine as Run(), avoiding a data race.
	callCount := 0
	svc := metricsmocks.NewMockService(ctrl)
	svc.EXPECT().ExportTopologyMetrics(gomock.Any()).AnyTimes().Do(func(_ context.Context) {
		callCount++
		if callCount == 1 {
			config.SDCTickInterval = 150 * time.Millisecond
			config.VolumeTickInterval = 150 * time.Millisecond
			config.StoragePoolTickInterval = 150 * time.Millisecond
			config.TopologyMetricsTickInterval = 150 * time.Millisecond
		}
	})

	exporter := exportermocks.NewMockOtlexporter(ctrl)
	exporter.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
	exporter.EXPECT().StopExporter().Return(nil)

	prev := entrypoint.ConfigValidatorFunc
	entrypoint.ConfigValidatorFunc = noCheckConfig
	defer func() { entrypoint.ConfigValidatorFunc = prev }()

	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	err := entrypoint.Run(ctx, config, exporter, svc)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func Test_Run_EnvVarOverrides(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	leaderElector := metricsmocks.NewMockLeaderElector(ctrl)
	leaderElector.EXPECT().InitLeaderElection("custom-endpoint", "custom-namespace").Return(nil)
	leaderElector.EXPECT().IsLeader().AnyTimes().Return(false)

	svc := metricsmocks.NewMockService(ctrl)

	exporter := exportermocks.NewMockOtlexporter(ctrl)
	exporter.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
	exporter.EXPECT().StopExporter().Return(nil)

	config := &entrypoint.Config{
		LeaderElector:               leaderElector,
		SDCMetricsEnabled:           false,
		VolumeMetricsEnabled:        false,
		StoragePoolMetricsEnabled:   false,
		TopologyMetricsEnabled:      false,
		SDCTickInterval:             100 * time.Millisecond,
		VolumeTickInterval:          100 * time.Millisecond,
		StoragePoolTickInterval:     100 * time.Millisecond,
		TopologyMetricsTickInterval: 100 * time.Millisecond,
		PowerFlexClient:             map[string]pflexServices.PowerFlexClient{},
		PowerFlexConfig:             map[string]sio.ConfigConnect{},
		SDCFinder:                   metricsmocks.NewMockSDCFinder(ctrl),
		NodeFinder:                  metricsmocks.NewMockNodeFinder(ctrl),
		Logger:                      logrus.New(),
	}

	prev := entrypoint.ConfigValidatorFunc
	entrypoint.ConfigValidatorFunc = noCheckConfig
	defer func() { entrypoint.ConfigValidatorFunc = prev }()

	t.Setenv("POWERFLEX_METRICS_ENDPOINT", "custom-endpoint")
	t.Setenv("POWERFLEX_METRICS_NAMESPACE", "custom-namespace")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := entrypoint.Run(ctx, config, exporter, svc)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
