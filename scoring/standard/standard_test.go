package standard

import (
	"context"
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
)

func TestNew(t *testing.T) {
	s := New(Options{})
	if s == nil {
		t.Fatal("New returned nil")
	}
}

func TestNewFromMap(t *testing.T) {
	s := NewFromMap(map[string]any{
		"pointWin":  3.0,
		"pointDraw": 1.0,
	})
	if s == nil {
		t.Fatal("NewFromMap returned nil")
	}
	if *s.opts.PointWin != 3.0 {
		t.Errorf("PointWin = %v, want 3.0", *s.opts.PointWin)
	}
	if *s.opts.PointDraw != 1.0 {
		t.Errorf("PointDraw = %v, want 1.0", *s.opts.PointDraw)
	}
}

func TestScoreNoPlayers(t *testing.T) {
	s := New(Options{})
	scores, err := s.Score(context.Background(), &chesspairing.TournamentState{})
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	if len(scores) != 0 {
		t.Errorf("expected no scores, got %d", len(scores))
	}
}

func TestScoreNoRounds(t *testing.T) {
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	if len(scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(scores))
	}
	// All scores should be zero, ranked by rating.
	for _, ps := range scores {
		if ps.Score != 0 {
			t.Errorf("player %s score = %v, want 0", ps.PlayerID, ps.Score)
		}
	}
	if scores[0].PlayerID != "p1" {
		t.Errorf("rank 1 = %s, want p1 (highest rated)", scores[0].PlayerID)
	}
}

func TestScoreOneRound(t *testing.T) {
	// 4 players, 1 round: p1 beats p2, p3 draws p4.
	s := New(Options{})
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

	// p1 wins: 1.0, p2 loses: 0.0, p3 draws: 0.5, p4 draws: 0.5.
	if scoreMap["p1"].Score != 1.0 {
		t.Errorf("p1 score = %v, want 1.0", scoreMap["p1"].Score)
	}
	if scoreMap["p2"].Score != 0.0 {
		t.Errorf("p2 score = %v, want 0.0", scoreMap["p2"].Score)
	}
	if scoreMap["p3"].Score != 0.5 {
		t.Errorf("p3 score = %v, want 0.5", scoreMap["p3"].Score)
	}
	if scoreMap["p4"].Score != 0.5 {
		t.Errorf("p4 score = %v, want 0.5", scoreMap["p4"].Score)
	}
	// Ranking: p1(1.0), p3(0.5, higher rating), p4(0.5), p2(0.0).
	if scoreMap["p1"].Rank != 1 {
		t.Errorf("p1 rank = %d, want 1", scoreMap["p1"].Rank)
	}
	if scoreMap["p2"].Rank != 4 {
		t.Errorf("p2 rank = %d, want 4", scoreMap["p2"].Rank)
	}
}

func TestScoreBlackWins(t *testing.T) {
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultBlackWins},
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

	if scoreMap["p1"].Score != 0.0 {
		t.Errorf("p1 score = %v, want 0.0 (lost)", scoreMap["p1"].Score)
	}
	if scoreMap["p2"].Score != 1.0 {
		t.Errorf("p2 score = %v, want 1.0 (won)", scoreMap["p2"].Score)
	}
	if scoreMap["p2"].Rank != 1 {
		t.Errorf("p2 rank = %d, want 1", scoreMap["p2"].Rank)
	}
}

func TestScoreMultipleRounds(t *testing.T) {
	// 4 players, 3 rounds.
	s := New(Options{})
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
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins}, // p1: 1.0
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},      // p3: 0.5, p4: 0.5
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},      // p1: 0.5
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins}, // p2: 1.0
				},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins}, // p1: 1.0
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultBlackWins}, // p3: 1.0
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

	// p1: 1.0 + 0.5 + 1.0 = 2.5
	// p2: 0.0 + 1.0 + 0.0 = 1.0
	// p3: 0.5 + 0.5 + 1.0 = 2.0
	// p4: 0.5 + 0.0 + 0.0 = 0.5
	if scoreMap["p1"].Score != 2.5 {
		t.Errorf("p1 score = %v, want 2.5", scoreMap["p1"].Score)
	}
	if scoreMap["p2"].Score != 1.0 {
		t.Errorf("p2 score = %v, want 1.0", scoreMap["p2"].Score)
	}
	if scoreMap["p3"].Score != 2.0 {
		t.Errorf("p3 score = %v, want 2.0", scoreMap["p3"].Score)
	}
	if scoreMap["p4"].Score != 0.5 {
		t.Errorf("p4 score = %v, want 0.5", scoreMap["p4"].Score)
	}
	// Rankings: p1(2.5), p3(2.0), p2(1.0), p4(0.5).
	if scoreMap["p1"].Rank != 1 {
		t.Errorf("p1 rank = %d, want 1", scoreMap["p1"].Rank)
	}
	if scoreMap["p3"].Rank != 2 {
		t.Errorf("p3 rank = %d, want 2", scoreMap["p3"].Rank)
	}
	if scoreMap["p2"].Rank != 3 {
		t.Errorf("p2 rank = %d, want 3", scoreMap["p2"].Rank)
	}
	if scoreMap["p4"].Rank != 4 {
		t.Errorf("p4 rank = %d, want 4", scoreMap["p4"].Rank)
	}
}

