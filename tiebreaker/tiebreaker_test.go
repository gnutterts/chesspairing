package tiebreaker

import (
	"context"
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
)

// Tournament setup for most tests:
// 4 players, 3 rounds (round-robin style).
//
// Round 1: p1 beats p2 (1-0), p3 draws p4 (½-½)
// Round 2: p1 draws p3 (½-½), p2 beats p4 (1-0)
// Round 3: p1 beats p4 (1-0), p3 beats p2 (0-1)
//
// Final scores (standard 1-½-0):
//
//	p1: 1.0 + 0.5 + 1.0 = 2.5
//	p3: 0.5 + 0.5 + 1.0 = 2.0
//	p2: 0.0 + 1.0 + 0.0 = 1.0
//	p4: 0.5 + 0.0 + 0.0 = 0.5
func standardState() *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultBlackWins},
				},
			},
		},
	}
}

func standardScores() []chesspairing.PlayerScore {
	return []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 2.5, Rank: 1},
		{PlayerID: "p3", Score: 2.0, Rank: 2},
		{PlayerID: "p2", Score: 1.0, Rank: 3},
		{PlayerID: "p4", Score: 0.5, Rank: 4},
	}
}

func valueMap(values []chesspairing.TieBreakValue) map[string]float64 {
	m := make(map[string]float64, len(values))
	for _, v := range values {
		m[v.PlayerID] = v.Value
	}
	return m
}

// --- Registry tests ---

func TestRegistryGet(t *testing.T) {
	ids := []string{
		"buchholz", "buchholz-cut1", "buchholz-cut2", "buchholz-median", "buchholz-median2",
		"sonneborn-berger", "direct-encounter", "wins", "win",
		"koya", "progressive", "aro", "black-games", "black-wins",
		"games-played", "rounds-played", "standard-points", "pairing-number",
		"fore-buchholz", "avg-opponent-buchholz",
		"performance-rating", "performance-points",
		"avg-opponent-tpr", "avg-opponent-ptp",
		"player-rating",
	}
	for _, id := range ids {
		tb, err := Get(id)
		if err != nil {
			t.Errorf("Get(%q) error: %v", id, err)
			continue
		}
		if tb.ID() != id {
			t.Errorf("Get(%q).ID() = %q", id, tb.ID())
		}
		if tb.Name() == "" {
			t.Errorf("Get(%q).Name() is empty", id)
		}
	}
}

func TestRegistryGetUnknown(t *testing.T) {
	_, err := Get("nonexistent")
	if err == nil {
		t.Error("Get(nonexistent) should return error")
	}
}

func TestRegistryAll(t *testing.T) {
	all := All()
	if len(all) < 25 {
		t.Errorf("expected at least 25 registered tiebreakers, got %d", len(all))
	}
}

// --- Buchholz tests ---

func TestBuchholzFull(t *testing.T) {
	tb, _ := Get("buchholz")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1 opponents: p2(1.0), p3(2.0), p4(0.5) → Buchholz = 3.5
	if vm["p1"] != 3.5 {
		t.Errorf("p1 Buchholz = %v, want 3.5", vm["p1"])
	}
	// p2 opponents: p1(2.5), p4(0.5), p3(2.0) → Buchholz = 5.0
	if vm["p2"] != 5.0 {
		t.Errorf("p2 Buchholz = %v, want 5.0", vm["p2"])
	}
	// p3 opponents: p4(0.5), p1(2.5), p2(1.0) → Buchholz = 4.0
	if vm["p3"] != 4.0 {
		t.Errorf("p3 Buchholz = %v, want 4.0", vm["p3"])
	}
	// p4 opponents: p3(2.0), p2(1.0), p1(2.5) → Buchholz = 5.5
	if vm["p4"] != 5.5 {
		t.Errorf("p4 Buchholz = %v, want 5.5", vm["p4"])
	}
}

func TestBuchholzCut1(t *testing.T) {
	tb, _ := Get("buchholz-cut1")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1 opponents sorted: [0.5, 1.0, 2.0] → drop lowest (0.5) → 3.0
	if vm["p1"] != 3.0 {
		t.Errorf("p1 Buchholz Cut-1 = %v, want 3.0", vm["p1"])
	}
	// p2 opponents sorted: [0.5, 2.0, 2.5] → drop lowest (0.5) → 4.5
	if vm["p2"] != 4.5 {
		t.Errorf("p2 Buchholz Cut-1 = %v, want 4.5", vm["p2"])
	}
}

func TestBuchholzCut2(t *testing.T) {
	tb, _ := Get("buchholz-cut2")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1 opponents sorted: [0.5, 1.0, 2.0] → drop 2 lowest (0.5, 1.0) → 2.0
	if vm["p1"] != 2.0 {
		t.Errorf("p1 Buchholz Cut-2 = %v, want 2.0", vm["p1"])
	}
	// p2 opponents sorted: [0.5, 2.0, 2.5] → drop 2 lowest (0.5, 2.0) → 2.5
	if vm["p2"] != 2.5 {
		t.Errorf("p2 Buchholz Cut-2 = %v, want 2.5", vm["p2"])
	}
}

func TestBuchholzMedian(t *testing.T) {
	tb, _ := Get("buchholz-median")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1 opponents sorted: [0.5, 1.0, 2.0] → drop lowest + highest → 1.0
	if vm["p1"] != 1.0 {
		t.Errorf("p1 Buchholz Median = %v, want 1.0", vm["p1"])
	}
	// p2 opponents sorted: [0.5, 2.0, 2.5] → drop lowest + highest → 2.0
	if vm["p2"] != 2.0 {
		t.Errorf("p2 Buchholz Median = %v, want 2.0", vm["p2"])
	}
}

func TestBuchholzNoRounds(t *testing.T) {
	tb, _ := Get("buchholz")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("Buchholz with no rounds = %v, want 0", values[0].Value)
	}
}

func TestBuchholzWithBye(t *testing.T) {
	tb, _ := Get("buchholz")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
		},
	}
	// p1: 1.0 (win), p2: 0.0 (loss), p3: 1.0 (bye, using standard bye=1.0)
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 1.0, Rank: 1},
		{PlayerID: "p3", Score: 1.0, Rank: 2},
		{PlayerID: "p2", Score: 0.0, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1 opponent: p2(0.0) → Buchholz = 0.0
	if vm["p1"] != 0.0 {
		t.Errorf("p1 Buchholz = %v, want 0.0", vm["p1"])
	}
	// p3 has bye → virtual opponent score = own score = 1.0 → Buchholz = 1.0
	if vm["p3"] != 1.0 {
		t.Errorf("p3 Buchholz = %v, want 1.0 (bye → virtual opponent = own score)", vm["p3"])
	}
}

// --- Sonneborn-Berger tests ---

func TestSonnebornBerger(t *testing.T) {
	tb, _ := Get("sonneborn-berger")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: beat p2(1.0) → +1.0, drew p3(2.0) → +1.0, beat p4(0.5) → +0.5 = 2.5
	if vm["p1"] != 2.5 {
		t.Errorf("p1 SB = %v, want 2.5", vm["p1"])
	}
	// p2: lost p1(2.5) → 0, beat p4(0.5) → +0.5, lost p3(2.0) → 0 = 0.5
	if vm["p2"] != 0.5 {
		t.Errorf("p2 SB = %v, want 0.5", vm["p2"])
	}
	// p3: drew p4(0.5) → +0.25, drew p1(2.5) → +1.25, beat p2(1.0) → +1.0 = 2.5
	if vm["p3"] != 2.5 {
		t.Errorf("p3 SB = %v, want 2.5", vm["p3"])
	}
	// p4: drew p3(2.0) → +1.0, lost p2(1.0) → 0, lost p1(2.5) → 0 = 1.0
	if vm["p4"] != 1.0 {
		t.Errorf("p4 SB = %v, want 1.0", vm["p4"])
	}
}

func TestSonnebornBergerNoGames(t *testing.T) {
	tb, _ := Get("sonneborn-berger")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("SB with no games = %v, want 0", values[0].Value)
	}
}

// --- Direct Encounter tests ---

func TestDirectEncounter(t *testing.T) {
	tb, _ := Get("direct-encounter")

	// Create a scenario with tied players.
	// p1 and p3 are tied at 1.5 each.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins}, // p1 +1
					// p3 bye → p3 +1
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw}, // p1 +0.5, p3 +0.5
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p2", Type: chesspairing.ByePAB}},
			},
		},
	}
	// p1: 1.5, p2: 1.0, p3: 1.5 — p1 and p3 are tied.
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 1.5, Rank: 1},
		{PlayerID: "p3", Score: 1.5, Rank: 2},
		{PlayerID: "p2", Score: 1.0, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1 vs p3: drew → p1 gets 0.5 from direct encounter.
	// p3 vs p1: drew → p3 gets 0.5 from direct encounter.
	if vm["p1"] != 0.5 {
		t.Errorf("p1 DE = %v, want 0.5", vm["p1"])
	}
	if vm["p3"] != 0.5 {
		t.Errorf("p3 DE = %v, want 0.5", vm["p3"])
	}
	// p2 is not tied with anyone → DE = 0.
	if vm["p2"] != 0 {
		t.Errorf("p2 DE = %v, want 0 (not tied)", vm["p2"])
	}
}

