// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func TestStageTrailingBlankRound(t *testing.T) {
	state := &cp.TournamentState{CurrentRound: 3}
	for i := 0; i < 10; i++ {
		state.Players = append(state.Players, cp.PlayerEntry{ID: string(rune('a' + i))})
	}
	first := cp.RoundData{Number: 1}
	for i := 0; i < 10; i += 2 {
		first.Games = append(first.Games, cp.GameData{WhiteID: state.Players[i].ID, BlackID: state.Players[i+1].ID, Result: cp.ResultDraw})
	}
	state.Rounds = []cp.RoundData{first, {Number: 2}}
	stageUnpairedRound(state)
	if state.CurrentRound != 2 || len(state.Rounds) != 1 || len(state.PreAssignedByes) != 0 {
		t.Errorf("trailing blank round not staged: %+v", state)
	}
}
