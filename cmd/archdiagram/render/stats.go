package render

import (
	"fmt"
	"strings"
)

// RenderStats produces the Engineering Highlights bullet list with live counts injected.
// The output replaces the GEN:STATS block in README.md.
func RenderStats(ruleCount, invariantCount int) string {
	var buf strings.Builder

	fmt.Fprintf(&buf,
		"- **Living Architecture** -- Every package has a [`doc.go`](docs/doc-go-spec.md) "+
			"with structured sections (`# Layer`, `# Key Types`, `# Dependencies`). "+
			"CI validates layer claims against the actual import graph. "+
			"[%d architecture rules](internal/) enforce boundaries at test time "+
			"-- import constraints, quality ratchets, and documentation coverage.",
		ruleCount,
	)
	buf.WriteString("\n\n")

	fmt.Fprintf(
		&buf,
		"- **Property-Based Testing** -- A custom "+
			"[invariant framework](docs/testing-philosophy.md) "+
			"simulates thousands of randomized games against a real "+
			"Postgres database, checking "+
			"[%d game-state invariants]"+
			"(internal/testing/invariant/doc.go) after every move. "+
			"Failures auto-shrink to minimal reproducers via "+
			"[rapid](https://pkg.go.dev/pgregory.net/rapid).",
		invariantCount,
	)
	buf.WriteString("\n\n")

	buf.WriteString(
		"- **Event-Driven Architecture** -- A custom [EventBus](internal/kernel/bus/doc.go) " +
			"with typed subscriptions, OTel linked spans, and sequential per-event dispatch. " +
			"Zero external dependencies. Decouples game logic from WebSocket broadcasting.",
	)
	buf.WriteString("\n\n")

	buf.WriteString(
		"- **Type-Safe Move Pipeline** -- Generic `Service[T, R]` interface with compile-time " +
			"enforcement across 5 move types. The [orchestration pipeline]" +
			"(internal/game/logic/move/orchestration/doc.go) runs validate, perform, log, " +
			"check mission, and advance in a single transaction.",
	)
	buf.WriteString("\n\n")

	buf.WriteString(
		"- **Auto-Generated Docs** -- A [Go tool](cmd/archdiagram/) reads the import graph " +
			"and `doc.go` labels to generate 5 outputs: " +
			"[D2 architecture diagram](docs/architecture-diagram.svg), " +
			"[architecture.md](docs/architecture.md), " +
			"[architecture-components.md](docs/architecture-components.md), " +
			"[doc-go-spec.md](docs/doc-go-spec.md), and the project tree below. " +
			"CI checks freshness via `make docs-check`.",
	)

	return buf.String()
}
