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

package service_test

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/dell/karavi-metrics-powerflex/internal/service"
	"github.com/dell/karavi-metrics-powerflex/internal/service/mocks"
	"github.com/sirupsen/logrus"

	"github.com/dell/karavi-metrics-powerflex/internal/k8s"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/metric/global"

	sio "github.com/dell/goscaleio"
	types "github.com/dell/goscaleio/types/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_GetSDCStatistics(t *testing.T) {
	type setup struct {
		Service *service.PowerFlexService
	}

	tests := map[string]func(t *testing.T) (setup, []service.StatisticsGetter, *gomock.Controller){
		"success": func(*testing.T) (setup, []service.StatisticsGetter, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			sdc1 := mocks.NewMockStatisticsGetter(ctrl)
			sdc1.EXPECT().GetStatistics().Return(&types.SdcStatistics{}, nil).Times(1)
			sdc2 := mocks.NewMockStatisticsGetter(ctrl)
			sdc2.EXPECT().GetStatistics().Return(&types.SdcStatistics{}, nil).Times(1)
			sdc3 := mocks.NewMockStatisticsGetter(ctrl)
			sdc3.EXPECT().GetStatistics().Return(&types.SdcStatistics{}, nil).Times(1)

			sdcs := []service.StatisticsGetter{sdc1, sdc2, sdc3}

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
			return setup{
				Service: &service,
			}, sdcs, ctrl
		},
		"nil list of sdcs": func(*testing.T) (setup, []service.StatisticsGetter, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			return setup{
				Service: &service,
			}, nil, ctrl
		},
		"error with 1 sdc": func(*testing.T) (setup, []service.StatisticsGetter, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			sdc1 := mocks.NewMockStatisticsGetter(ctrl)
			sdc1.EXPECT().GetStatistics().Return(nil, errors.New("error getting statistics")).Times(1)
			sdc2 := mocks.NewMockStatisticsGetter(ctrl)
			sdc2.EXPECT().GetStatistics().Return(&types.SdcStatistics{}, nil).Times(1)
			sdc3 := mocks.NewMockStatisticsGetter(ctrl)
			sdc3.EXPECT().GetStatistics().Return(&types.SdcStatistics{}, nil).Times(1)

			sdcs := []service.StatisticsGetter{sdc1, sdc2, sdc3}

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2)
			return setup{
				Service: &service,
			}, sdcs, ctrl
		},
		"error recording": func(*testing.T) (setup, []service.StatisticsGetter, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			sdc1 := mocks.NewMockStatisticsGetter(ctrl)
			sdc1.EXPECT().GetStatistics().Return(&types.SdcStatistics{}, nil).Times(1)

			sdcs := []service.StatisticsGetter{sdc1}

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("error"))
			return setup{
				Service: &service,
			}, sdcs, ctrl
		},
		"timing difference with sdc stats": func(t *testing.T) (setup, []service.StatisticsGetter, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			first, _ := time.ParseDuration("100ms")
			second, _ := time.ParseDuration("200ms")
			third, _ := time.ParseDuration("300ms")
			sdc1 := mocks.NewMockStatisticsGetter(ctrl)
			sdc1.EXPECT().GetStatistics().DoAndReturn(func() (*types.SdcStatistics, error) {
				time.Sleep(first)
				return &types.SdcStatistics{}, nil
			}).Times(1)
			sdc2 := mocks.NewMockStatisticsGetter(ctrl)
			sdc2.EXPECT().GetStatistics().DoAndReturn(func() (*types.SdcStatistics, error) {
				time.Sleep(second)
				return &types.SdcStatistics{}, nil
			}).Times(1)
			sdc3 := mocks.NewMockStatisticsGetter(ctrl)
			sdc3.EXPECT().GetStatistics().DoAndReturn(func() (*types.SdcStatistics, error) {
				time.Sleep(third)
				return &types.SdcStatistics{}, nil
			}).Times(1)

			sdcs := []service.StatisticsGetter{sdc1, sdc2, sdc3}

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
			return setup{
				Service: &service,
			}, sdcs, ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setup, sdcs, ctrl := tc(t)
			setup.Service.Logger = logrus.New()
			setup.Service.GetSDCStatistics(context.Background(), nil, sdcs)
			ctrl.Finish()
		})
	}
}

func Test_GetSDCBandwidth(t *testing.T) {
	tt := []struct {
		Name                   string
		Statistics             *types.SdcStatistics
		ExpectedReadBandwidth  float64
		ExpectedWriteBandwidth float64
	}{
		{
			"nil statistics",
			nil,
			0.0,
			0.0,
		},
		{
			"no data",
			&types.SdcStatistics{},
			0.0,
			0.0,
		},
		{
			"only read bandwidth",
			&types.SdcStatistics{
				UserDataReadBwc: types.BWC{TotalWeightInKb: 392040, NumSeconds: 110},
			},
			3.48046875,
			0.0,
		},
		{
			"only write bandwidth",
			&types.SdcStatistics{
				UserDataWriteBwc: types.BWC{TotalWeightInKb: 1958128, NumSeconds: 313},
			},
			0.0,
			6.109375,
		},
		{
			"read and write bandwidth",
			&types.SdcStatistics{
				UserDataReadBwc:  types.BWC{TotalWeightInKb: 1546272, NumSeconds: 236},
				UserDataWriteBwc: types.BWC{TotalWeightInKb: 12838, NumSeconds: 131},
			},
			6.3984375,
			0.095703125,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			readBandwidth, writeBandwidth := service.GetSDCBandwidth(tc.Statistics)
			assert.InDelta(t, tc.ExpectedReadBandwidth, readBandwidth, 0.001)
			assert.InDelta(t, tc.ExpectedWriteBandwidth, writeBandwidth, 0.001)
		})
	}
}

