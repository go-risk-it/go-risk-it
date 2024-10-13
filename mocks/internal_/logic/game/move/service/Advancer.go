// Code generated by mockery v2.46.2. DO NOT EDIT.

package service

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/db"

	mock "github.com/stretchr/testify/mock"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

// Advancer is an autogenerated mock type for the Advancer type
type Advancer[R any] struct {
	mock.Mock
}

type Advancer_Expecter[R any] struct {
	mock *mock.Mock
}

func (_m *Advancer[R]) EXPECT() *Advancer_Expecter[R] {
	return &Advancer_Expecter[R]{mock: &_m.Mock}
}

// AdvanceQ provides a mock function with given fields: _a0, querier, targetPhase, performResult
func (_m *Advancer[R]) AdvanceQ(_a0 ctx.GameContext, querier db.Querier, targetPhase sqlc.PhaseType, performResult R) error {
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

// Advancer_AdvanceQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AdvanceQ'
type Advancer_AdvanceQ_Call[R any] struct {
	*mock.Call
}

// AdvanceQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
//   - targetPhase sqlc.PhaseType
//   - performResult R
func (_e *Advancer_Expecter[R]) AdvanceQ(_a0 interface{}, querier interface{}, targetPhase interface{}, performResult interface{}) *Advancer_AdvanceQ_Call[R] {
	return &Advancer_AdvanceQ_Call[R]{Call: _e.mock.On("AdvanceQ", _a0, querier, targetPhase, performResult)}
}

func (_c *Advancer_AdvanceQ_Call[R]) Run(run func(_a0 ctx.GameContext, querier db.Querier, targetPhase sqlc.PhaseType, performResult R)) *Advancer_AdvanceQ_Call[R] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].(sqlc.PhaseType), args[3].(R))
	})
	return _c
}

func (_c *Advancer_AdvanceQ_Call[R]) Return(_a0 error) *Advancer_AdvanceQ_Call[R] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Advancer_AdvanceQ_Call[R]) RunAndReturn(run func(ctx.GameContext, db.Querier, sqlc.PhaseType, R) error) *Advancer_AdvanceQ_Call[R] {
	_c.Call.Return(run)
	return _c
}

// NewAdvancer creates a new instance of Advancer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAdvancer[R any](t interface {
	mock.TestingT
	Cleanup(func())
}) *Advancer[R] {
	mock := &Advancer[R]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
