// Code generated by mockery v2.50.1. DO NOT EDIT.

package controller

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	mock "github.com/stretchr/testify/mock"
)

// StartController is an autogenerated mock type for the StartController type
type StartController struct {
	mock.Mock
}

type StartController_Expecter struct {
	mock *mock.Mock
}

func (_m *StartController) EXPECT() *StartController_Expecter {
	return &StartController_Expecter{mock: &_m.Mock}
}

// StartGame provides a mock function with given fields: _a0
func (_m *StartController) StartGame(_a0 ctx.LobbyContext) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for StartGame")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.LobbyContext) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StartController_StartGame_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StartGame'
type StartController_StartGame_Call struct {
	*mock.Call
}

// StartGame is a helper method to define mock.On call
//   - _a0 ctx.LobbyContext
func (_e *StartController_Expecter) StartGame(_a0 interface{}) *StartController_StartGame_Call {
	return &StartController_StartGame_Call{Call: _e.mock.On("StartGame", _a0)}
}

func (_c *StartController_StartGame_Call) Run(run func(_a0 ctx.LobbyContext)) *StartController_StartGame_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.LobbyContext))
	})
	return _c
}

func (_c *StartController_StartGame_Call) Return(_a0 error) *StartController_StartGame_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *StartController_StartGame_Call) RunAndReturn(run func(ctx.LobbyContext) error) *StartController_StartGame_Call {
	_c.Call.Return(run)
	return _c
}

// NewStartController creates a new instance of StartController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStartController(t interface {
	mock.TestingT
	Cleanup(func())
}) *StartController {
	mock := &StartController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
