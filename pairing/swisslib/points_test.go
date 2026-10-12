// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestConfiguredPairingPointsAndScoreGroups(t *testing.T) {
	state := &chesspairing.TournamentState{
		ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringFootball},
		Players: []chesspairing.PlayerEntry{
			{ID: "winner", Rating: 2100},
			{ID: "loser", Rating: 2000},
			{ID: "draw", Rating: 1900},
			{ID: "other", Rating: 1800},
			{ID: "absent", Rating: 1700},
		},
		Rounds: []chesspairing.RoundData{{Games: []chesspairing.GameData{
			{WhiteID: "winner", BlackID: "loser", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "draw", BlackID: "other", Result: chesspairing.ResultDraw},
		}}},
		CurrentRound: 2,
	}
	players := mustBuildPlayerStates(t, state)
	want := map[string]float64{"winner": 3, "loser": 0, "draw": 1, "other": 1, "absent": 0}
	for _, p := range players {
		if p.Score != want[p.ID] {
			t.Errorf("%s: score = %g, want %g", p.ID, p.Score, want[p.ID])
		}
	}
	groups := BuildScoreGroups(players)
	if len(groups) != 3 || groups[0].Score != 3 || groups[1].Score != 1 || groups[2].Score != 0 || len(groups[1].Players) != 2 {
		t.Errorf("football groups = %+v", groups)
	}
}

func TestConfiguredByesAbsencesAndForfeits(t *testing.T) {
	state := &chesspairing.TournamentState{
		ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringStandard, Options: map[string]any{
			"pointBye": 0.5, "pointAbsent": 0.25, "pointForfeitWin": 0.75,
		}},
		Players: []chesspairing.PlayerEntry{
			{ID: "pab", Rating: 2000},
			{ID: "absent", Rating: 1900},
			{ID: "forfeit", Rating: 1800},
			{ID: "opponent", Rating: 1700},
		},
		Rounds: []chesspairing.RoundData{{
			Games: []chesspairing.GameData{{WhiteID: "forfeit", BlackID: "opponent", Result: chesspairing.ResultForfeitWhiteWins}},
			Byes:  []chesspairing.ByeEntry{{PlayerID: "pab", Type: chesspairing.ByePAB}},
		}},
		CurrentRound: 2,
	}
	players := mustBuildPlayerStates(t, state)
	want := map[string]float64{"pab": 0.5, "absent": 0.25, "forfeit": 0.75, "opponent": 0}
	for _, p := range players {
		if p.Score != want[p.ID] {
			t.Errorf("%s: score = %g, want %g", p.ID, p.Score, want[p.ID])
		}
		if p.ID == "forfeit" && p.PABIneligible.Any() {
			t.Error("sub-win forfeit must not trigger C2")
		}
	}
}

func TestDecimalScoreGroups(t *testing.T) {
	players := []PlayerState{{ID: "a", PairingScore: 0.1 + 0.2}, {ID: "b", PairingScore: 0.3}}
	groups := BuildScoreGroups(players)
	if len(groups) != 1 || len(groups[0].Players) != 2 {
		t.Errorf("decimal groups = %+v", groups)
	}
}

func TestPABDefaultsToConfiguredWin(t *testing.T) {
	for _, system := range []chesspairing.ScoringSystem{chesspairing.ScoringStandard, chesspairing.ScoringFootball} {
		cfg := chesspairing.ScoringConfig{System: system, Options: map[string]any{"pointWin": 2.0}}
		if got := pointsFor(cfg).bye(chesspairing.ByePAB); got != 2 {
			t.Errorf("%s: PAB = %g, want win value 2", system, got)
		}
	}
}

func TestBakuProportionalVirtualPoints(t *testing.T) {
	cfg := chesspairing.ScoringConfig{System: chesspairing.ScoringStandard, Options: map[string]any{"pointWin": 2.0, "pointDraw": 1.0}}
	if err := BakuPremise(cfg); err != nil {
		t.Fatal(err)
	}
	win := PairingWinPoints(cfg)
	if got := BakuVirtualPoints(win, 9, 1, true); got != 2 {
		t.Errorf("full VP = %g", got)
	}
	if got := BakuVirtualPoints(win, 9, 4, true); got != 1 {
		t.Errorf("half VP = %g", got)
	}
}
