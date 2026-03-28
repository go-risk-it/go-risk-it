// Package region manages game regions: creation with player assignment,
// querying by game or player, troop updates, and ownership transfers during
// conquest. Region assignment strategy (random or sequential) is injected
// via the assignment sub-package.
//
// # Layer
//
// Logic — region lifecycle and troop management.
//
// # Key Types
//
// [Service] is the primary interface for region operations: creation with
// player assignment, querying by game or player, troop updates, and
// ownership transfers.
//
// # Dependencies
//
//   - [assignment.Service] for the region-to-player assignment strategy
package region
