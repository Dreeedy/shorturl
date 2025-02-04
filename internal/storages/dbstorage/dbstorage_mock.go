// Code generated by MockGen. DO NOT EDIT.
// Source: F:\shorturl\internal\storages\dbstorage\dbstorage.go

// Package dbstorage is a generated GoMock package.
package dbstorage

import (
	reflect "reflect"

	common "github.com/Dreeedy/shorturl/internal/storages/common"
	gomock "github.com/golang/mock/gomock"
)

// MockDBStorage is a mock of DBStorage interface.
type MockDBStorage struct {
	ctrl     *gomock.Controller
	recorder *MockDBStorageMockRecorder
}

// MockDBStorageMockRecorder is the mock recorder for MockDBStorage.
type MockDBStorageMockRecorder struct {
	mock *MockDBStorage
}

// NewMockDBStorage creates a new mock instance.
func NewMockDBStorage(ctrl *gomock.Controller) *MockDBStorage {
	mock := &MockDBStorage{ctrl: ctrl}
	mock.recorder = &MockDBStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDBStorage) EXPECT() *MockDBStorageMockRecorder {
	return m.recorder
}

// DeleteURLsByUser mocks base method.
func (m *MockDBStorage) DeleteURLsByUser(hashes []string, userID int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteURLsByUser", hashes, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteURLsByUser indicates an expected call of DeleteURLsByUser.
func (mr *MockDBStorageMockRecorder) DeleteURLsByUser(hashes, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteURLsByUser", reflect.TypeOf((*MockDBStorage)(nil).DeleteURLsByUser), hashes, userID)
}

// GetURLWithDeletedFlag mocks base method.
func (m *MockDBStorage) GetURLWithDeletedFlag(shortURL string) (string, bool, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetURLWithDeletedFlag", shortURL)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(bool)
	return ret0, ret1, ret2
}

// GetURLWithDeletedFlag indicates an expected call of GetURLWithDeletedFlag.
func (mr *MockDBStorageMockRecorder) GetURLWithDeletedFlag(shortURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetURLWithDeletedFlag", reflect.TypeOf((*MockDBStorage)(nil).GetURLWithDeletedFlag), shortURL)
}

// GetURLsByUserID mocks base method.
func (m *MockDBStorage) GetURLsByUserID(userID int) (common.URLData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetURLsByUserID", userID)
	ret0, _ := ret[0].(common.URLData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetURLsByUserID indicates an expected call of GetURLsByUserID.
func (mr *MockDBStorageMockRecorder) GetURLsByUserID(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetURLsByUserID", reflect.TypeOf((*MockDBStorage)(nil).GetURLsByUserID), userID)
}
