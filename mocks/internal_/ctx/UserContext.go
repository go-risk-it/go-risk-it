// Code generated by mockery v2.50.1. DO NOT EDIT.

package ctx

import (
	time "time"

	mock "github.com/stretchr/testify/mock"

	trace "go.opentelemetry.io/otel/trace"

	zap "go.uber.org/zap"
)

// UserContext is an autogenerated mock type for the UserContext type
type UserContext struct {
	mock.Mock
}

type UserContext_Expecter struct {
	mock *mock.Mock
}

func (_m *UserContext) EXPECT() *UserContext_Expecter {
	return &UserContext_Expecter{mock: &_m.Mock}
}

// Deadline provides a mock function with no fields
func (_m *UserContext) Deadline() (time.Time, bool) {
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

// UserContext_Deadline_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Deadline'
type UserContext_Deadline_Call struct {
	*mock.Call
}

// Deadline is a helper method to define mock.On call
func (_e *UserContext_Expecter) Deadline() *UserContext_Deadline_Call {
	return &UserContext_Deadline_Call{Call: _e.mock.On("Deadline")}
}

func (_c *UserContext_Deadline_Call) Run(run func()) *UserContext_Deadline_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *UserContext_Deadline_Call) Return(deadline time.Time, ok bool) *UserContext_Deadline_Call {
	_c.Call.Return(deadline, ok)
	return _c
}

func (_c *UserContext_Deadline_Call) RunAndReturn(run func() (time.Time, bool)) *UserContext_Deadline_Call {
	_c.Call.Return(run)
	return _c
}

// Done provides a mock function with no fields
func (_m *UserContext) Done() <-chan struct{} {
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

// UserContext_Done_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Done'
type UserContext_Done_Call struct {
	*mock.Call
}

// Done is a helper method to define mock.On call
func (_e *UserContext_Expecter) Done() *UserContext_Done_Call {
	return &UserContext_Done_Call{Call: _e.mock.On("Done")}
}

func (_c *UserContext_Done_Call) Run(run func()) *UserContext_Done_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *UserContext_Done_Call) Return(_a0 <-chan struct{}) *UserContext_Done_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *UserContext_Done_Call) RunAndReturn(run func() <-chan struct{}) *UserContext_Done_Call {
	_c.Call.Return(run)
	return _c
}

// Err provides a mock function with no fields
func (_m *UserContext) Err() error {
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

// UserContext_Err_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Err'
type UserContext_Err_Call struct {
	*mock.Call
}

// Err is a helper method to define mock.On call
func (_e *UserContext_Expecter) Err() *UserContext_Err_Call {
	return &UserContext_Err_Call{Call: _e.mock.On("Err")}
}

func (_c *UserContext_Err_Call) Run(run func()) *UserContext_Err_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *UserContext_Err_Call) Return(_a0 error) *UserContext_Err_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *UserContext_Err_Call) RunAndReturn(run func() error) *UserContext_Err_Call {
	_c.Call.Return(run)
	return _c
}

// Log provides a mock function with no fields
func (_m *UserContext) Log() *zap.SugaredLogger {
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

// UserContext_Log_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Log'
type UserContext_Log_Call struct {
	*mock.Call
}

// Log is a helper method to define mock.On call
func (_e *UserContext_Expecter) Log() *UserContext_Log_Call {
	return &UserContext_Log_Call{Call: _e.mock.On("Log")}
}

func (_c *UserContext_Log_Call) Run(run func()) *UserContext_Log_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *UserContext_Log_Call) Return(_a0 *zap.SugaredLogger) *UserContext_Log_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *UserContext_Log_Call) RunAndReturn(run func() *zap.SugaredLogger) *UserContext_Log_Call {
	_c.Call.Return(run)
	return _c
}

// SetLog provides a mock function with given fields: log
func (_m *UserContext) SetLog(log *zap.SugaredLogger) {
	_m.Called(log)
}

// UserContext_SetLog_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetLog'
type UserContext_SetLog_Call struct {
	*mock.Call
}

// SetLog is a helper method to define mock.On call
//   - log *zap.SugaredLogger
func (_e *UserContext_Expecter) SetLog(log interface{}) *UserContext_SetLog_Call {
	return &UserContext_SetLog_Call{Call: _e.mock.On("SetLog", log)}
}

func (_c *UserContext_SetLog_Call) Run(run func(log *zap.SugaredLogger)) *UserContext_SetLog_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*zap.SugaredLogger))
	})
	return _c
}

func (_c *UserContext_SetLog_Call) Return() *UserContext_SetLog_Call {
	_c.Call.Return()
	return _c
}

func (_c *UserContext_SetLog_Call) RunAndReturn(run func(*zap.SugaredLogger)) *UserContext_SetLog_Call {
	_c.Run(run)
	return _c
}

// Span provides a mock function with no fields
func (_m *UserContext) Span() trace.Span {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Span")
	}

	var r0 trace.Span
	if rf, ok := ret.Get(0).(func() trace.Span); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(trace.Span)
		}
	}

	return r0
}

// UserContext_Span_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Span'
type UserContext_Span_Call struct {
	*mock.Call
}

// Span is a helper method to define mock.On call
func (_e *UserContext_Expecter) Span() *UserContext_Span_Call {
	return &UserContext_Span_Call{Call: _e.mock.On("Span")}
}

func (_c *UserContext_Span_Call) Run(run func()) *UserContext_Span_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *UserContext_Span_Call) Return(_a0 trace.Span) *UserContext_Span_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *UserContext_Span_Call) RunAndReturn(run func() trace.Span) *UserContext_Span_Call {
	_c.Call.Return(run)
	return _c
}

// UserID provides a mock function with no fields
func (_m *UserContext) UserID() string {
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

// UserContext_UserID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UserID'
type UserContext_UserID_Call struct {
	*mock.Call
}

// UserID is a helper method to define mock.On call
func (_e *UserContext_Expecter) UserID() *UserContext_UserID_Call {
	return &UserContext_UserID_Call{Call: _e.mock.On("UserID")}
}

func (_c *UserContext_UserID_Call) Run(run func()) *UserContext_UserID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *UserContext_UserID_Call) Return(_a0 string) *UserContext_UserID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *UserContext_UserID_Call) RunAndReturn(run func() string) *UserContext_UserID_Call {
	_c.Call.Return(run)
	return _c
}

// Value provides a mock function with given fields: key
func (_m *UserContext) Value(key any) any {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for Value")
	}

	var r0 any
	if rf, ok := ret.Get(0).(func(any) any); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(any)
		}
	}

	return r0
}

// UserContext_Value_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Value'
type UserContext_Value_Call struct {
	*mock.Call
}

// Value is a helper method to define mock.On call
//   - key any
func (_e *UserContext_Expecter) Value(key interface{}) *UserContext_Value_Call {
	return &UserContext_Value_Call{Call: _e.mock.On("Value", key)}
}

func (_c *UserContext_Value_Call) Run(run func(key any)) *UserContext_Value_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(any))
	})
	return _c
}

func (_c *UserContext_Value_Call) Return(_a0 any) *UserContext_Value_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *UserContext_Value_Call) RunAndReturn(run func(any) any) *UserContext_Value_Call {
	_c.Call.Return(run)
	return _c
}

// NewUserContext creates a new instance of UserContext. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUserContext(t interface {
	mock.TestingT
	Cleanup(func())
}) *UserContext {
	mock := &UserContext{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
