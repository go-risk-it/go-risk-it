// Code generated by mockery v2.50.1. DO NOT EDIT.

package deploy

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/game/db"

	deploy "github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"

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
func (_m *Service) AdvanceQ(_a0 ctx.GameContext, querier db.Querier, targetPhase sqlc.PhaseType, performResult *deploy.MoveResult) error {
	ret := _m.Called(_a0, querier, targetPhase, performResult)

	if len(ret) == 0 {
		panic("no return value specified for AdvanceQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, sqlc.PhaseType, *deploy.MoveResult) error); ok {
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
//   - performResult *deploy.MoveResult
func (_e *Service_Expecter) AdvanceQ(_a0 interface{}, querier interface{}, targetPhase interface{}, performResult interface{}) *Service_AdvanceQ_Call {
	return &Service_AdvanceQ_Call{Call: _e.mock.On("AdvanceQ", _a0, querier, targetPhase, performResult)}
}

func (_c *Service_AdvanceQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier, targetPhase sqlc.PhaseType, performResult *deploy.MoveResult)) *Service_AdvanceQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].(sqlc.PhaseType), args[3].(*deploy.MoveResult))
	})
	return _c
}

func (_c *Service_AdvanceQ_Call) Return(_a0 error) *Service_AdvanceQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_AdvanceQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier, sqlc.PhaseType, *deploy.MoveResult) error) *Service_AdvanceQ_Call {
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
func (_m *Service) PerformQ(_a0 ctx.GameContext, querier db.Querier, move deploy.Move) (*deploy.MoveResult, error) {
	ret := _m.Called(_a0, querier, move)

	if len(ret) == 0 {
		panic("no return value specified for PerformQ")
	}

	var r0 *deploy.MoveResult
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, deploy.Move) (*deploy.MoveResult, error)); ok {
		return rf(_a0, querier, move)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, deploy.Move) *deploy.MoveResult); ok {
		r0 = rf(_a0, querier, move)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*deploy.MoveResult)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier, deploy.Move) error); ok {
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
//   - move deploy.Move
func (_e *Service_Expecter) PerformQ(_a0 interface{}, querier interface{}, move interface{}) *Service_PerformQ_Call {
	return &Service_PerformQ_Call{Call: _e.mock.On("PerformQ", _a0, querier, move)}
}

func (_c *Service_PerformQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier, move deploy.Move)) *Service_PerformQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].(deploy.Move))
	})
	return _c
}

func (_c *Service_PerformQ_Call) Return(_a0 *deploy.MoveResult, _a1 error) *Service_PerformQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_PerformQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier, deploy.Move) (*deploy.MoveResult, error)) *Service_PerformQ_Call {
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
