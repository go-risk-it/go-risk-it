// Code generated by mockery v2.50.1. DO NOT EDIT.

package controller

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	mock "github.com/stretchr/testify/mock"

	request "github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
)

// MoveController is an autogenerated mock type for the MoveController type
type MoveController struct {
	mock.Mock
}

type MoveController_Expecter struct {
	mock *mock.Mock
}

func (_m *MoveController) EXPECT() *MoveController_Expecter {
	return &MoveController_Expecter{mock: &_m.Mock}
}

// PerformAttackMove provides a mock function with given fields: _a0, attackMove
func (_m *MoveController) PerformAttackMove(_a0 ctx.GameContext, attackMove request.AttackMove) error {
	ret := _m.Called(_a0, attackMove)

	if len(ret) == 0 {
		panic("no return value specified for PerformAttackMove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, request.AttackMove) error); ok {
		r0 = rf(_a0, attackMove)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MoveController_PerformAttackMove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PerformAttackMove'
type MoveController_PerformAttackMove_Call struct {
	*mock.Call
}

// PerformAttackMove is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - attackMove request.AttackMove
func (_e *MoveController_Expecter) PerformAttackMove(_a0 interface{}, attackMove interface{}) *MoveController_PerformAttackMove_Call {
	return &MoveController_PerformAttackMove_Call{Call: _e.mock.On("PerformAttackMove", _a0, attackMove)}
}

func (_c *MoveController_PerformAttackMove_Call) Run(run func(_a0 ctx.GameContext, attackMove request.AttackMove)) *MoveController_PerformAttackMove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(request.AttackMove))
	})
	return _c
}

func (_c *MoveController_PerformAttackMove_Call) Return(_a0 error) *MoveController_PerformAttackMove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MoveController_PerformAttackMove_Call) RunAndReturn(run func(ctx.GameContext, request.AttackMove) error) *MoveController_PerformAttackMove_Call {
	_c.Call.Return(run)
	return _c
}

// PerformCardsMove provides a mock function with given fields: _a0, cardsMove
func (_m *MoveController) PerformCardsMove(_a0 ctx.GameContext, cardsMove request.CardsMove) error {
	ret := _m.Called(_a0, cardsMove)

	if len(ret) == 0 {
		panic("no return value specified for PerformCardsMove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, request.CardsMove) error); ok {
		r0 = rf(_a0, cardsMove)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MoveController_PerformCardsMove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PerformCardsMove'
type MoveController_PerformCardsMove_Call struct {
	*mock.Call
}

// PerformCardsMove is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - cardsMove request.CardsMove
func (_e *MoveController_Expecter) PerformCardsMove(_a0 interface{}, cardsMove interface{}) *MoveController_PerformCardsMove_Call {
	return &MoveController_PerformCardsMove_Call{Call: _e.mock.On("PerformCardsMove", _a0, cardsMove)}
}

func (_c *MoveController_PerformCardsMove_Call) Run(run func(_a0 ctx.GameContext, cardsMove request.CardsMove)) *MoveController_PerformCardsMove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(request.CardsMove))
	})
	return _c
}

func (_c *MoveController_PerformCardsMove_Call) Return(_a0 error) *MoveController_PerformCardsMove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MoveController_PerformCardsMove_Call) RunAndReturn(run func(ctx.GameContext, request.CardsMove) error) *MoveController_PerformCardsMove_Call {
	_c.Call.Return(run)
	return _c
}

// PerformConquerMove provides a mock function with given fields: _a0, conquerMove
func (_m *MoveController) PerformConquerMove(_a0 ctx.GameContext, conquerMove request.ConquerMove) error {
	ret := _m.Called(_a0, conquerMove)

	if len(ret) == 0 {
		panic("no return value specified for PerformConquerMove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, request.ConquerMove) error); ok {
		r0 = rf(_a0, conquerMove)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MoveController_PerformConquerMove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PerformConquerMove'
type MoveController_PerformConquerMove_Call struct {
	*mock.Call
}

// PerformConquerMove is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - conquerMove request.ConquerMove
func (_e *MoveController_Expecter) PerformConquerMove(_a0 interface{}, conquerMove interface{}) *MoveController_PerformConquerMove_Call {
	return &MoveController_PerformConquerMove_Call{Call: _e.mock.On("PerformConquerMove", _a0, conquerMove)}
}

func (_c *MoveController_PerformConquerMove_Call) Run(run func(_a0 ctx.GameContext, conquerMove request.ConquerMove)) *MoveController_PerformConquerMove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(request.ConquerMove))
	})
	return _c
}

