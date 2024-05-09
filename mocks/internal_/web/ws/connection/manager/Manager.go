// Code generated by mockery v2.40.1. DO NOT EDIT.

package manager

import (
	websocket "github.com/lesismal/nbio/nbhttp/websocket"
	mock "github.com/stretchr/testify/mock"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

type Manager_Expecter struct {
	mock *mock.Mock
}

func (_m *Manager) EXPECT() *Manager_Expecter {
	return &Manager_Expecter{mock: &_m.Mock}
}

// ConnectPlayer provides a mock function with given fields: connection, gameID
func (_m *Manager) ConnectPlayer(connection *websocket.Conn, gameID int64) {
	_m.Called(connection, gameID)
}

// Manager_ConnectPlayer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ConnectPlayer'
type Manager_ConnectPlayer_Call struct {
	*mock.Call
}

// ConnectPlayer is a helper method to define mock.On call
//   - connection *websocket.Conn
//   - gameID int64
func (_e *Manager_Expecter) ConnectPlayer(connection interface{}, gameID interface{}) *Manager_ConnectPlayer_Call {
	return &Manager_ConnectPlayer_Call{Call: _e.mock.On("ConnectPlayer", connection, gameID)}
}

func (_c *Manager_ConnectPlayer_Call) Run(run func(connection *websocket.Conn, gameID int64)) *Manager_ConnectPlayer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*websocket.Conn), args[1].(int64))
	})
	return _c
}

func (_c *Manager_ConnectPlayer_Call) Return() *Manager_ConnectPlayer_Call {
	_c.Call.Return()
	return _c
}

func (_c *Manager_ConnectPlayer_Call) RunAndReturn(run func(*websocket.Conn, int64)) *Manager_ConnectPlayer_Call {
	_c.Call.Return(run)
	return _c
}

// DisconnectPlayer provides a mock function with given fields: connection, gameID
func (_m *Manager) DisconnectPlayer(connection *websocket.Conn, gameID int64) {
	_m.Called(connection, gameID)
}

// Manager_DisconnectPlayer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DisconnectPlayer'
type Manager_DisconnectPlayer_Call struct {
	*mock.Call
}

// DisconnectPlayer is a helper method to define mock.On call
//   - connection *websocket.Conn
//   - gameID int64
func (_e *Manager_Expecter) DisconnectPlayer(connection interface{}, gameID interface{}) *Manager_DisconnectPlayer_Call {
	return &Manager_DisconnectPlayer_Call{Call: _e.mock.On("DisconnectPlayer", connection, gameID)}
}

func (_c *Manager_DisconnectPlayer_Call) Run(run func(connection *websocket.Conn, gameID int64)) *Manager_DisconnectPlayer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*websocket.Conn), args[1].(int64))
	})
	return _c
}

func (_c *Manager_DisconnectPlayer_Call) Return() *Manager_DisconnectPlayer_Call {
	_c.Call.Return()
	return _c
}

func (_c *Manager_DisconnectPlayer_Call) RunAndReturn(run func(*websocket.Conn, int64)) *Manager_DisconnectPlayer_Call {
	_c.Call.Return(run)
	return _c
}

// NewManager creates a new instance of Manager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *Manager {
	mock := &Manager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}