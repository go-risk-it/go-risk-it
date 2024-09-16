// Code generated by mockery v2.44.1. DO NOT EDIT.

package orchestration

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	mock "github.com/stretchr/testify/mock"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

// Orchestrator is an autogenerated mock type for the Orchestrator type
type Orchestrator[T interface{}, R interface{}] struct {
	mock.Mock
}

type Orchestrator_Expecter[T interface{}, R interface{}] struct {
	mock *mock.Mock
}

func (_m *Orchestrator[T, R]) EXPECT() *Orchestrator_Expecter[T, R] {
	return &Orchestrator_Expecter[T, R]{mock: &_m.Mock}
}

// OrchestrateMove provides a mock function with given fields: _a0, phase, move
func (_m *Orchestrator[T, R]) OrchestrateMove(_a0 ctx.GameContext, phase sqlc.PhaseType, move T) error {
	ret := _m.Called(_a0, phase, move)

	if len(ret) == 0 {
		panic("no return value specified for OrchestrateMove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, sqlc.PhaseType, T) error); ok {
		r0 = rf(_a0, phase, move)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Orchestrator_OrchestrateMove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OrchestrateMove'
type Orchestrator_OrchestrateMove_Call[T interface{}, R interface{}] struct {
	*mock.Call
}

// OrchestrateMove is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - phase sqlc.PhaseType
//   - move T
func (_e *Orchestrator_Expecter[T, R]) OrchestrateMove(_a0 interface{}, phase interface{}, move interface{}) *Orchestrator_OrchestrateMove_Call[T, R] {
	return &Orchestrator_OrchestrateMove_Call[T, R]{Call: _e.mock.On("OrchestrateMove", _a0, phase, move)}
}

func (_c *Orchestrator_OrchestrateMove_Call[T, R]) Run(run func(_a0 ctx.GameContext, phase sqlc.PhaseType, move T)) *Orchestrator_OrchestrateMove_Call[T, R] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(sqlc.PhaseType), args[2].(T))
	})
	return _c
}

func (_c *Orchestrator_OrchestrateMove_Call[T, R]) Return(_a0 error) *Orchestrator_OrchestrateMove_Call[T, R] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Orchestrator_OrchestrateMove_Call[T, R]) RunAndReturn(run func(ctx.GameContext, sqlc.PhaseType, T) error) *Orchestrator_OrchestrateMove_Call[T, R] {
	_c.Call.Return(run)
	return _c
}

// NewOrchestrator creates a new instance of Orchestrator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOrchestrator[T interface{}, R interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *Orchestrator[T, R] {
	mock := &Orchestrator[T, R]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
