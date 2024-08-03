// Code generated by mockery v2.43.1. DO NOT EDIT.

package service

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/db"

	mock "github.com/stretchr/testify/mock"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

// Service is an autogenerated mock type for the Service type
type Service[T interface{}] struct {
	mock.Mock
}

type Service_Expecter[T interface{}] struct {
	mock *mock.Mock
}

func (_m *Service[T]) EXPECT() *Service_Expecter[T] {
	return &Service_Expecter[T]{mock: &_m.Mock}
}

// AdvanceQ provides a mock function with given fields: _a0, querier, targetPhase, move
func (_m *Service[T]) AdvanceQ(_a0 ctx.MoveContext, querier db.Querier, targetPhase sqlc.PhaseType, move T) error {
	ret := _m.Called(_a0, querier, targetPhase, move)

	if len(ret) == 0 {
		panic("no return value specified for AdvanceQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, sqlc.PhaseType, T) error); ok {
		r0 = rf(_a0, querier, targetPhase, move)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_AdvanceQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AdvanceQ'
type Service_AdvanceQ_Call[T interface{}] struct {
	*mock.Call
}

// AdvanceQ is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - querier db.Querier
//   - targetPhase sqlc.PhaseType
//   - move T
func (_e *Service_Expecter[T]) AdvanceQ(_a0 interface{}, querier interface{}, targetPhase interface{}, move interface{}) *Service_AdvanceQ_Call[T] {
	return &Service_AdvanceQ_Call[T]{Call: _e.mock.On("AdvanceQ", _a0, querier, targetPhase, move)}
}

func (_c *Service_AdvanceQ_Call[T]) Run(run func(_a0 ctx.MoveContext, querier db.Querier, targetPhase sqlc.PhaseType, move T)) *Service_AdvanceQ_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier), args[2].(sqlc.PhaseType), args[3].(T))
	})
	return _c
}

func (_c *Service_AdvanceQ_Call[T]) Return(_a0 error) *Service_AdvanceQ_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_AdvanceQ_Call[T]) RunAndReturn(run func(ctx.MoveContext, db.Querier, sqlc.PhaseType, T) error) *Service_AdvanceQ_Call[T] {
	_c.Call.Return(run)
	return _c
}

// PerformQ provides a mock function with given fields: _a0, querier, move
func (_m *Service[T]) PerformQ(_a0 ctx.MoveContext, querier db.Querier, move T) error {
	ret := _m.Called(_a0, querier, move)

	if len(ret) == 0 {
		panic("no return value specified for PerformQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, T) error); ok {
		r0 = rf(_a0, querier, move)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_PerformQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PerformQ'
type Service_PerformQ_Call[T interface{}] struct {
	*mock.Call
}

// PerformQ is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - querier db.Querier
//   - move T
func (_e *Service_Expecter[T]) PerformQ(_a0 interface{}, querier interface{}, move interface{}) *Service_PerformQ_Call[T] {
	return &Service_PerformQ_Call[T]{Call: _e.mock.On("PerformQ", _a0, querier, move)}
}

func (_c *Service_PerformQ_Call[T]) Run(run func(_a0 ctx.MoveContext, querier db.Querier, move T)) *Service_PerformQ_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier), args[2].(T))
	})
	return _c
}

func (_c *Service_PerformQ_Call[T]) Return(_a0 error) *Service_PerformQ_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_PerformQ_Call[T]) RunAndReturn(run func(ctx.MoveContext, db.Querier, T) error) *Service_PerformQ_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Walk provides a mock function with given fields: _a0, querier
func (_m *Service[T]) Walk(_a0 ctx.MoveContext, querier db.Querier) (sqlc.PhaseType, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for Walk")
	}

	var r0 sqlc.PhaseType
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier) (sqlc.PhaseType, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier) sqlc.PhaseType); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Get(0).(sqlc.PhaseType)
	}

	if rf, ok := ret.Get(1).(func(ctx.MoveContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_Walk_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Walk'
type Service_Walk_Call[T interface{}] struct {
	*mock.Call
}

// Walk is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - querier db.Querier
func (_e *Service_Expecter[T]) Walk(_a0 interface{}, querier interface{}) *Service_Walk_Call[T] {
	return &Service_Walk_Call[T]{Call: _e.mock.On("Walk", _a0, querier)}
}

func (_c *Service_Walk_Call[T]) Run(run func(_a0 ctx.MoveContext, querier db.Querier)) *Service_Walk_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_Walk_Call[T]) Return(_a0 sqlc.PhaseType, _a1 error) *Service_Walk_Call[T] {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_Walk_Call[T]) RunAndReturn(run func(ctx.MoveContext, db.Querier) (sqlc.PhaseType, error)) *Service_Walk_Call[T] {
	_c.Call.Return(run)
	return _c
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService[T interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *Service[T] {
	mock := &Service[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
