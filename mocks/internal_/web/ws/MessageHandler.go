// Code generated by mockery v2.40.1. DO NOT EDIT.

package ws

import (
	websocket "github.com/lesismal/nbio/nbhttp/websocket"
	mock "github.com/stretchr/testify/mock"
)

// MessageHandler is an autogenerated mock type for the MessageHandler type
type MessageHandler struct {
	mock.Mock
}

type MessageHandler_Expecter struct {
	mock *mock.Mock
}

func (_m *MessageHandler) EXPECT() *MessageHandler_Expecter {
	return &MessageHandler_Expecter{mock: &_m.Mock}
}

// OnMessage provides a mock function with given fields: connection, messageType, data
func (_m *MessageHandler) OnMessage(connection *websocket.Conn, messageType websocket.MessageType, data []byte) {
	_m.Called(connection, messageType, data)
}

// MessageHandler_OnMessage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OnMessage'
type MessageHandler_OnMessage_Call struct {
	*mock.Call
}

// OnMessage is a helper method to define mock.On call
//   - connection *websocket.Conn
//   - messageType websocket.MessageType
//   - data []byte
func (_e *MessageHandler_Expecter) OnMessage(connection interface{}, messageType interface{}, data interface{}) *MessageHandler_OnMessage_Call {
	return &MessageHandler_OnMessage_Call{Call: _e.mock.On("OnMessage", connection, messageType, data)}
}

func (_c *MessageHandler_OnMessage_Call) Run(run func(connection *websocket.Conn, messageType websocket.MessageType, data []byte)) *MessageHandler_OnMessage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*websocket.Conn), args[1].(websocket.MessageType), args[2].([]byte))
	})
	return _c
}

func (_c *MessageHandler_OnMessage_Call) Return() *MessageHandler_OnMessage_Call {
	_c.Call.Return()
	return _c
}

func (_c *MessageHandler_OnMessage_Call) RunAndReturn(run func(*websocket.Conn, websocket.MessageType, []byte)) *MessageHandler_OnMessage_Call {
	_c.Call.Return(run)
	return _c
}

// NewMessageHandler creates a new instance of MessageHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMessageHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MessageHandler {
	mock := &MessageHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
