// Code generated by mockery v2.36.1. DO NOT EDIT.

package mocks

import (
	context "context"

	cosmosfaucet "github.com/ignite/cli/v29/ignite/pkg/cosmosfaucet"

	mock "github.com/stretchr/testify/mock"
)

// FaucetClient is an autogenerated mock type for the FaucetClient type
type FaucetClient struct {
	mock.Mock
}

type FaucetClient_Expecter struct {
	mock *mock.Mock
}

func (_m *FaucetClient) EXPECT() *FaucetClient_Expecter {
	return &FaucetClient_Expecter{mock: &_m.Mock}
}

// Transfer provides a mock function with given fields: _a0, _a1
func (_m *FaucetClient) Transfer(_a0 context.Context, _a1 cosmosfaucet.TransferRequest) (cosmosfaucet.TransferResponse, error) {
	ret := _m.Called(_a0, _a1)

	var r0 cosmosfaucet.TransferResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, cosmosfaucet.TransferRequest) (cosmosfaucet.TransferResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, cosmosfaucet.TransferRequest) cosmosfaucet.TransferResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(cosmosfaucet.TransferResponse)
	}

	if rf, ok := ret.Get(1).(func(context.Context, cosmosfaucet.TransferRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FaucetClient_Transfer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Transfer'
type FaucetClient_Transfer_Call struct {
	*mock.Call
}

// Transfer is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 cosmosfaucet.TransferRequest
func (_e *FaucetClient_Expecter) Transfer(_a0 interface{}, _a1 interface{}) *FaucetClient_Transfer_Call {
	return &FaucetClient_Transfer_Call{Call: _e.mock.On("Transfer", _a0, _a1)}
}

func (_c *FaucetClient_Transfer_Call) Run(run func(_a0 context.Context, _a1 cosmosfaucet.TransferRequest)) *FaucetClient_Transfer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(cosmosfaucet.TransferRequest))
	})
	return _c
}

func (_c *FaucetClient_Transfer_Call) Return(_a0 cosmosfaucet.TransferResponse, _a1 error) *FaucetClient_Transfer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *FaucetClient_Transfer_Call) RunAndReturn(run func(context.Context, cosmosfaucet.TransferRequest) (cosmosfaucet.TransferResponse, error)) *FaucetClient_Transfer_Call {
	_c.Call.Return(run)
	return _c
}

// NewFaucetClient creates a new instance of FaucetClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFaucetClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *FaucetClient {
	mock := &FaucetClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
