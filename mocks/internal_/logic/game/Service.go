// Code generated by mockery v2.43.1. DO NOT EDIT.

package game

import (
	context "context"

	board "github.com/go-risk-it/go-risk-it/internal/logic/board"

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

// CreateGame provides a mock function with given fields: ctx, querier, _a2, players
func (_m *Service) CreateGame(ctx context.Context, querier db.Querier, _a2 *board.Board, players []request.Player) (int64, error) {
	ret := _m.Called(ctx, querier, _a2, players)

	if len(ret) == 0 {
		panic("no return value specified for CreateGame")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, db.Querier, *board.Board, []request.Player) (int64, error)); ok {
		return rf(ctx, querier, _a2, players)
	}
	if rf, ok := ret.Get(0).(func(context.Context, db.Querier, *board.Board, []request.Player) int64); ok {
		r0 = rf(ctx, querier, _a2, players)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, db.Querier, *board.Board, []request.Player) error); ok {
		r1 = rf(ctx, querier, _a2, players)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_CreateGame_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateGame'
type Service_CreateGame_Call struct {
	*mock.Call
}

// CreateGame is a helper method to define mock.On call
//   - ctx context.Context
//   - querier db.Querier
//   - _a2 *board.Board
//   - players []request.Player
func (_e *Service_Expecter) CreateGame(ctx interface{}, querier interface{}, _a2 interface{}, players interface{}) *Service_CreateGame_Call {
	return &Service_CreateGame_Call{Call: _e.mock.On("CreateGame", ctx, querier, _a2, players)}
}

func (_c *Service_CreateGame_Call) Run(run func(ctx context.Context, querier db.Querier, _a2 *board.Board, players []request.Player)) *Service_CreateGame_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(db.Querier), args[2].(*board.Board), args[3].([]request.Player))
	})
	return _c
}

func (_c *Service_CreateGame_Call) Return(_a0 int64, _a1 error) *Service_CreateGame_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_CreateGame_Call) RunAndReturn(run func(context.Context, db.Querier, *board.Board, []request.Player) (int64, error)) *Service_CreateGame_Call {
	_c.Call.Return(run)
	return _c
}

// CreateGameWithTx provides a mock function with given fields: ctx, _a1, players
func (_m *Service) CreateGameWithTx(ctx context.Context, _a1 *board.Board, players []request.Player) (int64, error) {
	ret := _m.Called(ctx, _a1, players)

	if len(ret) == 0 {
		panic("no return value specified for CreateGameWithTx")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *board.Board, []request.Player) (int64, error)); ok {
		return rf(ctx, _a1, players)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *board.Board, []request.Player) int64); ok {
		r0 = rf(ctx, _a1, players)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *board.Board, []request.Player) error); ok {
		r1 = rf(ctx, _a1, players)
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
//   - ctx context.Context
//   - _a1 *board.Board
//   - players []request.Player
func (_e *Service_Expecter) CreateGameWithTx(ctx interface{}, _a1 interface{}, players interface{}) *Service_CreateGameWithTx_Call {
	return &Service_CreateGameWithTx_Call{Call: _e.mock.On("CreateGameWithTx", ctx, _a1, players)}
}

func (_c *Service_CreateGameWithTx_Call) Run(run func(ctx context.Context, _a1 *board.Board, players []request.Player)) *Service_CreateGameWithTx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*board.Board), args[2].([]request.Player))
	})
	return _c
}

func (_c *Service_CreateGameWithTx_Call) Return(_a0 int64, _a1 error) *Service_CreateGameWithTx_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_CreateGameWithTx_Call) RunAndReturn(run func(context.Context, *board.Board, []request.Player) (int64, error)) *Service_CreateGameWithTx_Call {
	_c.Call.Return(run)
	return _c
}

// DecreaseDeployableTroopsQ provides a mock function with given fields: ctx, querier, _a2, troops
func (_m *Service) DecreaseDeployableTroopsQ(ctx context.Context, querier db.Querier, _a2 *sqlc.Game, troops int64) error {
	ret := _m.Called(ctx, querier, _a2, troops)

	if len(ret) == 0 {
		panic("no return value specified for DecreaseDeployableTroopsQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, db.Querier, *sqlc.Game, int64) error); ok {
		r0 = rf(ctx, querier, _a2, troops)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_DecreaseDeployableTroopsQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DecreaseDeployableTroopsQ'
type Service_DecreaseDeployableTroopsQ_Call struct {
	*mock.Call
}

// DecreaseDeployableTroopsQ is a helper method to define mock.On call
//   - ctx context.Context
//   - querier db.Querier
//   - _a2 *sqlc.Game
//   - troops int64
func (_e *Service_Expecter) DecreaseDeployableTroopsQ(ctx interface{}, querier interface{}, _a2 interface{}, troops interface{}) *Service_DecreaseDeployableTroopsQ_Call {
	return &Service_DecreaseDeployableTroopsQ_Call{Call: _e.mock.On("DecreaseDeployableTroopsQ", ctx, querier, _a2, troops)}
}

func (_c *Service_DecreaseDeployableTroopsQ_Call) Run(run func(ctx context.Context, querier db.Querier, _a2 *sqlc.Game, troops int64)) *Service_DecreaseDeployableTroopsQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(db.Querier), args[2].(*sqlc.Game), args[3].(int64))
	})
	return _c
}

