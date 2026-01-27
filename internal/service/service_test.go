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

package service_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/dell/karavi-metrics-powerflex/internal/k8s"

	"github.com/dell/karavi-metrics-powerflex/internal/service"
	"github.com/dell/karavi-metrics-powerflex/internal/service/mocks"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	sio "github.com/dell/goscaleio"
	types "github.com/dell/goscaleio/types/v1"
	"github.com/agiledragon/gomonkey/v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ecRetriever is a minimal test double for service.SdcMetricsRetriever that
// returns GenTypeEC and a mock PowerFlex client for the EC branch.
type ecRetriever struct {
	sdc    *sio.Sdc
	client service.PowerFlexClient
	stats  service.StatisticsGetter
	gen    string
}

func (e ecRetriever) GetSdc() *sio.Sdc                              { return e.sdc }
func (e ecRetriever) GetClient() service.PowerFlexClient            { return e.client }
func (e ecRetriever) GetGen() string                                { return e.gen }
func (e ecRetriever) GetStatisticsGetter() service.StatisticsGetter { return e.stats }

type ecPoolRetriever struct {
	client service.PowerFlexClient
	stats  service.StoragePoolStatisticsGetter
	gen    string
}

func (e ecPoolRetriever) GetClient() service.PowerFlexClient                       { return e.client }
func (e ecPoolRetriever) GetStatisticsGetter() service.StoragePoolStatisticsGetter { return e.stats }
func (e ecPoolRetriever) GetGen() string                                           { return e.gen }

type fakeSystemFinderTarget struct {
	byGUID map[string]*sio.Sdc
}

var _ service.PowerFlexSystem = (*fakeSystemFinderTarget)(nil)

type sdcToVolumes map[*sio.Sdc][]*sio.Volume

