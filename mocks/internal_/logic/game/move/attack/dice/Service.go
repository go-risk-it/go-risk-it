// Code generated by mockery v2.46.2. DO NOT EDIT.

package dice

import mock "github.com/stretchr/testify/mock"

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

// RollAttackingDices provides a mock function with given fields: n
func (_m *Service) RollAttackingDices(n int) []int {
	ret := _m.Called(n)

	if len(ret) == 0 {
		panic("no return value specified for RollAttackingDices")
	}

	var r0 []int
	if rf, ok := ret.Get(0).(func(int) []int); ok {
		r0 = rf(n)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]int)
		}
	}

	return r0
}

// Service_RollAttackingDices_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RollAttackingDices'
type Service_RollAttackingDices_Call struct {
	*mock.Call
}

// RollAttackingDices is a helper method to define mock.On call
//   - n int
func (_e *Service_Expecter) RollAttackingDices(n interface{}) *Service_RollAttackingDices_Call {
	return &Service_RollAttackingDices_Call{Call: _e.mock.On("RollAttackingDices", n)}
}

func (_c *Service_RollAttackingDices_Call) Run(run func(n int)) *Service_RollAttackingDices_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *Service_RollAttackingDices_Call) Return(_a0 []int) *Service_RollAttackingDices_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_RollAttackingDices_Call) RunAndReturn(run func(int) []int) *Service_RollAttackingDices_Call {
	_c.Call.Return(run)
	return _c
}

// RollDefendingDices provides a mock function with given fields: n
func (_m *Service) RollDefendingDices(n int) []int {
	ret := _m.Called(n)

	if len(ret) == 0 {
		panic("no return value specified for RollDefendingDices")
	}

	var r0 []int
	if rf, ok := ret.Get(0).(func(int) []int); ok {
		r0 = rf(n)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]int)
		}
	}

	return r0
}

// Service_RollDefendingDices_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RollDefendingDices'
type Service_RollDefendingDices_Call struct {
	*mock.Call
}

// RollDefendingDices is a helper method to define mock.On call
//   - n int
func (_e *Service_Expecter) RollDefendingDices(n interface{}) *Service_RollDefendingDices_Call {
	return &Service_RollDefendingDices_Call{Call: _e.mock.On("RollDefendingDices", n)}
}

func (_c *Service_RollDefendingDices_Call) Run(run func(n int)) *Service_RollDefendingDices_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *Service_RollDefendingDices_Call) Return(_a0 []int) *Service_RollDefendingDices_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_RollDefendingDices_Call) RunAndReturn(run func(int) []int) *Service_RollDefendingDices_Call {
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
