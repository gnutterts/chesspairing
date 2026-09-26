// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// With byePolicy="lowest-without-bye", a pairing-allocated bye should not be
// given twice in one season when a player without a prior bye is available.
func TestRepro_KZP03_PABByeIsNotRepeated(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "A", Rating: 2500},
			{ID: "B", Rating: 2400},
			{ID: "C", Rating: 2300},
			{ID: "D", Rating: 2200},
			{ID: "E", Rating: 2100},
		},
		CurrentRound: 1,
	}
	byePolicy := byePolicyLowestWithoutBye
	pairer := New(Options{ByePolicy: &byePolicy})

	round1, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(round1.Byes) != 1 || round1.Byes[0].PlayerID != "E" {
		t.Fatalf("round 1 bye = %+v, want E (the lowest-ranked player)", round1.Byes)
	}

	state.Rounds = []chesspairing.RoundData{{
		Number: 1,
		Games: []chesspairing.GameData{
			{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "C", BlackID: "D", Result: chesspairing.ResultWhiteWins},
		},
		Byes: round1.Byes,
	}}
	state.CurrentRound = 2
	round2, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("round 1 byes=%+v; round 2 byes=%+v", round1.Byes, round2.Byes)
	if len(round2.Byes) != 1 {
		t.Fatalf("round 2 byes = %+v, want exactly one bye", round2.Byes)
	}
	if round2.Byes[0].PlayerID == "E" {
		t.Fatalf("round 2 bye = %q, but E already received the round 1 bye; want a player without a prior bye", round2.Byes[0].PlayerID)
	}
}

// The default byePolicy="lowest" keeps the historical behaviour: the
// lowest-ranked player receives the bye again, even if they already had one.
func TestDefaultByePolicyStillChoosesLowest(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "A", Rating: 2500},
			{ID: "B", Rating: 2400},
			{ID: "C", Rating: 2300},
			{ID: "D", Rating: 2200},
			{ID: "E", Rating: 2100},
		},
		CurrentRound: 1,
	}
	pairer := New(Options{})

	round1, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	state.Rounds = []chesspairing.RoundData{{
		Number: 1,
		Games: []chesspairing.GameData{
			{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "C", BlackID: "D", Result: chesspairing.ResultWhiteWins},
		},
		Byes: round1.Byes,
	}}
	state.CurrentRound = 2
	round2, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(round2.Byes) != 1 || round2.Byes[0].PlayerID != "E" {
		t.Fatalf("default round 2 bye = %+v, want E (lowest ranked)", round2.Byes)
	}
}