func TestScoreBye(t *testing.T) {
	s := New(Options{})
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

	// p3 gets a full-point bye (1.0).
	if scoreMap["p3"].Score != 1.0 {
		t.Errorf("p3 bye score = %v, want 1.0", scoreMap["p3"].Score)
	}
	// p1 also 1.0 (win), so tiebreak by rating: p1 first.
	if scoreMap["p1"].Rank != 1 {
		t.Errorf("p1 rank = %d, want 1 (higher rating tiebreak)", scoreMap["p1"].Rank)
	}
	if scoreMap["p3"].Rank != 2 {
		t.Errorf("p3 rank = %d, want 2", scoreMap["p3"].Rank)
	}
}

func TestScoreAbsent(t *testing.T) {
	s := New(Options{})
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
				// p3 not in games, not in byes → absent
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

	// p3 absent: 0.0 (default).
	if scoreMap["p3"].Score != 0.0 {
		t.Errorf("p3 absent score = %v, want 0.0", scoreMap["p3"].Score)
	}
}

func TestScoreInactivePlayers(t *testing.T) {
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: false}, // withdrawn
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	if len(scores) != 1 {
		t.Fatalf("expected 1 score (inactive excluded), got %d", len(scores))
	}
	if scores[0].PlayerID != "p1" {
		t.Errorf("expected p1, got %s", scores[0].PlayerID)
	}
}

func TestScoreCustomOptions(t *testing.T) {
	// Football-style: 3 for win, 1 for draw, 0 for loss.
	win := 3.0
	draw := 1.0
	loss := 0.0
	bye := 3.0
	s := New(Options{
		PointWin:  &win,
		PointDraw: &draw,
		PointLoss: &loss,
		PointBye:  &bye,
	})
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
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
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

	if scoreMap["p1"].Score != 1.0 {
		t.Errorf("p1 draw = %v, want 1.0 (football draw)", scoreMap["p1"].Score)
	}
	if scoreMap["p2"].Score != 1.0 {
		t.Errorf("p2 draw = %v, want 1.0 (football draw)", scoreMap["p2"].Score)
	}
	if scoreMap["p3"].Score != 3.0 {
		t.Errorf("p3 bye = %v, want 3.0 (football bye)", scoreMap["p3"].Score)
	}
}

func TestScorePendingGame(t *testing.T) {
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultPending},
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

	// Pending game: no points for either player.
	if scoreMap["p1"].Score != 0.0 {
		t.Errorf("p1 score = %v, want 0.0 (pending)", scoreMap["p1"].Score)
	}
	if scoreMap["p2"].Score != 0.0 {
		t.Errorf("p2 score = %v, want 0.0 (pending)", scoreMap["p2"].Score)
	}
}

// PointsForResult tests

func TestPointsForResultWin(t *testing.T) {
	s := New(Options{})
	pts := s.PointsForResult(chesspairing.ResultWhiteWins, chesspairing.ResultContext{})
	if pts != 1.0 {
		t.Errorf("PointsForResult(win) = %v, want 1.0", pts)
	}
}

func TestPointsForResultDraw(t *testing.T) {
	s := New(Options{})
	pts := s.PointsForResult(chesspairing.ResultDraw, chesspairing.ResultContext{})
	if pts != 0.5 {
		t.Errorf("PointsForResult(draw) = %v, want 0.5", pts)
	}
}

