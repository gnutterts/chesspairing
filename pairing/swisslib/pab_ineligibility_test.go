// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPABIneligibilityFromResults(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "white"}, {ID: "black"}, {ID: "full"}, {ID: "double1"}, {ID: "double2"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "white", BlackID: "black", Result: chesspairing.ResultForfeitWhiteWins},
				{WhiteID: "double1", BlackID: "double2", Result: chesspairing.ResultDoubleForfeit},
			},
			Byes: []chesspairing.ByeEntry{{PlayerID: "full", Type: chesspairing.ByeFullPoint}},
		}},
		CurrentRound: 2,
	}
	players := mustBuildPlayerStates(t, state)
	byID := make(map[string]PlayerState, len(players))
	for _, player := range players {
		byID[player.ID] = player
	}
	if !byID["white"].PABIneligible.FullPointUnplayed {
		t.Error("forfeit winner recorded in Result must be ineligible")
	}
	if byID["black"].PABIneligible.Any() {
		t.Error("forfeit loser must remain eligible")
	}
	if !byID["full"].PABIneligible.FullPointUnplayed {
		t.Error("requested full-point bye must make player ineligible")
	}
	if byID["double1"].PABIneligible.Any() || byID["double2"].PABIneligible.Any() {
		t.Error("double forfeit must not make either player ineligible")
	}
}
