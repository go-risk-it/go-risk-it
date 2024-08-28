// Code generated by mockery v2.44.1. DO NOT EDIT.

package orchestration

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	mock "github.com/stretchr/testify/mock"

	service "github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

// Orchestator is an autogenerated mock type for the Orchestator type
type Orchestator[T interface{}, R interface{}] struct {
	mock.Mock
}

type Orchestator_Expecter[T interface{}, R interface{}] struct {
	mock *mock.Mock
}

func (_m *Orchestator[T, R]) EXPECT() *Orchestator_Expecter[T, R] {
	return &Orchestator_Expecter[T, R]{mock: &_m.Mock}
}

// OrchestrateMove provides a mock function with given fields: _a0, phase, _a2, move
func (_m *Orchestator[T, R]) OrchestrateMove(_a0 ctx.MoveContext, phase sqlc.PhaseType, _a2 service.Service[T, R], move T) error {
	ret := _m.Called(_a0, phase, _a2, move)

	if len(ret) == 0 {
		panic("no return value specified for OrchestrateMove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, sqlc.PhaseType, service.Service[T, R], T) error); ok {
		r0 = rf(_a0, phase, _a2, move)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Orchestator_OrchestrateMove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OrchestrateMove'
type Orchestator_OrchestrateMove_Call[T interface{}, R interface{}] struct {
	*mock.Call
}

// OrchestrateMove is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - phase sqlc.PhaseType
//   - _a2 service.Service[T,R]
//   - move T
func (_e *Orchestator_Expecter[T, R]) OrchestrateMove(_a0 interface{}, phase interface{}, _a2 interface{}, move interface{}) *Orchestator_OrchestrateMove_Call[T, R] {
	return &Orchestator_OrchestrateMove_Call[T, R]{Call: _e.mock.On("OrchestrateMove", _a0, phase, _a2, move)}
}

func (_c *Orchestator_OrchestrateMove_Call[T, R]) Run(run func(_a0 ctx.MoveContext, phase sqlc.PhaseType, _a2 service.Service[T, R], move T)) *Orchestator_OrchestrateMove_Call[T, R] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(sqlc.PhaseType), args[2].(service.Service[T, R]), args[3].(T))
	})
	return _c
}

func (_c *Orchestator_OrchestrateMove_Call[T, R]) Return(_a0 error) *Orchestator_OrchestrateMove_Call[T, R] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Orchestator_OrchestrateMove_Call[T, R]) RunAndReturn(run func(ctx.MoveContext, sqlc.PhaseType, service.Service[T, R], T) error) *Orchestator_OrchestrateMove_Call[T, R] {
	_c.Call.Return(run)
	return _c
}

// NewOrchestator creates a new instance of Orchestator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOrchestator[T interface{}, R interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *Orchestator[T, R] {
	mock := &Orchestator[T, R]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
