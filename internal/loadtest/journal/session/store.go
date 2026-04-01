package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Store manages session files on disk.
type Store struct {
	dir string
}

// NewStore creates a session store rooted at the given directory.
func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// GetOrCreate returns an existing active session for the branch, or creates a new one.
// If creating, it uses the latest entry in baselineDir as the baseline.
func (s *Store) GetOrCreate(branch, baselineDir string) (*Session, error) {
	existing, err := s.Load(branch)
	if err == nil && existing.Status == "active" {
		return existing, nil
	}

	// Find latest baseline entry.
	baselinePath, err := findLatestEntry(baselineDir)
	if err != nil {
		return nil, fmt.Errorf("find baseline: %w", err)
	}

	sess := &Session{
		ID:            sanitizeBranch(branch),
		Branch:        branch,
		BaselineEntry: baselinePath,
		StartedAt:     time.Now(),
		Status:        "active",
	}

	if err := s.save(sess); err != nil {
		return nil, err
	}

	return sess, nil
}

// AddRun appends a run reference to the session and saves.
func (s *Store) AddRun(branch string, ref RunRef) error {
	sess, err := s.Load(branch)
	if err != nil {
		return fmt.Errorf("load session: %w", err)
	}

	sess.Runs = append(sess.Runs, ref)

	return s.save(sess)
}

// Close marks a session with the given status ("merged" or "abandoned").
func (s *Store) Close(branch, status string) error {
	sess, err := s.Load(branch)
	if err != nil {
		return fmt.Errorf("load session: %w", err)
	}

	sess.Status = status

	return s.save(sess)
}

// Load reads a session file from disk.
func (s *Store) Load(branch string) (*Session, error) {
	path := s.path(branch)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read session %q: %w", path, err)
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("parse session %q: %w", path, err)
	}

	return &sess, nil
}

func (s *Store) path(branch string) string {
	return filepath.Join(s.dir, sanitizeBranch(branch)+".json")
}

func (s *Store) save(sess *Session) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("create session dir: %w", err)
	}

	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	path := s.path(sess.Branch)

	return os.WriteFile(path, data, 0o600) //nolint:wrapcheck // internal error
}

// sanitizeBranch converts a branch name to a safe filename.
func sanitizeBranch(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

// findLatestEntry returns the path to the newest entry file in the directory.
func findLatestEntry(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read dir %q: %w", dir, err)
	}

	var latest string

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		latest = filepath.Join(dir, e.Name())
	}

	if latest == "" {
		return "", fmt.Errorf("no entries found in %q", dir)
	}

	return latest, nil
}
