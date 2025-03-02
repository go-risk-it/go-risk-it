// Code generated by mockery v2.50.1. DO NOT EDIT.

package sqlc

import (
	context "context"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
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

// CanLobbyBeStarted provides a mock function with given fields: ctx, arg
func (_m *Querier) CanLobbyBeStarted(ctx context.Context, arg sqlc.CanLobbyBeStartedParams) (bool, error) {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for CanLobbyBeStarted")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.CanLobbyBeStartedParams) (bool, error)); ok {
		return rf(ctx, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.CanLobbyBeStartedParams) bool); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, sqlc.CanLobbyBeStartedParams) error); ok {
		r1 = rf(ctx, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_CanLobbyBeStarted_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CanLobbyBeStarted'
type Querier_CanLobbyBeStarted_Call struct {
	*mock.Call
}

// CanLobbyBeStarted is a helper method to define mock.On call
//   - ctx context.Context
//   - arg sqlc.CanLobbyBeStartedParams
func (_e *Querier_Expecter) CanLobbyBeStarted(ctx interface{}, arg interface{}) *Querier_CanLobbyBeStarted_Call {
	return &Querier_CanLobbyBeStarted_Call{Call: _e.mock.On("CanLobbyBeStarted", ctx, arg)}
}

func (_c *Querier_CanLobbyBeStarted_Call) Run(run func(ctx context.Context, arg sqlc.CanLobbyBeStartedParams)) *Querier_CanLobbyBeStarted_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(sqlc.CanLobbyBeStartedParams))
	})
	return _c
}

func (_c *Querier_CanLobbyBeStarted_Call) Return(_a0 bool, _a1 error) *Querier_CanLobbyBeStarted_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_CanLobbyBeStarted_Call) RunAndReturn(run func(context.Context, sqlc.CanLobbyBeStartedParams) (bool, error)) *Querier_CanLobbyBeStarted_Call {
	_c.Call.Return(run)
	return _c
}

// CreateLobby provides a mock function with given fields: ctx
func (_m *Querier) CreateLobby(ctx context.Context) (int64, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for CreateLobby")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (int64, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) int64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_CreateLobby_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateLobby'
type Querier_CreateLobby_Call struct {
	*mock.Call
}

// CreateLobby is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Querier_Expecter) CreateLobby(ctx interface{}) *Querier_CreateLobby_Call {
	return &Querier_CreateLobby_Call{Call: _e.mock.On("CreateLobby", ctx)}
}

func (_c *Querier_CreateLobby_Call) Run(run func(ctx context.Context)) *Querier_CreateLobby_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Querier_CreateLobby_Call) Return(_a0 int64, _a1 error) *Querier_CreateLobby_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_CreateLobby_Call) RunAndReturn(run func(context.Context) (int64, error)) *Querier_CreateLobby_Call {
	_c.Call.Return(run)
	return _c
}

// GetJoinableLobbies provides a mock function with given fields: ctx, userID
func (_m *Querier) GetJoinableLobbies(ctx context.Context, userID string) ([]sqlc.GetJoinableLobbiesRow, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetJoinableLobbies")
	}

	var r0 []sqlc.GetJoinableLobbiesRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]sqlc.GetJoinableLobbiesRow, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []sqlc.GetJoinableLobbiesRow); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GetJoinableLobbiesRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_GetJoinableLobbies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetJoinableLobbies'
type Querier_GetJoinableLobbies_Call struct {
	*mock.Call
}

// GetJoinableLobbies is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
func (_e *Querier_Expecter) GetJoinableLobbies(ctx interface{}, userID interface{}) *Querier_GetJoinableLobbies_Call {
	return &Querier_GetJoinableLobbies_Call{Call: _e.mock.On("GetJoinableLobbies", ctx, userID)}
}

func (_c *Querier_GetJoinableLobbies_Call) Run(run func(ctx context.Context, userID string)) *Querier_GetJoinableLobbies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Querier_GetJoinableLobbies_Call) Return(_a0 []sqlc.GetJoinableLobbiesRow, _a1 error) *Querier_GetJoinableLobbies_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_GetJoinableLobbies_Call) RunAndReturn(run func(context.Context, string) ([]sqlc.GetJoinableLobbiesRow, error)) *Querier_GetJoinableLobbies_Call {
	_c.Call.Return(run)
	return _c
}

// GetJoinedLobbies provides a mock function with given fields: ctx, userID
func (_m *Querier) GetJoinedLobbies(ctx context.Context, userID string) ([]sqlc.GetJoinedLobbiesRow, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetJoinedLobbies")
	}

	var r0 []sqlc.GetJoinedLobbiesRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]sqlc.GetJoinedLobbiesRow, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []sqlc.GetJoinedLobbiesRow); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GetJoinedLobbiesRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_GetJoinedLobbies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetJoinedLobbies'
type Querier_GetJoinedLobbies_Call struct {
	*mock.Call
}

