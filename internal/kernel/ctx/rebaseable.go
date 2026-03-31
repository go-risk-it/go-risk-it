package ctx

import (
	"context"
)

// Rebaseable is implemented by context types that can copy their domain metadata
// (e.g., GameID, LobbyID, UserID) onto a provided base context. Rebase returns a
// context.Context so that callers (such as the observe package and event bus) can
// work with contexts generically without knowing the concrete domain type. The
// caller owns timeout creation and cancellation — Rebase is metadata-only.
type Rebaseable interface {
	context.Context
	Rebase(base context.Context) context.Context
}