func TestDirectEncounterWinBreaksTie(t *testing.T) {
	tb, _ := Get("direct-encounter")

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins}, // p1 beats p2
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins}, // p2 beats p3
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByePAB}},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p3", BlackID: "p1", Result: chesspairing.ResultWhiteWins}, // p3 beats p1
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p2", Type: chesspairing.ByePAB}},
			},
		},
	}
	// All three: 1 win + 1 bye + 1 loss = 2.0 each (with standard bye = 1.0).
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 2.0, Rank: 1},
		{PlayerID: "p2", Score: 2.0, Rank: 2},
		{PlayerID: "p3", Score: 2.0, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// All three are in the same tied group. Each beat one and lost to one.
	// p1: beat p2 (1.0), lost to p3 (0) → DE = 1.0
	// p2: beat p3 (1.0), lost to p1 (0) → DE = 1.0
	// p3: beat p1 (1.0), lost to p2 (0) → DE = 1.0
	// All equal — direct encounter can't break this tie.
	if vm["p1"] != 1.0 {
		t.Errorf("p1 DE = %v, want 1.0", vm["p1"])
	}
	if vm["p2"] != 1.0 {
		t.Errorf("p2 DE = %v, want 1.0", vm["p2"])
	}
	if vm["p3"] != 1.0 {
		t.Errorf("p3 DE = %v, want 1.0", vm["p3"])
	}
}

func TestDirectEncounterNoTie(t *testing.T) {
	tb, _ := Get("direct-encounter")
	state := standardState()
	scores := standardScores() // all different scores

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// No one is tied → all DE values should be 0.
	for id, val := range vm {
		if val != 0 {
			t.Errorf("%s DE = %v, want 0 (no ties)", id, val)
		}
	}
}

// --- Wins tests ---

func TestWins(t *testing.T) {
	tb, _ := Get("wins")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: won vs p2, drew p3, won vs p4 → 2 wins
	if vm["p1"] != 2 {
		t.Errorf("p1 wins = %v, want 2", vm["p1"])
	}
	// p2: lost to p1, won vs p4, lost to p3 → 1 win
	if vm["p2"] != 1 {
		t.Errorf("p2 wins = %v, want 1", vm["p2"])
	}
	// p3: drew p4, drew p1, won vs p2 → 1 win
	if vm["p3"] != 1 {
		t.Errorf("p3 wins = %v, want 1", vm["p3"])
	}
	// p4: drew p3, lost to p2, lost to p1 → 0 wins
	if vm["p4"] != 0 {
		t.Errorf("p4 wins = %v, want 0", vm["p4"])
	}
}

func TestWinsNoGames(t *testing.T) {
	tb, _ := Get("wins")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("wins with no games = %v, want 0", values[0].Value)
	}
}

// --- Koya tests ---

func TestKoya(t *testing.T) {
	tb, _ := Get("koya")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// 3 rounds → threshold = 1.5
	// Qualifying players (score >= 1.5): p1(2.5), p3(2.0) — yes. p2(1.0), p4(0.5) — no.
	//
	// p1: vs p2(no) skip, vs p3(yes) drew → 0.5, vs p4(no) skip → Koya = 0.5
	if vm["p1"] != 0.5 {
		t.Errorf("p1 Koya = %v, want 0.5", vm["p1"])
	}
	// p3: vs p4(no) skip, vs p1(yes) drew → 0.5, vs p2(no) skip → Koya = 0.5
	if vm["p3"] != 0.5 {
		t.Errorf("p3 Koya = %v, want 0.5", vm["p3"])
	}
	// p2: vs p1(yes) lost → 0, vs p4(no) skip, vs p3(yes) lost → 0 → Koya = 0
	if vm["p2"] != 0 {
		t.Errorf("p2 Koya = %v, want 0", vm["p2"])
	}
	// p4: vs p3(yes) drew → 0.5, vs p2(no) skip, vs p1(yes) lost → 0 → Koya = 0.5
	if vm["p4"] != 0.5 {
		t.Errorf("p4 Koya = %v, want 0.5", vm["p4"])
	}
}

func TestKoyaAllQualifying(t *testing.T) {
	// 2 players, 1 round. Threshold = 0.5. A draw means both score 0.5, both qualify.
	tb, _ := Get("koya")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
				},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0.5, Rank: 1},
		{PlayerID: "p2", Score: 0.5, Rank: 2},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// Both qualify, both drew each other → each Koya = 0.5.
	if vm["p1"] != 0.5 {
		t.Errorf("p1 Koya = %v, want 0.5", vm["p1"])
	}
	if vm["p2"] != 0.5 {
		t.Errorf("p2 Koya = %v, want 0.5", vm["p2"])
	}
}

func TestKoyaNoGames(t *testing.T) {
	tb, _ := Get("koya")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("Koya with no games = %v, want 0", values[0].Value)
	}
}

// --- Progressive tests ---

func TestProgressive(t *testing.T) {
	tb, _ := Get("progressive")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: round scores [1.0, 0.5, 1.0]
	//   cumulative: [1.0, 1.5, 2.5]
	//   progressive: 1.0 + 1.5 + 2.5 = 5.0
	if vm["p1"] != 5.0 {
		t.Errorf("p1 Progressive = %v, want 5.0", vm["p1"])
	}
	// p3: round scores [0.5, 0.5, 1.0]
	//   cumulative: [0.5, 1.0, 2.0]
	//   progressive: 0.5 + 1.0 + 2.0 = 3.5
	if vm["p3"] != 3.5 {
		t.Errorf("p3 Progressive = %v, want 3.5", vm["p3"])
	}
	// p2: round scores [0.0, 1.0, 0.0]
	//   cumulative: [0.0, 1.0, 1.0]
	//   progressive: 0.0 + 1.0 + 1.0 = 2.0
	if vm["p2"] != 2.0 {
		t.Errorf("p2 Progressive = %v, want 2.0", vm["p2"])
	}
	// p4: round scores [0.5, 0.0, 0.0]
	//   cumulative: [0.5, 0.5, 0.5]
	//   progressive: 0.5 + 0.5 + 0.5 = 1.5
	if vm["p4"] != 1.5 {
		t.Errorf("p4 Progressive = %v, want 1.5", vm["p4"])
	}
}

func TestProgressiveEarlyWinsBetter(t *testing.T) {
	// Two players with same total score but different timing.
	// p1 wins early, p2 wins late → p1 should have higher progressive.
	tb, _ := Get("progressive")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins}, // p1 wins
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultBlackWins}, // p2 loses
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultBlackWins}, // p1 loses
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins}, // p2 wins
				},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 1.0, Rank: 1},
		{PlayerID: "p2", Score: 1.0, Rank: 2},
		{PlayerID: "p3", Score: 0.0, Rank: 3},
		{PlayerID: "p4", Score: 1.0, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: [1.0, 0.0] → cumulative [1.0, 1.0] → progressive = 2.0
	// p2: [0.0, 1.0] → cumulative [0.0, 1.0] → progressive = 1.0
	if vm["p1"] != 2.0 {
		t.Errorf("p1 Progressive = %v, want 2.0", vm["p1"])
	}
	if vm["p2"] != 1.0 {
		t.Errorf("p2 Progressive = %v, want 1.0", vm["p2"])
	}
	if vm["p1"] <= vm["p2"] {
		t.Errorf("p1 (early winner) should have higher progressive than p2 (late winner): %v vs %v", vm["p1"], vm["p2"])
	}
}

func TestProgressiveWithBye(t *testing.T) {
	tb, _ := Get("progressive")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByePAB}},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 2.0, Rank: 1},
		{PlayerID: "p3", Score: 1.5, Rank: 2},
		{PlayerID: "p2", Score: 0.5, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: [1.0(win), 1.0(bye)] → cumulative [1.0, 2.0] → progressive = 3.0
	if vm["p1"] != 3.0 {
		t.Errorf("p1 Progressive = %v, want 3.0", vm["p1"])
	}
	// p3: [1.0(bye), 0.5(draw)] → cumulative [1.0, 1.5] → progressive = 2.5
	if vm["p3"] != 2.5 {
		t.Errorf("p3 Progressive = %v, want 2.5", vm["p3"])
	}
}

