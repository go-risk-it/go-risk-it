// Code generated by mockery v2.44.1. DO NOT EDIT.

package orchestration

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	attack "github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"

	mock "github.com/stretchr/testify/mock"

	service "github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

// AttackOrchestrator is an autogenerated mock type for the AttackOrchestrator type
type AttackOrchestrator struct {
	mock.Mock
}

type AttackOrchestrator_Expecter struct {
	mock *mock.Mock
}

func (_m *AttackOrchestrator) EXPECT() *AttackOrchestrator_Expecter {
	return &AttackOrchestrator_Expecter{mock: &_m.Mock}
}

// OrchestrateMove provides a mock function with given fields: _a0, phase, _a2, move
func (_m *AttackOrchestrator) OrchestrateMove(_a0 ctx.MoveContext, phase sqlc.PhaseType, _a2 service.Service[attack.Move, *attack.MoveResult], move attack.Move) error {
	ret := _m.Called(_a0, phase, _a2, move)

	if len(ret) == 0 {
		panic("no return value specified for OrchestrateMove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, sqlc.PhaseType, service.Service[attack.Move, *attack.MoveResult], attack.Move) error); ok {
		r0 = rf(_a0, phase, _a2, move)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AttackOrchestrator_OrchestrateMove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OrchestrateMove'
type AttackOrchestrator_OrchestrateMove_Call struct {
	*mock.Call
}

// OrchestrateMove is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - phase sqlc.PhaseType
//   - _a2 service.Service[attack.Move,*attack.MoveResult]
//   - move attack.Move
func (_e *AttackOrchestrator_Expecter) OrchestrateMove(_a0 interface{}, phase interface{}, _a2 interface{}, move interface{}) *AttackOrchestrator_OrchestrateMove_Call {
	return &AttackOrchestrator_OrchestrateMove_Call{Call: _e.mock.On("OrchestrateMove", _a0, phase, _a2, move)}
}

func (_c *AttackOrchestrator_OrchestrateMove_Call) Run(run func(_a0 ctx.MoveContext, phase sqlc.PhaseType, _a2 service.Service[attack.Move, *attack.MoveResult], move attack.Move)) *AttackOrchestrator_OrchestrateMove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(sqlc.PhaseType), args[2].(service.Service[attack.Move, *attack.MoveResult]), args[3].(attack.Move))
	})
	return _c
}

func (_c *AttackOrchestrator_OrchestrateMove_Call) Return(_a0 error) *AttackOrchestrator_OrchestrateMove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AttackOrchestrator_OrchestrateMove_Call) RunAndReturn(run func(ctx.MoveContext, sqlc.PhaseType, service.Service[attack.Move, *attack.MoveResult], attack.Move) error) *AttackOrchestrator_OrchestrateMove_Call {
	_c.Call.Return(run)
	return _c
}

// NewAttackOrchestrator creates a new instance of AttackOrchestrator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAttackOrchestrator(t interface {
	mock.TestingT
	Cleanup(func())
}) *AttackOrchestrator {
	mock := &AttackOrchestrator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
