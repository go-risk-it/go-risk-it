package ctx

import (
	"context"
	"time"
)

// Detachable is implemented by context types that can create a copy of themselves
// detached from the parent's cancellation chain, preserving domain metadata
// (e.g., GameID, LobbyID, UserID) under a fresh timeout. This allows the event
// bus to detach contexts without type-switching on concrete context types.
type Detachable interface {
	Detach(timeout time.Duration) (context.Context, context.CancelFunc)
}