func TestProgressiveByeTypeDifferentiation(t *testing.T) {
	tb, _ := Get("progressive")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games:  []chesspairing.GameData{},
				Byes: []chesspairing.ByeEntry{
					{PlayerID: "p1", Type: chesspairing.ByePAB},
					{PlayerID: "p2", Type: chesspairing.ByeHalf},
					{PlayerID: "p3", Type: chesspairing.ByeZero},
					{PlayerID: "p4", Type: chesspairing.ByeAbsent},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 2.0, Rank: 1},
		{PlayerID: "p2", Score: 0.5, Rank: 2},
		{PlayerID: "p3", Score: 1.0, Rank: 3},
		{PlayerID: "p4", Score: 0.0, Rank: 4},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: [1.0(PAB), 1.0(win)] → cumulative [1.0, 2.0] → progressive = 3.0
	if vm["p1"] != 3.0 {
		t.Errorf("p1 Progressive = %v, want 3.0 (PAB=1.0)", vm["p1"])
	}
	// p2: [0.5(Half), 0.0(loss)] → cumulative [0.5, 0.5] → progressive = 1.0
	if vm["p2"] != 1.0 {
		t.Errorf("p2 Progressive = %v, want 1.0 (Half=0.5)", vm["p2"])
	}
	// p3: [0.0(Zero), 1.0(win)] → cumulative [0.0, 1.0] → progressive = 1.0
	if vm["p3"] != 1.0 {
		t.Errorf("p3 Progressive = %v, want 1.0 (Zero=0.0)", vm["p3"])
	}
	// p4: [0.0(Absent), 0.0(loss)] → cumulative [0.0, 0.0] → progressive = 0.0
	if vm["p4"] != 0.0 {
		t.Errorf("p4 Progressive = %v, want 0.0 (Absent=0.0)", vm["p4"])
	}
}

func TestProgressiveNoGames(t *testing.T) {
	tb, _ := Get("progressive")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("Progressive with no games = %v, want 0", values[0].Value)
	}
}

// --- ARO (Average Rating of Opponents) tests ---

func TestARO(t *testing.T) {
	tb, _ := Get("aro")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1 opponents: p2(1800), p3(1600), p4(1400) → ARO = (1800+1600+1400)/3 = 1600
	if vm["p1"] != 1600 {
		t.Errorf("p1 ARO = %v, want 1600", vm["p1"])
	}
	// p2 opponents: p1(2000), p4(1400), p3(1600) → ARO = (2000+1400+1600)/3 ≈ 1666.67
	expected := (2000.0 + 1400.0 + 1600.0) / 3.0
	if vm["p2"] != expected {
		t.Errorf("p2 ARO = %v, want %v", vm["p2"], expected)
	}
	// p3 opponents: p4(1400), p1(2000), p2(1800) → ARO = (1400+2000+1800)/3 ≈ 1733.33
	expected = (1400.0 + 2000.0 + 1800.0) / 3.0
	if vm["p3"] != expected {
		t.Errorf("p3 ARO = %v, want %v", vm["p3"], expected)
	}
	// p4 opponents: p3(1600), p2(1800), p1(2000) → ARO = (1600+1800+2000)/3 = 1800
	if vm["p4"] != 1800 {
		t.Errorf("p4 ARO = %v, want 1800", vm["p4"])
	}
}

func TestAROWithBye(t *testing.T) {
	// Player with bye should only average over actual opponents (not bye).
	tb, _ := Get("aro")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
				},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 1.0, Rank: 1},
		{PlayerID: "p3", Score: 1.5, Rank: 2},
		{PlayerID: "p2", Score: 0.5, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: opponent p2(1800) only (absent in round 2) → ARO = 1800
	if vm["p1"] != 1800 {
		t.Errorf("p1 ARO = %v, want 1800", vm["p1"])
	}
	// p3: opponent p2(1800) only (bye in round 1 has no opponent) → ARO = 1800
	if vm["p3"] != 1800 {
		t.Errorf("p3 ARO = %v, want 1800 (bye excluded)", vm["p3"])
	}
}

func TestARONoGames(t *testing.T) {
	tb, _ := Get("aro")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("ARO with no games = %v, want 0", values[0].Value)
	}
}

// --- BlackGames tests ---

func TestBlackGames(t *testing.T) {
	tb, _ := Get("black-games")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// Looking at standardState:
	// Round 1: p1(W) vs p2(B), p3(W) vs p4(B) → p2 +1B, p4 +1B
	// Round 2: p1(W) vs p3(B), p2(W) vs p4(B) → p3 +1B, p4 +1B
	// Round 3: p1(W) vs p4(B), p2(W) vs p3(B) → p4 +1B, p3 +1B
	//
	// p1: 0 black games (always white)
	// p2: 1 black game (round 1)
	// p3: 2 black games (rounds 2, 3)
	// p4: 3 black games (rounds 1, 2, 3)
	if vm["p1"] != 0 {
		t.Errorf("p1 BlackGames = %v, want 0", vm["p1"])
	}
	if vm["p2"] != 1 {
		t.Errorf("p2 BlackGames = %v, want 1", vm["p2"])
	}
	if vm["p3"] != 2 {
		t.Errorf("p3 BlackGames = %v, want 2", vm["p3"])
	}
	if vm["p4"] != 3 {
		t.Errorf("p4 BlackGames = %v, want 3", vm["p4"])
	}
}

func TestBlackGamesNoGames(t *testing.T) {
	tb, _ := Get("black-games")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("BlackGames with no games = %v, want 0", values[0].Value)
	}
}

func TestBlackGamesWithBye(t *testing.T) {
	// Byes should not count as black games.
	tb, _ := Get("black-games")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 1.0, Rank: 1},
		{PlayerID: "p3", Score: 1.0, Rank: 2},
		{PlayerID: "p2", Score: 0.0, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	if vm["p1"] != 0 {
		t.Errorf("p1 BlackGames = %v, want 0 (played white)", vm["p1"])
	}
	if vm["p2"] != 1 {
		t.Errorf("p2 BlackGames = %v, want 1 (played black)", vm["p2"])
	}
	if vm["p3"] != 0 {
		t.Errorf("p3 BlackGames = %v, want 0 (bye, not black)", vm["p3"])
	}
}

func TestBlackGamesExcludesForfeits(t *testing.T) {
	tb, _ := Get("black-games")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					// p3 plays Black but it's a forfeit win for White — should NOT count.
					{WhiteID: "p4", BlackID: "p3", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p3", BlackID: "p1", Result: chesspairing.ResultBlackWins}, // p1 is Black, OTB
					// p2 plays Black, double forfeit — should NOT count.
					{WhiteID: "p4", BlackID: "p2", Result: chesspairing.ResultDoubleForfeit, IsForfeit: true},
				},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 2.0, Rank: 1},
		{PlayerID: "p3", Score: 1.0, Rank: 2},
		{PlayerID: "p2", Score: 0.0, Rank: 3},
		{PlayerID: "p4", Score: 0.0, Rank: 4},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: Black in round 2 (OTB) only → 1
	if vm["p1"] != 1 {
		t.Errorf("p1 BlackGames = %v, want 1 (forfeits excluded)", vm["p1"])
	}
	// p2: Black in round 1 (OTB) + round 2 (forfeit, excluded) → 1
	if vm["p2"] != 1 {
		t.Errorf("p2 BlackGames = %v, want 1 (forfeit excluded)", vm["p2"])
	}
	// p3: Black in round 1 (forfeit, excluded) → 0
	if vm["p3"] != 0 {
		t.Errorf("p3 BlackGames = %v, want 0 (forfeit excluded)", vm["p3"])
	}
	// p4: 0 black games (always White)
	if vm["p4"] != 0 {
		t.Errorf("p4 BlackGames = %v, want 0", vm["p4"])
	}
}

// --- Win (rounds won) tests ---

func TestWin(t *testing.T) {
	tb, _ := Get("win")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// In standardState, standard scoring (1-½-0):
	// p1: R1 win(1.0), R2 draw(0.5), R3 win(1.0) — 2 rounds with win-points
	// p3: R1 draw(0.5), R2 draw(0.5), R3 win(1.0) — 1 round with win-points
	// p2: R1 loss(0.0), R2 win(1.0), R3 loss(0.0) — 1 round with win-points
	// p4: R1 draw(0.5), R2 loss(0.0), R3 loss(0.0) — 0 rounds with win-points
	if vm["p1"] != 2 {
		t.Errorf("p1 WIN = %v, want 2", vm["p1"])
	}
	if vm["p3"] != 1 {
		t.Errorf("p3 WIN = %v, want 1", vm["p3"])
	}
	if vm["p2"] != 1 {
		t.Errorf("p2 WIN = %v, want 1", vm["p2"])
	}
	if vm["p4"] != 0 {
		t.Errorf("p4 WIN = %v, want 0", vm["p4"])
	}
}

func TestWinIncludesByes(t *testing.T) {
	tb, _ := Get("win")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					// p1 wins by forfeit — counts as win-points awarded.
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p2", Type: chesspairing.ByeHalf}},
			},
		},
	}
	// Standard scoring: p1 gets 1.0 (win) + 1.0 (forfeit win) = 2.0
	// p3 gets 1.0 (PAB) + 0.0 (forfeit loss) = 1.0
	// p2 gets 0.0 (loss) + 0.5 (half-bye) = 0.5
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 2.0, Rank: 1},
		{PlayerID: "p3", Score: 1.0, Rank: 2},
		{PlayerID: "p2", Score: 0.5, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// WIN counts rounds where points == win points (1.0 in standard):
	// p1: R1 = 1.0 ✓, R2 = 1.0 ✓ (forfeit win gives full points) → 2
	// p3: R1 = 1.0 ✓ (PAB = full point), R2 = 0.0 ✗ → 1
	// p2: R1 = 0.0 ✗, R2 = 0.5 ✗ (half-bye ≠ win) → 0
	if vm["p1"] != 2 {
		t.Errorf("p1 WIN = %v, want 2 (OTB win + forfeit win)", vm["p1"])
	}
	if vm["p3"] != 1 {
		t.Errorf("p3 WIN = %v, want 1 (PAB counts as win-points)", vm["p3"])
	}
	if vm["p2"] != 0 {
		t.Errorf("p2 WIN = %v, want 0 (half-bye is not win-points)", vm["p2"])
	}
}

