package dutch

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPair_Round1_4Players(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2100, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1900, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2000, Active: true},
			{ID: "p4", DisplayName: "Diana", Rating: 1800, Active: true},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingSwiss,
		},
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Pairings) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 0 {
		t.Errorf("expected 0 byes, got %d", len(result.Byes))
	}

	// Round 1: S1={TPN1,TPN2} vs S2={TPN3,TPN4}
	// TPN order: Alice(2100)=1, Charlie(2000)=2, Bob(1900)=3, Diana(1800)=4
	// Expected: Alice vs Bob (1 vs 3), Charlie vs Diana (2 vs 4)
	pairedIDs := make(map[string]bool)
	pairingMap := make(map[string]string) // maps each player to their opponent
	for _, pair := range result.Pairings {
		pairedIDs[pair.WhiteID] = true
		pairedIDs[pair.BlackID] = true
		pairingMap[pair.WhiteID] = pair.BlackID
		pairingMap[pair.BlackID] = pair.WhiteID
	}
	for _, id := range []string{"p1", "p2", "p3", "p4"} {
		if !pairedIDs[id] {
			t.Errorf("player %s not paired", id)
		}
	}

	// Verify S1 vs S2 pairing: TPN1(p1/Alice) vs TPN3(p2/Bob), TPN2(p3/Charlie) vs TPN4(p4/Diana).
	if opp, ok := pairingMap["p1"]; !ok || opp != "p2" {
		t.Errorf("expected p1(Alice) paired with p2(Bob), got p1 paired with %s", pairingMap["p1"])
	}
	if opp, ok := pairingMap["p3"]; !ok || opp != "p4" {
		t.Errorf("expected p3(Charlie) paired with p4(Diana), got p3 paired with %s", pairingMap["p3"])
	}
}

func TestPair_Round1_OddPlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2100, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1900, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2000, Active: true},
		},
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingSwiss,
		},
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Pairings) != 1 {
		t.Errorf("expected 1 pairing, got %d", len(result.Pairings))
	}
	if len(result.Byes) != 1 {
		t.Errorf("expected 1 bye, got %d", len(result.Byes))
	}
}

func TestPair_TooFewPlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2100, Active: true},
		},
		CurrentRound: 1,
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	// Single player: should get bye, no error
	if err != nil {
		t.Fatalf("single player should not error: %v", err)
	}
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p1" {
		t.Errorf("single player should get bye, got byes=%v", result.Byes)
	}
}

func TestPair_AllWithdrawn(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2100, Active: false},
			{ID: "p2", DisplayName: "Bob", Rating: 1900, Active: false},
		},
		CurrentRound: 1,
	}

	p := New(Options{})
	_, err := p.Pair(context.Background(), state)
	if err == nil {
		t.Error("expected error for no active players")
	}
	if !errors.Is(err, ErrTooFewPlayers) {
		t.Errorf("expected ErrTooFewPlayers, got: %v", err)
	}
}

func TestForbiddenPairs(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "P2400", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "P2300", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "P2200", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "P2100", Rating: 2100, Active: true},
		},
		CurrentRound: 1,
	}

	p := New(Options{
		ForbiddenPairs: [][]string{{"p1", "p3"}},
	})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	for _, pairing := range result.Pairings {
		if (pairing.WhiteID == "p1" && pairing.BlackID == "p3") ||
			(pairing.WhiteID == "p3" && pairing.BlackID == "p1") {
			t.Error("p1 should not be paired with p3 (forbidden pair)")
		}
	}
}

func TestPair_Round2_WithHistory(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2100, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1900, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2000, Active: true},
			{ID: "p4", DisplayName: "Diana", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},
				},
			},
		},
		CurrentRound: 2,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingSwiss,
		},
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Pairings) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(result.Pairings))
	}

	// Verify no rematches from round 1.
	for _, pair := range result.Pairings {
		if (pair.WhiteID == "p1" && pair.BlackID == "p2") ||
			(pair.WhiteID == "p2" && pair.BlackID == "p1") {
			t.Error("p1 vs p2 is a rematch from round 1")
		}
		if (pair.WhiteID == "p3" && pair.BlackID == "p4") ||
			(pair.WhiteID == "p4" && pair.BlackID == "p3") {
			t.Error("p3 vs p4 is a rematch from round 1")
		}
	}
}

