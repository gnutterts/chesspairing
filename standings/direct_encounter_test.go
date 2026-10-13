// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package standings_test

import (
	"context"
	"testing"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
	"github.com/gnutterts/chesspairing/standings"
	"github.com/gnutterts/chesspairing/tiebreaker"
)

type fixedScores struct {
	cp.Scorer
	scores []cp.PlayerScore
}

func (f fixedScores) Score(context.Context, *cp.TournamentState) ([]cp.PlayerScore, error) {
	return f.scores, nil
}

func TestUnknownScoreDoesNotConsumeRank(t *testing.T) {
	state := &cp.TournamentState{Players: players("A", "B")}
	scores := []cp.PlayerScore{{PlayerID: "A", Score: 3}, {PlayerID: "missing", Score: 2}, {PlayerID: "B", Score: 1}}
	rows, err := standings.Build(context.Background(), state, fixedScores{scores: scores}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Rank != 1 || rows[1].Rank != 2 || rows[1].PlayerID != "B" {
		t.Errorf("ranked rows: %+v", rows)
	}
}

func TestSwissDirectEncounterGuaranteedLeader(t *testing.T) {
	state := &cp.TournamentState{
		Players: players("A", "B", "C"),
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{game("A", "B", cp.ResultWhiteWins)}},
			{Number: 2, Games: []cp.GameData{game("A", "C", cp.ResultWhiteWins)}},
		},
	}
	scores := []cp.PlayerScore{{PlayerID: "B", Score: 2}, {PlayerID: "C", Score: 2}, {PlayerID: "A", Score: 2}}
	de, _ := tiebreaker.Get("direct-encounter")
	rows, err := standings.Build(context.Background(), state, fixedScores{scores: scores}, []cp.TieBreaker{de})
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].PlayerID != "A" || rows[0].Rank != 1 || rows[1].Rank != 2 || rows[2].Rank != 2 {
		t.Errorf("guaranteed leader and unresolved remainder: %+v", rows)
	}
}

func TestDirectEncounterAfterBuchholzSubgroups(t *testing.T) {
	// Cross-group wins must not overturn the prior Buchholz order.
	state := &cp.TournamentState{
		Players: players("B", "F", "C", "E", "X1", "X2", "Y1", "Y2"),
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{game("B", "F", cp.ResultDraw), game("C", "E", cp.ResultDraw)}},
			{Number: 2, Games: []cp.GameData{game("E", "B", cp.ResultWhiteWins), game("C", "F", cp.ResultWhiteWins)}},
			{Number: 3, Games: []cp.GameData{game("B", "X1", cp.ResultDraw), game("F", "X2", cp.ResultDraw), game("C", "Y1", cp.ResultDraw), game("E", "Y2", cp.ResultDraw)}},
		},
	}
	de, _ := tiebreaker.Get("direct-encounter")
	scores := []cp.PlayerScore{{PlayerID: "B", Score: 2}, {PlayerID: "F", Score: 2}, {PlayerID: "C", Score: 2}, {PlayerID: "E", Score: 2}, {PlayerID: "X1", Score: 1}, {PlayerID: "X2", Score: 1}, {PlayerID: "Y1", Score: 0}, {PlayerID: "Y2", Score: 0}}
	bh, _ := tiebreaker.Get("buchholz")
	rows, err := standings.Build(context.Background(), state, fixedScores{standard.New(standard.Options{}.WithDefaults()), scores}, []cp.TieBreaker{bh, de})
	if err != nil {
		t.Fatal(err)
	}
	for i, id := range []string{"B", "F", "C", "E"} {
		if rows[i].PlayerID != id || rows[i].Rank != (i/2)*2+1 || rows[i].TieBreakers[1].Value != 0.5 {
			t.Errorf("row %d: %+v", i, rows[i])
		}
	}
}

func TestDirectEncounterReappliesWithinSubset(t *testing.T) {
	state := &cp.TournamentState{
		Players: players("A", "B", "C", "D"),
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{game("A", "B", cp.ResultWhiteWins), game("C", "D", cp.ResultWhiteWins)}},
			{Number: 2, Games: []cp.GameData{game("A", "C", cp.ResultDraw), game("B", "D", cp.ResultDraw)}},
			{Number: 3, Games: []cp.GameData{game("A", "D", cp.ResultDraw), game("B", "C", cp.ResultWhiteWins)}},
		},
	}
	de, _ := tiebreaker.Get("direct-encounter")
	scores := []cp.PlayerScore{{PlayerID: "A", Score: 2}, {PlayerID: "B", Score: 2}, {PlayerID: "C", Score: 2}, {PlayerID: "D", Score: 2}}
	rows, err := standings.Build(context.Background(), state, fixedScores{standard.New(standard.Options{}.WithDefaults()), scores}, []cp.TieBreaker{de})
	if err != nil {
		t.Fatal(err)
	}
	for i, id := range []string{"A", "B", "C", "D"} {
		if rows[i].PlayerID != id || rows[i].Rank != i+1 {
			t.Errorf("row %d: %+v", i, rows[i])
		}
	}
	if rows[1].TieBreakers[0].Value != 1 || rows[2].TieBreakers[0].Value != 0 {
		t.Errorf("reapplied values: B=%v C=%v", rows[1].TieBreakers, rows[2].TieBreakers)
	}
}
