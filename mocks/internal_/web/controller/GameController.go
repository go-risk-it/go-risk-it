// Code generated by mockery v2.42.1. DO NOT EDIT.

package controller

import (
	context "context"

	message "github.com/go-risk-it/go-risk-it/internal/api/game/message"

	mock "github.com/stretchr/testify/mock"

	request "github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
)

// GameController is an autogenerated mock type for the GameController type
type GameController struct {
	mock.Mock
}

type GameController_Expecter struct {
	mock *mock.Mock
}

func (_m *GameController) EXPECT() *GameController_Expecter {
	return &GameController_Expecter{mock: &_m.Mock}
}

// CreateGame provides a mock function with given fields: ctx, _a1
func (_m *GameController) CreateGame(ctx context.Context, _a1 request.CreateGame) (int64, error) {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for CreateGame")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, request.CreateGame) (int64, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, request.CreateGame) int64); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, request.CreateGame) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GameController_CreateGame_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateGame'
type GameController_CreateGame_Call struct {
	*mock.Call
}

// CreateGame is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 request.CreateGame
func (_e *GameController_Expecter) CreateGame(ctx interface{}, _a1 interface{}) *GameController_CreateGame_Call {
	return &GameController_CreateGame_Call{Call: _e.mock.On("CreateGame", ctx, _a1)}
}

func (_c *GameController_CreateGame_Call) Run(run func(ctx context.Context, _a1 request.CreateGame)) *GameController_CreateGame_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(request.CreateGame))
	})
	return _c
}

func (_c *GameController_CreateGame_Call) Return(_a0 int64, _a1 error) *GameController_CreateGame_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GameController_CreateGame_Call) RunAndReturn(run func(context.Context, request.CreateGame) (int64, error)) *GameController_CreateGame_Call {
	_c.Call.Return(run)
	return _c
}

// GetGameState provides a mock function with given fields: ctx, gameID
func (_m *GameController) GetGameState(ctx context.Context, gameID int64) (message.GameState, error) {
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

// GameController_GetGameState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGameState'
type GameController_GetGameState_Call struct {
	*mock.Call
}

// GetGameState is a helper method to define mock.On call
//   - ctx context.Context
//   - gameID int64
func (_e *GameController_Expecter) GetGameState(ctx interface{}, gameID interface{}) *GameController_GetGameState_Call {
	return &GameController_GetGameState_Call{Call: _e.mock.On("GetGameState", ctx, gameID)}
}

func (_c *GameController_GetGameState_Call) Run(run func(ctx context.Context, gameID int64)) *GameController_GetGameState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *GameController_GetGameState_Call) Return(_a0 message.GameState, _a1 error) *GameController_GetGameState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GameController_GetGameState_Call) RunAndReturn(run func(context.Context, int64) (message.GameState, error)) *GameController_GetGameState_Call {
	_c.Call.Return(run)
	return _c
}

// NewGameController creates a new instance of GameController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGameController(t interface {
	mock.TestingT
	Cleanup(func())
}) *GameController {
	mock := &GameController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}