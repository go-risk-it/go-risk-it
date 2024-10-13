// Code generated by mockery v2.46.2. DO NOT EDIT.

package testonly

import (
	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"
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

// TruncateTables provides a mock function with given fields: _a0
func (_m *Service) TruncateTables(_a0 ctx.LogContext) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for TruncateTables")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.LogContext) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_TruncateTables_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TruncateTables'
type Service_TruncateTables_Call struct {
	*mock.Call
}

// TruncateTables is a helper method to define mock.On call
//   - _a0 ctx.LogContext
func (_e *Service_Expecter) TruncateTables(_a0 interface{}) *Service_TruncateTables_Call {
	return &Service_TruncateTables_Call{Call: _e.mock.On("TruncateTables", _a0)}
}

func (_c *Service_TruncateTables_Call) Run(run func(_a0 ctx.LogContext)) *Service_TruncateTables_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.LogContext))
	})
	return _c
}

func (_c *Service_TruncateTables_Call) Return(_a0 error) *Service_TruncateTables_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_TruncateTables_Call) RunAndReturn(run func(ctx.LogContext) error) *Service_TruncateTables_Call {
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