func TestWinNoGames(t *testing.T) {
	tb, _ := Get("win")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("WIN with no games = %v, want 0", values[0].Value)
	}
}

// --- GamesPlayed tests ---

func TestGamesPlayed(t *testing.T) {
	tb, _ := Get("games-played")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// In standardState, all 4 players play exactly 3 games (round-robin).
	for _, id := range []string{"p1", "p2", "p3", "p4"} {
		if vm[id] != 3 {
			t.Errorf("%s GamesPlayed = %v, want 3", id, vm[id])
		}
	}
}

func TestGamesPlayedWithAbsence(t *testing.T) {
	// p3 has a bye in round 1, p1 is absent in round 2 — neither counts as a game.
	tb, _ := Get("games-played")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
				},
				// p1 absent
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 1.0, Rank: 1},
		{PlayerID: "p3", Score: 1.5, Rank: 2},
		{PlayerID: "p2", Score: 0.5, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: 1 game (round 1), absent round 2
	if vm["p1"] != 1 {
		t.Errorf("p1 GamesPlayed = %v, want 1 (absent round 2)", vm["p1"])
	}
	// p2: 2 games (rounds 1 and 2)
	if vm["p2"] != 2 {
		t.Errorf("p2 GamesPlayed = %v, want 2", vm["p2"])
	}
	// p3: 1 game (round 2), bye in round 1 doesn't count
	if vm["p3"] != 1 {
		t.Errorf("p3 GamesPlayed = %v, want 1 (bye doesn't count)", vm["p3"])
	}
}

func TestGamesPlayedNoGames(t *testing.T) {
	tb, _ := Get("games-played")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("GamesPlayed with no games = %v, want 0", values[0].Value)
	}
}

func TestGamesPlayedExcludesForfeits(t *testing.T) {
	tb, _ := Get("games-played")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},                         // OTB
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true}, // forfeit
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},                           // OTB
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultDoubleForfeit, IsForfeit: true}, // double forfeit
				},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},                         // OTB
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultForfeitBlackWins, IsForfeit: true}, // forfeit
				},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 2.5, Rank: 1},
		{PlayerID: "p3", Score: 1.5, Rank: 2},
		{PlayerID: "p2", Score: 1.0, Rank: 3},
		{PlayerID: "p4", Score: 0.0, Rank: 4},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: 3 OTB games (all rounds) → 3
	if vm["p1"] != 3 {
		t.Errorf("p1 GamesPlayed = %v, want 3", vm["p1"])
	}
	// p2: R1 OTB, R2 double forfeit (excluded), R3 forfeit (excluded) → 1
	if vm["p2"] != 1 {
		t.Errorf("p2 GamesPlayed = %v, want 1 (forfeits excluded)", vm["p2"])
	}
	// p3: R1 forfeit (excluded), R2 OTB, R3 forfeit (excluded) → 1
	if vm["p3"] != 1 {
		t.Errorf("p3 GamesPlayed = %v, want 1 (forfeits excluded)", vm["p3"])
	}
	// p4: R1 forfeit (excluded), R2 double forfeit (excluded), R3 OTB → 1
	if vm["p4"] != 1 {
		t.Errorf("p4 GamesPlayed = %v, want 1 (forfeits excluded)", vm["p4"])
	}
}

// --- BlackWins tests ---

func TestBlackWins(t *testing.T) {
	tb, _ := Get("black-wins")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// standardState:
	// Round 1: p1(W) beats p2(B), p3(W) draws p4(B) → no Black wins
	// Round 2: p1(W) draws p3(B), p2(W) beats p4(B) → no Black wins
	// Round 3: p1(W) beats p4(B), p2(W) vs p3(B) → p3 wins as Black!
	// p1: 0 black wins (always White)
	// p2: 0 black wins (Black in R1 = loss)
	// p3: 1 black win (R3 as Black, beat p2)
	// p4: 0 black wins (Black in R1=draw, R2=loss, R3=loss)
	if vm["p1"] != 0 {
		t.Errorf("p1 BlackWins = %v, want 0", vm["p1"])
	}
	if vm["p2"] != 0 {
		t.Errorf("p2 BlackWins = %v, want 0", vm["p2"])
	}
	if vm["p3"] != 1 {
		t.Errorf("p3 BlackWins = %v, want 1", vm["p3"])
	}
	if vm["p4"] != 0 {
		t.Errorf("p4 BlackWins = %v, want 0", vm["p4"])
	}
}

func TestBlackWinsExcludesForfeits(t *testing.T) {
	tb, _ := Get("black-wins")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultBlackWins}, // OTB Black win
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultForfeitBlackWins, IsForfeit: true}, // forfeit, excluded
				},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p2", Score: 2.0, Rank: 1},
		{PlayerID: "p1", Score: 0.0, Rank: 2},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p2: 1 OTB Black win (R1), forfeit win (R2) excluded
	if vm["p2"] != 1 {
		t.Errorf("p2 BlackWins = %v, want 1 (forfeit excluded)", vm["p2"])
	}
}

func TestBlackWinsNoGames(t *testing.T) {
	tb, _ := Get("black-wins")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("BlackWins with no games = %v, want 0", values[0].Value)
	}
}

// --- RoundsPlayed tests ---

func TestRoundsPlayed(t *testing.T) {
	tb, _ := Get("rounds-played")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// standardState: all 4 players play all 3 rounds OTB → REP = 3 each.
	for _, id := range []string{"p1", "p2", "p3", "p4"} {
		if vm[id] != 3 {
			t.Errorf("%s REP = %v, want 3", id, vm[id])
		}
	}
}

func TestRoundsPlayedWithUnplayedRounds(t *testing.T) {
	tb, _ := Get("rounds-played")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{
					{PlayerID: "p3", Type: chesspairing.ByePAB},  // PAB counts as played
					{PlayerID: "p4", Type: chesspairing.ByeHalf}, // half-bye = unplayed
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
					// p1 vs p2 forfeit — p2 gets forfeit loss = unplayed
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
				},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 2.5, Rank: 1},
		{PlayerID: "p3", Score: 2.0, Rank: 2},
		{PlayerID: "p2", Score: 1.0, Rank: 3},
		{PlayerID: "p4", Score: 0.5, Rank: 4},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: 3 rounds, no unplayed (forfeit WIN still counts) → 3
	if vm["p1"] != 3 {
		t.Errorf("p1 REP = %v, want 3", vm["p1"])
	}
	// p2: 3 rounds, R2 = forfeit loss → 3 - 1 = 2
	if vm["p2"] != 2 {
		t.Errorf("p2 REP = %v, want 2 (forfeit loss is unplayed)", vm["p2"])
	}
	// p3: 3 rounds, R1 = PAB (counts as played) → 3
	if vm["p3"] != 3 {
		t.Errorf("p3 REP = %v, want 3 (PAB counts as played)", vm["p3"])
	}
	// p4: 3 rounds, R1 = half-bye (unplayed) → 3 - 1 = 2
	if vm["p4"] != 2 {
		t.Errorf("p4 REP = %v, want 2 (half-bye is unplayed)", vm["p4"])
	}
}

func TestRoundsPlayedWithAbsence(t *testing.T) {
	tb, _ := Get("rounds-played")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				// p3 absent (not in byes, not in games)
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
				},
				Byes: []chesspairing.ByeEntry{
					{PlayerID: "p1", Type: chesspairing.ByeZero}, // zero-bye = unplayed
				},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 1.0, Rank: 1},
		{PlayerID: "p2", Score: 0.5, Rank: 2},
		{PlayerID: "p3", Score: 0.5, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: R1 played, R2 zero-bye (unplayed) → 2 - 1 = 1
	if vm["p1"] != 1 {
		t.Errorf("p1 REP = %v, want 1 (zero-bye is unplayed)", vm["p1"])
	}
	// p3: R1 absent (unplayed), R2 played → 2 - 1 = 1
	if vm["p3"] != 1 {
		t.Errorf("p3 REP = %v, want 1 (absent is unplayed)", vm["p3"])
	}
	// p2: both rounds played → 2
	if vm["p2"] != 2 {
		t.Errorf("p2 REP = %v, want 2", vm["p2"])
	}
}

// --- StandardPoints tests ---

