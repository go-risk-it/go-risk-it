// Code generated by mockery v2.42.1. DO NOT EDIT.

package rest

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// GameHandler is an autogenerated mock type for the GameHandler type
type GameHandler struct {
	mock.Mock
}

type GameHandler_Expecter struct {
	mock *mock.Mock
}

func (_m *GameHandler) EXPECT() *GameHandler_Expecter {
	return &GameHandler_Expecter{mock: &_m.Mock}
}

// Pattern provides a mock function with given fields:
func (_m *GameHandler) Pattern() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Pattern")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GameHandler_Pattern_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Pattern'
type GameHandler_Pattern_Call struct {
	*mock.Call
}

// Pattern is a helper method to define mock.On call
func (_e *GameHandler_Expecter) Pattern() *GameHandler_Pattern_Call {
	return &GameHandler_Pattern_Call{Call: _e.mock.On("Pattern")}
}

func (_c *GameHandler_Pattern_Call) Run(run func()) *GameHandler_Pattern_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *GameHandler_Pattern_Call) Return(_a0 string) *GameHandler_Pattern_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GameHandler_Pattern_Call) RunAndReturn(run func() string) *GameHandler_Pattern_Call {
	_c.Call.Return(run)
	return _c
}

// ServeHTTP provides a mock function with given fields: w, r
func (_m *GameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_m.Called(w, r)
}

// GameHandler_ServeHTTP_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ServeHTTP'
type GameHandler_ServeHTTP_Call struct {
	*mock.Call
}

// ServeHTTP is a helper method to define mock.On call
//   - w http.ResponseWriter
//   - r *http.Request
func (_e *GameHandler_Expecter) ServeHTTP(w interface{}, r interface{}) *GameHandler_ServeHTTP_Call {
	return &GameHandler_ServeHTTP_Call{Call: _e.mock.On("ServeHTTP", w, r)}
}

func (_c *GameHandler_ServeHTTP_Call) Run(run func(w http.ResponseWriter, r *http.Request)) *GameHandler_ServeHTTP_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.ResponseWriter), args[1].(*http.Request))
	})
	return _c
}

func (_c *GameHandler_ServeHTTP_Call) Return() *GameHandler_ServeHTTP_Call {
	_c.Call.Return()
	return _c
}

func (_c *GameHandler_ServeHTTP_Call) RunAndReturn(run func(http.ResponseWriter, *http.Request)) *GameHandler_ServeHTTP_Call {
	_c.Call.Return(run)
	return _c
}

// NewGameHandler creates a new instance of GameHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGameHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *GameHandler {
	mock := &GameHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
