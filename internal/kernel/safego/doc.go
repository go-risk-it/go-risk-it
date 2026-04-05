// Package safego provides panic-safe operation wrappers with OTel span
// integration.
//
// [SafeOp] runs an action inside a child span. On panic it records the error
// on the span and logs the recovered value and stack trace — the panic does
// not propagate. [TypedSafeOp] adds typed context propagation via the
// [ctx.Rebaseable] contract so callers receive their domain context type
// directly without type assertions.
//
// These wrappers are sequential (not goroutine-spawning). The event bus uses
// them inside its own goroutine lifecycle; other packages may use them for
// any panic-critical section that needs observability.
//
// # Layer
//
// Kernel — panic-safe operation wrappers with no logic dependencies.
package safego
