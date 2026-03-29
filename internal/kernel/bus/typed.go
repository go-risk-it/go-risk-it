package bus

import "context"

// OnEvent registers a typed handler on the bus for events of type E.
// It uses the nil-pointer method call pattern to extract the event type string
// at registration time: `var zero E; zero.EventType()`. This is safe because
// all EventType() methods return a string constant without dereferencing the receiver.
//
// The wrapped handler performs a type assertion from Event to E. If the assertion
// fails (which should not happen when used with OnType routing), the handler is a no-op.
func OnEvent[E Event](bus Bus, handler func(context.Context, E)) {
	var zero E
	eventType := zero.EventType()

	bus.OnType(eventType, func(ctx context.Context, event Event) {
		typed, ok := event.(E)
		if !ok {
			return
		}

		handler(ctx, typed)
	})
}
