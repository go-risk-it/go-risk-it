// Code generated by mockery v2.50.1. DO NOT EDIT.

package advancement

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/game/db"

	mock "github.com/stretchr/testify/mock"
)

// Service is an autogenerated mock type for the Service type
type Service[T any, R any] struct {
	mock.Mock
}

type Service_Expecter[T any, R any] struct {
	mock *mock.Mock
}

func (_m *Service[T, R]) EXPECT() *Service_Expecter[T, R] {
	return &Service_Expecter[T, R]{mock: &_m.Mock}
}

// Advance provides a mock function with given fields: _a0
func (_m *Service[T, R]) Advance(_a0 ctx.GameContext) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Advance")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_Advance_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Advance'
type Service_Advance_Call[T any, R any] struct {
	*mock.Call
}

// Advance is a helper method to define mock.On call
//   - _a0 ctx.GameContext
func (_e *Service_Expecter[T, R]) Advance(_a0 interface{}) *Service_Advance_Call[T, R] {
	return &Service_Advance_Call[T, R]{Call: _e.mock.On("Advance", _a0)}
}

func (_c *Service_Advance_Call[T, R]) Run(run func(_a0 ctx.GameContext)) *Service_Advance_Call[T, R] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext))
	})
	return _c
}

func (_c *Service_Advance_Call[T, R]) Return(_a0 error) *Service_Advance_Call[T, R] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_Advance_Call[T, R]) RunAndReturn(run func(ctx.GameContext) error) *Service_Advance_Call[T, R] {
	_c.Call.Return(run)
	return _c
}

// AdvanceQ provides a mock function with given fields: _a0, querier
func (_m *Service[T, R]) AdvanceQ(_a0 ctx.GameContext, querier db.Querier) error {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for AdvanceQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) error); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_AdvanceQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AdvanceQ'
type Service_AdvanceQ_Call[T any, R any] struct {
	*mock.Call
}

// AdvanceQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter[T, R]) AdvanceQ(_a0 interface{}, querier interface{}) *Service_AdvanceQ_Call[T, R] {
	return &Service_AdvanceQ_Call[T, R]{Call: _e.mock.On("AdvanceQ", _a0, querier)}
}

func (_c *Service_AdvanceQ_Call[T, R]) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_AdvanceQ_Call[T, R] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_AdvanceQ_Call[T, R]) Return(_a0 error) *Service_AdvanceQ_Call[T, R] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_AdvanceQ_Call[T, R]) RunAndReturn(run func(ctx.GameContext, db.Querier) error) *Service_AdvanceQ_Call[T, R] {
	_c.Call.Return(run)
	return _c
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService[T any, R any](t interface {
	mock.TestingT
	Cleanup(func())
}) *Service[T, R] {
	mock := &Service[T, R]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