func Test_GetSDCIOPS(t *testing.T) {
	tt := []struct {
		Name              string
		Statistics        *types.SdcStatistics
		ExpectedReadIOPS  float64
		ExpectedWriteIOPS float64
	}{
		{
			"nil statistics",
			nil,
			0.0,
			0.0,
		},
		{
			"no data",
			&types.SdcStatistics{},
			0.0,
			0.0,
		},
		{
			"only read IOPS",
			&types.SdcStatistics{
				UserDataReadBwc: types.BWC{NumOccured: 6856870, NumSeconds: 114},
			},
			60147.982456,
			0.0,
		},
		{
			"only write IOPS",
			&types.SdcStatistics{
				UserDataWriteBwc: types.BWC{NumOccured: 354139516, NumSeconds: 3131},
			},
			0.0,
			113107.478760,
		},
		{
			"read and write IOPS",
			&types.SdcStatistics{
				UserDataReadBwc:  types.BWC{NumOccured: 94729, NumSeconds: 236},
				UserDataWriteBwc: types.BWC{NumOccured: 68122431, NumSeconds: 131},
			},
			401.394068,
			520018.557251,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			readIOPS, writeIOPS := service.GetSDCIOPS(tc.Statistics)
			assert.InDelta(t, tc.ExpectedReadIOPS, readIOPS, 0.001)
			assert.InDelta(t, tc.ExpectedWriteIOPS, writeIOPS, 0.001)
		})
	}
}

func Test_GetSDCLatency(t *testing.T) {
	tt := []struct {
		Name                 string
		Statistics           *types.SdcStatistics
		ExpectedReadLatency  float64
		ExpectedWriteLatency float64
	}{
		{
			"nil statistics",
			nil,
			0.0,
			0.0,
		},
		{
			"no data",
			&types.SdcStatistics{},
			0.0,
			0.0,
		},
		{
			"only read latency",
			&types.SdcStatistics{
				UserDataSdcReadLatency: types.BWC{TotalWeightInKb: 6856870, NumOccured: 114},
			},
			58.738264,
			0.0,
		},
		{
			"only write latency",
			&types.SdcStatistics{
				UserDataSdcWriteLatency: types.BWC{TotalWeightInKb: 354139516, NumOccured: 313},
			},
			0.0,
			1104.918119,
		},
		{
			"read and write latency",
			&types.SdcStatistics{
				UserDataSdcReadLatency:  types.BWC{TotalWeightInKb: 94729, NumOccured: 236},
				UserDataSdcWriteLatency: types.BWC{TotalWeightInKb: 68122431, NumOccured: 131},
			},
			0.391986,
			507.830622,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			readLatency, writeLatency := service.GetSDCLatency(tc.Statistics)
			assert.InDelta(t, tc.ExpectedReadLatency, readLatency, 0.001)
			assert.InDelta(t, tc.ExpectedWriteLatency, writeLatency, 0.001)
		})
	}
}

func Test_GetSDCs(t *testing.T) {
	type checkFn func(*testing.T, []service.StatisticsGetter, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, sdcs []service.StatisticsGetter, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkSdcLength := func(length int) func(t *testing.T, sdcs []service.StatisticsGetter, err error) {
		return func(t *testing.T, sdcs []service.StatisticsGetter, err error) {
			assert.Equal(t, length, len(sdcs))
		}
	}

	hasError := func(t *testing.T, sdc []service.StatisticsGetter, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller){
		"success": func(*testing.T) (service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			powerflexSystem := mocks.NewMockPowerFlexSystem(ctrl)
			service.SystemFinder = func(service.PowerFlexClient, string, string, string) (service.PowerFlexSystem, error) {
				return powerflexSystem, nil
			}
			sdcFinder := mocks.NewMockSDCFinder(ctrl)

			sdcFinder.EXPECT().GetSDCGuids().Times(1).Return([]string{"1", "2"}, nil)
			powerflexClient.EXPECT().GetInstance(gomock.Any()).Times(1).Return([]*types.System{{}}, nil)
			powerflexSystem.EXPECT().FindSdc(gomock.Any(), gomock.Any()).Times(2).Return(&sio.Sdc{}, nil)

			return powerflexClient, sdcFinder, check(hasNoError, checkSdcLength(2)), ctrl
		},
		"error calling GetSDCGuids": func(*testing.T) (service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			powerflexSystem := mocks.NewMockPowerFlexSystem(ctrl)
			service.SystemFinder = func(service.PowerFlexClient, string, string, string) (service.PowerFlexSystem, error) {
				return powerflexSystem, nil
			}
			sdcFinder := mocks.NewMockSDCFinder(ctrl)

			sdcFinder.EXPECT().GetSDCGuids().Times(1).Return(nil, errors.New("error"))

			return powerflexClient, sdcFinder, check(hasError), ctrl
		},
		"error calling GetInstance": func(*testing.T) (service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			powerflexClient.EXPECT().GetInstance(gomock.Any()).Times(1).Return(nil, errors.New("error"))
			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().Times(1).Return([]string{"1", "2"}, nil)
			return powerflexClient, sdcFinder, check(hasError), ctrl
		},
		"error calling FindSystem": func(*testing.T) (service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			service.SystemFinder = func(service.PowerFlexClient, string, string, string) (service.PowerFlexSystem, error) {
				return nil, errors.New("error")
			}
			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().Times(1).Return([]string{"1", "2"}, nil)

			powerflexClient.EXPECT().GetInstance(gomock.Any()).Times(1).Return([]*types.System{{}}, nil)
			return powerflexClient, sdcFinder, check(hasError), ctrl
		},
		"calling FindSdc with error returns 0 SDCs": func(*testing.T) (service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			powerflexSystem := mocks.NewMockPowerFlexSystem(ctrl)
			service.SystemFinder = func(service.PowerFlexClient, string, string, string) (service.PowerFlexSystem, error) {
				return powerflexSystem, nil
			}
			sdcFinder := mocks.NewMockSDCFinder(ctrl)
			sdcFinder.EXPECT().GetSDCGuids().Times(1).Return([]string{"1", "2"}, nil)
			powerflexClient.EXPECT().GetInstance(gomock.Any()).Times(1).Return([]*types.System{{}}, nil)
			powerflexSystem.EXPECT().FindSdc(gomock.Any(), gomock.Any()).Times(2).Return(nil, errors.New("error"))
			return powerflexClient, sdcFinder, check(hasNoError, checkSdcLength(0)), ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			powerflexClient, sdcFinder, checkFns, ctrl := tc(t)
			svc := &service.PowerFlexService{
				MetricsWrapper: &service.MetricsWrapper{
					Meter: global.Meter("powerflex/sdc"),
				},
				Logger: logrus.New(),
			}
			sdcsList, err := svc.GetSDCs(context.Background(), powerflexClient, sdcFinder)
			// powerflexClient, sdcFinder, checkFns, ctrl := tc(t)
			// sdcsList, err := service.GetSDCs(context.Background(), powerflexClient, sdcFinder)
			for _, checkFn := range checkFns {
				checkFn(t, sdcsList, err)
			}
			ctrl.Finish()
		})
	}
}

