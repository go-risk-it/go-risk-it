// Package rand provides a seedable random number generator for game mechanics.
//
// It wraps math/rand/v2 with a fixed PCG seed to produce deterministic
// sequences for shuffling (region assignment, card dealing) and random
// selection (dice rolls). The [RNG] interface abstracts the generator,
// allowing tests to substitute controlled implementations.
//
// # Layer
//
// Kernel — stateless utility with no internal dependencies.
package rand
