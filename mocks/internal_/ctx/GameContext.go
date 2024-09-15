// Code generated by mockery v2.44.1. DO NOT EDIT.

package ctx

import (
	time "time"

	mock "github.com/stretchr/testify/mock"

	zap "go.uber.org/zap"
)

// GameContext is an autogenerated mock type for the GameContext type
type GameContext struct {
	mock.Mock
}

type GameContext_Expecter struct {
	mock *mock.Mock
}

func (_m *GameContext) EXPECT() *GameContext_Expecter {
	return &GameContext_Expecter{mock: &_m.Mock}
}

// Deadline provides a mock function with given fields:
func (_m *GameContext) Deadline() (time.Time, bool) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Deadline")
	}

	var r0 time.Time
	var r1 bool
	if rf, ok := ret.Get(0).(func() (time.Time, bool)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	if rf, ok := ret.Get(1).(func() bool); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// GameContext_Deadline_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Deadline'
type GameContext_Deadline_Call struct {
	*mock.Call
}

// Deadline is a helper method to define mock.On call
func (_e *GameContext_Expecter) Deadline() *GameContext_Deadline_Call {
	return &GameContext_Deadline_Call{Call: _e.mock.On("Deadline")}
}

func (_c *GameContext_Deadline_Call) Run(run func()) *GameContext_Deadline_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *GameContext_Deadline_Call) Return(deadline time.Time, ok bool) *GameContext_Deadline_Call {
	_c.Call.Return(deadline, ok)
	return _c
}

func (_c *GameContext_Deadline_Call) RunAndReturn(run func() (time.Time, bool)) *GameContext_Deadline_Call {
	_c.Call.Return(run)
	return _c
}

// Done provides a mock function with given fields:
func (_m *GameContext) Done() <-chan struct{} {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Done")
	}

	var r0 <-chan struct{}
	if rf, ok := ret.Get(0).(func() <-chan struct{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}

	return r0
}

// GameContext_Done_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Done'
type GameContext_Done_Call struct {
	*mock.Call
}

// Done is a helper method to define mock.On call
func (_e *GameContext_Expecter) Done() *GameContext_Done_Call {
	return &GameContext_Done_Call{Call: _e.mock.On("Done")}
}

func (_c *GameContext_Done_Call) Run(run func()) *GameContext_Done_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *GameContext_Done_Call) Return(_a0 <-chan struct{}) *GameContext_Done_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GameContext_Done_Call) RunAndReturn(run func() <-chan struct{}) *GameContext_Done_Call {
	_c.Call.Return(run)
	return _c
}

// Err provides a mock function with given fields:
func (_m *GameContext) Err() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Err")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GameContext_Err_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Err'
type GameContext_Err_Call struct {
	*mock.Call
}

// Err is a helper method to define mock.On call
func (_e *GameContext_Expecter) Err() *GameContext_Err_Call {
	return &GameContext_Err_Call{Call: _e.mock.On("Err")}
}

func (_c *GameContext_Err_Call) Run(run func()) *GameContext_Err_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *GameContext_Err_Call) Return(_a0 error) *GameContext_Err_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GameContext_Err_Call) RunAndReturn(run func() error) *GameContext_Err_Call {
	_c.Call.Return(run)
	return _c
}

// GameID provides a mock function with given fields:
func (_m *GameContext) GameID() int64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GameID")
	}

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// GameContext_GameID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GameID'
type GameContext_GameID_Call struct {
	*mock.Call
}

// GameID is a helper method to define mock.On call
func (_e *GameContext_Expecter) GameID() *GameContext_GameID_Call {
	return &GameContext_GameID_Call{Call: _e.mock.On("GameID")}
}

func (_c *GameContext_GameID_Call) Run(run func()) *GameContext_GameID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *GameContext_GameID_Call) Return(_a0 int64) *GameContext_GameID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GameContext_GameID_Call) RunAndReturn(run func() int64) *GameContext_GameID_Call {
	_c.Call.Return(run)
	return _c
}

// Log provides a mock function with given fields:
func (_m *GameContext) Log() *zap.SugaredLogger {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Log")
	}

	var r0 *zap.SugaredLogger
	if rf, ok := ret.Get(0).(func() *zap.SugaredLogger); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*zap.SugaredLogger)
		}
	}

	return r0
}

// GameContext_Log_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Log'
type GameContext_Log_Call struct {
	*mock.Call
}

// Log is a helper method to define mock.On call
func (_e *GameContext_Expecter) Log() *GameContext_Log_Call {
	return &GameContext_Log_Call{Call: _e.mock.On("Log")}
}

func (_c *GameContext_Log_Call) Run(run func()) *GameContext_Log_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *GameContext_Log_Call) Return(_a0 *zap.SugaredLogger) *GameContext_Log_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GameContext_Log_Call) RunAndReturn(run func() *zap.SugaredLogger) *GameContext_Log_Call {
	_c.Call.Return(run)
	return _c
}

// UserID provides a mock function with given fields:
func (_m *GameContext) UserID() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for UserID")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GameContext_UserID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UserID'
type GameContext_UserID_Call struct {
	*mock.Call
}

// UserID is a helper method to define mock.On call
func (_e *GameContext_Expecter) UserID() *GameContext_UserID_Call {
	return &GameContext_UserID_Call{Call: _e.mock.On("UserID")}
}

func (_c *GameContext_UserID_Call) Run(run func()) *GameContext_UserID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *GameContext_UserID_Call) Return(_a0 string) *GameContext_UserID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GameContext_UserID_Call) RunAndReturn(run func() string) *GameContext_UserID_Call {
	_c.Call.Return(run)
	return _c
}

// Value provides a mock function with given fields: key
func (_m *GameContext) Value(key interface{}) interface{} {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for Value")
	}

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(interface{}) interface{}); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// GameContext_Value_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Value'
type GameContext_Value_Call struct {
	*mock.Call
}

// Value is a helper method to define mock.On call
//   - key interface{}
func (_e *GameContext_Expecter) Value(key interface{}) *GameContext_Value_Call {
	return &GameContext_Value_Call{Call: _e.mock.On("Value", key)}
}

func (_c *GameContext_Value_Call) Run(run func(key interface{})) *GameContext_Value_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *GameContext_Value_Call) Return(_a0 interface{}) *GameContext_Value_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GameContext_Value_Call) RunAndReturn(run func(interface{}) interface{}) *GameContext_Value_Call {
	_c.Call.Return(run)
	return _c
}

// NewGameContext creates a new instance of GameContext. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGameContext(t interface {
	mock.TestingT
	Cleanup(func())
}) *GameContext {
	mock := &GameContext{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
