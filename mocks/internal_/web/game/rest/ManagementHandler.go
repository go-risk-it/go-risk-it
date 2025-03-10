// Code generated by mockery v2.50.1. DO NOT EDIT.

package rest

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// ManagementHandler is an autogenerated mock type for the ManagementHandler type
type ManagementHandler struct {
	mock.Mock
}

type ManagementHandler_Expecter struct {
	mock *mock.Mock
}

func (_m *ManagementHandler) EXPECT() *ManagementHandler_Expecter {
	return &ManagementHandler_Expecter{mock: &_m.Mock}
}

// Pattern provides a mock function with no fields
func (_m *ManagementHandler) Pattern() string {
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

// ManagementHandler_Pattern_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Pattern'
type ManagementHandler_Pattern_Call struct {
	*mock.Call
}

// Pattern is a helper method to define mock.On call
func (_e *ManagementHandler_Expecter) Pattern() *ManagementHandler_Pattern_Call {
	return &ManagementHandler_Pattern_Call{Call: _e.mock.On("Pattern")}
}

func (_c *ManagementHandler_Pattern_Call) Run(run func()) *ManagementHandler_Pattern_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ManagementHandler_Pattern_Call) Return(_a0 string) *ManagementHandler_Pattern_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ManagementHandler_Pattern_Call) RunAndReturn(run func() string) *ManagementHandler_Pattern_Call {
	_c.Call.Return(run)
	return _c
}

// RequiresAuth provides a mock function with no fields
func (_m *ManagementHandler) RequiresAuth() bool {
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

// ManagementHandler_RequiresAuth_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RequiresAuth'
type ManagementHandler_RequiresAuth_Call struct {
	*mock.Call
}

// RequiresAuth is a helper method to define mock.On call
func (_e *ManagementHandler_Expecter) RequiresAuth() *ManagementHandler_RequiresAuth_Call {
	return &ManagementHandler_RequiresAuth_Call{Call: _e.mock.On("RequiresAuth")}
}

func (_c *ManagementHandler_RequiresAuth_Call) Run(run func()) *ManagementHandler_RequiresAuth_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ManagementHandler_RequiresAuth_Call) Return(_a0 bool) *ManagementHandler_RequiresAuth_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ManagementHandler_RequiresAuth_Call) RunAndReturn(run func() bool) *ManagementHandler_RequiresAuth_Call {
	_c.Call.Return(run)
	return _c
}

// ServeHTTP provides a mock function with given fields: _a0, _a1
func (_m *ManagementHandler) ServeHTTP(_a0 http.ResponseWriter, _a1 *http.Request) {
	_m.Called(_a0, _a1)
}

// ManagementHandler_ServeHTTP_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ServeHTTP'
type ManagementHandler_ServeHTTP_Call struct {
	*mock.Call
}

// ServeHTTP is a helper method to define mock.On call
//   - _a0 http.ResponseWriter
//   - _a1 *http.Request
func (_e *ManagementHandler_Expecter) ServeHTTP(_a0 interface{}, _a1 interface{}) *ManagementHandler_ServeHTTP_Call {
	return &ManagementHandler_ServeHTTP_Call{Call: _e.mock.On("ServeHTTP", _a0, _a1)}
}

func (_c *ManagementHandler_ServeHTTP_Call) Run(run func(_a0 http.ResponseWriter, _a1 *http.Request)) *ManagementHandler_ServeHTTP_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.ResponseWriter), args[1].(*http.Request))
	})
	return _c
}

func (_c *ManagementHandler_ServeHTTP_Call) Return() *ManagementHandler_ServeHTTP_Call {
	_c.Call.Return()
	return _c
}

func (_c *ManagementHandler_ServeHTTP_Call) RunAndReturn(run func(http.ResponseWriter, *http.Request)) *ManagementHandler_ServeHTTP_Call {
	_c.Run(run)
	return _c
}

// NewManagementHandler creates a new instance of ManagementHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewManagementHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *ManagementHandler {
	mock := &ManagementHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
