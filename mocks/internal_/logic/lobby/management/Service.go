// Code generated by mockery v2.50.1. DO NOT EDIT.

package management

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	management "github.com/go-risk-it/go-risk-it/internal/logic/lobby/management"
	mock "github.com/stretchr/testify/mock"
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

// GetUserLobbies provides a mock function with given fields: _a0
func (_m *Service) GetUserLobbies(_a0 ctx.UserContext) (*management.UserLobbies, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetUserLobbies")
	}

	var r0 *management.UserLobbies
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.UserContext) (*management.UserLobbies, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(ctx.UserContext) *management.UserLobbies); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*management.UserLobbies)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.UserContext) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetUserLobbies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUserLobbies'
type Service_GetUserLobbies_Call struct {
	*mock.Call
}

// GetUserLobbies is a helper method to define mock.On call
//   - _a0 ctx.UserContext
func (_e *Service_Expecter) GetUserLobbies(_a0 interface{}) *Service_GetUserLobbies_Call {
	return &Service_GetUserLobbies_Call{Call: _e.mock.On("GetUserLobbies", _a0)}
}

func (_c *Service_GetUserLobbies_Call) Run(run func(_a0 ctx.UserContext)) *Service_GetUserLobbies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.UserContext))
	})
	return _c
}

func (_c *Service_GetUserLobbies_Call) Return(_a0 *management.UserLobbies, _a1 error) *Service_GetUserLobbies_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetUserLobbies_Call) RunAndReturn(run func(ctx.UserContext) (*management.UserLobbies, error)) *Service_GetUserLobbies_Call {
	_c.Call.Return(run)
	return _c
}

// JoinLobby provides a mock function with given fields: _a0, name
func (_m *Service) JoinLobby(_a0 ctx.LobbyContext, name string) error {
	ret := _m.Called(_a0, name)

	if len(ret) == 0 {
		panic("no return value specified for JoinLobby")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.LobbyContext, string) error); ok {
		r0 = rf(_a0, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_JoinLobby_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'JoinLobby'
type Service_JoinLobby_Call struct {
	*mock.Call
}

// JoinLobby is a helper method to define mock.On call
//   - _a0 ctx.LobbyContext
//   - name string
func (_e *Service_Expecter) JoinLobby(_a0 interface{}, name interface{}) *Service_JoinLobby_Call {
	return &Service_JoinLobby_Call{Call: _e.mock.On("JoinLobby", _a0, name)}
}

func (_c *Service_JoinLobby_Call) Run(run func(_a0 ctx.LobbyContext, name string)) *Service_JoinLobby_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.LobbyContext), args[1].(string))
	})
	return _c
}

func (_c *Service_JoinLobby_Call) Return(_a0 error) *Service_JoinLobby_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_JoinLobby_Call) RunAndReturn(run func(ctx.LobbyContext, string) error) *Service_JoinLobby_Call {
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
