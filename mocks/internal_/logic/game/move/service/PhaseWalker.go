// Code generated by mockery v2.50.1. DO NOT EDIT.

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

// WalkQ provides a mock function with given fields: _a0, querier, voluntaryAdvancement
func (_m *PhaseWalker) WalkQ(_a0 ctx.GameContext, querier db.Querier, voluntaryAdvancement bool) (sqlc.PhaseType, error) {
	ret := _m.Called(_a0, querier, voluntaryAdvancement)

	if len(ret) == 0 {
		panic("no return value specified for WalkQ")
	}

	var r0 sqlc.PhaseType
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, bool) (sqlc.PhaseType, error)); ok {
		return rf(_a0, querier, voluntaryAdvancement)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, bool) sqlc.PhaseType); ok {
		r0 = rf(_a0, querier, voluntaryAdvancement)
	} else {
		r0 = ret.Get(0).(sqlc.PhaseType)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier, bool) error); ok {
		r1 = rf(_a0, querier, voluntaryAdvancement)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PhaseWalker_WalkQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WalkQ'
type PhaseWalker_WalkQ_Call struct {
	*mock.Call
}

// WalkQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
//   - voluntaryAdvancement bool
func (_e *PhaseWalker_Expecter) WalkQ(_a0 interface{}, querier interface{}, voluntaryAdvancement interface{}) *PhaseWalker_WalkQ_Call {
	return &PhaseWalker_WalkQ_Call{Call: _e.mock.On("WalkQ", _a0, querier, voluntaryAdvancement)}
}

func (_c *PhaseWalker_WalkQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier, voluntaryAdvancement bool)) *PhaseWalker_WalkQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].(bool))
	})
	return _c
}

func (_c *PhaseWalker_WalkQ_Call) Return(_a0 sqlc.PhaseType, _a1 error) *PhaseWalker_WalkQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PhaseWalker_WalkQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier, bool) (sqlc.PhaseType, error)) *PhaseWalker_WalkQ_Call {
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