// goldenScenario describes a multi-round test scenario loaded from scenario.json.
type goldenScenario struct {
	Description        string                     `json:"description"`
	Players            []chesspairing.PlayerEntry `json:"players"`
	TotalRounds        int                        `json:"totalRounds"`
	ResultStrategy     string                     `json:"resultStrategy"`
	WithdrawAfterRound map[string]int             `json:"withdrawAfterRound,omitempty"`
}

// goldenDetermineResult picks a deterministic result for testing.
func goldenDetermineResult(whiteID, blackID string, ratings map[string]int, strategy string) chesspairing.GameResult {
	switch strategy {
	case "higher-rated-wins":
		if ratings[whiteID] > ratings[blackID] {
			return chesspairing.ResultWhiteWins
		}
		if ratings[blackID] > ratings[whiteID] {
			return chesspairing.ResultBlackWins
		}
		return chesspairing.ResultDraw
	case "lower-id-wins":
		if whiteID < blackID {
			return chesspairing.ResultWhiteWins
		}
		if blackID < whiteID {
			return chesspairing.ResultBlackWins
		}
		return chesspairing.ResultDraw
	default:
		return chesspairing.ResultDraw
	}
}

// goldenComparePairings compares the actual pairing result against the expected golden file.
func goldenComparePairings(t *testing.T, result, expected *chesspairing.PairingResult) {
	t.Helper()

	if len(result.Pairings) != len(expected.Pairings) {
		t.Errorf("expected %d pairings, got %d", len(expected.Pairings), len(result.Pairings))
		t.Logf("  expected: %v", formatPairings(expected))
		t.Logf("  got:      %v", formatPairings(result))
		return
	}

	for i, exp := range expected.Pairings {
		got := result.Pairings[i]
		if got.Board != exp.Board || got.WhiteID != exp.WhiteID || got.BlackID != exp.BlackID {
			t.Errorf("pairing[%d]: expected board %d %s-%s, got board %d %s-%s",
				i, exp.Board, exp.WhiteID, exp.BlackID,
				got.Board, got.WhiteID, got.BlackID)
		}
	}

	// Compare byes.
	expectedByes := make([]string, len(expected.Byes))
	for i, b := range expected.Byes {
		expectedByes[i] = b.PlayerID
	}
	gotByes := make([]string, len(result.Byes))
	for i, b := range result.Byes {
		gotByes[i] = b.PlayerID
	}
	sort.Strings(expectedByes)
	sort.Strings(gotByes)

	if len(gotByes) != len(expectedByes) {
		t.Errorf("expected byes %v, got byes %v", expectedByes, gotByes)
		return
	}
	for i := range expectedByes {
		if gotByes[i] != expectedByes[i] {
			t.Errorf("bye[%d]: expected %s, got %s", i, expectedByes[i], gotByes[i])
		}
	}
}

func formatPairings(r *chesspairing.PairingResult) string {
	var parts []string
	for _, p := range r.Pairings {
		parts = append(parts, p.WhiteID+"-"+p.BlackID)
	}
	if len(r.Byes) > 0 {
		byeIDs := make([]string, len(r.Byes))
		for i, b := range r.Byes {
			byeIDs[i] = b.PlayerID
		}
		parts = append(parts, "byes:"+strings.Join(byeIDs, ","))
	}
	return strings.Join(parts, " ")
}

func TestTopSeedColor_Black(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "P2400", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "P2300", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "P2200", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "P2100", Rating: 2100, Active: true},
		},
		CurrentRound: 1,
	}

	black := "black"
	p := New(Options{TopSeedColor: &black})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	if result.Pairings[0].BlackID != "p1" {
		t.Errorf("board 1: expected p1 as Black, got white=%s black=%s",
			result.Pairings[0].WhiteID, result.Pairings[0].BlackID)
	}
}

