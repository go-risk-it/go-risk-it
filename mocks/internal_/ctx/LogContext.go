// Code generated by mockery v2.50.1. DO NOT EDIT.

package ctx

import (
	time "time"

	mock "github.com/stretchr/testify/mock"

	zap "go.uber.org/zap"
)

// LogContext is an autogenerated mock type for the LogContext type
type LogContext struct {
	mock.Mock
}

type LogContext_Expecter struct {
	mock *mock.Mock
}

func (_m *LogContext) EXPECT() *LogContext_Expecter {
	return &LogContext_Expecter{mock: &_m.Mock}
}

// Deadline provides a mock function with no fields
func (_m *LogContext) Deadline() (time.Time, bool) {
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

// LogContext_Deadline_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Deadline'
type LogContext_Deadline_Call struct {
	*mock.Call
}

// Deadline is a helper method to define mock.On call
func (_e *LogContext_Expecter) Deadline() *LogContext_Deadline_Call {
	return &LogContext_Deadline_Call{Call: _e.mock.On("Deadline")}
}

func (_c *LogContext_Deadline_Call) Run(run func()) *LogContext_Deadline_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *LogContext_Deadline_Call) Return(deadline time.Time, ok bool) *LogContext_Deadline_Call {
	_c.Call.Return(deadline, ok)
	return _c
}

func (_c *LogContext_Deadline_Call) RunAndReturn(run func() (time.Time, bool)) *LogContext_Deadline_Call {
	_c.Call.Return(run)
	return _c
}

// Done provides a mock function with no fields
func (_m *LogContext) Done() <-chan struct{} {
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

// LogContext_Done_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Done'
type LogContext_Done_Call struct {
	*mock.Call
}

// Done is a helper method to define mock.On call
func (_e *LogContext_Expecter) Done() *LogContext_Done_Call {
	return &LogContext_Done_Call{Call: _e.mock.On("Done")}
}

func (_c *LogContext_Done_Call) Run(run func()) *LogContext_Done_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *LogContext_Done_Call) Return(_a0 <-chan struct{}) *LogContext_Done_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *LogContext_Done_Call) RunAndReturn(run func() <-chan struct{}) *LogContext_Done_Call {
	_c.Call.Return(run)
	return _c
}

// Err provides a mock function with no fields
func (_m *LogContext) Err() error {
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

// LogContext_Err_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Err'
type LogContext_Err_Call struct {
	*mock.Call
}

// Err is a helper method to define mock.On call
func (_e *LogContext_Expecter) Err() *LogContext_Err_Call {
	return &LogContext_Err_Call{Call: _e.mock.On("Err")}
}

func (_c *LogContext_Err_Call) Run(run func()) *LogContext_Err_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *LogContext_Err_Call) Return(_a0 error) *LogContext_Err_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *LogContext_Err_Call) RunAndReturn(run func() error) *LogContext_Err_Call {
	_c.Call.Return(run)
	return _c
}

// Log provides a mock function with no fields
func (_m *LogContext) Log() *zap.SugaredLogger {
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

// LogContext_Log_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Log'
type LogContext_Log_Call struct {
	*mock.Call
}

// Log is a helper method to define mock.On call
func (_e *LogContext_Expecter) Log() *LogContext_Log_Call {
	return &LogContext_Log_Call{Call: _e.mock.On("Log")}
}

func (_c *LogContext_Log_Call) Run(run func()) *LogContext_Log_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *LogContext_Log_Call) Return(_a0 *zap.SugaredLogger) *LogContext_Log_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *LogContext_Log_Call) RunAndReturn(run func() *zap.SugaredLogger) *LogContext_Log_Call {
	_c.Call.Return(run)
	return _c
}

// SetLog provides a mock function with given fields: log
func (_m *LogContext) SetLog(log *zap.SugaredLogger) {
	_m.Called(log)
}

// LogContext_SetLog_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetLog'
type LogContext_SetLog_Call struct {
	*mock.Call
}

// SetLog is a helper method to define mock.On call
//   - log *zap.SugaredLogger
func (_e *LogContext_Expecter) SetLog(log interface{}) *LogContext_SetLog_Call {
	return &LogContext_SetLog_Call{Call: _e.mock.On("SetLog", log)}
}

func (_c *LogContext_SetLog_Call) Run(run func(log *zap.SugaredLogger)) *LogContext_SetLog_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*zap.SugaredLogger))
	})
	return _c
}

func (_c *LogContext_SetLog_Call) Return() *LogContext_SetLog_Call {
	_c.Call.Return()
	return _c
}

func (_c *LogContext_SetLog_Call) RunAndReturn(run func(*zap.SugaredLogger)) *LogContext_SetLog_Call {
	_c.Run(run)
	return _c
}

// Value provides a mock function with given fields: key
func (_m *LogContext) Value(key interface{}) interface{} {
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

// LogContext_Value_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Value'
type LogContext_Value_Call struct {
	*mock.Call
}

// Value is a helper method to define mock.On call
//   - key interface{}
func (_e *LogContext_Expecter) Value(key interface{}) *LogContext_Value_Call {
	return &LogContext_Value_Call{Call: _e.mock.On("Value", key)}
}

func (_c *LogContext_Value_Call) Run(run func(key interface{})) *LogContext_Value_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *LogContext_Value_Call) Return(_a0 interface{}) *LogContext_Value_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *LogContext_Value_Call) RunAndReturn(run func(interface{}) interface{}) *LogContext_Value_Call {
	_c.Call.Return(run)
	return _c
}

// NewLogContext creates a new instance of LogContext. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLogContext(t interface {
	mock.TestingT
	Cleanup(func())
}) *LogContext {
	mock := &LogContext{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