func TestStandardPoints(t *testing.T) {
	tb, _ := Get("standard-points")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// standardState (standard 1-½-0 scoring):
	// p1: R1 beat p2 (1.0 vs 0.0 → +1), R2 drew p3 (0.5 vs 0.5 → +0.5), R3 beat p4 (1.0 vs 0.0 → +1) = 2.5
	// p3: R1 drew p4 (0.5 vs 0.5 → +0.5), R2 drew p1 (0.5 vs 0.5 → +0.5), R3 beat p2 (1.0 vs 0.0 → +1) = 2.0
	// p2: R1 lost p1 (0 vs 1.0 → +0), R2 beat p4 (1.0 vs 0.0 → +1), R3 lost p3 (0 vs 1.0 → +0) = 1.0
	// p4: R1 drew p3 (0.5 vs 0.5 → +0.5), R2 lost p2 (0 vs 1.0 → +0), R3 lost p1 (0 vs 1.0 → +0) = 0.5
	if vm["p1"] != 2.5 {
		t.Errorf("p1 STD = %v, want 2.5", vm["p1"])
	}
	if vm["p3"] != 2.0 {
		t.Errorf("p3 STD = %v, want 2.0", vm["p3"])
	}
	if vm["p2"] != 1.0 {
		t.Errorf("p2 STD = %v, want 1.0", vm["p2"])
	}
	if vm["p4"] != 0.5 {
		t.Errorf("p4 STD = %v, want 0.5", vm["p4"])
	}
}

func TestStandardPointsWithByes(t *testing.T) {
	tb, _ := Get("standard-points")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeHalf}},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 1.5, Rank: 1},
		{PlayerID: "p3", Score: 1.5, Rank: 2},
		{PlayerID: "p2", Score: 0.5, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: R1 win (+1), R2 half-bye: 0.5 vs draw(0.5) → +0.5 = 1.5
	if vm["p1"] != 1.5 {
		t.Errorf("p1 STD = %v, want 1.5", vm["p1"])
	}
	// p3: R1 PAB: 1.0 vs draw(0.5) → +1, R2 draw: 0.5 vs 0.5 → +0.5 = 1.5
	if vm["p3"] != 1.5 {
		t.Errorf("p3 STD = %v, want 1.5", vm["p3"])
	}
	// p2: R1 loss (+0), R2 draw (+0.5) = 0.5
	if vm["p2"] != 0.5 {
		t.Errorf("p2 STD = %v, want 0.5", vm["p2"])
	}
}

// --- PairingNumber tests ---

func TestPairingNumber(t *testing.T) {
	tb, _ := Get("pairing-number")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// TPN uses the player's index position (1-based) in state.Players as the
	// pairing number. Lower TPN = higher tiebreak. We negate so higher value
	// sorts first (consistent with all other tiebreakers where higher = better).
	// Players: p1(idx=1), p2(idx=2), p3(idx=3), p4(idx=4)
	// Negated: p1=-1, p2=-2, p3=-3, p4=-4
	// So p1 > p2 > p3 > p4 in tiebreak order.
	if vm["p1"] != -1 {
		t.Errorf("p1 TPN = %v, want -1", vm["p1"])
	}
	if vm["p2"] != -2 {
		t.Errorf("p2 TPN = %v, want -2", vm["p2"])
	}
	if vm["p3"] != -3 {
		t.Errorf("p3 TPN = %v, want -3", vm["p3"])
	}
	if vm["p4"] != -4 {
		t.Errorf("p4 TPN = %v, want -4", vm["p4"])
	}
}

// --- BuchholzMedian2 tests ---

func TestBuchholzMedian2(t *testing.T) {
	tb, _ := Get("buchholz-median2")

	// Need at least 5 opponents to meaningfully test drop-2.
	// 6 players, 5 rounds.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2200, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2000, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1800, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1600, Active: true},
			{ID: "p5", DisplayName: "Eve", Rating: 1400, Active: true},
			{ID: "p6", DisplayName: "Frank", Rating: 1200, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p2", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},
			}},
			{Number: 2, Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
				{WhiteID: "p3", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p4", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
			}},
			{Number: 3, Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p5", BlackID: "p6", Result: chesspairing.ResultDraw},
			}},
			{Number: 4, Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p2", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p3", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
			}},
			{Number: 5, Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
				{WhiteID: "p4", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
			}},
		},
	}
	// Scores: p1=4.5, p2=3.5, p3=2.5, p4=2.5, p5=0.5, p6=0.5
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 4.5, Rank: 1},
		{PlayerID: "p2", Score: 3.5, Rank: 2},
		{PlayerID: "p3", Score: 2.5, Rank: 3},
		{PlayerID: "p4", Score: 2.5, Rank: 4},
		{PlayerID: "p5", Score: 0.5, Rank: 5},
		{PlayerID: "p6", Score: 0.5, Rank: 6},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1 opponents: p6(0.5), p2(3.5), p3(2.5), p4(2.5), p5(0.5)
	// Sorted: [0.5, 0.5, 2.5, 2.5, 3.5]
	// Drop 2 lowest (0.5, 0.5) and 2 highest (2.5, 3.5) → keep [2.5]
	// BH-M2 = 2.5
	if vm["p1"] != 2.5 {
		t.Errorf("p1 BH-M2 = %v, want 2.5", vm["p1"])
	}
}

func TestBuchholzMedian2FewOpponents(t *testing.T) {
	// With fewer than 5 opponents, drop as many as possible (down to 0).
	tb, _ := Get("buchholz-median2")
	state := standardState() // 3 opponents per player
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1 opponents sorted: [0.5, 1.0, 2.0] — 3 opponents.
	// Drop 2 lowest (0.5, 1.0) and 2 highest (1.0, 2.0) — but only 3 total.
	// Drop min(2, 3) = 2 from bottom, then min(2, remaining) from top.
	// After dropping 2 lowest: [2.0]. Drop min(2, 1) = 1 from top: [].
	// BH-M2 = 0 (empty).
	if vm["p1"] != 0 {
		t.Errorf("p1 BH-M2 = %v, want 0 (too few opponents)", vm["p1"])
	}
}

// --- ForeBuchholz tests ---

func TestForeBuchholz(t *testing.T) {
	// All games completed → FB should equal regular Buchholz.
	tbFB, _ := Get("fore-buchholz")
	tbBH, _ := Get("buchholz")
	state := standardState()
	scores := standardScores()

	fbValues, err := tbFB.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("FB Compute error: %v", err)
	}
	bhValues, err := tbBH.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("BH Compute error: %v", err)
	}

	fbMap := valueMap(fbValues)
	bhMap := valueMap(bhValues)

	for _, id := range []string{"p1", "p2", "p3", "p4"} {
		if fbMap[id] != bhMap[id] {
			t.Errorf("%s FB = %v, BH = %v — should be equal when all games complete", id, fbMap[id], bhMap[id])
		}
	}
}

func TestForeBuchholzWithPendingFinalRound(t *testing.T) {
	tb, _ := Get("fore-buchholz")

	// 3 rounds. Rounds 1-2 complete, round 3 pending (all draws assumed).
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultPending}, // assume draw
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultPending}, // assume draw
				},
			},
		},
	}
	// Actual scores after 2 rounds (pending round not counted by scorer):
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 1.5, Rank: 1},
		{PlayerID: "p2", Score: 1.0, Rank: 2},
		{PlayerID: "p3", Score: 1.0, Rank: 3},
		{PlayerID: "p4", Score: 0.5, Rank: 4},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// FB computes virtual scores assuming pending final-round games are draws.
	// Virtual final scores: p1=2.0, p2=1.5, p3=1.5, p4=1.0
	// p1 opponents: p2(1.5), p3(1.5), p4(1.0) → FB = 4.0
	// p2 opponents: p1(2.0), p4(1.0), p3(1.5) → FB = 4.5
	if vm["p1"] != 4.0 {
		t.Errorf("p1 FB = %v, want 4.0", vm["p1"])
	}
	if vm["p2"] != 4.5 {
		t.Errorf("p2 FB = %v, want 4.5", vm["p2"])
	}
}

// --- AverageOpponentBuchholz tests ---

func TestAverageOpponentBuchholz(t *testing.T) {
	tb, _ := Get("avg-opponent-buchholz")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// First compute each player's Buchholz (from TestBuchholzFull):
	// p1 BH = 3.5, p2 BH = 5.0, p3 BH = 4.0, p4 BH = 5.5
	//
	// AOB = average of opponents' Buchholz:
	// p1 opponents: p2(BH=5.0), p3(BH=4.0), p4(BH=5.5) → AOB = 14.5/3 ≈ 4.833...
	// p2 opponents: p1(BH=3.5), p4(BH=5.5), p3(BH=4.0) → AOB = 13.0/3 ≈ 4.333...
	// p3 opponents: p4(BH=5.5), p1(BH=3.5), p2(BH=5.0) → AOB = 14.0/3 ≈ 4.666...
	// p4 opponents: p3(BH=4.0), p2(BH=5.0), p1(BH=3.5) → AOB = 12.5/3 ≈ 4.166...
	const epsilon = 0.001
	expected := map[string]float64{
		"p1": 14.5 / 3.0,
		"p2": 13.0 / 3.0,
		"p3": 14.0 / 3.0,
		"p4": 12.5 / 3.0,
	}
	for id, want := range expected {
		got := vm[id]
		if got < want-epsilon || got > want+epsilon {
			t.Errorf("%s AOB = %v, want ~%v", id, got, want)
		}
	}
}

