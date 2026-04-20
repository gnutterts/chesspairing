// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package standard

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
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
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
			{ID: "p3", DisplayName: "Carol", Rating: 1600},
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
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
			{ID: "p3", DisplayName: "Carol", Rating: 1600},
			{ID: "p4", DisplayName: "Dave", Rating: 1400},
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
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
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
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
			{ID: "p3", DisplayName: "Carol", Rating: 1600},
			{ID: "p4", DisplayName: "Dave", Rating: 1400},
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
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
			{ID: "p3", DisplayName: "Carol", Rating: 1600},
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
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
			{ID: "p3", DisplayName: "Carol", Rating: 1600},
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
	withdrawnAfter := 1
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, WithdrawnAfterRound: &withdrawnAfter},
		},
		CurrentRound: 2,
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
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
			{ID: "p3", DisplayName: "Carol", Rating: 1600},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
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
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
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

func TestScoreByeTypeDifferentiation(t *testing.T) {
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
			{ID: "p3", DisplayName: "Carol", Rating: 1600},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Byes: []chesspairing.ByeEntry{
					{PlayerID: "p1", Type: chesspairing.ByePAB},
					{PlayerID: "p2", Type: chesspairing.ByeHalf},
					{PlayerID: "p3", Type: chesspairing.ByeZero},
				},
			},
		},
	}

	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]float64)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps.Score
	}

	// PAB should get full point (1.0).
	if scoreMap["p1"] != 1.0 {
		t.Errorf("p1 (PAB) score = %v, want 1.0", scoreMap["p1"])
	}

	// Half-bye should get 0.5.
	if scoreMap["p2"] != 0.5 {
		t.Errorf("p2 (half-bye) score = %v, want 0.5", scoreMap["p2"])
	}

	// Zero-bye should get 0.0.
	if scoreMap["p3"] != 0.0 {
		t.Errorf("p3 (zero-bye) score = %v, want 0.0", scoreMap["p3"])
	}
}

func TestScoreSingleForfeit(t *testing.T) {
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{
						WhiteID:   "p1",
						BlackID:   "p2",
						Result:    chesspairing.ResultForfeitWhiteWins,
						IsForfeit: true,
					},
				},
			},
		},
	}

	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]float64)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps.Score
	}

	// Default PointForfeitWin = 1.0, PointForfeitLoss = 0.0.
	if scoreMap["p1"] != 1.0 {
		t.Errorf("p1 (forfeit winner) score = %v, want 1.0", scoreMap["p1"])
	}
	if scoreMap["p2"] != 0.0 {
		t.Errorf("p2 (forfeit loser) score = %v, want 0.0", scoreMap["p2"])
	}
}

func TestScoreDoubleForfeit(t *testing.T) {
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
			{ID: "p3", DisplayName: "Carol", Rating: 1600},
			{ID: "p4", DisplayName: "Dave", Rating: 1400},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{
						WhiteID:   "p3",
						BlackID:   "p4",
						Result:    chesspairing.ResultDoubleForfeit,
						IsForfeit: true,
					},
				},
			},
		},
	}

	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]float64)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps.Score
	}

	// Double forfeit: neither player gets points.
	if scoreMap["p3"] != 0.0 {
		t.Errorf("p3 (double forfeit) score = %v, want 0.0", scoreMap["p3"])
	}
	if scoreMap["p4"] != 0.0 {
		t.Errorf("p4 (double forfeit) score = %v, want 0.0", scoreMap["p4"])
	}

	// Normal game scored correctly.
	if scoreMap["p1"] != 1.0 {
		t.Errorf("p1 score = %v, want 1.0", scoreMap["p1"])
	}
	if scoreMap["p2"] != 0.0 {
		t.Errorf("p2 score = %v, want 0.0", scoreMap["p2"])
	}
}