func Test_GetSDCStatistics(t *testing.T) {
	type setup struct {
		Service *service.PowerFlexService
	}

	tests := map[string]func(t *testing.T) (setup, []service.SdcMetricsRetriever, *gomock.Controller){
		"success": func(*testing.T) (setup, []service.SdcMetricsRetriever, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			sdc1 := mocks.NewMockStatisticsGetter(ctrl)
			sdc1.EXPECT().GetStatistics().Return(&types.SdcStatistics{}, nil).Times(1)
			sdc2 := mocks.NewMockStatisticsGetter(ctrl)
			sdc2.EXPECT().GetStatistics().Return(&types.SdcStatistics{}, nil).Times(1)
			sdc3 := mocks.NewMockStatisticsGetter(ctrl)
			sdc3.EXPECT().GetStatistics().Return(&types.SdcStatistics{}, nil).Times(1)
			retrievers := []service.SdcMetricsRetriever{
				newSdcRetriever(t, ctrl, sdc1, "v1", &sio.Sdc{Sdc: &types.Sdc{
					SdcIP:   "1.2.3.4",
					ID:      "sdc-id-124",
					SdcGUID: "guid-xyz-789",
				}}),
				newSdcRetriever(t, ctrl, sdc2, "v1", &sio.Sdc{Sdc: &types.Sdc{
					SdcIP:   "1.2.3.5",
					ID:      "sdc-id-125",
					SdcGUID: "guid-xyz-790",
				}}),
				newSdcRetriever(t, ctrl, sdc3, "v1", &sio.Sdc{Sdc: &types.Sdc{
					SdcIP:   "1.2.3.6",
					ID:      "sdc-id-126",
					SdcGUID: "guid-xyz-791",
				}}),
			}
			metrics.EXPECT().
				Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(3)

			svc := service.PowerFlexService{MetricsWrapper: metrics}
			return setup{Service: &svc}, retrievers, ctrl
		},
		"ec metrics success": func(*testing.T) (setup, []service.SdcMetricsRetriever, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			client := mocks.NewMockPowerFlexClient(ctrl)

			sdcID := "sdc-ec-001"
			sdc := &sio.Sdc{Sdc: &types.Sdc{
				SdcIP:   "9.9.9.9",
				ID:      sdcID,
				SdcGUID: "guid-ec-001",
			}}

			client.EXPECT().
				GetMetrics("sdc", []string{sdcID}).
				Return(&types.MetricsResponse{
					Resources: []types.Resource{
						{
							ID: sdcID,
							Metrics: []types.Metric{
								{Name: "host_read_bandwidth", Values: []float64{1048576}},
								{Name: "host_write_bandwidth", Values: []float64{2097152}},
								{Name: "host_read_iops", Values: []float64{123}},
								{Name: "host_write_iops", Values: []float64{456}},
								{Name: "avg_host_read_latency", Values: []float64{5000}},
								{Name: "avg_host_write_latency", Values: []float64{7000}},
							},
						},
					},
				}, nil).
				Times(1)

			sg := mocks.NewMockStatisticsGetter(ctrl)

			retrievers := []service.SdcMetricsRetriever{
				ecRetriever{
					sdc:    sdc,
					client: client,
					stats:  sg,
					gen:    types.GenTypeEC,
				},
			}

			metrics.EXPECT().
				Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(1)

			svc := service.PowerFlexService{MetricsWrapper: metrics}
			return setup{Service: &svc}, retrievers, ctrl
		},
		"nil list of sdcs": func(*testing.T) (setup, []service.SdcMetricsRetriever, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			return setup{
				Service: &service,
			}, nil, ctrl
		},
		"error with 1 sdc": func(*testing.T) (setup, []service.SdcMetricsRetriever, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)

			sdc1 := mocks.NewMockStatisticsGetter(ctrl)
			sdc1.EXPECT().GetStatistics().Return(nil, errors.New("error getting statistics")).Times(1)
			retrievers := []service.SdcMetricsRetriever{
				newSdcRetriever(t, ctrl, sdc1, "v1", &sio.Sdc{Sdc: &types.Sdc{
					SdcIP:   "1.2.3.4",
					ID:      "sdc-id-124",
					SdcGUID: "guid-xyz-789",
				}}),
			}
			svc := service.PowerFlexService{MetricsWrapper: metrics}
			return setup{Service: &svc}, retrievers, ctrl
		},
		"timing difference with sdc stats": func(t *testing.T) (setup, []service.SdcMetricsRetriever, *gomock.Controller) {
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
			metrics.EXPECT().
				Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(3)
			retrievers := []service.SdcMetricsRetriever{
				newSdcRetriever(t, ctrl, sdc1, "v1", &sio.Sdc{Sdc: &types.Sdc{
					SdcIP:   "1.2.3.4",
					ID:      "sdc-id-124",
					SdcGUID: "guid-xyz-789",
				}}),
				newSdcRetriever(t, ctrl, sdc2, "v1", &sio.Sdc{Sdc: &types.Sdc{
					SdcIP:   "1.2.3.5",
					ID:      "sdc-id-125",
					SdcGUID: "guid-xyz-790",
				}}),
				newSdcRetriever(t, ctrl, sdc3, "v1", &sio.Sdc{Sdc: &types.Sdc{
					SdcIP:   "1.2.3.6",
					ID:      "sdc-id-126",
					SdcGUID: "guid-xyz-791",
				}}),
			}
			svc := service.PowerFlexService{MetricsWrapper: metrics}
			return setup{Service: &svc}, retrievers, ctrl
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

func (f *fakeSystemFinderTarget) FindSdc(_, value string) (*sio.Sdc, error) {
	if s, ok := f.byGUID[value]; ok {
		return s, nil
	}
	return nil, errors.New("not found")
}

func Test_GetSDCs(t *testing.T) {
	type checkFn func(*testing.T, []service.SdcMetricsRetriever, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasError := func(t *testing.T, _ []service.SdcMetricsRetriever, err error) {
		require.Error(t, err)
	}
	noErrorAndLen := func(n int) checkFn {
		return func(t *testing.T, got []service.SdcMetricsRetriever, err error) {
			require.NoError(t, err)
			assert.Len(t, got, n)
		}
	}

	tests := map[string]func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller){
		"error from SDCFinder.GetSDCGuids": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			finder := mocks.NewMockSDCFinder(ctrl)
			client := mocks.NewMockPowerFlexClient(ctrl)

			finder.EXPECT().GetSDCGuids().Return(nil, errors.New("boom")).Times(1)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, finder, check(hasError), ctrl
		},

		"empty SDC GUIDs returns empty": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			finder := mocks.NewMockSDCFinder(ctrl)
			client := mocks.NewMockPowerFlexClient(ctrl)

			finder.EXPECT().GetSDCGuids().Return([]string{}, nil).Times(1)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, finder, check(noErrorAndLen(0)), ctrl
		},

		"error from client.GetInstance": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			finder := mocks.NewMockSDCFinder(ctrl)
			client := mocks.NewMockPowerFlexClient(ctrl)

			finder.EXPECT().GetSDCGuids().Return([]string{"g1"}, nil).Times(1)
			client.EXPECT().GetInstance("").Return(nil, errors.New("instances down")).Times(1)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, finder, check(hasError), ctrl
		},

		"error from service.SystemFinder": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			finder := mocks.NewMockSDCFinder(ctrl)
			client := mocks.NewMockPowerFlexClient(ctrl)

			finder.EXPECT().GetSDCGuids().Return([]string{"g1"}, nil).Times(1)
			client.EXPECT().GetInstance("").Return([]*types.System{{Name: "sys1", ID: "sid1"}}, nil).Times(1)

			// Patch SystemFinder to fail, so we return early before FindSystem/GetGenType
			patches := gomonkey.NewPatches()
			patches.ApplyFunc(service.SystemFinder, func(service.PowerFlexClient, string, string, string) (service.PowerFlexSystem, error) {
				return nil, errors.New("systemfinder oops")
			})
			t.Cleanup(patches.Reset)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, finder, check(hasError), ctrl
		},
		"error from service.GetGenType": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			finder := mocks.NewMockSDCFinder(ctrl)
			client := mocks.NewMockPowerFlexClient(ctrl)

			finder.EXPECT().GetSDCGuids().Return([]string{"g1"}, nil).Times(1)
			client.EXPECT().GetInstance("").Return([]*types.System{{Name: "sys1", ID: "sid1"}}, nil).Times(1)
			// client.EXPECT().FindSystem("sid1", "sys1", "").Return(&types.System{}, nil).Times(1)
			client.EXPECT().FindSystem("sid1", "sys1", "").Return((*sio.System)(nil), nil).Times(1)

			patches := gomonkey.NewPatches()
			patches.ApplyFunc(service.SystemFinder, func(service.PowerFlexClient, string, string, string) (service.PowerFlexSystem, error) {
				var sf service.PowerFlexSystem = &fakeSystemFinderTarget{byGUID: map[string]*sio.Sdc{}}
				return sf, nil
			})
			// patches.ApplyFunc(service.GetGenType, func(*types.System) (string, error) {
			// 	return "", errors.New("gen fail")
			// })
			patches.ApplyFunc(service.GetGenType, func(*sio.System) (string, error) {
				return "", errors.New("gen fail")
			})
			t.Cleanup(patches.Reset)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, finder, check(hasError), ctrl
		},

		"SDC not found is skipped; found one is returned": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			finder := mocks.NewMockSDCFinder(ctrl)
			client := mocks.NewMockPowerFlexClient(ctrl)

			finder.EXPECT().GetSDCGuids().Return([]string{"g1", "g2"}, nil).Times(1)
			client.EXPECT().GetInstance("").Return([]*types.System{{Name: "sys1", ID: "sid1"}}, nil).Times(1)
			// client.EXPECT().FindSystem("sid1", "sys1", "").Return(&types.System{}, nil).Times(1)
			client.EXPECT().FindSystem("sid1", "sys1", "").Return((*sio.System)(nil), nil).Times(1)

			patches := gomonkey.NewPatches()
			patches.ApplyFunc(service.SystemFinder, func(service.PowerFlexClient, string, string, string) (service.PowerFlexSystem, error) {
				var sf service.PowerFlexSystem = &fakeSystemFinderTarget{
					byGUID: map[string]*sio.Sdc{
						"g2": {Sdc: &types.Sdc{SdcGUID: "g2", ID: "sdc-id-2", SdcIP: "1.2.3.5"}},
					},
				}
				return sf, nil
			})
			// patches.ApplyFunc(service.GetGenType, func(*types.System) (string, error) {
			// 	return "v1", nil
			// })
			patches.ApplyFunc(service.GetGenType, func(*sio.System) (string, error) {
				return "v1", nil
			})
			t.Cleanup(patches.Reset)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, finder, check(noErrorAndLen(1)), ctrl
		},

		"multiple systems and GUIDs → accumulates": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, service.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			finder := mocks.NewMockSDCFinder(ctrl)
			client := mocks.NewMockPowerFlexClient(ctrl)

			finder.EXPECT().GetSDCGuids().Return([]string{"g1", "g2"}, nil).Times(1)
			client.EXPECT().GetInstance("").Return([]*types.System{
				{Name: "sys1", ID: "sid1"},
				{Name: "sys2", ID: "sid2"},
			}, nil).Times(1)
			// client.EXPECT().FindSystem("sid1", "sys1", "").Return(&types.System{}, nil).Times(1)
			client.EXPECT().FindSystem("sid1", "sys1", "").Return((*sio.System)(nil), nil).Times(1)
			// client.EXPECT().FindSystem("sid2", "sys2", "").Return(&types.System{}, nil).Times(1)
			client.EXPECT().FindSystem("sid2", "sys2", "").Return((*sio.System)(nil), nil).Times(1)

			patches := gomonkey.NewPatches()
			patches.ApplyFunc(service.SystemFinder, func(service.PowerFlexClient, string, string, string) (service.PowerFlexSystem, error) {
				var sf service.PowerFlexSystem = &fakeSystemFinderTarget{
					byGUID: map[string]*sio.Sdc{
						"g1": {Sdc: &types.Sdc{SdcGUID: "g1", ID: "sdc-id-1", SdcIP: "1.2.3.4"}},
						"g2": {Sdc: &types.Sdc{SdcGUID: "g2", ID: "sdc-id-2", SdcIP: "1.2.3.5"}},
					},
				}
				return sf, nil
			})
			// patches.ApplyFunc(service.GetGenType, func(*types.System) (string, error) {
			// 	return "v1", nil
			// })
			patches.ApplyFunc(service.GetGenType, func(*sio.System) (string, error) {
				return "v1", nil
			})
			t.Cleanup(patches.Reset)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			// 2 systems × 2 GUIDs → 4 retrievers
			return svc, client, finder, check(noErrorAndLen(4)), ctrl
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			svc, client, finder, checks, ctrl := tc(t)
			defer ctrl.Finish()
			// NB: ctx unused by GetSDCs internally, but pass a real context.Background()
			got, err := svc.GetSDCs(context.Background(), client, finder)
			for _, c := range checks {
				c(t, got, err)
			}
		})
	}

	// optional: ensure we patched function symbols of correct kind, similar to your style of sanity checks
	assert.Equal(t, reflect.Func, reflect.ValueOf(service.SystemFinder).Kind())
	assert.Equal(t, reflect.Func, reflect.ValueOf(service.GetGenType).Kind())
}