func TestGoldenFiles(t *testing.T) {
	// Find old-style input.json fixtures (single state, round files).
	inputs, _ := filepath.Glob("testdata/golden/*/input.json")
	for _, inputFile := range inputs {
		dir := filepath.Dir(inputFile)
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			inputData, err := os.ReadFile(inputFile) //nolint:gosec // test fixture
			if err != nil {
				t.Fatalf("read input.json: %v", err)
			}
			var state chesspairing.TournamentState
			if err := json.Unmarshal(inputData, &state); err != nil {
				t.Fatalf("unmarshal input.json: %v", err)
			}

			roundFiles, err := filepath.Glob(filepath.Join(dir, "round-*.json"))
			if err != nil {
				t.Fatalf("glob round files: %v", err)
			}
			sort.Strings(roundFiles)

			p := New(Options{})
			for _, roundFile := range roundFiles {
				roundName := filepath.Base(roundFile)
				t.Run(roundName, func(t *testing.T) {
					expectedData, err := os.ReadFile(roundFile) //nolint:gosec // test fixture
					if err != nil {
						t.Fatalf("read %s: %v", roundName, err)
					}
					var expected chesspairing.PairingResult
					if err := json.Unmarshal(expectedData, &expected); err != nil {
						t.Fatalf("unmarshal %s: %v", roundName, err)
					}

					result, err := p.Pair(context.Background(), &state)
					if err != nil {
						t.Fatalf("Pair() error: %v", err)
					}
					goldenComparePairings(t, result, &expected)
				})
			}
		})
	}

	// Find multi-round scenario.json fixtures.
	scenarios, _ := filepath.Glob("testdata/golden/*/scenario.json")
	for _, scenarioFile := range scenarios {
		dir := filepath.Dir(scenarioFile)
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			scenarioData, err := os.ReadFile(scenarioFile) //nolint:gosec // test fixture
			if err != nil {
				t.Fatalf("read scenario.json: %v", err)
			}
			var scenario goldenScenario
			if err := json.Unmarshal(scenarioData, &scenario); err != nil {
				t.Fatalf("unmarshal scenario.json: %v", err)
			}

			// Build rating lookup.
			ratingByID := make(map[string]int, len(scenario.Players))
			for _, p := range scenario.Players {
				ratingByID[p.ID] = p.Rating
			}

			// Sort players by rating descending (same as generator).
			players := make([]chesspairing.PlayerEntry, len(scenario.Players))
			copy(players, scenario.Players)
			sort.Slice(players, func(i, j int) bool {
				return players[i].Rating > players[j].Rating
			})

			state := chesspairing.TournamentState{
				Players:      players,
				CurrentRound: 0,
				PairingConfig: chesspairing.PairingConfig{
					System:  chesspairing.PairingSwiss,
					Options: map[string]any{},
				},
				ScoringConfig: chesspairing.ScoringConfig{
					System:  chesspairing.ScoringStandard,
					Options: map[string]any{},
				},
			}

			roundFiles, err := filepath.Glob(filepath.Join(dir, "round-*.json"))
			if err != nil {
				t.Fatalf("glob round files: %v", err)
			}
			sort.Strings(roundFiles)

			p := New(Options{})

			for roundNum, roundFile := range roundFiles {
				round := roundNum + 1
				roundName := filepath.Base(roundFile)

				// Apply withdrawals before this round.
				if scenario.WithdrawAfterRound != nil && round > 1 {
					for pid, afterRound := range scenario.WithdrawAfterRound {
						if afterRound == round-1 {
							for i := range state.Players {
								if state.Players[i].ID == pid {
									state.Players[i].Active = false
								}
							}
						}
					}
				}

				state.CurrentRound = round

				// Read expected golden pairings.
				expectedData, err := os.ReadFile(roundFile) //nolint:gosec // test fixture
				if err != nil {
					t.Fatalf("read %s: %v", roundName, err)
				}
				var expected chesspairing.PairingResult
				if err := json.Unmarshal(expectedData, &expected); err != nil {
					t.Fatalf("unmarshal %s: %v", roundName, err)
				}

				t.Run(roundName, func(t *testing.T) {
					result, err := p.Pair(context.Background(), &state)
					if err != nil {
						t.Fatalf("Pair() error: %v", err)
					}
					goldenComparePairings(t, result, &expected)
				})

				// Feed the EXPECTED (golden) pairings' results into state for the
				// next round. This ensures each round is tested against the same
				// history that the reference engine used, so a difference in round N
				// doesn't cascade into false failures in round N+1.
				rd := chesspairing.RoundData{
					Number: round,
					Games:  make([]chesspairing.GameData, len(expected.Pairings)),
					Byes:   expected.Byes,
				}
				for i, ep := range expected.Pairings {
					rd.Games[i] = chesspairing.GameData{
						WhiteID:   ep.WhiteID,
						BlackID:   ep.BlackID,
						Result:    goldenDetermineResult(ep.WhiteID, ep.BlackID, ratingByID, scenario.ResultStrategy),
						IsForfeit: false,
					}
				}
				state.Rounds = append(state.Rounds, rd)
			}
		})
	}
}