func TestScoreAbsencePenalty(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
			{ID: "p3", DisplayName: "Carol", Rating: 1600},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				// p3 is active but has no game and no bye → absent.
			},
		},
	}

	// Default PointAbsent = 0.0.
	scorer := New(Options{})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]float64)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps.Score
	}

	if scoreMap["p3"] != 0.0 {
		t.Errorf("p3 (absent, default penalty) score = %v, want 0.0", scoreMap["p3"])
	}

	// Now test with a custom absent penalty (e.g., -0.5).
	penalty := -0.5
	scorer2 := New(Options{PointAbsent: &penalty})
	scores2, err := scorer2.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap2 := make(map[string]float64)
	for _, ps := range scores2 {
		scoreMap2[ps.PlayerID] = ps.Score
	}

	if scoreMap2["p3"] != -0.5 {
		t.Errorf("p3 (absent, -0.5 penalty) score = %v, want -0.5", scoreMap2["p3"])
	}
}

func TestRatingTiebreak(t *testing.T) {
	// Two players with identical scores — higher rated should rank first.
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 1800},
			{ID: "p2", DisplayName: "Bob", Rating: 2000},
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

// TestAllByeTypes enumerates every ByeType value and asserts the
// scorer awards the expected points with default options. This is a
// sentinel test: when a new ByeType is added, this test will fail
// until the scorer's switch is extended.
func TestAllByeTypes(t *testing.T) {
	cases := []struct {
		name string
		typ  chesspairing.ByeType
		want float64
	}{
		{"PAB", chesspairing.ByePAB, 1.0},
		{"Half", chesspairing.ByeHalf, 0.5},
		{"Zero", chesspairing.ByeZero, 0.0},
		{"Absent", chesspairing.ByeAbsent, 0.0},
		{"Excused", chesspairing.ByeExcused, 0.0},
		{"ClubCommitment", chesspairing.ByeClubCommitment, 0.0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "p1", DisplayName: "Alice", Rating: 1800},
				},
				Rounds: []chesspairing.RoundData{
					{
						Number: 1,
						Byes:   []chesspairing.ByeEntry{{PlayerID: "p1", Type: c.typ}},
					},
				},
			}
			s := New(Options{})
			scores, err := s.Score(context.Background(), state)
			if err != nil {
				t.Fatalf("Score: %v", err)
			}
			if len(scores) != 1 {
				t.Fatalf("got %d scores, want 1", len(scores))
			}
			if scores[0].Score != c.want {
				t.Errorf("score for %s bye = %v, want %v", c.name, scores[0].Score, c.want)
			}
		})
	}

	// Verify the test covers every value the type can take. If a new
	// bye type is added without updating this test, the count check
	// catches it.
	covered := len(cases)
	expected := int(chesspairing.ByeClubCommitment) - int(chesspairing.ByePAB) + 1
	if covered != expected {
		t.Errorf("test covers %d bye types, expected %d (a new ByeType was added without updating TestAllByeTypes)", covered, expected)
	}
}

// TestExcusedByeConfigurable verifies that PointExcused overrides the
// default 0.0 award for ByeExcused.
func TestExcusedByeConfigurable(t *testing.T) {
	half := 0.5
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 1800},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeExcused}},
			},
		},
	}
	s := New(Options{PointExcused: &half})
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	if scores[0].Score != 0.5 {
		t.Errorf("excused bye with PointExcused=0.5 = %v, want 0.5", scores[0].Score)
	}
}

// TestClubCommitmentByeConfigurable verifies that PointClubCommitment
// overrides the default 0.0 award.
func TestClubCommitmentByeConfigurable(t *testing.T) {
	half := 0.5
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 1800},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeClubCommitment}},
			},
		},
	}
	s := New(Options{PointClubCommitment: &half})
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	if scores[0].Score != 0.5 {
		t.Errorf("club-commitment bye with PointClubCommitment=0.5 = %v, want 0.5", scores[0].Score)
	}
}

// TestParseOptionsExcusedClubCommitment verifies that the new keys
// are honored by ParseOptions.
func TestParseOptionsExcusedClubCommitment(t *testing.T) {
	o := ParseOptions(map[string]any{
		"pointExcused":        0.5,
		"pointClubCommitment": 0.25,
	})
	if o.PointExcused == nil || *o.PointExcused != 0.5 {
		t.Errorf("PointExcused = %v, want 0.5", o.PointExcused)
	}
	if o.PointClubCommitment == nil || *o.PointClubCommitment != 0.25 {
		t.Errorf("PointClubCommitment = %v, want 0.25", o.PointClubCommitment)
	}
}
