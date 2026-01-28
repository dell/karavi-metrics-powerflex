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
	"testing"

	v1 "github.com/dell/goscaleio/types/v1"
	service "github.com/dell/karavi-metrics-powerflex/internal/service"
	"github.com/dell/karavi-metrics-powerflex/internal/service/mocks"
	"go.uber.org/mock/gomock"
)

type mockStatsGetter struct{}

func (m *mockStatsGetter) GetStatistics() (*v1.Statistics, error) {
	return &v1.Statistics{}, nil
}

type mockMetricsRetriever struct {
	sg service.StoragePoolStatisticsGetter
	c  service.PowerFlexClient
	g  string
}

func (m *mockMetricsRetriever) GetStatisticsGetter() service.StoragePoolStatisticsGetter { return m.sg }
func (m *mockMetricsRetriever) GetClient() service.PowerFlexClient                       { return m.c }
func (m *mockMetricsRetriever) GetGen() string                                           { return m.g }

var (
	_ service.StoragePoolStatisticsGetter = (*mockStatsGetter)(nil)
	_ service.StoragePoolMetricsRetriever = (*mockMetricsRetriever)(nil)
)

func newRetriever(t *testing.T, gen string) *mockMetricsRetriever {
	t.Helper()
	ctrl := gomock.NewController(t)
	pfClient := mocks.NewMockPowerFlexClient(ctrl)
	return &mockMetricsRetriever{
		sg: &mockStatsGetter{},
		c:  pfClient,
		g:  gen,
	}
}
