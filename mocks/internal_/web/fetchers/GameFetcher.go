// Code generated by mockery v2.40.1. DO NOT EDIT.

package fetchers

import (
	context "context"

	json "encoding/json"

	mock "github.com/stretchr/testify/mock"
)

// GameFetcher is an autogenerated mock type for the GameFetcher type
type GameFetcher struct {
	mock.Mock
}

type GameFetcher_Expecter struct {
	mock *mock.Mock
}

func (_m *GameFetcher) EXPECT() *GameFetcher_Expecter {
	return &GameFetcher_Expecter{mock: &_m.Mock}
}

// FetchState provides a mock function with given fields: ctx, gameID, stateChannel
func (_m *GameFetcher) FetchState(ctx context.Context, gameID int64, stateChannel chan json.RawMessage) {
	_m.Called(ctx, gameID, stateChannel)
}

// GameFetcher_FetchState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FetchState'
type GameFetcher_FetchState_Call struct {
	*mock.Call
}

// FetchState is a helper method to define mock.On call
//   - ctx context.Context
//   - gameID int64
//   - stateChannel chan json.RawMessage
func (_e *GameFetcher_Expecter) FetchState(ctx interface{}, gameID interface{}, stateChannel interface{}) *GameFetcher_FetchState_Call {
	return &GameFetcher_FetchState_Call{Call: _e.mock.On("FetchState", ctx, gameID, stateChannel)}
}

func (_c *GameFetcher_FetchState_Call) Run(run func(ctx context.Context, gameID int64, stateChannel chan json.RawMessage)) *GameFetcher_FetchState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64), args[2].(chan json.RawMessage))
	})
	return _c
}

func (_c *GameFetcher_FetchState_Call) Return() *GameFetcher_FetchState_Call {
	_c.Call.Return()
	return _c
}

func (_c *GameFetcher_FetchState_Call) RunAndReturn(run func(context.Context, int64, chan json.RawMessage)) *GameFetcher_FetchState_Call {
	_c.Call.Return(run)
	return _c
}

// NewGameFetcher creates a new instance of GameFetcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGameFetcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *GameFetcher {
	mock := &GameFetcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
