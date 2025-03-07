// Code generated by mockery v2.50.1. DO NOT EDIT.

package player

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/game/db"

	mock "github.com/stretchr/testify/mock"

	request "github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"

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

// CreatePlayersQ provides a mock function with given fields: _a0, querier, gameID, players
func (_m *Service) CreatePlayersQ(_a0 ctx.GameContext, querier db.Querier, gameID int64, players []request.Player) ([]sqlc.GamePlayer, error) {
	ret := _m.Called(_a0, querier, gameID, players)

	if len(ret) == 0 {
		panic("no return value specified for CreatePlayersQ")
	}

	var r0 []sqlc.GamePlayer
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, int64, []request.Player) ([]sqlc.GamePlayer, error)); ok {
		return rf(_a0, querier, gameID, players)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier, int64, []request.Player) []sqlc.GamePlayer); ok {
		r0 = rf(_a0, querier, gameID, players)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GamePlayer)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier, int64, []request.Player) error); ok {
		r1 = rf(_a0, querier, gameID, players)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_CreatePlayersQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreatePlayersQ'
type Service_CreatePlayersQ_Call struct {
	*mock.Call
}

// CreatePlayersQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
//   - gameID int64
//   - players []request.Player
func (_e *Service_Expecter) CreatePlayersQ(_a0 interface{}, querier interface{}, gameID interface{}, players interface{}) *Service_CreatePlayersQ_Call {
	return &Service_CreatePlayersQ_Call{Call: _e.mock.On("CreatePlayersQ", _a0, querier, gameID, players)}
}

func (_c *Service_CreatePlayersQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier, gameID int64, players []request.Player)) *Service_CreatePlayersQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier), args[2].(int64), args[3].([]request.Player))
	})
	return _c
}

func (_c *Service_CreatePlayersQ_Call) Return(_a0 []sqlc.GamePlayer, _a1 error) *Service_CreatePlayersQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_CreatePlayersQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier, int64, []request.Player) ([]sqlc.GamePlayer, error)) *Service_CreatePlayersQ_Call {
	_c.Call.Return(run)
	return _c
}

// GetCurrentPlayerQ provides a mock function with given fields: _a0, querier
func (_m *Service) GetCurrentPlayerQ(_a0 ctx.GameContext, querier db.Querier) (sqlc.GamePlayer, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for GetCurrentPlayerQ")
	}

	var r0 sqlc.GamePlayer
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) (sqlc.GamePlayer, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) sqlc.GamePlayer); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Get(0).(sqlc.GamePlayer)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetCurrentPlayerQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCurrentPlayerQ'
type Service_GetCurrentPlayerQ_Call struct {
	*mock.Call
}

// GetCurrentPlayerQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter) GetCurrentPlayerQ(_a0 interface{}, querier interface{}) *Service_GetCurrentPlayerQ_Call {
	return &Service_GetCurrentPlayerQ_Call{Call: _e.mock.On("GetCurrentPlayerQ", _a0, querier)}
}

func (_c *Service_GetCurrentPlayerQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_GetCurrentPlayerQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_GetCurrentPlayerQ_Call) Return(_a0 sqlc.GamePlayer, _a1 error) *Service_GetCurrentPlayerQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetCurrentPlayerQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) (sqlc.GamePlayer, error)) *Service_GetCurrentPlayerQ_Call {
	_c.Call.Return(run)
	return _c
}

// GetNextPlayerQ provides a mock function with given fields: _a0, querier
func (_m *Service) GetNextPlayerQ(_a0 ctx.GameContext, querier db.Querier) (sqlc.GamePlayer, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for GetNextPlayerQ")
	}

	var r0 sqlc.GamePlayer
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) (sqlc.GamePlayer, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) sqlc.GamePlayer); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Get(0).(sqlc.GamePlayer)
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetNextPlayerQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNextPlayerQ'
type Service_GetNextPlayerQ_Call struct {
	*mock.Call
}

// GetNextPlayerQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter) GetNextPlayerQ(_a0 interface{}, querier interface{}) *Service_GetNextPlayerQ_Call {
	return &Service_GetNextPlayerQ_Call{Call: _e.mock.On("GetNextPlayerQ", _a0, querier)}
}

