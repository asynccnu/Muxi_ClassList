// Code generated by MockGen. DO NOT EDIT.
// Source: ./log.go

// Package mock_log is a generated GoMock package.
package mock_log

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockLogerPrinter is a mock of LogerPrinter interface.
type MockLogerPrinter struct {
	ctrl     *gomock.Controller
	recorder *MockLogerPrinterMockRecorder
}

// MockLogerPrinterMockRecorder is the mock recorder for MockLogerPrinter.
type MockLogerPrinterMockRecorder struct {
	mock *MockLogerPrinter
}

// NewMockLogerPrinter creates a new mock instance.
func NewMockLogerPrinter(ctrl *gomock.Controller) *MockLogerPrinter {
	mock := &MockLogerPrinter{ctrl: ctrl}
	mock.recorder = &MockLogerPrinterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogerPrinter) EXPECT() *MockLogerPrinterMockRecorder {
	return m.recorder
}

// FuncError mocks base method.
func (m *MockLogerPrinter) FuncError(f interface{}, err error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "FuncError", f, err)
}

// FuncError indicates an expected call of FuncError.
func (mr *MockLogerPrinterMockRecorder) FuncError(f, err interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FuncError", reflect.TypeOf((*MockLogerPrinter)(nil).FuncError), f, err)
}
