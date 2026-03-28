// Package attack implements the attack move: a player sends troops from an
// owned region to an adjacent enemy region. Dice rolls determine casualties,
// and the phase walker decides whether the game transitions to CONQUER (if a
// region was conquered), REINFORCE (if the player chooses to stop or has no
// eligible attackers), or stays in ATTACK.
//
// # Layer
//
// Logic — attack move validation, dice combat, and phase transitions.
//
// # Key Types
//
// [Service] extends [service.Service] with attack-specific queries:
// [Service.HasConquered] detects zero-troop enemy regions and
// [Service.CanContinueAttacking] checks if the player has eligible attackers.
//
// [Move] is the input DTO carrying attacking/defending region IDs, declared
// troop counts, and the number of attacking troops.
//
// [MoveResult] is the output carrying the conquering troop count, used by
// the orchestrator to create a conquer phase when applicable.
//
// # Dependencies
//
//   - [board.Service] for adjacency checks between attacker and defender
//   - [dice.Service] for rolling attacking and defending dice
//   - [phase.Service] for inserting new phases on advancement
//   - [region.Service] for reading and updating region troop counts
package attack
