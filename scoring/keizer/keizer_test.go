package keizer

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
		"winFraction":  0.8,
		"drawFraction": 0.4,
	})
	if s == nil {
		t.Fatal("NewFromMap returned nil")
	}
	if *s.opts.WinFraction != 0.8 {
		t.Errorf("WinFraction = %v, want 0.8", *s.opts.WinFraction)
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
	// With no rounds, all scores should be zero, ranked by rating.
	for _, ps := range scores {
		if ps.Score != 0 {
			t.Errorf("player %s score = %v, want 0", ps.PlayerID, ps.Score)
		}
	}
	if scores[0].PlayerID != "p1" {
		t.Errorf("rank 1 = %s, want p1 (highest rated)", scores[0].PlayerID)
	}
	if scores[1].PlayerID != "p2" {
		t.Errorf("rank 2 = %s, want p2", scores[1].PlayerID)
	}
	if scores[2].PlayerID != "p3" {
		t.Errorf("rank 3 = %s, want p3 (lowest rated)", scores[2].PlayerID)
	}
}

func TestScoreOneRoundAllPlay(t *testing.T) {
	// 4 players, 1 round: p1 beats p2, p3 beats p4.
	// Default options: value numbers = 4,3,2,1 (by initial rating rank).
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
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	if len(scores) != 4 {
		t.Fatalf("expected 4 scores, got %d", len(scores))
	}

	// p1 should be rank 1, p3 rank 2 (beat someone), p2 and p4 at 0.
	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps
	}

	if scoreMap["p1"].Rank != 1 {
		t.Errorf("p1 rank = %d, want 1", scoreMap["p1"].Rank)
	}
	if scoreMap["p3"].Rank != 2 {
		t.Errorf("p3 rank = %d, want 2", scoreMap["p3"].Rank)
	}
	if scoreMap["p1"].Score <= scoreMap["p3"].Score {
		t.Errorf("p1 score (%v) should be > p3 score (%v)", scoreMap["p1"].Score, scoreMap["p3"].Score)
	}
	if scoreMap["p3"].Score <= 0 {
		t.Errorf("p3 score should be > 0, got %v", scoreMap["p3"].Score)
	}
	if scoreMap["p2"].Score != 0 {
		t.Errorf("p2 score = %v, want 0 (lost)", scoreMap["p2"].Score)
	}
	if scoreMap["p4"].Score != 0 {
		t.Errorf("p4 score = %v, want 0 (lost)", scoreMap["p4"].Score)
	}
}

func TestScoreDraws(t *testing.T) {
	// 2 players draw: each gets 50% of opponent's value.
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

	// Both drew. The iterative Keizer algorithm oscillates (each iteration
	// the player who was ranked lower gets more from the draw, flipping the
	// ranking). Oscillation detection averages the two iterations' scores,
	// giving both players equal scores. Tiebreak by rating puts p1 first.
	if scoreMap["p1"].Score != scoreMap["p2"].Score {
		t.Errorf("p1 score (%v) should equal p2 score (%v) — oscillation averaged",
			scoreMap["p1"].Score, scoreMap["p2"].Score)
	}
	if scoreMap["p1"].Rank != 1 {
		t.Errorf("p1 rank = %d, want 1 (higher rating tiebreak)", scoreMap["p1"].Rank)
	}
	if scoreMap["p2"].Rank != 2 {
		t.Errorf("p2 rank = %d, want 2", scoreMap["p2"].Rank)
	}
}

func TestScoreAbsentPlayer(t *testing.T) {
	// 4 players, 2 rounds. p1 beats p3, p2 beats p4 in round 1.
	// Round 2: p1 beats p2, p3 and p4 are absent.
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
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				// p3, p4 absent from round 2
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

	// Absent players should get some points (absent penalty > 0).
	if scoreMap["p3"].Score <= 0 {
		t.Errorf("absent player p3 score = %v, want > 0", scoreMap["p3"].Score)
	}
	if scoreMap["p4"].Score <= 0 {
		t.Errorf("absent player p4 score = %v, want > 0", scoreMap["p4"].Score)
	}
	// p1 won both rounds — should be ranked first.
	if scoreMap["p1"].Rank != 1 {
		t.Errorf("p1 rank = %d, want 1 (won both rounds)", scoreMap["p1"].Rank)
	}
	// p2 won round 1 and lost round 2 — should outscore at least one absent player.
	if scoreMap["p2"].Rank >= scoreMap["p3"].Rank && scoreMap["p2"].Rank >= scoreMap["p4"].Rank {
		t.Errorf("p2 rank = %d, expected to outscore at least one absent player", scoreMap["p2"].Rank)
	}
}

func TestScoreByePlayer(t *testing.T) {
	// 3 players, 1 round. p1 plays p2, p3 gets a bye.
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
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
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

	// p3 gets bye: 50% of own value (same as absent default).
	if scoreMap["p3"].Score <= 0 {
		t.Errorf("bye player p3 score = %v, want > 0", scoreMap["p3"].Score)
	}
}

