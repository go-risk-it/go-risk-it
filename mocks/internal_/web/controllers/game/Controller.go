// Code generated by mockery v2.40.1. DO NOT EDIT.

package game

import (
	context "context"

	message "github.com/tomfran/go-risk-it/internal/api/game/message"

	mock "github.com/stretchr/testify/mock"
)

// Controller is an autogenerated mock type for the Controller type
type Controller struct {
	mock.Mock
}

type Controller_Expecter struct {
	mock *mock.Mock
}

func (_m *Controller) EXPECT() *Controller_Expecter {
	return &Controller_Expecter{mock: &_m.Mock}
}

// GetGameState provides a mock function with given fields: ctx, gameID
func (_m *Controller) GetGameState(ctx context.Context, gameID int64) (message.GameState, error) {
	ret := _m.Called(ctx, gameID)

	if len(ret) == 0 {
		panic("no return value specified for GetGameState")
	}

	var r0 message.GameState
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (message.GameState, error)); ok {
		return rf(ctx, gameID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) message.GameState); ok {
		r0 = rf(ctx, gameID)
	} else {
		r0 = ret.Get(0).(message.GameState)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, gameID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Controller_GetGameState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGameState'
type Controller_GetGameState_Call struct {
	*mock.Call
}

// GetGameState is a helper method to define mock.On call
//   - ctx context.Context
//   - gameID int64
func (_e *Controller_Expecter) GetGameState(ctx interface{}, gameID interface{}) *Controller_GetGameState_Call {
	return &Controller_GetGameState_Call{Call: _e.mock.On("GetGameState", ctx, gameID)}
}

func (_c *Controller_GetGameState_Call) Run(run func(ctx context.Context, gameID int64)) *Controller_GetGameState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *Controller_GetGameState_Call) Return(_a0 message.GameState, _a1 error) *Controller_GetGameState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Controller_GetGameState_Call) RunAndReturn(run func(context.Context, int64) (message.GameState, error)) *Controller_GetGameState_Call {
	_c.Call.Return(run)
	return _c
}

// NewController creates a new instance of Controller. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewController(t interface {
	mock.TestingT
	Cleanup(func())
}) *Controller {
	mock := &Controller{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
