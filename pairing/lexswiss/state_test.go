package lexswiss

import (
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
)

func TestBuildParticipantStates_BasicFourPlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "Diana", Rating: 2100, Active: true},
		},
		CurrentRound: 1,
	}

	participants := BuildParticipantStates(state)
	if len(participants) != 4 {
		t.Fatalf("expected 4 participants, got %d", len(participants))
	}

	// Should be sorted by rating desc → TPN 1=Alice, 2=Bob, 3=Charlie, 4=Diana.
	if participants[0].ID != "p1" || participants[0].TPN != 1 {
		t.Errorf("expected p1 with TPN 1, got %s with TPN %d", participants[0].ID, participants[0].TPN)
	}
	if participants[3].ID != "p4" || participants[3].TPN != 4 {
		t.Errorf("expected p4 with TPN 4, got %s with TPN %d", participants[3].ID, participants[3].TPN)
	}

	// Round 1: all scores should be 0.
	for _, p := range participants {
		if p.Score != 0 {
			t.Errorf("round 1: expected score 0 for %s, got %f", p.ID, p.Score)
		}
	}
}

func TestBuildParticipantStates_WithHistory(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, Active: true},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200, Active: true},
			{ID: "p4", DisplayName: "Diana", Rating: 2100, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p4", BlackID: "p2", Result: chesspairing.ResultBlackWins},
				},
			},
		},
		CurrentRound: 2,
	}

	participants := BuildParticipantStates(state)

	// After round 1: p1 scored 1.0, p2 scored 1.0, p3 scored 0.0, p4 scored 0.0
	// TPN ordering: score desc, then initial rank asc.
	// p1 (1.0, rank 1) → TPN 1, p2 (1.0, rank 2) → TPN 2
	// p3 (0.0, rank 3) → TPN 3, p4 (0.0, rank 4) → TPN 4
	scoreMap := make(map[string]float64)
	for _, p := range participants {
		scoreMap[p.ID] = p.Score
	}
	if scoreMap["p1"] != 1.0 {
		t.Errorf("p1 score: expected 1.0, got %f", scoreMap["p1"])
	}
	if scoreMap["p3"] != 0.0 {
		t.Errorf("p3 score: expected 0.0, got %f", scoreMap["p3"])
	}
}

func TestBuildParticipantStates_OpponentHistory(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
				},
			},
		},
		CurrentRound: 2,
	}

	participants := BuildParticipantStates(state)
	p1 := findParticipant(participants, "p1")
	if p1 == nil {
		t.Fatal("p1 not found")
	}
	if len(p1.Opponents) != 1 || p1.Opponents[0] != "p2" {
		t.Errorf("p1 opponents: expected [p2], got %v", p1.Opponents)
	}
}

func TestBuildParticipantStates_ForfeitExcludedFromOpponents(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, Active: true},
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

	participants := BuildParticipantStates(state)
	p1 := findParticipant(participants, "p1")
	if p1 == nil {
		t.Fatal("p1 not found")
	}
	// Forfeits excluded from opponent history — can be re-paired.
	if len(p1.Opponents) != 0 {
		t.Errorf("forfeit should not add opponent: got %v", p1.Opponents)
	}
	// But forfeit winner still gets points.
	if p1.Score != 1.0 {
		t.Errorf("forfeit winner score: expected 1.0, got %f", p1.Score)
	}
}

func TestBuildParticipantStates_ByeTracking(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByePAB}},
			},
		},
		CurrentRound: 2,
	}

	participants := BuildParticipantStates(state)
	if len(participants) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(participants))
	}
	if !participants[0].ByeReceived {
		t.Error("expected ByeReceived=true after PAB")
	}
}

func TestBuildParticipantStates_InactivePlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, Active: false},
		},
		CurrentRound: 1,
	}

	participants := BuildParticipantStates(state)
	if len(participants) != 1 {
		t.Fatalf("expected 1 active participant, got %d", len(participants))
	}
	if participants[0].ID != "p1" {
		t.Errorf("expected p1, got %s", participants[0].ID)
	}
}

func TestBuildParticipantStates_ColorHistory(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}

	participants := BuildParticipantStates(state)
	p1 := findParticipant(participants, "p1")
	if p1 == nil {
		t.Fatal("p1 not found")
	}
	if len(p1.ColorHistory) != 1 || p1.ColorHistory[0] != ColorWhite {
		t.Errorf("p1 color history: expected [White], got %v", p1.ColorHistory)
	}
}

// findParticipant is a test helper to locate a participant by ID.
func findParticipant(participants []ParticipantState, id string) *ParticipantState {
	for i := range participants {
		if participants[i].ID == id {
			return &participants[i]
		}
	}
	return nil
}