func Test_GetSDCMeta(t *testing.T) {
	type checkFn func(*testing.T, *service.SDCMeta, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	checkSdcMeta := func(expectedOutput *service.SDCMeta) func(t *testing.T, sdcMeta *service.SDCMeta, err error) {
		return func(t *testing.T, sdcMeta *service.SDCMeta, err error) {
			require.NoError(t, err)
			assert.Equal(t, expectedOutput, sdcMeta)
		}
	}

	tests := map[string]func(t *testing.T) (*sio.Sdc, []corev1.Node, []checkFn){
		"success": func(*testing.T) (*sio.Sdc, []corev1.Node, []checkFn) {
			sdc := &sio.Sdc{
				Sdc: &types.Sdc{
					SdcIP:   "1.2.3.4",
					ID:      "sdc-id-123",
					SdcGUID: "guid-xyz-789",
				},
			}
			nodes := []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node1"},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{{Address: "1.2.3.4"}},
					},
				},
			}
			expectedOutput := &service.SDCMeta{
				Name:    "node1",
				ID:      "sdc-id-123",
				IP:      "1.2.3.4",
				SdcGUID: "guid-xyz-789",
			}
			return sdc, nodes, check(checkSdcMeta(expectedOutput))
		},
		"no-match": func(*testing.T) (*sio.Sdc, []corev1.Node, []checkFn) {
			sdc := &sio.Sdc{
				Sdc: &types.Sdc{
					SdcIP:   "5.6.7.8",
					ID:      "sdc-id-456",
					SdcGUID: "guid-abc-123",
				},
			}
			nodes := []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node1"},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{{Address: "1.2.3.4"}},
					},
				},
			}
			expectedOutput := &service.SDCMeta{
				Name:    "",
				ID:      "sdc-id-456",
				IP:      "5.6.7.8",
				SdcGUID: "guid-abc-123",
			}
			return sdc, nodes, check(checkSdcMeta(expectedOutput))
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			sdc, nodes, checks := tc(t)
			sdcMeta, err := service.GetSDCMeta(sdc, nodes)
			for _, c := range checks {
				c(t, sdcMeta, err)
			}
		})
	}

	t.Run("nil sdc", func(t *testing.T) {
		meta, err := service.GetSDCMeta(nil, nil)
		require.Error(t, err)
		assert.Nil(t, meta)
	})

	t.Run("unsupported sdc type", func(t *testing.T) {
		type bogus struct{}
		meta, err := service.GetSDCMeta(bogus{}, nil) // anything not *sio.Sdc triggers default
		require.Error(t, err)
		assert.Nil(t, meta)
		assert.Contains(t, err.Error(), "unsupported sdc type")
	})
}

