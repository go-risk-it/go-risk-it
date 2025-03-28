// Code generated by mockery v2.50.1. DO NOT EDIT.

package advancement

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/game/db"

	mock "github.com/stretchr/testify/mock"
)

// CardsAdvancer is an autogenerated mock type for the CardsAdvancer type
type CardsAdvancer struct {
	mock.Mock
}

type CardsAdvancer_Expecter struct {
	mock *mock.Mock
}

func (_m *CardsAdvancer) EXPECT() *CardsAdvancer_Expecter {
	return &CardsAdvancer_Expecter{mock: &_m.Mock}
}

// Advance provides a mock function with given fields: _a0
func (_m *CardsAdvancer) Advance(_a0 ctx.GameContext) error {
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

// CardsAdvancer_Advance_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Advance'
type CardsAdvancer_Advance_Call struct {
	*mock.Call
}

// Advance is a helper method to define mock.On call
//   - _a0 ctx.GameContext
func (_e *CardsAdvancer_Expecter) Advance(_a0 interface{}) *CardsAdvancer_Advance_Call {
	return &CardsAdvancer_Advance_Call{Call: _e.mock.On("Advance", _a0)}
}

func (_c *CardsAdvancer_Advance_Call) Run(run func(_a0 ctx.GameContext)) *CardsAdvancer_Advance_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext))
	})
	return _c
}

func (_c *CardsAdvancer_Advance_Call) Return(_a0 error) *CardsAdvancer_Advance_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CardsAdvancer_Advance_Call) RunAndReturn(run func(ctx.GameContext) error) *CardsAdvancer_Advance_Call {
	_c.Call.Return(run)
	return _c
}

// AdvanceQ provides a mock function with given fields: _a0, querier
func (_m *CardsAdvancer) AdvanceQ(_a0 ctx.GameContext, querier db.Querier) error {
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

// CardsAdvancer_AdvanceQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AdvanceQ'
type CardsAdvancer_AdvanceQ_Call struct {
	*mock.Call
}

// AdvanceQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *CardsAdvancer_Expecter) AdvanceQ(_a0 interface{}, querier interface{}) *CardsAdvancer_AdvanceQ_Call {
	return &CardsAdvancer_AdvanceQ_Call{Call: _e.mock.On("AdvanceQ", _a0, querier)}
}

func (_c *CardsAdvancer_AdvanceQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *CardsAdvancer_AdvanceQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *CardsAdvancer_AdvanceQ_Call) Return(_a0 error) *CardsAdvancer_AdvanceQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CardsAdvancer_AdvanceQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) error) *CardsAdvancer_AdvanceQ_Call {
	_c.Call.Return(run)
	return _c
}

// NewCardsAdvancer creates a new instance of CardsAdvancer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCardsAdvancer(t interface {
	mock.TestingT
	Cleanup(func())
}) *CardsAdvancer {
	mock := &CardsAdvancer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
