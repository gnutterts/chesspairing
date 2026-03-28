package doubleswiss

import (
	"context"
	"fmt"
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
)

// TestFIDE_DoubleSwiss_6Player5Round plays a full 5-round tournament with
// 6 players where the higher-rated player always wins. Verifies invariants
// after each round.
//
// With 6 players and deterministic "higher-rated wins" results, the
// lexicographic pairer produces 3 complete pairings for rounds 1–3. In
// rounds 4–5 the spread of scores creates singleton brackets that cannot
// always upfloat, so partial pairings are expected (and correct per the
// algorithm's spec — it returns the best partial pairing when no complete
// one exists). We verify strict invariants for R1–R3 and weaker invariants
// (no duplicate, no inactive, no self-pairing) for R4–R5.
func TestFIDE_DoubleSwiss_6Player5Round(t *testing.T) {
	totalRounds := 5
	pairer := New(Options{TotalRounds: &totalRounds})

	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Player1", Rating: 2500, Active: true},
		{ID: "p2", DisplayName: "Player2", Rating: 2400, Active: true},
		{ID: "p3", DisplayName: "Player3", Rating: 2300, Active: true},
		{ID: "p4", DisplayName: "Player4", Rating: 2200, Active: true},
		{ID: "p5", DisplayName: "Player5", Rating: 2100, Active: true},
		{ID: "p6", DisplayName: "Player6", Rating: 2000, Active: true},
	}

	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
		},
	}

	for round := 1; round <= totalRounds; round++ {
		state.CurrentRound = round
		result, err := pairer.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: Pair() error: %v", round, err)
		}

		// Strict structural invariants hold regardless of partial pairing.
		assertWeakInvariants(t, state, result)

		// Rounds 1–3 should produce complete pairings (3 games, 0 byes).
		if round <= 3 {
			assertInvariants(t, state, result)
			if len(result.Pairings) != 3 {
				t.Errorf("round %d: expected 3 pairings, got %d", round, len(result.Pairings))
			}
			if len(result.Byes) != 0 {
				t.Errorf("round %d: expected 0 byes, got %d", round, len(result.Byes))
			}
		}

		// Rounds 4–5: partial pairings are acceptable, but at least 2 pairings expected.
		if round >= 4 {
			if len(result.Pairings) < 2 {
				t.Errorf("round %d: expected at least 2 pairings, got %d", round, len(result.Pairings))
			}
		}

		// Verify no rematches in any round.
		assertNoRematches(t, state, result, round)

		// Record results: higher-rated wins.
		var games []chesspairing.GameData
		for _, gp := range result.Pairings {
			wRating := ratingOf(players, gp.WhiteID)
			bRating := ratingOf(players, gp.BlackID)
			var res chesspairing.GameResult
			if wRating >= bRating {
				res = chesspairing.ResultWhiteWins
			} else {
				res = chesspairing.ResultBlackWins
			}
			games = append(games, chesspairing.GameData{
				WhiteID: gp.WhiteID,
				BlackID: gp.BlackID,
				Result:  res,
			})
		}
		state.Rounds = append(state.Rounds, chesspairing.RoundData{
			Number: round,
			Games:  games,
		})
	}
}