func Test_GetStorageClasses(t *testing.T) {
	type checkFn func(*testing.T, []service.StorageClassMeta, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, _ []service.StorageClassMeta, err error) {
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	checkPoolLength := func(class string, length int) func(t *testing.T, classes []service.StorageClassMeta, err error) {
		return func(t *testing.T, classes []service.StorageClassMeta, _ error) {
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

	hasError := func(t *testing.T, _ []service.StorageClassMeta, err error) {
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
						ID:   "123",
						Name: "test",
					},
				}, nil)

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

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]k8s.StorageClass{sc1}, nil)

			storageClassFinder.EXPECT().GetStoragePools(sc1).Times(1).
				Return([]string{"pool-1"})

			powerflexClient.EXPECT().GetStoragePool("").Times(1).
				Return([]*types.StoragePool{
					{
						ID:      "pool-id-1",
						Name:    "pool-1",
						GenType: "v1",
					},
					{
						ID:      "pool-id-2",
						Name:    "pool-2",
						GenType: "v1",
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
					{ID: "1234", Name: "sys-a"},
					{ID: "5678", Name: "sys-b"},
				}, nil)

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
				SystemID: "1234",
			}

			sc2 := k8s.StorageClass{
				StorageClass: v1.StorageClass{
					ObjectMeta: metav1.ObjectMeta{
						UID:  "123",
						Name: "class-1-xfs",
					},
					Provisioner: "csi-vxflexos.dellemc.com",
					Parameters: map[string]string{
						"storagepool": "pool-1",
					},
				},
				SystemID: "5678",
			}

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]k8s.StorageClass{sc1, sc2}, nil)

			storageClassFinder.EXPECT().GetStoragePools(sc1).Times(1).
				Return([]string{"pool-1"})
			storageClassFinder.EXPECT().GetStoragePools(sc2).Times(1).
				Return([]string{"pool-1"})

			powerflexClient.EXPECT().GetStoragePool("").Times(1).
				Return([]*types.StoragePool{
					{
						ID:      "pool-id-1",
						Name:    "pool-1",
						GenType: "v1",
					},
					{
						ID:      "pool-id-2",
						Name:    "pool-2",
						GenType: "v1",
					},
				}, nil)

			return powerflexClient, storageClassFinder,
				check(hasNoError, checkPoolLength("class-1", 1), checkPoolLength("class-1-xfs", 1)),
				ctrl
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

			sc2 := k8s.StorageClass{
				StorageClass: v1.StorageClass{
					ObjectMeta: metav1.ObjectMeta{
						UID:  "123",
						Name: "class-1-xfs",
					},
					Provisioner: "csi-vxflexos.dellemc.com",
					Parameters: map[string]string{
						"storagepool": "pool-1",
					},
				},
				SystemID: "5678",
			}

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]k8s.StorageClass{sc1, sc2}, nil)

			powerflexClient.EXPECT().GetInstance("").Times(1).
				Return(nil, errors.New("error"))

			return powerflexClient, storageClassFinder, check(hasError), ctrl
		},
		"error calling GetStoragePool": func(*testing.T) (service.PowerFlexClient, service.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			storageClassFinder := mocks.NewMockStorageClassFinder(ctrl)

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
				SystemID: "1234",
			}

			sc2 := k8s.StorageClass{
				StorageClass: v1.StorageClass{
					ObjectMeta: metav1.ObjectMeta{
						UID:  "123",
						Name: "class-1-xfs",
					},
					Provisioner: "csi-vxflexos.dellemc.com",
					Parameters: map[string]string{
						"storagepool": "pool-1",
					},
				},
				SystemID: "5678",
			}

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]k8s.StorageClass{sc1, sc2}, nil)

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
				SystemID: "1234",
			}

			sc2 := k8s.StorageClass{
				StorageClass: v1.StorageClass{
					ObjectMeta: metav1.ObjectMeta{
						UID:  "123",
						Name: "class-1-xfs",
					},
					Provisioner: "csi-vxflexos.dellemc.com",
					Parameters: map[string]string{
						"storagepool": "pool-1",
					},
				},
				SystemID: "5678",
			}

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]k8s.StorageClass{sc1, sc2}, nil)

			powerflexClient.EXPECT().GetInstance("").Times(1).
				Return([]*types.System{}, nil)

			return powerflexClient, storageClassFinder, check(hasError), ctrl
		},
		"calling GetStorageClasses returns 0 classes": func(*testing.T) (service.PowerFlexClient, service.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			powerflexClient := mocks.NewMockPowerFlexClient(ctrl)
			storageClassFinder := mocks.NewMockStorageClassFinder(ctrl)

			storageClassFinder.EXPECT().GetStorageClasses().Times(1).
				Return([]k8s.StorageClass{}, nil)

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
					Meter: otel.Meter("powerflex/sdc"),
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

			pool1 := newPoolRetriever(t, ctrl, sp1, "v1")
			pool2 := newPoolRetriever(t, ctrl, sp2, "v1")

			scMetas := []service.StorageClassMeta{
				{
					ID:              "123",
					Name:            "class-1",
					Driver:          "driver",
					StorageSystemID: "system1",
					StoragePools: map[string]service.StoragePoolMetricsRetriever{
						"poolID-1": pool1,
						"poolID-2": pool2,
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

			pool1 := newPoolRetriever(t, ctrl, sp1, "v1")
			pool2 := newPoolRetriever(t, ctrl, sp2, "v1")

			scMetas := []service.StorageClassMeta{
				{
					ID:              "123",
					Name:            "class-1",
					Driver:          "driver",
					StorageSystemID: "system1",
					StoragePools: map[string]service.StoragePoolMetricsRetriever{
						"poolID-1": pool1,
						"poolID-2": pool2,
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

			pool1 := newPoolRetriever(t, ctrl, sp1, "v1")

			scMetas := []service.StorageClassMeta{
				{
					ID:              "123",
					Name:            "class-1",
					Driver:          "driver",
					StorageSystemID: "system1",
					StoragePools: map[string]service.StoragePoolMetricsRetriever{
						"poolID-1": pool1,
					},
				},
			}

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().RecordCapacity(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("error"))
			return setup{
				Service: &service,
			}, scMetas, ctrl
		},
		"EC metrics path success": func(*testing.T) (setup, []service.StorageClassMeta, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			pfClient := mocks.NewMockPowerFlexClient(ctrl)

			// Build one EC storage pool retriever (poolID-ec)
			// Stats getter is not used for EC path, but we provide a mock to satisfy interface.
			spStats := mocks.NewMockStoragePoolStatisticsGetter(ctrl)

			// Metrics payload for EC path (bytes → will be converted to GiB inside gatherPoolStatistics)
			giB := float64(1 << 30)
			pfClient.EXPECT().
				GetMetrics("storage_pool", []string{"poolID-ec"}).
				Return(&types.MetricsResponse{
					Resources: []types.Resource{
						{
							ID: "poolID-ec",
							Metrics: []types.Metric{
								{Name: "physical_total", Values: []float64{100 * giB}},
								{Name: "physical_free", Values: []float64{40 * giB}},
								{Name: "physical_used", Values: []float64{60 * giB}},
								{Name: "logical_provisioned", Values: []float64{120 * giB}},
							},
						},
					},
				}, nil).
				Times(1)

			// Our EC retriever (returns GenTypeEC and the mocked client)
			ecPool := ecPoolRetriever{
				client: pfClient,
				stats:  spStats,
				gen:    types.GenTypeEC,
			}

			scMetas := []service.StorageClassMeta{
				{
					ID:              "123",
					Name:            "class-ec",
					Driver:          "driver",
					StorageSystemID: "system-ec",
					StoragePools: map[string]service.StoragePoolMetricsRetriever{
						"poolID-ec": ecPool,
					},
				},
			}

			// One pool → expect one RecordCapacity call (we don't assert its values here)
			metrics.EXPECT().
				RecordCapacity(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(1)

			svc := service.PowerFlexService{MetricsWrapper: metrics}
			return setup{Service: &svc}, scMetas, ctrl
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

	ctrl := gomock.NewController(b) // gomock accepts testing.TB, so *testing.B is fine
	b.Cleanup(ctrl.Finish)

	metrics := mocks.NewMockMetricsRecorder(ctrl)

	// Build StoragePoolStatisticsGetter mocks
	stats := make(map[string]service.StoragePoolStatisticsGetter, numOfPools)
	for i := 0; i < numOfPools; i++ {
		sg := mocks.NewMockStoragePoolStatisticsGetter(ctrl)
		sg.EXPECT().GetStatistics().DoAndReturn(func() (*types.Statistics, error) {
			d, _ := time.ParseDuration(poolQueryTime)
			time.Sleep(d)
			return &types.Statistics{}, nil
		}).AnyTimes()
		stats["poolID-"+strconv.Itoa(i)] = sg
	}

	// Wrap into StoragePoolMetricsRetriever
	retr := make(map[string]service.StoragePoolMetricsRetriever, numOfPools)
	for k, sg := range stats {
		retr[k] = newPoolRetriever(b, ctrl, sg, "v1") // <-- b implements testing.TB
	}

	scMetas := []service.StorageClassMeta{
		{
			ID:              "123",
			Name:            "class-1",
			Driver:          "driver",
			StorageSystemID: "system1",
			StoragePools:    retr,
		},
	}

	metrics.EXPECT().
		RecordCapacity(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()

	svc := service.PowerFlexService{MetricsWrapper: metrics}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GetStoragePoolStatistics(context.Background(), scMetas)
	}
}

func Test_GetVolumes(t *testing.T) {
	type checkFn func(t *testing.T, out []*service.VolumeMetaMetrics, err error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasError := func(t *testing.T, _ []*service.VolumeMetaMetrics, err error) {
		require.Error(t, err)
	}
	noErrorAndLen := func(n int) checkFn {
		return func(t *testing.T, out []*service.VolumeMetaMetrics, err error) {
			require.NoError(t, err)
			assert.Len(t, out, n)
		}
	}
	type ecVals struct {
		readBW, writeBW, readIOPS, writeIOPS, readLat, writeLat float64
	}
	checkECValues := func(expect map[string]ecVals) checkFn {
		return func(t *testing.T, out []*service.VolumeMetaMetrics, err error) {
			require.NoError(t, err)
			seen := map[string]*service.VolumeMetaMetrics{}
			for _, m := range out {
				seen[m.ID] = m
			}
			for id, exp := range expect {
				got, ok := seen[id]
				require.True(t, ok, "expected volume %s in results", id)

				assert.InDelta(t, exp.readBW, got.HostReadBandwith, 1e-6, "read BW for %s", id)
				assert.InDelta(t, exp.writeBW, got.HostWriteBandwith, 1e-6, "write BW for %s", id)
				assert.InDelta(t, exp.readIOPS, got.HostReadIOPS, 1e-6, "read IOPS for %s", id)
				assert.InDelta(t, exp.writeIOPS, got.HostWriteIOPS, 1e-6, "write IOPS for %s", id)
				assert.InDelta(t, exp.readLat, got.AvgHostReadLatency, 1e-6, "read latency for %s", id)
				assert.InDelta(t, exp.writeLat, got.AvgHostWriteLatency, 1e-6, "write latency for %s", id)
				assert.Equal(t, types.GenTypeEC, got.GenType, "GenType for %s", id)
			}
		}
	}

	mkMapped := func(ids ...string) []*types.MappedSdcInfo {
		out := make([]*types.MappedSdcInfo, 0, len(ids))
		for _, id := range ids {
			out = append(out, &types.MappedSdcInfo{SdcID: id, SdcIP: "10.0.0." + id})
		}
		return out
	}

	vol1 := &sio.Volume{Volume: &types.Volume{ID: "1", Name: "vol-1", GenType: "", MappedSdcInfo: mkMapped("60001")}}
	vol2 := &sio.Volume{Volume: &types.Volume{ID: "2", Name: "vol-2", GenType: "", MappedSdcInfo: mkMapped("60001", "60002")}}

	ecVolNoID := &sio.Volume{Volume: &types.Volume{ID: "", Name: "ec-empty", GenType: types.GenTypeEC}}
	ecVol1 := &sio.Volume{Volume: &types.Volume{ID: "ec1", Name: "ec-1", GenType: types.GenTypeEC}}
	ecVol2 := &sio.Volume{Volume: &types.Volume{ID: "ec2", Name: "ec-2", GenType: types.GenTypeEC}}

	bwc := types.BWC{NumOccured: 100, NumSeconds: 10, TotalWeightInKb: 2048}
	vm1 := &types.SdcVolumeMetrics{
		VolumeID:        "1",
		ReadBwc:         bwc,
		WriteBwc:        bwc,
		ReadLatencyBwc:  bwc,
		WriteLatencyBwc: bwc,
		TrimBwc:         bwc,
		TrimLatencyBwc:  bwc,
	}
	vm2 := &types.SdcVolumeMetrics{
		VolumeID:        "2",
		ReadBwc:         bwc,
		WriteBwc:        bwc,
		ReadLatencyBwc:  bwc,
		WriteLatencyBwc: bwc,
		TrimBwc:         bwc,
		TrimLatencyBwc:  bwc,
	}

	tests := map[string]func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, []service.SdcMetricsRetriever, []checkFn, *gomock.Controller, *gomonkey.Patches){
		"error from FindVolumes": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, []service.SdcMetricsRetriever, []checkFn, *gomock.Controller, *gomonkey.Patches) {
			ctrl := gomock.NewController(t)
			client := mocks.NewMockPowerFlexClient(ctrl)

			stats := mocks.NewMockStatisticsGetter(ctrl)
			sdc := &sio.Sdc{Sdc: &types.Sdc{ID: "sdc-1", SdcIP: "1.1.1.1"}}
			retr := newSdcRetriever(t, ctrl, stats, "v1", sdc)

			patches := gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(&sio.Sdc{}), "FindVolumes", func(_ *sio.Sdc) ([]*sio.Volume, error) {
				return nil, errors.New("find-volumes-fail")
			})
			t.Cleanup(patches.Reset)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, []service.SdcMetricsRetriever{retr}, check(hasError), ctrl, patches
		},

		"non-EC: success, metrics mapped and volumes de-duplicated across SDCs": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, []service.SdcMetricsRetriever, []checkFn, *gomock.Controller, *gomonkey.Patches) {
			ctrl := gomock.NewController(t)
			client := mocks.NewMockPowerFlexClient(ctrl)

			stats1 := mocks.NewMockStatisticsGetter(ctrl)
			stats2 := mocks.NewMockStatisticsGetter(ctrl)

			sdc1 := &sio.Sdc{Sdc: &types.Sdc{ID: "sdc-1", SdcIP: "1.1.1.1"}}
			sdc2 := &sio.Sdc{Sdc: &types.Sdc{ID: "sdc-2", SdcIP: "1.1.1.2"}}

			stats1.EXPECT().GetVolumeMetrics().Return([]*types.SdcVolumeMetrics{vm1, vm2}, nil).Times(1)
			stats2.EXPECT().GetVolumeMetrics().Return([]*types.SdcVolumeMetrics{vm1, vm2}, nil).Times(1)

			r1 := newSdcRetriever(t, ctrl, stats1, "v1", sdc1)
			r2 := newSdcRetriever(t, ctrl, stats2, "v1", sdc2)

			mapping := sdcToVolumes{
				sdc1: []*sio.Volume{vol1, vol2},
				sdc2: []*sio.Volume{vol1},
			}
			patches := gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(&sio.Sdc{}), "FindVolumes", func(s *sio.Sdc) ([]*sio.Volume, error) {
				return mapping[s], nil
			})
			t.Cleanup(patches.Reset)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, []service.SdcMetricsRetriever{r1, r2}, check(noErrorAndLen(2)), ctrl, patches
		},

		"non-EC: GetVolumeMetrics error": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, []service.SdcMetricsRetriever, []checkFn, *gomock.Controller, *gomonkey.Patches) {
			ctrl := gomock.NewController(t)
			client := mocks.NewMockPowerFlexClient(ctrl)

			stats := mocks.NewMockStatisticsGetter(ctrl)
			sdc := &sio.Sdc{Sdc: &types.Sdc{ID: "sdc-1", SdcIP: "1.1.1.1"}}
			r := newSdcRetriever(t, ctrl, stats, "v1", sdc)

			patches := gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(&sio.Sdc{}), "FindVolumes", func(_ *sio.Sdc) ([]*sio.Volume, error) {
				return []*sio.Volume{vol1}, nil
			})
			t.Cleanup(patches.Reset)

			stats.EXPECT().GetVolumeMetrics().Return(nil, errors.New("metrics-fail")).Times(1)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, []service.SdcMetricsRetriever{r}, check(hasError), ctrl, patches
		},

		"EC: no valid IDs -> early return (no error, empty result)": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, []service.SdcMetricsRetriever, []checkFn, *gomock.Controller, *gomonkey.Patches) {
			ctrl := gomock.NewController(t)
			client := mocks.NewMockPowerFlexClient(ctrl)

			stats := mocks.NewMockStatisticsGetter(ctrl)
			sdc := &sio.Sdc{Sdc: &types.Sdc{ID: "sdc-ec", SdcIP: "1.1.1.3"}}
			r := newSdcRetriever(t, ctrl, stats, "v1", sdc)

			patches := gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(&sio.Sdc{}), "FindVolumes", func(_ *sio.Sdc) ([]*sio.Volume, error) {
				return []*sio.Volume{ecVolNoID, ecVolNoID}, nil
			})
			t.Cleanup(patches.Reset)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, []service.SdcMetricsRetriever{r}, check(noErrorAndLen(0)), ctrl, patches
		},

		"EC: GetMetrics error": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, []service.SdcMetricsRetriever, []checkFn, *gomock.Controller, *gomonkey.Patches) {
			ctrl := gomock.NewController(t)
			client := mocks.NewMockPowerFlexClient(ctrl)

			stats := mocks.NewMockStatisticsGetter(ctrl)
			sdc := &sio.Sdc{Sdc: &types.Sdc{ID: "sdc-ec-err", SdcIP: "1.1.1.9"}}
			r := newSdcRetriever(t, ctrl, stats, "v1", sdc)

			patches := gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(&sio.Sdc{}), "FindVolumes", func(_ *sio.Sdc) ([]*sio.Volume, error) {
				return []*sio.Volume{ecVol1, ecVol2}, nil
			})
			t.Cleanup(patches.Reset)

			client.EXPECT().
				GetMetrics("volume", []string{"ec1", "ec2"}).
				Return((*types.MetricsResponse)(nil), fmt.Errorf("metrics failed")).
				Times(1)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, []service.SdcMetricsRetriever{r}, check(hasError), ctrl, patches
		},

		"EC: empty resources -> returns no-volume-metrics error": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, []service.SdcMetricsRetriever, []checkFn, *gomock.Controller, *gomonkey.Patches) {
			ctrl := gomock.NewController(t)
			client := mocks.NewMockPowerFlexClient(ctrl)

			stats := mocks.NewMockStatisticsGetter(ctrl)
			sdc := &sio.Sdc{Sdc: &types.Sdc{ID: "sdc-ec-empty", SdcIP: "1.1.1.10"}}
			r := newSdcRetriever(t, ctrl, stats, "v1", sdc)

			patches := gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(&sio.Sdc{}), "FindVolumes", func(_ *sio.Sdc) ([]*sio.Volume, error) {
				return []*sio.Volume{ecVol1, ecVol2}, nil
			})
			t.Cleanup(patches.Reset)

			client.EXPECT().
				GetMetrics("volume", []string{"ec1", "ec2"}).
				Return(&types.MetricsResponse{Resources: []types.Resource{}}, nil).
				Times(1)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			return svc, client, []service.SdcMetricsRetriever{r}, check(hasError), ctrl, patches
		},

		"EC: metrics success -> populates & normalizes values": func(t *testing.T) (*service.PowerFlexService, service.PowerFlexClient, []service.SdcMetricsRetriever, []checkFn, *gomock.Controller, *gomonkey.Patches) {
			ctrl := gomock.NewController(t)
			client := mocks.NewMockPowerFlexClient(ctrl)

			stats := mocks.NewMockStatisticsGetter(ctrl)
			sdc := &sio.Sdc{Sdc: &types.Sdc{ID: "sdc-ec", SdcIP: "1.1.1.11"}}
			r := newSdcRetriever(t, ctrl, stats, "v1", sdc)

			patches := gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(&sio.Sdc{}), "FindVolumes", func(_ *sio.Sdc) ([]*sio.Volume, error) {
				ecEmpty := &sio.Volume{Volume: &types.Volume{ID: "", Name: "ec-empty-x", GenType: types.GenTypeEC}}
				return []*sio.Volume{ecVol1, ecVol2, ecEmpty}, nil
			})
			t.Cleanup(patches.Reset)
			client.EXPECT().
				GetMetrics("volume", []string{"ec1", "ec2"}).
				Return(&types.MetricsResponse{
					Resources: []types.Resource{
						{
							ID: "ec1",
							Metrics: []types.Metric{
								{Name: "host_read_bandwidth", Values: []float64{1048576}},
								{Name: "host_write_bandwidth", Values: []float64{2097152}},
								{Name: "host_read_iops", Values: []float64{111}},
								{Name: "host_write_iops", Values: []float64{222}},
								{Name: "avg_host_read_latency", Values: []float64{5000}},
								{Name: "avg_host_write_latency", Values: []float64{7000}},
							},
						},
						{
							ID: "ec2",
							Metrics: []types.Metric{
								{Name: "host_read_bandwidth", Values: []float64{3145728}},
								{Name: "host_write_bandwidth", Values: []float64{4194304}},
								{Name: "host_read_iops", Values: []float64{333}},
								{Name: "host_write_iops", Values: []float64{444}},
								{Name: "avg_host_read_latency", Values: []float64{9000}},
								{Name: "avg_host_write_latency", Values: []float64{11000}},
							},
						},
					},
				}, nil).
				Times(1)

			svc := &service.PowerFlexService{Logger: logrus.New()}
			expect := map[string]ecVals{
				"ec1": {readBW: 1.0, writeBW: 2.0, readIOPS: 111, writeIOPS: 222, readLat: 5.0, writeLat: 7.0},
				"ec2": {readBW: 3.0, writeBW: 4.0, readIOPS: 333, writeIOPS: 444, readLat: 9.0, writeLat: 11.0},
			}
			checkHasEmptyID := func() checkFn {
				return func(t *testing.T, out []*service.VolumeMetaMetrics, err error) {
					require.NoError(t, err)
					found := false
					for _, m := range out {
						if m.ID == "" {
							found = true
							assert.Zero(t, m.HostReadBandwith)
							assert.Zero(t, m.HostWriteBandwith)
							assert.Zero(t, m.HostReadIOPS)
							assert.Zero(t, m.HostWriteIOPS)
							assert.Zero(t, m.AvgHostReadLatency)
							assert.Zero(t, m.AvgHostWriteLatency)
							assert.Empty(t, m.GenType)
							break
						}
					}
					assert.True(t, found, "expected an empty-ID volume in results")
				}
			}
			return svc, client, []service.SdcMetricsRetriever{r}, check(noErrorAndLen(3), checkECValues(expect), checkHasEmptyID()), ctrl, patches
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			svc, client, sdcs, checks, ctrl, patches := tc(t)
			defer ctrl.Finish()
			if patches != nil {
				defer patches.Reset()
			}
			out, err := svc.GetVolumes(context.Background(), client, sdcs)
			for _, c := range checks {
				c(t, out, err)
			}
		})
	}
}

