// Code generated by mockery v2.44.1. DO NOT EDIT.

package deploy

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/db"

	deploy "github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"

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
func (_m *Service) AdvanceQ(_a0 ctx.MoveContext, querier db.Querier, targetPhase sqlc.PhaseType, move deploy.Move) error {
	ret := _m.Called(_a0, querier, targetPhase, move)

	if len(ret) == 0 {
		panic("no return value specified for AdvanceQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, sqlc.PhaseType, deploy.Move) error); ok {
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
//   - move deploy.Move
func (_e *Service_Expecter) AdvanceQ(_a0 interface{}, querier interface{}, targetPhase interface{}, move interface{}) *Service_AdvanceQ_Call {
	return &Service_AdvanceQ_Call{Call: _e.mock.On("AdvanceQ", _a0, querier, targetPhase, move)}
}

func (_c *Service_AdvanceQ_Call) Run(run func(_a0 ctx.MoveContext, querier db.Querier, targetPhase sqlc.PhaseType, move deploy.Move)) *Service_AdvanceQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier), args[2].(sqlc.PhaseType), args[3].(deploy.Move))
	})
	return _c
}

func (_c *Service_AdvanceQ_Call) Return(_a0 error) *Service_AdvanceQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_AdvanceQ_Call) RunAndReturn(run func(ctx.MoveContext, db.Querier, sqlc.PhaseType, deploy.Move) error) *Service_AdvanceQ_Call {
	_c.Call.Return(run)
	return _c
}

// GetDeployableTroops provides a mock function with given fields: _a0
func (_m *Service) GetDeployableTroops(_a0 ctx.GameContext) (int64, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetDeployableTroops")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext) (int64, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext) int64); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetDeployableTroops_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDeployableTroops'
type Service_GetDeployableTroops_Call struct {
	*mock.Call
}

// GetDeployableTroops is a helper method to define mock.On call
//   - _a0 ctx.GameContext
func (_e *Service_Expecter) GetDeployableTroops(_a0 interface{}) *Service_GetDeployableTroops_Call {
	return &Service_GetDeployableTroops_Call{Call: _e.mock.On("GetDeployableTroops", _a0)}
}

func (_c *Service_GetDeployableTroops_Call) Run(run func(_a0 ctx.GameContext)) *Service_GetDeployableTroops_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext))
	})
	return _c
}

func (_c *Service_GetDeployableTroops_Call) Return(_a0 int64, _a1 error) *Service_GetDeployableTroops_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetDeployableTroops_Call) RunAndReturn(run func(ctx.GameContext) (int64, error)) *Service_GetDeployableTroops_Call {
	_c.Call.Return(run)
	return _c
}

// GetDeployableTroopsQ provides a mock function with given fields: _a0, querier
func (_m *Service) GetDeployableTroopsQ(_a0 ctx.GameContext, querier db.Querier) (int64, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for GetDeployableTroopsQ")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) (int64, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) int64); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetDeployableTroopsQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDeployableTroopsQ'
type Service_GetDeployableTroopsQ_Call struct {
	*mock.Call
}

// GetDeployableTroopsQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter) GetDeployableTroopsQ(_a0 interface{}, querier interface{}) *Service_GetDeployableTroopsQ_Call {
	return &Service_GetDeployableTroopsQ_Call{Call: _e.mock.On("GetDeployableTroopsQ", _a0, querier)}
}

func (_c *Service_GetDeployableTroopsQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_GetDeployableTroopsQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_GetDeployableTroopsQ_Call) Return(_a0 int64, _a1 error) *Service_GetDeployableTroopsQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetDeployableTroopsQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) (int64, error)) *Service_GetDeployableTroopsQ_Call {
	_c.Call.Return(run)
	return _c
}

// PerformQ provides a mock function with given fields: _a0, querier, move
func (_m *Service) PerformQ(_a0 ctx.MoveContext, querier db.Querier, move deploy.Move) error {
	ret := _m.Called(_a0, querier, move)

	if len(ret) == 0 {
		panic("no return value specified for PerformQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, deploy.Move) error); ok {
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
//   - move deploy.Move
func (_e *Service_Expecter) PerformQ(_a0 interface{}, querier interface{}, move interface{}) *Service_PerformQ_Call {
	return &Service_PerformQ_Call{Call: _e.mock.On("PerformQ", _a0, querier, move)}
}

func (_c *Service_PerformQ_Call) Run(run func(_a0 ctx.MoveContext, querier db.Querier, move deploy.Move)) *Service_PerformQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier), args[2].(deploy.Move))
	})
	return _c
}

func (_c *Service_PerformQ_Call) Return(_a0 error) *Service_PerformQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_PerformQ_Call) RunAndReturn(run func(ctx.MoveContext, db.Querier, deploy.Move) error) *Service_PerformQ_Call {
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
