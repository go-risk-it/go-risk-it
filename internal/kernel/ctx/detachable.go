package ctx

import (
	"context"
)

// Detachable is implemented by context types that can copy their domain metadata
// (e.g., GameID, LobbyID, UserID) onto a provided base context. This allows the
// event bus to detach contexts without type-switching on concrete context types.
// The caller owns timeout creation and cancellation — DetachOnto is metadata-only.
type Detachable interface {
	DetachOnto(base context.Context) context.Context
}
