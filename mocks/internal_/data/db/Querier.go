// Code generated by mockery v2.43.1. DO NOT EDIT.

package db

import (
	context "context"

	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/db"

	mock "github.com/stretchr/testify/mock"

	pgx "github.com/jackc/pgx/v5"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
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

// DecreaseDeployableTroops provides a mock function with given fields: _a0, arg
func (_m *Querier) DecreaseDeployableTroops(_a0 context.Context, arg sqlc.DecreaseDeployableTroopsParams) error {
	ret := _m.Called(_a0, arg)

	if len(ret) == 0 {
		panic("no return value specified for DecreaseDeployableTroops")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.DecreaseDeployableTroopsParams) error); ok {
		r0 = rf(_a0, arg)
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
//   - _a0 context.Context
//   - arg sqlc.DecreaseDeployableTroopsParams
func (_e *Querier_Expecter) DecreaseDeployableTroops(_a0 interface{}, arg interface{}) *Querier_DecreaseDeployableTroops_Call {
	return &Querier_DecreaseDeployableTroops_Call{Call: _e.mock.On("DecreaseDeployableTroops", _a0, arg)}
}

func (_c *Querier_DecreaseDeployableTroops_Call) Run(run func(_a0 context.Context, arg sqlc.DecreaseDeployableTroopsParams)) *Querier_DecreaseDeployableTroops_Call {
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

// ExecuteInTransaction provides a mock function with given fields: _a0, txFunc
func (_m *Querier) ExecuteInTransaction(_a0 ctx.LogContext, txFunc func(db.Querier) (interface{}, error)) (interface{}, error) {
	ret := _m.Called(_a0, txFunc)

	if len(ret) == 0 {
		panic("no return value specified for ExecuteInTransaction")
	}

	var r0 interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.LogContext, func(db.Querier) (interface{}, error)) (interface{}, error)); ok {
		return rf(_a0, txFunc)
	}
	if rf, ok := ret.Get(0).(func(ctx.LogContext, func(db.Querier) (interface{}, error)) interface{}); ok {
		r0 = rf(_a0, txFunc)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.LogContext, func(db.Querier) (interface{}, error)) error); ok {
		r1 = rf(_a0, txFunc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_ExecuteInTransaction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExecuteInTransaction'
type Querier_ExecuteInTransaction_Call struct {
	*mock.Call
}

// ExecuteInTransaction is a helper method to define mock.On call
//   - _a0 ctx.LogContext
//   - txFunc func(db.Querier)(interface{} , error)
func (_e *Querier_Expecter) ExecuteInTransaction(_a0 interface{}, txFunc interface{}) *Querier_ExecuteInTransaction_Call {
	return &Querier_ExecuteInTransaction_Call{Call: _e.mock.On("ExecuteInTransaction", _a0, txFunc)}
}

func (_c *Querier_ExecuteInTransaction_Call) Run(run func(_a0 ctx.LogContext, txFunc func(db.Querier) (interface{}, error))) *Querier_ExecuteInTransaction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.LogContext), args[1].(func(db.Querier) (interface{}, error)))
	})
	return _c
}

func (_c *Querier_ExecuteInTransaction_Call) Return(_a0 interface{}, _a1 error) *Querier_ExecuteInTransaction_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_ExecuteInTransaction_Call) RunAndReturn(run func(ctx.LogContext, func(db.Querier) (interface{}, error)) (interface{}, error)) *Querier_ExecuteInTransaction_Call {
	_c.Call.Return(run)
	return _c
}

