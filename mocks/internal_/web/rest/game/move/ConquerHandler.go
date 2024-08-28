// Code generated by mockery v2.44.1. DO NOT EDIT.

package move

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// ConquerHandler is an autogenerated mock type for the ConquerHandler type
type ConquerHandler struct {
	mock.Mock
}

type ConquerHandler_Expecter struct {
	mock *mock.Mock
}

func (_m *ConquerHandler) EXPECT() *ConquerHandler_Expecter {
	return &ConquerHandler_Expecter{mock: &_m.Mock}
}

// Pattern provides a mock function with given fields:
func (_m *ConquerHandler) Pattern() string {
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

// ConquerHandler_Pattern_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Pattern'
type ConquerHandler_Pattern_Call struct {
	*mock.Call
}

// Pattern is a helper method to define mock.On call
func (_e *ConquerHandler_Expecter) Pattern() *ConquerHandler_Pattern_Call {
	return &ConquerHandler_Pattern_Call{Call: _e.mock.On("Pattern")}
}

func (_c *ConquerHandler_Pattern_Call) Run(run func()) *ConquerHandler_Pattern_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ConquerHandler_Pattern_Call) Return(_a0 string) *ConquerHandler_Pattern_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ConquerHandler_Pattern_Call) RunAndReturn(run func() string) *ConquerHandler_Pattern_Call {
	_c.Call.Return(run)
	return _c
}

// RequiresAuth provides a mock function with given fields:
func (_m *ConquerHandler) RequiresAuth() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RequiresAuth")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ConquerHandler_RequiresAuth_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RequiresAuth'
type ConquerHandler_RequiresAuth_Call struct {
	*mock.Call
}

// RequiresAuth is a helper method to define mock.On call
func (_e *ConquerHandler_Expecter) RequiresAuth() *ConquerHandler_RequiresAuth_Call {
	return &ConquerHandler_RequiresAuth_Call{Call: _e.mock.On("RequiresAuth")}
}

func (_c *ConquerHandler_RequiresAuth_Call) Run(run func()) *ConquerHandler_RequiresAuth_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ConquerHandler_RequiresAuth_Call) Return(_a0 bool) *ConquerHandler_RequiresAuth_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ConquerHandler_RequiresAuth_Call) RunAndReturn(run func() bool) *ConquerHandler_RequiresAuth_Call {
	_c.Call.Return(run)
	return _c
}

// ServeHTTP provides a mock function with given fields: _a0, _a1
func (_m *ConquerHandler) ServeHTTP(_a0 http.ResponseWriter, _a1 *http.Request) {
	_m.Called(_a0, _a1)
}

// ConquerHandler_ServeHTTP_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ServeHTTP'
type ConquerHandler_ServeHTTP_Call struct {
	*mock.Call
}

// ServeHTTP is a helper method to define mock.On call
//   - _a0 http.ResponseWriter
//   - _a1 *http.Request
func (_e *ConquerHandler_Expecter) ServeHTTP(_a0 interface{}, _a1 interface{}) *ConquerHandler_ServeHTTP_Call {
	return &ConquerHandler_ServeHTTP_Call{Call: _e.mock.On("ServeHTTP", _a0, _a1)}
}

func (_c *ConquerHandler_ServeHTTP_Call) Run(run func(_a0 http.ResponseWriter, _a1 *http.Request)) *ConquerHandler_ServeHTTP_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.ResponseWriter), args[1].(*http.Request))
	})
	return _c
}

func (_c *ConquerHandler_ServeHTTP_Call) Return() *ConquerHandler_ServeHTTP_Call {
	_c.Call.Return()
	return _c
}

func (_c *ConquerHandler_ServeHTTP_Call) RunAndReturn(run func(http.ResponseWriter, *http.Request)) *ConquerHandler_ServeHTTP_Call {
	_c.Call.Return(run)
	return _c
}

// NewConquerHandler creates a new instance of ConquerHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewConquerHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *ConquerHandler {
	mock := &ConquerHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}