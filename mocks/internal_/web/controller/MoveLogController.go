// Code generated by mockery v2.50.1. DO NOT EDIT.

package controller

import (
	messaging "github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	mock "github.com/stretchr/testify/mock"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

// MoveLogController is an autogenerated mock type for the MoveLogController type
type MoveLogController struct {
	mock.Mock
}

type MoveLogController_Expecter struct {
	mock *mock.Mock
}

func (_m *MoveLogController) EXPECT() *MoveLogController_Expecter {
	return &MoveLogController_Expecter{mock: &_m.Mock}
}

// ConvertMoveLogs provides a mock function with given fields: _a0, sqlcLogs
func (_m *MoveLogController) ConvertMoveLogs(_a0 ctx.GameContext, sqlcLogs []sqlc.MoveLog) (messaging.MoveHistory, error) {
	ret := _m.Called(_a0, sqlcLogs)

	if len(ret) == 0 {
		panic("no return value specified for ConvertMoveLogs")
	}

	var r0 messaging.MoveHistory
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, []sqlc.MoveLog) (messaging.MoveHistory, error)); ok {
		return rf(_a0, sqlcLogs)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, []sqlc.MoveLog) messaging.MoveHistory); ok {
		r0 = rf(_a0, sqlcLogs)
	} else {
		r0 = ret.Get(0).(messaging.MoveHistory)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, []sqlc.MoveLog) error); ok {
		r1 = rf(_a0, sqlcLogs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MoveLogController_ConvertMoveLogs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ConvertMoveLogs'
type MoveLogController_ConvertMoveLogs_Call struct {
	*mock.Call
}

// ConvertMoveLogs is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - sqlcLogs []sqlc.MoveLog
func (_e *MoveLogController_Expecter) ConvertMoveLogs(_a0 interface{}, sqlcLogs interface{}) *MoveLogController_ConvertMoveLogs_Call {
	return &MoveLogController_ConvertMoveLogs_Call{Call: _e.mock.On("ConvertMoveLogs", _a0, sqlcLogs)}
}

func (_c *MoveLogController_ConvertMoveLogs_Call) Run(run func(_a0 ctx.GameContext, sqlcLogs []sqlc.MoveLog)) *MoveLogController_ConvertMoveLogs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].([]sqlc.MoveLog))
	})
	return _c
}

func (_c *MoveLogController_ConvertMoveLogs_Call) Return(_a0 messaging.MoveHistory, _a1 error) *MoveLogController_ConvertMoveLogs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MoveLogController_ConvertMoveLogs_Call) RunAndReturn(run func(ctx.GameContext, []sqlc.MoveLog) (messaging.MoveHistory, error)) *MoveLogController_ConvertMoveLogs_Call {
	_c.Call.Return(run)
	return _c
}

// GetMoveLogs provides a mock function with given fields: _a0, limit
func (_m *MoveLogController) GetMoveLogs(_a0 ctx.GameContext, limit int64) (messaging.MoveHistory, error) {
	ret := _m.Called(_a0, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetMoveLogs")
	}

	var r0 messaging.MoveHistory
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, int64) (messaging.MoveHistory, error)); ok {
		return rf(_a0, limit)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, int64) messaging.MoveHistory); ok {
		r0 = rf(_a0, limit)
	} else {
		r0 = ret.Get(0).(messaging.MoveHistory)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, int64) error); ok {
		r1 = rf(_a0, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MoveLogController_GetMoveLogs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetMoveLogs'
type MoveLogController_GetMoveLogs_Call struct {
	*mock.Call
}

// GetMoveLogs is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - limit int64
func (_e *MoveLogController_Expecter) GetMoveLogs(_a0 interface{}, limit interface{}) *MoveLogController_GetMoveLogs_Call {
	return &MoveLogController_GetMoveLogs_Call{Call: _e.mock.On("GetMoveLogs", _a0, limit)}
}

func (_c *MoveLogController_GetMoveLogs_Call) Run(run func(_a0 ctx.GameContext, limit int64)) *MoveLogController_GetMoveLogs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(int64))
	})
	return _c
}

func (_c *MoveLogController_GetMoveLogs_Call) Return(_a0 messaging.MoveHistory, _a1 error) *MoveLogController_GetMoveLogs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MoveLogController_GetMoveLogs_Call) RunAndReturn(run func(ctx.GameContext, int64) (messaging.MoveHistory, error)) *MoveLogController_GetMoveLogs_Call {
	_c.Call.Return(run)
	return _c
}

// NewMoveLogController creates a new instance of MoveLogController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMoveLogController(t interface {
	mock.TestingT
	Cleanup(func())
}) *MoveLogController {
	mock := &MoveLogController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}