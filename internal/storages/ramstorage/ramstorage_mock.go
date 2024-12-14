// Code generated by MockGen. DO NOT EDIT.
// Source: F:\shorturl\internal\storages\ramstorage\ramstorage.go

// Package ramstorage is a generated GoMock package.
package ramstorage

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// GetURL mocks base method.
func (m *MockStorage) GetURL(hash string) (string, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetURL", hash)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetURL indicates an expected call of GetURL.
func (mr *MockStorageMockRecorder) GetURL(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetURL", reflect.TypeOf((*MockStorage)(nil).GetURL), hash)
}

// SetURL mocks base method.
func (m *MockStorage) SetURL(hash, originalURL string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetURL", hash, originalURL)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetURL indicates an expected call of SetURL.
func (mr *MockStorageMockRecorder) SetURL(hash, originalURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetURL", reflect.TypeOf((*MockStorage)(nil).SetURL), hash, originalURL)
}