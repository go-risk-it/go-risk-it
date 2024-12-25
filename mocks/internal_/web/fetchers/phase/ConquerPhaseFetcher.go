// Code generated by mockery v2.50.1. DO NOT EDIT.

package phase

import (
	json "encoding/json"

	ctx "github.com/go-risk-it/go-risk-it/internal/ctx"

	mock "github.com/stretchr/testify/mock"

	state "github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

// ConquerPhaseFetcher is an autogenerated mock type for the ConquerPhaseFetcher type
type ConquerPhaseFetcher struct {
	mock.Mock
}

type ConquerPhaseFetcher_Expecter struct {
	mock *mock.Mock
}

func (_m *ConquerPhaseFetcher) EXPECT() *ConquerPhaseFetcher_Expecter {
	return &ConquerPhaseFetcher_Expecter{mock: &_m.Mock}
}

// FetchState provides a mock function with given fields: _a0, game, stateChannel
func (_m *ConquerPhaseFetcher) FetchState(_a0 ctx.GameContext, game *state.Game, stateChannel chan json.RawMessage) {
	_m.Called(_a0, game, stateChannel)
}

// ConquerPhaseFetcher_FetchState_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FetchState'
type ConquerPhaseFetcher_FetchState_Call struct {
	*mock.Call
}

// FetchState is a helper method to define mock.On call
//   - _a0 ctx.GameContext
//   - game *state.Game
//   - stateChannel chan json.RawMessage
func (_e *ConquerPhaseFetcher_Expecter) FetchState(_a0 interface{}, game interface{}, stateChannel interface{}) *ConquerPhaseFetcher_FetchState_Call {
	return &ConquerPhaseFetcher_FetchState_Call{Call: _e.mock.On("FetchState", _a0, game, stateChannel)}
}

func (_c *ConquerPhaseFetcher_FetchState_Call) Run(run func(_a0 ctx.GameContext, game *state.Game, stateChannel chan json.RawMessage)) *ConquerPhaseFetcher_FetchState_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.GameContext), args[1].(*state.Game), args[2].(chan json.RawMessage))
	})
	return _c
}

func (_c *ConquerPhaseFetcher_FetchState_Call) Return() *ConquerPhaseFetcher_FetchState_Call {
	_c.Call.Return()
	return _c
}

func (_c *ConquerPhaseFetcher_FetchState_Call) RunAndReturn(run func(ctx.GameContext, *state.Game, chan json.RawMessage)) *ConquerPhaseFetcher_FetchState_Call {
	_c.Run(run)
	return _c
}

// NewConquerPhaseFetcher creates a new instance of ConquerPhaseFetcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewConquerPhaseFetcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *ConquerPhaseFetcher {
	mock := &ConquerPhaseFetcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
