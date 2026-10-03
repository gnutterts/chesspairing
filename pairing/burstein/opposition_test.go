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
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
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

func TestComputeOppositionIndex_ForfeitIsSelfOpponent(t *testing.T) {
	t.Parallel()

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000},
			{ID: "p2", DisplayName: "Bob", Rating: 1800},
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

	// Article 1.7.2 treats an unplayed forfeit as a game against oneself.
	player := &swisslib.PlayerState{
		ID:        "p1",
		TPN:       1,
		Score:     1.0,
		Opponents: nil, // forfeits excluded from opponent history
	}

	idx := ComputeOppositionIndex(player, state)

	if idx.Buchholz != 1 {
		t.Errorf("Buchholz: got %f, want 1", idx.Buchholz)
	}
	if idx.SonnebornBerger != 1 {
		t.Errorf("SonnebornBerger: got %f, want 1", idx.SonnebornBerger)
	}
}

func TestComputeOppositionIndex_ZeroByeSeries(t *testing.T) {
	t.Parallel()

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p"}, {ID: "opponent"}},
		Rounds: []chesspairing.RoundData{
			{Games: []chesspairing.GameData{{WhiteID: "opponent", BlackID: "p", Result: chesspairing.ResultWhiteWins}}},
			{Byes: []chesspairing.ByeEntry{{PlayerID: "p", Type: chesspairing.ByeZero}}},
			{Byes: []chesspairing.ByeEntry{{PlayerID: "p", Type: chesspairing.ByeZero}}},
		},
		CurrentRound: 4,
	}
	// Under the Article 1.7.2 interpretation, p's two current-series byes make
	// p contribute 1 to its over-the-board opponent. Opponent's absent rounds
	// are self-played at its registered score 1, totaling 3.
	index := ComputeOppositionIndex(&swisslib.PlayerState{ID: "opponent"}, state)
	if index.Buchholz != 3 {
		t.Errorf("opponent Buchholz = %v, want 3", index.Buchholz)
	}
}

func TestComputeOppositionIndex_ZeroByeSeriesStaged(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p"}, {ID: "opponent"}},
		Rounds: []chesspairing.RoundData{
			{Games: []chesspairing.GameData{{WhiteID: "opponent", BlackID: "p", Result: chesspairing.ResultWhiteWins}}},
			{Byes: []chesspairing.ByeEntry{{PlayerID: "p", Type: chesspairing.ByeZero}}},
			{Number: 3, Byes: []chesspairing.ByeEntry{{PlayerID: "p", Type: chesspairing.ByeZero}}},
		},
		CurrentRound: 3,
	}
	// The staged round is not complete, so only round 2 benefits opponent.
	// p contributes .5 and opponent's absent second round is self-played at 1.
	index := ComputeOppositionIndex(&swisslib.PlayerState{ID: "opponent"}, state)
	if index.Buchholz != 1.5 {
		t.Errorf("opponent Buchholz = %v, want 1.5", index.Buchholz)
	}
}

func TestComputeOppositionIndex_ZeroByeSeriesBrokenByGame(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p"}, {ID: "opponent"}, {ID: "other"}},
		Rounds: []chesspairing.RoundData{
			{Games: []chesspairing.GameData{{WhiteID: "opponent", BlackID: "p", Result: chesspairing.ResultWhiteWins}}},
			{Byes: []chesspairing.ByeEntry{{PlayerID: "p", Type: chesspairing.ByeZero}}},
			{Games: []chesspairing.GameData{{WhiteID: "p", BlackID: "other", Result: chesspairing.ResultDraw}}},
		},
		CurrentRound: 4,
	}
	// p's played round 3 breaks the zero-bye series; p contributes its
	// registered .5, while opponent's two absent rounds add 1+1 by self-play.
	index := ComputeOppositionIndex(&swisslib.PlayerState{ID: "opponent"}, state)
	if index.Buchholz != 2.5 {
		t.Errorf("opponent Buchholz = %v, want 2.5", index.Buchholz)
	}
}

