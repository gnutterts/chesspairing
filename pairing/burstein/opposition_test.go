// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestComputeOppositionIndex(t *testing.T) {
	t.Parallel()

	// Setup: 4 players, 2 completed rounds.
	// P1 beat P2, P3 beat P4 in round 1.
	// P1 beat P3, P2 beat P4 in round 2.
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
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 3,
	}

	// Scores: P1=2.0, P2=1.0, P3=1.0, P4=0.0
	player := &swisslib.PlayerState{
		ID:        "p1",
		TPN:       1,
		Score:     2.0,
		Opponents: []string{"p2", "p3"}, // played P2 and P3
	}

	idx := ComputeOppositionIndex(player, state)

	// Buchholz for P1: score(P2) + score(P3) = 1.0 + 1.0 = 2.0
	if idx.Buchholz != 2.0 {
		t.Errorf("Buchholz: got %f, want 2.0", idx.Buchholz)
	}

	// SB for P1: 1.0 * score(P2) + 1.0 * score(P3) = 1.0 * 1.0 + 1.0 * 1.0 = 2.0
	if idx.SonnebornBerger != 2.0 {
		t.Errorf("SonnebornBerger: got %f, want 2.0", idx.SonnebornBerger)
	}

	if idx.TPN != 1 {
		t.Errorf("TPN: got %d, want 1", idx.TPN)
	}
}

func TestComputeOppositionIndex_NoOpponents(t *testing.T) {
	t.Parallel()

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		},
		CurrentRound: 1,
	}

	player := &swisslib.PlayerState{
		ID:        "p1",
		TPN:       1,
		Score:     0.0,
		Opponents: nil,
	}

	idx := ComputeOppositionIndex(player, state)

	if idx.Buchholz != 0 {
		t.Errorf("Buchholz: got %f, want 0", idx.Buchholz)
	}
	if idx.SonnebornBerger != 0 {
		t.Errorf("SonnebornBerger: got %f, want 0", idx.SonnebornBerger)
	}
}

func TestComputeOppositionIndex_ExcludesForfeits(t *testing.T) {
	t.Parallel()

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
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

	// Forfeits are excluded from SB calculation but scores still count.
	player := &swisslib.PlayerState{
		ID:        "p1",
		TPN:       1,
		Score:     1.0,
		Opponents: nil, // forfeits excluded from opponent history
	}

	idx := ComputeOppositionIndex(player, state)

	// No opponents → Buchholz = 0.
	if idx.Buchholz != 0 {
		t.Errorf("Buchholz: got %f, want 0", idx.Buchholz)
	}

	// Forfeit game excluded from SB → SB = 0.
	if idx.SonnebornBerger != 0 {
		t.Errorf("SonnebornBerger: got %f, want 0", idx.SonnebornBerger)
	}
}

func TestRankByOppositionIndex(t *testing.T) {
	t.Parallel()

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
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}

	// After round 1:
	// P1: score=1.0, opponents=[p4], Buchholz=score(p4)=0.0
	// P2: score=1.0, opponents=[p3], Buchholz=score(p3)=0.0
	// P3: score=0.0, opponents=[p2], Buchholz=score(p2)=1.0
	// P4: score=0.0, opponents=[p1], Buchholz=score(p1)=1.0
	//
	// P1 and P2 have same score and same Buchholz (0.0).
	// P1 SB = 1.0 * 0.0 = 0.0
	// P2 SB = 1.0 * 0.0 = 0.0
	// Tiebreak by original TPN: P1 (TPN=1) before P2 (TPN=2).
	//
	// P3 and P4 have same score (0.0) and same Buchholz (1.0).
	// P3 SB = 0.0 * 1.0 = 0.0
	// P4 SB = 0.0 * 1.0 = 0.0
	// Tiebreak by original TPN: P3 (TPN=3) before P4 (TPN=4).

	players := []swisslib.PlayerState{
		{ID: "p1", TPN: 1, Score: 1.0, Opponents: []string{"p4"}},
		{ID: "p2", TPN: 2, Score: 1.0, Opponents: []string{"p3"}},
		{ID: "p3", TPN: 3, Score: 0.0, Opponents: []string{"p2"}},
		{ID: "p4", TPN: 4, Score: 0.0, Opponents: []string{"p1"}},
	}

	result := RankByOppositionIndex(players, state)

	// Expected order: P1, P2, P3, P4 (same as original in this symmetric case).
	expectedOrder := []string{"p1", "p2", "p3", "p4"}
	for i, id := range expectedOrder {
		if result[i].ID != id {
			t.Errorf("position %d: got %s, want %s", i, result[i].ID, id)
		}
		if result[i].TPN != i+1 {
			t.Errorf("player %s: TPN=%d, want %d", result[i].ID, result[i].TPN, i+1)
		}
	}
}

func TestRankByOppositionIndex_DifferentBuchholz(t *testing.T) {
	t.Parallel()

	// P1 and P2 both have 1.0 points, but P2 faced a stronger opponent.
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
					// P1 beats P4 (weak opponent).
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
					// P2 draws with P3 (stronger opponent).
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
				},
			},
		},
		CurrentRound: 2,
	}

	// Scores: P1=1.0, P2=0.5, P3=0.5, P4=0.0
	// P1 Buchholz = score(P4) = 0.0
	// P2 Buchholz = score(P3) = 0.5
	//
	// P2 has higher Buchholz than P1. But P2 has lower score (0.5 vs 1.0),
	// so P1 should still be ranked first (score is primary).

	players := []swisslib.PlayerState{
		{ID: "p1", TPN: 1, Score: 1.0, Opponents: []string{"p4"}},
		{ID: "p2", TPN: 2, Score: 0.5, Opponents: []string{"p3"}},
		{ID: "p3", TPN: 3, Score: 0.5, Opponents: []string{"p2"}},
		{ID: "p4", TPN: 4, Score: 0.0, Opponents: []string{"p1"}},
	}

	result := RankByOppositionIndex(players, state)

	// P1 (1.0) first, then P2 and P3 (both 0.5), then P4 (0.0).
	if result[0].ID != "p1" {
		t.Errorf("position 0: got %s, want p1", result[0].ID)
	}
	if result[3].ID != "p4" {
		t.Errorf("position 3: got %s, want p4", result[3].ID)
	}
}
