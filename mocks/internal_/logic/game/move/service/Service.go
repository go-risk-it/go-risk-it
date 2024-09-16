// Code generated by mockery v2.44.1. DO NOT EDIT.

package service

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/db"

	mock "github.com/stretchr/testify/mock"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

// Service is an autogenerated mock type for the Service type
type Service[T interface{}, R interface{}] struct {
	mock.Mock
}

type Service_Expecter[T interface{}, R interface{}] struct {
	mock *mock.Mock
}

func (_m *Service[T, R]) EXPECT() *Service_Expecter[T, R] {
	return &Service_Expecter[T, R]{mock: &_m.Mock}
}

// AdvanceQ provides a mock function with given fields: _a0, querier, targetPhase, performResult
func (_m *Service[T, R]) AdvanceQ(_a0 ctx.GameContext, querier db.Querier, targetPhase sqlc.PhaseType, performResult R) error {
	ret := _m.Called(_a0, querier, targetPhase, performResult)

	if len(ret) == 0 {
		panic("no return value specified for AdvanceQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, sqlc.PhaseType, R) error); ok {
		r0 = rf(_a0, querier, targetPhase, performResult)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_AdvanceQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AdvanceQ'
type Service_AdvanceQ_Call[T interface{}, R interface{}] struct {
	*mock.Call
}

// AdvanceQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
//   - targetPhase sqlc.PhaseType
//   - performResult R
func (_e *Service_Expecter[T, R]) AdvanceQ(_a0 interface{}, querier interface{}, targetPhase interface{}, performResult interface{}) *Service_AdvanceQ_Call[T, R] {
	return &Service_AdvanceQ_Call[T, R]{Call: _e.mock.On("AdvanceQ", _a0, querier, targetPhase, performResult)}
}

func (_c *Service_AdvanceQ_Call[T, R]) Run(run func(_a0 ctx.GameContext, querier db.Querier, targetPhase sqlc.PhaseType, performResult R)) *Service_AdvanceQ_Call[T, R] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].(sqlc.PhaseType), args[3].(R))
	})
	return _c
}

func (_c *Service_AdvanceQ_Call[T, R]) Return(_a0 error) *Service_AdvanceQ_Call[T, R] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_AdvanceQ_Call[T, R]) RunAndReturn(run func(ctx.GameContext, db.Querier, sqlc.PhaseType, R) error) *Service_AdvanceQ_Call[T, R] {
	_c.Call.Return(run)
	return _c
}

// PerformQ provides a mock function with given fields: _a0, querier, move
func (_m *Service[T, R]) PerformQ(_a0 ctx.GameContext, querier db.Querier, move T) (R, error) {
	ret := _m.Called(_a0, querier, move)

	if len(ret) == 0 {
		panic("no return value specified for PerformQ")
	}

	var r0 R
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, T) (R, error)); ok {
		return rf(_a0, querier, move)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, T) R); ok {
		r0 = rf(_a0, querier, move)
	} else {
		r0 = ret.Get(0).(R)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier, T) error); ok {
		r1 = rf(_a0, querier, move)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_PerformQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PerformQ'
type Service_PerformQ_Call[T interface{}, R interface{}] struct {
	*mock.Call
}

// PerformQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
//   - move T
func (_e *Service_Expecter[T, R]) PerformQ(_a0 interface{}, querier interface{}, move interface{}) *Service_PerformQ_Call[T, R] {
	return &Service_PerformQ_Call[T, R]{Call: _e.mock.On("PerformQ", _a0, querier, move)}
}

func (_c *Service_PerformQ_Call[T, R]) Run(run func(_a0 ctx.GameContext, querier db.Querier, move T)) *Service_PerformQ_Call[T, R] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].(T))
	})
	return _c
}

func (_c *Service_PerformQ_Call[T, R]) Return(_a0 R, _a1 error) *Service_PerformQ_Call[T, R] {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_PerformQ_Call[T, R]) RunAndReturn(run func(ctx.GameContext, db.Querier, T) (R, error)) *Service_PerformQ_Call[T, R] {
	_c.Call.Return(run)
	return _c
}

// Walk provides a mock function with given fields: _a0, querier
func (_m *Service[T, R]) Walk(_a0 ctx.GameContext, querier db.Querier) (sqlc.PhaseType, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for Walk")
	}

	var r0 sqlc.PhaseType
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) (sqlc.PhaseType, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) sqlc.PhaseType); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Get(0).(sqlc.PhaseType)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_Walk_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Walk'
type Service_Walk_Call[T interface{}, R interface{}] struct {
	*mock.Call
}

// Walk is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter[T, R]) Walk(_a0 interface{}, querier interface{}) *Service_Walk_Call[T, R] {
	return &Service_Walk_Call[T, R]{Call: _e.mock.On("Walk", _a0, querier)}
}

func (_c *Service_Walk_Call[T, R]) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_Walk_Call[T, R] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_Walk_Call[T, R]) Return(_a0 sqlc.PhaseType, _a1 error) *Service_Walk_Call[T, R] {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_Walk_Call[T, R]) RunAndReturn(run func(ctx.GameContext, db.Querier) (sqlc.PhaseType, error)) *Service_Walk_Call[T, R] {
	_c.Call.Return(run)
	return _c
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService[T interface{}, R interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *Service[T, R] {
	mock := &Service[T, R]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