func Test_GetSDCMeta(t *testing.T) {
	type checkFn func(*testing.T, *service.SDCMeta)
	check := func(fns ...checkFn) []checkFn { return fns }

	checkSdcMeta := func(expectedOutput *service.SDCMeta) func(t *testing.T, sdcMeta *service.SDCMeta) {
		return func(t *testing.T, sdcMeta *service.SDCMeta) {
			assert.Equal(t, sdcMeta, expectedOutput)
		}
	}

	tests := map[string]func(t *testing.T) (*sio.Sdc, []corev1.Node, []checkFn){
		"success": func(*testing.T) (*sio.Sdc, []corev1.Node, []checkFn) {
			sdc := &sio.Sdc{
				Sdc: &types.Sdc{
					SdcIP: "1.2.3.4",
				},
			}

			nodes := []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Address: "1.2.3.4",
							},
						},
					},
				},
			}

			expectedOutput := &service.SDCMeta{
				Name: "node1",
				IP:   "1.2.3.4",
			}

			return sdc, nodes, check(checkSdcMeta(expectedOutput))
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			sdc, nodes, checkFns := tc(t)

			sdcMeta := service.GetSDCMeta(sdc, nodes)
			for _, checkFn := range checkFns {
				checkFn(t, sdcMeta)
			}
		})
	}
}

