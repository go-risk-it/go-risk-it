// Code generated by mockery v2.40.1. DO NOT EDIT.

package game

import (
	context "context"

	board "github.com/tomfran/go-risk-it/internal/logic/board"

	db "github.com/tomfran/go-risk-it/internal/db"

	mock "github.com/stretchr/testify/mock"
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

// CreateGame provides a mock function with given fields: ctx, q, _a2, users
func (_m *Service) CreateGame(ctx context.Context, q db.Querier, _a2 *board.Board, users []string) error {
	ret := _m.Called(ctx, q, _a2, users)

	if len(ret) == 0 {
		panic("no return value specified for CreateGame")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, db.Querier, *board.Board, []string) error); ok {
		r0 = rf(ctx, q, _a2, users)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_CreateGame_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateGame'
type Service_CreateGame_Call struct {
	*mock.Call
}

// CreateGame is a helper method to define mock.On call
//   - ctx context.Context
//   - q db.Querier
//   - _a2 *board.Board
//   - users []string
func (_e *Service_Expecter) CreateGame(ctx interface{}, q interface{}, _a2 interface{}, users interface{}) *Service_CreateGame_Call {
	return &Service_CreateGame_Call{Call: _e.mock.On("CreateGame", ctx, q, _a2, users)}
}

func (_c *Service_CreateGame_Call) Run(run func(ctx context.Context, q db.Querier, _a2 *board.Board, users []string)) *Service_CreateGame_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(db.Querier), args[2].(*board.Board), args[3].([]string))
	})
	return _c
}

func (_c *Service_CreateGame_Call) Return(_a0 error) *Service_CreateGame_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_CreateGame_Call) RunAndReturn(run func(context.Context, db.Querier, *board.Board, []string) error) *Service_CreateGame_Call {
	_c.Call.Return(run)
	return _c
}

// GetGameState provides a mock function with given fields: ctx, gameID
func (_m *Service) GetGameState(ctx context.Context, gameID int64) (*db.Game, error) {
	ret := _m.Called(ctx, gameID)

	if len(ret) == 0 {
		panic("no return value specified for GetGameState")
	}

	var r0 *db.Game
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (*db.Game, error)); ok {
		return rf(ctx, gameID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) *db.Game); ok {
		r0 = rf(ctx, gameID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*db.Game)
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

func (_c *Service_GetGameState_Call) Return(_a0 *db.Game, _a1 error) *Service_GetGameState_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_GetGameState_Call) RunAndReturn(run func(context.Context, int64) (*db.Game, error)) *Service_GetGameState_Call {
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