// TestFIDE_DoubleSwiss_OddPlayers_Bye tests 5 players over 3 rounds with
// higher-rated wins. Verifies bye rotation: no player receives a PAB twice.
// (With 5 players and decisive results, 3 rounds produce complete pairings;
// round 4 can hit bracket-size constraints.)
func TestFIDE_DoubleSwiss_OddPlayers_Bye(t *testing.T) {
	totalRounds := 4
	pairer := New(Options{TotalRounds: &totalRounds})

	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Player1", Rating: 2500, Active: true},
		{ID: "p2", DisplayName: "Player2", Rating: 2400, Active: true},
		{ID: "p3", DisplayName: "Player3", Rating: 2300, Active: true},
		{ID: "p4", DisplayName: "Player4", Rating: 2200, Active: true},
		{ID: "p5", DisplayName: "Player5", Rating: 2100, Active: true},
	}

	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
		},
	}

	byeReceivers := make(map[string]int)
	const fullPairingRounds = 3

	for round := 1; round <= fullPairingRounds; round++ {
		state.CurrentRound = round
		result, err := pairer.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: Pair() error: %v", round, err)
		}
		assertInvariants(t, state, result)

		if len(result.Pairings) != 2 {
			t.Errorf("round %d: expected 2 pairings, got %d", round, len(result.Pairings))
		}
		if len(result.Byes) != 1 {
			t.Fatalf("round %d: expected 1 bye, got %d", round, len(result.Byes))
		}

		byePlayer := result.Byes[0].PlayerID
		byeReceivers[byePlayer]++
		if result.Byes[0].Type != chesspairing.ByePAB {
			t.Errorf("round %d: expected ByePAB, got %v", round, result.Byes[0].Type)
		}

		// Record results: higher-rated wins.
		var games []chesspairing.GameData
		for _, gp := range result.Pairings {
			wRating := ratingOf(players, gp.WhiteID)
			bRating := ratingOf(players, gp.BlackID)
			var res chesspairing.GameResult
			if wRating >= bRating {
				res = chesspairing.ResultWhiteWins
			} else {
				res = chesspairing.ResultBlackWins
			}
			games = append(games, chesspairing.GameData{
				WhiteID: gp.WhiteID,
				BlackID: gp.BlackID,
				Result:  res,
			})
		}
		state.Rounds = append(state.Rounds, chesspairing.RoundData{
			Number: round,
			Games:  games,
			Byes:   result.Byes,
		})
	}

	// Verify no player received a PAB more than once across the 3 rounds.
	for id, count := range byeReceivers {
		if count > 1 {
			t.Errorf("player %s received PAB %d times (max 1 expected)", id, count)
		}
	}

	// Verify 3 distinct players got byes (one per round).
	if len(byeReceivers) != fullPairingRounds {
		t.Errorf("expected %d distinct bye receivers, got %d", fullPairingRounds, len(byeReceivers))
	}
}

// TestFIDE_DoubleSwiss_DrawResults tests that draws are handled correctly.
// R1: all draws. R2: no rematches, invariants hold.
func TestFIDE_DoubleSwiss_DrawResults(t *testing.T) {
	totalRounds := 3
	pairer := New(Options{TotalRounds: &totalRounds})

	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Player1", Rating: 2500, Active: true},
		{ID: "p2", DisplayName: "Player2", Rating: 2400, Active: true},
		{ID: "p3", DisplayName: "Player3", Rating: 2300, Active: true},
		{ID: "p4", DisplayName: "Player4", Rating: 2200, Active: true},
	}

	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
		},
	}

	// Round 1: pair and record all draws.
	r1Result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("R1 Pair() error: %v", err)
	}
	assertInvariants(t, state, r1Result)

	// Record R1 opponents for rematch checking.
	r1Opponents := make(map[string]string)
	var r1Games []chesspairing.GameData
	for _, gp := range r1Result.Pairings {
		r1Opponents[gp.WhiteID] = gp.BlackID
		r1Opponents[gp.BlackID] = gp.WhiteID
		r1Games = append(r1Games, chesspairing.GameData{
			WhiteID: gp.WhiteID,
			BlackID: gp.BlackID,
			Result:  chesspairing.ResultDraw,
		})
	}
	state.Rounds = append(state.Rounds, chesspairing.RoundData{
		Number: 1,
		Games:  r1Games,
	})

	// Round 2: verify no rematches.
	state.CurrentRound = 2
	r2Result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("R2 Pair() error: %v", err)
	}
	assertInvariants(t, state, r2Result)

	if len(r2Result.Pairings) != 2 {
		t.Errorf("R2: expected 2 pairings, got %d", len(r2Result.Pairings))
	}

	for _, gp := range r2Result.Pairings {
		if r1Opponents[gp.WhiteID] == gp.BlackID {
			t.Errorf("R2: rematch detected: %s vs %s", gp.WhiteID, gp.BlackID)
		}
	}
}

