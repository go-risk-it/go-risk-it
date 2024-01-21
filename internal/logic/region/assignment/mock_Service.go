// Code generated by mockery v2.40.1. DO NOT EDIT.

package assignment

import (
	db "github.com/tomfran/go-risk-it/internal/db"
	board "github.com/tomfran/go-risk-it/internal/logic/board"

	mock "github.com/stretchr/testify/mock"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

type MockService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockService) EXPECT() *MockService_Expecter {
	return &MockService_Expecter{mock: &_m.Mock}
}

// AssignRegionsToPlayers provides a mock function with given fields: players, regions
func (_m *MockService) AssignRegionsToPlayers(players []db.Player, regions []board.Region) RegionAssignment {
	ret := _m.Called(players, regions)

	if len(ret) == 0 {
		panic("no return value specified for AssignRegionsToPlayers")
	}

	var r0 RegionAssignment
	if rf, ok := ret.Get(0).(func([]db.Player, []board.Region) RegionAssignment); ok {
		r0 = rf(players, regions)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(RegionAssignment)
		}
	}

	return r0
}

// MockService_AssignRegionsToPlayers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AssignRegionsToPlayers'
type MockService_AssignRegionsToPlayers_Call struct {
	*mock.Call
}

// AssignRegionsToPlayers is a helper method to define mock.On call
//   - players []db.Player
//   - regions []board.Region
func (_e *MockService_Expecter) AssignRegionsToPlayers(players interface{}, regions interface{}) *MockService_AssignRegionsToPlayers_Call {
	return &MockService_AssignRegionsToPlayers_Call{Call: _e.mock.On("AssignRegionsToPlayers", players, regions)}
}

func (_c *MockService_AssignRegionsToPlayers_Call) Run(run func(players []db.Player, regions []board.Region)) *MockService_AssignRegionsToPlayers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]db.Player), args[1].([]board.Region))
	})
	return _c
}

func (_c *MockService_AssignRegionsToPlayers_Call) Return(_a0 RegionAssignment) *MockService_AssignRegionsToPlayers_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_AssignRegionsToPlayers_Call) RunAndReturn(run func([]db.Player, []board.Region) RegionAssignment) *MockService_AssignRegionsToPlayers_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockService creates a new instance of MockService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockService {
	mock := &MockService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
