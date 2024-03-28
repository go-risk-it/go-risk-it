// Code generated by mockery v2.42.1. DO NOT EDIT.

package controller

import (
	context "context"

	message "github.com/tomfran/go-risk-it/internal/api/game/message"

	mock "github.com/stretchr/testify/mock"
)

// PlayerController is an autogenerated mock type for the PlayerController type
type PlayerController struct {
	mock.Mock
}

type PlayerController_Expecter struct {
	mock *mock.Mock
}

func (_m *PlayerController) EXPECT() *PlayerController_Expecter {
	return &PlayerController_Expecter{mock: &_m.Mock}
}

// GetPlayerState provides a mock function with given fields: ctx, gameID
func (_m *PlayerController) GetPlayerState(ctx context.Context, gameID int64) (message.PlayersState, error) {
	ret := _m.Called(ctx, gameID)

	if len(ret) == 0 {
		panic("no return value specified for GetPlayerState")
	}

	var r0 message.PlayersState
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (message.PlayersState, error)); ok {
		return rf(ctx, gameID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) message.PlayersState); ok {
		r0 = rf(ctx, gameID)
	} else {
		r0 = ret.Get(0).(message.PlayersState)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, gameID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PlayerController_GetPlayerState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPlayerState'
type PlayerController_GetPlayerState_Call struct {
	*mock.Call
}

// GetPlayerState is a helper method to define mock.On call
//   - ctx context.Context
//   - gameID int64
func (_e *PlayerController_Expecter) GetPlayerState(ctx interface{}, gameID interface{}) *PlayerController_GetPlayerState_Call {
	return &PlayerController_GetPlayerState_Call{Call: _e.mock.On("GetPlayerState", ctx, gameID)}
}

func (_c *PlayerController_GetPlayerState_Call) Run(run func(ctx context.Context, gameID int64)) *PlayerController_GetPlayerState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *PlayerController_GetPlayerState_Call) Return(_a0 message.PlayersState, _a1 error) *PlayerController_GetPlayerState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PlayerController_GetPlayerState_Call) RunAndReturn(run func(context.Context, int64) (message.PlayersState, error)) *PlayerController_GetPlayerState_Call {
	_c.Call.Return(run)
	return _c
}

// NewPlayerController creates a new instance of PlayerController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPlayerController(t interface {
	mock.TestingT
	Cleanup(func())
}) *PlayerController {
	mock := &PlayerController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
