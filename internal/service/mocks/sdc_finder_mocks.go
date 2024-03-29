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
// Source: github.com/dell/karavi-metrics-powerflex/internal/service (interfaces: SDCFinder)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockSDCFinder is a mock of SDCFinder interface.
type MockSDCFinder struct {
	ctrl     *gomock.Controller
	recorder *MockSDCFinderMockRecorder
}

// MockSDCFinderMockRecorder is the mock recorder for MockSDCFinder.
type MockSDCFinderMockRecorder struct {
	mock *MockSDCFinder
}

// NewMockSDCFinder creates a new mock instance.
func NewMockSDCFinder(ctrl *gomock.Controller) *MockSDCFinder {
	mock := &MockSDCFinder{ctrl: ctrl}
	mock.recorder = &MockSDCFinderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSDCFinder) EXPECT() *MockSDCFinderMockRecorder {
	return m.recorder
}

// GetSDCGuids mocks base method.
func (m *MockSDCFinder) GetSDCGuids() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSDCGuids")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSDCGuids indicates an expected call of GetSDCGuids.
func (mr *MockSDCFinderMockRecorder) GetSDCGuids() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSDCGuids", reflect.TypeOf((*MockSDCFinder)(nil).GetSDCGuids))
}
