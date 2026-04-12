package snapshot

import "go.uber.org/fx"

// Module provides the snapshot Service, Reader, and ReaderFactory for
// dependency injection.
var Module = fx.Options(
	fx.Provide(NewService),
	fx.Provide(NewReader),
	fx.Provide(NewReaderFactory),
)
