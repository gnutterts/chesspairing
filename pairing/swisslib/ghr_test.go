// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

// TestForfeitUnion verifies that both Result.IsForfeit() with IsForfeit:false
// and IsForfeit:true with plain Result (TRF W) result in no colour and no opponent.
func TestForfeitUnion(t *testing.T) {
	// Case 1: Result is forfeit, IsForfeit is false
	state1 := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: false},
			},
		}},
		CurrentRound: 2,
	}
	players1, err := BuildPlayerStates(state1)
	if err != nil {
		t.Fatalf("BuildPlayerStates failed: %v", err)
	}
	byID1 := make(map[string]PlayerState)
	for _, p := range players1 {
		byID1[p.ID] = p
	}
	if len(byID1["p1"].ColorHistory) != 0 {
		t.Errorf("p1 should have no colour history, got %v", byID1["p1"].ColorHistory)
	}
	if len(byID1["p2"].ColorHistory) != 0 {
		t.Errorf("p2 should have no colour history, got %v", byID1["p2"].ColorHistory)
	}
	if len(byID1["p1"].Opponents) != 0 {
		t.Errorf("p1 should have no opponents (rematch allowed), got %v", byID1["p1"].Opponents)
	}
	if len(byID1["p2"].Opponents) != 0 {
		t.Errorf("p2 should have no opponents (rematch allowed), got %v", byID1["p2"].Opponents)
	}

	// Case 2: IsForfeit is true, Result is plain win (TRF W)
	state2 := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p3"}, {ID: "p4"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins, IsForfeit: true},
			},
		}},
		CurrentRound: 2,
	}
	players2, err := BuildPlayerStates(state2)
	if err != nil {
		t.Fatalf("BuildPlayerStates failed: %v", err)
	}
	byID2 := make(map[string]PlayerState)
	for _, p := range players2 {
		byID2[p.ID] = p
	}
	if len(byID2["p3"].ColorHistory) != 0 {
		t.Errorf("p3 should have no colour history, got %v", byID2["p3"].ColorHistory)
	}
	if len(byID2["p4"].ColorHistory) != 0 {
		t.Errorf("p4 should have no colour history, got %v", byID2["p4"].ColorHistory)
	}
	if len(byID2["p3"].Opponents) != 0 {
		t.Errorf("p3 should have no opponents (rematch allowed), got %v", byID2["p3"].Opponents)
	}
	if len(byID2["p4"].Opponents) != 0 {
		t.Errorf("p4 should have no opponents (rematch allowed), got %v", byID2["p4"].Opponents)
	}
}

// TestPendingGameScore verifies that an adjourned (pending) game awards 0.5
// pairing score to both players in swisslib.
func TestPendingGameScore(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultPending},
			},
		}},
		CurrentRound: 2,
	}
	players, err := BuildPlayerStates(state)
	if err != nil {
		t.Fatalf("BuildPlayerStates failed: %v", err)
	}
	byID := make(map[string]PlayerState)
	for _, p := range players {
		byID[p.ID] = p
	}
	if byID["p1"].Score != 0.5 {
		t.Errorf("p1 score: want 0.5, got %v", byID["p1"].Score)
	}
	if byID["p2"].Score != 0.5 {
		t.Errorf("p2 score: want 0.5, got %v", byID["p2"].Score)
	}
	// Colors and opponents should still be recorded (game was started/adjourned)
	if len(byID["p1"].ColorHistory) != 1 || byID["p1"].ColorHistory[0] != ColorWhite {
		t.Errorf("p1 color history: want [White], got %v", byID["p1"].ColorHistory)
	}
	if len(byID["p2"].ColorHistory) != 1 || byID["p2"].ColorHistory[0] != ColorBlack {
		t.Errorf("p2 color history: want [Black], got %v", byID["p2"].ColorHistory)
	}
	if len(byID["p1"].Opponents) != 1 || byID["p1"].Opponents[0] != "p2" {
		t.Errorf("p1 opponents: want [p2], got %v", byID["p1"].Opponents)
	}
}
