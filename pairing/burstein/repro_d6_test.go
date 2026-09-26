// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// C.04.4.2 art. 1.7.2 says: "the round shall be considered as one in
// which that player played against himself getting the result (win, draw,
// loss) that yields the same number of points as registered for the
// standings". A half-point bye is therefore a self-opponent draw, giving
// BH=own score and SB=half of own score.
func TestRepro_D6_HalfPointByeCountsAsSelfOpponent(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p", DisplayName: "P", Rating: 2200},
			{ID: "q", DisplayName: "Q", Rating: 2100},
			{ID: "r", DisplayName: "R", Rating: 2000},
		},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games:  []chesspairing.GameData{{WhiteID: "q", BlackID: "r", Result: chesspairing.ResultDraw}},
			Byes:   []chesspairing.ByeEntry{{PlayerID: "p", Type: chesspairing.ByeHalf}},
		}},
		CurrentRound: 2,
	}
	players, err := swisslib.BuildPlayerStates(state)
	if err != nil {
		t.Fatal(err)
	}
	var p swisslib.PlayerState
	for _, player := range players {
		if player.ID == "p" {
			p = player
		}
	}
	idx := ComputeOppositionIndex(&p, state)
	if idx.Buchholz != 0.5 {
		t.Fatalf("BH(P) = %v, want 0.5 (half-bye is a self-opponent draw)", idx.Buchholz)
	}
	if idx.SonnebornBerger != 0.25 {
		t.Fatalf("SB(P) = %v, want 0.25 (half point against own score)", idx.SonnebornBerger)
	}
}

// C.04.4.2 art. 1.8.1 orders equal-score players by Buchholz then
// Sonneborn-Berger. With P's half-bye counted as a self-opponent draw, all
// three players tie on score (0.5), BH (0.5) and SB (0.25); the original TPN
// then decides in favour of P, Q, R.
func TestRepro_D6_HalfPointByeChangesRanking(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players:      []chesspairing.PlayerEntry{{ID: "p", DisplayName: "P", Rating: 2200}, {ID: "q", DisplayName: "Q", Rating: 2100}, {ID: "r", DisplayName: "R", Rating: 2000}},
		Rounds:       []chesspairing.RoundData{{Number: 1, Games: []chesspairing.GameData{{WhiteID: "q", BlackID: "r", Result: chesspairing.ResultDraw}}, Byes: []chesspairing.ByeEntry{{PlayerID: "p", Type: chesspairing.ByeHalf}}}},
		CurrentRound: 2,
	}
	states, err := swisslib.BuildPlayerStates(state)
	if err != nil {
		t.Fatal(err)
	}
	ranked := RankByOppositionIndex(states, state)
	if len(ranked) != 3 {
		t.Fatalf("got %d ranked players, want 3", len(ranked))
	}
	if ranked[0].ID != "p" || ranked[1].ID != "q" || ranked[2].ID != "r" {
		t.Fatalf("ranking = %s, %s, %s; want P, Q, R", ranked[0].ID, ranked[1].ID, ranked[2].ID)
	}
}
