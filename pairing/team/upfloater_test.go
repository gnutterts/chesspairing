// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

func TestPairAvoidsPreviousRoundFloaterAsUpfloater(t *testing.T) {
	totalRounds := 5
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1", Rating: 2500},
			{ID: "t2", Rating: 2400},
			{ID: "t3", Rating: 2300},
			{ID: "t4", Rating: 2200},
			{ID: "t5", Rating: 2100},
			{ID: "t6", Rating: 2000},
		},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{
				{WhiteID: "t1", BlackID: "t6", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "t2", BlackID: "t5", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "t3", BlackID: "t4", Result: chesspairing.ResultWhiteWins},
			}},
			{Number: 2, Games: []chesspairing.GameData{
				{WhiteID: "t1", BlackID: "t2", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "t3", BlackID: "t6", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "t4", BlackID: "t5", Result: chesspairing.ResultDraw},
			}},
		},
		CurrentRound: 3,
	}
	result, err := New(Options{TotalRounds: &totalRounds, ColorPreferenceType: stringPtr("none")}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if hasTeamPair(result.Pairings, "t2", "t6") {
		t.Fatalf("previous-round floater t6 was selected as upfloater: %v", result.Pairings)
	}
}

func hasTeamPair(pairings []chesspairing.GamePairing, first, second string) bool {
	for _, pairing := range pairings {
		if (pairing.WhiteID == first && pairing.BlackID == second) || (pairing.WhiteID == second && pairing.BlackID == first) {
			return true
		}
	}
	return false
}

func TestByeClearsPreviousRoundFloaterStatus(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultWhiteWins}}},
			{Number: 2, TeamByes: []chesspairing.ByeEntry{{PlayerID: "a", Type: chesspairing.ByePAB}}},
		},
		CurrentRound: 3,
	}
	participants, err := buildParticipantStates(state, "match")
	if err != nil {
		t.Fatal(err)
	}
	for _, participant := range participants {
		if participant.ID == "a" && participant.WasFloater {
			t.Fatal("a was marked as a floater after a bye in the previous round")
		}
	}
}

func TestSelectTeamUpfloaterAvoidsPreviousFloater(t *testing.T) {
	floater := &lexswiss.ParticipantState{ID: "floater", TPN: 4, WasFloater: true}
	resident := &lexswiss.ParticipantState{ID: "resident", TPN: 3}
	target := &lexswiss.ParticipantState{ID: "target", TPN: 1}
	selected := selectTeamUpfloater([]*lexswiss.ParticipantState{floater, resident}, []*lexswiss.ParticipantState{target}, nil, false)
	if selected != resident {
		t.Fatalf("selected %v, want non-floater resident", selected)
	}
}

func TestSelectTeamUpfloaterIgnoresC7InLastTwoRounds(t *testing.T) {
	floater := &lexswiss.ParticipantState{ID: "floater", TPN: 4, WasFloater: true}
	resident := &lexswiss.ParticipantState{ID: "resident", TPN: 3}
	target := &lexswiss.ParticipantState{ID: "target", TPN: 1}
	selected := selectTeamUpfloater([]*lexswiss.ParticipantState{floater, resident}, []*lexswiss.ParticipantState{target}, nil, true)
	if selected != resident {
		t.Fatalf("selected %v, want the lowest TPN (Art. 3.5.4) once C7 is relaxed", selected)
	}
}
