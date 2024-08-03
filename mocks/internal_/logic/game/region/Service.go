// Code generated by mockery v2.44.1. DO NOT EDIT.

package region

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
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

// CreateRegions provides a mock function with given fields: _a0, querier, players, regions
func (_m *Service) CreateRegions(_a0 ctx.UserContext, querier db.Querier, players []sqlc.Player, regions []string) error {
	ret := _m.Called(_a0, querier, players, regions)

	if len(ret) == 0 {
		panic("no return value specified for CreateRegions")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.UserContext, db.Querier, []sqlc.Player, []string) error); ok {
		r0 = rf(_a0, querier, players, regions)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_CreateRegions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateRegions'
type Service_CreateRegions_Call struct {
	*mock.Call
}

// CreateRegions is a helper method to define mock.On call
//   - _a0 ctx.UserContext
//   - querier db.Querier
//   - players []sqlc.Player
//   - regions []string
func (_e *Service_Expecter) CreateRegions(_a0 interface{}, querier interface{}, players interface{}, regions interface{}) *Service_CreateRegions_Call {
	return &Service_CreateRegions_Call{Call: _e.mock.On("CreateRegions", _a0, querier, players, regions)}
}

func (_c *Service_CreateRegions_Call) Run(run func(_a0 ctx.UserContext, querier db.Querier, players []sqlc.Player, regions []string)) *Service_CreateRegions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.UserContext), args[1].(db.Querier), args[2].([]sqlc.Player), args[3].([]string))
	})
	return _c
}

func (_c *Service_CreateRegions_Call) Return(_a0 error) *Service_CreateRegions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_CreateRegions_Call) RunAndReturn(run func(ctx.UserContext, db.Querier, []sqlc.Player, []string) error) *Service_CreateRegions_Call {
	_c.Call.Return(run)
	return _c
}

// GetRegionQ provides a mock function with given fields: _a0, querier, _a2
func (_m *Service) GetRegionQ(_a0 ctx.MoveContext, querier db.Querier, _a2 string) (*sqlc.GetRegionsByGameRow, error) {
	ret := _m.Called(_a0, querier, _a2)

	if len(ret) == 0 {
		panic("no return value specified for GetRegionQ")
	}

	var r0 *sqlc.GetRegionsByGameRow
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, string) (*sqlc.GetRegionsByGameRow, error)); ok {
		return rf(_a0, querier, _a2)
	}
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, string) *sqlc.GetRegionsByGameRow); ok {
		r0 = rf(_a0, querier, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sqlc.GetRegionsByGameRow)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.MoveContext, db.Querier, string) error); ok {
		r1 = rf(_a0, querier, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetRegionQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRegionQ'
type Service_GetRegionQ_Call struct {
	*mock.Call
}

// GetRegionQ is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - querier db.Querier
//   - _a2 string
func (_e *Service_Expecter) GetRegionQ(_a0 interface{}, querier interface{}, _a2 interface{}) *Service_GetRegionQ_Call {
	return &Service_GetRegionQ_Call{Call: _e.mock.On("GetRegionQ", _a0, querier, _a2)}
}

func (_c *Service_GetRegionQ_Call) Run(run func(_a0 ctx.MoveContext, querier db.Querier, _a2 string)) *Service_GetRegionQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier), args[2].(string))
	})
	return _c
}

func (_c *Service_GetRegionQ_Call) Return(_a0 *sqlc.GetRegionsByGameRow, _a1 error) *Service_GetRegionQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetRegionQ_Call) RunAndReturn(run func(ctx.MoveContext, db.Querier, string) (*sqlc.GetRegionsByGameRow, error)) *Service_GetRegionQ_Call {
	_c.Call.Return(run)
	return _c
}

// GetRegions provides a mock function with given fields: _a0
func (_m *Service) GetRegions(_a0 ctx.GameContext) ([]sqlc.GetRegionsByGameRow, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetRegions")
	}

	var r0 []sqlc.GetRegionsByGameRow
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext) ([]sqlc.GetRegionsByGameRow, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext) []sqlc.GetRegionsByGameRow); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GetRegionsByGameRow)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetRegions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRegions'
type Service_GetRegions_Call struct {
	*mock.Call
}

