// Code generated by mockery v2.50.1. DO NOT EDIT.

package mission

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
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

// CreateMissionsQ provides a mock function with given fields: _a0, querier, players
func (_m *Service) CreateMissionsQ(_a0 ctx.GameContext, querier db.Querier, players []sqlc.GamePlayer) error {
	ret := _m.Called(_a0, querier, players)

	if len(ret) == 0 {
		panic("no return value specified for CreateMissionsQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, []sqlc.GamePlayer) error); ok {
		r0 = rf(_a0, querier, players)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_CreateMissionsQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateMissionsQ'
type Service_CreateMissionsQ_Call struct {
	*mock.Call
}

// CreateMissionsQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
//   - players []sqlc.GamePlayer
func (_e *Service_Expecter) CreateMissionsQ(_a0 interface{}, querier interface{}, players interface{}) *Service_CreateMissionsQ_Call {
	return &Service_CreateMissionsQ_Call{Call: _e.mock.On("CreateMissionsQ", _a0, querier, players)}
}

func (_c *Service_CreateMissionsQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier, players []sqlc.GamePlayer)) *Service_CreateMissionsQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].([]sqlc.GamePlayer))
	})
	return _c
}

func (_c *Service_CreateMissionsQ_Call) Return(_a0 error) *Service_CreateMissionsQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_CreateMissionsQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier, []sqlc.GamePlayer) error) *Service_CreateMissionsQ_Call {
	_c.Call.Return(run)
	return _c
}

// IsMissionAccomplishedQ provides a mock function with given fields: _a0, querier
func (_m *Service) IsMissionAccomplishedQ(_a0 ctx.GameContext, querier db.Querier) (bool, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for IsMissionAccomplishedQ")
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

// Service_IsMissionAccomplishedQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsMissionAccomplishedQ'
type Service_IsMissionAccomplishedQ_Call struct {
	*mock.Call
}

// IsMissionAccomplishedQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter) IsMissionAccomplishedQ(_a0 interface{}, querier interface{}) *Service_IsMissionAccomplishedQ_Call {
	return &Service_IsMissionAccomplishedQ_Call{Call: _e.mock.On("IsMissionAccomplishedQ", _a0, querier)}
}

func (_c *Service_IsMissionAccomplishedQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_IsMissionAccomplishedQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_IsMissionAccomplishedQ_Call) Return(_a0 bool, _a1 error) *Service_IsMissionAccomplishedQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_IsMissionAccomplishedQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) (bool, error)) *Service_IsMissionAccomplishedQ_Call {
	_c.Call.Return(run)
	return _c
}

// ReassignMissionsQ provides a mock function with given fields: _a0, querier, eliminatedPlayerID
func (_m *Service) ReassignMissionsQ(_a0 ctx.GameContext, querier db.Querier, eliminatedPlayerID int64) error {
	ret := _m.Called(_a0, querier, eliminatedPlayerID)

	if len(ret) == 0 {
		panic("no return value specified for ReassignMissionsQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, int64) error); ok {
		r0 = rf(_a0, querier, eliminatedPlayerID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_ReassignMissionsQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReassignMissionsQ'
type Service_ReassignMissionsQ_Call struct {
	*mock.Call
}

// ReassignMissionsQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
//   - eliminatedPlayerID int64
func (_e *Service_Expecter) ReassignMissionsQ(_a0 interface{}, querier interface{}, eliminatedPlayerID interface{}) *Service_ReassignMissionsQ_Call {
	return &Service_ReassignMissionsQ_Call{Call: _e.mock.On("ReassignMissionsQ", _a0, querier, eliminatedPlayerID)}
}

func (_c *Service_ReassignMissionsQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier, eliminatedPlayerID int64)) *Service_ReassignMissionsQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].(int64))
	})
	return _c
}

func (_c *Service_ReassignMissionsQ_Call) Return(_a0 error) *Service_ReassignMissionsQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_ReassignMissionsQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier, int64) error) *Service_ReassignMissionsQ_Call {
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
