package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
)

// layerEmoji maps layer names to display emoji.
//
//nolint:gochecknoglobals // static lookup table
var layerEmoji = map[string]string{
	"Web":           "\U0001F310",   // 🌐
	"API":           "\U0001F4CB",   // 📋
	"Logic":         "\u2699\uFE0F", // ⚙️
	"Events-domain": "\U0001F4E1",   // 📡
	"Data":          "\U0001F4BE",   // 💾
	"Kernel":        "\U0001F527",   // 🔧
	"Game-domain":   "\U0001F3AE",   // 🎮
	"Game-support":  "\U0001F3AE",   // 🎮
	"Lobby-domain":  "\U0001F3AE",   // 🎮
	"Test":          "\U0001F9EA",   // 🧪
}

// layerStroke maps layer names to Mermaid stroke colors for styling.
//
//nolint:gochecknoglobals // static lookup table
var layerStroke = map[string]string{
	"Web":           "#1565C0",
	"API":           "#3949AB",
	"Logic":         "#2E7D32",
	"Events-domain": "#C62828",
	"Data":          "#E65100",
	"Kernel":        "#6A1B9A",
	"Game-domain":   "#00838F",
	"Game-support":  "#F9A825",
	"Lobby-domain":  "#00838F",
	"Test":          "#424242",
}

// componentLayerID returns a Mermaid-safe subgraph ID for a layer.
func componentLayerID(layerName string) string {
	return strings.ReplaceAll(strings.ReplaceAll(layerName, "-", ""), " ", "")
}

// RenderComponentArch produces a Mermaid graph LR component architecture diagram.
func RenderComponentArch(archModel *model.ArchModel) string {
	var buf strings.Builder

	buf.WriteString("```mermaid\ngraph LR\n")

	writeComponentSubgraphs(&buf, archModel)
	writeComponentEdges(&buf, archModel)
	writeComponentStyles(&buf, archModel)

	buf.WriteString("```")

	return buf.String()
}

// writeComponentSubgraphs emits styled subgraphs per layer with subsystem nodes.
func writeComponentSubgraphs(
	buf *strings.Builder,
	archModel *model.ArchModel,
) {
	// Group subsystems by layer.
	byLayer := make(map[string][]*model.SubsystemInfo)
	for _, sub := range archModel.Subsystems {
		byLayer[sub.Layer] = append(byLayer[sub.Layer], sub)
	}

	// Sort layers by order.
	sortedLayers := sortLayersByOrder(archModel, byLayer)

	for _, layerKey := range sortedLayers {
		info, ok := archModel.Layers[layerKey]
		if !ok {
			continue
		}

		emoji := layerEmoji[layerKey]
		subgraphID := componentLayerID(layerKey)

		label := info.Name
		if emoji != "" {
			label = fmt.Sprintf("%s %s", emoji, info.Name)
		}

		fmt.Fprintf(buf, "    subgraph %s[\"%s\"]\n", subgraphID, label)
		buf.WriteString("        direction TB\n")

		subs := byLayer[layerKey]
		sort.Slice(subs, func(i, j int) bool {
			return subs[i].ID < subs[j].ID
		})

		for _, sub := range subs {
			nodeID := strings.ReplaceAll(sub.ID, "_", "")
			fmt.Fprintf(
				buf,
				"        %s[\"%s\"]\n",
				nodeID,
				sub.Label,
			)
		}

		buf.WriteString("    end\n\n")
	}
}

// writeComponentEdges emits cross-layer edges with labels.
func writeComponentEdges(
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

	if len(edges) > 0 {
		buf.WriteString("    %% Cross-layer dependencies\n")
	}

	for _, e := range edges {
		// Pick representative subsystem from each layer for edge endpoints.
		fromNode := pickRepresentativeNode(archModel, e.From)
		toNode := pickRepresentativeNode(archModel, e.To)

		if fromNode == "" || toNode == "" {
			continue
		}

		fmt.Fprintf(buf, "    %s --> %s\n", fromNode, toNode)
	}

	if len(edges) > 0 {
		buf.WriteString("\n")
	}
}

// pickRepresentativeNode returns the Mermaid node ID of the first subsystem in a layer.
func pickRepresentativeNode(m *model.ArchModel, layerName string) string {
	var candidates []*model.SubsystemInfo

	for _, sub := range m.Subsystems {
		if sub.Layer == layerName {
			candidates = append(candidates, sub)
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].ID < candidates[j].ID
	})

	return strings.ReplaceAll(candidates[0].ID, "_", "")
}

// writeComponentStyles emits style directives matching layer colors.
func writeComponentStyles(
	buf *strings.Builder,
	archModel *model.ArchModel,
) {
	byLayer := make(map[string]bool)
	for _, sub := range archModel.Subsystems {
		byLayer[sub.Layer] = true
	}

	sortedLayers := make([]string, 0, len(byLayer))
	for l := range byLayer {
		sortedLayers = append(sortedLayers, l)
	}

	sort.Strings(sortedLayers)

	buf.WriteString("    %% Styling\n")

	for _, layerKey := range sortedLayers {
		info, ok := archModel.Layers[layerKey]
		if !ok {
			continue
		}

		subgraphID := componentLayerID(layerKey)
		stroke := layerStroke[layerKey]

		if stroke == "" {
			stroke = "#333"
		}

		fmt.Fprintf(
			buf,
			"    style %s fill:%s,stroke:%s,color:#000\n",
			subgraphID,
			info.Color,
			stroke,
		)
	}
}

// sortLayersByOrder returns layer names sorted by their Order field.
func sortLayersByOrder(
	archModel *model.ArchModel,
	byLayer map[string][]*model.SubsystemInfo,
) []string {
	sortedLayers := make([]string, 0, len(byLayer))
	for l := range byLayer {
		sortedLayers = append(sortedLayers, l)
	}

	sort.Slice(sortedLayers, func(i, j int) bool {
		layerI, okI := archModel.Layers[sortedLayers[i]]
		layerJ, okJ := archModel.Layers[sortedLayers[j]]

		if !okI || !okJ {
			return sortedLayers[i] < sortedLayers[j]
		}

		return layerI.Order < layerJ.Order
	})

	return sortedLayers
}
