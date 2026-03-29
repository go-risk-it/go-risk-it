// Package kernel groups domain-agnostic infrastructure packages that
// every layer may depend on: context types, event bus, configuration,
// database primitives, observability, and randomness.
//
// Kernel packages have zero imports of logic/, web/, data/game/, data/lobby/,
// events/game/, or events/lobby/ — they sit below the entire domain stack.
//
// # Layer
//
// Kernel — foundational infrastructure with no domain dependencies.
package kernel
