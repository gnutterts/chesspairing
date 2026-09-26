// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// c07AssertFIDEValues computes one tie-break and compares every supplied
// value with the FIDE value from C.07:2026.
func c07AssertFIDEValues(t *testing.T, abbreviation, id string, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore, want map[string]float64) {
	t.Helper()

	tb, err := Get(id)
	if err != nil {
		t.Fatalf("Get(%q): %v", id, err)
	}
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("%s.Compute: %v", id, err)
	}

	got := make(map[string]float64, len(values))
	for _, value := range values {
		got[value.PlayerID] = value.Value
	}
	if len(got) != len(want) {
		t.Errorf("%s: got %d values, want %d", abbreviation, len(got), len(want))
	}
	for playerID, expected := range want {
		actual, ok := got[playerID]
		if !ok {
			t.Errorf("%s: player %s is missing, want FIDE %v", abbreviation, playerID, expected)
			continue
		}
		if actual != expected {
			t.Errorf("C.07:2026 %s: player %s = %v, want FIDE %v", abbreviation, playerID, actual, expected)
		}
	}
}

type c07Check struct {
	abbreviation string
	id           string
	want         map[string]float64
}

type c07Case struct {
	name   string
	state  *chesspairing.TournamentState
	scores []chesspairing.PlayerScore
	checks []c07Check
}