func (_c *Service_DecreaseDeployableTroopsQ_Call) Return(_a0 error) *Service_DecreaseDeployableTroopsQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_DecreaseDeployableTroopsQ_Call) RunAndReturn(run func(context.Context, db.Querier, *sqlc.Game, int64) error) *Service_DecreaseDeployableTroopsQ_Call {
	_c.Call.Return(run)
	return _c
}

// GetGameState provides a mock function with given fields: ctx, gameID
func (_m *Service) GetGameState(ctx context.Context, gameID int64) (*sqlc.Game, error) {
	ret := _m.Called(ctx, gameID)

	if len(ret) == 0 {
		panic("no return value specified for GetGameState")
	}

	var r0 *sqlc.Game
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (*sqlc.Game, error)); ok {
		return rf(ctx, gameID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) *sqlc.Game); ok {
		r0 = rf(ctx, gameID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sqlc.Game)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, gameID)
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
//   - ctx context.Context
//   - gameID int64
func (_e *Service_Expecter) GetGameState(ctx interface{}, gameID interface{}) *Service_GetGameState_Call {
	return &Service_GetGameState_Call{Call: _e.mock.On("GetGameState", ctx, gameID)}
}

func (_c *Service_GetGameState_Call) Run(run func(ctx context.Context, gameID int64)) *Service_GetGameState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *Service_GetGameState_Call) Return(_a0 *sqlc.Game, _a1 error) *Service_GetGameState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetGameState_Call) RunAndReturn(run func(context.Context, int64) (*sqlc.Game, error)) *Service_GetGameState_Call {
	_c.Call.Return(run)
	return _c
}

// GetGameStateQ provides a mock function with given fields: ctx, querier, gameID
func (_m *Service) GetGameStateQ(ctx context.Context, querier db.Querier, gameID int64) (*sqlc.Game, error) {
	ret := _m.Called(ctx, querier, gameID)

	if len(ret) == 0 {
		panic("no return value specified for GetGameStateQ")
	}

	var r0 *sqlc.Game
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, db.Querier, int64) (*sqlc.Game, error)); ok {
		return rf(ctx, querier, gameID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, db.Querier, int64) *sqlc.Game); ok {
		r0 = rf(ctx, querier, gameID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sqlc.Game)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, db.Querier, int64) error); ok {
		r1 = rf(ctx, querier, gameID)
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
//   - ctx context.Context
//   - querier db.Querier
//   - gameID int64
func (_e *Service_Expecter) GetGameStateQ(ctx interface{}, querier interface{}, gameID interface{}) *Service_GetGameStateQ_Call {
	return &Service_GetGameStateQ_Call{Call: _e.mock.On("GetGameStateQ", ctx, querier, gameID)}
}

func (_c *Service_GetGameStateQ_Call) Run(run func(ctx context.Context, querier db.Querier, gameID int64)) *Service_GetGameStateQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(db.Querier), args[2].(int64))
	})
	return _c
}

func (_c *Service_GetGameStateQ_Call) Return(_a0 *sqlc.Game, _a1 error) *Service_GetGameStateQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetGameStateQ_Call) RunAndReturn(run func(context.Context, db.Querier, int64) (*sqlc.Game, error)) *Service_GetGameStateQ_Call {
	_c.Call.Return(run)
	return _c
}

// SetGamePhaseQ provides a mock function with given fields: ctx, querier, gameID, phase
func (_m *Service) SetGamePhaseQ(ctx context.Context, querier db.Querier, gameID int64, phase sqlc.Phase) error {
	ret := _m.Called(ctx, querier, gameID, phase)

	if len(ret) == 0 {
		panic("no return value specified for SetGamePhaseQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, db.Querier, int64, sqlc.Phase) error); ok {
		r0 = rf(ctx, querier, gameID, phase)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_SetGamePhaseQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetGamePhaseQ'
type Service_SetGamePhaseQ_Call struct {
	*mock.Call
}

// SetGamePhaseQ is a helper method to define mock.On call
//   - ctx context.Context
//   - querier db.Querier
//   - gameID int64
//   - phase sqlc.Phase
func (_e *Service_Expecter) SetGamePhaseQ(ctx interface{}, querier interface{}, gameID interface{}, phase interface{}) *Service_SetGamePhaseQ_Call {
	return &Service_SetGamePhaseQ_Call{Call: _e.mock.On("SetGamePhaseQ", ctx, querier, gameID, phase)}
}

func (_c *Service_SetGamePhaseQ_Call) Run(run func(ctx context.Context, querier db.Querier, gameID int64, phase sqlc.Phase)) *Service_SetGamePhaseQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(db.Querier), args[2].(int64), args[3].(sqlc.Phase))
	})
	return _c
}

func (_c *Service_SetGamePhaseQ_Call) Return(_a0 error) *Service_SetGamePhaseQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_SetGamePhaseQ_Call) RunAndReturn(run func(context.Context, db.Querier, int64, sqlc.Phase) error) *Service_SetGamePhaseQ_Call {
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
