// Package deploy implements the deploy move: a player places available
// troops onto regions they own. Each move validates region ownership,
// declared troop counts, and deployable troop availability. The phase
// transitions to ATTACK once all deployable troops are placed.
//
// # Layer
//
// Logic — troop deployment validation and execution.
//
// # Key Types
//
// [Service] extends [service.Service] with [Service.GetDeployableTroops]
// for querying how many troops remain to be placed in the current deploy
// phase.
//
// [Move] is the input DTO carrying the target region, current troop count,
// and desired troop count.
//
// # Dependencies
//
//   - [phase.Service] for inserting the attack phase on advancement
//   - [region.Service] for region ownership checks and troop updates
package deploy
