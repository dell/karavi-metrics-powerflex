/*
Copyright (c) 2021-2023 Dell Inc. or its subsidiaries. All Rights Reserved.

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
	testingpkg "testing"

	service "github.com/dell/karavi-metrics-powerflex/internal/service"
	"github.com/dell/karavi-metrics-powerflex/internal/service/mocks"

	sio "github.com/dell/goscaleio"
	"go.uber.org/mock/gomock"
)

type mockSdcMetricsRetriever struct {
	sg  service.StatisticsGetter
	gen string
	sdc *sio.Sdc
	c   service.PowerFlexClient
}

func (m *mockSdcMetricsRetriever) GetStatisticsGetter() service.StatisticsGetter { return m.sg }
func (m *mockSdcMetricsRetriever) GetGen() string                                { return m.gen }
func (m *mockSdcMetricsRetriever) GetSdc() *sio.Sdc                              { return m.sdc }
func (m *mockSdcMetricsRetriever) GetClient() service.PowerFlexClient            { return m.c }

var _ service.SdcMetricsRetriever = (*mockSdcMetricsRetriever)(nil)

func newSdcRetriever(
	t testingpkg.TB,
	ctrl *gomock.Controller,
	sg service.StatisticsGetter,
	gen string,
	sdc *sio.Sdc,
) service.SdcMetricsRetriever {
	t.Helper()
	pfClient := mocks.NewMockPowerFlexClient(ctrl)
	return &mockSdcMetricsRetriever{
		sg:  sg,
		gen: gen,
		sdc: sdc,
		c:   pfClient,
	}
}
