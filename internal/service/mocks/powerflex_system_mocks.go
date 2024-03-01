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
// Source: github.com/dell/karavi-metrics-powerflex/internal/service (interfaces: PowerFlexSystem)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	goscaleio "github.com/dell/goscaleio"
	gomock "github.com/golang/mock/gomock"
)

// MockPowerFlexSystem is a mock of PowerFlexSystem interface.
type MockPowerFlexSystem struct {
	ctrl     *gomock.Controller
	recorder *MockPowerFlexSystemMockRecorder
}

// MockPowerFlexSystemMockRecorder is the mock recorder for MockPowerFlexSystem.
type MockPowerFlexSystemMockRecorder struct {
	mock *MockPowerFlexSystem
}

// NewMockPowerFlexSystem creates a new mock instance.
func NewMockPowerFlexSystem(ctrl *gomock.Controller) *MockPowerFlexSystem {
	mock := &MockPowerFlexSystem{ctrl: ctrl}
	mock.recorder = &MockPowerFlexSystemMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPowerFlexSystem) EXPECT() *MockPowerFlexSystemMockRecorder {
	return m.recorder
}

// FindSdc mocks base method.
func (m *MockPowerFlexSystem) FindSdc(arg0, arg1 string) (*goscaleio.Sdc, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindSdc", arg0, arg1)
	ret0, _ := ret[0].(*goscaleio.Sdc)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindSdc indicates an expected call of FindSdc.
func (mr *MockPowerFlexSystemMockRecorder) FindSdc(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindSdc", reflect.TypeOf((*MockPowerFlexSystem)(nil).FindSdc), arg0, arg1)
}
