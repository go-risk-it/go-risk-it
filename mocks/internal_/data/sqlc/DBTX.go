// Code generated by mockery v2.46.3. DO NOT EDIT.

package sqlc

import (
	context "context"

	pgconn "github.com/jackc/pgx/v5/pgconn"
	mock "github.com/stretchr/testify/mock"

	pgx "github.com/jackc/pgx/v5"
)

// DBTX is an autogenerated mock type for the DBTX type
type DBTX struct {
	mock.Mock
}

type DBTX_Expecter struct {
	mock *mock.Mock
}

func (_m *DBTX) EXPECT() *DBTX_Expecter {
	return &DBTX_Expecter{mock: &_m.Mock}
}

// CopyFrom provides a mock function with given fields: ctx, tableName, columnNames, rowSrc
func (_m *DBTX) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	ret := _m.Called(ctx, tableName, columnNames, rowSrc)

	if len(ret) == 0 {
		panic("no return value specified for CopyFrom")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)); ok {
		return rf(ctx, tableName, columnNames, rowSrc)
	}
	if rf, ok := ret.Get(0).(func(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) int64); ok {
		r0 = rf(ctx, tableName, columnNames, rowSrc)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) error); ok {
		r1 = rf(ctx, tableName, columnNames, rowSrc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DBTX_CopyFrom_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CopyFrom'
type DBTX_CopyFrom_Call struct {
	*mock.Call
}

// CopyFrom is a helper method to define mock.On call
//   - ctx context.Context
//   - tableName pgx.Identifier
//   - columnNames []string
//   - rowSrc pgx.CopyFromSource
func (_e *DBTX_Expecter) CopyFrom(ctx interface{}, tableName interface{}, columnNames interface{}, rowSrc interface{}) *DBTX_CopyFrom_Call {
	return &DBTX_CopyFrom_Call{Call: _e.mock.On("CopyFrom", ctx, tableName, columnNames, rowSrc)}
}

func (_c *DBTX_CopyFrom_Call) Run(run func(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource)) *DBTX_CopyFrom_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(pgx.Identifier), args[2].([]string), args[3].(pgx.CopyFromSource))
	})
	return _c
}

func (_c *DBTX_CopyFrom_Call) Return(_a0 int64, _a1 error) *DBTX_CopyFrom_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DBTX_CopyFrom_Call) RunAndReturn(run func(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)) *DBTX_CopyFrom_Call {
	_c.Call.Return(run)
	return _c
}

// Exec provides a mock function with given fields: _a0, _a1, _a2
func (_m *DBTX) Exec(_a0 context.Context, _a1 string, _a2 ...interface{}) (pgconn.CommandTag, error) {
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _a2...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Exec")
	}

	var r0 pgconn.CommandTag
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) (pgconn.CommandTag, error)); ok {
		return rf(_a0, _a1, _a2...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) pgconn.CommandTag); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		r0 = ret.Get(0).(pgconn.CommandTag)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...interface{}) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DBTX_Exec_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Exec'
type DBTX_Exec_Call struct {
	*mock.Call
}

// Exec is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
//   - _a2 ...interface{}
func (_e *DBTX_Expecter) Exec(_a0 interface{}, _a1 interface{}, _a2 ...interface{}) *DBTX_Exec_Call {
	return &DBTX_Exec_Call{Call: _e.mock.On("Exec",
		append([]interface{}{_a0, _a1}, _a2...)...)}
}

func (_c *DBTX_Exec_Call) Run(run func(_a0 context.Context, _a1 string, _a2 ...interface{})) *DBTX_Exec_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *DBTX_Exec_Call) Return(_a0 pgconn.CommandTag, _a1 error) *DBTX_Exec_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DBTX_Exec_Call) RunAndReturn(run func(context.Context, string, ...interface{}) (pgconn.CommandTag, error)) *DBTX_Exec_Call {
	_c.Call.Return(run)
	return _c
}

// Query provides a mock function with given fields: _a0, _a1, _a2
func (_m *DBTX) Query(_a0 context.Context, _a1 string, _a2 ...interface{}) (pgx.Rows, error) {
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _a2...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Query")
	}

	var r0 pgx.Rows
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) (pgx.Rows, error)); ok {
		return rf(_a0, _a1, _a2...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) pgx.Rows); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pgx.Rows)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...interface{}) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DBTX_Query_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Query'
type DBTX_Query_Call struct {
	*mock.Call
}

// Query is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
//   - _a2 ...interface{}
func (_e *DBTX_Expecter) Query(_a0 interface{}, _a1 interface{}, _a2 ...interface{}) *DBTX_Query_Call {
	return &DBTX_Query_Call{Call: _e.mock.On("Query",
		append([]interface{}{_a0, _a1}, _a2...)...)}
}

func (_c *DBTX_Query_Call) Run(run func(_a0 context.Context, _a1 string, _a2 ...interface{})) *DBTX_Query_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *DBTX_Query_Call) Return(_a0 pgx.Rows, _a1 error) *DBTX_Query_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DBTX_Query_Call) RunAndReturn(run func(context.Context, string, ...interface{}) (pgx.Rows, error)) *DBTX_Query_Call {
	_c.Call.Return(run)
	return _c
}

// QueryRow provides a mock function with given fields: _a0, _a1, _a2
func (_m *DBTX) QueryRow(_a0 context.Context, _a1 string, _a2 ...interface{}) pgx.Row {
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _a2...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for QueryRow")
	}

	var r0 pgx.Row
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) pgx.Row); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pgx.Row)
		}
	}

	return r0
}

// DBTX_QueryRow_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'QueryRow'
type DBTX_QueryRow_Call struct {
	*mock.Call
}

// QueryRow is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
//   - _a2 ...interface{}
func (_e *DBTX_Expecter) QueryRow(_a0 interface{}, _a1 interface{}, _a2 ...interface{}) *DBTX_QueryRow_Call {
	return &DBTX_QueryRow_Call{Call: _e.mock.On("QueryRow",
		append([]interface{}{_a0, _a1}, _a2...)...)}
}

func (_c *DBTX_QueryRow_Call) Run(run func(_a0 context.Context, _a1 string, _a2 ...interface{})) *DBTX_QueryRow_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *DBTX_QueryRow_Call) Return(_a0 pgx.Row) *DBTX_QueryRow_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DBTX_QueryRow_Call) RunAndReturn(run func(context.Context, string, ...interface{}) pgx.Row) *DBTX_QueryRow_Call {
	_c.Call.Return(run)
	return _c
}

// NewDBTX creates a new instance of DBTX. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDBTX(t interface {
	mock.TestingT
	Cleanup(func())
}) *DBTX {
	mock := &DBTX{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
