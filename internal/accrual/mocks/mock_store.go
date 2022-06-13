// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/go-rfe/loyalty-system/internal/accrual (interfaces: Client)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	accrual "github.com/go-rfe/loyalty-system/internal/accrual"
	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// GetOrder mocks base method.
func (m *MockClient) GetOrder(arg0 context.Context, arg1 string) (*accrual.Accrual, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrder", arg0, arg1)
	ret0, _ := ret[0].(*accrual.Accrual)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrder indicates an expected call of GetOrder.
func (mr *MockClientMockRecorder) GetOrder(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrder", reflect.TypeOf((*MockClient)(nil).GetOrder), arg0, arg1)
}