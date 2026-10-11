// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestBuildParticipantStatesKeepsTeamTPN(t *testing.T) {
	withdrawnAfter := 1
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", PairingNumber: 1, WithdrawnAfterRound: &withdrawnAfter},
			{ID: "t2", PairingNumber: 2},
			{ID: "t3", PairingNumber: 3},
			{ID: "t4", PairingNumber: 4, WithdrawnAfterRound: &withdrawnAfter},
			{ID: "t5", PairingNumber: 5, WithdrawnAfterRound: &withdrawnAfter},
			{ID: "t6", PairingNumber: 6},
			{ID: "t7", PairingNumber: 7, WithdrawnAfterRound: &withdrawnAfter},
			{ID: "t8", PairingNumber: 8, WithdrawnAfterRound: &withdrawnAfter},
			{ID: "t9", PairingNumber: 9, WithdrawnAfterRound: &withdrawnAfter},
			{ID: "t10", PairingNumber: 10},
		},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "t6", BlackID: "t2", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "t10", BlackID: "t3", Result: chesspairing.ResultDraw},
			},
		}},
		CurrentRound: 2,
	}

	participants, err := buildParticipantStates(state, "match")
	if err != nil {
		t.Fatal(err)
	}
	tpn := make(map[string]int, len(participants))
	for _, participant := range participants {
		tpn[participant.ID] = participant.TPN
	}
	if tpn["t2"] != 2 || tpn["t3"] != 3 || tpn["t6"] != 6 || tpn["t10"] != 10 {
		t.Fatalf("team TPNs = %v; want fixed tournament TPNs", tpn)
	}
}

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
