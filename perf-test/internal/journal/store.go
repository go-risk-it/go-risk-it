package journal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// SaveEntry writes entry as JSON to dir/NNN-slug-commit.json with
// auto-incrementing sequence. Returns the path of the written file.
func SaveEntry(dir, slug string, entry Entry) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	seq, err := nextSequenceNumber(dir)
	if err != nil {
		return "", fmt.Errorf("sequence number: %w", err)
	}

	slug = sanitizeSlug(slug)
	filename := fmt.Sprintf("%03d-%s-%s.json", seq, slug, entry.CommitSHA)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal entry: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write entry: %w", err)
	}

	return path, nil
}

// LoadEntry reads a journal entry from JSON.
func LoadEntry(path string) (Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Entry{}, fmt.Errorf("read entry: %w", err)
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return Entry{}, fmt.Errorf("unmarshal entry: %w", err)
	}

	return entry, nil
}

// ListEntries returns paths of all journal entries in dir, sorted by
// sequence number.
func ListEntries(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("read dir: %w", err)
	}

	re := regexp.MustCompile(`^\d{3}-.*\.json$`)

	var paths []string

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		if re.MatchString(e.Name()) {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}

	sort.Strings(paths)

	return paths, nil
}

// LatestEntry loads the highest-numbered entry in dir.
func LatestEntry(dir string) (Entry, error) {
	paths, err := ListEntries(dir)
	if err != nil {
		return Entry{}, err
	}

	if len(paths) == 0 {
		return Entry{}, fmt.Errorf("no journal entries in %s", dir)
	}

	return LoadEntry(paths[len(paths)-1])
}

// sanitizeSlug normalizes a name for use in filenames: lowercase, hyphens,
// no special chars.
func sanitizeSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

	re := regexp.MustCompile(`[^a-z0-9-]`)
	s = re.ReplaceAllString(s, "")

	re = regexp.MustCompile(`-+`)
	s = re.ReplaceAllString(s, "-")

	return strings.Trim(s, "-")
}

// nextSequenceNumber scans dir for files matching NNN-* and returns the
// next number.
func nextSequenceNumber(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, nil
	}

	highest := -1
	re := regexp.MustCompile(`^(\d{3})-`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		n, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		if n > highest {
			highest = n
		}
	}

	return highest + 1, nil
}
