package render

import (
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
)

// RenderPackageTables produces markdown classification tables: wiring roots,
// full-tier packages, and lightweight-tier packages.
func RenderPackageTables(
	archModel *model.ArchModel,
	pkgs []model.GoPackage,
) string {
	var buf strings.Builder

	wiringRoots := collectWiringRoots(pkgs)
	fullTier, lightTier := classifyPackages(archModel)

	writeWiringRootsTable(&buf, wiringRoots)
	writeFullTierTable(&buf, fullTier, archModel)
	writeLightweightTable(&buf, lightTier, archModel)

	buf.WriteString("Note: Packages with `service.go` and 2 or fewer non-test GoFiles are\n")
	buf.WriteString(
		"classified as lightweight. The `game/logic/move/service` package has only\n",
	)
	buf.WriteString("1 file and is classified as a wiring root.")

	return buf.String()
}

// wiringRoot represents a wiring root package.
type wiringRoot struct {
	suffix string
	file   string
}

// collectWiringRoots finds wiring root packages from the raw go list output.
func collectWiringRoots(pkgs []model.GoPackage) []wiringRoot {
	var roots []wiringRoot

	for _, pkg := range pkgs {
		if !strings.HasPrefix(pkg.ImportPath, model.ModulePrefix) {
			continue
		}

		suffix := model.PackageSuffix(pkg.ImportPath)
		if suffix == "" {
			continue
		}

		// Wiring root detection: len(GoFiles) == 1 AND the single file matches
		// the last path segment with ".go" suffix.
		if len(pkg.GoFiles) != 1 {
			continue
		}

		lastSegment := path.Base(suffix)
		expectedFile := lastSegment + ".go"

		if pkg.GoFiles[0] == expectedFile {
			roots = append(roots, wiringRoot{suffix: suffix, file: pkg.GoFiles[0]})
		}
	}

	sort.Slice(roots, func(i, j int) bool {
		return roots[i].suffix < roots[j].suffix
	})

	return roots
}

// classifyPackages separates model packages into full-tier and lightweight-tier.
func classifyPackages(
	archModel *model.ArchModel,
) ([]*model.PackageInfo, []*model.PackageInfo) {
	var fullTier, lightTier []*model.PackageInfo

	suffixes := make([]string, 0, len(archModel.Packages))
	for s := range archModel.Packages {
		suffixes = append(suffixes, s)
	}

	sort.Strings(suffixes)

	for _, suffix := range suffixes {
		info := archModel.Packages[suffix]

		if IsFullTier(info) {
			fullTier = append(fullTier, info)
		} else {
			lightTier = append(lightTier, info)
		}
	}

	return fullTier, lightTier
}

// IsFullTier returns true if a package qualifies as full-tier:
// has "service.go" in GoFiles AND len(GoFiles) > 2.
func IsFullTier(info *model.PackageInfo) bool {
	return slices.Contains(info.GoFiles, "service.go") && len(info.GoFiles) > 2
}

// writeWiringRootsTable writes the wiring roots markdown table.
func writeWiringRootsTable(buf *strings.Builder, roots []wiringRoot) {
	buf.WriteString("### Wiring roots (no doc.go required)\n\n")
	buf.WriteString("Single-file packages whose sole purpose is fx.Module aggregation:\n\n")
	buf.WriteString("| Package | File |\n")
	buf.WriteString("|---------|------|\n")

	for _, root := range roots {
		fmt.Fprintf(buf, "| `%s` | `%s` |\n", root.suffix, root.file)
	}

	buf.WriteString(
		"\nDetection rule: `len(GoFiles) == 1` AND the single file is named after the\n",
	)
	buf.WriteString("last path segment (e.g., `game/logic/game.go`).\n\n")
}

// writeFullTierTable writes the full-tier packages markdown table.
func writeFullTierTable(
	buf *strings.Builder,
	pkgs []*model.PackageInfo,
	archModel *model.ArchModel,
) {
	buf.WriteString("### Full-tier packages\n\n")
	buf.WriteString(
		"Packages that define a service boundary: has `service.go` AND `len(GoFiles) > 2`\n",
	)
	buf.WriteString("AND is not a wiring root or generated package.\n\n")
	buf.WriteString("Required doc.go sections:\n")
	buf.WriteString("1. Package summary (first paragraph)\n")
	buf.WriteString("2. `# Layer` — layer name from taxonomy\n")
	buf.WriteString("3. `# Key Types` — exported interfaces and key structs\n")
	buf.WriteString("4. `# Dependencies` — key service dependencies\n\n")
	buf.WriteString("| Package | Layer | GoFiles |\n")
	buf.WriteString("|---------|-------|---------|\n")

	for _, info := range pkgs {
		layer := strings.ToLower(info.Layer)

		fmt.Fprintf(buf, "| `%s` | %s | %d |\n", info.Suffix, layer, len(info.GoFiles))
	}

	_ = archModel // used for consistency; layer lookup is from info directly

	buf.WriteString("\n")
}

// writeLightweightTable writes the lightweight-tier packages markdown table.
func writeLightweightTable(
	buf *strings.Builder,
	pkgs []*model.PackageInfo,
	archModel *model.ArchModel,
) {
	buf.WriteString("### Lightweight-tier packages\n\n")
	buf.WriteString("All other non-excluded packages. Required doc.go sections:\n")
	buf.WriteString("1. Package summary (first paragraph)\n")
	buf.WriteString("2. `# Layer` — layer name from taxonomy\n\n")
	buf.WriteString("| Package | Layer |\n")
	buf.WriteString("|---------|-------|\n")

	for _, info := range pkgs {
		layer := strings.ToLower(info.Layer)

		fmt.Fprintf(buf, "| `%s` | %s |\n", info.Suffix, layer)
	}

	_ = archModel // used for consistency

	buf.WriteString("\n")
}