// ExecuteInTransactionWithIsolation provides a mock function with given fields: _a0, isolationLevel, txFunc
func (_m *Querier) ExecuteInTransactionWithIsolation(_a0 ctx.LogContext, isolationLevel pgx.TxIsoLevel, txFunc func(db.Querier) (interface{}, error)) (interface{}, error) {
	ret := _m.Called(_a0, isolationLevel, txFunc)

	if len(ret) == 0 {
		panic("no return value specified for ExecuteInTransactionWithIsolation")
	}

	var r0 interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.LogContext, pgx.TxIsoLevel, func(db.Querier) (interface{}, error)) (interface{}, error)); ok {
		return rf(_a0, isolationLevel, txFunc)
	}
	if rf, ok := ret.Get(0).(func(ctx.LogContext, pgx.TxIsoLevel, func(db.Querier) (interface{}, error)) interface{}); ok {
		r0 = rf(_a0, isolationLevel, txFunc)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.LogContext, pgx.TxIsoLevel, func(db.Querier) (interface{}, error)) error); ok {
		r1 = rf(_a0, isolationLevel, txFunc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_ExecuteInTransactionWithIsolation_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExecuteInTransactionWithIsolation'
type Querier_ExecuteInTransactionWithIsolation_Call struct {
	*mock.Call
}

// ExecuteInTransactionWithIsolation is a helper method to define mock.On call
//   - _a0 ctx.LogContext
//   - isolationLevel pgx.TxIsoLevel
//   - txFunc func(db.Querier)(interface{} , error)
func (_e *Querier_Expecter) ExecuteInTransactionWithIsolation(_a0 interface{}, isolationLevel interface{}, txFunc interface{}) *Querier_ExecuteInTransactionWithIsolation_Call {
	return &Querier_ExecuteInTransactionWithIsolation_Call{Call: _e.mock.On("ExecuteInTransactionWithIsolation", _a0, isolationLevel, txFunc)}
}

func (_c *Querier_ExecuteInTransactionWithIsolation_Call) Run(run func(_a0 ctx.LogContext, isolationLevel pgx.TxIsoLevel, txFunc func(db.Querier) (interface{}, error))) *Querier_ExecuteInTransactionWithIsolation_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.LogContext), args[1].(pgx.TxIsoLevel), args[2].(func(db.Querier) (interface{}, error)))
	})
	return _c
}

func (_c *Querier_ExecuteInTransactionWithIsolation_Call) Return(_a0 interface{}, _a1 error) *Querier_ExecuteInTransactionWithIsolation_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_ExecuteInTransactionWithIsolation_Call) RunAndReturn(run func(ctx.LogContext, pgx.TxIsoLevel, func(db.Querier) (interface{}, error)) (interface{}, error)) *Querier_ExecuteInTransactionWithIsolation_Call {
	_c.Call.Return(run)
	return _c
}

// GetDeployableTroops provides a mock function with given fields: _a0, id
func (_m *Querier) GetDeployableTroops(_a0 context.Context, id int64) (int64, error) {
	ret := _m.Called(_a0, id)

	if len(ret) == 0 {
		panic("no return value specified for GetDeployableTroops")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (int64, error)); ok {
		return rf(_a0, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) int64); ok {
		r0 = rf(_a0, id)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(_a0, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_GetDeployableTroops_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDeployableTroops'
type Querier_GetDeployableTroops_Call struct {
	*mock.Call
}

// GetDeployableTroops is a helper method to define mock.On call
//   - _a0 context.Context
//   - id int64
func (_e *Querier_Expecter) GetDeployableTroops(_a0 interface{}, id interface{}) *Querier_GetDeployableTroops_Call {
	return &Querier_GetDeployableTroops_Call{Call: _e.mock.On("GetDeployableTroops", _a0, id)}
}

func (_c *Querier_GetDeployableTroops_Call) Run(run func(_a0 context.Context, id int64)) *Querier_GetDeployableTroops_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *Querier_GetDeployableTroops_Call) Return(_a0 int64, _a1 error) *Querier_GetDeployableTroops_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_GetDeployableTroops_Call) RunAndReturn(run func(context.Context, int64) (int64, error)) *Querier_GetDeployableTroops_Call {
	_c.Call.Return(run)
	return _c
}

// GetGame provides a mock function with given fields: _a0, id
func (_m *Querier) GetGame(_a0 context.Context, id int64) (sqlc.GetGameRow, error) {
	ret := _m.Called(_a0, id)

	if len(ret) == 0 {
		panic("no return value specified for GetGame")
	}

	var r0 sqlc.GetGameRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (sqlc.GetGameRow, error)); ok {
		return rf(_a0, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) sqlc.GetGameRow); ok {
		r0 = rf(_a0, id)
	} else {
		r0 = ret.Get(0).(sqlc.GetGameRow)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(_a0, id)
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
//   - _a0 context.Context
//   - id int64
func (_e *Querier_Expecter) GetGame(_a0 interface{}, id interface{}) *Querier_GetGame_Call {
	return &Querier_GetGame_Call{Call: _e.mock.On("GetGame", _a0, id)}
}

func (_c *Querier_GetGame_Call) Run(run func(_a0 context.Context, id int64)) *Querier_GetGame_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *Querier_GetGame_Call) Return(_a0 sqlc.GetGameRow, _a1 error) *Querier_GetGame_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_GetGame_Call) RunAndReturn(run func(context.Context, int64) (sqlc.GetGameRow, error)) *Querier_GetGame_Call {
	_c.Call.Return(run)
	return _c
}

