// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPreAssignedPAB_C2Validation(t *testing.T) {
	pairer := New(Options{})

	// 1. Second PAB rejected with an error
	stateSecondPAB := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "B", BlackID: "C", Result: chesspairing.ResultDraw},
			},
			Byes: []chesspairing.ByeEntry{{PlayerID: "A", Type: chesspairing.ByePAB}},
		}},
		CurrentRound: 2,
		PreAssignedByes: []chesspairing.ByeEntry{
			{PlayerID: "A", Type: chesspairing.ByePAB},
		},
	}
	_, err := pairer.Pair(context.Background(), stateSecondPAB)
	if err == nil {
		t.Fatal("expected error for second PAB on player A, got nil")
	}
	if !strings.Contains(err.Error(), "A") || !strings.Contains(err.Error(), "C2") {
		t.Errorf("error should name player A and C2, got %v", err)
	}

	// 2. Player with forfeit win rejected
	stateForfeitWin := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultForfeitWhiteWins},
			},
			Byes: []chesspairing.ByeEntry{{PlayerID: "C", Type: chesspairing.ByeHalf}},
		}},
		CurrentRound: 2,
		PreAssignedByes: []chesspairing.ByeEntry{
			{PlayerID: "A", Type: chesspairing.ByePAB},
		},
	}
	_, err = pairer.Pair(context.Background(), stateForfeitWin)
	if err == nil {
		t.Fatal("expected error for forfeit winner A receiving pre-assigned PAB, got nil")
	}
	if !strings.Contains(err.Error(), "A") || !strings.Contains(err.Error(), "C2") {
		t.Errorf("error should name player A and C2, got %v", err)
	}

	// 3. Player without PAB history accepted
	stateValid := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultDraw},
			},
			Byes: []chesspairing.ByeEntry{{PlayerID: "C", Type: chesspairing.ByeZero}},
		}},
		CurrentRound: 2,
		PreAssignedByes: []chesspairing.ByeEntry{
			{PlayerID: "A", Type: chesspairing.ByePAB},
		},
	}
	res, err := pairer.Pair(context.Background(), stateValid)
	if err != nil {
		t.Fatalf("expected player A without PAB history to be accepted, got error: %v", err)
	}
	if len(res.Byes) != 1 || res.Byes[0].PlayerID != "A" || res.Byes[0].Type != chesspairing.ByePAB {
		t.Errorf("unexpected byes: %v", res.Byes)
	}

	// 4. Pre-assigned half-point bye for a PAB holder still accepted
	stateHalfBye := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"}, {ID: "E"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "B", BlackID: "C", Result: chesspairing.ResultDraw},
				{WhiteID: "D", BlackID: "E", Result: chesspairing.ResultDraw},
			},
			Byes: []chesspairing.ByeEntry{{PlayerID: "A", Type: chesspairing.ByePAB}},
		}},
		CurrentRound: 2,
		PreAssignedByes: []chesspairing.ByeEntry{
			{PlayerID: "A", Type: chesspairing.ByeHalf},
		},
	}
	res, err = pairer.Pair(context.Background(), stateHalfBye)
	if err != nil {
		t.Fatalf("expected pre-assigned half-point bye for PAB holder to be accepted, got error: %v", err)
	}
	if len(res.Byes) != 1 || res.Byes[0].PlayerID != "A" || res.Byes[0].Type != chesspairing.ByeHalf {
		t.Errorf("unexpected byes: %v", res.Byes)
	}
}

// TestForfeitUnionInDutch verifies Dutch state respects forfeit union:
// both a forfeit result with IsForfeit: false and the TRF-W shape (plain win
// with IsForfeit: true) leave the encounter unplayed, so the rematch is
// allowed. PairingConfig.System is set to PairingDutch, as the rematch
// validation (C.04.2 3.5) only forbids repeats for a configured system.
func TestForfeitUnionInDutch(t *testing.T) {
	tests := []struct {
		name string
		game chesspairing.GameData
	}{
		{
			name: "forfeit result without the flag",
			game: chesspairing.GameData{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: false},
		},
		{
			name: "flag with a plain win (TRF W)",
			game: chesspairing.GameData{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins, IsForfeit: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}},
				Rounds: []chesspairing.RoundData{{
					Number: 1,
					Games:  []chesspairing.GameData{tt.game},
				}},
				CurrentRound:  2,
				PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDutch},
			}
			pairer := New(Options{})
			res, err := pairer.Pair(context.Background(), state)
			if err != nil {
				t.Fatalf("Pair failed: %v", err)
			}
			// Since round 1 was a forfeit, rematch between p1 and p2 is allowed!
			if len(res.Pairings) != 1 {
				t.Fatalf("expected 1 pairing (rematch allowed), got %d", len(res.Pairings))
			}
			if (res.Pairings[0].WhiteID != "p1" || res.Pairings[0].BlackID != "p2") &&
				(res.Pairings[0].WhiteID != "p2" || res.Pairings[0].BlackID != "p1") {
				t.Errorf("expected the p1/p2 rematch, got %v", res.Pairings[0])
			}
		})
	}
}