// TestC07_2026 asserts the C.07:2026 tie-break rules from Articles 15.2,
// 16.2-16.5, 7.7-7.8, 8.2-8.3 and 10.6.
func TestC07_2026(t *testing.T) {
	cases := []c07Case{
		{
			// Articles 16.2 (categories), 16.3 (adjusted opponent score)
			// and 16.4 (dummy score caps) in BH, BH-C1 and SB.
			name: "Article16_2_To_16_4",
			state: &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "p1", DisplayName: "Alice", Rating: 2000},
					{ID: "p2", DisplayName: "Bob", Rating: 1800},
					{ID: "p3", DisplayName: "Carol", Rating: 1600},
					{ID: "p4", DisplayName: "Dave", Rating: 1400},
				},
				Rounds: []chesspairing.RoundData{
					{
						Number: 1,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
							{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},
						},
					},
					{
						Number: 2,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
							{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
						},
					},
					{
						Number: 3,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
						},
						Byes: []chesspairing.ByeEntry{
							{PlayerID: "p2", Type: chesspairing.ByePAB},
							{PlayerID: "p3", Type: chesspairing.ByeZero},
						},
					},
				},
			},
			scores: []chesspairing.PlayerScore{
				{PlayerID: "p1", Score: 3.0, Rank: 1},
				{PlayerID: "p2", Score: 2.0, Rank: 2},
				{PlayerID: "p3", Score: 0.5, Rank: 3},
				{PlayerID: "p4", Score: 0.5, Rank: 3},
			},
			checks: []c07Check{
				{
					abbreviation: "BH",
					id:           "buchholz",
					want: map[string]float64{
						"p1": 3.5,
						"p2": 5.0,
						"p3": 1.5,
						"p4": 6.0,
					},
				},
				{
					abbreviation: "BH-C1",
					id:           "buchholz-cut1",
					want: map[string]float64{
						"p1": 3.0,
						"p2": 4.5,
						"p3": 1.0,
						"p4": 5.0,
					},
				},
				{
					abbreviation: "SB",
					id:           "sonneborn-berger",
					want: map[string]float64{
						"p1": 3.5,
						"p2": 2.0,
						"p3": 0.25,
						"p4": 0.5,
					},
				},
			},
		},
		{
			// Article 16.2.5 (requested bye in the last round) and the
			// Article 16.4.2 dummy cap in BH, BH-C1 and SB.
			name: "Article16_2_5",
			state: &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "p1", DisplayName: "Alice", Rating: 2000},
					{ID: "p2", DisplayName: "Bob", Rating: 1800},
					{ID: "p3", DisplayName: "Carol", Rating: 1600},
					{ID: "p4", DisplayName: "Dave", Rating: 1400},
				},
				Rounds: []chesspairing.RoundData{
					{
						Number: 1,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
							{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},
						},
					},
					{
						Number: 2,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},
							{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
						},
					},
					{
						Number: 3,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
						},
						Byes: []chesspairing.ByeEntry{
							{PlayerID: "p2", Type: chesspairing.ByePAB},
							{PlayerID: "p3", Type: chesspairing.ByeHalf},
						},
					},
				},
			},
			scores: []chesspairing.PlayerScore{
				{PlayerID: "p1", Score: 2.5, Rank: 1},
				{PlayerID: "p2", Score: 2.0, Rank: 2},
				{PlayerID: "p3", Score: 1.5, Rank: 3},
				{PlayerID: "p4", Score: 0.5, Rank: 4},
			},
			checks: []c07Check{
				{
					abbreviation: "BH",
					id:           "buchholz",
					want: map[string]float64{
						"p1": 4.0,
						"p2": 4.5,
						"p3": 4.5,
						"p4": 6.0,
					},
				},
				{
					abbreviation: "BH-C1",
					id:           "buchholz-cut1",
					want: map[string]float64{
						"p1": 3.5,
						"p2": 4.0,
						"p3": 3.0,
						"p4": 4.5,
					},
				},
				{
					abbreviation: "SB",
					id:           "sonneborn-berger",
					want: map[string]float64{
						"p1": 3.25,
						"p2": 2.0,
						"p3": 2.25,
						"p4": 0.75,
					},
				},
			},
		},
		{
			// Article 15.2: in pre-determined pairings forfeits are regular
			// games for BH and SB.
			name: "Article15_2",
			state: &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "p1", DisplayName: "Alice", Rating: 2000},
					{ID: "p2", DisplayName: "Bob", Rating: 1800},
					{ID: "p3", DisplayName: "Carol", Rating: 1600},
					{ID: "p4", DisplayName: "Dave", Rating: 1400},
				},
				Rounds: []chesspairing.RoundData{
					{
						Number: 1,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
							{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw},
						},
					},
					{
						Number: 2,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultForfeitBlackWins, IsForfeit: true},
							{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
						},
					},
					{
						Number: 3,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
							{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
						},
					},
				},
				PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingRoundRobin},
			},
			scores: []chesspairing.PlayerScore{
				{PlayerID: "p1", Score: 2.0, Rank: 1},
				{PlayerID: "p2", Score: 2.0, Rank: 1},
				{PlayerID: "p3", Score: 1.5, Rank: 3},
				{PlayerID: "p4", Score: 0.5, Rank: 4},
			},
			checks: []c07Check{
				{
					abbreviation: "BH",
					id:           "buchholz",
					want: map[string]float64{
						"p1": 4.0,
						"p2": 4.0,
						"p3": 4.5,
						"p4": 5.5,
					},
				},
				{
					abbreviation: "SB",
					id:           "sonneborn-berger",
					want: map[string]float64{
						"p1": 2.5,
						"p2": 2.0,
						"p3": 2.25,
						"p4": 0.75,
					},
				},
				{
					abbreviation: "STD",
					id:           "standard-points",
					want: map[string]float64{
						"p1": 2.0,
						"p2": 2.0,
						"p3": 1.5,
						"p4": 0.5,
					},
				},
				{
					abbreviation: "WIN",
					id:           "win",
					want: map[string]float64{
						"p1": 2.0,
						"p2": 2.0,
						"p3": 1.0,
						"p4": 0.0,
					},
				},
			},
		},
		{
			// Article 16.5.1: Cut-1 exception for participants with VURs in
			// BH-C1.
			name: "Article16_5_1",
			state: &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "p1", DisplayName: "Alice", Rating: 2200},
					{ID: "p2", DisplayName: "Bob", Rating: 2000},
					{ID: "p3", DisplayName: "Carol", Rating: 1800},
					{ID: "p4", DisplayName: "Dave", Rating: 1600},
					{ID: "p5", DisplayName: "Eve", Rating: 1400},
				},
				Rounds: []chesspairing.RoundData{
					{
						Number: 1,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
							{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
						},
						Byes: []chesspairing.ByeEntry{{PlayerID: "p5", Type: chesspairing.ByePAB}},
					},
					{
						Number: 2,
						Games: []chesspairing.GameData{
							{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
							{WhiteID: "p3", BlackID: "p5", Result: chesspairing.ResultWhiteWins},
						},
						Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeHalf}},
					},
					{
						Number: 3,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
							{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
						},
						Byes: []chesspairing.ByeEntry{{PlayerID: "p5", Type: chesspairing.ByeZero}},
					},
				},
			},
			scores: []chesspairing.PlayerScore{
				{PlayerID: "p2", Score: 2.5, Rank: 1},
				{PlayerID: "p1", Score: 2.0, Rank: 2},
				{PlayerID: "p3", Score: 2.0, Rank: 2},
				{PlayerID: "p5", Score: 1.0, Rank: 4},
				{PlayerID: "p4", Score: 0.0, Rank: 5},
			},
			checks: []c07Check{
				{
					abbreviation: "BH",
					id:           "buchholz",
					want: map[string]float64{
						"p1": 4.0,
						"p2": 4.0,
						"p3": 4.0,
						"p4": 6.5,
						"p5": 4.0,
					},
				},
				{
					abbreviation: "BH-C1",
					id:           "buchholz-cut1",
					want: map[string]float64{
						"p1": 2.5,
						"p2": 4.0,
						"p3": 4.0,
						"p4": 4.5,
						"p5": 3.0,
					},
				},
			},
		},
		{
			// STD (7.7), TPN (7.8) and RTNG (10.6). The registry stores TPN
			// negated so that higher values rank higher.
			name: "TypeB",
			state: &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "p1", DisplayName: "Alice", Rating: 2200},
					{ID: "p2", DisplayName: "Bob", Rating: 2000},
					{ID: "p3", DisplayName: "Carol", Rating: 1800},
					{ID: "p4", DisplayName: "Dave", Rating: 1600},
				},
				Rounds: []chesspairing.RoundData{
					{
						Number: 1,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
						},
						Byes: []chesspairing.ByeEntry{
							{PlayerID: "p3", Type: chesspairing.ByePAB},
							{PlayerID: "p4", Type: chesspairing.ByeHalf},
						},
					},
					{
						Number: 2,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
							{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
						},
					},
					{
						Number: 3,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultDraw},
							{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
						},
					},
				},
			},
			scores: []chesspairing.PlayerScore{
				{PlayerID: "p1", Score: 2.5, Rank: 1},
				{PlayerID: "p2", Score: 1.5, Rank: 2},
				{PlayerID: "p3", Score: 1.5, Rank: 2},
				{PlayerID: "p4", Score: 1.0, Rank: 4},
			},
			checks: []c07Check{
				{
					abbreviation: "STD",
					id:           "standard-points",
					want: map[string]float64{
						"p1": 2.5,
						"p2": 1.5,
						"p3": 1.5,
						"p4": 1.0,
					},
				},
				{
					abbreviation: "TPN",
					id:           "pairing-number",
					want: map[string]float64{
						"p1": -1.0,
						"p2": -2.0,
						"p3": -3.0,
						"p4": -4.0,
					},
				},
				{
					abbreviation: "RTNG",
					id:           "player-rating",
					want: map[string]float64{
						"p1": 2200.0,
						"p2": 2000.0,
						"p3": 1800.0,
						"p4": 1600.0,
					},
				},
			},
		},
		{
			// Article 8.2 with Article 8.3 (AOB(FB), reading A: the pending
			// final-round opponent is not counted).
			name: "AOBForeBuchholz",
			state: &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "p1", DisplayName: "Alice", Rating: 2000},
					{ID: "p2", DisplayName: "Bob", Rating: 1800},
					{ID: "p3", DisplayName: "Carol", Rating: 1600},
					{ID: "p4", DisplayName: "Dave", Rating: 1400},
				},
				Rounds: []chesspairing.RoundData{
					{
						Number: 1,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
							{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
						},
					},
					{
						Number: 2,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
							{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultDraw},
						},
					},
					{
						Number: 3,
						Games: []chesspairing.GameData{
							{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultPending},
							{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultPending},
						},
					},
				},
			},
			scores: []chesspairing.PlayerScore{
				{PlayerID: "p1", Score: 2.0, Rank: 1},
				{PlayerID: "p3", Score: 1.0, Rank: 2},
				{PlayerID: "p2", Score: 0.5, Rank: 3},
				{PlayerID: "p4", Score: 0.5, Rank: 3},
			},
			checks: []c07Check{
				{
					abbreviation: "FB",
					id:           "fore-buchholz",
					want: map[string]float64{
						"p1": 3.5,
						"p2": 5.0,
						"p3": 4.5,
						"p4": 5.0,
					},
				},
				{
					abbreviation: "AOB(FB)",
					id:           "avg-opponent-fore-buchholz",
					want: map[string]float64{
						"p1": 4.75,
						"p2": 4.25,
						"p3": 4.25,
						"p4": 4.75,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, check := range tc.checks {
				t.Run(check.abbreviation, func(t *testing.T) {
					c07AssertFIDEValues(t, check.abbreviation, check.id, tc.state, tc.scores, check.want)
				})
			}
		})
	}
}
