// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import "testing"

// TestDummyScoreCaps is the table of hand-computed Article 16.4.1/16.4.2
// dummy-score ceilings. Each case builds the smallest record that exercises
// one ceiling; the expected value is worked out by hand in the comment.
//
// Article 16.4.1 caps the dummy at the scheduled opponent's adjusted score
// for forfeits (16.2.2 and 16.2.4). Article 16.4.2 caps the dummy at half a
// draw point times the number of rounds (0.5 × rounds) for all other unplayed
// rounds (16.2.1, 16.2.3, 16.2.5).
func TestDummyScoreCaps(t *testing.T) {
	cases := []struct {
		name     string
		record   OpponentRecord
		table    opponentTable
		capped   bool
		expected float64
	}{
		{
			// Forfeit win: own 3.0 > scheduled opponent adjusted 1.0 → cap 1.0.
			name:     "forfeit-win-capped-at-opponent-adjusted-score",
			record:   OpponentRecord{Category: ForfeitWin, OpponentID: "o", Points: 1},
			table:    opponentTable{scores: map[string]float64{"p": 3.0}, adjustedScores: map[string]float64{"o": 1.0}, totalRounds: 5},
			capped:   true,
			expected: 1.0,
		},
		{
			// Forfeit loss: own 2.0 > scheduled opponent adjusted 1.5 → cap 1.5.
			name:     "forfeit-loss-capped-at-opponent-adjusted-score",
			record:   OpponentRecord{Category: ForfeitLoss, OpponentID: "o", Points: 0},
			table:    opponentTable{scores: map[string]float64{"p": 2.0}, adjustedScores: map[string]float64{"o": 1.5}, totalRounds: 5},
			capped:   true,
			expected: 1.5,
		},
		{
			// PAB: own 3.0 > 0.5 × 5 = 2.5 → cap 2.5.
			name:     "pab-capped-at-half-draw-times-rounds",
			record:   OpponentRecord{Category: PABOrFullPoint, Points: 1},
			table:    opponentTable{scores: map[string]float64{"p": 3.0}, adjustedScores: map[string]float64{}, totalRounds: 5},
			capped:   true,
			expected: 2.5,
		},
		{
			// Requested bye in the last round: own 2.5 > 0.5 × 3 = 1.5 → cap 1.5.
			name:     "requested-bye-capped-at-half-draw-times-rounds",
			record:   OpponentRecord{Category: RequestedByeFinal, Points: 0.5},
			table:    opponentTable{scores: map[string]float64{"p": 2.5}, adjustedScores: map[string]float64{}, totalRounds: 3},
			capped:   true,
			expected: 1.5,
		},
		{
			// Own score below the ceiling is returned unchanged.
			name:     "own-score-below-ceiling-unchanged",
			record:   OpponentRecord{Category: PABOrFullPoint, Points: 1},
			table:    opponentTable{scores: map[string]float64{"p": 1.0}, adjustedScores: map[string]float64{}, totalRounds: 5},
			capped:   true,
			expected: 1.0,
		},
		{
			// Legacy (uncapped) dummy always returns the player's own score.
			name:     "legacy-uncapped-returns-own-score",
			record:   OpponentRecord{Category: PABOrFullPoint, Points: 1},
			table:    opponentTable{scores: map[string]float64{"p": 3.0}, adjustedScores: map[string]float64{}, totalRounds: 1},
			capped:   false,
			expected: 3.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := dummyScore("p", tc.record, tc.table, tc.capped); got != tc.expected {
				t.Errorf("dummyScore() = %v, want %v", got, tc.expected)
			}
		})
	}
}
