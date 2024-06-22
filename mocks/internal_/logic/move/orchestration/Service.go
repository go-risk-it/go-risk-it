// Code generated by mockery v2.43.1. DO NOT EDIT.

package orchestration

import (
	context "context"

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

// AdvancePhaseQ provides a mock function with given fields: ctx, querier, gameID
func (_m *Service) AdvancePhaseQ(ctx context.Context, querier db.Querier, gameID int64) error {
	ret := _m.Called(ctx, querier, gameID)

	if len(ret) == 0 {
		panic("no return value specified for AdvancePhaseQ")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, db.Querier, int64) error); ok {
		r0 = rf(ctx, querier, gameID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_AdvancePhaseQ_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AdvancePhaseQ'
type Service_AdvancePhaseQ_Call struct {
	*mock.Call
}

// AdvancePhaseQ is a helper method to define mock.On call
//   - ctx context.Context
//   - querier db.Querier
//   - gameID int64
func (_e *Service_Expecter) AdvancePhaseQ(ctx interface{}, querier interface{}, gameID interface{}) *Service_AdvancePhaseQ_Call {
	return &Service_AdvancePhaseQ_Call{Call: _e.mock.On("AdvancePhaseQ", ctx, querier, gameID)}
}

func (_c *Service_AdvancePhaseQ_Call) Run(run func(ctx context.Context, querier db.Querier, gameID int64)) *Service_AdvancePhaseQ_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(db.Querier), args[2].(int64))
	})
	return _c
}

func (_c *Service_AdvancePhaseQ_Call) Return(_a0 error) *Service_AdvancePhaseQ_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_AdvancePhaseQ_Call) RunAndReturn(run func(context.Context, db.Querier, int64) error) *Service_AdvancePhaseQ_Call {
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
