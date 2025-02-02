// Code generated by mockery v2.50.1. DO NOT EDIT.

package controller

import (
	messaging "github.com/go-risk-it/go-risk-it/internal/api/lobby/messaging"
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	mock "github.com/stretchr/testify/mock"
)

// StateController is an autogenerated mock type for the StateController type
type StateController struct {
	mock.Mock
}

type StateController_Expecter struct {
	mock *mock.Mock
}

func (_m *StateController) EXPECT() *StateController_Expecter {
	return &StateController_Expecter{mock: &_m.Mock}
}

// GetLobbyState provides a mock function with given fields: _a0
func (_m *StateController) GetLobbyState(_a0 ctx.LobbyContext) (messaging.LobbyState, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetLobbyState")
	}

	var r0 messaging.LobbyState
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.LobbyContext) (messaging.LobbyState, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(ctx.LobbyContext) messaging.LobbyState); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(messaging.LobbyState)
	}

	if rf, ok := ret.Get(1).(func(ctx.LobbyContext) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StateController_GetLobbyState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLobbyState'
type StateController_GetLobbyState_Call struct {
	*mock.Call
}

// GetLobbyState is a helper method to define mock.On call
//   - _a0 ctx.LobbyContext
func (_e *StateController_Expecter) GetLobbyState(_a0 interface{}) *StateController_GetLobbyState_Call {
	return &StateController_GetLobbyState_Call{Call: _e.mock.On("GetLobbyState", _a0)}
}

func (_c *StateController_GetLobbyState_Call) Run(run func(_a0 ctx.LobbyContext)) *StateController_GetLobbyState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.LobbyContext))
	})
	return _c
}

func (_c *StateController_GetLobbyState_Call) Return(_a0 messaging.LobbyState, _a1 error) *StateController_GetLobbyState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StateController_GetLobbyState_Call) RunAndReturn(run func(ctx.LobbyContext) (messaging.LobbyState, error)) *StateController_GetLobbyState_Call {
	_c.Call.Return(run)
	return _c
}

// NewStateController creates a new instance of StateController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStateController(t interface {
	mock.TestingT
	Cleanup(func())
}) *StateController {
	mock := &StateController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
