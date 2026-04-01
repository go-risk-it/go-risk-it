// Package fileutil provides shared file-naming utilities for the loadtest
// subsystem: slug sanitization and auto-incrementing sequence numbers.
package fileutil

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Pre-compiled regexes for file operations.
var (
	reSequenceNum  = regexp.MustCompile(`^(\d{3})-`)
	reNonAlphaNum  = regexp.MustCompile(`[^a-z0-9-]`)
	reMultiHyphens = regexp.MustCompile(`-+`)
)

// SanitizeSlug normalizes a name for use in filenames: lowercase, hyphens,
// no special chars.
func SanitizeSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = reNonAlphaNum.ReplaceAllString(s, "")
	s = reMultiHyphens.ReplaceAllString(s, "-")

	return strings.Trim(s, "-")
}

// NextSequenceNumber scans dir for files matching NNN-* and returns the
// next number.
func NextSequenceNumber(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil // Directory doesn't exist yet — start at 0.
		}

		return 0, fmt.Errorf("read dir: %w", err)
	}

	highest := -1

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := reSequenceNum.FindStringSubmatch(entry.Name())
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
