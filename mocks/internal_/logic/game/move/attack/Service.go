// Code generated by mockery v2.50.1. DO NOT EDIT.

package attack

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	attack "github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"

	db "github.com/go-risk-it/go-risk-it/internal/data/game/db"

	mock "github.com/stretchr/testify/mock"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

type Service_Expecter struct {
	mock *mock.Mock
}

func (_m *Service) EXPECT() *Service_Expecter {
	return &Service_Expecter{mock: &_m.Mock}
}

// AdvanceQ provides a mock function with given fields: _a0, querier, targetPhase, performResult
func (_m *Service) AdvanceQ(_a0 ctx.GameContext, querier db.Querier, targetPhase sqlc.PhaseType, performResult *attack.MoveResult) error {
	ret := _m.Called(_a0, querier, targetPhase, performResult)

	if len(ret) == 0 {
		panic("no return value specified for AdvanceQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, sqlc.PhaseType, *attack.MoveResult) error); ok {
		r0 = rf(_a0, querier, targetPhase, performResult)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_AdvanceQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AdvanceQ'
type Service_AdvanceQ_Call struct {
	*mock.Call
}

// AdvanceQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
//   - targetPhase sqlc.PhaseType
//   - performResult *attack.MoveResult
func (_e *Service_Expecter) AdvanceQ(_a0 interface{}, querier interface{}, targetPhase interface{}, performResult interface{}) *Service_AdvanceQ_Call {
	return &Service_AdvanceQ_Call{Call: _e.mock.On("AdvanceQ", _a0, querier, targetPhase, performResult)}
}

func (_c *Service_AdvanceQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier, targetPhase sqlc.PhaseType, performResult *attack.MoveResult)) *Service_AdvanceQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].(sqlc.PhaseType), args[3].(*attack.MoveResult))
	})
	return _c
}

func (_c *Service_AdvanceQ_Call) Return(_a0 error) *Service_AdvanceQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_AdvanceQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier, sqlc.PhaseType, *attack.MoveResult) error) *Service_AdvanceQ_Call {
	_c.Call.Return(run)
	return _c
}

// CanContinueAttackingQ provides a mock function with given fields: _a0, querier
func (_m *Service) CanContinueAttackingQ(_a0 ctx.GameContext, querier db.Querier) (bool, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for CanContinueAttackingQ")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) (bool, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) bool); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_CanContinueAttackingQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CanContinueAttackingQ'
type Service_CanContinueAttackingQ_Call struct {
	*mock.Call
}

// CanContinueAttackingQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter) CanContinueAttackingQ(_a0 interface{}, querier interface{}) *Service_CanContinueAttackingQ_Call {
	return &Service_CanContinueAttackingQ_Call{Call: _e.mock.On("CanContinueAttackingQ", _a0, querier)}
}

func (_c *Service_CanContinueAttackingQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_CanContinueAttackingQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_CanContinueAttackingQ_Call) Return(_a0 bool, _a1 error) *Service_CanContinueAttackingQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_CanContinueAttackingQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) (bool, error)) *Service_CanContinueAttackingQ_Call {
	_c.Call.Return(run)
	return _c
}

// HasConqueredQ provides a mock function with given fields: _a0, querier
func (_m *Service) HasConqueredQ(_a0 ctx.GameContext, querier db.Querier) (bool, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for HasConqueredQ")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) (bool, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) bool); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_HasConqueredQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasConqueredQ'
type Service_HasConqueredQ_Call struct {
	*mock.Call
}

// HasConqueredQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter) HasConqueredQ(_a0 interface{}, querier interface{}) *Service_HasConqueredQ_Call {
	return &Service_HasConqueredQ_Call{Call: _e.mock.On("HasConqueredQ", _a0, querier)}
}

func (_c *Service_HasConqueredQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_HasConqueredQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_HasConqueredQ_Call) Return(_a0 bool, _a1 error) *Service_HasConqueredQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_HasConqueredQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) (bool, error)) *Service_HasConqueredQ_Call {
	_c.Call.Return(run)
	return _c
}

// PerformQ provides a mock function with given fields: _a0, querier, move
func (_m *Service) PerformQ(_a0 ctx.GameContext, querier db.Querier, move attack.Move) (*attack.MoveResult, error) {
	ret := _m.Called(_a0, querier, move)

	if len(ret) == 0 {
		panic("no return value specified for PerformQ")
	}

	var r0 *attack.MoveResult
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, attack.Move) (*attack.MoveResult, error)); ok {
		return rf(_a0, querier, move)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, attack.Move) *attack.MoveResult); ok {
		r0 = rf(_a0, querier, move)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*attack.MoveResult)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier, attack.Move) error); ok {
		r1 = rf(_a0, querier, move)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_PerformQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PerformQ'
type Service_PerformQ_Call struct {
	*mock.Call
}

// PerformQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
//   - move attack.Move
func (_e *Service_Expecter) PerformQ(_a0 interface{}, querier interface{}, move interface{}) *Service_PerformQ_Call {
	return &Service_PerformQ_Call{Call: _e.mock.On("PerformQ", _a0, querier, move)}
}

func (_c *Service_PerformQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier, move attack.Move)) *Service_PerformQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].(attack.Move))
	})
	return _c
}

func (_c *Service_PerformQ_Call) Return(_a0 *attack.MoveResult, _a1 error) *Service_PerformQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_PerformQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier, attack.Move) (*attack.MoveResult, error)) *Service_PerformQ_Call {
	_c.Call.Return(run)
	return _c
}

// PhaseType provides a mock function with no fields
func (_m *Service) PhaseType() sqlc.PhaseType {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for PhaseType")
	}

	var r0 sqlc.PhaseType
	if rf, ok := ret.Get(0).(func() sqlc.PhaseType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(sqlc.PhaseType)
	}

	return r0
}

// Service_PhaseType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PhaseType'
type Service_PhaseType_Call struct {
	*mock.Call
}

// PhaseType is a helper method to define mock.On call
func (_e *Service_Expecter) PhaseType() *Service_PhaseType_Call {
	return &Service_PhaseType_Call{Call: _e.mock.On("PhaseType")}
}

func (_c *Service_PhaseType_Call) Run(run func()) *Service_PhaseType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Service_PhaseType_Call) Return(_a0 sqlc.PhaseType) *Service_PhaseType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_PhaseType_Call) RunAndReturn(run func() sqlc.PhaseType) *Service_PhaseType_Call {
	_c.Call.Return(run)
	return _c
}

// WalkQ provides a mock function with given fields: _a0, querier, voluntaryAdvancement
func (_m *Service) WalkQ(_a0 ctx.GameContext, querier db.Querier, voluntaryAdvancement bool) (sqlc.PhaseType, error) {
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

// Service_WalkQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WalkQ'
type Service_WalkQ_Call struct {
	*mock.Call
}

// WalkQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
//   - voluntaryAdvancement bool
func (_e *Service_Expecter) WalkQ(_a0 interface{}, querier interface{}, voluntaryAdvancement interface{}) *Service_WalkQ_Call {
	return &Service_WalkQ_Call{Call: _e.mock.On("WalkQ", _a0, querier, voluntaryAdvancement)}
}

func (_c *Service_WalkQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier, voluntaryAdvancement bool)) *Service_WalkQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].(bool))
	})
	return _c
}

func (_c *Service_WalkQ_Call) Return(_a0 sqlc.PhaseType, _a1 error) *Service_WalkQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_WalkQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier, bool) (sqlc.PhaseType, error)) *Service_WalkQ_Call {
	_c.Call.Return(run)
	return _c
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService(t interface {
	mock.TestingT
	Cleanup(func())
}) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