func TestPointsForResultBye(t *testing.T) {
	s := New(Options{})
	pts := s.PointsForResult(chesspairing.ResultPending, chesspairing.ResultContext{IsBye: true})
	if pts != 1.0 {
		t.Errorf("PointsForResult(bye) = %v, want 1.0", pts)
	}
}

func TestPointsForResultAbsent(t *testing.T) {
	s := New(Options{})
	pts := s.PointsForResult(chesspairing.ResultPending, chesspairing.ResultContext{IsAbsent: true})
	if pts != 0.0 {
		t.Errorf("PointsForResult(absent) = %v, want 0.0", pts)
	}
}

func TestPointsForResultForfeitWin(t *testing.T) {
	s := New(Options{})
	pts := s.PointsForResult(chesspairing.ResultWhiteWins, chesspairing.ResultContext{IsForfeit: true})
	if pts != 1.0 {
		t.Errorf("PointsForResult(forfeit win) = %v, want 1.0", pts)
	}
}

func TestPointsForResultForfeitLoss(t *testing.T) {
	s := New(Options{})
	pts := s.PointsForResult(chesspairing.ResultPending, chesspairing.ResultContext{IsForfeit: true})
	if pts != 0.0 {
		t.Errorf("PointsForResult(forfeit loss) = %v, want 0.0", pts)
	}
}

// Options tests

func TestOptionsWithDefaults(t *testing.T) {
	o := Options{}.WithDefaults()
	if *o.PointWin != 1.0 {
		t.Errorf("PointWin = %v, want 1.0", *o.PointWin)
	}
	if *o.PointDraw != 0.5 {
		t.Errorf("PointDraw = %v, want 0.5", *o.PointDraw)
	}
	if *o.PointLoss != 0.0 {
		t.Errorf("PointLoss = %v, want 0.0", *o.PointLoss)
	}
	if *o.PointBye != 1.0 {
		t.Errorf("PointBye = %v, want 1.0", *o.PointBye)
	}
	if *o.PointForfeitWin != 1.0 {
		t.Errorf("PointForfeitWin = %v, want 1.0", *o.PointForfeitWin)
	}
	if *o.PointForfeitLoss != 0.0 {
		t.Errorf("PointForfeitLoss = %v, want 0.0", *o.PointForfeitLoss)
	}
	if *o.PointAbsent != 0.0 {
		t.Errorf("PointAbsent = %v, want 0.0", *o.PointAbsent)
	}
}

func TestOptionsWithDefaultsPreservesExplicit(t *testing.T) {
	win := 3.0
	draw := 1.0
	o := Options{
		PointWin:  &win,
		PointDraw: &draw,
	}.WithDefaults()
	if *o.PointWin != 3.0 {
		t.Errorf("PointWin = %v, want 3.0 (explicit)", *o.PointWin)
	}
	if *o.PointDraw != 1.0 {
		t.Errorf("PointDraw = %v, want 1.0 (explicit)", *o.PointDraw)
	}
	// Unset fields should get defaults.
	if *o.PointLoss != 0.0 {
		t.Errorf("PointLoss = %v, want 0.0 (default)", *o.PointLoss)
	}
}

func TestParseOptions(t *testing.T) {
	m := map[string]any{
		"pointWin":        3,
		"pointDraw":       1.0,
		"pointForfeitWin": 2.5,
		"unknownField":    "ignored",
	}
	o := ParseOptions(m)
	if o.PointWin == nil || *o.PointWin != 3.0 {
		t.Errorf("PointWin = %v, want 3.0", o.PointWin)
	}
	if o.PointDraw == nil || *o.PointDraw != 1.0 {
		t.Errorf("PointDraw = %v, want 1.0", o.PointDraw)
	}
	if o.PointForfeitWin == nil || *o.PointForfeitWin != 2.5 {
		t.Errorf("PointForfeitWin = %v, want 2.5", o.PointForfeitWin)
	}
	// Fields not in the map should remain nil.
	if o.PointLoss != nil {
		t.Errorf("PointLoss = %v, want nil (not in map)", o.PointLoss)
	}
}

func TestRatingTiebreak(t *testing.T) {
	// Two players with identical scores — higher rated should rank first.
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 1800, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 2000, Active: true},
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
	if scores[0].PlayerID != "p2" {
		t.Errorf("rank 1 = %s, want p2 (higher rated)", scores[0].PlayerID)
	}
}
