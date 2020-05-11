// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi (interfaces: File)

// Package mock_pluginapi is a generated GoMock package.
package mock_pluginapi

import (
	gomock "github.com/golang/mock/gomock"
	model "github.com/mattermost/mattermost-server/v5/model"
	io "io"
	reflect "reflect"
)

// MockFile is a mock of File interface
type MockFile struct {
	ctrl     *gomock.Controller
	recorder *MockFileMockRecorder
}

// MockFileMockRecorder is the mock recorder for MockFile
type MockFileMockRecorder struct {
	mock *MockFile
}

// NewMockFile creates a new mock instance
func NewMockFile(ctrl *gomock.Controller) *MockFile {
	mock := &MockFile{ctrl: ctrl}
	mock.recorder = &MockFileMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFile) EXPECT() *MockFileMockRecorder {
	return m.recorder
}

// Upload mocks base method
func (m *MockFile) Upload(arg0 io.Reader, arg1, arg2 string) (*model.FileInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upload", arg0, arg1, arg2)
	ret0, _ := ret[0].(*model.FileInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Upload indicates an expected call of Upload
func (mr *MockFileMockRecorder) Upload(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockFile)(nil).Upload), arg0, arg1, arg2)
}
