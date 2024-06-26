// Code generated by mockery v2.43.1. DO NOT EDIT.

package sqlc

import (
	context "context"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	mock "github.com/stretchr/testify/mock"
)

// Querier is an autogenerated mock type for the Querier type
type Querier struct {
	mock.Mock
}

type Querier_Expecter struct {
	mock *mock.Mock
}

func (_m *Querier) EXPECT() *Querier_Expecter {
	return &Querier_Expecter{mock: &_m.Mock}
}

// DecreaseDeployableTroops provides a mock function with given fields: ctx, arg
func (_m *Querier) DecreaseDeployableTroops(ctx context.Context, arg sqlc.DecreaseDeployableTroopsParams) error {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for DecreaseDeployableTroops")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.DecreaseDeployableTroopsParams) error); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Querier_DecreaseDeployableTroops_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DecreaseDeployableTroops'
type Querier_DecreaseDeployableTroops_Call struct {
	*mock.Call
}

// DecreaseDeployableTroops is a helper method to define mock.On call
//   - ctx context.Context
//   - arg sqlc.DecreaseDeployableTroopsParams
func (_e *Querier_Expecter) DecreaseDeployableTroops(ctx interface{}, arg interface{}) *Querier_DecreaseDeployableTroops_Call {
	return &Querier_DecreaseDeployableTroops_Call{Call: _e.mock.On("DecreaseDeployableTroops", ctx, arg)}
}

func (_c *Querier_DecreaseDeployableTroops_Call) Run(run func(ctx context.Context, arg sqlc.DecreaseDeployableTroopsParams)) *Querier_DecreaseDeployableTroops_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(sqlc.DecreaseDeployableTroopsParams))
	})
	return _c
}

func (_c *Querier_DecreaseDeployableTroops_Call) Return(_a0 error) *Querier_DecreaseDeployableTroops_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Querier_DecreaseDeployableTroops_Call) RunAndReturn(run func(context.Context, sqlc.DecreaseDeployableTroopsParams) error) *Querier_DecreaseDeployableTroops_Call {
	_c.Call.Return(run)
	return _c
}

// GetGame provides a mock function with given fields: ctx, id
func (_m *Querier) GetGame(ctx context.Context, id int64) (sqlc.Game, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetGame")
	}

	var r0 sqlc.Game
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (sqlc.Game, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) sqlc.Game); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(sqlc.Game)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_GetGame_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGame'
type Querier_GetGame_Call struct {
	*mock.Call
}

// GetGame is a helper method to define mock.On call
//   - ctx context.Context
//   - id int64
func (_e *Querier_Expecter) GetGame(ctx interface{}, id interface{}) *Querier_GetGame_Call {
	return &Querier_GetGame_Call{Call: _e.mock.On("GetGame", ctx, id)}
}

func (_c *Querier_GetGame_Call) Run(run func(ctx context.Context, id int64)) *Querier_GetGame_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *Querier_GetGame_Call) Return(_a0 sqlc.Game, _a1 error) *Querier_GetGame_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_GetGame_Call) RunAndReturn(run func(context.Context, int64) (sqlc.Game, error)) *Querier_GetGame_Call {
	_c.Call.Return(run)
	return _c
}

// GetPlayerByUserId provides a mock function with given fields: ctx, userID
func (_m *Querier) GetPlayerByUserId(ctx context.Context, userID string) (sqlc.Player, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetPlayerByUserId")
	}

	var r0 sqlc.Player
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (sqlc.Player, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) sqlc.Player); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Get(0).(sqlc.Player)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_GetPlayerByUserId_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPlayerByUserId'
type Querier_GetPlayerByUserId_Call struct {
	*mock.Call
}

// GetPlayerByUserId is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
func (_e *Querier_Expecter) GetPlayerByUserId(ctx interface{}, userID interface{}) *Querier_GetPlayerByUserId_Call {
	return &Querier_GetPlayerByUserId_Call{Call: _e.mock.On("GetPlayerByUserId", ctx, userID)}
}

func (_c *Querier_GetPlayerByUserId_Call) Run(run func(ctx context.Context, userID string)) *Querier_GetPlayerByUserId_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Querier_GetPlayerByUserId_Call) Return(_a0 sqlc.Player, _a1 error) *Querier_GetPlayerByUserId_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_GetPlayerByUserId_Call) RunAndReturn(run func(context.Context, string) (sqlc.Player, error)) *Querier_GetPlayerByUserId_Call {
	_c.Call.Return(run)
	return _c
}

