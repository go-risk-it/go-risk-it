// Code generated by mockery v2.40.1. DO NOT EDIT.

package signals

import (
	context "context"
	logicsignals "github.com/tomfran/go-risk-it/internal/signals"

	signals "github.com/maniartech/signals"
	mock "github.com/stretchr/testify/mock"
)

// BoardStateChangedSignal is an autogenerated mock type for the BoardStateChangedSignal type
type BoardStateChangedSignal struct {
	mock.Mock
}

type BoardStateChangedSignal_Expecter struct {
	mock *mock.Mock
}

func (_m *BoardStateChangedSignal) EXPECT() *BoardStateChangedSignal_Expecter {
	return &BoardStateChangedSignal_Expecter{mock: &_m.Mock}
}

// AddListener provides a mock function with given fields: handler, key
func (_m *BoardStateChangedSignal) AddListener(handler signals.SignalListener[logicsignals.BoardStateChangedData], key ...string) int {
	_va := make([]interface{}, len(key))
	for _i := range key {
		_va[_i] = key[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, handler)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for AddListener")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func(signals.SignalListener[logicsignals.BoardStateChangedData], ...string) int); ok {
		r0 = rf(handler, key...)
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// BoardStateChangedSignal_AddListener_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddListener'
type BoardStateChangedSignal_AddListener_Call struct {
	*mock.Call
}

// AddListener is a helper method to define mock.On call
//   - handler signals.SignalListener[logicsignals.BoardStateChangedData]
//   - key ...string
func (_e *BoardStateChangedSignal_Expecter) AddListener(handler interface{}, key ...interface{}) *BoardStateChangedSignal_AddListener_Call {
	return &BoardStateChangedSignal_AddListener_Call{Call: _e.mock.On("AddListener",
		append([]interface{}{handler}, key...)...)}
}

func (_c *BoardStateChangedSignal_AddListener_Call) Run(run func(handler signals.SignalListener[logicsignals.BoardStateChangedData], key ...string)) *BoardStateChangedSignal_AddListener_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(signals.SignalListener[logicsignals.BoardStateChangedData]), variadicArgs...)
	})
	return _c
}

func (_c *BoardStateChangedSignal_AddListener_Call) Return(_a0 int) *BoardStateChangedSignal_AddListener_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BoardStateChangedSignal_AddListener_Call) RunAndReturn(run func(signals.SignalListener[logicsignals.BoardStateChangedData], ...string) int) *BoardStateChangedSignal_AddListener_Call {
	_c.Call.Return(run)
	return _c
}

// Emit provides a mock function with given fields: ctx, payload
func (_m *BoardStateChangedSignal) Emit(ctx context.Context, payload logicsignals.BoardStateChangedData) {
	_m.Called(ctx, payload)
}

// BoardStateChangedSignal_Emit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Emit'
type BoardStateChangedSignal_Emit_Call struct {
	*mock.Call
}

// Emit is a helper method to define mock.On call
//   - ctx context.Context
//   - payload logicsignals.BoardStateChangedData
func (_e *BoardStateChangedSignal_Expecter) Emit(ctx interface{}, payload interface{}) *BoardStateChangedSignal_Emit_Call {
	return &BoardStateChangedSignal_Emit_Call{Call: _e.mock.On("Emit", ctx, payload)}
}

func (_c *BoardStateChangedSignal_Emit_Call) Run(run func(ctx context.Context, payload logicsignals.BoardStateChangedData)) *BoardStateChangedSignal_Emit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(logicsignals.BoardStateChangedData))
	})
	return _c
}

func (_c *BoardStateChangedSignal_Emit_Call) Return() *BoardStateChangedSignal_Emit_Call {
	_c.Call.Return()
	return _c
}

func (_c *BoardStateChangedSignal_Emit_Call) RunAndReturn(run func(context.Context, logicsignals.BoardStateChangedData)) *BoardStateChangedSignal_Emit_Call {
	_c.Call.Return(run)
	return _c
}

// IsEmpty provides a mock function with given fields:
func (_m *BoardStateChangedSignal) IsEmpty() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsEmpty")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// BoardStateChangedSignal_IsEmpty_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsEmpty'
type BoardStateChangedSignal_IsEmpty_Call struct {
	*mock.Call
}