func TestScoreInactivePlayersExcluded(t *testing.T) {
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
	// Custom: draws worth 40%, absence penalty 0 (no penalty for missing).
	draw := 0.4
	absent := 0.0
	s := New(Options{
		DrawFraction:          &draw,
		AbsentPenaltyFraction: &absent,
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
				// p3 absent, but penalty is 0.
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

	// p3 absent with 0 penalty → score should be 0.
	if scoreMap["p3"].Score != 0 {
		t.Errorf("p3 score = %v, want 0 (no absent penalty)", scoreMap["p3"].Score)
	}
}

func TestScoreMultipleRounds(t *testing.T) {
	// 4 players, 2 rounds. Tests that scoring accumulates across rounds
	// and rankings can shift.
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
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins}, // p2 bounces back
					{WhiteID: "p4", BlackID: "p1", Result: chesspairing.ResultBlackWins}, // p1 wins again
				},
			},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	if len(scores) != 4 {
		t.Fatalf("expected 4 scores, got %d", len(scores))
	}

	// p1 won both games — should be rank 1.
	if scores[0].PlayerID != "p1" {
		t.Errorf("rank 1 = %s, want p1 (won both games)", scores[0].PlayerID)
	}
	// All ranks should be 1-4.
	for i, ps := range scores {
		if ps.Rank != i+1 {
			t.Errorf("scores[%d].Rank = %d, want %d", i, ps.Rank, i+1)
		}
	}
}

func TestPointsForResultWin(t *testing.T) {
	s := New(Options{})
	rctx := chesspairing.ResultContext{
		OpponentValueNumber: 5,
	}
	pts := s.PointsForResult(chesspairing.ResultWhiteWins, rctx)
	if pts != 5.0 {
		t.Errorf("PointsForResult(win) = %v, want 5.0", pts)
	}
}

func TestPointsForResultDraw(t *testing.T) {
	s := New(Options{})
	rctx := chesspairing.ResultContext{
		OpponentValueNumber: 4,
	}
	pts := s.PointsForResult(chesspairing.ResultDraw, rctx)
	if pts != 2.0 {
		t.Errorf("PointsForResult(draw) = %v, want 2.0", pts)
	}
}

func TestPointsForResultAbsent(t *testing.T) {
	s := New(Options{})
	rctx := chesspairing.ResultContext{
		PlayerValueNumber: 6,
		IsAbsent:          true,
	}
	pts := s.PointsForResult(chesspairing.ResultPending, rctx)
	if pts != 3.0 {
		t.Errorf("PointsForResult(absent) = %v, want 3.0 (50%% of 6)", pts)
	}
}

func TestPointsForResultBye(t *testing.T) {
	s := New(Options{})
	rctx := chesspairing.ResultContext{
		PlayerValueNumber: 4,
		IsBye:             true,
	}
	pts := s.PointsForResult(chesspairing.ResultPending, rctx)
	if pts != 2.0 {
		t.Errorf("PointsForResult(bye) = %v, want 2.0 (50%% of 4)", pts)
	}
}

// TestOptionsWithDefaults verifies that defaults are applied correctly.
func TestOptionsWithDefaults(t *testing.T) {
	o := Options{}.WithDefaults(10)
	if *o.ValueNumberBase != 10 {
		t.Errorf("ValueNumberBase = %d, want 10", *o.ValueNumberBase)
	}
	if *o.ValueNumberStep != 1 {
		t.Errorf("ValueNumberStep = %d, want 1", *o.ValueNumberStep)
	}
	if *o.WinFraction != 1.0 {
		t.Errorf("WinFraction = %v, want 1.0", *o.WinFraction)
	}
	if *o.DrawFraction != 0.5 {
		t.Errorf("DrawFraction = %v, want 0.5", *o.DrawFraction)
	}
	if *o.LossFraction != 0.0 {
		t.Errorf("LossFraction = %v, want 0.0", *o.LossFraction)
	}
	if *o.AbsentPenaltyFraction != 0.5 {
		t.Errorf("AbsentPenaltyFraction = %v, want 0.5", *o.AbsentPenaltyFraction)
	}
}

// TestOptionsWithDefaultsPreservesExplicit verifies that explicit values are kept.
func TestOptionsWithDefaultsPreservesExplicit(t *testing.T) {
	win := 0.75
	base := 20
	o := Options{
		WinFraction:     &win,
		ValueNumberBase: &base,
	}.WithDefaults(10)
	if *o.WinFraction != 0.75 {
		t.Errorf("WinFraction = %v, want 0.75 (explicit)", *o.WinFraction)
	}
	if *o.ValueNumberBase != 20 {
		t.Errorf("ValueNumberBase = %d, want 20 (explicit, not default 10)", *o.ValueNumberBase)
	}
}

func TestValueNumber(t *testing.T) {
	base := 10
	step := 1
	o := Options{ValueNumberBase: &base, ValueNumberStep: &step}
	tests := []struct {
		rank int
		want int
	}{
		{1, 10},
		{2, 9},
		{5, 6},
		{10, 1},
	}
	for _, tt := range tests {
		got := o.ValueNumber(tt.rank)
		if got != tt.want {
			t.Errorf("ValueNumber(%d) = %d, want %d", tt.rank, got, tt.want)
		}
	}
}

func TestParseOptions(t *testing.T) {
	m := map[string]any{
		"valueNumberBase":       12,
		"absentPenaltyFraction": 0.3,
		"winFraction":           0.9,
		"unknownField":          "ignored",
	}
	o := ParseOptions(m)
	if o.ValueNumberBase == nil || *o.ValueNumberBase != 12 {
		t.Errorf("ValueNumberBase = %v, want 12", o.ValueNumberBase)
	}
	if o.AbsentPenaltyFraction == nil || *o.AbsentPenaltyFraction != 0.3 {
		t.Errorf("AbsentPenaltyFraction = %v, want 0.3", o.AbsentPenaltyFraction)
	}
	if o.WinFraction == nil || *o.WinFraction != 0.9 {
		t.Errorf("WinFraction = %v, want 0.9", o.WinFraction)
	}
	// Fields not in the map should remain nil.
	if o.DrawFraction != nil {
		t.Errorf("DrawFraction = %v, want nil (not in map)", o.DrawFraction)
	}
}
