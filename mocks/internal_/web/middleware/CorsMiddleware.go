// Code generated by mockery v2.50.1. DO NOT EDIT.

package middleware

import (
	route "github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	mock "github.com/stretchr/testify/mock"
)

// CorsMiddleware is an autogenerated mock type for the CorsMiddleware type
type CorsMiddleware struct {
	mock.Mock
}

type CorsMiddleware_Expecter struct {
	mock *mock.Mock
}

func (_m *CorsMiddleware) EXPECT() *CorsMiddleware_Expecter {
	return &CorsMiddleware_Expecter{mock: &_m.Mock}
}

// Wrap provides a mock function with given fields: routeToWrap
func (_m *CorsMiddleware) Wrap(routeToWrap route.Route) route.Route {
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

// CorsMiddleware_Wrap_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Wrap'
type CorsMiddleware_Wrap_Call struct {
	*mock.Call
}

// Wrap is a helper method to define mock.On call
//   - routeToWrap route.Route
func (_e *CorsMiddleware_Expecter) Wrap(routeToWrap interface{}) *CorsMiddleware_Wrap_Call {
	return &CorsMiddleware_Wrap_Call{Call: _e.mock.On("Wrap", routeToWrap)}
}

func (_c *CorsMiddleware_Wrap_Call) Run(run func(routeToWrap route.Route)) *CorsMiddleware_Wrap_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(route.Route))
	})
	return _c
}

func (_c *CorsMiddleware_Wrap_Call) Return(_a0 route.Route) *CorsMiddleware_Wrap_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CorsMiddleware_Wrap_Call) RunAndReturn(run func(route.Route) route.Route) *CorsMiddleware_Wrap_Call {
	_c.Call.Return(run)
	return _c
}

// NewCorsMiddleware creates a new instance of CorsMiddleware. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCorsMiddleware(t interface {
	mock.TestingT
	Cleanup(func())
}) *CorsMiddleware {
	mock := &CorsMiddleware{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
