// Code generated by mockery v2.50.1. DO NOT EDIT.

package orchestration

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	conquer "github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"

	mock "github.com/stretchr/testify/mock"
)

// ConquerOrchestrator is an autogenerated mock type for the ConquerOrchestrator type
type ConquerOrchestrator struct {
	mock.Mock
}

type ConquerOrchestrator_Expecter struct {
	mock *mock.Mock
}

func (_m *ConquerOrchestrator) EXPECT() *ConquerOrchestrator_Expecter {
	return &ConquerOrchestrator_Expecter{mock: &_m.Mock}
}

// OrchestrateMove provides a mock function with given fields: _a0, move
func (_m *ConquerOrchestrator) OrchestrateMove(_a0 ctx.GameContext, move conquer.Move) error {
	ret := _m.Called(_a0, move)

	if len(ret) == 0 {
		panic("no return value specified for OrchestrateMove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, conquer.Move) error); ok {
		r0 = rf(_a0, move)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ConquerOrchestrator_OrchestrateMove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OrchestrateMove'
type ConquerOrchestrator_OrchestrateMove_Call struct {
	*mock.Call
}

// OrchestrateMove is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - move conquer.Move
func (_e *ConquerOrchestrator_Expecter) OrchestrateMove(_a0 interface{}, move interface{}) *ConquerOrchestrator_OrchestrateMove_Call {
	return &ConquerOrchestrator_OrchestrateMove_Call{Call: _e.mock.On("OrchestrateMove", _a0, move)}
}

func (_c *ConquerOrchestrator_OrchestrateMove_Call) Run(run func(_a0 ctx.GameContext, move conquer.Move)) *ConquerOrchestrator_OrchestrateMove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(conquer.Move))
	})
	return _c
}

func (_c *ConquerOrchestrator_OrchestrateMove_Call) Return(_a0 error) *ConquerOrchestrator_OrchestrateMove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ConquerOrchestrator_OrchestrateMove_Call) RunAndReturn(run func(ctx.GameContext, conquer.Move) error) *ConquerOrchestrator_OrchestrateMove_Call {
	_c.Call.Return(run)
	return _c
}

// NewConquerOrchestrator creates a new instance of ConquerOrchestrator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewConquerOrchestrator(t interface {
	mock.TestingT
	Cleanup(func())
}) *ConquerOrchestrator {
	mock := &ConquerOrchestrator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
