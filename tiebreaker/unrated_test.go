// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// unratedState is a one-round tournament in which rated Alice drew the
// unrated Uma over the board.
func unratedState() *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "A", DisplayName: "Alice", Rating: 2000},
			{ID: "U", DisplayName: "Uma", Rating: 0},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "A", BlackID: "U", Result: chesspairing.ResultDraw},
				},
			},
		},
	}
}

// TestUnratedOpponentErrorsRatingTiebreaks asserts C.07:2026 Article 10:
// ratings-based tie-breaks are dropped when unrated players are present,
// unless the tournament regulations define detailed rules. Computing ARO,
// TPR, PTP, APRO or APPO with an unrated player (rating 0) in a game played
// over the board returns an error naming that player. Before, the missing
// rating was averaged as 0.
func TestUnratedOpponentErrorsRatingTiebreaks(t *testing.T) {
	state := unratedState()

	for _, id := range []string{"aro", "performance-rating", "performance-points", "avg-opponent-tpr", "avg-opponent-ptp"} {
		tb, err := Get(id)
		if err != nil {
			t.Fatalf("Get(%q): %v", id, err)
		}
		_, err = tb.Compute(context.Background(), state, []chesspairing.PlayerScore{
			{PlayerID: "A", Score: 0.5, Rank: 1},
			{PlayerID: "U", Score: 0.5, Rank: 1},
		})
		if err == nil {
			t.Errorf("%s: want an error for the unrated player, got none", id)
			continue
		}
		if !strings.Contains(err.Error(), "Uma") {
			t.Errorf("%s: error %q does not name the unrated player Uma", id, err)
		}
	}
}

// TestUnratedOverrideUsed asserts the detailed rules of Article 10: with
// the scoring option "unratedRating" (1900), Uma is rated 1900 and the
// ratings-based tie-breaks compute normally: ARO = 1900 (0.5 rounds up),
// TPR = round(1900 + dp(0.5)) = 1900, APRO = 2000 (Uma's TPR against Alice
// at 0.5 score). PTP and APPO look for the lowest rating whose expected
// score still reaches the score: the B.02 table keeps expected score 0.5
// down to a difference of -3, so PTP = 1900 - 3 = 1897 and APPO =
// 2000 - 3 = 1997.
func TestUnratedOverrideUsed(t *testing.T) {
	state := unratedState()
	state.ScoringConfig = chesspairing.ScoringConfig{
		Options: map[string]any{"unratedRating": 1900},
	}
	assertUnratedValue(t, "aro", state, "A", 1900)
	assertUnratedValue(t, "performance-rating", state, "A", 1900)
	assertUnratedValue(t, "performance-points", state, "A", 1897)
	assertUnratedValue(t, "avg-opponent-tpr", state, "A", 2000)
	assertUnratedValue(t, "avg-opponent-ptp", state, "A", 1997)
}

// assertUnratedValue computes one tie-break and asserts Alice's value.
func assertUnratedValue(t *testing.T, id string, state *chesspairing.TournamentState, playerID string, want float64) {
	t.Helper()
	tb, err := Get(id)
	if err != nil {
		t.Fatalf("Get(%q): %v", id, err)
	}
	values, err := tb.Compute(context.Background(), state, []chesspairing.PlayerScore{
		{PlayerID: "A", Score: 0.5, Rank: 1},
		{PlayerID: "U", Score: 0.5, Rank: 1},
	})
	if err != nil {
		t.Fatalf("%s.Compute: %v", id, err)
	}
	for _, value := range values {
		if value.PlayerID != playerID {
			continue
		}
		if value.Value != want {
			t.Errorf("%s: player %s = %v, want %v", id, playerID, value.Value, want)
		}
		return
	}
	t.Errorf("%s: no value for player %s", id, playerID)
}
