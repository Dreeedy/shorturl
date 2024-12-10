// Code generated by MockGen. DO NOT EDIT.
// Source: F:\shorturl\internal\config\config.go

// Package config is a generated GoMock package.
package config

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockConfig is a mock of Config interface.
type MockConfig struct {
	ctrl     *gomock.Controller
	recorder *MockConfigMockRecorder
}

// MockConfigMockRecorder is the mock recorder for MockConfig.
type MockConfigMockRecorder struct {
	mock *MockConfig
}

// NewMockConfig creates a new mock instance.
func NewMockConfig(ctrl *gomock.Controller) *MockConfig {
	mock := &MockConfig{ctrl: ctrl}
	mock.recorder = &MockConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConfig) EXPECT() *MockConfigMockRecorder {
	return m.recorder
}

// GetConfig mocks base method.
func (m *MockConfig) GetConfig() HTTPConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfig")
	ret0, _ := ret[0].(HTTPConfig)
	return ret0
}

// GetConfig indicates an expected call of GetConfig.
func (mr *MockConfigMockRecorder) GetConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfig", reflect.TypeOf((*MockConfig)(nil).GetConfig))
}