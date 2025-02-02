// Code generated by mockery v2.50.1. DO NOT EDIT.

package state

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/lobby/db"

	mock "github.com/stretchr/testify/mock"

	state "github.com/go-risk-it/go-risk-it/internal/logic/lobby/state"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

type Service_Expecter struct {
	mock *mock.Mock
}

func (_m *Service) EXPECT() *Service_Expecter {
	return &Service_Expecter{mock: &_m.Mock}
}

// GetLobbyState provides a mock function with given fields: _a0
func (_m *Service) GetLobbyState(_a0 ctx.LobbyContext) (*state.Lobby, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetLobbyState")
	}

	var r0 *state.Lobby
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.LobbyContext) (*state.Lobby, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(ctx.LobbyContext) *state.Lobby); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*state.Lobby)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.LobbyContext) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetLobbyState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLobbyState'
type Service_GetLobbyState_Call struct {
	*mock.Call
}

// GetLobbyState is a helper method to define mock.On call
//   - _a0 ctx.LobbyContext
func (_e *Service_Expecter) GetLobbyState(_a0 interface{}) *Service_GetLobbyState_Call {
	return &Service_GetLobbyState_Call{Call: _e.mock.On("GetLobbyState", _a0)}
}

func (_c *Service_GetLobbyState_Call) Run(run func(_a0 ctx.LobbyContext)) *Service_GetLobbyState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.LobbyContext))
	})
	return _c
}

func (_c *Service_GetLobbyState_Call) Return(_a0 *state.Lobby, _a1 error) *Service_GetLobbyState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetLobbyState_Call) RunAndReturn(run func(ctx.LobbyContext) (*state.Lobby, error)) *Service_GetLobbyState_Call {
	_c.Call.Return(run)
	return _c
}

// GetLobbyStateQ provides a mock function with given fields: _a0, querier
func (_m *Service) GetLobbyStateQ(_a0 ctx.LobbyContext, querier db.Querier) (*state.Lobby, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for GetLobbyStateQ")
	}

	var r0 *state.Lobby
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.LobbyContext, db.Querier) (*state.Lobby, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.LobbyContext, db.Querier) *state.Lobby); ok {
		r0 = rf(_a0, querier)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*state.Lobby)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.LobbyContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetLobbyStateQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLobbyStateQ'
type Service_GetLobbyStateQ_Call struct {
	*mock.Call
}

// GetLobbyStateQ is a helper method to define mock.On call
//   - _a0 ctx.LobbyContext
//   - querier db.Querier
func (_e *Service_Expecter) GetLobbyStateQ(_a0 interface{}, querier interface{}) *Service_GetLobbyStateQ_Call {
	return &Service_GetLobbyStateQ_Call{Call: _e.mock.On("GetLobbyStateQ", _a0, querier)}
}

func (_c *Service_GetLobbyStateQ_Call) Run(run func(_a0 ctx.LobbyContext, querier db.Querier)) *Service_GetLobbyStateQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.LobbyContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_GetLobbyStateQ_Call) Return(_a0 *state.Lobby, _a1 error) *Service_GetLobbyStateQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetLobbyStateQ_Call) RunAndReturn(run func(ctx.LobbyContext, db.Querier) (*state.Lobby, error)) *Service_GetLobbyStateQ_Call {
	_c.Call.Return(run)
	return _c
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService(t interface {
	mock.TestingT
	Cleanup(func())
}) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
