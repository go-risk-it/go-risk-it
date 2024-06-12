package move

import "context"

type Move[T any] struct {
	UserID  string
	GameID  int64
	Payload T
}

type Performer[T any] interface {
	Perform(
		ctx context.Context,
		move Move[T],
	) error
}
