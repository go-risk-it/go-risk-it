// Code generated by mockery v2.50.1. DO NOT EDIT.

package signals

import (
	context "context"

	gamesignals "github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
	mock "github.com/stretchr/testify/mock"

	signals "github.com/maniartech/signals"
)

// MovePerformedSignal is an autogenerated mock type for the MovePerformedSignal type
type MovePerformedSignal struct {
	mock.Mock
}

type MovePerformedSignal_Expecter struct {
	mock *mock.Mock
}

func (_m *MovePerformedSignal) EXPECT() *MovePerformedSignal_Expecter {
	return &MovePerformedSignal_Expecter{mock: &_m.Mock}
}

// AddListener provides a mock function with given fields: handler, key
func (_m *MovePerformedSignal) AddListener(handler signals.SignalListener[gamesignals.MovePerformedData], key ...string) int {
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
	if rf, ok := ret.Get(0).(func(signals.SignalListener[gamesignals.MovePerformedData], ...string) int); ok {
		r0 = rf(handler, key...)
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// MovePerformedSignal_AddListener_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddListener'
type MovePerformedSignal_AddListener_Call struct {
	*mock.Call
}

// AddListener is a helper method to define mock.On call
//   - handler signals.SignalListener[gamesignals.MovePerformedData]
//   - key ...string
func (_e *MovePerformedSignal_Expecter) AddListener(handler interface{}, key ...interface{}) *MovePerformedSignal_AddListener_Call {
	return &MovePerformedSignal_AddListener_Call{Call: _e.mock.On("AddListener",
		append([]interface{}{handler}, key...)...)}
}

func (_c *MovePerformedSignal_AddListener_Call) Run(run func(handler signals.SignalListener[gamesignals.MovePerformedData], key ...string)) *MovePerformedSignal_AddListener_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(signals.SignalListener[gamesignals.MovePerformedData]), variadicArgs...)
	})
	return _c
}

func (_c *MovePerformedSignal_AddListener_Call) Return(_a0 int) *MovePerformedSignal_AddListener_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MovePerformedSignal_AddListener_Call) RunAndReturn(run func(signals.SignalListener[gamesignals.MovePerformedData], ...string) int) *MovePerformedSignal_AddListener_Call {
	_c.Call.Return(run)
	return _c
}

// Emit provides a mock function with given fields: ctx, payload
func (_m *MovePerformedSignal) Emit(ctx context.Context, payload gamesignals.MovePerformedData) {
	_m.Called(ctx, payload)
}

// MovePerformedSignal_Emit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Emit'
type MovePerformedSignal_Emit_Call struct {
	*mock.Call
}

// Emit is a helper method to define mock.On call
//   - ctx context.Context
//   - payload gamesignals.MovePerformedData
func (_e *MovePerformedSignal_Expecter) Emit(ctx interface{}, payload interface{}) *MovePerformedSignal_Emit_Call {
	return &MovePerformedSignal_Emit_Call{Call: _e.mock.On("Emit", ctx, payload)}
}

func (_c *MovePerformedSignal_Emit_Call) Run(run func(ctx context.Context, payload gamesignals.MovePerformedData)) *MovePerformedSignal_Emit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(gamesignals.MovePerformedData))
	})
	return _c
}

func (_c *MovePerformedSignal_Emit_Call) Return() *MovePerformedSignal_Emit_Call {
	_c.Call.Return()
	return _c
}

func (_c *MovePerformedSignal_Emit_Call) RunAndReturn(run func(context.Context, gamesignals.MovePerformedData)) *MovePerformedSignal_Emit_Call {
	_c.Run(run)
	return _c
}

// IsEmpty provides a mock function with no fields
func (_m *MovePerformedSignal) IsEmpty() bool {
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

// MovePerformedSignal_IsEmpty_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsEmpty'
type MovePerformedSignal_IsEmpty_Call struct {
	*mock.Call
}

// IsEmpty is a helper method to define mock.On call
func (_e *MovePerformedSignal_Expecter) IsEmpty() *MovePerformedSignal_IsEmpty_Call {
	return &MovePerformedSignal_IsEmpty_Call{Call: _e.mock.On("IsEmpty")}
}

func (_c *MovePerformedSignal_IsEmpty_Call) Run(run func()) *MovePerformedSignal_IsEmpty_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MovePerformedSignal_IsEmpty_Call) Return(_a0 bool) *MovePerformedSignal_IsEmpty_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MovePerformedSignal_IsEmpty_Call) RunAndReturn(run func() bool) *MovePerformedSignal_IsEmpty_Call {
	_c.Call.Return(run)
	return _c
}

// Len provides a mock function with no fields
func (_m *MovePerformedSignal) Len() int {
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

// MovePerformedSignal_Len_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Len'
type MovePerformedSignal_Len_Call struct {
	*mock.Call
}

// Len is a helper method to define mock.On call
func (_e *MovePerformedSignal_Expecter) Len() *MovePerformedSignal_Len_Call {
	return &MovePerformedSignal_Len_Call{Call: _e.mock.On("Len")}
}

func (_c *MovePerformedSignal_Len_Call) Run(run func()) *MovePerformedSignal_Len_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MovePerformedSignal_Len_Call) Return(_a0 int) *MovePerformedSignal_Len_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MovePerformedSignal_Len_Call) RunAndReturn(run func() int) *MovePerformedSignal_Len_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveListener provides a mock function with given fields: key
func (_m *MovePerformedSignal) RemoveListener(key string) int {
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

// MovePerformedSignal_RemoveListener_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveListener'
type MovePerformedSignal_RemoveListener_Call struct {
	*mock.Call
}

// RemoveListener is a helper method to define mock.On call
//   - key string
func (_e *MovePerformedSignal_Expecter) RemoveListener(key interface{}) *MovePerformedSignal_RemoveListener_Call {
	return &MovePerformedSignal_RemoveListener_Call{Call: _e.mock.On("RemoveListener", key)}
}

func (_c *MovePerformedSignal_RemoveListener_Call) Run(run func(key string)) *MovePerformedSignal_RemoveListener_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MovePerformedSignal_RemoveListener_Call) Return(_a0 int) *MovePerformedSignal_RemoveListener_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MovePerformedSignal_RemoveListener_Call) RunAndReturn(run func(string) int) *MovePerformedSignal_RemoveListener_Call {
	_c.Call.Return(run)
	return _c
}

// Reset provides a mock function with no fields
func (_m *MovePerformedSignal) Reset() {
	_m.Called()
}

// MovePerformedSignal_Reset_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reset'
type MovePerformedSignal_Reset_Call struct {
	*mock.Call
}

// Reset is a helper method to define mock.On call
func (_e *MovePerformedSignal_Expecter) Reset() *MovePerformedSignal_Reset_Call {
	return &MovePerformedSignal_Reset_Call{Call: _e.mock.On("Reset")}
}

func (_c *MovePerformedSignal_Reset_Call) Run(run func()) *MovePerformedSignal_Reset_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MovePerformedSignal_Reset_Call) Return() *MovePerformedSignal_Reset_Call {
	_c.Call.Return()
	return _c
}

func (_c *MovePerformedSignal_Reset_Call) RunAndReturn(run func()) *MovePerformedSignal_Reset_Call {
	_c.Run(run)
	return _c
}

// NewMovePerformedSignal creates a new instance of MovePerformedSignal. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMovePerformedSignal(t interface {
	mock.TestingT
	Cleanup(func())
}) *MovePerformedSignal {
	mock := &MovePerformedSignal{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
