//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/dutch"
)

// GoldenScenario describes a test scenario written to scenario.json.
type GoldenScenario struct {
	Description        string                     `json:"description"`
	Players            []chesspairing.PlayerEntry `json:"players"`
	TotalRounds        int                        `json:"totalRounds"`
	ResultStrategy     string                     `json:"resultStrategy"`
	WithdrawAfterRound map[string]int             `json:"withdrawAfterRound,omitempty"`
}

// scenarioDef is the internal definition used to drive generation.
type scenarioDef struct {
	dirName            string
	description        string
	players            []chesspairing.PlayerEntry
	totalRounds        int
	resultStrategy     string
	withdrawAfterRound map[string]int // playerID -> round number after which they withdraw
}

func main() {
	scenarios := buildScenarios()

	// Resolve goldenDir relative to this source file's location,
	// so the generator works regardless of CWD.
	_, thisFile, _, _ := runtime.Caller(0)
	goldenDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "golden")

	for _, sc := range scenarios {
		fmt.Printf("=== Generating: %s ===\n", sc.dirName)
		if err := generateScenario(goldenDir, sc); err != nil {
			log.Fatalf("scenario %s failed: %v", sc.dirName, err)
		}
		fmt.Printf("    Done (%d rounds)\n", sc.totalRounds)
	}

	fmt.Println("\nAll scenarios generated successfully.")
}

func buildScenarios() []scenarioDef {
	return []scenarioDef{
		{
			dirName:        "8-players-5-rounds",
			description:    "8 players, 5 rounds, higher-rated always wins",
			players:        makePlayers([]int{2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700}),
			totalRounds:    5,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "10-players-7-rounds",
			description:    "10 players, 7 rounds, higher-rated always wins",
			players:        makePlayers([]int{2500, 2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700, 1600}),
			totalRounds:    7,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "9-players-5-rounds",
			description:    "9 players (odd count), 5 rounds, bye rotation, higher-rated always wins",
			players:        makePlayers([]int{2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700, 1600}),
			totalRounds:    5,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "12-players-7-rounds",
			description:    "12 players, 7 rounds, higher-rated always wins",
			players:        makePlayers([]int{2600, 2500, 2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700, 1600, 1500}),
			totalRounds:    7,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "20-players-9-rounds",
			description:    "20 players, 9 rounds, higher-rated always wins",
			players:        make20Players(),
			totalRounds:    9,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "equal-ratings",
			description:    "6 players all rated 2000, 3 rounds, lower player ID wins ties",
			players:        makeEqualPlayers(6, 2000),
			totalRounds:    3,
			resultStrategy: "lower-id-wins",
		},
		{
			dirName:        "withdrawal",
			description:    "8 players, 5 rounds, p8 withdraws after round 2, higher-rated always wins",
			players:        makePlayers([]int{2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700}),
			totalRounds:    5,
			resultStrategy: "higher-rated-wins",
			withdrawAfterRound: map[string]int{
				"p8": 2,
			},
		},
	}
}

// makePlayers creates players p1..pN with the given ratings (must be in descending order).
func makePlayers(ratings []int) []chesspairing.PlayerEntry {
	players := make([]chesspairing.PlayerEntry, len(ratings))
	for i, r := range ratings {
		id := fmt.Sprintf("p%d", i+1)
		players[i] = chesspairing.PlayerEntry{
			ID:          id,
			DisplayName: fmt.Sprintf("Player %d", r),
			Rating:      r,
			Active:      true,
		}
	}
	return players
}

// make20Players creates 20 players with distinct ratings from 2700 down.
func make20Players() []chesspairing.PlayerEntry {
	ratings := make([]int, 20)
	for i := 0; i < 20; i++ {
		ratings[i] = 2700 - i*47
	}
	return makePlayers(ratings)
}

// makeEqualPlayers creates n players all with the same rating.
func makeEqualPlayers(n, rating int) []chesspairing.PlayerEntry {
	players := make([]chesspairing.PlayerEntry, n)
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("p%d", i+1)
		players[i] = chesspairing.PlayerEntry{
			ID:          id,
			DisplayName: fmt.Sprintf("Player %s", id),
			Rating:      rating,
			Active:      true,
		}
	}
	return players
}

