// Code generated by mockery v2.50.1. DO NOT EDIT.

package phase

import (
	json "encoding/json"

	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"

	mock "github.com/stretchr/testify/mock"

	state "github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

// DeployPhaseFetcher is an autogenerated mock type for the DeployPhaseFetcher type
type DeployPhaseFetcher struct {
	mock.Mock
}

type DeployPhaseFetcher_Expecter struct {
	mock *mock.Mock
}

func (_m *DeployPhaseFetcher) EXPECT() *DeployPhaseFetcher_Expecter {
	return &DeployPhaseFetcher_Expecter{mock: &_m.Mock}
}

// FetchState provides a mock function with given fields: _a0, game, stateChannel
func (_m *DeployPhaseFetcher) FetchState(_a0 ctx.GameContext, game *state.Game, stateChannel chan json.RawMessage) {
	_m.Called(_a0, game, stateChannel)
}

// DeployPhaseFetcher_FetchState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FetchState'
type DeployPhaseFetcher_FetchState_Call struct {
	*mock.Call
}

// FetchState is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - game *state.Game
//   - stateChannel chan json.RawMessage
func (_e *DeployPhaseFetcher_Expecter) FetchState(_a0 interface{}, game interface{}, stateChannel interface{}) *DeployPhaseFetcher_FetchState_Call {
	return &DeployPhaseFetcher_FetchState_Call{Call: _e.mock.On("FetchState", _a0, game, stateChannel)}
}

func (_c *DeployPhaseFetcher_FetchState_Call) Run(run func(_a0 ctx.GameContext, game *state.Game, stateChannel chan json.RawMessage)) *DeployPhaseFetcher_FetchState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(*state.Game), args[2].(chan json.RawMessage))
	})
	return _c
}

func (_c *DeployPhaseFetcher_FetchState_Call) Return() *DeployPhaseFetcher_FetchState_Call {
	_c.Call.Return()
	return _c
}

func (_c *DeployPhaseFetcher_FetchState_Call) RunAndReturn(run func(ctx.GameContext, *state.Game, chan json.RawMessage)) *DeployPhaseFetcher_FetchState_Call {
	_c.Run(run)
	return _c
}

// NewDeployPhaseFetcher creates a new instance of DeployPhaseFetcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDeployPhaseFetcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *DeployPhaseFetcher {
	mock := &DeployPhaseFetcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
