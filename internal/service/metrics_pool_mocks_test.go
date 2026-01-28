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
	testingpkg "testing"

	"github.com/dell/karavi-metrics-powerflex/internal/service"
	"github.com/dell/karavi-metrics-powerflex/internal/service/mocks"

	"go.uber.org/mock/gomock"
)

type mockPoolMetricsRetriever struct {
	sg  service.StoragePoolStatisticsGetter
	c   service.PowerFlexClient
	gen string
}

func (m *mockPoolMetricsRetriever) GetStatisticsGetter() service.StoragePoolStatisticsGetter {
	return m.sg
}
func (m *mockPoolMetricsRetriever) GetClient() service.PowerFlexClient { return m.c }
func (m *mockPoolMetricsRetriever) GetGen() string                     { return m.gen }

var _ service.StoragePoolMetricsRetriever = (*mockPoolMetricsRetriever)(nil)

func newPoolRetriever(t testingpkg.TB, ctrl *gomock.Controller, sg service.StoragePoolStatisticsGetter, gen string) service.StoragePoolMetricsRetriever {
	t.Helper()
	pfClient := mocks.NewMockPowerFlexClient(ctrl)
	return &mockPoolMetricsRetriever{
		sg:  sg,
		c:   pfClient,
		gen: gen,
	}
}
