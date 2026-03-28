// Package checker provides the mission-checking framework. Each mission type
// implements [MissionChecker], and a [Registry] maps mission types to their
// checkers for dispatch during win-condition evaluation.
//
// # Layer
//
// Logic — mission win-condition checking.
package checker
