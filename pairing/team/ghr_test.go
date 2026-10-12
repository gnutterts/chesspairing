// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"strings"
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

// C.04.2 3.1: an adjourned match is a drawn match for pairing purposes only.
// A match with one decided board and one pending board must give both teams
// the match points of a draw (1 under the default 2-1-0 rule) in the built
// pairing score, not a full win for the leading team.
func TestPendingTeamMatchIsDrawnForPairing(t *testing.T) {
	st := &cp.TournamentState{
		Players: []cp.PlayerEntry{
			{ID: "A1", TeamID: "A", Rating: 2500},
			{ID: "A2", TeamID: "A", Rating: 2400},
			{ID: "B1", TeamID: "B", Rating: 2300},
			{ID: "B2", TeamID: "B", Rating: 2200},
		},
		Rounds: []cp.RoundData{{
			Number: 1,
			Matches: []cp.MatchData{{
				HomeID: "A", AwayID: "B",
				Boards: []cp.GameData{
					{WhiteID: "A1", BlackID: "B1", Result: cp.ResultWhiteWins},
					{WhiteID: "A2", BlackID: "B2", Result: cp.ResultPending},
				},
			}},
		}},
		CurrentRound:  2,
		PairingConfig: cp.PairingConfig{System: cp.PairingTeam},
	}
	participants, err := buildParticipantStates(st, "match")
	if err != nil {
		t.Fatal(err)
	}
	scores := make(map[string]float64, len(participants))
	for _, participant := range participants {
		scores[participant.ID] = participant.Score
	}
	if scores["A"] != 1 || scores["B"] != 1 {
		t.Errorf("adjourned match must score a draw for pairing: A=%v B=%v, want 1, 1", scores["A"], scores["B"])
	}
}

// C.04.2 3.1: an adjourned game is a draw for pairing purposes only. With
// game points as the pairing score, the pending board must still score each
// team the configured draw points (the match totals are summed from the
// boards); with match points as the pairing score, both teams receive the
// configured draw match points, as before.
func TestPendingBoardScoresGamePoints(t *testing.T) {
	draw := 0.5
	matchDraw := 1.0
	st := &cp.TournamentState{
		Players: []cp.PlayerEntry{
			{ID: "A1", TeamID: "A", Rating: 2500},
			{ID: "A2", TeamID: "A", Rating: 2400},
			{ID: "B1", TeamID: "B", Rating: 2300},
			{ID: "B2", TeamID: "B", Rating: 2200},
		},
		Rounds: []cp.RoundData{{
			Number: 1,
			Matches: []cp.MatchData{{
				HomeID: "A", AwayID: "B",
				Boards: []cp.GameData{
					{WhiteID: "A1", BlackID: "B1", Result: cp.ResultWhiteWins},
					{WhiteID: "A2", BlackID: "B2", Result: cp.ResultPending},
				},
			}},
		}},
		CurrentRound:  2,
		PairingConfig: cp.PairingConfig{System: cp.PairingTeam},
		ScoringConfig: cp.ScoringConfig{Options: map[string]any{"pointDraw": &draw, "pointMatchDraw": &matchDraw}},
	}
	wantGameHome := 1 + draw
	wantGameAway := draw
	wantMatch := matchDraw

	gameParticipants, err := buildParticipantStates(st, "game")
	if err != nil {
		t.Fatal(err)
	}
	scores := make(map[string]float64, len(gameParticipants))
	for _, participant := range gameParticipants {
		scores[participant.ID] = participant.Score
	}
	if scores["A"] != wantGameHome || scores["B"] != wantGameAway {
		t.Errorf("game-primary: A=%v B=%v, want %v, %v (board win for A, adjourned-board draw for both)", scores["A"], scores["B"], wantGameHome, wantGameAway)
	}

	matchParticipants, err := buildParticipantStates(st, "match")
	if err != nil {
		t.Fatal(err)
	}
	for _, participant := range matchParticipants {
		scores[participant.ID] = participant.Score
	}
	if scores["A"] != wantMatch || scores["B"] != wantMatch {
		t.Errorf("match-primary: A=%v B=%v, want %v, %v (drawn match)", scores["A"], scores["B"], wantMatch, wantMatch)
	}
}

// C2 with the legacy per-player-boards state shape: a team whose round-1
// boards were all forfeit-won must be rejected for a pre-assigned PAB, using
// the same team-keyed rule the pairer applies for its own PAB choice.
func TestPreAssignedPAB_LegacyBoardsShape_C2Validation(t *testing.T) {
	tr := 5
	st := &cp.TournamentState{
		Players: []cp.PlayerEntry{
			{ID: "A1", TeamID: "A", Rating: 2500},
			{ID: "A2", TeamID: "A", Rating: 2400},
			{ID: "B1", TeamID: "B", Rating: 2300},
			{ID: "B2", TeamID: "B", Rating: 2200},
		},
		Rounds: []cp.RoundData{{
			Number: 1,
			Games: []cp.GameData{
				{WhiteID: "A1", BlackID: "B1", Result: cp.ResultForfeitWhiteWins},
				{WhiteID: "A2", BlackID: "B2", Result: cp.ResultForfeitWhiteWins},
			},
		}},
		CurrentRound:    2,
		PreAssignedByes: []cp.ByeEntry{{PlayerID: "A", Type: cp.ByePAB}},
		PairingConfig:   cp.PairingConfig{System: cp.PairingTeam},
	}
	_, err := New(Options{TotalRounds: &tr}).Pair(context.Background(), st)
	if err == nil {
		t.Fatal("expected error for forfeit-winning team A receiving a pre-assigned PAB, got nil")
	}
	if !strings.Contains(err.Error(), "A") || !strings.Contains(err.Error(), "C2") {
		t.Errorf("error should name team A and C2, got %v", err)
	}
}
