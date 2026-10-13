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

func TestRoundRobinExcludedByeDoesNotConsumeRank(t *testing.T) {
	withdrawn := 1
	state := &cp.TournamentState{
		PairingConfig: cp.PairingConfig{System: cp.PairingRoundRobin},
		Players:       players("w", "a", "b", "c", "d"),
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{game("w", "a", cp.ResultWhiteWins), game("a", "b", cp.ResultWhiteWins), game("a", "c", cp.ResultWhiteWins), game("a", "d", cp.ResultWhiteWins), game("b", "c", cp.ResultWhiteWins), game("b", "d", cp.ResultWhiteWins), game("c", "d", cp.ResultWhiteWins)}},
			{Number: 2, Byes: []cp.ByeEntry{{PlayerID: "w", Type: cp.ByeFullPoint}}},
		},
	}
	state.Players[0].WithdrawnAfterRound = &withdrawn
	rows, err := standings.Build(context.Background(), state, standard.New(standard.Options{}.WithDefaults()), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 5 || rows[4].PlayerID != "w" || rows[4].Rank != 0 {
		t.Fatalf("excluded player: %+v", rows)
	}
	for i := 0; i < 4; i++ {
		if rows[i].Rank != i+1 {
			t.Errorf("row %d: %+v", i, rows[i])
		}
	}
}

func TestRoundRobinWithdrawalThreshold(t *testing.T) {
	for _, tc := range []struct {
		name      string
		completed int
		excluded  bool
	}{
		{"below half", 1, true}, {"exactly half", 2, false}, {"above half", 3, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withdrawn := tc.completed
			state := &cp.TournamentState{
				PairingConfig: cp.PairingConfig{System: cp.PairingRoundRobin},
				Players:       players("w", "a", "b", "c", "d"),
				CurrentRound:  5,
			}
			state.Players[0].WithdrawnAfterRound = &withdrawn
			for i, id := range []string{"a", "b", "c", "d"} {
				result := cp.ResultForfeitBlackWins
				if i < tc.completed {
					result = cp.ResultWhiteWins
				}
				state.Rounds = append(state.Rounds, cp.RoundData{Number: i + 1, Games: []cp.GameData{game("w", id, result)}})
			}
			ctx := context.Background()
			scorer := standard.New(standard.Options{}.WithDefaults())
			rows, err := standings.BuildByID(ctx, state, scorer, []string{"buchholz"})
			if err != nil {
				t.Fatal(err)
			}
			w, a := findStanding(rows, "w"), findStanding(rows, "a")
			if w == nil || a == nil {
				t.Fatalf("missing rows: %v", rows)
			}
			if tc.excluded {
				if w.Rank != 0 || rows[len(rows)-1].PlayerID != "w" || w.Score != 0 || a.Losses != 0 || a.GamesPlayed != 0 || len(w.TieBreakers) != 0 {
					t.Errorf("excluded withdrawal: w=%+v a=%+v", *w, *a)
				}
			} else {
				if w.Rank == 0 || w.Score != float64(tc.completed) || a.Losses != 1 || a.GamesPlayed != 1 || len(w.TieBreakers) != 1 {
					t.Fatalf("counted withdrawal: w=%+v a=%+v", *w, *a)
				}
				scoringState := *state
				scoringState.Players = append([]cp.PlayerEntry(nil), state.Players...)
				scoringState.Players[0].WithdrawnAfterRound = nil
				scores, err := scorer.Score(ctx, &scoringState)
				if err != nil {
					t.Fatal(err)
				}
				buchholz, err := tiebreaker.Get("buchholz")
				if err != nil {
					t.Fatal(err)
				}
				values, err := buchholz.Compute(ctx, state, scores)
				if err != nil {
					t.Fatal(err)
				}
				for _, value := range values {
					if value.PlayerID == "w" {
						if w.TieBreakers[0].Value != value.Value {
							t.Errorf("withdrawn player's Buchholz = %v, want %v from unmodified state", w.TieBreakers[0].Value, value.Value)
						}
						return
					}
				}
				t.Fatal("Buchholz missing withdrawn player")
			}
		})
	}
}
