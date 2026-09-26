// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// d13State builds a three-round, three-player tournament in which player p1
// has exactly two games over the board (a draw and a loss, both against
// 2200-rated opponents) plus one half-point bye. Standard scoring therefore
// gives p1 a tournament score of 1.0 even though only 0.5 points were scored
// in games played over the board.
func d13State() (*chesspairing.TournamentState, []chesspairing.PlayerScore) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alpha", Rating: 2000},
			{ID: "p2", DisplayName: "Beta", Rating: 2200},
			{ID: "p3", DisplayName: "Gamma", Rating: 2200},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games:  []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw}},
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
			{
				Number: 2,
				Games:  []chesspairing.GameData{{WhiteID: "p3", BlackID: "p1", Result: chesspairing.ResultWhiteWins}},
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p2", Type: chesspairing.ByePAB}},
			},
			{
				Number: 3,
				Games:  []chesspairing.GameData{{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins}},
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeHalf}},
			},
		},
		CurrentRound: 4,
	}

	scores, err := standard.New(standard.Options{}).Score(context.Background(), state)
	if err != nil {
		panic(err)
	}
	return state, scores
}

func d13ValueFor(values []chesspairing.TieBreakValue, id string) float64 {
	for _, v := range values {
		if v.PlayerID == id {
			return v.Value
		}
	}
	return -1
}

// TestD13_TPR_HalfPointBye
//
// FIDE C.07:2026 Article 10.2 (Tournament Performance Rating, TPR):
// "Calculated adding to ARO a number (called rating difference (RD) - which
// may be negative) resulting from the conversion of the fractional score
// (number of points achieved in games played over the board divided by the
// number of games) into RD".
//
// For the input below FIDE therefore prescribes:
//   - points achieved over the board by p1 = 0.5 (round 1 draw) + 0 (round 2 loss) = 0.5
//   - number of games = 2
//   - fractional score p = 0.5 / 2 = 0.25 -> RD = -193
//   - ARO = roundHalfUp((2200 + 2200) / 2) = 2200
//   - TPR = 2200 - 193 = 2007
//
// The half-point bye must not contribute board points: p is derived from the
// games actually played over the board, not from the tournament score.
func TestD13_TPR_HalfPointBye(t *testing.T) {
	state, scores := d13State()
	tb, err := Get("performance-rating")
	if err != nil {
		t.Fatalf("Get(performance-rating): %v", err)
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	score := 0.0
	for _, ps := range scores {
		if ps.PlayerID == "p1" {
			score = ps.Score
		}
	}

	got := d13ValueFor(values, "p1")
	t.Logf("input: p1 rating=2000, round 1 draw vs p2(2200), round 2 loss vs p3(2200), round 3 half-point bye")
	t.Logf("score(p1) via standard.Score = %.1f (0.5 OTB + 0.5 bye)", score)
	t.Logf("actual TPR(p1) = %.0f", got)
	t.Logf("FIDE 10.2: p = 0.5 OTB / 2 games = 0.25 -> RD=-193, ARO=2200 -> TPR=2007")

	if got != 2007 {
		t.Fatalf("TPR(p1) = %.0f, want 2007 (FIDE 10.2: board points / games)", got)
	}
}

// TestD13_PTP_HalfPointBye
//
// FIDE C.07:2026 Article 10.3 (Perfect Tournament Performance, PTP):
// "This is a whole number corresponding to the lowest rating that a player
// should have for their expected score to be greater than or equal to their
// tournament score."
//
// Unlike TPR (Article 10.2), the PTP definition does not restrict the score
// to games played over the board: it uses "their tournament score". A
// half-point bye is part of the tournament score, so for p1 the expected
// score over the two opponents faced must reach 1.0, not 0.5. Against two
// 2200-rated opponents that threshold is reached at rating 2200.
func TestD13_PTP_HalfPointBye(t *testing.T) {
	state, scores := d13State()
	tb, err := Get("performance-points")
	if err != nil {
		t.Fatalf("Get(performance-points): %v", err)
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	got := d13ValueFor(values, "p1")
	t.Logf("input: p1 rating=2000, round 1 draw vs p2(2200), round 2 loss vs p3(2200), round 3 half-point bye")
	t.Logf("actual PTP(p1) = %.0f", got)
	t.Logf("FIDE 10.3: lowest rating with expected score >= tournament score 1.0 against 2200,2200 -> 2197 (Table 8.1b)")

	if got != 2197 {
		t.Fatalf("PTP(p1) = %.0f, want 2197 (FIDE 10.3: tournament score)", got)
	}
}
