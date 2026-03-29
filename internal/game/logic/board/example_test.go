package board_test

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/logic/board"
)

func ExampleNewGraph() {
	boardDto := &board.BoardDto{
		Regions: []board.RegionDto{
			{ExternalReference: "alaska"},
			{ExternalReference: "kamchatka"},
			{ExternalReference: "brazil"},
		},
		Borders: []board.BorderDto{
			{Source: "alaska", Target: "kamchatka"},
		},
	}

	graph, err := board.NewGraph(boardDto)
	if err != nil {
		panic(err)
	}

	fmt.Println("regions:", graph.GetRegions())
	fmt.Println("alaska-kamchatka neighbours:", graph.AreNeighbours("alaska", "kamchatka"))
	fmt.Println("alaska-brazil neighbours:", graph.AreNeighbours("alaska", "brazil"))
	// Output:
	// regions: [alaska brazil kamchatka]
	// alaska-kamchatka neighbours: true
	// alaska-brazil neighbours: false
}
