package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/journal"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/journal/session"
	"go.opentelemetry.io/otel/attribute"
)

// saveJournalEntry saves the journal entry and handles session integration.
func saveJournalEntry(
	entry journal.Entry,
	slug string,
	branch, commitSHA, hypothesis string,
) {
	ctx := context.Background()

	path, err := journal.SaveEntry("perf-journal/entries", slug, entry)
	if err != nil {
		observe.Error(ctx, err, "failed to save journal entry")

		return
	}

	observe.Info(ctx, "journal entry saved", attribute.String("path", path))

	if branch != "" && branch != "main" && branch != "master" {
		handleSession(branch, commitSHA, path, entry.SLOCeiling.Games, hypothesis)
	}
}

// compareJournalEntries prints a comparison between two journal entries.
func compareJournalEntries(compareFile string, entry journal.Entry) error {
	prevEntry, err := journal.LoadEntry(compareFile)
	if err != nil {
		return fmt.Errorf("load journal entry for comparison: %w", err)
	}

	journal.PrintCeilingComparison(os.Stdout, prevEntry, entry)

	return nil
}

// handleSession manages session state for branch-scoped optimization tracking.
func handleSession(branch, commitSHA, entryPath string, ceilingGames int, hypothesis string) {
	ctx := context.Background()
	store := session.NewStore("perf-journal/sessions")

	sess, err := store.GetOrCreate(branch, "perf-journal/entries")
	if err != nil {
		observe.Error(ctx, err, "failed to create/load session")

		return
	}

	// Compute delta vs session baseline ceiling.
	baselineCeiling := 0
	if len(sess.Runs) > 0 {
		baselineCeiling = sess.Runs[0].CeilingGames
	}

	ref := session.RunRef{
		EntryPath:    entryPath,
		CommitSHA:    commitSHA,
		CeilingGames: ceilingGames,
		CeilingDelta: ceilingGames - baselineCeiling,
		Hypothesis:   hypothesis,
		Timestamp:    time.Now(),
	}

	if err := store.AddRun(branch, ref); err != nil {
		observe.Error(ctx, err, "failed to add session run")

		return
	}

	runNum := len(sess.Runs) + 1
	observe.Info(ctx, "session run recorded",
		attribute.String("branch", branch),
		attribute.Int("run_num", runNum),
		attribute.Int("ceiling", ceilingGames),
		attribute.Int("delta", ref.CeilingDelta),
	)

	if hypothesis != "" {
		observe.Info(ctx, "session hypothesis", attribute.String("hypothesis", hypothesis))
	}
}
