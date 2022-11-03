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

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/dell/karavi-metrics-powerflex/internal/service (interfaces: StatisticsGetter)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	goscaleio "github.com/dell/goscaleio"
	goscaleio0 "github.com/dell/goscaleio/types/v1"
	gomock "github.com/golang/mock/gomock"
)

// MockStatisticsGetter is a mock of StatisticsGetter interface.
type MockStatisticsGetter struct {
	ctrl     *gomock.Controller
	recorder *MockStatisticsGetterMockRecorder
}

// MockStatisticsGetterMockRecorder is the mock recorder for MockStatisticsGetter.
type MockStatisticsGetterMockRecorder struct {
	mock *MockStatisticsGetter
}

// NewMockStatisticsGetter creates a new mock instance.
func NewMockStatisticsGetter(ctrl *gomock.Controller) *MockStatisticsGetter {
	mock := &MockStatisticsGetter{ctrl: ctrl}
	mock.recorder = &MockStatisticsGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStatisticsGetter) EXPECT() *MockStatisticsGetterMockRecorder {
	return m.recorder
}

// FindVolumes mocks base method.
func (m *MockStatisticsGetter) FindVolumes() ([]*goscaleio.Volume, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindVolumes")
	ret0, _ := ret[0].([]*goscaleio.Volume)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindVolumes indicates an expected call of FindVolumes.
func (mr *MockStatisticsGetterMockRecorder) FindVolumes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindVolumes", reflect.TypeOf((*MockStatisticsGetter)(nil).FindVolumes))
}

// GetStatistics mocks base method.
func (m *MockStatisticsGetter) GetStatistics() (*goscaleio0.SdcStatistics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatistics")
	ret0, _ := ret[0].(*goscaleio0.SdcStatistics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatistics indicates an expected call of GetStatistics.
func (mr *MockStatisticsGetterMockRecorder) GetStatistics() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatistics", reflect.TypeOf((*MockStatisticsGetter)(nil).GetStatistics))
}

// GetVolume mocks base method.
func (m *MockStatisticsGetter) GetVolume() ([]*goscaleio0.Volume, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVolume")
	ret0, _ := ret[0].([]*goscaleio0.Volume)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVolume indicates an expected call of GetVolume.
func (mr *MockStatisticsGetterMockRecorder) GetVolume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVolume", reflect.TypeOf((*MockStatisticsGetter)(nil).GetVolume))
}
