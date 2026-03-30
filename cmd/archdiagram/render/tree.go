package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
)

// treeNode represents a directory in the tree.
type treeNode struct {
	name     string
	summary  string
	children map[string]*treeNode
}

// RenderProjectTree produces a box-drawing project structure tree for internal/ packages.
// It walks sorted package suffixes, builds a directory tree, and annotates directories
// with summary comments from PackageInfo. Wiring roots and excluded packages are omitted.
func RenderProjectTree(archModel *model.ArchModel, pkgs []model.GoPackage) string {
	root := &treeNode{children: make(map[string]*treeNode)}

	// Build suffix set of all internal packages (including wiring roots, for tree structure).
	allSuffixes := collectAllInternalSuffixes(pkgs)

	// Sort suffixes for deterministic tree building.
	sort.Strings(allSuffixes)

	// Build tree from all non-excluded suffixes.
	for _, suffix := range allSuffixes {
		if isExcludedFromTree(suffix) {
			continue
		}

		parts := strings.Split(suffix, "/")
		insertPath(root, parts)
	}

	// Annotate tree nodes with summaries from the model.
	annotateSummaries(root, archModel, "")

	var buf strings.Builder

	buf.WriteString("```\ninternal/\n")

	renderChildren(&buf, root, "")

	buf.WriteString("```")

	return buf.String()
}

// collectAllInternalSuffixes returns suffixes of all packages under the module's internal/.
func collectAllInternalSuffixes(pkgs []model.GoPackage) []string {
	var result []string

	for _, pkg := range pkgs {
		if !strings.HasPrefix(pkg.ImportPath, model.ModulePrefix) {
			continue
		}

		suffix := model.PackageSuffix(pkg.ImportPath)
		if suffix == "" {
			continue
		}

		result = append(result, suffix)
	}

	return result
}

// isExcludedFromTree returns true for suffixes that should not appear in the tree.
func isExcludedFromTree(suffix string) bool {
	return strings.Contains(suffix, "/sqlc") || strings.Contains(suffix, "/mocks")
}

// insertPath creates tree nodes for each path segment.
func insertPath(root *treeNode, parts []string) {
	current := root

	for _, part := range parts {
		if current.children[part] == nil {
			current.children[part] = &treeNode{
				name:     part,
				children: make(map[string]*treeNode),
			}
		}

		current = current.children[part]
	}
}

// annotateSummaries walks the tree and attaches summaries from the model.
func annotateSummaries(
	node *treeNode,
	archModel *model.ArchModel,
	prefix string,
) {
	for _, child := range node.children {
		childPath := child.name
		if prefix != "" {
			childPath = prefix + "/" + child.name
		}

		// Look up summary from the model — only for non-wiring packages.
		if info, ok := archModel.Packages[childPath]; ok && info.Summary != "" {
			child.summary = info.Summary
		}

		annotateSummaries(child, archModel, childPath)
	}
}

// renderChildren renders the sorted children of a node with box-drawing characters.
func renderChildren(buf *strings.Builder, node *treeNode, indent string) {
	names := sortedChildNames(node)

	for i, name := range names {
		child := node.children[name]
		isLast := i == len(names)-1

		connector := "\u251C\u2500\u2500" // ├──
		if isLast {
			connector = "\u2514\u2500\u2500" // └──
		}

		line := fmt.Sprintf("%s%s %s/", indent, connector, name)

		if child.summary != "" {
			// Pad to alignment column and add summary.
			line = padToColumn(line, 30+len(indent)) + "# " + child.summary
		}

		buf.WriteString(line + "\n")

		childIndent := indent + "\u2502   " // │   (pipe + 3 spaces)
		if isLast {
			childIndent = indent + "    "
		}

		if len(child.children) > 0 {
			renderChildren(buf, child, childIndent)
		}
	}
}

// padToColumn pads s with spaces to reach at least the given column width.
func padToColumn(s string, col int) string {
	if len(s) >= col {
		return s + " "
	}

	return s + strings.Repeat(" ", col-len(s))
}

// sortedChildNames returns sorted child names of a node.
func sortedChildNames(node *treeNode) []string {
	names := make([]string, 0, len(node.children))
	for name := range node.children {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}
