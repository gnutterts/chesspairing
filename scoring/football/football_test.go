package football

import (
	"context"
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

func TestNew(t *testing.T) {
	s := New(standard.Options{})
	if s == nil {
		t.Fatal("New returned nil")
	}
}

func TestNewFromMap(t *testing.T) {
	s := NewFromMap(map[string]any{
		"pointWin": 4.0,
	})
	if s == nil {
		t.Fatal("NewFromMap returned nil")
	}
}

func TestScoreDefaults(t *testing.T) {
	s := New(standard.Options{})
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
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps
	}

	// Football defaults: win=3, draw=1, loss=0, bye=3.
	if scoreMap["p1"].Score != 3.0 {
		t.Errorf("p1 win = %v, want 3.0", scoreMap["p1"].Score)
	}
	if scoreMap["p2"].Score != 0.0 {
		t.Errorf("p2 loss = %v, want 0.0", scoreMap["p2"].Score)
	}
	if scoreMap["p3"].Score != 3.0 {
		t.Errorf("p3 bye = %v, want 3.0", scoreMap["p3"].Score)
	}
}

func TestScoreDraw(t *testing.T) {
	s := New(standard.Options{})
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
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps
	}

	// Draw = 1 point each in football scoring.
	if scoreMap["p1"].Score != 1.0 {
		t.Errorf("p1 draw = %v, want 1.0", scoreMap["p1"].Score)
	}
	if scoreMap["p2"].Score != 1.0 {
		t.Errorf("p2 draw = %v, want 1.0", scoreMap["p2"].Score)
	}
}

func TestScoreCustomOverride(t *testing.T) {
	// Override win to 4 points, keep other football defaults.
	win := 4.0
	s := New(standard.Options{PointWin: &win})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps
	}

	if scoreMap["p1"].Score != 4.0 {
		t.Errorf("p1 win = %v, want 4.0 (custom override)", scoreMap["p1"].Score)
	}
}

func TestScoreMultipleRounds(t *testing.T) {
	s := New(standard.Options{})
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
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins}, // p1: 3
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},      // p3: 1, p4: 1
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},      // p1: 1, p3: 1
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultBlackWins}, // p4: 3
				},
			},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps
	}

	// p1: 3+1 = 4, p2: 0+0 = 0, p3: 1+1 = 2, p4: 1+3 = 4
	if scoreMap["p1"].Score != 4.0 {
		t.Errorf("p1 score = %v, want 4.0", scoreMap["p1"].Score)
	}
	if scoreMap["p2"].Score != 0.0 {
		t.Errorf("p2 score = %v, want 0.0", scoreMap["p2"].Score)
	}
	if scoreMap["p3"].Score != 2.0 {
		t.Errorf("p3 score = %v, want 2.0", scoreMap["p3"].Score)
	}
	if scoreMap["p4"].Score != 4.0 {
		t.Errorf("p4 score = %v, want 4.0", scoreMap["p4"].Score)
	}
}

// PointsForResult tests

func TestPointsForResultWin(t *testing.T) {
	s := New(standard.Options{})
	pts := s.PointsForResult(chesspairing.ResultWhiteWins, chesspairing.ResultContext{})
	if pts != 3.0 {
		t.Errorf("PointsForResult(win) = %v, want 3.0", pts)
	}
}

func TestPointsForResultDraw(t *testing.T) {
	s := New(standard.Options{})
	pts := s.PointsForResult(chesspairing.ResultDraw, chesspairing.ResultContext{})
	if pts != 1.0 {
		t.Errorf("PointsForResult(draw) = %v, want 1.0", pts)
	}
}

func TestPointsForResultBye(t *testing.T) {
	s := New(standard.Options{})
	pts := s.PointsForResult(chesspairing.ResultPending, chesspairing.ResultContext{IsBye: true})
	if pts != 3.0 {
		t.Errorf("PointsForResult(bye) = %v, want 3.0", pts)
	}
}

func TestPointsForResultAbsent(t *testing.T) {
	s := New(standard.Options{})
	pts := s.PointsForResult(chesspairing.ResultPending, chesspairing.ResultContext{IsAbsent: true})
	if pts != 0.0 {
		t.Errorf("PointsForResult(absent) = %v, want 0.0", pts)
	}
}

func TestPointsForResultForfeitWin(t *testing.T) {
	s := New(standard.Options{})
	pts := s.PointsForResult(chesspairing.ResultWhiteWins, chesspairing.ResultContext{IsForfeit: true})
	if pts != 3.0 {
		t.Errorf("PointsForResult(forfeit win) = %v, want 3.0", pts)
	}
}

func TestPointsForResultForfeitLoss(t *testing.T) {
	s := New(standard.Options{})
	pts := s.PointsForResult(chesspairing.ResultPending, chesspairing.ResultContext{IsForfeit: true})
	if pts != 0.0 {
		t.Errorf("PointsForResult(forfeit loss) = %v, want 0.0", pts)
	}
}
