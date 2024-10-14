// Code generated by mockery v2.46.3. DO NOT EDIT.

package middleware

import (
	route "github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	mock "github.com/stretchr/testify/mock"
)

// LogMiddleware is an autogenerated mock type for the LogMiddleware type
type LogMiddleware struct {
	mock.Mock
}

type LogMiddleware_Expecter struct {
	mock *mock.Mock
}

func (_m *LogMiddleware) EXPECT() *LogMiddleware_Expecter {
	return &LogMiddleware_Expecter{mock: &_m.Mock}
}

// Wrap provides a mock function with given fields: routeToWrap
func (_m *LogMiddleware) Wrap(routeToWrap route.Route) route.Route {
	ret := _m.Called(routeToWrap)

	if len(ret) == 0 {
		panic("no return value specified for Wrap")
	}

	var r0 route.Route
	if rf, ok := ret.Get(0).(func(route.Route) route.Route); ok {
		r0 = rf(routeToWrap)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(route.Route)
		}
	}

	return r0
}

// LogMiddleware_Wrap_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Wrap'
type LogMiddleware_Wrap_Call struct {
	*mock.Call
}

// Wrap is a helper method to define mock.On call
//   - routeToWrap route.Route
func (_e *LogMiddleware_Expecter) Wrap(routeToWrap interface{}) *LogMiddleware_Wrap_Call {
	return &LogMiddleware_Wrap_Call{Call: _e.mock.On("Wrap", routeToWrap)}
}

func (_c *LogMiddleware_Wrap_Call) Run(run func(routeToWrap route.Route)) *LogMiddleware_Wrap_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(route.Route))
	})
	return _c
}

func (_c *LogMiddleware_Wrap_Call) Return(_a0 route.Route) *LogMiddleware_Wrap_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *LogMiddleware_Wrap_Call) RunAndReturn(run func(route.Route) route.Route) *LogMiddleware_Wrap_Call {
	_c.Call.Return(run)
	return _c
}

// NewLogMiddleware creates a new instance of LogMiddleware. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLogMiddleware(t interface {
	mock.TestingT
	Cleanup(func())
}) *LogMiddleware {
	mock := &LogMiddleware{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
