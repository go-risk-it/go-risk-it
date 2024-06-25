// Code generated by mockery v2.43.1. DO NOT EDIT.

package validation

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

// Validate provides a mock function with given fields: _a0, querier, game
func (_m *Service) Validate(_a0 ctx.MoveContext, querier db.Querier, game *sqlc.Game) error {
	ret := _m.Called(_a0, querier, game)

	if len(ret) == 0 {
		panic("no return value specified for Validate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.MoveContext, db.Querier, *sqlc.Game) error); ok {
		r0 = rf(_a0, querier, game)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_Validate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Validate'
type Service_Validate_Call struct {
	*mock.Call
}

// Validate is a helper method to define mock.On call
//   - _a0 ctx.MoveContext
//   - querier db.Querier
//   - game *sqlc.Game
func (_e *Service_Expecter) Validate(_a0 interface{}, querier interface{}, game interface{}) *Service_Validate_Call {
	return &Service_Validate_Call{Call: _e.mock.On("Validate", _a0, querier, game)}
}

func (_c *Service_Validate_Call) Run(run func(_a0 ctx.MoveContext, querier db.Querier, game *sqlc.Game)) *Service_Validate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.MoveContext), args[1].(db.Querier), args[2].(*sqlc.Game))
	})
	return _c
}

func (_c *Service_Validate_Call) Return(_a0 error) *Service_Validate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_Validate_Call) RunAndReturn(run func(ctx.MoveContext, db.Querier, *sqlc.Game) error) *Service_Validate_Call {
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
