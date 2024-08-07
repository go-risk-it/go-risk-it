// Code generated by mockery v2.44.1. DO NOT EDIT.

package controller

import (
	messaging "github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	mock "github.com/stretchr/testify/mock"

	state "github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

// PhaseController is an autogenerated mock type for the PhaseController type
type PhaseController struct {
	mock.Mock
}

type PhaseController_Expecter struct {
	mock *mock.Mock
}

func (_m *PhaseController) EXPECT() *PhaseController_Expecter {
	return &PhaseController_Expecter{mock: &_m.Mock}
}

// GetAttackPhaseState provides a mock function with given fields: _a0, game
func (_m *PhaseController) GetAttackPhaseState(_a0 ctx.GameContext, game *state.Game) (messaging.GameState[messaging.EmptyState], error) {
	ret := _m.Called(_a0, game)

	if len(ret) == 0 {
		panic("no return value specified for GetAttackPhaseState")
	}

	var r0 messaging.GameState[messaging.EmptyState]
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, *state.Game) (messaging.GameState[messaging.EmptyState], error)); ok {
		return rf(_a0, game)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, *state.Game) messaging.GameState[messaging.EmptyState]); ok {
		r0 = rf(_a0, game)
	} else {
		r0 = ret.Get(0).(messaging.GameState[messaging.EmptyState])
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, *state.Game) error); ok {
		r1 = rf(_a0, game)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PhaseController_GetAttackPhaseState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAttackPhaseState'
type PhaseController_GetAttackPhaseState_Call struct {
	*mock.Call
}

// GetAttackPhaseState is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - game *state.Game
func (_e *PhaseController_Expecter) GetAttackPhaseState(_a0 interface{}, game interface{}) *PhaseController_GetAttackPhaseState_Call {
	return &PhaseController_GetAttackPhaseState_Call{Call: _e.mock.On("GetAttackPhaseState", _a0, game)}
}

func (_c *PhaseController_GetAttackPhaseState_Call) Run(run func(_a0 ctx.GameContext, game *state.Game)) *PhaseController_GetAttackPhaseState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(*state.Game))
	})
	return _c
}

func (_c *PhaseController_GetAttackPhaseState_Call) Return(_a0 messaging.GameState[messaging.EmptyState], _a1 error) *PhaseController_GetAttackPhaseState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PhaseController_GetAttackPhaseState_Call) RunAndReturn(run func(ctx.GameContext, *state.Game) (messaging.GameState[messaging.EmptyState], error)) *PhaseController_GetAttackPhaseState_Call {
	_c.Call.Return(run)
	return _c
}

// GetConquerPhaseState provides a mock function with given fields: _a0, game
func (_m *PhaseController) GetConquerPhaseState(_a0 ctx.GameContext, game *state.Game) (messaging.GameState[messaging.ConquerPhaseState], error) {
	ret := _m.Called(_a0, game)

	if len(ret) == 0 {
		panic("no return value specified for GetConquerPhaseState")
	}

	var r0 messaging.GameState[messaging.ConquerPhaseState]
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, *state.Game) (messaging.GameState[messaging.ConquerPhaseState], error)); ok {
		return rf(_a0, game)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, *state.Game) messaging.GameState[messaging.ConquerPhaseState]); ok {
		r0 = rf(_a0, game)
	} else {
		r0 = ret.Get(0).(messaging.GameState[messaging.ConquerPhaseState])
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, *state.Game) error); ok {
		r1 = rf(_a0, game)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PhaseController_GetConquerPhaseState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetConquerPhaseState'
type PhaseController_GetConquerPhaseState_Call struct {
	*mock.Call
}

// GetConquerPhaseState is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - game *state.Game
func (_e *PhaseController_Expecter) GetConquerPhaseState(_a0 interface{}, game interface{}) *PhaseController_GetConquerPhaseState_Call {
	return &PhaseController_GetConquerPhaseState_Call{Call: _e.mock.On("GetConquerPhaseState", _a0, game)}
}

func (_c *PhaseController_GetConquerPhaseState_Call) Run(run func(_a0 ctx.GameContext, game *state.Game)) *PhaseController_GetConquerPhaseState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(*state.Game))
	})
	return _c
}

func (_c *PhaseController_GetConquerPhaseState_Call) Return(_a0 messaging.GameState[messaging.ConquerPhaseState], _a1 error) *PhaseController_GetConquerPhaseState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PhaseController_GetConquerPhaseState_Call) RunAndReturn(run func(ctx.GameContext, *state.Game) (messaging.GameState[messaging.ConquerPhaseState], error)) *PhaseController_GetConquerPhaseState_Call {
	_c.Call.Return(run)
	return _c
}

// GetDeployPhaseState provides a mock function with given fields: _a0, game
func (_m *PhaseController) GetDeployPhaseState(_a0 ctx.GameContext, game *state.Game) (messaging.GameState[messaging.DeployPhaseState], error) {
	ret := _m.Called(_a0, game)

	if len(ret) == 0 {
		panic("no return value specified for GetDeployPhaseState")
	}

	var r0 messaging.GameState[messaging.DeployPhaseState]
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, *state.Game) (messaging.GameState[messaging.DeployPhaseState], error)); ok {
		return rf(_a0, game)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, *state.Game) messaging.GameState[messaging.DeployPhaseState]); ok {
		r0 = rf(_a0, game)
	} else {
		r0 = ret.Get(0).(messaging.GameState[messaging.DeployPhaseState])
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, *state.Game) error); ok {
		r1 = rf(_a0, game)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PhaseController_GetDeployPhaseState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDeployPhaseState'
type PhaseController_GetDeployPhaseState_Call struct {
	*mock.Call
}

// GetDeployPhaseState is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - game *state.Game
func (_e *PhaseController_Expecter) GetDeployPhaseState(_a0 interface{}, game interface{}) *PhaseController_GetDeployPhaseState_Call {
	return &PhaseController_GetDeployPhaseState_Call{Call: _e.mock.On("GetDeployPhaseState", _a0, game)}
}

func (_c *PhaseController_GetDeployPhaseState_Call) Run(run func(_a0 ctx.GameContext, game *state.Game)) *PhaseController_GetDeployPhaseState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(*state.Game))
	})
	return _c
}

func (_c *PhaseController_GetDeployPhaseState_Call) Return(_a0 messaging.GameState[messaging.DeployPhaseState], _a1 error) *PhaseController_GetDeployPhaseState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PhaseController_GetDeployPhaseState_Call) RunAndReturn(run func(ctx.GameContext, *state.Game) (messaging.GameState[messaging.DeployPhaseState], error)) *PhaseController_GetDeployPhaseState_Call {
	_c.Call.Return(run)
	return _c
}

// NewPhaseController creates a new instance of PhaseController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPhaseController(t interface {
	mock.TestingT
	Cleanup(func())
}) *PhaseController {
	mock := &PhaseController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
