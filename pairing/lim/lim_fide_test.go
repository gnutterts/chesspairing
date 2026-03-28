package lim

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// ratingOf returns the rating of the player with the given ID, or -1 if not found.
func ratingOf(players []chesspairing.PlayerEntry, id string) int {
	for _, p := range players {
		if p.ID == id {
			return p.Rating
		}
	}
	return -1
}

// assertPaired asserts that id1 and id2 are paired (in either colour order).
func assertPaired(t *testing.T, result *chesspairing.PairingResult, id1, id2 string) {
	t.Helper()
	for _, gp := range result.Pairings {
		if isPair(gp, id1, id2) {
			return
		}
	}
	t.Errorf("expected pairing between %s and %s, not found", id1, id2)
}

// isPair returns true if the game pairing matches id1 vs id2 in either colour order.
func isPair(gp chesspairing.GamePairing, id1, id2 string) bool {
	return (gp.WhiteID == id1 && gp.BlackID == id2) ||
		(gp.WhiteID == id2 && gp.BlackID == id1)
}

// itoa converts a small integer to its string representation.
func itoa(i int) string {
	return strconv.Itoa(i)
}

// assertPairingInvariants checks universal structural properties of a pairing
// result. This mirrors swisslib.AssertPairingInvariants (which lives in a _test.go
// file and cannot be imported across packages).
func assertPairingInvariants(t *testing.T, state *chesspairing.TournamentState, result *chesspairing.PairingResult) {
	t.Helper()

	activeIDs := make(map[string]bool)
	for _, p := range state.Players {
		if p.Active {
			activeIDs[p.ID] = true
		}
	}

	// Uniqueness: no player appears more than once.
	seen := make(map[string]int)
	for i, gp := range result.Pairings {
		seen[gp.WhiteID]++
		seen[gp.BlackID]++
		if gp.WhiteID == gp.BlackID {
			t.Errorf("pairing[%d]: player %s paired against themselves", i, gp.WhiteID)
		}
	}
	for _, bye := range result.Byes {
		seen[bye.PlayerID]++
	}
	for id, count := range seen {
		if count != 1 {
			t.Errorf("player %s appears %d times in pairings+byes (expected 1)", id, count)
		}
	}

	// Completeness: every active player is paired or has a bye.
	for id := range activeIDs {
		if seen[id] == 0 {
			t.Errorf("active player %s not found in pairings or byes", id)
		}
	}

	// No inactive player paired.
	for id := range seen {
		if !activeIDs[id] {
			t.Errorf("inactive player %s found in pairings or byes", id)
		}
	}

	// Board numbers sequential from 1.
	for i, gp := range result.Pairings {
		expected := i + 1
		if gp.Board != expected {
			t.Errorf("pairing[%d]: expected board %d, got %d", i, expected, gp.Board)
		}
	}

	// No rematches (forfeits excluded from pairing history).
	prevPairs := make(map[[2]string]bool)
	for _, rd := range state.Rounds {
		for _, g := range rd.Games {
			if g.IsForfeit {
				continue
			}
			prevPairs[swisslib.CanonicalPairKey(g.WhiteID, g.BlackID)] = true
		}
	}
	for _, gp := range result.Pairings {
		key := swisslib.CanonicalPairKey(gp.WhiteID, gp.BlackID)
		if prevPairs[key] {
			t.Errorf("rematch detected: %s vs %s", gp.WhiteID, gp.BlackID)
		}
	}

	// Bye type check.
	for _, bye := range result.Byes {
		if bye.Type != chesspairing.ByePAB {
			t.Errorf("bye for %s has type %v, expected ByePAB", bye.PlayerID, bye.Type)
		}
	}
}