func TestAverageOpponentBuchholzNoGames(t *testing.T) {
	tb, _ := Get("avg-opponent-buchholz")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("AOB with no games = %v, want 0", values[0].Value)
	}
}

// --- PerformanceRating tests ---

func TestPerformanceRating(t *testing.T) {
	tb, _ := Get("performance-rating")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p1: ARO = (1800+1600+1400)/3 = 1600, p = 2.5/3 ≈ 0.8333
	// dpFromP(0.8333) → interpolated between 0.83(273) and 0.84(284):
	//   fraction = 0.3333, dp = 273 + 0.3333*11 ≈ 276.67
	// TPR = 1600 + 276.67 = 1876.67 → rounded to 1877
	if vm["p1"] != 1877 {
		t.Errorf("p1 TPR = %v, want 1877", vm["p1"])
	}

	// p3: ARO = (1400+2000+1800)/3 ≈ 1733.33, p = 2.0/3 ≈ 0.6667
	// dpFromP(0.6667) → interpolated between 0.66(117) and 0.67(125):
	//   fraction = 0.667, dp = 117 + 0.667*8 ≈ 122.33
	// TPR = 1733.33 + 122.33 = 1855.67 → rounded to 1856
	if vm["p3"] != 1856 {
		t.Errorf("p3 TPR = %v, want 1856", vm["p3"])
	}

	// p2: ARO = (2000+1400+1600)/3 ≈ 1666.67, p = 1.0/3 ≈ 0.3333
	// dpFromP(0.3333) → interpolated between 0.33(-125) and 0.34(-117):
	//   fraction = 0.333, dp = -125 + 0.333*8 ≈ -122.33
	// TPR = 1666.67 - 122.33 = 1544.33 → rounded to 1544
	if vm["p2"] != 1544 {
		t.Errorf("p2 TPR = %v, want 1544", vm["p2"])
	}

	// p4: ARO = (1600+1800+2000)/3 = 1800, p = 0.5/3 ≈ 0.1667
	// dpFromP(0.1667) → interpolated between 0.16(-284) and 0.17(-273):
	//   fraction = 0.667, dp = -284 + 0.667*11 ≈ -276.67
	// TPR = 1800 - 276.67 = 1523.33 → rounded to 1523
	if vm["p4"] != 1523 {
		t.Errorf("p4 TPR = %v, want 1523", vm["p4"])
	}
}

func TestPerformanceRatingNoGames(t *testing.T) {
	tb, _ := Get("performance-rating")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("TPR with no games = %v, want 0", values[0].Value)
	}
}

// --- PerformancePoints tests ---

func TestPerformancePoints(t *testing.T) {
	tb, _ := Get("performance-points")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// PTP: lowest rating R where sum(expectedScore(R - oppRating)) >= score.
	// p1: opponents 1800,1600,1400; score=2.5 → PTP=1921
	if vm["p1"] != 1921 {
		t.Errorf("p1 PTP = %v, want 1921", vm["p1"])
	}
	// p3: opponents 1400,2000,1800; score=2.0 → PTP=1914
	if vm["p3"] != 1914 {
		t.Errorf("p3 PTP = %v, want 1914", vm["p3"])
	}
	// p2: opponents 2000,1400,1600; score=1.0 → PTP=1486
	if vm["p2"] != 1486 {
		t.Errorf("p2 PTP = %v, want 1486", vm["p2"])
	}
	// p4: opponents 1600,1800,2000; score=0.5 → PTP=1479
	if vm["p4"] != 1479 {
		t.Errorf("p4 PTP = %v, want 1479", vm["p4"])
	}
}

func TestPerformancePointsZeroScore(t *testing.T) {
	tb, _ := Get("performance-points")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 1.0, Rank: 1},
		{PlayerID: "p2", Score: 1.0, Rank: 2},
		{PlayerID: "p3", Score: 0.0, Rank: 3},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// p3: score=0, opponents p1(2000) and p2(1800). Lowest = 1800.
	// PTP = 1800 - 800 = 1000.
	if vm["p3"] != 1000 {
		t.Errorf("p3 PTP = %v, want 1000 (lowest opp 1800 - 800)", vm["p3"])
	}
}

func TestPerformancePointsNoGames(t *testing.T) {
	tb, _ := Get("performance-points")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("PTP with no games = %v, want 0", values[0].Value)
	}
}

// --- AvgOpponentTPR tests ---

func TestAvgOpponentTPR(t *testing.T) {
	tb, _ := Get("avg-opponent-tpr")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// APRO = average of opponents' TPR values (rounded).
	// TPR: p1=1877, p2=1544, p3=1856, p4=1523
	// p1 opponents: p2(1544), p3(1856), p4(1523) → avg = 4923/3 = 1641
	if vm["p1"] != 1641 {
		t.Errorf("p1 APRO = %v, want 1641", vm["p1"])
	}
	// p3 opponents: p4(1523), p1(1877), p2(1544) → avg = 4944/3 = 1648
	if vm["p3"] != 1648 {
		t.Errorf("p3 APRO = %v, want 1648", vm["p3"])
	}
	// p2 opponents: p1(1877), p4(1523), p3(1856) → avg = 5256/3 = 1752
	if vm["p2"] != 1752 {
		t.Errorf("p2 APRO = %v, want 1752", vm["p2"])
	}
	// p4 opponents: p3(1856), p2(1544), p1(1877) → avg = 5277/3 = 1759
	if vm["p4"] != 1759 {
		t.Errorf("p4 APRO = %v, want 1759", vm["p4"])
	}
}

func TestAvgOpponentTPRNoGames(t *testing.T) {
	tb, _ := Get("avg-opponent-tpr")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("APRO with no games = %v, want 0", values[0].Value)
	}
}

// --- AvgOpponentPTP tests ---

func TestAvgOpponentPTP(t *testing.T) {
	tb, _ := Get("avg-opponent-ptp")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// APPO = average of opponents' PTP values (rounded).
	// PTP: p1=1921, p2=1486, p3=1914, p4=1479
	// p1 opponents: p2(1486), p3(1914), p4(1479) → avg = 4879/3 ≈ 1626.33 → 1626
	if vm["p1"] != 1626 {
		t.Errorf("p1 APPO = %v, want 1626", vm["p1"])
	}
	// p3 opponents: p4(1479), p1(1921), p2(1486) → avg = 4886/3 ≈ 1628.67 → 1629
	if vm["p3"] != 1629 {
		t.Errorf("p3 APPO = %v, want 1629", vm["p3"])
	}
	// p2 opponents: p1(1921), p4(1479), p3(1914) → avg = 5314/3 ≈ 1771.33 → 1771
	if vm["p2"] != 1771 {
		t.Errorf("p2 APPO = %v, want 1771", vm["p2"])
	}
	// p4 opponents: p3(1914), p2(1486), p1(1921) → avg = 5321/3 ≈ 1773.67 → 1774
	if vm["p4"] != 1774 {
		t.Errorf("p4 APPO = %v, want 1774", vm["p4"])
	}
}

func TestAvgOpponentPTPNoGames(t *testing.T) {
	tb, _ := Get("avg-opponent-ptp")
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 0, Rank: 1},
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}
	if values[0].Value != 0 {
		t.Errorf("APPO with no games = %v, want 0", values[0].Value)
	}
}

// --- PlayerRating tests ---

func TestPlayerRating(t *testing.T) {
	tb, _ := Get("player-rating")
	state := standardState()
	scores := standardScores()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	if vm["p1"] != 2000 {
		t.Errorf("p1 RTNG = %v, want 2000", vm["p1"])
	}
	if vm["p2"] != 1800 {
		t.Errorf("p2 RTNG = %v, want 1800", vm["p2"])
	}
	if vm["p3"] != 1600 {
		t.Errorf("p3 RTNG = %v, want 1600", vm["p3"])
	}
	if vm["p4"] != 1400 {
		t.Errorf("p4 RTNG = %v, want 1400", vm["p4"])
	}
}

// --- Forfeit edge cases for buildOpponentData-dependent tiebreakers ---

