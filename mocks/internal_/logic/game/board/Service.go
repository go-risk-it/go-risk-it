// Code generated by mockery v2.44.1. DO NOT EDIT.

package board

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
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

// AreNeighbours provides a mock function with given fields: context, source, target
func (_m *Service) AreNeighbours(context ctx.LogContext, source string, target string) (bool, error) {
	ret := _m.Called(context, source, target)

	if len(ret) == 0 {
		panic("no return value specified for AreNeighbours")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.LogContext, string, string) (bool, error)); ok {
		return rf(context, source, target)
	}
	if rf, ok := ret.Get(0).(func(ctx.LogContext, string, string) bool); ok {
		r0 = rf(context, source, target)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(ctx.LogContext, string, string) error); ok {
		r1 = rf(context, source, target)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_AreNeighbours_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AreNeighbours'
type Service_AreNeighbours_Call struct {
	*mock.Call
}

// AreNeighbours is a helper method to define mock.On call
//   - context ctx.LogContext
//   - source string
//   - target string
func (_e *Service_Expecter) AreNeighbours(context interface{}, source interface{}, target interface{}) *Service_AreNeighbours_Call {
	return &Service_AreNeighbours_Call{Call: _e.mock.On("AreNeighbours", context, source, target)}
}

func (_c *Service_AreNeighbours_Call) Run(run func(context ctx.LogContext, source string, target string)) *Service_AreNeighbours_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.LogContext), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *Service_AreNeighbours_Call) Return(_a0 bool, _a1 error) *Service_AreNeighbours_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_AreNeighbours_Call) RunAndReturn(run func(ctx.LogContext, string, string) (bool, error)) *Service_AreNeighbours_Call {
	_c.Call.Return(run)
	return _c
}

// CanPlayerReach provides a mock function with given fields: context, source, target
func (_m *Service) CanPlayerReach(context ctx.GameContext, source string, target string) (bool, error) {
	ret := _m.Called(context, source, target)

	if len(ret) == 0 {
		panic("no return value specified for CanPlayerReach")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, string, string) (bool, error)); ok {
		return rf(context, source, target)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, string, string) bool); ok {
		r0 = rf(context, source, target)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, string, string) error); ok {
		r1 = rf(context, source, target)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_CanPlayerReach_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CanPlayerReach'
type Service_CanPlayerReach_Call struct {
	*mock.Call
}

// CanPlayerReach is a helper method to define mock.On call
//   - context ctx.GameContext
//   - source string
//   - target string
func (_e *Service_Expecter) CanPlayerReach(context interface{}, source interface{}, target interface{}) *Service_CanPlayerReach_Call {
	return &Service_CanPlayerReach_Call{Call: _e.mock.On("CanPlayerReach", context, source, target)}
}

func (_c *Service_CanPlayerReach_Call) Run(run func(context ctx.GameContext, source string, target string)) *Service_CanPlayerReach_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *Service_CanPlayerReach_Call) Return(_a0 bool, _a1 error) *Service_CanPlayerReach_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_CanPlayerReach_Call) RunAndReturn(run func(ctx.GameContext, string, string) (bool, error)) *Service_CanPlayerReach_Call {
	_c.Call.Return(run)
	return _c
}

// GetBoardRegions provides a mock function with given fields: _a0
func (_m *Service) GetBoardRegions(_a0 ctx.LogContext) ([]string, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetBoardRegions")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.LogContext) ([]string, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(ctx.LogContext) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.LogContext) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetBoardRegions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetBoardRegions'
type Service_GetBoardRegions_Call struct {
	*mock.Call
}

// GetBoardRegions is a helper method to define mock.On call
//   - _a0 ctx.LogContext
func (_e *Service_Expecter) GetBoardRegions(_a0 interface{}) *Service_GetBoardRegions_Call {
	return &Service_GetBoardRegions_Call{Call: _e.mock.On("GetBoardRegions", _a0)}
}

func (_c *Service_GetBoardRegions_Call) Run(run func(_a0 ctx.LogContext)) *Service_GetBoardRegions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.LogContext))
	})
	return _c
}

func (_c *Service_GetBoardRegions_Call) Return(_a0 []string, _a1 error) *Service_GetBoardRegions_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetBoardRegions_Call) RunAndReturn(run func(ctx.LogContext) ([]string, error)) *Service_GetBoardRegions_Call {
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
