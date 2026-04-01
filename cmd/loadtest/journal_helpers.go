package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/journal"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/journal/session"
)

// saveJournalEntry saves the journal entry and handles session integration.
func saveJournalEntry(
	entry journal.Entry,
	slug string,
	branch, commitSHA, hypothesis string,
) {
	path, err := journal.SaveEntry("perf-journal/entries", slug, entry)
	if err != nil {
		log.Printf("failed to save journal entry: %v", err)

		return
	}

	log.Printf("journal entry saved: %s", path)

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
	store := session.NewStore("perf-journal/sessions")

	sess, err := store.GetOrCreate(branch, "perf-journal/entries")
	if err != nil {
		log.Printf("[session] failed to create/load session: %v", err)

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
		log.Printf("[session] failed to add run: %v", err)

		return
	}

	runNum := len(sess.Runs) + 1
	log.Printf(
		"[session] %s: run #%d, ceiling=%d (delta=%+d from session start)",
		branch, runNum, ceilingGames, ref.CeilingDelta,
	)

	if hypothesis != "" {
		log.Printf("[session] hypothesis: %s", hypothesis)
	}
}