// GetRegions is a helper method to define mock.On call
//   - _a0 ctx.GameContext
func (_e *Service_Expecter) GetRegions(_a0 interface{}) *Service_GetRegions_Call {
	return &Service_GetRegions_Call{Call: _e.mock.On("GetRegions", _a0)}
}

func (_c *Service_GetRegions_Call) Run(run func(_a0 ctx.GameContext)) *Service_GetRegions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext))
	})
	return _c
}

func (_c *Service_GetRegions_Call) Return(_a0 []sqlc.GetRegionsByGameRow, _a1 error) *Service_GetRegions_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetRegions_Call) RunAndReturn(run func(ctx.GameContext) ([]sqlc.GetRegionsByGameRow, error)) *Service_GetRegions_Call {
	_c.Call.Return(run)
	return _c
}

// GetRegionsQ provides a mock function with given fields: _a0, querier
func (_m *Service) GetRegionsQ(_a0 ctx.GameContext, querier db.Querier) ([]sqlc.GetRegionsByGameRow, error) {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for GetRegionsQ")
	}

	var r0 []sqlc.GetRegionsByGameRow
	var r1 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) ([]sqlc.GetRegionsByGameRow, error)); ok {
		return rf(_a0, querier)
	}
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) []sqlc.GetRegionsByGameRow); ok {
		r0 = rf(_a0, querier)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GetRegionsByGameRow)
		}
	}

	if rf, ok := ret.Get(1).(func(ctx.GameContext, db.Querier) error); ok {
		r1 = rf(_a0, querier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_GetRegionsQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRegionsQ'
type Service_GetRegionsQ_Call struct {
	*mock.Call
}

// GetRegionsQ is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter) GetRegionsQ(_a0 interface{}, querier interface{}) *Service_GetRegionsQ_Call {
	return &Service_GetRegionsQ_Call{Call: _e.mock.On("GetRegionsQ", _a0, querier)}
}

func (_c *Service_GetRegionsQ_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_GetRegionsQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_GetRegionsQ_Call) Return(_a0 []sqlc.GetRegionsByGameRow, _a1 error) *Service_GetRegionsQ_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetRegionsQ_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) ([]sqlc.GetRegionsByGameRow, error)) *Service_GetRegionsQ_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateTroopsInRegion provides a mock function with given fields: _a0, querier, regionID, troopsToAdd
func (_m *Service) UpdateTroopsInRegion(_a0 ctx.MoveContext, querier db.Querier, regionID int64, troopsToAdd int64) error {
	ret := _m.Called(_a0, querier, regionID, troopsToAdd)

	if len(ret) == 0 {
		panic("no return value specified for UpdateTroopsInRegion")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, int64, int64) error); ok {
		r0 = rf(_a0, querier, regionID, troopsToAdd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_UpdateTroopsInRegion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateTroopsInRegion'
type Service_UpdateTroopsInRegion_Call struct {
	*mock.Call
}

// UpdateTroopsInRegion is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - querier db.Querier
//   - regionID int64
//   - troopsToAdd int64
func (_e *Service_Expecter) UpdateTroopsInRegion(_a0 interface{}, querier interface{}, regionID interface{}, troopsToAdd interface{}) *Service_UpdateTroopsInRegion_Call {
	return &Service_UpdateTroopsInRegion_Call{Call: _e.mock.On("UpdateTroopsInRegion", _a0, querier, regionID, troopsToAdd)}
}

func (_c *Service_UpdateTroopsInRegion_Call) Run(run func(_a0 ctx.MoveContext, querier db.Querier, regionID int64, troopsToAdd int64)) *Service_UpdateTroopsInRegion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier), args[2].(int64), args[3].(int64))
	})
	return _c
}

func (_c *Service_UpdateTroopsInRegion_Call) Return(_a0 error) *Service_UpdateTroopsInRegion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_UpdateTroopsInRegion_Call) RunAndReturn(run func(ctx.MoveContext, db.Querier, int64, int64) error) *Service_UpdateTroopsInRegion_Call {
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