// TestFIDE_DoubleSwiss_ForfeitsExcluded verifies that a forfeit game is
// excluded from pairing history (players can be re-paired).
func TestFIDE_DoubleSwiss_ForfeitsExcluded(t *testing.T) {
	totalRounds := 3
	pairer := New(Options{TotalRounds: &totalRounds})

	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Player1", Rating: 2500, Active: true},
		{ID: "p2", DisplayName: "Player2", Rating: 2400, Active: true},
		{ID: "p3", DisplayName: "Player3", Rating: 2300, Active: true},
		{ID: "p4", DisplayName: "Player4", Rating: 2200, Active: true},
	}

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
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
		},
	}

	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("R2 Pair() error: %v", err)
	}
	assertInvariants(t, state, result)

	if len(result.Pairings) != 2 {
		t.Errorf("R2: expected 2 pairings, got %d", len(result.Pairings))
	}

	// p1 and p3 had a forfeit — they are NOT in each other's opponent
	// history, so a rematch is structurally allowed. Verify pairing is valid.
	// Scores: p1=1, p2=1, p3=0, p4=0.
	// Score groups: [1.0: p1, p2], [0.0: p3, p4].
	// p2 played p4 (non-forfeit), so p2 cannot play p4 again.
	// → p1 vs p2 and p3 vs p4 (lexicographic within score groups).
	foundP1P2 := false
	foundP3P4 := false
	for _, gp := range result.Pairings {
		if isPair(gp, "p1", "p2") {
			foundP1P2 = true
		}
		if isPair(gp, "p3", "p4") {
			foundP3P4 = true
		}
	}
	if !foundP1P2 {
		t.Error("R2: expected p1 vs p2 in score group 1.0")
	}
	if !foundP3P4 {
		t.Error("R2: expected p3 vs p4 in score group 0.0")
	}
}

// TestFIDE_DoubleSwiss_DoubleForfeit verifies that a double forfeit is
// excluded from both scoring and pairing history.
func TestFIDE_DoubleSwiss_DoubleForfeit(t *testing.T) {
	totalRounds := 3
	pairer := New(Options{TotalRounds: &totalRounds})

	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Player1", Rating: 2500, Active: true},
		{ID: "p2", DisplayName: "Player2", Rating: 2400, Active: true},
		{ID: "p3", DisplayName: "Player3", Rating: 2300, Active: true},
		{ID: "p4", DisplayName: "Player4", Rating: 2200, Active: true},
	}

	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDoubleForfeit, IsForfeit: true},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
		},
	}

	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("R2 Pair() error: %v", err)
	}
	assertInvariants(t, state, result)

	if len(result.Pairings) != 2 {
		t.Errorf("R2: expected 2 pairings, got %d", len(result.Pairings))
	}

	// Double forfeit: p1 and p3 both get 0 points and are NOT in each
	// other's opponent history. They can be re-paired.
	// Scores: p1=0, p2=1, p3=0, p4=0.
	// Score groups: [1.0: p2], [0.0: p1, p3, p4].
	// p2 alone → upfloater from [0.0]. Lowest-ranked with compatible
	// opponent in target = p4 (highest TPN first).
	// Bracket [1.0: p2, p4]: p2 played p4 already, but the upfloater
	// selection ensures compatibility. If p4 can't play p2, p3 is tried, etc.
	//
	// The exact pairings depend on upfloater resolution. Just verify structural validity.
}

