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
	ids := []string{"buchholz", "buchholz-cut1", "buchholz-cut2", "buchholz-median",
		"sonneborn-berger", "direct-encounter", "wins",
		"koya", "progressive", "aro", "black-games", "games-played"}
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
	if len(all) < 12 {
		t.Errorf("expected at least 12 registered tiebreakers, got %d", len(all))
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
				Byes: []string{"p3"},
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
				Byes: []string{"p3"},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw}, // p1 +0.5, p3 +0.5
				},
				Byes: []string{"p2"},
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
				Byes: []string{"p3"},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins}, // p2 beats p3
				},
				Byes: []string{"p1"},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p3", BlackID: "p1", Result: chesspairing.ResultWhiteWins}, // p3 beats p1
				},
				Byes: []string{"p2"},
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
				Byes: []string{"p3"},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
				},
				Byes: []string{"p1"},
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
				Byes: []string{"p3"},
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
				Byes: []string{"p3"},
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
				Byes: []string{"p3"},
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