// GetPlayersByGame provides a mock function with given fields: ctx, gameID
func (_m *Querier) GetPlayersByGame(ctx context.Context, gameID int64) ([]sqlc.Player, error) {
	ret := _m.Called(ctx, gameID)

	if len(ret) == 0 {
		panic("no return value specified for GetPlayersByGame")
	}

	var r0 []sqlc.Player
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) ([]sqlc.Player, error)); ok {
		return rf(ctx, gameID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) []sqlc.Player); ok {
		r0 = rf(ctx, gameID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.Player)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, gameID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_GetPlayersByGame_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPlayersByGame'
type Querier_GetPlayersByGame_Call struct {
	*mock.Call
}

// GetPlayersByGame is a helper method to define mock.On call
//   - ctx context.Context
//   - gameID int64
func (_e *Querier_Expecter) GetPlayersByGame(ctx interface{}, gameID interface{}) *Querier_GetPlayersByGame_Call {
	return &Querier_GetPlayersByGame_Call{Call: _e.mock.On("GetPlayersByGame", ctx, gameID)}
}

func (_c *Querier_GetPlayersByGame_Call) Run(run func(ctx context.Context, gameID int64)) *Querier_GetPlayersByGame_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *Querier_GetPlayersByGame_Call) Return(_a0 []sqlc.Player, _a1 error) *Querier_GetPlayersByGame_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_GetPlayersByGame_Call) RunAndReturn(run func(context.Context, int64) ([]sqlc.Player, error)) *Querier_GetPlayersByGame_Call {
	_c.Call.Return(run)
	return _c
}

// GetRegionsByGame provides a mock function with given fields: ctx, id
func (_m *Querier) GetRegionsByGame(ctx context.Context, id int64) ([]sqlc.GetRegionsByGameRow, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetRegionsByGame")
	}

	var r0 []sqlc.GetRegionsByGameRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) ([]sqlc.GetRegionsByGameRow, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) []sqlc.GetRegionsByGameRow); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GetRegionsByGameRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_GetRegionsByGame_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRegionsByGame'
type Querier_GetRegionsByGame_Call struct {
	*mock.Call
}

// GetRegionsByGame is a helper method to define mock.On call
//   - ctx context.Context
//   - id int64
func (_e *Querier_Expecter) GetRegionsByGame(ctx interface{}, id interface{}) *Querier_GetRegionsByGame_Call {
	return &Querier_GetRegionsByGame_Call{Call: _e.mock.On("GetRegionsByGame", ctx, id)}
}

func (_c *Querier_GetRegionsByGame_Call) Run(run func(ctx context.Context, id int64)) *Querier_GetRegionsByGame_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *Querier_GetRegionsByGame_Call) Return(_a0 []sqlc.GetRegionsByGameRow, _a1 error) *Querier_GetRegionsByGame_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_GetRegionsByGame_Call) RunAndReturn(run func(context.Context, int64) ([]sqlc.GetRegionsByGameRow, error)) *Querier_GetRegionsByGame_Call {
	_c.Call.Return(run)
	return _c
}

// IncreaseRegionTroops provides a mock function with given fields: ctx, arg
func (_m *Querier) IncreaseRegionTroops(ctx context.Context, arg sqlc.IncreaseRegionTroopsParams) error {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for IncreaseRegionTroops")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.IncreaseRegionTroopsParams) error); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Querier_IncreaseRegionTroops_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IncreaseRegionTroops'
type Querier_IncreaseRegionTroops_Call struct {
	*mock.Call
}

// IncreaseRegionTroops is a helper method to define mock.On call
//   - ctx context.Context
//   - arg sqlc.IncreaseRegionTroopsParams
func (_e *Querier_Expecter) IncreaseRegionTroops(ctx interface{}, arg interface{}) *Querier_IncreaseRegionTroops_Call {
	return &Querier_IncreaseRegionTroops_Call{Call: _e.mock.On("IncreaseRegionTroops", ctx, arg)}
}

func (_c *Querier_IncreaseRegionTroops_Call) Run(run func(ctx context.Context, arg sqlc.IncreaseRegionTroopsParams)) *Querier_IncreaseRegionTroops_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(sqlc.IncreaseRegionTroopsParams))
	})
	return _c
}

func (_c *Querier_IncreaseRegionTroops_Call) Return(_a0 error) *Querier_IncreaseRegionTroops_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Querier_IncreaseRegionTroops_Call) RunAndReturn(run func(context.Context, sqlc.IncreaseRegionTroopsParams) error) *Querier_IncreaseRegionTroops_Call {
	_c.Call.Return(run)
	return _c
}

// InsertGame provides a mock function with given fields: ctx, deployableTroops
func (_m *Querier) InsertGame(ctx context.Context, deployableTroops int64) (sqlc.Game, error) {
	ret := _m.Called(ctx, deployableTroops)

	if len(ret) == 0 {
		panic("no return value specified for InsertGame")
	}

	var r0 sqlc.Game
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (sqlc.Game, error)); ok {
		return rf(ctx, deployableTroops)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) sqlc.Game); ok {
		r0 = rf(ctx, deployableTroops)
	} else {
		r0 = ret.Get(0).(sqlc.Game)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, deployableTroops)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_InsertGame_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InsertGame'
type Querier_InsertGame_Call struct {
	*mock.Call
}

// InsertGame is a helper method to define mock.On call
//   - ctx context.Context
//   - deployableTroops int64
func (_e *Querier_Expecter) InsertGame(ctx interface{}, deployableTroops interface{}) *Querier_InsertGame_Call {
	return &Querier_InsertGame_Call{Call: _e.mock.On("InsertGame", ctx, deployableTroops)}
}