func generateScenario(goldenDir string, sc scenarioDef) error {
	dir := filepath.Join(goldenDir, sc.dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	// Write scenario.json.
	scenarioFile := GoldenScenario{
		Description:        sc.description,
		Players:            sc.players,
		TotalRounds:        sc.totalRounds,
		ResultStrategy:     sc.resultStrategy,
		WithdrawAfterRound: sc.withdrawAfterRound,
	}
	if err := writeJSON(filepath.Join(dir, "scenario.json"), scenarioFile); err != nil {
		return err
	}

	// Build a rating lookup from the original player list.
	ratingByID := make(map[string]int, len(sc.players))
	for _, p := range sc.players {
		ratingByID[p.ID] = p.Rating
	}

	// Start with a copy of the players sorted by rating descending.
	players := make([]chesspairing.PlayerEntry, len(sc.players))
	copy(players, sc.players)
	sort.Slice(players, func(i, j int) bool {
		return players[i].Rating > players[j].Rating
	})

	pairer := dutch.New(dutch.Options{})

	state := chesspairing.TournamentState{
		Players:      players,
		Rounds:       nil,
		CurrentRound: 0,
		PairingConfig: chesspairing.PairingConfig{
			System:  chesspairing.PairingDutch,
			Options: map[string]any{},
		},
		ScoringConfig: chesspairing.ScoringConfig{
			System:      chesspairing.ScoringStandard,
			Tiebreakers: nil,
			Options:     map[string]any{},
		},
	}

	ctx := context.Background()

	for round := 1; round <= sc.totalRounds; round++ {
		// Apply any withdrawals that happen after the previous round.
		if sc.withdrawAfterRound != nil && round > 1 {
			for pid, afterRound := range sc.withdrawAfterRound {
				if afterRound == round-1 {
					for i := range state.Players {
						if state.Players[i].ID == pid {
							state.Players[i].Active = false
							fmt.Printf("    Round %d: player %s withdrawn\n", round, pid)
						}
					}
				}
			}
		}

		state.CurrentRound = round

		result, err := pairer.Pair(ctx, &state)
		if err != nil {
			return fmt.Errorf("pairing round %d: %w", round, err)
		}

		fmt.Printf("    Round %d: %d pairings, %d byes\n", round, len(result.Pairings), len(result.Byes))

		// Write round file.
		roundFile := filepath.Join(dir, fmt.Sprintf("round-%d.json", round))
		if err := writeJSON(roundFile, result); err != nil {
			return err
		}

		// Build round data with deterministic results and append to state.
		rd := chesspairing.RoundData{
			Number: round,
			Games:  make([]chesspairing.GameData, len(result.Pairings)),
			Byes:   result.Byes,
		}

		for i, p := range result.Pairings {
			res := determineResult(p.WhiteID, p.BlackID, ratingByID, sc.resultStrategy)
			rd.Games[i] = chesspairing.GameData{
				WhiteID:   p.WhiteID,
				BlackID:   p.BlackID,
				Result:    res,
				IsForfeit: false,
			}
		}

		state.Rounds = append(state.Rounds, rd)
	}

	return nil
}

// determineResult picks a result based on the strategy.
func determineResult(whiteID, blackID string, ratings map[string]int, strategy string) chesspairing.GameResult {
	switch strategy {
	case "higher-rated-wins":
		wr := ratings[whiteID]
		br := ratings[blackID]
		if wr > br {
			return chesspairing.ResultWhiteWins
		}
		if br > wr {
			return chesspairing.ResultBlackWins
		}
		// Equal ratings: draw.
		return chesspairing.ResultDraw

	case "lower-id-wins":
		// All ratings equal — lower ID (alphabetically) wins.
		if whiteID < blackID {
			return chesspairing.ResultWhiteWins
		}
		if blackID < whiteID {
			return chesspairing.ResultBlackWins
		}
		return chesspairing.ResultDraw

	default:
		log.Fatalf("unknown result strategy: %s", strategy)
		return ""
	}
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
