// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/shubhamdubey02/subnet-evm/warp/aggregator (interfaces: SignatureGetter)

// Package aggregator is a generated GoMock package.
package aggregator

import (
	context "context"
	reflect "reflect"

	ids "github.com/MetalBlockchain/metalgo/ids"
	bls "github.com/MetalBlockchain/metalgo/utils/crypto/bls"
	warp "github.com/MetalBlockchain/metalgo/vms/platformvm/warp"
	gomock "go.uber.org/mock/gomock"
)

// MockSignatureGetter is a mock of SignatureGetter interface.
type MockSignatureGetter struct {
	ctrl     *gomock.Controller
	recorder *MockSignatureGetterMockRecorder
}

// MockSignatureGetterMockRecorder is the mock recorder for MockSignatureGetter.
type MockSignatureGetterMockRecorder struct {
	mock *MockSignatureGetter
}

// NewMockSignatureGetter creates a new mock instance.
func NewMockSignatureGetter(ctrl *gomock.Controller) *MockSignatureGetter {
	mock := &MockSignatureGetter{ctrl: ctrl}
	mock.recorder = &MockSignatureGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSignatureGetter) EXPECT() *MockSignatureGetterMockRecorder {
	return m.recorder
}

// GetSignature mocks base method.
func (m *MockSignatureGetter) GetSignature(arg0 context.Context, arg1 ids.NodeID, arg2 *warp.UnsignedMessage) (*bls.Signature, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSignature", arg0, arg1, arg2)
	ret0, _ := ret[0].(*bls.Signature)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSignature indicates an expected call of GetSignature.
func (mr *MockSignatureGetterMockRecorder) GetSignature(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSignature", reflect.TypeOf((*MockSignatureGetter)(nil).GetSignature), arg0, arg1, arg2)
}