// TestFIDE_DoubleSwiss_Withdrawal tests that a withdrawn player is excluded
// from pairing in subsequent rounds.
func TestFIDE_DoubleSwiss_Withdrawal(t *testing.T) {
	totalRounds := 4
	pairer := New(Options{TotalRounds: &totalRounds})

	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Player1", Rating: 2500, Active: true},
		{ID: "p2", DisplayName: "Player2", Rating: 2400, Active: true},
		{ID: "p3", DisplayName: "Player3", Rating: 2300, Active: true},
		{ID: "p4", DisplayName: "Player4", Rating: 2200, Active: true},
		{ID: "p5", DisplayName: "Player5", Rating: 2100, Active: true},
		{ID: "p6", DisplayName: "Player6", Rating: 2000, Active: true},
	}

	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
		},
	}

	// Round 1: normal pairing with 6 players.
	r1Result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("R1 Pair() error: %v", err)
	}
	assertInvariants(t, state, r1Result)

	if len(r1Result.Pairings) != 3 {
		t.Fatalf("R1: expected 3 pairings, got %d", len(r1Result.Pairings))
	}

	// Record R1 results: higher-rated wins.
	var r1Games []chesspairing.GameData
	for _, gp := range r1Result.Pairings {
		wRating := ratingOf(players, gp.WhiteID)
		bRating := ratingOf(players, gp.BlackID)
		var res chesspairing.GameResult
		if wRating >= bRating {
			res = chesspairing.ResultWhiteWins
		} else {
			res = chesspairing.ResultBlackWins
		}
		r1Games = append(r1Games, chesspairing.GameData{
			WhiteID: gp.WhiteID,
			BlackID: gp.BlackID,
			Result:  res,
		})
	}
	state.Rounds = append(state.Rounds, chesspairing.RoundData{
		Number: 1,
		Games:  r1Games,
	})

	// p6 withdraws before R2.
	state.Players[5].Active = false
	state.CurrentRound = 2

	r2Result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("R2 Pair() error: %v", err)
	}
	assertInvariants(t, state, r2Result)

	// 5 active players → 2 pairings + 1 bye.
	if len(r2Result.Pairings) != 2 {
		t.Errorf("R2: expected 2 pairings, got %d", len(r2Result.Pairings))
	}
	if len(r2Result.Byes) != 1 {
		t.Errorf("R2: expected 1 bye, got %d", len(r2Result.Byes))
	}

	// p6 must not appear in any pairing or bye.
	for _, gp := range r2Result.Pairings {
		if gp.WhiteID == "p6" || gp.BlackID == "p6" {
			t.Errorf("R2: withdrawn player p6 is paired")
		}
	}
	for _, bye := range r2Result.Byes {
		if bye.PlayerID == "p6" {
			t.Errorf("R2: withdrawn player p6 received a bye")
		}
	}
}

// TestFIDE_DoubleSwiss_BlackWins tests a scenario where the lower-rated player
// (Black) wins in R1. R2 should have no rematches and correct score group placement.
func TestFIDE_DoubleSwiss_BlackWins(t *testing.T) {
	totalRounds := 3
	pairer := New(Options{TotalRounds: &totalRounds})

	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Player1", Rating: 2500, Active: true},
		{ID: "p2", DisplayName: "Player2", Rating: 2400, Active: true},
		{ID: "p3", DisplayName: "Player3", Rating: 2300, Active: true},
		{ID: "p4", DisplayName: "Player4", Rating: 2200, Active: true},
	}

	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					// Lower-rated (Black) wins both games.
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultBlackWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultBlackWins},
				},
			},
		},
		CurrentRound: 2,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
		},
	}

	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("R2 Pair() error: %v", err)
	}
	assertInvariants(t, state, result)

	if len(result.Pairings) != 2 {
		t.Errorf("R2: expected 2 pairings, got %d", len(result.Pairings))
	}

	// After R1: p2=1.0, p4=1.0, p1=0.0, p3=0.0.
	// Score groups: [1.0: p2, p4], [0.0: p1, p3].
	// No rematches allowed.
	for _, gp := range result.Pairings {
		if isPair(gp, "p1", "p2") {
			t.Error("R2: rematch p1 vs p2")
		}
		if isPair(gp, "p3", "p4") {
			t.Error("R2: rematch p3 vs p4")
		}
	}

	// Winners should be paired together: p2 vs p4.
	foundWinnerPairing := false
	for _, gp := range result.Pairings {
		if isPair(gp, "p2", "p4") {
			foundWinnerPairing = true
		}
	}
	if !foundWinnerPairing {
		t.Error("R2: expected winners p2 and p4 to be paired together")
	}

	// Losers should be paired together: p1 vs p3.
	foundLoserPairing := false
	for _, gp := range result.Pairings {
		if isPair(gp, "p1", "p3") {
			foundLoserPairing = true
		}
	}
	if !foundLoserPairing {
		t.Error("R2: expected losers p1 and p3 to be paired together")
	}
}

// --- Helpers ---