// GetPlayerByUserId provides a mock function with given fields: _a0, userID
func (_m *Querier) GetPlayerByUserId(_a0 context.Context, userID string) (sqlc.Player, error) {
	ret := _m.Called(_a0, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetPlayerByUserId")
	}

	var r0 sqlc.Player
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (sqlc.Player, error)); ok {
		return rf(_a0, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) sqlc.Player); ok {
		r0 = rf(_a0, userID)
	} else {
		r0 = ret.Get(0).(sqlc.Player)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, userID)
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
//   - _a0 context.Context
//   - userID string
func (_e *Querier_Expecter) GetPlayerByUserId(_a0 interface{}, userID interface{}) *Querier_GetPlayerByUserId_Call {
	return &Querier_GetPlayerByUserId_Call{Call: _e.mock.On("GetPlayerByUserId", _a0, userID)}
}

func (_c *Querier_GetPlayerByUserId_Call) Run(run func(_a0 context.Context, userID string)) *Querier_GetPlayerByUserId_Call {
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

// GetPlayersByGame provides a mock function with given fields: _a0, gameID
func (_m *Querier) GetPlayersByGame(_a0 context.Context, gameID int64) ([]sqlc.Player, error) {
	ret := _m.Called(_a0, gameID)

	if len(ret) == 0 {
		panic("no return value specified for GetPlayersByGame")
	}

	var r0 []sqlc.Player
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) ([]sqlc.Player, error)); ok {
		return rf(_a0, gameID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) []sqlc.Player); ok {
		r0 = rf(_a0, gameID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.Player)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(_a0, gameID)
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
//   - _a0 context.Context
//   - gameID int64
func (_e *Querier_Expecter) GetPlayersByGame(_a0 interface{}, gameID interface{}) *Querier_GetPlayersByGame_Call {
	return &Querier_GetPlayersByGame_Call{Call: _e.mock.On("GetPlayersByGame", _a0, gameID)}
}

func (_c *Querier_GetPlayersByGame_Call) Run(run func(_a0 context.Context, gameID int64)) *Querier_GetPlayersByGame_Call {
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

// GetRegionsByGame provides a mock function with given fields: _a0, id
func (_m *Querier) GetRegionsByGame(_a0 context.Context, id int64) ([]sqlc.GetRegionsByGameRow, error) {
	ret := _m.Called(_a0, id)

	if len(ret) == 0 {
		panic("no return value specified for GetRegionsByGame")
	}

	var r0 []sqlc.GetRegionsByGameRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) ([]sqlc.GetRegionsByGameRow, error)); ok {
		return rf(_a0, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) []sqlc.GetRegionsByGameRow); ok {
		r0 = rf(_a0, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GetRegionsByGameRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(_a0, id)
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
//   - _a0 context.Context
//   - id int64
func (_e *Querier_Expecter) GetRegionsByGame(_a0 interface{}, id interface{}) *Querier_GetRegionsByGame_Call {
	return &Querier_GetRegionsByGame_Call{Call: _e.mock.On("GetRegionsByGame", _a0, id)}
}

func (_c *Querier_GetRegionsByGame_Call) Run(run func(_a0 context.Context, id int64)) *Querier_GetRegionsByGame_Call {
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

// IncreaseRegionTroops provides a mock function with given fields: _a0, arg
func (_m *Querier) IncreaseRegionTroops(_a0 context.Context, arg sqlc.IncreaseRegionTroopsParams) error {
	ret := _m.Called(_a0, arg)

	if len(ret) == 0 {
		panic("no return value specified for IncreaseRegionTroops")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.IncreaseRegionTroopsParams) error); ok {
		r0 = rf(_a0, arg)
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
//   - _a0 context.Context
//   - arg sqlc.IncreaseRegionTroopsParams
func (_e *Querier_Expecter) IncreaseRegionTroops(_a0 interface{}, arg interface{}) *Querier_IncreaseRegionTroops_Call {
	return &Querier_IncreaseRegionTroops_Call{Call: _e.mock.On("IncreaseRegionTroops", _a0, arg)}
}

func (_c *Querier_IncreaseRegionTroops_Call) Run(run func(_a0 context.Context, arg sqlc.IncreaseRegionTroopsParams)) *Querier_IncreaseRegionTroops_Call {
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

// InsertDeployPhase provides a mock function with given fields: _a0, arg
func (_m *Querier) InsertDeployPhase(_a0 context.Context, arg sqlc.InsertDeployPhaseParams) (sqlc.DeployPhase, error) {
	ret := _m.Called(_a0, arg)

	if len(ret) == 0 {
		panic("no return value specified for InsertDeployPhase")
	}

	var r0 sqlc.DeployPhase
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.InsertDeployPhaseParams) (sqlc.DeployPhase, error)); ok {
		return rf(_a0, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.InsertDeployPhaseParams) sqlc.DeployPhase); ok {
		r0 = rf(_a0, arg)
	} else {
		r0 = ret.Get(0).(sqlc.DeployPhase)
	}

	if rf, ok := ret.Get(1).(func(context.Context, sqlc.InsertDeployPhaseParams) error); ok {
		r1 = rf(_a0, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_InsertDeployPhase_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InsertDeployPhase'
type Querier_InsertDeployPhase_Call struct {
	*mock.Call
}

// InsertDeployPhase is a helper method to define mock.On call
//   - _a0 context.Context
//   - arg sqlc.InsertDeployPhaseParams
func (_e *Querier_Expecter) InsertDeployPhase(_a0 interface{}, arg interface{}) *Querier_InsertDeployPhase_Call {
	return &Querier_InsertDeployPhase_Call{Call: _e.mock.On("InsertDeployPhase", _a0, arg)}
}

func (_c *Querier_InsertDeployPhase_Call) Run(run func(_a0 context.Context, arg sqlc.InsertDeployPhaseParams)) *Querier_InsertDeployPhase_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(sqlc.InsertDeployPhaseParams))
	})
	return _c
}

func (_c *Querier_InsertDeployPhase_Call) Return(_a0 sqlc.DeployPhase, _a1 error) *Querier_InsertDeployPhase_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_InsertDeployPhase_Call) RunAndReturn(run func(context.Context, sqlc.InsertDeployPhaseParams) (sqlc.DeployPhase, error)) *Querier_InsertDeployPhase_Call {
	_c.Call.Return(run)
	return _c
}

// InsertGame provides a mock function with given fields: _a0
func (_m *Querier) InsertGame(_a0 context.Context) (sqlc.Game, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for InsertGame")
	}

	var r0 sqlc.Game
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (sqlc.Game, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) sqlc.Game); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(sqlc.Game)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
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
//   - _a0 context.Context
func (_e *Querier_Expecter) InsertGame(_a0 interface{}) *Querier_InsertGame_Call {
	return &Querier_InsertGame_Call{Call: _e.mock.On("InsertGame", _a0)}
}

func (_c *Querier_InsertGame_Call) Run(run func(_a0 context.Context)) *Querier_InsertGame_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Querier_InsertGame_Call) Return(_a0 sqlc.Game, _a1 error) *Querier_InsertGame_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_InsertGame_Call) RunAndReturn(run func(context.Context) (sqlc.Game, error)) *Querier_InsertGame_Call {
	_c.Call.Return(run)
	return _c
}