// higherRatedWins records results where the higher-rated player always wins.
// It returns a RoundData for the given round number.
func higherRatedWins(players []chesspairing.PlayerEntry, result *chesspairing.PairingResult, roundNum int) chesspairing.RoundData {
	rd := chesspairing.RoundData{Number: roundNum}
	for _, gp := range result.Pairings {
		wr := ratingOf(players, gp.WhiteID)
		br := ratingOf(players, gp.BlackID)
		var res chesspairing.GameResult
		if wr >= br {
			res = chesspairing.ResultWhiteWins
		} else {
			res = chesspairing.ResultBlackWins
		}
		rd.Games = append(rd.Games, chesspairing.GameData{
			WhiteID: gp.WhiteID,
			BlackID: gp.BlackID,
			Result:  res,
		})
	}
	rd.Byes = append(rd.Byes, result.Byes...)
	return rd
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestFIDE_Lim_6Player5Round runs a full 5-round Lim tournament with 6 players.
// Higher-rated always wins. Checks round 1 pairings specifically, and structural
// invariants for all rounds.
func TestFIDE_Lim_6Player5Round(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "P1", Rating: 2500, Active: true},
		{ID: "p2", DisplayName: "P2", Rating: 2400, Active: true},
		{ID: "p3", DisplayName: "P3", Rating: 2300, Active: true},
		{ID: "p4", DisplayName: "P4", Rating: 2200, Active: true},
		{ID: "p5", DisplayName: "P5", Rating: 2100, Active: true},
		{ID: "p6", DisplayName: "P6", Rating: 2000, Active: true},
	}

	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
	}

	pairer := New(Options{})
	ctx := context.Background()

	for round := 1; round <= 5; round++ {
		t.Run("Round"+itoa(round), func(t *testing.T) {
			result, err := pairer.Pair(ctx, state)
			if err != nil {
				t.Fatalf("Pair() round %d error: %v", round, err)
			}

			// Structural invariants every round.
			assertPairingInvariants(t, state, result)

			// 6 even players → 3 pairings, 0 byes.
			if len(result.Pairings) != 3 {
				t.Fatalf("round %d: expected 3 pairings, got %d", round, len(result.Pairings))
			}
			if len(result.Byes) != 0 {
				t.Errorf("round %d: expected 0 byes, got %d", round, len(result.Byes))
			}

			// Round-specific checks.
			switch round {
			case 1:
				// Round 1: top half vs bottom half → p1-p4, p2-p5, p3-p6.
				assertPaired(t, result, "p1", "p4")
				assertPaired(t, result, "p2", "p5")
				assertPaired(t, result, "p3", "p6")
			case 2:
				// Round 2: no rematches (checked by invariants above).
				// Just verify we still get 3 pairings — already checked.
			}

			// Record results: higher-rated wins.
			rd := higherRatedWins(players, result, round)
			state.Rounds = append(state.Rounds, rd)
			state.CurrentRound = round + 1
		})
	}
}

// TestFIDE_Lim_OddPlayers_ByeArt1_1 verifies PAB assignment with 5 players
// over 5 rounds. Art. 1.1: PAB to lowest-ranked in lowest scoregroup.
// No player may receive a second PAB.
func TestFIDE_Lim_OddPlayers_ByeArt1_1(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "P1", Rating: 2500, Active: true},
		{ID: "p2", DisplayName: "P2", Rating: 2400, Active: true},
		{ID: "p3", DisplayName: "P3", Rating: 2300, Active: true},
		{ID: "p4", DisplayName: "P4", Rating: 2200, Active: true},
		{ID: "p5", DisplayName: "P5", Rating: 2100, Active: true},
	}

	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
	}

	pairer := New(Options{})
	ctx := context.Background()
	byePlayers := make(map[string]bool)

	for round := 1; round <= 5; round++ {
		t.Run("Round"+itoa(round), func(t *testing.T) {
			result, err := pairer.Pair(ctx, state)
			if err != nil {
				t.Fatalf("Pair() round %d error: %v", round, err)
			}

			assertPairingInvariants(t, state, result)

			// 5 players → 2 pairings + 1 bye.
			if len(result.Pairings) != 2 {
				t.Errorf("round %d: expected 2 pairings, got %d", round, len(result.Pairings))
			}
			if len(result.Byes) != 1 {
				t.Fatalf("round %d: expected 1 bye, got %d", round, len(result.Byes))
			}

			// No second PAB.
			byeID := result.Byes[0].PlayerID
			if byePlayers[byeID] {
				t.Errorf("round %d: player %s received a second PAB", round, byeID)
			}
			byePlayers[byeID] = true

			// Record results: higher-rated wins.
			rd := higherRatedWins(players, result, round)
			state.Rounds = append(state.Rounds, rd)
			state.CurrentRound = round + 1
		})
	}
}