// IsEmpty is a helper method to define mock.On call
func (_e *BoardStateChangedSignal_Expecter) IsEmpty() *BoardStateChangedSignal_IsEmpty_Call {
	return &BoardStateChangedSignal_IsEmpty_Call{Call: _e.mock.On("IsEmpty")}
}

func (_c *BoardStateChangedSignal_IsEmpty_Call) Run(run func()) *BoardStateChangedSignal_IsEmpty_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BoardStateChangedSignal_IsEmpty_Call) Return(_a0 bool) *BoardStateChangedSignal_IsEmpty_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BoardStateChangedSignal_IsEmpty_Call) RunAndReturn(run func() bool) *BoardStateChangedSignal_IsEmpty_Call {
	_c.Call.Return(run)
	return _c
}

// Len provides a mock function with given fields:
func (_m *BoardStateChangedSignal) Len() int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Len")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// BoardStateChangedSignal_Len_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Len'
type BoardStateChangedSignal_Len_Call struct {
	*mock.Call
}

// Len is a helper method to define mock.On call
func (_e *BoardStateChangedSignal_Expecter) Len() *BoardStateChangedSignal_Len_Call {
	return &BoardStateChangedSignal_Len_Call{Call: _e.mock.On("Len")}
}

func (_c *BoardStateChangedSignal_Len_Call) Run(run func()) *BoardStateChangedSignal_Len_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BoardStateChangedSignal_Len_Call) Return(_a0 int) *BoardStateChangedSignal_Len_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BoardStateChangedSignal_Len_Call) RunAndReturn(run func() int) *BoardStateChangedSignal_Len_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveListener provides a mock function with given fields: key
func (_m *BoardStateChangedSignal) RemoveListener(key string) int {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for RemoveListener")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func(string) int); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// BoardStateChangedSignal_RemoveListener_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveListener'
type BoardStateChangedSignal_RemoveListener_Call struct {
	*mock.Call
}

// RemoveListener is a helper method to define mock.On call
//   - key string
func (_e *BoardStateChangedSignal_Expecter) RemoveListener(key interface{}) *BoardStateChangedSignal_RemoveListener_Call {
	return &BoardStateChangedSignal_RemoveListener_Call{Call: _e.mock.On("RemoveListener", key)}
}

func (_c *BoardStateChangedSignal_RemoveListener_Call) Run(run func(key string)) *BoardStateChangedSignal_RemoveListener_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *BoardStateChangedSignal_RemoveListener_Call) Return(_a0 int) *BoardStateChangedSignal_RemoveListener_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BoardStateChangedSignal_RemoveListener_Call) RunAndReturn(run func(string) int) *BoardStateChangedSignal_RemoveListener_Call {
	_c.Call.Return(run)
	return _c
}

// Reset provides a mock function with given fields:
func (_m *BoardStateChangedSignal) Reset() {
	_m.Called()
}

// BoardStateChangedSignal_Reset_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reset'
type BoardStateChangedSignal_Reset_Call struct {
	*mock.Call
}

// Reset is a helper method to define mock.On call
func (_e *BoardStateChangedSignal_Expecter) Reset() *BoardStateChangedSignal_Reset_Call {
	return &BoardStateChangedSignal_Reset_Call{Call: _e.mock.On("Reset")}
}

func (_c *BoardStateChangedSignal_Reset_Call) Run(run func()) *BoardStateChangedSignal_Reset_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BoardStateChangedSignal_Reset_Call) Return() *BoardStateChangedSignal_Reset_Call {
	_c.Call.Return()
	return _c
}

func (_c *BoardStateChangedSignal_Reset_Call) RunAndReturn(run func()) *BoardStateChangedSignal_Reset_Call {
	_c.Call.Return(run)
	return _c
}

// NewBoardStateChangedSignal creates a new instance of BoardStateChangedSignal. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBoardStateChangedSignal(t interface {
	mock.TestingT
	Cleanup(func())
}) *BoardStateChangedSignal {
	mock := &BoardStateChangedSignal{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