func Test_GetStorageClasses(t *testing.T) {
	type checkFn func(*testing.T, []service.StorageClassMeta, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, classes []service.StorageClassMeta, err error) {
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	checkPoolLength := func(class string, length int) func(t *testing.T, classes []service.StorageClassMeta, err error) {
		return func(t *testing.T, classes []service.StorageClassMeta, err error) {
			if class == "" {
				assert.Equal(t, 0, len(classes))
				return
			}

			for _, c := range classes {
				if c.Name == class {
					assert.Equal(t, length, len(c.StoragePools))
				}
			}
		}
	}

	hasError := func(t *testing.T, classes []service.StorageClassMeta, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (service.PowerFlexClient, service.StorageClassFinder, []checkFn, *gomock.Controller){
		"success one storage class one pool": func(*testing.T) (service.PowerFlexClient, service.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			storageClassFinder := mocks.NewMockStorageClassFinder(ctrl)

			powerflexClient.EXPECT().GetInstance("").Times(1).
				Return([]*types.System{
					{
						Name: "test",
					},
				}, nil)

			sc1 := v1.StorageClass{}
			sc1.Provisioner = "csi-vxflexos.dellemc.com"
			sc1.ObjectMeta = metav1.ObjectMeta{
				UID:  "123",
				Name: "class-1",
			}
			sc1.Parameters = map[string]string{
				"storagepool": "pool-1",
			}

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]v1.StorageClass{sc1}, nil)

			powerflexClient.EXPECT().GetStoragePool("").Times(1).
				Return([]*types.StoragePool{
					{
						Name: "pool-1",
					},
					{
						Name: "pool-2",
					},
				}, nil)

			return powerflexClient, storageClassFinder, check(hasNoError, checkPoolLength("class-1", 1)), ctrl
		},
		"success two storage classes one pool": func(*testing.T) (service.PowerFlexClient, service.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			storageClassFinder := mocks.NewMockStorageClassFinder(ctrl)

			powerflexClient.EXPECT().GetInstance("").Times(1).
				Return([]*types.System{
					{
						Name: "test",
					},
				}, nil)

			sc1 := v1.StorageClass{}
			sc1.Provisioner = "csi-vxflexos.dellemc.com"
			sc1.ObjectMeta = metav1.ObjectMeta{
				UID:  "123",
				Name: "class-1",
			}
			sc1.Parameters = map[string]string{
				"storagepool": "pool-1",
				"systemID":    "1234",
			}

			sc2 := v1.StorageClass{}
			sc2.Provisioner = "csi-vxflexos.dellemc.com"
			sc2.ObjectMeta = metav1.ObjectMeta{
				UID:  "1234",
				Name: "class-1-xfs",
			}
			sc2.Parameters = map[string]string{
				"storagepool": "pool-1",
				"systemID":    "5678",
			}

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]v1.StorageClass{sc1, sc2}, nil)

			powerflexClient.EXPECT().GetStoragePool("").Times(1).
				Return([]*types.StoragePool{
					{
						Name: "pool-1",
					},
					{
						Name: "pool-2",
					},
				}, nil)

			return powerflexClient, storageClassFinder, check(hasNoError, checkPoolLength("class-1", 1), checkPoolLength("class-1-xfs", 1)), ctrl
		},
		"error calling GetStorageClasses": func(*testing.T) (service.PowerFlexClient, service.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			storageClassFinder := mocks.NewMockStorageClassFinder(ctrl)

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).Return(nil, errors.New("error"))

			return powerflexClient, storageClassFinder, check(hasError), ctrl
		},
		"error calling GetInstances": func(*testing.T) (service.PowerFlexClient, service.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			storageClassFinder := mocks.NewMockStorageClassFinder(ctrl)

			sc1 := v1.StorageClass{}
			sc1.Provisioner = "csi-vxflexos.dellemc.com"
			sc1.ObjectMeta = metav1.ObjectMeta{
				UID:  "123",
				Name: "class-1",
			}
			sc1.Parameters = map[string]string{
				"storagepool": "pool-1",
			}

			sc2 := v1.StorageClass{}
			sc2.Provisioner = "csi-vxflexos.dellemc.com"
			sc2.ObjectMeta = metav1.ObjectMeta{
				UID:  "1234",
				Name: "class-1-xfs",
			}
			sc2.Parameters = map[string]string{
				"storagepool": "pool-1",
			}

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]v1.StorageClass{sc1, sc2}, nil)

			powerflexClient.EXPECT().GetInstance("").Times(1).
				Return(nil, errors.New("error"))

			return powerflexClient, storageClassFinder, check(hasError), ctrl
		},
		"error calling GetStoragePool": func(*testing.T) (service.PowerFlexClient, service.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			storageClassFinder := mocks.NewMockStorageClassFinder(ctrl)

			sc1 := v1.StorageClass{}
			sc1.Provisioner = "csi-vxflexos.dellemc.com"
			sc1.ObjectMeta = metav1.ObjectMeta{
				UID:  "123",
				Name: "class-1",
			}
			sc1.Parameters = map[string]string{
				"storagepool": "pool-1",
			}

			sc2 := v1.StorageClass{}
			sc2.Provisioner = "csi-vxflexos.dellemc.com"
			sc2.ObjectMeta = metav1.ObjectMeta{
				UID:  "1234",
				Name: "class-1-xfs",
			}
			sc2.Parameters = map[string]string{
				"storagepool": "pool-1",
			}

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]v1.StorageClass{sc1, sc2}, nil)

			powerflexClient.EXPECT().GetInstance("").Times(1).
				Return([]*types.System{
					{
						Name: "test",
					},
				}, nil)

			powerflexClient.EXPECT().GetStoragePool(gomock.Any()).Times(1).Return(nil, errors.New("error"))

			return powerflexClient, storageClassFinder, check(hasError), ctrl
		},
		"calling GetInstances returns 0 systems": func(*testing.T) (service.PowerFlexClient, service.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			storageClassFinder := mocks.NewMockStorageClassFinder(ctrl)

			sc1 := v1.StorageClass{}
			sc1.Provisioner = "csi-vxflexos.dellemc.com"
			sc1.ObjectMeta = metav1.ObjectMeta{
				UID:  "123",
				Name: "class-1",
			}
			sc1.Parameters = map[string]string{
				"storagepool": "pool-1",
			}

			sc2 := v1.StorageClass{}
			sc2.Provisioner = "csi-vxflexos.dellemc.com"
			sc2.ObjectMeta = metav1.ObjectMeta{
				UID:  "1234",
				Name: "class-1-xfs",
			}
			sc2.Parameters = map[string]string{
				"storagepool": "pool-1",
			}

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]v1.StorageClass{sc1, sc2}, nil)

			powerflexClient.EXPECT().GetInstance("").Times(1).
				Return([]*types.System{}, nil)

			return powerflexClient, storageClassFinder, check(hasError), ctrl
		},
		"calling GetStorageClasses returns 0 classes": func(*testing.T) (service.PowerFlexClient, service.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			storageClassFinder := mocks.NewMockStorageClassFinder(ctrl)

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]v1.StorageClass{}, nil)

			powerflexClient.EXPECT().GetInstance("").Times(1).
				Return([]*types.System{
					{
						Name: "test",
					},
				}, nil)

			powerflexClient.EXPECT().GetStoragePool("").Times(1).
				Return([]*types.StoragePool{
					{
						Name: "pool-1",
					},
					{
						Name: "pool-2",
					},
				}, nil)

			return powerflexClient, storageClassFinder, check(hasNoError, checkPoolLength("", 0)), ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			powerflexClient, storageClassFinder, checkFns, ctrl := tc(t)
			svc := &service.PowerFlexService{
				MetricsWrapper: &service.MetricsWrapper{
					Meter: global.Meter("powerflex/sdc"),
				},
				Logger: logrus.New(),
			}
			classToPools, err := svc.GetStorageClasses(context.Background(), powerflexClient, storageClassFinder)
			for _, checkFn := range checkFns {
				checkFn(t, classToPools, err)
			}
			ctrl.Finish()
		})
	}
}

