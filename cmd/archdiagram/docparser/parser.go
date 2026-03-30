// Package docparser extracts architectural metadata from Go package doc comments.
//
// It reads the doc.go file in a package directory, parses the package comment
// using [go/doc/comment.Parser], and extracts the layer name and summary from
// the structured comment format used by the go-risk-it living architecture spec.
//
// The expected format is:
//
//	// Package foo does something.
//	//
//	// # Layer
//	//
//	// Logic — description of the package's role.
//	package foo
//
// [ParseLayerAndSummary] returns the layer name (text before " —" in the
// paragraph after # Layer) and the first paragraph of the doc comment as the
// summary.
package docparser

import (
	"fmt"
	"go/doc/comment"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ParseLayerAndSummary reads the doc.go file in dir, parses its package
// comment, and extracts the architectural layer name and package summary.
//
// The layer is extracted from the first paragraph after a "# Layer" heading.
// The summary is the first paragraph of the doc comment (the package
// description before any headings).
//
// Returns empty strings (not an error) when:
//   - dir has no doc.go file
//   - doc.go has no package comment
//   - the package comment has no "# Layer" heading
//
// Returns an error only for I/O failures (directory doesn't exist, permission
// denied, etc.).
func ParseLayerAndSummary(dir string) (string, string, error) {
	doc, err := parseDocGo(dir)
	if err != nil {
		return "", "", err
	}

	if doc == nil {
		return "", "", nil
	}

	layer := extractLayer(doc)
	if layer == "" {
		// No # Layer heading means this is a pre-spec doc.go.
		// Don't extract summary from non-conforming files.
		return "", "", nil
	}

	return layer, extractSummary(doc), nil
}

// parseDocGo reads and parses the doc.go file in dir.
// Returns nil (not an error) if doc.go doesn't exist or has no package comment.
// Returns an error if dir itself doesn't exist or for other I/O or parse failures.
func parseDocGo(dir string) (*comment.Doc, error) {
	// Verify the directory exists — a missing directory is an I/O error,
	// not "package without doc.go".
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("accessing directory %s: %w", dir, err)
	}

	path := filepath.Join(dir, "doc.go")

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, nil //nolint:nilnil // missing doc.go is expected — not an error
		}

		return nil, fmt.Errorf("checking doc.go in %s: %w", dir, err)
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing doc.go in %s: %w", dir, err)
	}

	if file.Doc == nil || file.Doc.Text() == "" {
		return nil, nil //nolint:nilnil // empty package comment is expected — not an error
	}

	var p comment.Parser
	doc := p.Parse(file.Doc.Text())

	return doc, nil
}

// extractLayer returns the layer name from the paragraph after the "# Layer"
// heading. The layer name is the text before " —" (em-dash), or the full
// paragraph text if no em-dash is present. Returns empty string if no Layer
// heading or no paragraph after it.
func extractLayer(doc *comment.Doc) string {
	para := findParagraphAfterHeading(doc, "Layer")
	if para == nil {
		return ""
	}

	text := paragraphText(para)

	// Layer paragraphs use "LayerName — description" format
	if before, _, ok := strings.Cut(text, " \u2014"); ok {
		return before
	}

	// Fallback: use the full paragraph text, trimmed
	return strings.TrimSpace(text)
}

// extractSummary returns the first paragraph of the doc comment — the package
// description before any headings or other block elements.
func extractSummary(doc *comment.Doc) string {
	if len(doc.Content) == 0 {
		return ""
	}

	if p, ok := doc.Content[0].(*comment.Paragraph); ok {
		return paragraphText(p)
	}

	return ""
}

// findParagraphAfterHeading returns the first Paragraph block following a
// heading with the given text. Returns nil if not found.
func findParagraphAfterHeading(doc *comment.Doc, heading string) *comment.Paragraph {
	for idx, block := range doc.Content {
		h, ok := block.(*comment.Heading)
		if !ok {
			continue
		}

		if textContent(h.Text) != heading {
			continue
		}

		for j := idx + 1; j < len(doc.Content); j++ {
			if p, ok := doc.Content[j].(*comment.Paragraph); ok {
				return p
			}
			// Stop if we hit another heading
			if _, ok := doc.Content[j].(*comment.Heading); ok {
				break
			}
		}
	}

	return nil
}

// paragraphText returns the text content of a Paragraph block.
func paragraphText(p *comment.Paragraph) string {
	return textContent(p.Text)
}

// textContent renders a slice of comment.Text elements to a plain string.
func textContent(texts []comment.Text) string {
	var builder strings.Builder

	for _, text := range texts {
		switch val := text.(type) {
		case comment.Plain:
			builder.WriteString(string(val))
		case comment.Italic:
			builder.WriteString(string(val))
		case *comment.Link:
			builder.WriteString(textContent(val.Text))
		case *comment.DocLink:
			builder.WriteString(textContent(val.Text))
		}
	}

	return builder.String()
}
