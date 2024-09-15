// Code generated by mockery v2.44.1. DO NOT EDIT.

package connection

import (
	json "encoding/json"

	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"

	mock "github.com/stretchr/testify/mock"

	websocket "github.com/lesismal/nbio/nbhttp/websocket"
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

// Broadcast provides a mock function with given fields: _a0, message
func (_m *Manager) Broadcast(_a0 ctx.GameContext, message json.RawMessage) {
	_m.Called(_a0, message)
}

// Manager_Broadcast_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Broadcast'
type Manager_Broadcast_Call struct {
	*mock.Call
}

// Broadcast is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - message json.RawMessage
func (_e *Manager_Expecter) Broadcast(_a0 interface{}, message interface{}) *Manager_Broadcast_Call {
	return &Manager_Broadcast_Call{Call: _e.mock.On("Broadcast", _a0, message)}
}

func (_c *Manager_Broadcast_Call) Run(run func(_a0 ctx.GameContext, message json.RawMessage)) *Manager_Broadcast_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(json.RawMessage))
	})
	return _c
}

func (_c *Manager_Broadcast_Call) Return() *Manager_Broadcast_Call {
	_c.Call.Return()
	return _c
}

func (_c *Manager_Broadcast_Call) RunAndReturn(run func(ctx.GameContext, json.RawMessage)) *Manager_Broadcast_Call {
	_c.Call.Return(run)
	return _c
}

// ConnectPlayer provides a mock function with given fields: _a0, _a1
func (_m *Manager) ConnectPlayer(_a0 ctx.GameContext, _a1 *websocket.Conn) {
	_m.Called(_a0, _a1)
}

// Manager_ConnectPlayer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ConnectPlayer'
type Manager_ConnectPlayer_Call struct {
	*mock.Call
}

// ConnectPlayer is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - _a1 *websocket.Conn
func (_e *Manager_Expecter) ConnectPlayer(_a0 interface{}, _a1 interface{}) *Manager_ConnectPlayer_Call {
	return &Manager_ConnectPlayer_Call{Call: _e.mock.On("ConnectPlayer", _a0, _a1)}
}

func (_c *Manager_ConnectPlayer_Call) Run(run func(_a0 ctx.GameContext, _a1 *websocket.Conn)) *Manager_ConnectPlayer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(*websocket.Conn))
	})
	return _c
}

func (_c *Manager_ConnectPlayer_Call) Return() *Manager_ConnectPlayer_Call {
	_c.Call.Return()
	return _c
}

func (_c *Manager_ConnectPlayer_Call) RunAndReturn(run func(ctx.GameContext, *websocket.Conn)) *Manager_ConnectPlayer_Call {
	_c.Call.Return(run)
	return _c
}

// WriteMessage provides a mock function with given fields: _a0, message
func (_m *Manager) WriteMessage(_a0 ctx.GameContext, message json.RawMessage) {
	_m.Called(_a0, message)
}

// Manager_WriteMessage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteMessage'
type Manager_WriteMessage_Call struct {
	*mock.Call
}

// WriteMessage is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - message json.RawMessage
func (_e *Manager_Expecter) WriteMessage(_a0 interface{}, message interface{}) *Manager_WriteMessage_Call {
	return &Manager_WriteMessage_Call{Call: _e.mock.On("WriteMessage", _a0, message)}
}

func (_c *Manager_WriteMessage_Call) Run(run func(_a0 ctx.GameContext, message json.RawMessage)) *Manager_WriteMessage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(json.RawMessage))
	})
	return _c
}

func (_c *Manager_WriteMessage_Call) Return() *Manager_WriteMessage_Call {
	_c.Call.Return()
	return _c
}

func (_c *Manager_WriteMessage_Call) RunAndReturn(run func(ctx.GameContext, json.RawMessage)) *Manager_WriteMessage_Call {
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
