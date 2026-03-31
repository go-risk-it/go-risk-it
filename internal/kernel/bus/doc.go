// Package bus provides the domain event bus for decoupled communication.
//
// The bus exposes two directional sub-interfaces: [Publisher] (Emit) for event
// producers and [Subscriber] (OnAll, OnType) for event consumers. The composite
// [Bus] interface is used only at the FX composition root — all other packages
// should depend on Publisher or Subscriber to enforce the directional contract.
//
// Each emitted event spawns one goroutine that runs all matching handlers
// sequentially — OnAll handlers first, then OnType handlers — with per-handler
// panic recovery. Handler contexts are detached from the emitter's cancellation
// chain but carry a linked OTel span and preserved domain metadata (via
// [ctx.Rebaseable]).
//
// [OnEvent] provides type-safe generic subscriptions, extracting the event
// type string at registration time via the nil-pointer method call pattern.
// [TestBus] is a synchronous implementation for unit tests that captures
// emitted events in order without goroutines.
//
// # Layer
//
// Kernel — event bus infrastructure with no logic dependencies.
package bus