func (_c *MoveController_PerformConquerMove_Call) Return(_a0 error) *MoveController_PerformConquerMove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MoveController_PerformConquerMove_Call) RunAndReturn(run func(ctx.GameContext, request.ConquerMove) error) *MoveController_PerformConquerMove_Call {
	_c.Call.Return(run)
	return _c
}

// PerformDeployMove provides a mock function with given fields: _a0, deployMove
func (_m *MoveController) PerformDeployMove(_a0 ctx.GameContext, deployMove request.DeployMove) error {
	ret := _m.Called(_a0, deployMove)

	if len(ret) == 0 {
		panic("no return value specified for PerformDeployMove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, request.DeployMove) error); ok {
		r0 = rf(_a0, deployMove)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MoveController_PerformDeployMove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PerformDeployMove'
type MoveController_PerformDeployMove_Call struct {
	*mock.Call
}

// PerformDeployMove is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - deployMove request.DeployMove
func (_e *MoveController_Expecter) PerformDeployMove(_a0 interface{}, deployMove interface{}) *MoveController_PerformDeployMove_Call {
	return &MoveController_PerformDeployMove_Call{Call: _e.mock.On("PerformDeployMove", _a0, deployMove)}
}

func (_c *MoveController_PerformDeployMove_Call) Run(run func(_a0 ctx.GameContext, deployMove request.DeployMove)) *MoveController_PerformDeployMove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(request.DeployMove))
	})
	return _c
}

func (_c *MoveController_PerformDeployMove_Call) Return(_a0 error) *MoveController_PerformDeployMove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MoveController_PerformDeployMove_Call) RunAndReturn(run func(ctx.GameContext, request.DeployMove) error) *MoveController_PerformDeployMove_Call {
	_c.Call.Return(run)
	return _c
}

// PerformReinforceMove provides a mock function with given fields: _a0, reinforceMove
func (_m *MoveController) PerformReinforceMove(_a0 ctx.GameContext, reinforceMove request.ReinforceMove) error {
	ret := _m.Called(_a0, reinforceMove)

	if len(ret) == 0 {
		panic("no return value specified for PerformReinforceMove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, request.ReinforceMove) error); ok {
		r0 = rf(_a0, reinforceMove)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MoveController_PerformReinforceMove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PerformReinforceMove'
type MoveController_PerformReinforceMove_Call struct {
	*mock.Call
}

// PerformReinforceMove is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - reinforceMove request.ReinforceMove
func (_e *MoveController_Expecter) PerformReinforceMove(_a0 interface{}, reinforceMove interface{}) *MoveController_PerformReinforceMove_Call {
	return &MoveController_PerformReinforceMove_Call{Call: _e.mock.On("PerformReinforceMove", _a0, reinforceMove)}
}

func (_c *MoveController_PerformReinforceMove_Call) Run(run func(_a0 ctx.GameContext, reinforceMove request.ReinforceMove)) *MoveController_PerformReinforceMove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(request.ReinforceMove))
	})
	return _c
}

func (_c *MoveController_PerformReinforceMove_Call) Return(_a0 error) *MoveController_PerformReinforceMove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MoveController_PerformReinforceMove_Call) RunAndReturn(run func(ctx.GameContext, request.ReinforceMove) error) *MoveController_PerformReinforceMove_Call {
	_c.Call.Return(run)
	return _c
}

// NewMoveController creates a new instance of MoveController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMoveController(t interface {
	mock.TestingT
	Cleanup(func())
}) *MoveController {
	mock := &MoveController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