// GetJoinedLobbies is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
func (_e *Querier_Expecter) GetJoinedLobbies(ctx interface{}, userID interface{}) *Querier_GetJoinedLobbies_Call {
	return &Querier_GetJoinedLobbies_Call{Call: _e.mock.On("GetJoinedLobbies", ctx, userID)}
}

func (_c *Querier_GetJoinedLobbies_Call) Run(run func(ctx context.Context, userID string)) *Querier_GetJoinedLobbies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Querier_GetJoinedLobbies_Call) Return(_a0 []sqlc.GetJoinedLobbiesRow, _a1 error) *Querier_GetJoinedLobbies_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_GetJoinedLobbies_Call) RunAndReturn(run func(context.Context, string) ([]sqlc.GetJoinedLobbiesRow, error)) *Querier_GetJoinedLobbies_Call {
	_c.Call.Return(run)
	return _c
}

// GetLobby provides a mock function with given fields: ctx, id
func (_m *Querier) GetLobby(ctx context.Context, id int64) ([]sqlc.GetLobbyRow, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetLobby")
	}

	var r0 []sqlc.GetLobbyRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) ([]sqlc.GetLobbyRow, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) []sqlc.GetLobbyRow); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GetLobbyRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_GetLobby_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLobby'
type Querier_GetLobby_Call struct {
	*mock.Call
}

// GetLobby is a helper method to define mock.On call
//   - ctx context.Context
//   - id int64
func (_e *Querier_Expecter) GetLobby(ctx interface{}, id interface{}) *Querier_GetLobby_Call {
	return &Querier_GetLobby_Call{Call: _e.mock.On("GetLobby", ctx, id)}
}

func (_c *Querier_GetLobby_Call) Run(run func(ctx context.Context, id int64)) *Querier_GetLobby_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *Querier_GetLobby_Call) Return(_a0 []sqlc.GetLobbyRow, _a1 error) *Querier_GetLobby_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_GetLobby_Call) RunAndReturn(run func(context.Context, int64) ([]sqlc.GetLobbyRow, error)) *Querier_GetLobby_Call {
	_c.Call.Return(run)
	return _c
}

// GetLobbyPlayers provides a mock function with given fields: ctx, id
func (_m *Querier) GetLobbyPlayers(ctx context.Context, id int64) ([]sqlc.GetLobbyPlayersRow, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetLobbyPlayers")
	}

	var r0 []sqlc.GetLobbyPlayersRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) ([]sqlc.GetLobbyPlayersRow, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) []sqlc.GetLobbyPlayersRow); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GetLobbyPlayersRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_GetLobbyPlayers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLobbyPlayers'
type Querier_GetLobbyPlayers_Call struct {
	*mock.Call
}

// GetLobbyPlayers is a helper method to define mock.On call
//   - ctx context.Context
//   - id int64
func (_e *Querier_Expecter) GetLobbyPlayers(ctx interface{}, id interface{}) *Querier_GetLobbyPlayers_Call {
	return &Querier_GetLobbyPlayers_Call{Call: _e.mock.On("GetLobbyPlayers", ctx, id)}
}

func (_c *Querier_GetLobbyPlayers_Call) Run(run func(ctx context.Context, id int64)) *Querier_GetLobbyPlayers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *Querier_GetLobbyPlayers_Call) Return(_a0 []sqlc.GetLobbyPlayersRow, _a1 error) *Querier_GetLobbyPlayers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_GetLobbyPlayers_Call) RunAndReturn(run func(context.Context, int64) ([]sqlc.GetLobbyPlayersRow, error)) *Querier_GetLobbyPlayers_Call {
	_c.Call.Return(run)
	return _c
}

// GetOwnedLobbies provides a mock function with given fields: ctx, userID
func (_m *Querier) GetOwnedLobbies(ctx context.Context, userID string) ([]sqlc.GetOwnedLobbiesRow, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetOwnedLobbies")
	}

	var r0 []sqlc.GetOwnedLobbiesRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]sqlc.GetOwnedLobbiesRow, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []sqlc.GetOwnedLobbiesRow); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.GetOwnedLobbiesRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_GetOwnedLobbies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOwnedLobbies'
type Querier_GetOwnedLobbies_Call struct {
	*mock.Call
}

// GetOwnedLobbies is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
func (_e *Querier_Expecter) GetOwnedLobbies(ctx interface{}, userID interface{}) *Querier_GetOwnedLobbies_Call {
	return &Querier_GetOwnedLobbies_Call{Call: _e.mock.On("GetOwnedLobbies", ctx, userID)}
}

func (_c *Querier_GetOwnedLobbies_Call) Run(run func(ctx context.Context, userID string)) *Querier_GetOwnedLobbies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Querier_GetOwnedLobbies_Call) Return(_a0 []sqlc.GetOwnedLobbiesRow, _a1 error) *Querier_GetOwnedLobbies_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_GetOwnedLobbies_Call) RunAndReturn(run func(context.Context, string) ([]sqlc.GetOwnedLobbiesRow, error)) *Querier_GetOwnedLobbies_Call {
	_c.Call.Return(run)
	return _c
}

