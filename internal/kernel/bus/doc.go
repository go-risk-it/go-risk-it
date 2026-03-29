// Package bus provides the domain event bus for decoupled communication.
//
// The [Bus] interface dispatches [Event] values to registered [Handler]
// functions. Each emitted event spawns one goroutine that runs all matching
// handlers sequentially — OnAll handlers first, then OnType handlers — with
// per-handler panic recovery. Handler contexts are detached from the
// emitter's cancellation chain but carry a linked OTel span and preserved
// domain metadata (via [ctx.Detachable]).
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
