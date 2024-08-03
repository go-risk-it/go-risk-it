// Code generated by mockery v2.44.1. DO NOT EDIT.

package attack

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	attack "github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"

	db "github.com/go-risk-it/go-risk-it/internal/data/db"

	mock "github.com/stretchr/testify/mock"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
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

// AdvanceQ provides a mock function with given fields: _a0, querier, targetPhase, move
func (_m *Service) AdvanceQ(_a0 ctx.MoveContext, querier db.Querier, targetPhase sqlc.PhaseType, move attack.Move) error {
	ret := _m.Called(_a0, querier, targetPhase, move)

	if len(ret) == 0 {
		panic("no return value specified for AdvanceQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, sqlc.PhaseType, attack.Move) error); ok {
		r0 = rf(_a0, querier, targetPhase, move)
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
//   - _a0 ctx.MoveContext
//   - querier db.Querier
//   - targetPhase sqlc.PhaseType
//   - move attack.Move
func (_e *Service_Expecter) AdvanceQ(_a0 interface{}, querier interface{}, targetPhase interface{}, move interface{}) *Service_AdvanceQ_Call {
	return &Service_AdvanceQ_Call{Call: _e.mock.On("AdvanceQ", _a0, querier, targetPhase, move)}
}

func (_c *Service_AdvanceQ_Call) Run(run func(_a0 ctx.MoveContext, querier db.Querier, targetPhase sqlc.PhaseType, move attack.Move)) *Service_AdvanceQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier), args[2].(sqlc.PhaseType), args[3].(attack.Move))
	})
	return _c
}

func (_c *Service_AdvanceQ_Call) Return(_a0 error) *Service_AdvanceQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_AdvanceQ_Call) RunAndReturn(run func(ctx.MoveContext, db.Querier, sqlc.PhaseType, attack.Move) error) *Service_AdvanceQ_Call {
	_c.Call.Return(run)
	return _c
}

// CanContinueAttackingQ provides a mock function with given fields: _a0, querier
func (_m *Service) CanContinueAttackingQ(_a0 ctx.MoveContext, querier db.Querier) (bool, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for CanContinueAttackingQ")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier) (bool, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier) bool); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(ctx.MoveContext, db.Querier) error); ok {
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
//   - _a0 ctx.MoveContext
//   - querier db.Querier
func (_e *Service_Expecter) CanContinueAttackingQ(_a0 interface{}, querier interface{}) *Service_CanContinueAttackingQ_Call {
	return &Service_CanContinueAttackingQ_Call{Call: _e.mock.On("CanContinueAttackingQ", _a0, querier)}
}

func (_c *Service_CanContinueAttackingQ_Call) Run(run func(_a0 ctx.MoveContext, querier db.Querier)) *Service_CanContinueAttackingQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_CanContinueAttackingQ_Call) Return(_a0 bool, _a1 error) *Service_CanContinueAttackingQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_CanContinueAttackingQ_Call) RunAndReturn(run func(ctx.MoveContext, db.Querier) (bool, error)) *Service_CanContinueAttackingQ_Call {
	_c.Call.Return(run)
	return _c
}

// HasConqueredQ provides a mock function with given fields: _a0, querier
func (_m *Service) HasConqueredQ(_a0 ctx.MoveContext, querier db.Querier) (bool, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for HasConqueredQ")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier) (bool, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier) bool); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(ctx.MoveContext, db.Querier) error); ok {
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
//   - _a0 ctx.MoveContext
//   - querier db.Querier
func (_e *Service_Expecter) HasConqueredQ(_a0 interface{}, querier interface{}) *Service_HasConqueredQ_Call {
	return &Service_HasConqueredQ_Call{Call: _e.mock.On("HasConqueredQ", _a0, querier)}
}

func (_c *Service_HasConqueredQ_Call) Run(run func(_a0 ctx.MoveContext, querier db.Querier)) *Service_HasConqueredQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_HasConqueredQ_Call) Return(_a0 bool, _a1 error) *Service_HasConqueredQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_HasConqueredQ_Call) RunAndReturn(run func(ctx.MoveContext, db.Querier) (bool, error)) *Service_HasConqueredQ_Call {
	_c.Call.Return(run)
	return _c
}

// PerformQ provides a mock function with given fields: _a0, querier, move
func (_m *Service) PerformQ(_a0 ctx.MoveContext, querier db.Querier, move attack.Move) error {
	ret := _m.Called(_a0, querier, move)

	if len(ret) == 0 {
		panic("no return value specified for PerformQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, attack.Move) error); ok {
		r0 = rf(_a0, querier, move)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_PerformQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PerformQ'
type Service_PerformQ_Call struct {
	*mock.Call
}

// PerformQ is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - querier db.Querier
//   - move attack.Move
func (_e *Service_Expecter) PerformQ(_a0 interface{}, querier interface{}, move interface{}) *Service_PerformQ_Call {
	return &Service_PerformQ_Call{Call: _e.mock.On("PerformQ", _a0, querier, move)}
}

func (_c *Service_PerformQ_Call) Run(run func(_a0 ctx.MoveContext, querier db.Querier, move attack.Move)) *Service_PerformQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier), args[2].(attack.Move))
	})
	return _c
}

func (_c *Service_PerformQ_Call) Return(_a0 error) *Service_PerformQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_PerformQ_Call) RunAndReturn(run func(ctx.MoveContext, db.Querier, attack.Move) error) *Service_PerformQ_Call {
	_c.Call.Return(run)
	return _c
}

// Walk provides a mock function with given fields: _a0, querier
func (_m *Service) Walk(_a0 ctx.MoveContext, querier db.Querier) (sqlc.PhaseType, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for Walk")
	}

	var r0 sqlc.PhaseType
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier) (sqlc.PhaseType, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier) sqlc.PhaseType); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Get(0).(sqlc.PhaseType)
	}

	if rf, ok := ret.Get(1).(func(ctx.MoveContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_Walk_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Walk'
type Service_Walk_Call struct {
	*mock.Call
}

// Walk is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - querier db.Querier
func (_e *Service_Expecter) Walk(_a0 interface{}, querier interface{}) *Service_Walk_Call {
	return &Service_Walk_Call{Call: _e.mock.On("Walk", _a0, querier)}
}

func (_c *Service_Walk_Call) Run(run func(_a0 ctx.MoveContext, querier db.Querier)) *Service_Walk_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_Walk_Call) Return(_a0 sqlc.PhaseType, _a1 error) *Service_Walk_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_Walk_Call) RunAndReturn(run func(ctx.MoveContext, db.Querier) (sqlc.PhaseType, error)) *Service_Walk_Call {
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