func Test_ExportVolumeStatistics(t *testing.T) {
	type setup struct {
		Service *service.PowerFlexService
	}

	tests := map[string]func(t *testing.T) (setup, []*service.VolumeMetaMetrics, service.VolumeFinder, *gomock.Controller){
		"success": func(*testing.T) (setup, []*service.VolumeMetaMetrics, service.VolumeFinder, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			vol1 := &service.VolumeMetaMetrics{
				ID: "vol1",
			}
			vol2 := &service.VolumeMetaMetrics{
				ID: "vol2",
			}
			vol3 := &service.VolumeMetaMetrics{
				ID: "vol3",
			}

			vols := []*service.VolumeMetaMetrics{vol1, vol2, vol3}

			volFinder.EXPECT().GetPersistentVolumes().Return([]k8s.VolumeInfo{}, nil)

			service := service.PowerFlexService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
			return setup{
				Service: &service,
			}, vols, volFinder, ctrl
		},
		"nil list of vols": func(*testing.T) (setup, []*service.VolumeMetaMetrics, service.VolumeFinder, *gomock.Controller) {
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
		"error recording": func(*testing.T) (setup, []*service.VolumeMetaMetrics, service.VolumeFinder, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			vol1 := &service.VolumeMetaMetrics{
				ID: "vol1",
			}
			vols := []*service.VolumeMetaMetrics{vol1}

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

func Benchmark_GetVolumes(b *testing.B) {
	numOfSDCs, sdcQueryTime := 500, "100ms"
	b.Logf("For %d SDCs and assuming each sdc query takes %s\n", numOfSDCs, sdcQueryTime)
	b.ReportAllocs()

	ctrl := gomock.NewController(b)
	b.Cleanup(ctrl.Finish)

	client := mocks.NewMockPowerFlexClient(ctrl)

	svc := &service.PowerFlexService{
		Logger:         logrus.New(),
		MetricsWrapper: mocks.NewMockMetricsRecorder(ctrl),
	}

	vols := []*sio.Volume{
		{Volume: &types.Volume{ID: "vol1", Name: "vol1", GenType: ""}},
		{Volume: &types.Volume{ID: "vol2", Name: "vol2", GenType: ""}},
	}
	patches := gomonkey.NewPatches()
	patches.ApplyMethod(reflect.TypeOf(&sio.Sdc{}), "FindVolumes", func(_ *sio.Sdc) ([]*sio.Volume, error) {
		return vols, nil
	})
	b.Cleanup(patches.Reset)

	retrievers := make([]service.SdcMetricsRetriever, 0, numOfSDCs)
	for i := 0; i < numOfSDCs; i++ {
		sdcStats := mocks.NewMockStatisticsGetter(gomock.NewController(b))
		sdcStats.EXPECT().GetVolumeMetrics().DoAndReturn(func() ([]*types.SdcVolumeMetrics, error) {
			dur, _ := time.ParseDuration(sdcQueryTime)
			time.Sleep(dur)
			return []*types.SdcVolumeMetrics{
				{VolumeID: "vol1"},
				{VolumeID: "vol2"},
			}, nil
		}).AnyTimes()

		sdcObj := &sio.Sdc{Sdc: &types.Sdc{ID: fmt.Sprintf("sdc-%d", i), SdcIP: "1.2.3.4"}}

		retr := newSdcRetriever(b, ctrl, sdcStats, "v1", sdcObj)
		retrievers = append(retrievers, retr)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.GetVolumes(context.Background(), client, retrievers)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Test_GetVolumeBandwidth(t *testing.T) {
	tt := []struct {
		Name                   string
		Statistics             *service.VolumeMetaMetrics
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
			&service.VolumeMetaMetrics{},
			0.0,
			0.0,
		},
		{
			"only read bandwidth",
			&service.VolumeMetaMetrics{
				ReadBwc: types.BWC{TotalWeightInKb: 392040, NumSeconds: 110},
			},
			3.48046875,
			0.0,
		},
		{
			"only write bandwidth",
			&service.VolumeMetaMetrics{
				WriteBwc: types.BWC{TotalWeightInKb: 1958128, NumSeconds: 313},
			},
			0.0,
			6.109375,
		},
		{
			"read and write bandwidth",
			&service.VolumeMetaMetrics{
				ReadBwc:  types.BWC{TotalWeightInKb: 1546272, NumSeconds: 236},
				WriteBwc: types.BWC{TotalWeightInKb: 12838, NumSeconds: 131},
			},
			6.3984375,
			0.095703125,
		},
		{
			"EC gen type uses host bandwidths",
			&service.VolumeMetaMetrics{
				GenType:           types.GenTypeEC,
				HostReadBandwith:  12.75,
				HostWriteBandwith: 34.5,
				ReadBwc:           types.BWC{TotalWeightInKb: 999999, NumSeconds: 1},
				WriteBwc:          types.BWC{TotalWeightInKb: 888888, NumSeconds: 2},
			},
			12.75,
			34.5,
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
		Statistics        *service.VolumeMetaMetrics
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
			&service.VolumeMetaMetrics{},
			0.0,
			0.0,
		},
		{
			"only read IOPS",
			&service.VolumeMetaMetrics{
				ReadBwc: types.BWC{NumOccured: 6856870, NumSeconds: 114},
			},
			60147.982456,
			0.0,
		},
		{
			"only write IOPS",
			&service.VolumeMetaMetrics{
				WriteBwc: types.BWC{NumOccured: 354139516, NumSeconds: 3131},
			},
			0.0,
			113107.478760,
		},
		{
			"read and write IOPS",
			&service.VolumeMetaMetrics{
				ReadBwc:  types.BWC{NumOccured: 94729, NumSeconds: 236},
				WriteBwc: types.BWC{NumOccured: 68122431, NumSeconds: 131},
			},
			401.394068,
			520018.557251,
		},
		{
			"EC gen type uses host IOPS",
			&service.VolumeMetaMetrics{
				GenType:       types.GenTypeEC,
				HostReadIOPS:  12345.67,
				HostWriteIOPS: 76543.21,
				ReadBwc:       types.BWC{NumOccured: 999, NumSeconds: 1},
				WriteBwc:      types.BWC{NumOccured: 888, NumSeconds: 2},
			},
			12345.67,
			76543.21,
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
		Statistics           *service.VolumeMetaMetrics
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
			&service.VolumeMetaMetrics{},
			0.0,
			0.0,
		},
		{
			"only read latency",
			&service.VolumeMetaMetrics{
				ReadLatencyBwc: types.BWC{TotalWeightInKb: 6856870, NumOccured: 114},
			},
			58.738264,
			0.0,
		},
		{
			"only write latency",
			&service.VolumeMetaMetrics{
				WriteLatencyBwc: types.BWC{TotalWeightInKb: 354139516, NumOccured: 313},
			},
			0.0,
			1104.918119,
		},
		{
			"read and write latency",
			&service.VolumeMetaMetrics{
				ReadLatencyBwc:  types.BWC{TotalWeightInKb: 94729, NumOccured: 236},
				WriteLatencyBwc: types.BWC{TotalWeightInKb: 68122431, NumOccured: 131},
			},
			0.391986,
			507.830622,
		},
		{
			"EC gen type uses host latencies",
			&service.VolumeMetaMetrics{
				GenType:             types.GenTypeEC,
				AvgHostReadLatency:  12.345,
				AvgHostWriteLatency: 67.89,
				ReadLatencyBwc:      types.BWC{TotalWeightInKb: 999999, NumOccured: 1},
				WriteLatencyBwc:     types.BWC{TotalWeightInKb: 888888, NumOccured: 2},
			},
			12.345,
			67.89,
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

func TestExportTopologyMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVolumeFinder := mocks.NewMockVolumeFinder(ctrl)
	mockMetricsWrapper := mocks.NewMockMetricsRecorder(ctrl)

	s := &service.PowerFlexService{
		VolumeFinder:   mockVolumeFinder,
		MetricsWrapper: mockMetricsWrapper,
		Logger:         logrus.New(),
	}

	ctx := context.Background()
	volumes := []k8s.VolumeInfo{
		{
			Namespace:               "ns1",
			VolumeClaimName:         "pvc1",
			PersistentVolume:        "pv1",
			PersistentVolumeStatus:  "Bound",
			StorageClass:            "sc1",
			Driver:                  "csi-driver",
			ProvisionedSize:         "100Gi",
			StorageSystemVolumeName: "vol1",
			StoragePoolName:         "pool1",
			StorageSystem:           "sys1",
			Protocol:                "NFS",
			CreatedTime:             "2022-01-01T00:00:00Z",
			VolumeHandle:            "vol1-sys1",
		},
	}

	mockVolumeFinder.EXPECT().GetPersistentVolumes().Return(volumes, nil)
	mockMetricsWrapper.EXPECT().RecordTopologyMetrics(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	s.ExportTopologyMetrics(ctx)
}

func Test_GetGenType(t *testing.T) {
	t.Run("success - returns first PD GenType", func(t *testing.T) {
		sys := &sio.System{}
		pds := []*types.ProtectionDomain{
			{ID: "pd-1", Name: "pd1", GenType: "v2"},
			{ID: "pd-2", Name: "pd2", GenType: "v1"},
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyMethod(reflect.TypeOf(sys), "GetProtectionDomain",
			func(_ *sio.System, _ string) ([]*types.ProtectionDomain, error) {
				return pds, nil
			})

		got, err := service.GetGenType(sys)
		require.NoError(t, err)
		assert.Equal(t, "v2", got, "should return the GenType of the first PD")
	})

	t.Run("empty list - returns empty string with no error", func(t *testing.T) {
		sys := &sio.System{}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyMethod(reflect.TypeOf(sys), "GetProtectionDomain",
			func(_ *sio.System, _ string) ([]*types.ProtectionDomain, error) {
				return []*types.ProtectionDomain{}, nil
			})

		got, err := service.GetGenType(sys)
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("error from GetProtectionDomain is propagated", func(t *testing.T) {
		sys := &sio.System{}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyMethod(reflect.TypeOf(sys), "GetProtectionDomain",
			func(_ *sio.System, _ string) ([]*types.ProtectionDomain, error) {
				return nil, errors.New("pd call failed")
			})

		got, err := service.GetGenType(sys)
		require.Error(t, err)
		assert.Equal(t, "", got)
		assert.Contains(t, err.Error(), "pd call failed")
	})

	t.Run("nil system - panics", func(t *testing.T) {
		require.Panics(t, func() {
			_, _ = service.GetGenType(nil)
		})
	})
}
