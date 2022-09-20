// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	client "github.com/cosmos/cosmos-sdk/client"

	mock "github.com/stretchr/testify/mock"

	tx "github.com/cosmos/cosmos-sdk/client/tx"
)

// Signer is an autogenerated mock type for the Signer type
type Signer struct {
	mock.Mock
}

type Signer_Expecter struct {
	mock *mock.Mock
}

func (_m *Signer) EXPECT() *Signer_Expecter {
	return &Signer_Expecter{mock: &_m.Mock}
}

// Sign provides a mock function with given fields: txf, name, txBuilder, overwriteSig
func (_m *Signer) Sign(txf tx.Factory, name string, txBuilder client.TxBuilder, overwriteSig bool) error {
	ret := _m.Called(txf, name, txBuilder, overwriteSig)

	var r0 error
	if rf, ok := ret.Get(0).(func(tx.Factory, string, client.TxBuilder, bool) error); ok {
		r0 = rf(txf, name, txBuilder, overwriteSig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Signer_Sign_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Sign'
type Signer_Sign_Call struct {
	*mock.Call
}

// Sign is a helper method to define mock.On call
//   - txf tx.Factory
//   - name string
//   - txBuilder client.TxBuilder
//   - overwriteSig bool
func (_e *Signer_Expecter) Sign(txf interface{}, name interface{}, txBuilder interface{}, overwriteSig interface{}) *Signer_Sign_Call {
	return &Signer_Sign_Call{Call: _e.mock.On("Sign", txf, name, txBuilder, overwriteSig)}
}

func (_c *Signer_Sign_Call) Run(run func(txf tx.Factory, name string, txBuilder client.TxBuilder, overwriteSig bool)) *Signer_Sign_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(tx.Factory), args[1].(string), args[2].(client.TxBuilder), args[3].(bool))
	})
	return _c
}

func (_c *Signer_Sign_Call) Return(_a0 error) *Signer_Sign_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewSigner interface {
	mock.TestingT
	Cleanup(func())
}

// NewSigner creates a new instance of Signer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSigner(t mockConstructorTestingTNewSigner) *Signer {
	mock := &Signer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
