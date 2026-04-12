// Package snapshot defines the clean domain types for game state snapshots.
//
// These types represent the public contract for game state — they are what
// consumers (WebSocket publishers, REST controllers, read-model projectors)
// receive after the data layer has been queried and mapped. They contain
// zero database or ORM dependencies.
//
// # Sealed interfaces
//
// [Phase] and [PlayerMission] use sealed interfaces ([PhaseState] and
// [MissionDetail]) with unexported marker methods to prevent external
// implementations. This gives runtime polymorphism (unlike Go type-set
// constraints) while keeping the variant set closed. Custom MarshalJSON /
// UnmarshalJSON on the wrapper structs handle discriminator-based routing.
//
// # Wire format
//
// JSON tags use camelCase and match the existing WebSocket wire format
// byte-for-byte. Changing a tag here is a breaking change for all connected
// clients.
//
// # Layer
//
// API — pure value types with JSON serialization. Must not import from
// logic/, data/, sqlc/, or any internal implementation package.
package snapshot
