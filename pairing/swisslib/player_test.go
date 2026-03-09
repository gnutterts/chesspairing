package swisslib

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestBuildPlayerStates_Round1_NoHistory(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2100, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1900, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2000, Active: true},
			{ID: "p4", DisplayName: "Diana", Rating: 1800, Active: true},
		},
		Rounds:       nil,
		CurrentRound: 1,
	}

	players := BuildPlayerStates(state)

	if len(players) != 4 {
		t.Fatalf("expected 4 players, got %d", len(players))
	}

	// Players should be sorted by rating desc, then name asc.
	// TPN should be 1-based sequential after sorting.
	// Expected order: Alice(2100), Charlie(2000), Bob(1900), Diana(1800)
	wantOrder := []struct {
		id     string
		tpn    int
		score  float64
		rating int
	}{
		{"p1", 1, 0.0, 2100}, // Alice
		{"p3", 2, 0.0, 2000}, // Charlie
		{"p2", 3, 0.0, 1900}, // Bob
		{"p4", 4, 0.0, 1800}, // Diana
	}

	for i, want := range wantOrder {
		got := players[i]
		if got.ID != want.id {
			t.Errorf("position %d: want ID %s, got %s", i, want.id, got.ID)
		}
		if got.TPN != want.tpn {
			t.Errorf("player %s: want TPN %d, got %d", got.ID, want.tpn, got.TPN)
		}
		if got.Score != want.score {
			t.Errorf("player %s: want score %.1f, got %.1f", got.ID, want.score, got.Score)
		}
		if got.Rating != want.rating {
			t.Errorf("player %s: want rating %d, got %d", got.ID, want.rating, got.Rating)
		}
		if len(got.ColorHistory) != 0 {
			t.Errorf("player %s: expected empty color history, got %v", got.ID, got.ColorHistory)
		}
		if len(got.Opponents) != 0 {
			t.Errorf("player %s: expected empty opponents, got %v", got.ID, got.Opponents)
		}
		if got.ByeReceived {
			t.Errorf("player %s: expected no bye received", got.ID)
		}
	}
}

func TestBuildPlayerStates_WithdrawnExcluded(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2100, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1900, Active: false}, // withdrawn
			{ID: "p3", DisplayName: "Charlie", Rating: 2000, Active: true},
		},
		CurrentRound: 1,
	}

	players := BuildPlayerStates(state)

	if len(players) != 2 {
		t.Fatalf("expected 2 active players, got %d", len(players))
	}
	if players[0].ID != "p1" || players[1].ID != "p3" {
		t.Errorf("unexpected player order: %s, %s", players[0].ID, players[1].ID)
	}
}

func TestBuildPlayerStates_WithHistory(t *testing.T) {
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
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p2", Result: chesspairing.ResultDraw},
				},
			},
		},
		CurrentRound: 2,
	}

	players := BuildPlayerStates(state)

	// After round 1:
	// Alice: 1.0pt (white, beat Diana)
	// Charlie: 0.5pt (white, drew Bob)
	// Bob: 0.5pt (black, drew Charlie)
	// Diana: 0.0pt (black, lost to Alice)
	//
	// TPN order by score desc, then initial rank (rating desc):
	// 1. Alice (1.0, rating 2100)
	// 2. Charlie (0.5, rating 2000)
	// 3. Bob (0.5, rating 1900)
	// 4. Diana (0.0, rating 1800)

	byID := make(map[string]*PlayerState, len(players))
	for i := range players {
		byID[players[i].ID] = &players[i]
	}

	alice := byID["p1"]
	if alice.Score != 1.0 {
		t.Errorf("Alice score: want 1.0, got %.1f", alice.Score)
	}
	if len(alice.ColorHistory) != 1 || alice.ColorHistory[0] != ColorWhite {
		t.Errorf("Alice color history: want [White], got %v", alice.ColorHistory)
	}
	if len(alice.Opponents) != 1 || alice.Opponents[0] != "p4" {
		t.Errorf("Alice opponents: want [p4], got %v", alice.Opponents)
	}

	diana := byID["p4"]
	if diana.Score != 0.0 {
		t.Errorf("Diana score: want 0.0, got %.1f", diana.Score)
	}
	if len(diana.ColorHistory) != 1 || diana.ColorHistory[0] != ColorBlack {
		t.Errorf("Diana color history: want [Black], got %v", diana.ColorHistory)
	}
}

func TestBuildPlayerStates_ByeTracking(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2100, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1900, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2000, Active: true},
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
		CurrentRound: 2,
	}

	players := BuildPlayerStates(state)

	byID := make(map[string]*PlayerState, len(players))
	for i := range players {
		byID[players[i].ID] = &players[i]
	}

	charlie := byID["p3"]
	if !charlie.ByeReceived {
		t.Error("Charlie should have ByeReceived=true")
	}
	if charlie.Score != 1.0 {
		t.Errorf("Charlie (bye) score: want 1.0, got %.1f", charlie.Score)
	}
	// Bye gives ColorNone in history
	if len(charlie.ColorHistory) != 1 || charlie.ColorHistory[0] != ColorNone {
		t.Errorf("Charlie color history: want [None], got %v", charlie.ColorHistory)
	}
}

func TestBuildPlayerStates_ForfeitExcludedFromOpponents(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2100, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1900, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
				},
			},
		},
		CurrentRound: 2,
	}

	players := BuildPlayerStates(state)

	byID := make(map[string]*PlayerState, len(players))
	for i := range players {
		byID[players[i].ID] = &players[i]
	}

	// Forfeit games are excluded from opponent history (can be paired again)
	if len(byID["p1"].Opponents) != 0 {
		t.Errorf("Alice should have no opponents (forfeit excluded), got %v", byID["p1"].Opponents)
	}
	// But winner still gets the point
	if byID["p1"].Score != 1.0 {
		t.Errorf("Alice score: want 1.0 (forfeit win), got %.1f", byID["p1"].Score)
	}
}

func TestBuildPlayerStates_SameRatingTiebreak(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Zara", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Adam", Rating: 2000, Active: true},
		},
		CurrentRound: 1,
	}

	players := BuildPlayerStates(state)

	// Same rating: alphabetical by DisplayName (Adam before Zara)
	if players[0].ID != "p2" {
		t.Errorf("expected Adam (p2) first when ratings tie, got %s", players[0].ID)
	}
	if players[0].TPN != 1 || players[1].TPN != 2 {
		t.Errorf("TPN should be 1,2 got %d,%d", players[0].TPN, players[1].TPN)
	}
}
