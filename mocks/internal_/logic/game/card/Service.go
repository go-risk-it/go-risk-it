// Code generated by mockery v2.44.1. DO NOT EDIT.

package card

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
	db "github.com/go-risk-it/go-risk-it/internal/data/db"

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

// CreateCards provides a mock function with given fields: _a0, querier
func (_m *Service) CreateCards(_a0 ctx.GameContext, querier db.Querier) error {
	ret := _m.Called(_a0, querier)

	if len(ret) == 0 {
		panic("no return value specified for CreateCards")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.GameContext, db.Querier) error); ok {
		r0 = rf(_a0, querier)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_CreateCards_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateCards'
type Service_CreateCards_Call struct {
	*mock.Call
}

// CreateCards is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - querier db.Querier
func (_e *Service_Expecter) CreateCards(_a0 interface{}, querier interface{}) *Service_CreateCards_Call {
	return &Service_CreateCards_Call{Call: _e.mock.On("CreateCards", _a0, querier)}
}

func (_c *Service_CreateCards_Call) Run(run func(_a0 ctx.GameContext, querier db.Querier)) *Service_CreateCards_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(db.Querier))
	})
	return _c
}

func (_c *Service_CreateCards_Call) Return(_a0 error) *Service_CreateCards_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_CreateCards_Call) RunAndReturn(run func(ctx.GameContext, db.Querier) error) *Service_CreateCards_Call {
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
