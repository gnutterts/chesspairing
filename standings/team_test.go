// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package standings

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/team"
	"github.com/gnutterts/chesspairing/tiebreaker"
)

func TestBuildKeepsTeamPoints(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Matches: []chesspairing.MatchData{{
				HomeID: "A",
				AwayID: "B",
				Result: &chesspairing.TeamMatchResult{HomeGame: 2.5, AwayGame: 1.5},
			}},
		}},
	}
	rows, err := Build(context.Background(), state, team.New(team.Options{}), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Team == nil {
		t.Fatalf("team points missing from standings: %#v", rows)
	}
	if *rows[0].Team != (chesspairing.TeamPoints{Match: 2, Game: 2.5}) {
		t.Errorf("team points = %#v, want MP 2, GP 2.5", rows[0].Team)
	}
}

func TestBuildBoardCountOrder(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Matches: []chesspairing.MatchData{{
				HomeID: "A",
				AwayID: "B",
				Boards: []chesspairing.GameData{
					{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins}, // A wins bd1 -> count 1, B gets 0
					{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultBlackWins}, // B wins bd2 -> count 2, A gets 0
				},
			}},
		}},
	}
	// A gets MP 1, GP 1. B gets MP 1, GP 1.
	// Board count for A: 1 * 1 + 2 * 0 = 1. Inverted: -1.
	// Board count for B: 1 * 0 + 2 * 1 = 2. Inverted: -2.
	// A (-1) > B (-2), so A should be ranked higher!
	tb, _ := tiebreaker.Get("board-count")
	rows, err := Build(context.Background(), state, team.New(team.Options{}), []chesspairing.TieBreaker{tb})
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].PlayerID != "A" {
		t.Errorf("expected A to rank first, got %s", rows[0].PlayerID)
	}
	if rows[1].PlayerID != "B" {
		t.Errorf("expected B to rank second, got %s", rows[1].PlayerID)
	}
}