func (_c *Service_GetNextPlayerQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_GetNextPlayerQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_GetNextPlayerQ_Call) Return(_a0 sqlc.GamePlayer, _a1 error) *Service_GetNextPlayerQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetNextPlayerQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) (sqlc.GamePlayer, error)) *Service_GetNextPlayerQ_Call {
	_c.Call.Return(run)
	return _c
}

// GetPlayersQ provides a mock function with given fields: _a0, querier
func (_m *Service) GetPlayersQ(_a0 ctx.GameContext, querier db.Querier) ([]sqlc.GamePlayer, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for GetPlayersQ")
	}

	var r0 []sqlc.GamePlayer
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) ([]sqlc.GamePlayer, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) []sqlc.GamePlayer); ok {
		r0 = rf(_a0, querier)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GamePlayer)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetPlayersQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPlayersQ'
type Service_GetPlayersQ_Call struct {
	*mock.Call
}

// GetPlayersQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter) GetPlayersQ(_a0 interface{}, querier interface{}) *Service_GetPlayersQ_Call {
	return &Service_GetPlayersQ_Call{Call: _e.mock.On("GetPlayersQ", _a0, querier)}
}

func (_c *Service_GetPlayersQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_GetPlayersQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_GetPlayersQ_Call) Return(_a0 []sqlc.GamePlayer, _a1 error) *Service_GetPlayersQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetPlayersQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) ([]sqlc.GamePlayer, error)) *Service_GetPlayersQ_Call {
	_c.Call.Return(run)
	return _c
}

// GetPlayersState provides a mock function with given fields: _a0
func (_m *Service) GetPlayersState(_a0 ctx.GameContext) ([]sqlc.GetPlayersStateRow, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetPlayersState")
	}

	var r0 []sqlc.GetPlayersStateRow
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext) ([]sqlc.GetPlayersStateRow, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext) []sqlc.GetPlayersStateRow); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GetPlayersStateRow)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetPlayersState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPlayersState'
type Service_GetPlayersState_Call struct {
	*mock.Call
}

// GetPlayersState is a helper method to define mock.On call
//   - _a0 ctx.GameContext
func (_e *Service_Expecter) GetPlayersState(_a0 interface{}) *Service_GetPlayersState_Call {
	return &Service_GetPlayersState_Call{Call: _e.mock.On("GetPlayersState", _a0)}
}

func (_c *Service_GetPlayersState_Call) Run(run func(_a0 ctx.GameContext)) *Service_GetPlayersState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext))
	})
	return _c
}

func (_c *Service_GetPlayersState_Call) Return(_a0 []sqlc.GetPlayersStateRow, _a1 error) *Service_GetPlayersState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetPlayersState_Call) RunAndReturn(run func(ctx.GameContext) ([]sqlc.GetPlayersStateRow, error)) *Service_GetPlayersState_Call {
	_c.Call.Return(run)
	return _c
}

// GetPlayersStateQ provides a mock function with given fields: _a0, querier
func (_m *Service) GetPlayersStateQ(_a0 ctx.GameContext, querier db.Querier) ([]sqlc.GetPlayersStateRow, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for GetPlayersStateQ")
	}

	var r0 []sqlc.GetPlayersStateRow
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) ([]sqlc.GetPlayersStateRow, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) []sqlc.GetPlayersStateRow); ok {
		r0 = rf(_a0, querier)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GetPlayersStateRow)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetPlayersStateQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPlayersStateQ'
type Service_GetPlayersStateQ_Call struct {
	*mock.Call
}

// GetPlayersStateQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter) GetPlayersStateQ(_a0 interface{}, querier interface{}) *Service_GetPlayersStateQ_Call {
	return &Service_GetPlayersStateQ_Call{Call: _e.mock.On("GetPlayersStateQ", _a0, querier)}
}

func (_c *Service_GetPlayersStateQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_GetPlayersStateQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_GetPlayersStateQ_Call) Return(_a0 []sqlc.GetPlayersStateRow, _a1 error) *Service_GetPlayersStateQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetPlayersStateQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) ([]sqlc.GetPlayersStateRow, error)) *Service_GetPlayersStateQ_Call {
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