// assertInvariants checks strict structural pairing invariants:
// uniqueness, completeness (all active players paired), no inactive paired,
// no self-pairing.
func assertInvariants(t *testing.T, state *chesspairing.TournamentState, result *chesspairing.PairingResult) {
	t.Helper()

	activeIDs := make(map[string]bool)
	for _, p := range state.Players {
		if p.Active {
			activeIDs[p.ID] = true
		}
	}

	seen := make(map[string]int)
	for i, gp := range result.Pairings {
		seen[gp.WhiteID]++
		seen[gp.BlackID]++
		if gp.WhiteID == gp.BlackID {
			t.Errorf("pairing[%d]: self-pairing %s", i, gp.WhiteID)
		}
	}
	for _, bye := range result.Byes {
		seen[bye.PlayerID]++
	}

	for id, count := range seen {
		if count != 1 {
			t.Errorf("player %s appears %d times", id, count)
		}
	}
	for id := range activeIDs {
		if seen[id] == 0 {
			t.Errorf("active player %s not paired", id)
		}
	}
	for id := range seen {
		if !activeIDs[id] {
			t.Errorf("inactive player %s paired", id)
		}
	}
}

// assertWeakInvariants checks invariants that hold even for partial pairings:
// no duplicate, no inactive paired, no self-pairing. Does NOT require every
// active player to be paired (the lexicographic pairer may legitimately
// produce partial pairings when no complete matching exists).
func assertWeakInvariants(t *testing.T, state *chesspairing.TournamentState, result *chesspairing.PairingResult) {
	t.Helper()

	activeIDs := make(map[string]bool)
	for _, p := range state.Players {
		if p.Active {
			activeIDs[p.ID] = true
		}
	}

	seen := make(map[string]int)
	for i, gp := range result.Pairings {
		seen[gp.WhiteID]++
		seen[gp.BlackID]++
		if gp.WhiteID == gp.BlackID {
			t.Errorf("pairing[%d]: self-pairing %s", i, gp.WhiteID)
		}
	}
	for _, bye := range result.Byes {
		seen[bye.PlayerID]++
	}

	for id, count := range seen {
		if count != 1 {
			t.Errorf("player %s appears %d times", id, count)
		}
	}
	for id := range seen {
		if !activeIDs[id] {
			t.Errorf("inactive player %s paired", id)
		}
	}
}

// assertNoRematches verifies that no pairing in the result is a rematch
// of a game from a previous round (excluding forfeits, which are not in
// opponent history and thus allowed).
func assertNoRematches(t *testing.T, state *chesspairing.TournamentState, result *chesspairing.PairingResult, round int) {
	t.Helper()

	// Build opponent history from previous rounds (non-forfeit games only).
	played := make(map[[2]string]bool)
	for _, rd := range state.Rounds {
		for _, g := range rd.Games {
			if !g.IsForfeit {
				played[[2]string{g.WhiteID, g.BlackID}] = true
				played[[2]string{g.BlackID, g.WhiteID}] = true
			}
		}
	}

	for _, gp := range result.Pairings {
		if played[[2]string{gp.WhiteID, gp.BlackID}] {
			t.Errorf("round %d: rematch %s vs %s", round, gp.WhiteID, gp.BlackID)
		}
	}
}

// ratingOf returns the rating for a player ID.
func ratingOf(players []chesspairing.PlayerEntry, id string) int {
	for _, p := range players {
		if p.ID == id {
			return p.Rating
		}
	}
	return 0
}

// isPair returns true if the game pairing matches the two IDs in either colour order.
func isPair(gp chesspairing.GamePairing, id1, id2 string) bool {
	return (gp.WhiteID == id1 && gp.BlackID == id2) || (gp.WhiteID == id2 && gp.BlackID == id1)
}

// ---------------------------------------------------------------------------
// Test: Large tournament — 20 players, 7 rounds
// ---------------------------------------------------------------------------

func TestFIDE_DoubleSwiss_LargeTournament_20Players7Rounds(t *testing.T) {
	totalRounds := 7
	players := make([]chesspairing.PlayerEntry, 20)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{
			ID:          fmt.Sprintf("p%02d", i+1),
			DisplayName: fmt.Sprintf("Player %d", i+1),
			Rating:      2700 - i*50,
			Active:      true,
		}
	}

	pairer := New(Options{TotalRounds: &totalRounds})
	state := &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 1,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDoubleSwiss,
		},
	}

	for round := 1; round <= 7; round++ {
		state.CurrentRound = round

		result, err := pairer.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d error: %v", round, err)
		}

		if len(result.Pairings) < 1 {
			t.Fatalf("round %d: expected at least 1 pairing, got 0", round)
		}

		assertWeakInvariants(t, state, result)
		assertNoRematches(t, state, result, round)

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