// InsertPhase provides a mock function with given fields: _a0, arg
func (_m *Querier) InsertPhase(_a0 context.Context, arg sqlc.InsertPhaseParams) (sqlc.Phase, error) {
	ret := _m.Called(_a0, arg)

	if len(ret) == 0 {
		panic("no return value specified for InsertPhase")
	}

	var r0 sqlc.Phase
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.InsertPhaseParams) (sqlc.Phase, error)); ok {
		return rf(_a0, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.InsertPhaseParams) sqlc.Phase); ok {
		r0 = rf(_a0, arg)
	} else {
		r0 = ret.Get(0).(sqlc.Phase)
	}

	if rf, ok := ret.Get(1).(func(context.Context, sqlc.InsertPhaseParams) error); ok {
		r1 = rf(_a0, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_InsertPhase_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InsertPhase'
type Querier_InsertPhase_Call struct {
	*mock.Call
}

// InsertPhase is a helper method to define mock.On call
//   - _a0 context.Context
//   - arg sqlc.InsertPhaseParams
func (_e *Querier_Expecter) InsertPhase(_a0 interface{}, arg interface{}) *Querier_InsertPhase_Call {
	return &Querier_InsertPhase_Call{Call: _e.mock.On("InsertPhase", _a0, arg)}
}

func (_c *Querier_InsertPhase_Call) Run(run func(_a0 context.Context, arg sqlc.InsertPhaseParams)) *Querier_InsertPhase_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(sqlc.InsertPhaseParams))
	})
	return _c
}

