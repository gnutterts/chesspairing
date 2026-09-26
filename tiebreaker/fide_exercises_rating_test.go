// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// ratingExerciseState builds a four-round tournament in which player "p"
// faces four opponents whose ratings sum to 7550, so the raw ARO is
// 7550/4 = 1887.5. results[i] is the result of p's game in round i+1,
// with p always playing White.
func ratingExerciseState(results []chesspairing.GameResult) *chesspairing.TournamentState {
	opponents := []struct {
		id     string
		rating int
	}{
		{id: "o1", rating: 1890},
		{id: "o2", rating: 1890},
		{id: "o3", rating: 1885},
		{id: "o4", rating: 1885},
	}

	players := []chesspairing.PlayerEntry{{ID: "p", DisplayName: "Player", Rating: 2000}}
	for _, o := range opponents {
		players = append(players, chesspairing.PlayerEntry{ID: o.id, DisplayName: o.id, Rating: o.rating})
	}

	state := &chesspairing.TournamentState{Players: players}
	for i, res := range results {
		state.Rounds = append(state.Rounds, chesspairing.RoundData{
			Number: i + 1,
			Games:  []chesspairing.GameData{{WhiteID: "p", BlackID: opponents[i].id, Result: res}},
		})
	}
	return state
}

// ratingExerciseValue computes the tiebreaker tbID for player "p" and
// returns its value.
func ratingExerciseValue(t *testing.T, tbID string, state *chesspairing.TournamentState, score float64) float64 {
	t.Helper()
	tb, err := Get(tbID)
	if err != nil {
		t.Fatalf("Get(%q): %v", tbID, err)
	}
	values, err := tb.Compute(context.Background(), state, []chesspairing.PlayerScore{{PlayerID: "p", Score: score, Rank: 1}})
	if err != nil {
		t.Fatalf("%s.Compute: %v", tbID, err)
	}
	if len(values) != 1 || values[0].PlayerID != "p" {
		t.Fatalf("%s.Compute: want one value for p, got %v", tbID, values)
	}
	return values[0].Value
}

// TestFIDEExercise_17_ARORoundsHalfUp asserts FIDE C.07:2026 Article 10.1:
// ARO = 7550/4 = 1887.5 rounds half up to 1888.
func TestFIDEExercise_17_ARORoundsHalfUp(t *testing.T) {
	state := ratingExerciseState([]chesspairing.GameResult{
		chesspairing.ResultDraw,
		chesspairing.ResultDraw,
		chesspairing.ResultDraw,
		chesspairing.ResultDraw,
	})

	got := ratingExerciseValue(t, "aro", state, 2.0)
	if got != 1888 {
		t.Errorf("ARO = %v, want 1888 (7550/4 = 1887.5 rounds half up)", got)
	}
}

// TestFIDEExercise_19_21_TPR asserts FIDE C.07:2026 Article 10.2 for three
// fractional scores over the four opponents from ratingExerciseState, whose
// rounded ARO is 1888. The expected values are exact B.02 table entries:
// p = 0.00 -> dp = -800 -> TPR = 1088; p = 0.50 -> dp = 0 -> TPR = 1888;
// p = 1.00 -> dp = 800 -> TPR = 2688.
func TestFIDEExercise_19_21_TPR(t *testing.T) {
	tests := []struct {
		name    string
		results []chesspairing.GameResult
		score   float64
		want    float64
	}{
		{
			name: "exercise 19 score 0 of 4",
			results: []chesspairing.GameResult{
				chesspairing.ResultBlackWins,
				chesspairing.ResultBlackWins,
				chesspairing.ResultBlackWins,
				chesspairing.ResultBlackWins,
			},
			score: 0,
			want:  1088,
		},
		{
			name: "exercise 20 score 2 of 4",
			results: []chesspairing.GameResult{
				chesspairing.ResultWhiteWins,
				chesspairing.ResultWhiteWins,
				chesspairing.ResultBlackWins,
				chesspairing.ResultBlackWins,
			},
			score: 2,
			want:  1888,
		},
		{
			name: "exercise 21 score 4 of 4",
			results: []chesspairing.GameResult{
				chesspairing.ResultWhiteWins,
				chesspairing.ResultWhiteWins,
				chesspairing.ResultWhiteWins,
				chesspairing.ResultWhiteWins,
			},
			score: 4,
			want:  2688,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := ratingExerciseState(tt.results)
			got := ratingExerciseValue(t, "performance-rating", state, tt.score)
			if got != tt.want {
				t.Errorf("TPR = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTPRWithOnlyByesAndForfeitsIsZero asserts FIDE C.07:2026 Article 15.2
// in combination with Article 10.2: forfeits and byes remain unplayed
// rounds, so a player without any game over the board has no TPR.
func TestTPRWithOnlyByesAndForfeitsIsZero(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p", DisplayName: "Player", Rating: 2000},
			{ID: "o1", DisplayName: "Opponent 1", Rating: 2200},
			{ID: "o2", DisplayName: "Opponent 2", Rating: 2200},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p", Type: chesspairing.ByeHalf}},
			},
			{
				Number: 2,
				Games:  []chesspairing.GameData{{WhiteID: "p", BlackID: "o1", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true}},
			},
			{
				Number: 3,
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p", Type: chesspairing.ByePAB}},
			},
		},
	}

	got := ratingExerciseValue(t, "performance-rating", state, 2.5)
	if got != 0 {
		t.Errorf("TPR = %v, want 0 (no games played over the board)", got)
	}
}
