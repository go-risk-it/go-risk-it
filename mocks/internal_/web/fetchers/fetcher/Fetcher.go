// Code generated by mockery v2.50.1. DO NOT EDIT.

package fetcher

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"

	json "encoding/json"

	mock "github.com/stretchr/testify/mock"
)

// Fetcher is an autogenerated mock type for the Fetcher type
type Fetcher struct {
	mock.Mock
}

type Fetcher_Expecter struct {
	mock *mock.Mock
}

func (_m *Fetcher) EXPECT() *Fetcher_Expecter {
	return &Fetcher_Expecter{mock: &_m.Mock}
}

// FetchState provides a mock function with given fields: _a0, stateChannel
func (_m *Fetcher) FetchState(_a0 ctx.GameContext, stateChannel chan json.RawMessage) {
	_m.Called(_a0, stateChannel)
}

// Fetcher_FetchState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FetchState'
type Fetcher_FetchState_Call struct {
	*mock.Call
}

// FetchState is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - stateChannel chan json.RawMessage
func (_e *Fetcher_Expecter) FetchState(_a0 interface{}, stateChannel interface{}) *Fetcher_FetchState_Call {
	return &Fetcher_FetchState_Call{Call: _e.mock.On("FetchState", _a0, stateChannel)}
}

func (_c *Fetcher_FetchState_Call) Run(run func(_a0 ctx.GameContext, stateChannel chan json.RawMessage)) *Fetcher_FetchState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(chan json.RawMessage))
	})
	return _c
}

func (_c *Fetcher_FetchState_Call) Return() *Fetcher_FetchState_Call {
	_c.Call.Return()
	return _c
}

func (_c *Fetcher_FetchState_Call) RunAndReturn(run func(ctx.GameContext, chan json.RawMessage)) *Fetcher_FetchState_Call {
	_c.Run(run)
	return _c
}

// NewFetcher creates a new instance of Fetcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFetcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *Fetcher {
	mock := &Fetcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
