package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create a fake baseline entry.
	entriesDir := filepath.Join(dir, "entries")
	if err := os.MkdirAll(entriesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(entriesDir, "000-baseline.json"),
		[]byte(`{}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestStore_NewSession(t *testing.T) {
	dir := setupTestDir(t)
	store := NewStore(filepath.Join(dir, "sessions"))
	entriesDir := filepath.Join(dir, "entries")

	sess, err := store.GetOrCreate("feature/optimize-db", entriesDir)
	if err != nil {
		t.Fatalf("GetOrCreate: %v", err)
	}

	if sess.Branch != "feature/optimize-db" {
		t.Errorf("branch: expected feature/optimize-db, got %q", sess.Branch)
	}

	if sess.Status != "active" {
		t.Errorf("status: expected active, got %q", sess.Status)
	}

	if sess.BaselineEntry == "" {
		t.Error("baseline entry should not be empty")
	}

	if len(sess.Runs) != 0 {
		t.Errorf("new session should have 0 runs, got %d", len(sess.Runs))
	}
}

func TestStore_ExistingSession(t *testing.T) {
	dir := setupTestDir(t)
	store := NewStore(filepath.Join(dir, "sessions"))
	entriesDir := filepath.Join(dir, "entries")

	// Create session.
	sess1, err := store.GetOrCreate("feature/optimize", entriesDir)
	if err != nil {
		t.Fatalf("first GetOrCreate: %v", err)
	}

	startedAt := sess1.StartedAt

	// Get again — should return same session.
	sess2, err := store.GetOrCreate("feature/optimize", entriesDir)
	if err != nil {
		t.Fatalf("second GetOrCreate: %v", err)
	}

	if !sess2.StartedAt.Equal(startedAt) {
		t.Error("second call should return same session, got different StartedAt")
	}
}

func TestStore_AddRun(t *testing.T) {
	dir := setupTestDir(t)
	store := NewStore(filepath.Join(dir, "sessions"))
	entriesDir := filepath.Join(dir, "entries")

	_, err := store.GetOrCreate("feature/test", entriesDir)
	if err != nil {
		t.Fatalf("GetOrCreate: %v", err)
	}

	ref := RunRef{
		EntryPath:    "entries/001-test.json",
		CommitSHA:    "abc1234",
		CeilingGames: 40,
		CeilingDelta: 0,
		Hypothesis:   "optimize query batching",
		Timestamp:    time.Now(),
	}

	if err := store.AddRun("feature/test", ref); err != nil {
		t.Fatalf("AddRun: %v", err)
	}

	// Reload and verify.
	sess, err := store.Load("feature/test")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(sess.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(sess.Runs))
	}

	if sess.Runs[0].CeilingGames != 40 {
		t.Errorf("ceiling: expected 40, got %d", sess.Runs[0].CeilingGames)
	}

	if sess.Runs[0].Hypothesis != "optimize query batching" {
		t.Errorf("hypothesis: expected 'optimize query batching', got %q", sess.Runs[0].Hypothesis)
	}
}

func TestStore_Close(t *testing.T) {
	dir := setupTestDir(t)
	store := NewStore(filepath.Join(dir, "sessions"))
	entriesDir := filepath.Join(dir, "entries")

	_, err := store.GetOrCreate("feature/done", entriesDir)
	if err != nil {
		t.Fatalf("GetOrCreate: %v", err)
	}

	if err := store.Close("feature/done", "merged"); err != nil {
		t.Fatalf("Close: %v", err)
	}

	sess, err := store.Load("feature/done")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if sess.Status != "merged" {
		t.Errorf("status: expected merged, got %q", sess.Status)
	}

	// GetOrCreate should create new session since old one is closed.
	sess2, err := store.GetOrCreate("feature/done", entriesDir)
	if err != nil {
		t.Fatalf("GetOrCreate after close: %v", err)
	}

	if sess2.Status != "active" {
		t.Errorf("new session status: expected active, got %q", sess2.Status)
	}
}

func TestStore_NoBaseline(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "sessions"))
	emptyDir := filepath.Join(dir, "empty-entries")

	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := store.GetOrCreate("feature/test", emptyDir)
	if err == nil {
		t.Error("expected error when no baseline entries exist")
	}
}

func TestSession_LastCeiling(t *testing.T) {
	sess := &Session{
		Runs: []RunRef{
			{CeilingGames: 20},
			{CeilingGames: 40},
			{CeilingGames: 35},
		},
	}

	if got := sess.LastCeiling(); got != 35 {
		t.Errorf("LastCeiling: expected 35, got %d", got)
	}

	empty := &Session{}
	if got := empty.LastCeiling(); got != 0 {
		t.Errorf("empty session LastCeiling: expected 0, got %d", got)
	}
}
