// Code generated by mockery v2.43.1. DO NOT EDIT.

package controller

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
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

// CreateGame provides a mock function with given fields: _a0, _a1
func (_m *GameController) CreateGame(_a0 ctx.UserContext, _a1 request.CreateGame) (int64, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for CreateGame")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.UserContext, request.CreateGame) (int64, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(ctx.UserContext, request.CreateGame) int64); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(ctx.UserContext, request.CreateGame) error); ok {
		r1 = rf(_a0, _a1)
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
//   - _a0 ctx.UserContext
//   - _a1 request.CreateGame
func (_e *GameController_Expecter) CreateGame(_a0 interface{}, _a1 interface{}) *GameController_CreateGame_Call {
	return &GameController_CreateGame_Call{Call: _e.mock.On("CreateGame", _a0, _a1)}
}

func (_c *GameController_CreateGame_Call) Run(run func(_a0 ctx.UserContext, _a1 request.CreateGame)) *GameController_CreateGame_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.UserContext), args[1].(request.CreateGame))
	})
	return _c
}

func (_c *GameController_CreateGame_Call) Return(_a0 int64, _a1 error) *GameController_CreateGame_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GameController_CreateGame_Call) RunAndReturn(run func(ctx.UserContext, request.CreateGame) (int64, error)) *GameController_CreateGame_Call {
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