func TestBakuAcceleration_Round1(t *testing.T) {
	// 8 players, 5 rounds, Baku acceleration.
	// GA = BakuGASize(8) = 2 * ceil(8/4) = 4 (top 4 players).
	// Round 1 is a full VP round → GA players get +1.0 virtual points.
	// GA (PairingScore 1.0): p1(2400), p2(2300), p3(2200), p4(2100)
	// GB (PairingScore 0.0): p5(2000), p6(1900), p7(1800), p8(1700)
	// Expected: GA pairs within GA, GB pairs within GB (no mixing).
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "P2400", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "P2300", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "P2200", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "P2100", Rating: 2100, Active: true},
			{ID: "p5", DisplayName: "P2000", Rating: 2000, Active: true},
			{ID: "p6", DisplayName: "P1900", Rating: 1900, Active: true},
			{ID: "p7", DisplayName: "P1800", Rating: 1800, Active: true},
			{ID: "p8", DisplayName: "P1700", Rating: 1700, Active: true},
		},
		CurrentRound: 1,
	}

	baku := "baku"
	p := New(Options{Acceleration: &baku})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	if len(result.Pairings) != 4 {
		t.Fatalf("expected 4 pairings, got %d", len(result.Pairings))
	}

	// Verify no GA/GB mixing: each pairing must have both players from GA or both from GB.
	ga := map[string]bool{"p1": true, "p2": true, "p3": true, "p4": true}
	for _, pair := range result.Pairings {
		whiteGA := ga[pair.WhiteID]
		blackGA := ga[pair.BlackID]
		if whiteGA != blackGA {
			t.Errorf("GA/GB mixing: board %d has %s (GA=%v) vs %s (GA=%v)",
				pair.Board, pair.WhiteID, whiteGA, pair.BlackID, blackGA)
		}
	}

	// Verify acceleration note is present.
	foundAccelNote := false
	for _, note := range result.Notes {
		if strings.Contains(note, "Baku acceleration") {
			foundAccelNote = true
		}
	}
	if !foundAccelNote {
		t.Errorf("expected Baku acceleration note, got notes: %v", result.Notes)
	}
}

func TestBakuAcceleration_NoAcceleration(t *testing.T) {
	// Same 8 players, no acceleration. Standard Dutch: p1 vs p5 on board 1.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "P2400", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "P2300", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "P2200", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "P2100", Rating: 2100, Active: true},
			{ID: "p5", DisplayName: "P2000", Rating: 2000, Active: true},
			{ID: "p6", DisplayName: "P1900", Rating: 1900, Active: true},
			{ID: "p7", DisplayName: "P1800", Rating: 1800, Active: true},
			{ID: "p8", DisplayName: "P1700", Rating: 1700, Active: true},
		},
		CurrentRound: 1,
	}

	p := New(Options{})
	result, err := p.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	if len(result.Pairings) != 4 {
		t.Fatalf("expected 4 pairings, got %d", len(result.Pairings))
	}

	// Standard Dutch round 1: S1={TPN1-4} vs S2={TPN5-8}.
	// Board 1: TPN1(p1) vs TPN5(p5).
	board1 := result.Pairings[0]
	if (board1.WhiteID != "p1" || board1.BlackID != "p5") &&
		(board1.WhiteID != "p5" || board1.BlackID != "p1") {
		t.Errorf("expected p1 vs p5 on board 1, got %s vs %s", board1.WhiteID, board1.BlackID)
	}
}