func (_c *Querier_InsertPhase_Call) Return(_a0 sqlc.Phase, _a1 error) *Querier_InsertPhase_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_InsertPhase_Call) RunAndReturn(run func(context.Context, sqlc.InsertPhaseParams) (sqlc.Phase, error)) *Querier_InsertPhase_Call {
	_c.Call.Return(run)
	return _c
}

// InsertPlayers provides a mock function with given fields: _a0, arg
func (_m *Querier) InsertPlayers(_a0 context.Context, arg []sqlc.InsertPlayersParams) (int64, error) {
	ret := _m.Called(_a0, arg)

	if len(ret) == 0 {
		panic("no return value specified for InsertPlayers")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []sqlc.InsertPlayersParams) (int64, error)); ok {
		return rf(_a0, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []sqlc.InsertPlayersParams) int64); ok {
		r0 = rf(_a0, arg)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, []sqlc.InsertPlayersParams) error); ok {
		r1 = rf(_a0, arg)
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
//   - _a0 context.Context
//   - arg []sqlc.InsertPlayersParams
func (_e *Querier_Expecter) InsertPlayers(_a0 interface{}, arg interface{}) *Querier_InsertPlayers_Call {
	return &Querier_InsertPlayers_Call{Call: _e.mock.On("InsertPlayers", _a0, arg)}
}

func (_c *Querier_InsertPlayers_Call) Run(run func(_a0 context.Context, arg []sqlc.InsertPlayersParams)) *Querier_InsertPlayers_Call {
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

// InsertRegions provides a mock function with given fields: _a0, arg
func (_m *Querier) InsertRegions(_a0 context.Context, arg []sqlc.InsertRegionsParams) (int64, error) {
	ret := _m.Called(_a0, arg)

	if len(ret) == 0 {
		panic("no return value specified for InsertRegions")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []sqlc.InsertRegionsParams) (int64, error)); ok {
		return rf(_a0, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []sqlc.InsertRegionsParams) int64); ok {
		r0 = rf(_a0, arg)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, []sqlc.InsertRegionsParams) error); ok {
		r1 = rf(_a0, arg)
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
//   - _a0 context.Context
//   - arg []sqlc.InsertRegionsParams
func (_e *Querier_Expecter) InsertRegions(_a0 interface{}, arg interface{}) *Querier_InsertRegions_Call {
	return &Querier_InsertRegions_Call{Call: _e.mock.On("InsertRegions", _a0, arg)}
}

func (_c *Querier_InsertRegions_Call) Run(run func(_a0 context.Context, arg []sqlc.InsertRegionsParams)) *Querier_InsertRegions_Call {
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

// SetGamePhase provides a mock function with given fields: _a0, arg
func (_m *Querier) SetGamePhase(_a0 context.Context, arg sqlc.SetGamePhaseParams) error {
	ret := _m.Called(_a0, arg)

	if len(ret) == 0 {
		panic("no return value specified for SetGamePhase")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.SetGamePhaseParams) error); ok {
		r0 = rf(_a0, arg)
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
//   - _a0 context.Context
//   - arg sqlc.SetGamePhaseParams
func (_e *Querier_Expecter) SetGamePhase(_a0 interface{}, arg interface{}) *Querier_SetGamePhase_Call {
	return &Querier_SetGamePhase_Call{Call: _e.mock.On("SetGamePhase", _a0, arg)}
}

func (_c *Querier_SetGamePhase_Call) Run(run func(_a0 context.Context, arg sqlc.SetGamePhaseParams)) *Querier_SetGamePhase_Call {
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