func Test_GetStoragePoolStatistics(t *testing.T) {
	type setup struct {
		Service *service.PowerFlexService
	}

	tests := map[string]func(t *testing.T) (setup, []service.StorageClassMeta, *gomock.Controller){
		"success": func(*testing.T) (setup, []service.StorageClassMeta, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			sp1 := mocks.NewMockStoragePoolStatisticsGetter(ctrl)
			sp1.EXPECT().GetStatistics().Return(&types.Statistics{}, nil).Times(1)

			sp2 := mocks.NewMockStoragePoolStatisticsGetter(ctrl)
			sp2.EXPECT().GetStatistics().Return(&types.Statistics{}, nil).Times(1)

			scMetas := []service.StorageClassMeta{
				{
					"123",
					"class-1",
					"driver",
					"system1",
					map[string]service.StoragePoolStatisticsGetter{
						"poolID-1": sp1,
						"poolID-2": sp2,
					},
				},
			}

			metrics.EXPECT().RecordCapacity(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2)

			service := service.PowerFlexService{MetricsWrapper: metrics}
			return setup{
				Service: &service,
			}, scMetas, ctrl
		},
		"nil list of storage class metas": func(*testing.T) (setup, []service.StorageClassMeta, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().RecordCapacity(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			return setup{
				Service: &service,
			}, nil, ctrl
		},
		"error with 1 storage pool": func(*testing.T) (setup, []service.StorageClassMeta, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			sp1 := mocks.NewMockStoragePoolStatisticsGetter(ctrl)
			sp1.EXPECT().GetStatistics().Return(nil, errors.New("error getting statistics")).Times(1)
			sp2 := mocks.NewMockStoragePoolStatisticsGetter(ctrl)
			sp2.EXPECT().GetStatistics().Return(&types.Statistics{}, nil).Times(1)

			scMetas := []service.StorageClassMeta{
				{
					"123",
					"class-1",
					"driver",
					"system1",
					map[string]service.StoragePoolStatisticsGetter{
						"poolID-1": sp1,
						"poolID-2": sp2,
					},
				},
			}

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().RecordCapacity(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			return setup{
				Service: &service,
			}, scMetas, ctrl
		},
		"error recording": func(*testing.T) (setup, []service.StorageClassMeta, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			sp1 := mocks.NewMockStoragePoolStatisticsGetter(ctrl)
			sp1.EXPECT().GetStatistics().Return(&types.Statistics{}, nil).Times(1)

			scMetas := []service.StorageClassMeta{
				{
					"123",
					"class-1",
					"driver",
					"system1",
					map[string]service.StoragePoolStatisticsGetter{
						"poolID-1": sp1,
					},
				},
			}

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().RecordCapacity(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("error"))
			return setup{
				Service: &service,
			}, scMetas, ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setup, storageClassMetas, ctrl := tc(t)
			setup.Service.Logger = logrus.New()
			setup.Service.GetStoragePoolStatistics(context.Background(), storageClassMetas)
			ctrl.Finish()
		})
	}
}

func Benchmark_GetStoragePoolStatistics(b *testing.B) {
	numOfPools, poolQueryTime := 1024, "100ms"
	b.Logf("For %d pools and assuming each pool query takes %s\n", numOfPools, poolQueryTime)

	b.ReportAllocs()

	type setup struct {
		Service *service.PowerFlexService
	}

	storagePools := make(map[string]service.StoragePoolStatisticsGetter)
	ctrl := gomock.NewController(b)
	metrics := mocks.NewMockMetricsRecorder(ctrl)

	for i := 0; i < numOfPools; i++ {
		i := i
		tmpSp := mocks.NewMockStoragePoolStatisticsGetter(ctrl)
		tmpSp.EXPECT().GetStatistics().DoAndReturn(func() (*types.Statistics, error) {
			dur, _ := time.ParseDuration(poolQueryTime)
			time.Sleep(dur)
			return &types.Statistics{}, nil
		})
		storagePools["poolID-"+strconv.Itoa(i)] = tmpSp
	}

	scMetas := []service.StorageClassMeta{
		{
			"123",
			"class-1",
			"driver",
			"system1",
			storagePools,
		},
	}

	metrics.EXPECT().RecordCapacity(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(b.N * numOfPools)
	service := service.PowerFlexService{MetricsWrapper: metrics}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetStoragePoolStatistics(context.Background(), scMetas)
	}
}

func Test_GetVolumes(t *testing.T) {
	type setup struct {
		Service *service.PowerFlexService
	}
	type checkFn func(*testing.T, []service.VolumeStatisticsGetter, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasError := func(t *testing.T, vols []service.VolumeStatisticsGetter, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}
	hasNoError := func(t *testing.T, vols []service.VolumeStatisticsGetter, err error) {
		if err != nil {
			t.Fatalf("did not expected error but got %v", err)
		}
	}
	checkVolumeLength := func(length int) func(t *testing.T, vols []service.VolumeStatisticsGetter, err error) {
		return func(t *testing.T, vols []service.VolumeStatisticsGetter, err error) {
			assert.Equal(t, length, len(vols))
		}
	}

	tests := map[string]func(t *testing.T) (setup, []service.StatisticsGetter, []checkFn, *gomock.Controller){
		"success": func(*testing.T) (setup, []service.StatisticsGetter, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)

			mappedInfos := []*types.MappedSdcInfo{
				{SdcID: "60001", SdcIP: "10.234"},
				{SdcID: "60002", SdcIP: "10.235"},
			}
			volumes := []*types.Volume{
				{ID: "1", Name: "name_testing1", MappedSdcInfo: mappedInfos},
				{ID: "2", Name: "name_testing2", MappedSdcInfo: mappedInfos[:1]},
			}

			volumeClient := []*sio.Volume{
				{Volume: volumes[0]},
				{Volume: volumes[1]},
			}

			sdc1 := mocks.NewMockStatisticsGetter(ctrl)
			sdc1.EXPECT().GetVolume().Return(volumes, nil).AnyTimes()
			sdc1.EXPECT().FindVolumes().Return(volumeClient, nil).AnyTimes()
			sdc2 := mocks.NewMockStatisticsGetter(ctrl)
			sdc2.EXPECT().GetVolume().Return(volumes[:1], nil).AnyTimes()
			sdc2.EXPECT().FindVolumes().Return(volumeClient, nil).AnyTimes()
			sdc3 := mocks.NewMockStatisticsGetter(ctrl)
			sdc3.EXPECT().GetVolume().Return(append(volumes, volumes...), nil).AnyTimes()
			sdc3.EXPECT().FindVolumes().Return(append(volumeClient, volumeClient...), nil).AnyTimes()
			sdc4 := mocks.NewMockStatisticsGetter(ctrl)
			sdc4.EXPECT().GetVolume().Return(volumes[:0], nil).AnyTimes()
			sdc4.EXPECT().FindVolumes().Return(volumeClient[:0], nil).AnyTimes()

			sdcs := []service.StatisticsGetter{sdc1, sdc2, sdc3, sdc4}

			return setup{
				Service: &service.PowerFlexService{},
			}, sdcs, check(hasNoError, checkVolumeLength(len(volumes))), ctrl
		},
		"Failed GetVolume": func(*testing.T) (setup, []service.StatisticsGetter, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)

			sdc1 := mocks.NewMockStatisticsGetter(ctrl)
			sdc1.EXPECT().GetVolume().Return(nil, errors.New("error")).AnyTimes()
			sdc1.EXPECT().FindVolumes().Return(nil, errors.New("error")).AnyTimes()

			sdcs := []service.StatisticsGetter{sdc1}

			return setup{
				Service: &service.PowerFlexService{},
			}, sdcs, check(hasError), ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setup, sdcs, checkFns, ctrl := tc(t)
			setup.Service.Logger = logrus.New()
			volumes, err := setup.Service.GetVolumes(context.Background(), sdcs)
			for _, checkFn := range checkFns {
				checkFn(t, volumes, err)
			}
			ctrl.Finish()
		})
	}
}

