// Code generated by mockery v2.50.1. DO NOT EDIT.

package ctx

import (
	time "time"

	mock "github.com/stretchr/testify/mock"

	trace "go.opentelemetry.io/otel/trace"

	zap "go.uber.org/zap"
)

// TraceContext is an autogenerated mock type for the TraceContext type
type TraceContext struct {
	mock.Mock
}

type TraceContext_Expecter struct {
	mock *mock.Mock
}

func (_m *TraceContext) EXPECT() *TraceContext_Expecter {
	return &TraceContext_Expecter{mock: &_m.Mock}
}

// Deadline provides a mock function with no fields
func (_m *TraceContext) Deadline() (time.Time, bool) {
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

// TraceContext_Deadline_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Deadline'
type TraceContext_Deadline_Call struct {
	*mock.Call
}

// Deadline is a helper method to define mock.On call
func (_e *TraceContext_Expecter) Deadline() *TraceContext_Deadline_Call {
	return &TraceContext_Deadline_Call{Call: _e.mock.On("Deadline")}
}

func (_c *TraceContext_Deadline_Call) Run(run func()) *TraceContext_Deadline_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *TraceContext_Deadline_Call) Return(deadline time.Time, ok bool) *TraceContext_Deadline_Call {
	_c.Call.Return(deadline, ok)
	return _c
}

func (_c *TraceContext_Deadline_Call) RunAndReturn(run func() (time.Time, bool)) *TraceContext_Deadline_Call {
	_c.Call.Return(run)
	return _c
}

// Done provides a mock function with no fields
func (_m *TraceContext) Done() <-chan struct{} {
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

// TraceContext_Done_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Done'
type TraceContext_Done_Call struct {
	*mock.Call
}

// Done is a helper method to define mock.On call
func (_e *TraceContext_Expecter) Done() *TraceContext_Done_Call {
	return &TraceContext_Done_Call{Call: _e.mock.On("Done")}
}

func (_c *TraceContext_Done_Call) Run(run func()) *TraceContext_Done_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *TraceContext_Done_Call) Return(_a0 <-chan struct{}) *TraceContext_Done_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TraceContext_Done_Call) RunAndReturn(run func() <-chan struct{}) *TraceContext_Done_Call {
	_c.Call.Return(run)
	return _c
}

// Err provides a mock function with no fields
func (_m *TraceContext) Err() error {
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

// TraceContext_Err_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Err'
type TraceContext_Err_Call struct {
	*mock.Call
}

// Err is a helper method to define mock.On call
func (_e *TraceContext_Expecter) Err() *TraceContext_Err_Call {
	return &TraceContext_Err_Call{Call: _e.mock.On("Err")}
}

func (_c *TraceContext_Err_Call) Run(run func()) *TraceContext_Err_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *TraceContext_Err_Call) Return(_a0 error) *TraceContext_Err_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TraceContext_Err_Call) RunAndReturn(run func() error) *TraceContext_Err_Call {
	_c.Call.Return(run)
	return _c
}

// Log provides a mock function with no fields
func (_m *TraceContext) Log() *zap.SugaredLogger {
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

// TraceContext_Log_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Log'
type TraceContext_Log_Call struct {
	*mock.Call
}

// Log is a helper method to define mock.On call
func (_e *TraceContext_Expecter) Log() *TraceContext_Log_Call {
	return &TraceContext_Log_Call{Call: _e.mock.On("Log")}
}

func (_c *TraceContext_Log_Call) Run(run func()) *TraceContext_Log_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *TraceContext_Log_Call) Return(_a0 *zap.SugaredLogger) *TraceContext_Log_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TraceContext_Log_Call) RunAndReturn(run func() *zap.SugaredLogger) *TraceContext_Log_Call {
	_c.Call.Return(run)
	return _c
}

// SetLog provides a mock function with given fields: log
func (_m *TraceContext) SetLog(log *zap.SugaredLogger) {
	_m.Called(log)
}

// TraceContext_SetLog_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetLog'
type TraceContext_SetLog_Call struct {
	*mock.Call
}

// SetLog is a helper method to define mock.On call
//   - log *zap.SugaredLogger
func (_e *TraceContext_Expecter) SetLog(log interface{}) *TraceContext_SetLog_Call {
	return &TraceContext_SetLog_Call{Call: _e.mock.On("SetLog", log)}
}

func (_c *TraceContext_SetLog_Call) Run(run func(log *zap.SugaredLogger)) *TraceContext_SetLog_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*zap.SugaredLogger))
	})
	return _c
}

func (_c *TraceContext_SetLog_Call) Return() *TraceContext_SetLog_Call {
	_c.Call.Return()
	return _c
}

func (_c *TraceContext_SetLog_Call) RunAndReturn(run func(*zap.SugaredLogger)) *TraceContext_SetLog_Call {
	_c.Run(run)
	return _c
}

// Span provides a mock function with no fields
func (_m *TraceContext) Span() trace.Span {
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

// TraceContext_Span_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Span'
type TraceContext_Span_Call struct {
	*mock.Call
}

// Span is a helper method to define mock.On call
func (_e *TraceContext_Expecter) Span() *TraceContext_Span_Call {
	return &TraceContext_Span_Call{Call: _e.mock.On("Span")}
}

func (_c *TraceContext_Span_Call) Run(run func()) *TraceContext_Span_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *TraceContext_Span_Call) Return(_a0 trace.Span) *TraceContext_Span_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TraceContext_Span_Call) RunAndReturn(run func() trace.Span) *TraceContext_Span_Call {
	_c.Call.Return(run)
	return _c
}

// Value provides a mock function with given fields: key
func (_m *TraceContext) Value(key any) any {
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

// TraceContext_Value_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Value'
type TraceContext_Value_Call struct {
	*mock.Call
}

// Value is a helper method to define mock.On call
//   - key any
func (_e *TraceContext_Expecter) Value(key interface{}) *TraceContext_Value_Call {
	return &TraceContext_Value_Call{Call: _e.mock.On("Value", key)}
}

func (_c *TraceContext_Value_Call) Run(run func(key any)) *TraceContext_Value_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(any))
	})
	return _c
}

func (_c *TraceContext_Value_Call) Return(_a0 any) *TraceContext_Value_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TraceContext_Value_Call) RunAndReturn(run func(any) any) *TraceContext_Value_Call {
	_c.Call.Return(run)
	return _c
}

// NewTraceContext creates a new instance of TraceContext. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTraceContext(t interface {
	mock.TestingT
	Cleanup(func())
}) *TraceContext {
	mock := &TraceContext{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
