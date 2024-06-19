// Code generated by mockery v2.43.1. DO NOT EDIT.

package move

import (
	context "context"

	db "github.com/go-risk-it/go-risk-it/internal/data/db"
	mock "github.com/stretchr/testify/mock"

	move "github.com/go-risk-it/go-risk-it/internal/logic/move/move"
)

// Performer is an autogenerated mock type for the Performer type
type Performer[T interface{}] struct {
	mock.Mock
}

type Performer_Expecter[T interface{}] struct {
	mock *mock.Mock
}

func (_m *Performer[T]) EXPECT() *Performer_Expecter[T] {
	return &Performer_Expecter[T]{mock: &_m.Mock}
}

// PerformQ provides a mock function with given fields: ctx, querier, _a2
func (_m *Performer[T]) PerformQ(ctx context.Context, querier db.Querier, _a2 move.Move[T]) error {
	ret := _m.Called(ctx, querier, _a2)

	if len(ret) == 0 {
		panic("no return value specified for PerformQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, db.Querier, move.Move[T]) error); ok {
		r0 = rf(ctx, querier, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Performer_PerformQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PerformQ'
type Performer_PerformQ_Call[T interface{}] struct {
	*mock.Call
}

// PerformQ is a helper method to define mock.On call
//   - ctx context.Context
//   - querier db.Querier
//   - _a2 move.Move[T]
func (_e *Performer_Expecter[T]) PerformQ(ctx interface{}, querier interface{}, _a2 interface{}) *Performer_PerformQ_Call[T] {
	return &Performer_PerformQ_Call[T]{Call: _e.mock.On("PerformQ", ctx, querier, _a2)}
}

func (_c *Performer_PerformQ_Call[T]) Run(run func(ctx context.Context, querier db.Querier, _a2 move.Move[T])) *Performer_PerformQ_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(db.Querier), args[2].(move.Move[T]))
	})
	return _c
}

func (_c *Performer_PerformQ_Call[T]) Return(_a0 error) *Performer_PerformQ_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Performer_PerformQ_Call[T]) RunAndReturn(run func(context.Context, db.Querier, move.Move[T]) error) *Performer_PerformQ_Call[T] {
	_c.Call.Return(run)
	return _c
}

// NewPerformer creates a new instance of Performer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPerformer[T interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *Performer[T] {
	mock := &Performer[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
