// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// c07AssertFIDEValues computes one tie-break and compares every supplied
// value with the FIDE value.
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
			t.Errorf("C.07:2026 Article 16.2.5 %s: player %s = %v, want FIDE %v", abbreviation, playerID, actual, expected)
		}
	}
}

// TestC07_2026_02_Article16_2_5 asserts FIDE C.07:2026 Article 16.2.5 for a
// requested last-round bye. The participant's unplayed round is evaluated as
// a draw for opponents' tie-breaks (Article 16.3.2), and its dummy score is
// capped at one half of the number of rounds (Article 16.4.2).
//
// R1: p1 1-0 p2, p3 1/2-1/2 p4
// R2: p1 1/2-1/2 p3, p4 1-0 p2
// R3: p1 1-0 p4; p2 PAB; p3 requested half-bye (Article 16.2.5).
func TestC07_2026_02_Article16_2_5(t *testing.T) {
	state := &chesspairing.TournamentState{
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
					{WhiteID: "p4", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
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
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "p1", Score: 2.5, Rank: 1},
		{PlayerID: "p3", Score: 1.5, Rank: 2},
		{PlayerID: "p4", Score: 1.5, Rank: 2},
		{PlayerID: "p2", Score: 1.0, Rank: 4},
	}

	c07AssertFIDEValues(t, "BH", "buchholz", state, scores, map[string]float64{
		"p1": 4.0,
		"p2": 5.0,
		"p3": 5.5,
		"p4": 5.0,
	})
	c07AssertFIDEValues(t, "BH-C1", "buchholz-cut1", state, scores, map[string]float64{
		"p1": 3.0,
		"p2": 4.0,
		"p3": 4.0,
		"p4": 4.0,
	})
	c07AssertFIDEValues(t, "SB", "sonneborn-berger", state, scores, map[string]float64{
		"p1": 3.25,
		"p2": 0.0,
		"p3": 2.0,
		"p4": 1.75,
	})
}