func Test_ExportVolumeStatistics(t *testing.T) {
	type setup struct {
		Service *service.PowerFlexService
	}

	tests := map[string]func(t *testing.T) (setup, []service.VolumeStatisticsGetter, service.VolumeFinder, *gomock.Controller){
		"success": func(*testing.T) (setup, []service.VolumeStatisticsGetter, service.VolumeFinder, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			vol1 := mocks.NewMockVolumeStatisticsGetter(ctrl)
			vol1.EXPECT().GetVolumeStatistics().Return(&types.VolumeStatistics{}, nil).Times(1)
			vol2 := mocks.NewMockVolumeStatisticsGetter(ctrl)
			vol2.EXPECT().GetVolumeStatistics().Return(&types.VolumeStatistics{}, nil).Times(1)
			vol3 := mocks.NewMockVolumeStatisticsGetter(ctrl)
			vol3.EXPECT().GetVolumeStatistics().Return(&types.VolumeStatistics{}, nil).Times(1)

			vols := []service.VolumeStatisticsGetter{vol1, vol2, vol3}

			volFinder.EXPECT().GetPersistentVolumes().Return([]k8s.VolumeInfo{}, nil)

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
			return setup{
				Service: &service,
			}, vols, volFinder, ctrl
		},
		"sucess even with timing difference with volume stats": func(t *testing.T) (setup, []service.VolumeStatisticsGetter, service.VolumeFinder, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			first, _ := time.ParseDuration("100ms")
			second, _ := time.ParseDuration("200ms")
			third, _ := time.ParseDuration("300ms")
			vol1 := mocks.NewMockVolumeStatisticsGetter(ctrl)
			vol1.EXPECT().GetVolumeStatistics().DoAndReturn(func() (*types.VolumeStatistics, error) {
				time.Sleep(first)
				return &types.VolumeStatistics{}, nil
			}).Times(1)
			vol2 := mocks.NewMockVolumeStatisticsGetter(ctrl)
			vol2.EXPECT().GetVolumeStatistics().DoAndReturn(func() (*types.VolumeStatistics, error) {
				time.Sleep(second)
				return &types.VolumeStatistics{}, nil
			}).Times(1)
			vol3 := mocks.NewMockVolumeStatisticsGetter(ctrl)
			vol3.EXPECT().GetVolumeStatistics().DoAndReturn(func() (*types.VolumeStatistics, error) {
				time.Sleep(third)
				return &types.VolumeStatistics{}, nil
			}).Times(1)

			vols := []service.VolumeStatisticsGetter{vol1, vol2, vol3}

			volFinder.EXPECT().GetPersistentVolumes().Return([]k8s.VolumeInfo{}, nil)

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
			return setup{
				Service: &service,
			}, vols, volFinder, ctrl
		},
		"nil list of vols": func(*testing.T) (setup, []service.VolumeStatisticsGetter, service.VolumeFinder, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			volFinder.EXPECT().GetPersistentVolumes().Return([]k8s.VolumeInfo{}, nil)

			svc := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), &service.VolumeMeta{}, float64(0), float64(0), float64(0), float64(0), float64(0), float64(0)).Times(1)
			return setup{
				Service: &svc,
			}, nil, volFinder, ctrl
		},
		"error with 1 vol": func(*testing.T) (setup, []service.VolumeStatisticsGetter, service.VolumeFinder, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			vol1 := mocks.NewMockVolumeStatisticsGetter(ctrl)
			vol1.EXPECT().GetVolumeStatistics().Return(nil, errors.New("error getting statistics")).Times(1)
			vol2 := mocks.NewMockVolumeStatisticsGetter(ctrl)
			vol2.EXPECT().GetVolumeStatistics().Return(&types.VolumeStatistics{}, nil).Times(1)
			vol3 := mocks.NewMockVolumeStatisticsGetter(ctrl)
			vol3.EXPECT().GetVolumeStatistics().Return(&types.VolumeStatistics{}, nil).Times(1)

			vols := []service.VolumeStatisticsGetter{vol1, vol2, vol3}

			volFinder.EXPECT().GetPersistentVolumes().Return([]k8s.VolumeInfo{}, nil)

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2)
			return setup{
				Service: &service,
			}, vols, volFinder, ctrl
		},
		"error recording": func(*testing.T) (setup, []service.VolumeStatisticsGetter, service.VolumeFinder, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			vol1 := mocks.NewMockVolumeStatisticsGetter(ctrl)
			vol1.EXPECT().GetVolumeStatistics().Return(&types.VolumeStatistics{}, nil).Times(1)

			vols := []service.VolumeStatisticsGetter{vol1}

			volFinder.EXPECT().GetPersistentVolumes().Return([]k8s.VolumeInfo{}, nil)

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("error"))
			return setup{
				Service: &service,
			}, vols, volFinder, ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setup, vols, volFinder, ctrl := tc(t)
			setup.Service.Logger = logrus.New()
			setup.Service.ExportVolumeStatistics(context.Background(), vols, volFinder)
			ctrl.Finish()
		})
	}
}

