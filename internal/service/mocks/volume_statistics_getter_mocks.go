// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/dell/karavi-metrics-powerflex/internal/service (interfaces: VolumeStatisticsGetter)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	goscaleio "github.com/dell/goscaleio/types/v1"
	gomock "github.com/golang/mock/gomock"
)

// MockVolumeStatisticsGetter is a mock of VolumeStatisticsGetter interface.
type MockVolumeStatisticsGetter struct {
	ctrl     *gomock.Controller
	recorder *MockVolumeStatisticsGetterMockRecorder
}

// MockVolumeStatisticsGetterMockRecorder is the mock recorder for MockVolumeStatisticsGetter.
type MockVolumeStatisticsGetterMockRecorder struct {
	mock *MockVolumeStatisticsGetter
}

// NewMockVolumeStatisticsGetter creates a new mock instance.
func NewMockVolumeStatisticsGetter(ctrl *gomock.Controller) *MockVolumeStatisticsGetter {
	mock := &MockVolumeStatisticsGetter{ctrl: ctrl}
	mock.recorder = &MockVolumeStatisticsGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVolumeStatisticsGetter) EXPECT() *MockVolumeStatisticsGetterMockRecorder {
	return m.recorder
}

// GetVolumeStatistics mocks base method.
func (m *MockVolumeStatisticsGetter) GetVolumeStatistics() (*goscaleio.VolumeStatistics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVolumeStatistics")
	ret0, _ := ret[0].(*goscaleio.VolumeStatistics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVolumeStatistics indicates an expected call of GetVolumeStatistics.
func (mr *MockVolumeStatisticsGetterMockRecorder) GetVolumeStatistics() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVolumeStatistics", reflect.TypeOf((*MockVolumeStatisticsGetter)(nil).GetVolumeStatistics))
}
