// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPairPrimaryScoreChangesScoregroups(t *testing.T) {
	players := []chesspairing.PlayerEntry{{ID: "t1", Rating: 2500}, {ID: "t2", Rating: 2400}, {ID: "t3", Rating: 2300}, {ID: "t4", Rating: 2200}, {ID: "t5", Rating: 2100}, {ID: "t6", Rating: 2000}}
	round := chesspairing.RoundData{Number: 1, Games: []chesspairing.GameData{
		{WhiteID: "t1", BlackID: "t6", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "t1", BlackID: "t6", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "t1", BlackID: "t6", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "t1", BlackID: "t6", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "t2", BlackID: "t5", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "t2", BlackID: "t5", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "t2", BlackID: "t5", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "t2", BlackID: "t5", Result: chesspairing.ResultDraw},
		{WhiteID: "t3", BlackID: "t4", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "t3", BlackID: "t4", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "t3", BlackID: "t4", Result: chesspairing.ResultDraw},
		{WhiteID: "t3", BlackID: "t4", Result: chesspairing.ResultDraw},
	}}
	state := &chesspairing.TournamentState{Players: players, Rounds: []chesspairing.RoundData{round}, CurrentRound: 2}
	match, err := New(Options{PrimaryScore: stringPtr("match"), ColorPreferenceType: stringPtr("none")}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	game, err := New(Options{PrimaryScore: stringPtr("game"), ColorPreferenceType: stringPtr("none")}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if pairingsEqual(match.Pairings, game.Pairings) {
		t.Errorf("match and game primary scores produced the same pairings: %v", match.Pairings)
	}
}

func stringPtr(value string) *string { return &value }

func pairingsEqual(left, right []chesspairing.GamePairing) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