func (_c *Querier_InsertGame_Call) Run(run func(ctx context.Context, deployableTroops int64)) *Querier_InsertGame_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *Querier_InsertGame_Call) Return(_a0 sqlc.Game, _a1 error) *Querier_InsertGame_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_InsertGame_Call) RunAndReturn(run func(context.Context, int64) (sqlc.Game, error)) *Querier_InsertGame_Call {
	_c.Call.Return(run)
	return _c
}

// InsertPlayers provides a mock function with given fields: ctx, arg
func (_m *Querier) InsertPlayers(ctx context.Context, arg []sqlc.InsertPlayersParams) (int64, error) {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for InsertPlayers")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []sqlc.InsertPlayersParams) (int64, error)); ok {
		return rf(ctx, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []sqlc.InsertPlayersParams) int64); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, []sqlc.InsertPlayersParams) error); ok {
		r1 = rf(ctx, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_InsertPlayers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InsertPlayers'
type Querier_InsertPlayers_Call struct {
	*mock.Call
}

// InsertPlayers is a helper method to define mock.On call
//   - ctx context.Context
//   - arg []sqlc.InsertPlayersParams
func (_e *Querier_Expecter) InsertPlayers(ctx interface{}, arg interface{}) *Querier_InsertPlayers_Call {
	return &Querier_InsertPlayers_Call{Call: _e.mock.On("InsertPlayers", ctx, arg)}
}

func (_c *Querier_InsertPlayers_Call) Run(run func(ctx context.Context, arg []sqlc.InsertPlayersParams)) *Querier_InsertPlayers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]sqlc.InsertPlayersParams))
	})
	return _c
}

func (_c *Querier_InsertPlayers_Call) Return(_a0 int64, _a1 error) *Querier_InsertPlayers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_InsertPlayers_Call) RunAndReturn(run func(context.Context, []sqlc.InsertPlayersParams) (int64, error)) *Querier_InsertPlayers_Call {
	_c.Call.Return(run)
	return _c
}

// InsertRegions provides a mock function with given fields: ctx, arg
func (_m *Querier) InsertRegions(ctx context.Context, arg []sqlc.InsertRegionsParams) (int64, error) {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for InsertRegions")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []sqlc.InsertRegionsParams) (int64, error)); ok {
		return rf(ctx, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []sqlc.InsertRegionsParams) int64); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, []sqlc.InsertRegionsParams) error); ok {
		r1 = rf(ctx, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_InsertRegions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InsertRegions'
type Querier_InsertRegions_Call struct {
	*mock.Call
}

// InsertRegions is a helper method to define mock.On call
//   - ctx context.Context
//   - arg []sqlc.InsertRegionsParams
func (_e *Querier_Expecter) InsertRegions(ctx interface{}, arg interface{}) *Querier_InsertRegions_Call {
	return &Querier_InsertRegions_Call{Call: _e.mock.On("InsertRegions", ctx, arg)}
}

func (_c *Querier_InsertRegions_Call) Run(run func(ctx context.Context, arg []sqlc.InsertRegionsParams)) *Querier_InsertRegions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]sqlc.InsertRegionsParams))
	})
	return _c
}

func (_c *Querier_InsertRegions_Call) Return(_a0 int64, _a1 error) *Querier_InsertRegions_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_InsertRegions_Call) RunAndReturn(run func(context.Context, []sqlc.InsertRegionsParams) (int64, error)) *Querier_InsertRegions_Call {
	_c.Call.Return(run)
	return _c
}

// SetGamePhase provides a mock function with given fields: ctx, arg
func (_m *Querier) SetGamePhase(ctx context.Context, arg sqlc.SetGamePhaseParams) error {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for SetGamePhase")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.SetGamePhaseParams) error); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Querier_SetGamePhase_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetGamePhase'
type Querier_SetGamePhase_Call struct {
	*mock.Call
}

// SetGamePhase is a helper method to define mock.On call
//   - ctx context.Context
//   - arg sqlc.SetGamePhaseParams
func (_e *Querier_Expecter) SetGamePhase(ctx interface{}, arg interface{}) *Querier_SetGamePhase_Call {
	return &Querier_SetGamePhase_Call{Call: _e.mock.On("SetGamePhase", ctx, arg)}
}

func (_c *Querier_SetGamePhase_Call) Run(run func(ctx context.Context, arg sqlc.SetGamePhaseParams)) *Querier_SetGamePhase_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(sqlc.SetGamePhaseParams))
	})
	return _c
}

func (_c *Querier_SetGamePhase_Call) Return(_a0 error) *Querier_SetGamePhase_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Querier_SetGamePhase_Call) RunAndReturn(run func(context.Context, sqlc.SetGamePhaseParams) error) *Querier_SetGamePhase_Call {
	_c.Call.Return(run)
	return _c
}

// NewQuerier creates a new instance of Querier. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewQuerier(t interface {
	mock.TestingT
	Cleanup(func())
}) *Querier {
	mock := &Querier{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