// InsertParticipant provides a mock function with given fields: ctx, arg
func (_m *Querier) InsertParticipant(ctx context.Context, arg sqlc.InsertParticipantParams) (int64, error) {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for InsertParticipant")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.InsertParticipantParams) (int64, error)); ok {
		return rf(ctx, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.InsertParticipantParams) int64); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, sqlc.InsertParticipantParams) error); ok {
		r1 = rf(ctx, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Querier_InsertParticipant_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InsertParticipant'
type Querier_InsertParticipant_Call struct {
	*mock.Call
}

// InsertParticipant is a helper method to define mock.On call
//   - ctx context.Context
//   - arg sqlc.InsertParticipantParams
func (_e *Querier_Expecter) InsertParticipant(ctx interface{}, arg interface{}) *Querier_InsertParticipant_Call {
	return &Querier_InsertParticipant_Call{Call: _e.mock.On("InsertParticipant", ctx, arg)}
}

func (_c *Querier_InsertParticipant_Call) Run(run func(ctx context.Context, arg sqlc.InsertParticipantParams)) *Querier_InsertParticipant_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(sqlc.InsertParticipantParams))
	})
	return _c
}

func (_c *Querier_InsertParticipant_Call) Return(_a0 int64, _a1 error) *Querier_InsertParticipant_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Querier_InsertParticipant_Call) RunAndReturn(run func(context.Context, sqlc.InsertParticipantParams) (int64, error)) *Querier_InsertParticipant_Call {
	_c.Call.Return(run)
	return _c
}

// MarkLobbyAsStarted provides a mock function with given fields: ctx, arg
func (_m *Querier) MarkLobbyAsStarted(ctx context.Context, arg sqlc.MarkLobbyAsStartedParams) error {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for MarkLobbyAsStarted")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.MarkLobbyAsStartedParams) error); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Querier_MarkLobbyAsStarted_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MarkLobbyAsStarted'
type Querier_MarkLobbyAsStarted_Call struct {
	*mock.Call
}

// MarkLobbyAsStarted is a helper method to define mock.On call
//   - ctx context.Context
//   - arg sqlc.MarkLobbyAsStartedParams
func (_e *Querier_Expecter) MarkLobbyAsStarted(ctx interface{}, arg interface{}) *Querier_MarkLobbyAsStarted_Call {
	return &Querier_MarkLobbyAsStarted_Call{Call: _e.mock.On("MarkLobbyAsStarted", ctx, arg)}
}

func (_c *Querier_MarkLobbyAsStarted_Call) Run(run func(ctx context.Context, arg sqlc.MarkLobbyAsStartedParams)) *Querier_MarkLobbyAsStarted_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(sqlc.MarkLobbyAsStartedParams))
	})
	return _c
}

func (_c *Querier_MarkLobbyAsStarted_Call) Return(_a0 error) *Querier_MarkLobbyAsStarted_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Querier_MarkLobbyAsStarted_Call) RunAndReturn(run func(context.Context, sqlc.MarkLobbyAsStartedParams) error) *Querier_MarkLobbyAsStarted_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateLobbyOwner provides a mock function with given fields: ctx, arg
func (_m *Querier) UpdateLobbyOwner(ctx context.Context, arg sqlc.UpdateLobbyOwnerParams) error {
	ret := _m.Called(ctx, arg)

	if len(ret) == 0 {
		panic("no return value specified for UpdateLobbyOwner")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.UpdateLobbyOwnerParams) error); ok {
		r0 = rf(ctx, arg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Querier_UpdateLobbyOwner_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateLobbyOwner'
type Querier_UpdateLobbyOwner_Call struct {
	*mock.Call
}

// UpdateLobbyOwner is a helper method to define mock.On call
//   - ctx context.Context
//   - arg sqlc.UpdateLobbyOwnerParams
func (_e *Querier_Expecter) UpdateLobbyOwner(ctx interface{}, arg interface{}) *Querier_UpdateLobbyOwner_Call {
	return &Querier_UpdateLobbyOwner_Call{Call: _e.mock.On("UpdateLobbyOwner", ctx, arg)}
}

func (_c *Querier_UpdateLobbyOwner_Call) Run(run func(ctx context.Context, arg sqlc.UpdateLobbyOwnerParams)) *Querier_UpdateLobbyOwner_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(sqlc.UpdateLobbyOwnerParams))
	})
	return _c
}

func (_c *Querier_UpdateLobbyOwner_Call) Return(_a0 error) *Querier_UpdateLobbyOwner_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Querier_UpdateLobbyOwner_Call) RunAndReturn(run func(context.Context, sqlc.UpdateLobbyOwnerParams) error) *Querier_UpdateLobbyOwner_Call {
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