func Benchmark_ExportVolumeStatistics(b *testing.B) {
	numOfVolumes, volumeQueryTime := 500, "100ms"
	b.Logf("For %d volumes and assuming each volume query takes %s\n", numOfVolumes, volumeQueryTime)

	b.ReportAllocs()

	type setup struct {
		Service *service.PowerFlexService
	}

	var volumes []service.VolumeStatisticsGetter
	ctrl := gomock.NewController(b)
	metrics := mocks.NewMockMetricsRecorder(ctrl)
	volFinder := mocks.NewMockVolumeFinder(ctrl)

	for i := 0; i < numOfVolumes; i++ {
		tmpVol := mocks.NewMockVolumeStatisticsGetter(ctrl)
		tmpVol.EXPECT().GetVolumeStatistics().DoAndReturn(func() (*types.VolumeStatistics, error) {
			dur, _ := time.ParseDuration(volumeQueryTime)
			time.Sleep(dur)
			return &types.VolumeStatistics{}, nil
		})
		volumes = append(volumes, tmpVol)
	}

	service := service.PowerFlexService{MetricsWrapper: metrics, Logger: logrus.New()}
	metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(b.N * numOfVolumes)
	volFinder.EXPECT().GetPersistentVolumes().Return([]k8s.VolumeInfo{}, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ExportVolumeStatistics(context.Background(), volumes, volFinder)
	}
}

func Test_GetVolumeBandwidth(t *testing.T) {
	tt := []struct {
		Name                   string
		Statistics             *types.VolumeStatistics
		ExpectedReadBandwidth  float64
		ExpectedWriteBandwidth float64
	}{
		{
			"nil statistics",
			nil,
			0.0,
			0.0,
		},
		{
			"no data",
			&types.VolumeStatistics{},
			0.0,
			0.0,
		},
		{
			"only read bandwidth",
			&types.VolumeStatistics{
				UserDataReadBwc: types.BWC{TotalWeightInKb: 392040, NumSeconds: 110},
			},
			3.48046875,
			0.0,
		},
		{
			"only write bandwidth",
			&types.VolumeStatistics{
				UserDataWriteBwc: types.BWC{TotalWeightInKb: 1958128, NumSeconds: 313},
			},
			0.0,
			6.109375,
		},
		{
			"read and write bandwidth",
			&types.VolumeStatistics{
				UserDataReadBwc:  types.BWC{TotalWeightInKb: 1546272, NumSeconds: 236},
				UserDataWriteBwc: types.BWC{TotalWeightInKb: 12838, NumSeconds: 131},
			},
			6.3984375,
			0.095703125,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			readBandwidth, writeBandwidth := service.GetVolumeBandwidth(tc.Statistics)
			assert.InDelta(t, tc.ExpectedReadBandwidth, readBandwidth, 0.001)
			assert.InDelta(t, tc.ExpectedWriteBandwidth, writeBandwidth, 0.001)
		})
	}
}

