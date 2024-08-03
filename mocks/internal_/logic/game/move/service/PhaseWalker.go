// Code generated by mockery v2.43.1. DO NOT EDIT.

package service

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/db"

	mock "github.com/stretchr/testify/mock"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

// PhaseWalker is an autogenerated mock type for the PhaseWalker type
type PhaseWalker struct {
	mock.Mock
}

type PhaseWalker_Expecter struct {
	mock *mock.Mock
}

func (_m *PhaseWalker) EXPECT() *PhaseWalker_Expecter {
	return &PhaseWalker_Expecter{mock: &_m.Mock}
}

// Walk provides a mock function with given fields: _a0, querier
func (_m *PhaseWalker) Walk(_a0 ctx.MoveContext, querier db.Querier) (sqlc.PhaseType, error) {
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

// PhaseWalker_Walk_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Walk'
type PhaseWalker_Walk_Call struct {
	*mock.Call
}

// Walk is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - querier db.Querier
func (_e *PhaseWalker_Expecter) Walk(_a0 interface{}, querier interface{}) *PhaseWalker_Walk_Call {
	return &PhaseWalker_Walk_Call{Call: _e.mock.On("Walk", _a0, querier)}
}

func (_c *PhaseWalker_Walk_Call) Run(run func(_a0 ctx.MoveContext, querier db.Querier)) *PhaseWalker_Walk_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *PhaseWalker_Walk_Call) Return(_a0 sqlc.PhaseType, _a1 error) *PhaseWalker_Walk_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PhaseWalker_Walk_Call) RunAndReturn(run func(ctx.MoveContext, db.Querier) (sqlc.PhaseType, error)) *PhaseWalker_Walk_Call {
	_c.Call.Return(run)
	return _c
}

// NewPhaseWalker creates a new instance of PhaseWalker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPhaseWalker(t interface {
	mock.TestingT
	Cleanup(func())
}) *PhaseWalker {
	mock := &PhaseWalker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
