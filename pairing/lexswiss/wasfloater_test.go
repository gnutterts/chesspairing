// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestBuildParticipantStates_WasFloaterUsesPreviousRoundScores(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}, {ID: "p3"}, {ID: "p4"}, {ID: "p5"}, {ID: "p6"}, {ID: "p7"}},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},
					{WhiteID: "p6", BlackID: "p7", Result: chesspairing.ResultDraw},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultDraw},
					{WhiteID: "p6", BlackID: "p7", Result: chesspairing.ResultDraw},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p5", Type: chesspairing.ByePAB}},
			},
		},
		CurrentRound: 3,
	}
	participants := mustBuildParticipantStates(t, state)
	for _, want := range []struct {
		id         string
		wasFloater bool
	}{
		{id: "p1", wasFloater: true},
		{id: "p3", wasFloater: true},
		{id: "p2", wasFloater: true},
		{id: "p4", wasFloater: true},
		{id: "p5", wasFloater: false},
		{id: "p6", wasFloater: false},
		{id: "p7", wasFloater: false},
	} {
		participant := findParticipant(participants, want.id)
		if participant == nil {
			t.Fatalf("participant %s not found", want.id)
		}
		if participant.WasFloater != want.wasFloater {
			t.Errorf("%s WasFloater = %t, want %t", want.id, participant.WasFloater, want.wasFloater)
		}
	}
}