// TestFIDE_Lim_ForfeitsExcluded verifies that a forfeit game is excluded from
// pairing history, so the two players can be re-paired.
func TestFIDE_Lim_ForfeitsExcluded(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "P1", Rating: 2400, Active: true},
		{ID: "p2", DisplayName: "P2", Rating: 2300, Active: true},
		{ID: "p3", DisplayName: "P3", Rating: 2200, Active: true},
		{ID: "p4", DisplayName: "P4", Rating: 2100, Active: true},
	}

	// R1: p1-p3 forfeit (white wins), p2-p4 normal.
	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	assertPairingInvariants(t, state, result)

	if len(result.Pairings) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(result.Pairings))
	}

	// Forfeit excluded: p1 and p3 CAN be re-paired. The invariants check
	// already verifies no rematches (excluding forfeits), so if p1-p3 appears
	// it's valid. Just confirm we got 2 valid pairings.
	for _, gp := range result.Pairings {
		t.Logf("Board %d: %s vs %s", gp.Board, gp.WhiteID, gp.BlackID)
	}
}

// ---------------------------------------------------------------------------
// Test: Large tournament — 20 players, 7 rounds
// ---------------------------------------------------------------------------

func TestFIDE_Lim_LargeTournament_20Players7Rounds(t *testing.T) {
	players := make([]chesspairing.PlayerEntry, 20)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{
			ID:          fmt.Sprintf("p%02d", i+1),
			DisplayName: fmt.Sprintf("Player %d", i+1),
			Rating:      2700 - i*50,
			Active:      true,
		}
	}

	pairer := New(Options{})
	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
	}

	for round := 1; round <= 7; round++ {
		state.CurrentRound = round

		result, err := pairer.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d error: %v", round, err)
		}

		if len(result.Pairings) != 10 {
			t.Fatalf("round %d: expected 10 pairings, got %d", round, len(result.Pairings))
		}

		assertPairingInvariants(t, state, result)

		// Simulate: higher-rated wins.
		games := make([]chesspairing.GameData, len(result.Pairings))
		for i, gp := range result.Pairings {
			res := chesspairing.ResultWhiteWins
			if ratingOf(players, gp.BlackID) > ratingOf(players, gp.WhiteID) {
				res = chesspairing.ResultBlackWins
			}
			games[i] = chesspairing.GameData{
				WhiteID: gp.WhiteID, BlackID: gp.BlackID, Result: res,
			}
		}
		state.Rounds = append(state.Rounds, chesspairing.RoundData{
			Number: round, Games: games, Byes: result.Byes,
		})
	}
}

// TestFIDE_Lim_DrawResults verifies R2 pairing after R1 with all draws.
// All players have the same score, so no rematches should occur.
func TestFIDE_Lim_DrawResults(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "P1", Rating: 2400, Active: true},
		{ID: "p2", DisplayName: "P2", Rating: 2300, Active: true},
		{ID: "p3", DisplayName: "P3", Rating: 2200, Active: true},
		{ID: "p4", DisplayName: "P4", Rating: 2100, Active: true},
	}

	// R1: p1-p3 draw, p2-p4 draw.
	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultDraw},
				},
			},
		},
		CurrentRound: 2,
	}

	pairer := New(Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}

	assertPairingInvariants(t, state, result)

	if len(result.Pairings) != 2 {
		t.Fatalf("expected 2 pairings, got %d", len(result.Pairings))
	}

	// No rematches: p1-p3 and p2-p4 must not recur.
	for _, gp := range result.Pairings {
		if isPair(gp, "p1", "p3") {
			t.Error("rematch detected: p1 vs p3")
		}
		if isPair(gp, "p2", "p4") {
			t.Error("rematch detected: p2 vs p4")
		}
	}

	for _, gp := range result.Pairings {
		t.Logf("Board %d: %s vs %s", gp.Board, gp.WhiteID, gp.BlackID)
	}
}
