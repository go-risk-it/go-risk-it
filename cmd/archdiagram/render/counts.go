package render

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// rulePattern matches "// Rule " comment lines in arch_test.go.
var rulePattern = regexp.MustCompile(`(?m)^// Rule `)

// CountArchRules counts architecture rules by finding "// Rule " comments
// in all arch_*_test.go files under internal/.
func CountArchRules(repoRoot string) (int, error) {
	dir := filepath.Join(repoRoot, "internal")

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("reading internal/ directory: %w", err)
	}

	total := 0

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "arch_") || !strings.HasSuffix(name, "_test.go") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return 0, fmt.Errorf("reading %s: %w", name, err)
		}

		matches := rulePattern.FindAll(data, -1)
		total += len(matches)
	}

	return total, nil
}

// invariantEntryPattern matches "{" entries within the AllInvariants slice literal.
// It looks for lines starting with optional whitespace followed by "{" that contain
// a quoted invariant name.
var invariantEntryPattern = regexp.MustCompile(`(?m)^\s*\{`)

// CountInvariants counts game invariants by finding entries in the AllInvariants
// slice in invariant.go.
func CountInvariants(repoRoot string) (int, error) {
	path := filepath.Join(repoRoot, "internal", "game", "testing", "invariant", "invariant.go")

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("reading invariant.go: %w", err)
	}

	content := string(data)

	// Find the AllInvariants slice literal.
	startMarker := "var AllInvariants = []Invariant{"
	startIdx := strings.Index(content, startMarker)

	if startIdx < 0 {
		return 0, errors.New("AllInvariants slice not found in invariant.go")
	}

	// Find the closing brace of the slice literal by counting braces.
	sliceStart := startIdx + len(startMarker)
	depth := 1
	sliceEnd := sliceStart

	for idx := sliceStart; idx < len(content) && depth > 0; idx++ {
		switch content[idx] {
		case '{':
			depth++
		case '}':
			depth--
		}

		sliceEnd = idx
	}

	sliceBody := content[sliceStart:sliceEnd]

	// Count the struct literal entries: lines starting with `{` inside the slice.
	matches := invariantEntryPattern.FindAllString(sliceBody, -1)

	return len(matches), nil
}
