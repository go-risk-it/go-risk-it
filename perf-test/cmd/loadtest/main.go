package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/mapgraph"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player/heuristic"
)

func main() {
	url := flag.String("url", "http://localhost:8000", "Base URL of the server (Kong gateway)")
	anonKey := flag.String("anon-key", "", "Supabase anon key")
	players := flag.Int("players", 4, "Number of players per game")
	gameTimeout := flag.Duration("game-timeout", 10*time.Minute, "Timeout per game")
	mapFile := flag.String("map", "map.json", "Path to map.json")
	flag.Parse()

	if *anonKey == "" {
		// Try reading from env.
		*anonKey = os.Getenv("ANON_KEY")
	}

	if *anonKey == "" {
		log.Fatal("--anon-key flag or ANON_KEY env var is required")
	}

	// Parse map.
	graph, err := mapgraph.LoadFromFile(*mapFile)
	if err != nil {
		log.Fatalf("load map: %v", err)
	}

	log.Printf("loaded map: %d regions, %d continents", len(graph.Regions), len(graph.Continents))

	// Build WS URL from HTTP URL.
	wsURL := strings.Replace(*url, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	// Create strategy.
	strategy := heuristic.New(graph)

	// Run a single game.
	runner := orchestrator.NewGameRunner(*url, wsURL, *anonKey, strategy, *gameTimeout)
	result := runner.Run(0, *players)

	// Print results.
	fmt.Println()
	fmt.Println("=== Performance Test Results ===")
	fmt.Printf("Duration:  %v\n", result.Duration.Round(time.Millisecond))
	fmt.Printf("Moves:     %d\n", result.Moves)
	fmt.Printf("Errors:    %d\n", result.Errors)
	fmt.Printf("Timed out: %v\n", result.TimedOut)

	if result.Winner != "" {
		fmt.Printf("Winner:    %s\n", result.Winner)
	}

	if result.FatalError != nil {
		fmt.Printf("Fatal:     %v\n", result.FatalError)
		os.Exit(1)
	}
}