// forfeitState creates a scenario with both OTB and forfeit games to verify
// that forfeited games are excluded from Buchholz, SB, and Wins calculations.
//
// 4 players, 3 rounds:
//
//	Round 1: p1 beats p2 (OTB),       p3 beats p4 (OTB)
//	Round 2: p1 beats p3 (forfeit),    p2 draws p4 (OTB)
//	Round 3: p1 draws p4 (OTB),        p2 beats p3 (forfeit)
//
// OTB-only scores (for tiebreaker purposes):
//
//	p1: R1 win(1) + R2 forfeit(excluded) + R3 draw(0.5) = 1.5 OTB
//	p2: R1 loss(0) + R2 draw(0.5) + R3 forfeit(excluded) = 0.5 OTB
//	p3: R1 win(1) + R2 forfeit(excluded) + R3 forfeit(excluded) = 1.0 OTB
//	p4: R1 loss(0) + R2 draw(0.5) + R3 draw(0.5) = 1.0 OTB
//
// Tournament scores (including forfeit points):
//
//	p1: 2.5, p3: 1.0, p2: 1.5, p4: 1.0
func forfeitState() (*chesspairing.TournamentState, []chesspairing.PlayerScore) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultDraw},
				},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultDraw},
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
				},
			},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 2.5, Rank: 1},
		{PlayerID: "p2", Score: 1.5, Rank: 2},
		{PlayerID: "p3", Score: 1.0, Rank: 3},
		{PlayerID: "p4", Score: 1.0, Rank: 4},
	}
	return state, scores
}

func TestBuchholzExcludesForfeits(t *testing.T) {
	tb, _ := Get("buchholz")
	state, scores := forfeitState()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// buildOpponentData skips forfeit games. Forfeited rounds become absences,
	// and Buchholz uses virtual opponent (own score) for absent rounds.
	//
	// p1: OTB opponents p2(1.5), p4(1.0); R2 absent → virtual(2.5)
	//   → [1.0, 1.5, 2.5] → BH = 5.0
	if vm["p1"] != 5.0 {
		t.Errorf("p1 Buchholz = %v, want 5.0", vm["p1"])
	}
	// p2: OTB opponents p1(2.5), p4(1.0); R3 absent → virtual(1.5)
	//   → [1.0, 1.5, 2.5] → BH = 5.0
	if vm["p2"] != 5.0 {
		t.Errorf("p2 Buchholz = %v, want 5.0", vm["p2"])
	}
	// p3: OTB opponent p4(1.0); R2+R3 absent → 2x virtual(1.0)
	//   → [1.0, 1.0, 1.0] → BH = 3.0
	if vm["p3"] != 3.0 {
		t.Errorf("p3 Buchholz = %v, want 3.0 (forfeits become absences, virtual=own score)", vm["p3"])
	}
	// p4: all OTB: opponents p3(1.0), p2(1.5), p1(2.5) → BH = 5.0
	if vm["p4"] != 5.0 {
		t.Errorf("p4 Buchholz = %v, want 5.0", vm["p4"])
	}
}

func TestSonnebornBergerExcludesForfeits(t *testing.T) {
	tb, _ := Get("sonneborn-berger")
	state, scores := forfeitState()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// SB uses buildOpponentData (forfeit games excluded).
	// p1: OTB: beat p2(1.5)→+1.5, drew p4(1.0)→+0.5 = 2.0
	if vm["p1"] != 2.0 {
		t.Errorf("p1 SB = %v, want 2.0 (forfeit excluded)", vm["p1"])
	}
	// p2: OTB: lost p1(2.5)→0, drew p4(1.0)→0.5 = 0.5
	if vm["p2"] != 0.5 {
		t.Errorf("p2 SB = %v, want 0.5 (forfeit excluded)", vm["p2"])
	}
	// p3: OTB: beat p4(1.0)→+1.0 = 1.0
	if vm["p3"] != 1.0 {
		t.Errorf("p3 SB = %v, want 1.0 (forfeit games excluded)", vm["p3"])
	}
	// p4: OTB: lost p3(1.0)→0, drew p2(1.5)→0.75, drew p1(2.5)→1.25 = 2.0
	if vm["p4"] != 2.0 {
		t.Errorf("p4 SB = %v, want 2.0", vm["p4"])
	}
}

func TestBuchholzWithWithdrawnPlayer(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: false}, // withdrawn after round 1
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p2", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
				},
			},
		},
	}

	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 1.5, Rank: 1},
		{PlayerID: "p2", Score: 1.5, Rank: 2},
	}

	tb, _ := Get("buchholz")
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	vm := valueMap(values)

	// p1's R1 opponent (p3) is withdrawn/inactive. buildOpponentData skips games
	// involving inactive players, so p1 has 1 real opponent (p2 in R2) + 1 absence (R1).
	// For absent rounds, Buchholz uses the virtual opponent score = player's own score.
	// Buchholz(p1) = score(p2) + virtual(own=1.5) = 1.5 + 1.5 = 3.0
	if vm["p1"] != 3.0 {
		t.Errorf("p1 buchholz = %v, want 3.0 (withdrawn opponent excluded, virtual used)", vm["p1"])
	}

	// p2 had a PAB bye in R1 and played p1 in R2.
	// 1 real opponent (p1) + 1 bye round.
	// For bye rounds, Buchholz uses the virtual opponent score = player's own score.
	// Buchholz(p2) = score(p1) + virtual(own=1.5) = 1.5 + 1.5 = 3.0
	if vm["p2"] != 3.0 {
		t.Errorf("p2 buchholz = %v, want 3.0 (bye round uses virtual opponent)", vm["p2"])
	}

	t.Logf("p1=%v p2=%v", vm["p1"], vm["p2"])
}

func TestWinsExcludesForfeits(t *testing.T) {
	tb, _ := Get("wins")
	state, scores := forfeitState()

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute error: %v", err)
	}

	vm := valueMap(values)

	// Wins = OTB wins only.
	// p1: beat p2 (OTB), beat p3 (forfeit, excluded), drew p4 → 1 OTB win
	if vm["p1"] != 1 {
		t.Errorf("p1 Wins = %v, want 1 (forfeit win excluded)", vm["p1"])
	}
	// p2: lost p1, drew p4, beat p3 (forfeit, excluded) → 0 OTB wins
	if vm["p2"] != 0 {
		t.Errorf("p2 Wins = %v, want 0 (forfeit win excluded)", vm["p2"])
	}
	// p3: beat p4 (OTB) → 1 OTB win
	if vm["p3"] != 1 {
		t.Errorf("p3 Wins = %v, want 1", vm["p3"])
	}
	// p4: lost p3, drew p2, drew p1 → 0 OTB wins
	if vm["p4"] != 0 {
		t.Errorf("p4 Wins = %v, want 0", vm["p4"])
	}
}

// referenceState8Player returns an 8-player, 5-round Swiss tournament with
// all decisive results and draws, no forfeits, no byes. Every player plays
// a unique opponent each round (valid Swiss pairings).
//
// Results:
//
//	R1: p1-p8 1-0, p2-p7 1-0, p3-p6 ½-½, p4-p5 0-1
//	R2: p5-p1 0-1, p2-p3 ½-½, p6-p4 1-0, p7-p8 1-0
//	R3: p1-p6 1-0, p3-p5 ½-½, p7-p4 ½-½, p2-p8 1-0
//	R4: p4-p1 0-1, p5-p2 0-1, p3-p7 1-0, p8-p6 0-1
//	R5: p1-p3 ½-½, p6-p2 0-1, p4-p8 1-0, p5-p7 1-0
//
// Standings: p1=4.5, p2=4.5, p3=3.0, p5=2.5, p6=2.5, p4=1.5, p7=1.5, p8=0.0
func referenceState8Player() (*chesspairing.TournamentState, []chesspairing.PlayerScore) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alpha", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Beta", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "Gamma", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "Delta", Rating: 2100, Active: true},
			{ID: "p5", DisplayName: "Epsilon", Rating: 2000, Active: true},
			{ID: "p6", DisplayName: "Zeta", Rating: 1900, Active: true},
			{ID: "p7", DisplayName: "Eta", Rating: 1800, Active: true},
			{ID: "p8", DisplayName: "Theta", Rating: 1700, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p8", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p2", BlackID: "p7", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p6", Result: chesspairing.ResultDraw},
					{WhiteID: "p4", BlackID: "p5", Result: chesspairing.ResultBlackWins},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p5", BlackID: "p1", Result: chesspairing.ResultBlackWins},
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
					{WhiteID: "p6", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p7", BlackID: "p8", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p5", Result: chesspairing.ResultDraw},
					{WhiteID: "p7", BlackID: "p4", Result: chesspairing.ResultDraw},
					{WhiteID: "p2", BlackID: "p8", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 4,
				Games: []chesspairing.GameData{
					{WhiteID: "p4", BlackID: "p1", Result: chesspairing.ResultBlackWins},
					{WhiteID: "p5", BlackID: "p2", Result: chesspairing.ResultBlackWins},
					{WhiteID: "p3", BlackID: "p7", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p8", BlackID: "p6", Result: chesspairing.ResultBlackWins},
				},
			},
			{
				Number: 5,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},
					{WhiteID: "p6", BlackID: "p2", Result: chesspairing.ResultBlackWins},
					{WhiteID: "p4", BlackID: "p8", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p5", BlackID: "p7", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
	}

	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 4.5, Rank: 1},
		{PlayerID: "p2", Score: 4.5, Rank: 2},
		{PlayerID: "p3", Score: 3.0, Rank: 3},
		{PlayerID: "p5", Score: 2.5, Rank: 4},
		{PlayerID: "p6", Score: 2.5, Rank: 5},
		{PlayerID: "p4", Score: 1.5, Rank: 6},
		{PlayerID: "p7", Score: 1.5, Rank: 7},
		{PlayerID: "p8", Score: 0.0, Rank: 8},
	}

	return state, scores
}

