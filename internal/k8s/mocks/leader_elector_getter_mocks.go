// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/dell/karavi-powerflex-metrics/internal/k8s (interfaces: LeaderElectorGetter)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockLeaderElectorGetter is a mock of LeaderElectorGetter interface
type MockLeaderElectorGetter struct {
	ctrl     *gomock.Controller
	recorder *MockLeaderElectorGetterMockRecorder
}

// MockLeaderElectorGetterMockRecorder is the mock recorder for MockLeaderElectorGetter
type MockLeaderElectorGetterMockRecorder struct {
	mock *MockLeaderElectorGetter
}

// NewMockLeaderElectorGetter creates a new mock instance
func NewMockLeaderElectorGetter(ctrl *gomock.Controller) *MockLeaderElectorGetter {
	mock := &MockLeaderElectorGetter{ctrl: ctrl}
	mock.recorder = &MockLeaderElectorGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLeaderElectorGetter) EXPECT() *MockLeaderElectorGetterMockRecorder {
	return m.recorder
}

// InitLeaderElection mocks base method
func (m *MockLeaderElectorGetter) InitLeaderElection(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitLeaderElection", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InitLeaderElection indicates an expected call of InitLeaderElection
func (mr *MockLeaderElectorGetterMockRecorder) InitLeaderElection(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitLeaderElection", reflect.TypeOf((*MockLeaderElectorGetter)(nil).InitLeaderElection), arg0, arg1)
}

// IsLeader mocks base method
func (m *MockLeaderElectorGetter) IsLeader() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsLeader")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsLeader indicates an expected call of IsLeader
func (mr *MockLeaderElectorGetterMockRecorder) IsLeader() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsLeader", reflect.TypeOf((*MockLeaderElectorGetter)(nil).IsLeader))
}
