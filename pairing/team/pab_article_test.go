// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPABUsesTeamMatchesPlayedForGamesFormHistory(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", TeamID: "A"},
			{ID: "b", TeamID: "B"},
			{ID: "c", TeamID: "C"},
			{ID: "d", TeamID: "D"},
			{ID: "e", TeamID: "E"},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "c", BlackID: "d", Result: chesspairing.ResultWhiteWins},
				},
				TeamByes: []chesspairing.ByeEntry{{PlayerID: "E", Type: chesspairing.ByeFullPoint}},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "a", BlackID: "d", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "c", BlackID: "e", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 3,
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingTeam,
		},
	}

	result, err := New(Options{}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.TeamByes) != 1 || result.TeamByes[0].PlayerID != "D" {
		t.Errorf("PAB = %v, want D: B and D are tied C2 candidates, but D played two team matches to B's one (Art. 3.4.3)", result.TeamByes)
	}
}

func TestPABLeavesLegalPairing(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Matches: []chesspairing.MatchData{
				m("a", "c", 2, 0),
				m("a", "d", 2, 0),
				m("c", "d", 2, 0),
			},
		}},
		CurrentRound: 2,
	}

	result, err := New(Options{}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.TeamByes) != 1 || result.TeamByes[0].PlayerID != "d" {
		t.Errorf("PAB = %v, want d: removing lower-ranked e leaves no C1 pairing (Art. 3.4.1)", result.TeamByes)
	}
}
