package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
)

// RenderPackageArch produces a Mermaid graph TD showing layered package architecture.
func RenderPackageArch(archModel *model.ArchModel) string {
	var buf strings.Builder

	buf.WriteString("```mermaid\ngraph TD\n")

	writePackageSubgraphs(&buf, archModel)
	writePackageEdges(&buf, archModel)

	buf.WriteString("```")

	return buf.String()
}

// writePackageSubgraphs emits subgraphs for major layers with subsystem nodes inside.
func writePackageSubgraphs(
	buf *strings.Builder,
	archModel *model.ArchModel,
) {
	// Group subsystems by layer.
	byLayer := make(map[string][]*model.SubsystemInfo)
	for _, sub := range archModel.Subsystems {
		byLayer[sub.Layer] = append(byLayer[sub.Layer], sub)
	}

	sortedLayers := sortLayersByOrder(archModel, byLayer)

	for _, layerKey := range sortedLayers {
		info, ok := archModel.Layers[layerKey]
		if !ok {
			continue
		}

		subgraphID := packageLayerID(layerKey)
		label := info.Name + " Layer"

		fmt.Fprintf(buf, "    subgraph %s[\"%s\"]\n", subgraphID, label)
		buf.WriteString("        direction LR\n")

		subs := byLayer[layerKey]
		sort.Slice(subs, func(i, j int) bool {
			return subs[i].ID < subs[j].ID
		})

		for _, sub := range subs {
			nodeID := "pkg_" + sub.ID
			fmt.Fprintf(buf, "        %s[\"%s\"]\n", nodeID, sub.Label)
		}

		buf.WriteString("    end\n\n")
	}
}

// writePackageEdges emits layer-to-layer edges.
func writePackageEdges(
	buf *strings.Builder,
	archModel *model.ArchModel,
) {
	edges := make([]model.Edge, len(archModel.Edges))
	copy(edges, archModel.Edges)

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}

		return edges[i].To < edges[j].To
	})

	for _, e := range edges {
		from := packageLayerID(e.From)
		to := packageLayerID(e.To)

		fmt.Fprintf(buf, "    %s --> %s\n", from, to)
	}
}

// packageLayerID returns a Mermaid-safe subgraph ID for a layer in package diagrams.
func packageLayerID(layerName string) string {
	return "layer_" + strings.ReplaceAll(strings.ToLower(layerName), "-", "_")
}
