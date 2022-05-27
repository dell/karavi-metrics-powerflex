// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/dell/karavi-metrics-powerflex/internal/service (interfaces: VolumeFinder)

// Package mocks is a generated GoMock package.
package mocks

import (
	k8s "github.com/dell/karavi-metrics-powerflex/internal/k8s"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockVolumeFinder is a mock of VolumeFinder interface
type MockVolumeFinder struct {
	ctrl     *gomock.Controller
	recorder *MockVolumeFinderMockRecorder
}

// MockVolumeFinderMockRecorder is the mock recorder for MockVolumeFinder
type MockVolumeFinderMockRecorder struct {
	mock *MockVolumeFinder
}

// NewMockVolumeFinder creates a new mock instance
func NewMockVolumeFinder(ctrl *gomock.Controller) *MockVolumeFinder {
	mock := &MockVolumeFinder{ctrl: ctrl}
	mock.recorder = &MockVolumeFinderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockVolumeFinder) EXPECT() *MockVolumeFinderMockRecorder {
	return m.recorder
}

// GetPersistentVolumes mocks base method
func (m *MockVolumeFinder) GetPersistentVolumes() ([]k8s.VolumeInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPersistentVolumes")
	ret0, _ := ret[0].([]k8s.VolumeInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPersistentVolumes indicates an expected call of GetPersistentVolumes
func (mr *MockVolumeFinderMockRecorder) GetPersistentVolumes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPersistentVolumes", reflect.TypeOf((*MockVolumeFinder)(nil).GetPersistentVolumes))
}
