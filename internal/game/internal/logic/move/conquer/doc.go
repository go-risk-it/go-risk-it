// Package conquer implements the conquer move: after a successful attack
// reduces a defending region to zero troops, the attacker must move troops
// into the conquered region. The move validates minimum troop requirements,
// transfers region ownership, and handles player elimination (card transfer
// and mission reassignment).
//
// # Layer
//
// Logic — conquest troop movement, ownership transfer, and elimination.
//
// # Key Types
//
// [Service] extends [service.Service] with [Service.GetPhaseState] for
// reading the conquer phase constraints (source/target regions and minimum
// troop count) set by the attack advancer.
//
// [Move] is the input DTO carrying only the number of troops to move.
//
// # Dependencies
//
//   - [attack.Service] for checking if the player can continue attacking
//   - [card.Service] for transferring cards from an eliminated player
//   - [mission.Service] for reassigning missions targeting an eliminated player
//   - [phase.Service] for inserting the next phase on advancement
//   - [region.Service] for troop updates and ownership transfer
package conquer