func TestComputeOppositionIndex_MissingRoundIsSelfPlay(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players:      []chesspairing.PlayerEntry{{ID: "late", JoinedRound: 2}, {ID: "absent"}},
		Rounds:       []chesspairing.RoundData{{Number: 1, Byes: []chesspairing.ByeEntry{{PlayerID: "absent", Type: chesspairing.ByeAbsent}}}},
		CurrentRound: 2,
	}
	late := ComputeOppositionIndex(&swisslib.PlayerState{ID: "late"}, state)
	absent := ComputeOppositionIndex(&swisslib.PlayerState{ID: "absent"}, state)
	if late.Buchholz != absent.Buchholz {
		t.Errorf("late-entry Buchholz = %v, explicit absent bye = %v", late.Buchholz, absent.Buchholz)
	}
}

func TestComputeOppositionIndex_PreAssignedByeOpponentSurvivesFilter(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}},
		Rounds: []chesspairing.RoundData{
			{Games: []chesspairing.GameData{{WhiteID: "1", BlackID: "2", Result: chesspairing.ResultWhiteWins}, {WhiteID: "3", BlackID: "4", Result: chesspairing.ResultWhiteWins}}},
			{Games: []chesspairing.GameData{{WhiteID: "1", BlackID: "3", Result: chesspairing.ResultWhiteWins}, {WhiteID: "2", BlackID: "4", Result: chesspairing.ResultDraw}}},
		},
		CurrentRound:    3,
		PreAssignedByes: []chesspairing.ByeEntry{{PlayerID: "3", Type: chesspairing.ByeHalf}},
	}
	filtered, _ := swisslib.FilterPreAssignedByes(state)
	// Player 3 is absent from filtered.Players, but remains player 1's round-2
	// opponent. Its registered score is 1, so player 1's Buchholz is 0.5+1.
	if got := ComputeOppositionIndex(&swisslib.PlayerState{ID: "1"}, filtered).Buchholz; got != 1.5 {
		t.Errorf("Buchholz = %v, want 1.5", got)
	}
}

func TestComputeOppositionIndex_IgnoresStagedRound(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}, {ID: "p3"}, {ID: "p4"}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins}}},
			{Number: 2, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins}}},
			{Number: 3, Byes: []chesspairing.ByeEntry{{PlayerID: "p4", Type: chesspairing.ByeHalf}}},
		},
		CurrentRound: 3,
	}
	without := *state
	without.Rounds = without.Rounds[:2]
	for _, id := range []string{"p1", "p2", "p3", "p4"} {
		if got, want := ComputeOppositionIndex(&swisslib.PlayerState{ID: id}, state), ComputeOppositionIndex(&swisslib.PlayerState{ID: id}, &without); got != want {
			t.Errorf("%s index with staged round = %+v, without = %+v", id, got, want)
		}
	}
}

func TestRankByOppositionIndex(t *testing.T) {
	t.Parallel()

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

	// Article 1.8 does not use score for this ranking.
	expectedOrder := []string{"p3", "p4", "p1", "p2"}
	for i, id := range expectedOrder {
		if result[i].ID != id {
			t.Errorf("position %d: got %s, want %s", i, result[i].ID, id)
		}
		if result[i].TPN != players[result[i].TPN-1].TPN {
			t.Errorf("player %s: fixed TPN changed to %d", result[i].ID, result[i].TPN)
		}
	}
}

func TestRankByOppositionIndex_DifferentBuchholz(t *testing.T) {
	t.Parallel()

	// P1 and P2 both have 1.0 points, but P2 faced a stronger opponent.
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

	// Article 1.8 ranks by opposition index, not score.
	if result[0].ID != "p4" {
		t.Errorf("position 0: got %s, want p4", result[0].ID)
	}
	if result[3].ID != "p1" {
		t.Errorf("position 3: got %s, want p1", result[3].ID)
	}
}