func Test_GetVolumeIOPS(t *testing.T) {
	tt := []struct {
		Name              string
		Statistics        *types.VolumeStatistics
		ExpectedReadIOPS  float64
		ExpectedWriteIOPS float64
	}{
		{
			"nil statistics",
			nil,
			0.0,
			0.0,
		},
		{
			"no data",
			&types.VolumeStatistics{},
			0.0,
			0.0,
		},
		{
			"only read IOPS",
			&types.VolumeStatistics{
				UserDataReadBwc: types.BWC{NumOccured: 6856870, NumSeconds: 114},
			},
			60147.982456,
			0.0,
		},
		{
			"only write IOPS",
			&types.VolumeStatistics{
				UserDataWriteBwc: types.BWC{NumOccured: 354139516, NumSeconds: 3131},
			},
			0.0,
			113107.478760,
		},
		{
			"read and write IOPS",
			&types.VolumeStatistics{
				UserDataReadBwc:  types.BWC{NumOccured: 94729, NumSeconds: 236},
				UserDataWriteBwc: types.BWC{NumOccured: 68122431, NumSeconds: 131},
			},
			401.394068,
			520018.557251,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			readIOPS, writeIOPS := service.GetVolumeIOPS(tc.Statistics)
			assert.InDelta(t, tc.ExpectedReadIOPS, readIOPS, 0.001)
			assert.InDelta(t, tc.ExpectedWriteIOPS, writeIOPS, 0.001)
		})
	}
}

func Test_GetVolumeLatency(t *testing.T) {
	tt := []struct {
		Name                 string
		Statistics           *types.VolumeStatistics
		ExpectedReadLatency  float64
		ExpectedWriteLatency float64
	}{
		{
			"nil statistics",
			nil,
			0.0,
			0.0,
		},
		{
			"no data",
			&types.VolumeStatistics{},
			0.0,
			0.0,
		},
		{
			"only read latency",
			&types.VolumeStatistics{
				UserDataSdcReadLatency: types.BWC{TotalWeightInKb: 6856870, NumOccured: 114},
			},
			58.738264,
			0.0,
		},
		{
			"only write latency",
			&types.VolumeStatistics{
				UserDataSdcWriteLatency: types.BWC{TotalWeightInKb: 354139516, NumOccured: 313},
			},
			0.0,
			1104.918119,
		},
		{
			"read and write latency",
			&types.VolumeStatistics{
				UserDataSdcReadLatency:  types.BWC{TotalWeightInKb: 94729, NumOccured: 236},
				UserDataSdcWriteLatency: types.BWC{TotalWeightInKb: 68122431, NumOccured: 131},
			},
			0.391986,
			507.830622,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			readLatency, writeLatency := service.GetVolumeLatency(tc.Statistics)
			assert.InDelta(t, tc.ExpectedReadLatency, readLatency, 0.001)
			assert.InDelta(t, tc.ExpectedWriteLatency, writeLatency, 0.001)
		})
	}
}

func Test_GetTotalLogicalCapacity(t *testing.T) {
	tt := []struct {
		Name             string
		Statistics       *types.Statistics
		ExpectedCapacity float64
	}{
		{
			"success",
			&types.Statistics{
				NetUserDataCapacityInKb: 16783360,
				NetUnusedCapacityInKb:   264777216,
			},
			268.517,
		},
		{
			"nil statistics",
			nil,
			0.0,
		},
		{
			"no data",
			&types.Statistics{},
			0.0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			capacity := service.GetTotalLogicalCapacity(tc.Statistics)
			assert.InDelta(t, tc.ExpectedCapacity, capacity, 0.001)
		})
	}
}

func Test_GetLogicalCapacityAvailable(t *testing.T) {
	tt := []struct {
		Name             string
		Statistics       *types.Statistics
		ExpectedCapacity float64
	}{
		{
			"success",
			&types.Statistics{
				NetUnusedCapacityInKb: 264777216,
			},
			252.511,
		},
		{
			"nil statistics",
			nil,
			0.0,
		},
		{
			"no data",
			&types.Statistics{},
			0.0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			capacity := service.GetLogicalCapacityAvailable(tc.Statistics)
			assert.InDelta(t, tc.ExpectedCapacity, capacity, 0.001)
		})
	}
}

func Test_GetLogicalCapacityInUse(t *testing.T) {
	tt := []struct {
		Name             string
		Statistics       *types.Statistics
		ExpectedCapacity float64
	}{
		{
			"success",
			&types.Statistics{
				NetUserDataCapacityInKb: 16783360,
			},
			16.005,
		},
		{
			"nil statistics",
			nil,
			0.0,
		},
		{
			"no data",
			&types.Statistics{},
			0.0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			capacity := service.GetLogicalCapacityInUse(tc.Statistics)
			assert.InDelta(t, tc.ExpectedCapacity, capacity, 0.001)
		})
	}
}

func Test_GetLogicalProvisioned(t *testing.T) {
	tt := []struct {
		Name             string
		Statistics       *types.Statistics
		ExpectedCapacity float64
	}{
		{
			"success",
			&types.Statistics{
				VolumeAddressSpaceInKb: 58720256,
			},
			56,
		},
		{
			"nil statistics",
			nil,
			0.0,
		},
		{
			"no data",
			&types.Statistics{},
			0.0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			capacity := service.GetLogicalProvisioned(tc.Statistics)
			assert.InDelta(t, tc.ExpectedCapacity, capacity, 0.001)
		})
	}
}
