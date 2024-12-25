// Code generated by mockery v2.50.1. DO NOT EDIT.

package fetcher

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"

	json "encoding/json"

	mock "github.com/stretchr/testify/mock"
)

// PlayerFetcher is an autogenerated mock type for the PlayerFetcher type
type PlayerFetcher struct {
	mock.Mock
}

type PlayerFetcher_Expecter struct {
	mock *mock.Mock
}

func (_m *PlayerFetcher) EXPECT() *PlayerFetcher_Expecter {
	return &PlayerFetcher_Expecter{mock: &_m.Mock}
}

// FetchState provides a mock function with given fields: _a0, stateChannel
func (_m *PlayerFetcher) FetchState(_a0 ctx.GameContext, stateChannel chan json.RawMessage) {
	_m.Called(_a0, stateChannel)
}

// PlayerFetcher_FetchState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FetchState'
type PlayerFetcher_FetchState_Call struct {
	*mock.Call
}

// FetchState is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - stateChannel chan json.RawMessage
func (_e *PlayerFetcher_Expecter) FetchState(_a0 interface{}, stateChannel interface{}) *PlayerFetcher_FetchState_Call {
	return &PlayerFetcher_FetchState_Call{Call: _e.mock.On("FetchState", _a0, stateChannel)}
}

func (_c *PlayerFetcher_FetchState_Call) Run(run func(_a0 ctx.GameContext, stateChannel chan json.RawMessage)) *PlayerFetcher_FetchState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(chan json.RawMessage))
	})
	return _c
}

func (_c *PlayerFetcher_FetchState_Call) Return() *PlayerFetcher_FetchState_Call {
	_c.Call.Return()
	return _c
}

func (_c *PlayerFetcher_FetchState_Call) RunAndReturn(run func(ctx.GameContext, chan json.RawMessage)) *PlayerFetcher_FetchState_Call {
	_c.Run(run)
	return _c
}

// NewPlayerFetcher creates a new instance of PlayerFetcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPlayerFetcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *PlayerFetcher {
	mock := &PlayerFetcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
