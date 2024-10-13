// Code generated by mockery v2.46.2. DO NOT EDIT.

package graph

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"

	mock "github.com/stretchr/testify/mock"
)

// Graph is an autogenerated mock type for the Graph type
type Graph struct {
	mock.Mock
}

type Graph_Expecter struct {
	mock *mock.Mock
}

func (_m *Graph) EXPECT() *Graph_Expecter {
	return &Graph_Expecter{mock: &_m.Mock}
}

// AreNeighbours provides a mock function with given fields: source, target
func (_m *Graph) AreNeighbours(source string, target string) bool {
	ret := _m.Called(source, target)

	if len(ret) == 0 {
		panic("no return value specified for AreNeighbours")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(source, target)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Graph_AreNeighbours_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AreNeighbours'
type Graph_AreNeighbours_Call struct {
	*mock.Call
}

// AreNeighbours is a helper method to define mock.On call
//   - source string
//   - target string
func (_e *Graph_Expecter) AreNeighbours(source interface{}, target interface{}) *Graph_AreNeighbours_Call {
	return &Graph_AreNeighbours_Call{Call: _e.mock.On("AreNeighbours", source, target)}
}

func (_c *Graph_AreNeighbours_Call) Run(run func(source string, target string)) *Graph_AreNeighbours_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Graph_AreNeighbours_Call) Return(_a0 bool) *Graph_AreNeighbours_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Graph_AreNeighbours_Call) RunAndReturn(run func(string, string) bool) *Graph_AreNeighbours_Call {
	_c.Call.Return(run)
	return _c
}

// CanReach provides a mock function with given fields: context, source, target, usableRegions
func (_m *Graph) CanReach(context ctx.LogContext, source string, target string, usableRegions map[string]struct{}) bool {
	ret := _m.Called(context, source, target, usableRegions)

	if len(ret) == 0 {
		panic("no return value specified for CanReach")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(ctx.LogContext, string, string, map[string]struct{}) bool); ok {
		r0 = rf(context, source, target, usableRegions)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Graph_CanReach_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CanReach'
type Graph_CanReach_Call struct {
	*mock.Call
}

// CanReach is a helper method to define mock.On call
//   - context ctx.LogContext
//   - source string
//   - target string
//   - usableRegions map[string]struct{}
func (_e *Graph_Expecter) CanReach(context interface{}, source interface{}, target interface{}, usableRegions interface{}) *Graph_CanReach_Call {
	return &Graph_CanReach_Call{Call: _e.mock.On("CanReach", context, source, target, usableRegions)}
}

func (_c *Graph_CanReach_Call) Run(run func(context ctx.LogContext, source string, target string, usableRegions map[string]struct{})) *Graph_CanReach_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.LogContext), args[1].(string), args[2].(string), args[3].(map[string]struct{}))
	})
	return _c
}

func (_c *Graph_CanReach_Call) Return(_a0 bool) *Graph_CanReach_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Graph_CanReach_Call) RunAndReturn(run func(ctx.LogContext, string, string, map[string]struct{}) bool) *Graph_CanReach_Call {
	_c.Call.Return(run)
	return _c
}

// GetRegions provides a mock function with given fields:
func (_m *Graph) GetRegions() []string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetRegions")
	}

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Graph_GetRegions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRegions'
type Graph_GetRegions_Call struct {
	*mock.Call
}

// GetRegions is a helper method to define mock.On call
func (_e *Graph_Expecter) GetRegions() *Graph_GetRegions_Call {
	return &Graph_GetRegions_Call{Call: _e.mock.On("GetRegions")}
}

func (_c *Graph_GetRegions_Call) Run(run func()) *Graph_GetRegions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Graph_GetRegions_Call) Return(_a0 []string) *Graph_GetRegions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Graph_GetRegions_Call) RunAndReturn(run func() []string) *Graph_GetRegions_Call {
	_c.Call.Return(run)
	return _c
}

// NewGraph creates a new instance of Graph. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGraph(t interface {
	mock.TestingT
	Cleanup(func())
}) *Graph {
	mock := &Graph{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
