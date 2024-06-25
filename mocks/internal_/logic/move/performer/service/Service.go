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

// MustAdvanceQ provides a mock function with given fields: _a0, querier, game
func (_m *Service[T]) MustAdvanceQ(_a0 ctx.MoveContext, querier db.Querier, game *sqlc.Game) bool {
	ret := _m.Called(_a0, querier, game)

	if len(ret) == 0 {
		panic("no return value specified for MustAdvanceQ")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, *sqlc.Game) bool); ok {
		r0 = rf(_a0, querier, game)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Service_MustAdvanceQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MustAdvanceQ'
type Service_MustAdvanceQ_Call[T interface{}] struct {
	*mock.Call
}

// MustAdvanceQ is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - querier db.Querier
//   - game *sqlc.Game
func (_e *Service_Expecter[T]) MustAdvanceQ(_a0 interface{}, querier interface{}, game interface{}) *Service_MustAdvanceQ_Call[T] {
	return &Service_MustAdvanceQ_Call[T]{Call: _e.mock.On("MustAdvanceQ", _a0, querier, game)}
}

func (_c *Service_MustAdvanceQ_Call[T]) Run(run func(_a0 ctx.MoveContext, querier db.Querier, game *sqlc.Game)) *Service_MustAdvanceQ_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier), args[2].(*sqlc.Game))
	})
	return _c
}

func (_c *Service_MustAdvanceQ_Call[T]) Return(_a0 bool) *Service_MustAdvanceQ_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_MustAdvanceQ_Call[T]) RunAndReturn(run func(ctx.MoveContext, db.Querier, *sqlc.Game) bool) *Service_MustAdvanceQ_Call[T] {
	_c.Call.Return(run)
	return _c
}

// PerformQ provides a mock function with given fields: _a0, querier, game, move
func (_m *Service[T]) PerformQ(_a0 ctx.MoveContext, querier db.Querier, game *sqlc.Game, move T) error {
	ret := _m.Called(_a0, querier, game, move)

	if len(ret) == 0 {
		panic("no return value specified for PerformQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, *sqlc.Game, T) error); ok {
		r0 = rf(_a0, querier, game, move)
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
//   - game *sqlc.Game
//   - move T
func (_e *Service_Expecter[T]) PerformQ(_a0 interface{}, querier interface{}, game interface{}, move interface{}) *Service_PerformQ_Call[T] {
	return &Service_PerformQ_Call[T]{Call: _e.mock.On("PerformQ", _a0, querier, game, move)}
}

func (_c *Service_PerformQ_Call[T]) Run(run func(_a0 ctx.MoveContext, querier db.Querier, game *sqlc.Game, move T)) *Service_PerformQ_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier), args[2].(*sqlc.Game), args[3].(T))
	})
	return _c
}

func (_c *Service_PerformQ_Call[T]) Return(_a0 error) *Service_PerformQ_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_PerformQ_Call[T]) RunAndReturn(run func(ctx.MoveContext, db.Querier, *sqlc.Game, T) error) *Service_PerformQ_Call[T] {
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