func TestReferenceScenarioBuchholz(t *testing.T) {
	state, scores := referenceState8Player()

	tb, _ := Get("buchholz")
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	vm := valueMap(values)

	// Hand-computed Buchholz (sum of opponents' scores):
	// p1 opps: p8(0)+p5(2.5)+p6(2.5)+p4(1.5)+p3(3.0) = 9.5
	// p2 opps: p7(1.5)+p3(3.0)+p8(0)+p5(2.5)+p6(2.5) = 9.5
	// p3 opps: p6(2.5)+p2(4.5)+p5(2.5)+p7(1.5)+p1(4.5) = 15.5
	// p4 opps: p5(2.5)+p6(2.5)+p7(1.5)+p1(4.5)+p8(0) = 11.0
	// p5 opps: p4(1.5)+p1(4.5)+p3(3.0)+p2(4.5)+p7(1.5) = 15.0
	// p6 opps: p3(3.0)+p4(1.5)+p1(4.5)+p8(0)+p2(4.5) = 13.5
	// p7 opps: p2(4.5)+p8(0)+p4(1.5)+p3(3.0)+p5(2.5) = 11.5
	// p8 opps: p1(4.5)+p7(1.5)+p2(4.5)+p6(2.5)+p4(1.5) = 14.5
	expected := map[string]float64{
		"p1": 9.5, "p2": 9.5, "p3": 15.5, "p4": 11.0,
		"p5": 15.0, "p6": 13.5, "p7": 11.5, "p8": 14.5,
	}

	for id, want := range expected {
		if vm[id] != want {
			t.Errorf("%s buchholz = %v, want %v", id, vm[id], want)
		}
	}
}

func TestReferenceScenarioSonnebornBerger(t *testing.T) {
	state, scores := referenceState8Player()

	tb, _ := Get("sonneborn-berger")
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	vm := valueMap(values)

	// Hand-computed SB (win→opp score, draw→half opp score, loss→0):
	// p1: beat p8(0)→0 + beat p5(2.5)→2.5 + beat p6(2.5)→2.5 + beat p4(1.5)→1.5 + draw p3(3.0)→1.5 = 8.0
	// p2: beat p7(1.5)→1.5 + draw p3(3.0)→1.5 + beat p8(0)→0 + beat p5(2.5)→2.5 + beat p6(2.5)→2.5 = 8.0
	// p3: draw p6(2.5)→1.25 + draw p2(4.5)→2.25 + draw p5(2.5)→1.25 + beat p7(1.5)→1.5 + draw p1(4.5)→2.25 = 8.5
	// p4: lose p5→0 + lose p6→0 + draw p7(1.5)→0.75 + lose p1→0 + beat p8(0)→0 = 0.75
	// p5: beat p4(1.5)→1.5 + lose p1→0 + draw p3(3.0)→1.5 + lose p2→0 + beat p7(1.5)→1.5 = 4.5
	// p6: draw p3(3.0)→1.5 + beat p4(1.5)→1.5 + lose p1→0 + beat p8(0)→0 + lose p2→0 = 3.0
	// p7: lose p2→0 + beat p8(0)→0 + draw p4(1.5)→0.75 + lose p3→0 + lose p5→0 = 0.75
	// p8: all losses → 0
	expected := map[string]float64{
		"p1": 8.0, "p2": 8.0, "p3": 8.5, "p4": 0.75,
		"p5": 4.5, "p6": 3.0, "p7": 0.75, "p8": 0.0,
	}

	for id, want := range expected {
		if vm[id] != want {
			t.Errorf("%s SB = %v, want %v", id, vm[id], want)
		}
	}
}

func TestReferenceScenarioWins(t *testing.T) {
	state, scores := referenceState8Player()

	tb, _ := Get("wins")
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	vm := valueMap(values)

	// Hand-computed OTB wins:
	// p1: beat p8, p5, p6, p4 = 4
	// p2: beat p7, p8, p5, p6 = 4
	// p3: beat p7 = 1
	// p4: beat p8 = 1
	// p5: beat p4, p7 = 2
	// p6: beat p4, p8 = 2
	// p7: beat p8 = 1
	// p8: 0
	expected := map[string]float64{
		"p1": 4, "p2": 4, "p3": 1, "p4": 1,
		"p5": 2, "p6": 2, "p7": 1, "p8": 0,
	}

	for id, want := range expected {
		if vm[id] != want {
			t.Errorf("%s wins = %v, want %v", id, vm[id], want)
		}
	}
}

func TestReferenceScenarioProgressive(t *testing.T) {
	state, scores := referenceState8Player()

	tb, _ := Get("progressive")
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	vm := valueMap(values)

	// Hand-computed progressive (cumulative score after each round):
	// p1: 1, 2, 3, 4, 4.5 → sum = 14.5
	// p2: 1, 1.5, 2.5, 3.5, 4.5 → sum = 13.0
	// p3: 0.5, 1, 1.5, 2.5, 3.0 → sum = 8.5
	// p5: 1, 1, 1.5, 1.5, 2.5 → sum = 7.5
	// p6: 0.5, 1.5, 1.5, 2.5, 2.5 → sum = 8.5
	// p4: 0, 0, 0.5, 0.5, 1.5 → sum = 2.5
	// p7: 0, 1, 1.5, 1.5, 1.5 → sum = 5.5
	// p8: 0, 0, 0, 0, 0 → sum = 0
	expected := map[string]float64{
		"p1": 14.5, "p2": 13.0, "p3": 8.5, "p4": 2.5,
		"p5": 7.5, "p6": 8.5, "p7": 5.5, "p8": 0.0,
	}

	for id, want := range expected {
		if vm[id] != want {
			t.Errorf("%s progressive = %v, want %v", id, vm[id], want)
		}
	}
}

func TestReferenceScenarioBuchholzCut1(t *testing.T) {
	state, scores := referenceState8Player()

	tb, _ := Get("buchholz-cut1")
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	vm := valueMap(values)

	// Buchholz Cut-1: drop the lowest opponent score.
	// p1 opps sorted: 0, 1.5, 2.5, 2.5, 3.0 → drop 0 → 1.5+2.5+2.5+3.0 = 9.5
	// p2 opps sorted: 0, 1.5, 2.5, 2.5, 3.0 → drop 0 → 1.5+2.5+2.5+3.0 = 9.5
	// p3 opps sorted: 1.5, 2.5, 2.5, 4.5, 4.5 → drop 1.5 → 2.5+2.5+4.5+4.5 = 14.0
	// p4 opps sorted: 0, 1.5, 2.5, 2.5, 4.5 → drop 0 → 1.5+2.5+2.5+4.5 = 11.0
	// p5 opps sorted: 1.5, 1.5, 3.0, 4.5, 4.5 → drop 1.5 → 1.5+3.0+4.5+4.5 = 13.5
	// p6 opps sorted: 0, 1.5, 3.0, 4.5, 4.5 → drop 0 → 1.5+3.0+4.5+4.5 = 13.5
	// p7 opps sorted: 0, 1.5, 2.5, 3.0, 4.5 → drop 0 → 1.5+2.5+3.0+4.5 = 11.5
	// p8 opps sorted: 1.5, 1.5, 2.5, 4.5, 4.5 → drop 1.5 → 1.5+2.5+4.5+4.5 = 13.0
	expected := map[string]float64{
		"p1": 9.5, "p2": 9.5, "p3": 14.0, "p4": 11.0,
		"p5": 13.5, "p6": 13.5, "p7": 11.5, "p8": 13.0,
	}

	for id, want := range expected {
		if vm[id] != want {
			t.Errorf("%s buchholz-cut1 = %v, want %v", id, vm[id], want)
		}
	}
}

func TestReferenceScenarioDirectEncounter(t *testing.T) {
	state, scores := referenceState8Player()

	tb, _ := Get("direct-encounter")
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	vm := valueMap(values)

	// Direct encounter applies to tied players only.
	// Ties: p1=p2=4.5, p5=p6=2.5, p4=p7=1.5
	//
	// p1 vs p2: they didn't play each other (p1 opps: p8,p5,p6,p4,p3; p2 opps: p7,p3,p8,p5,p6)
	// → DE = 0 for both (no head-to-head)
	// p5 vs p6: they didn't play each other (p5 opps: p4,p1,p3,p2,p7; p6 opps: p3,p4,p1,p8,p2)
	// → DE = 0 for both
	// p4 vs p7: R3 draw → DE = 0.5 each
	// p3 is alone at 3.0 → DE = 0
	// p8 is alone at 0.0 → DE = 0
	expected := map[string]float64{
		"p1": 0, "p2": 0, "p3": 0, "p4": 0.5,
		"p5": 0, "p6": 0, "p7": 0.5, "p8": 0,
	}

	for id, want := range expected {
		if vm[id] != want {
			t.Errorf("%s direct-encounter = %v, want %v", id, vm[id], want)
		}
	}
}
