// Code generated by mockery v2.43.1. DO NOT EDIT.

package gamestate

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/db"

	mock "github.com/stretchr/testify/mock"

	request "github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"

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

// CreateGameQ provides a mock function with given fields: _a0, querier, regions, players
func (_m *Service) CreateGameQ(_a0 ctx.UserContext, querier db.Querier, regions []string, players []request.Player) (int64, error) {
	ret := _m.Called(_a0, querier, regions, players)

	if len(ret) == 0 {
		panic("no return value specified for CreateGameQ")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.UserContext, db.Querier, []string, []request.Player) (int64, error)); ok {
		return rf(_a0, querier, regions, players)
	}
	if rf, ok := ret.Get(0).(func(ctx.UserContext, db.Querier, []string, []request.Player) int64); ok {
		r0 = rf(_a0, querier, regions, players)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(ctx.UserContext, db.Querier, []string, []request.Player) error); ok {
		r1 = rf(_a0, querier, regions, players)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_CreateGameQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateGameQ'
type Service_CreateGameQ_Call struct {
	*mock.Call
}

// CreateGameQ is a helper method to define mock.On call
//   - _a0 ctx.UserContext
//   - querier db.Querier
//   - regions []string
//   - players []request.Player
func (_e *Service_Expecter) CreateGameQ(_a0 interface{}, querier interface{}, regions interface{}, players interface{}) *Service_CreateGameQ_Call {
	return &Service_CreateGameQ_Call{Call: _e.mock.On("CreateGameQ", _a0, querier, regions, players)}
}

func (_c *Service_CreateGameQ_Call) Run(run func(_a0 ctx.UserContext, querier db.Querier, regions []string, players []request.Player)) *Service_CreateGameQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.UserContext), args[1].(db.Querier), args[2].([]string), args[3].([]request.Player))
	})
	return _c
}

func (_c *Service_CreateGameQ_Call) Return(_a0 int64, _a1 error) *Service_CreateGameQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_CreateGameQ_Call) RunAndReturn(run func(ctx.UserContext, db.Querier, []string, []request.Player) (int64, error)) *Service_CreateGameQ_Call {
	_c.Call.Return(run)
	return _c
}

// CreateGameWithTx provides a mock function with given fields: _a0, regions, players
func (_m *Service) CreateGameWithTx(_a0 ctx.UserContext, regions []string, players []request.Player) (int64, error) {
	ret := _m.Called(_a0, regions, players)

	if len(ret) == 0 {
		panic("no return value specified for CreateGameWithTx")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.UserContext, []string, []request.Player) (int64, error)); ok {
		return rf(_a0, regions, players)
	}
	if rf, ok := ret.Get(0).(func(ctx.UserContext, []string, []request.Player) int64); ok {
		r0 = rf(_a0, regions, players)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(ctx.UserContext, []string, []request.Player) error); ok {
		r1 = rf(_a0, regions, players)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_CreateGameWithTx_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateGameWithTx'
type Service_CreateGameWithTx_Call struct {
	*mock.Call
}

// CreateGameWithTx is a helper method to define mock.On call
//   - _a0 ctx.UserContext
//   - regions []string
//   - players []request.Player
func (_e *Service_Expecter) CreateGameWithTx(_a0 interface{}, regions interface{}, players interface{}) *Service_CreateGameWithTx_Call {
	return &Service_CreateGameWithTx_Call{Call: _e.mock.On("CreateGameWithTx", _a0, regions, players)}
}

func (_c *Service_CreateGameWithTx_Call) Run(run func(_a0 ctx.UserContext, regions []string, players []request.Player)) *Service_CreateGameWithTx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.UserContext), args[1].([]string), args[2].([]request.Player))
	})
	return _c
}

func (_c *Service_CreateGameWithTx_Call) Return(_a0 int64, _a1 error) *Service_CreateGameWithTx_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_CreateGameWithTx_Call) RunAndReturn(run func(ctx.UserContext, []string, []request.Player) (int64, error)) *Service_CreateGameWithTx_Call {
	_c.Call.Return(run)
	return _c
}

// GetGameState provides a mock function with given fields: _a0
func (_m *Service) GetGameState(_a0 ctx.GameContext) (*sqlc.Game, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetGameState")
	}

	var r0 *sqlc.Game
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext) (*sqlc.Game, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext) *sqlc.Game); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sqlc.Game)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetGameState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGameState'
type Service_GetGameState_Call struct {
	*mock.Call
}

// GetGameState is a helper method to define mock.On call
//   - _a0 ctx.GameContext
func (_e *Service_Expecter) GetGameState(_a0 interface{}) *Service_GetGameState_Call {
	return &Service_GetGameState_Call{Call: _e.mock.On("GetGameState", _a0)}
}

func (_c *Service_GetGameState_Call) Run(run func(_a0 ctx.GameContext)) *Service_GetGameState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext))
	})
	return _c
}

func (_c *Service_GetGameState_Call) Return(_a0 *sqlc.Game, _a1 error) *Service_GetGameState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetGameState_Call) RunAndReturn(run func(ctx.GameContext) (*sqlc.Game, error)) *Service_GetGameState_Call {
	_c.Call.Return(run)
	return _c
}

// GetGameStateQ provides a mock function with given fields: _a0, querier
func (_m *Service) GetGameStateQ(_a0 ctx.GameContext, querier db.Querier) (*sqlc.Game, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for GetGameStateQ")
	}

	var r0 *sqlc.Game
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) (*sqlc.Game, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) *sqlc.Game); ok {
		r0 = rf(_a0, querier)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sqlc.Game)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetGameStateQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGameStateQ'
type Service_GetGameStateQ_Call struct {
	*mock.Call
}

// GetGameStateQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter) GetGameStateQ(_a0 interface{}, querier interface{}) *Service_GetGameStateQ_Call {
	return &Service_GetGameStateQ_Call{Call: _e.mock.On("GetGameStateQ", _a0, querier)}
}

func (_c *Service_GetGameStateQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_GetGameStateQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_GetGameStateQ_Call) Return(_a0 *sqlc.Game, _a1 error) *Service_GetGameStateQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetGameStateQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) (*sqlc.Game, error)) *Service_GetGameStateQ_Call {
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
