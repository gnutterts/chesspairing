package tiebreaker

import (
	"context"
	"math"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("performance-points", func() chesspairing.TieBreaker { return &PerformancePoints{} })
}

// PerformancePoints computes the Performance with Tournament Points
// tiebreaker (FIDE Art. 10.3, PTP).
//
// PTP is the lowest rating R such that the sum of expected scores
// (from the FIDE table) against all opponents is >= the player's actual score.
//
// Special case: if score = 0, PTP = (lowest opponent rating) - 800.
// For players with no games, PTP = 0.
//
// FIDE Category D tiebreaker.
type PerformancePoints struct{}

func (pp *PerformancePoints) ID() string   { return "performance-points" }
func (pp *PerformancePoints) Name() string { return "Performance Points" }

func (pp *PerformancePoints) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	data := buildOpponentData(state, scores)

	// Build rating lookup.
	ratings := make(map[string]int, len(state.Players))
	for _, p := range state.Players {
		ratings[p.ID] = p.Rating
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		games := data.playerGames[ps.PlayerID]
		if len(games) == 0 {
			result[i] = chesspairing.TieBreakValue{PlayerID: ps.PlayerID, Value: 0}
			continue
		}

		// Collect opponent ratings.
		oppRatings := make([]float64, len(games))
		minOppRating := math.MaxFloat64
		for j, g := range games {
			r := float64(ratings[g.opponentID])
			oppRatings[j] = r
			if r < minOppRating {
				minOppRating = r
			}
		}

		// Special case: zero score.
		if ps.Score <= 0 {
			result[i] = chesspairing.TieBreakValue{
				PlayerID: ps.PlayerID,
				Value:    math.Round(minOppRating - 800),
			}
			continue
		}

		// Special case: perfect score.
		if ps.Score >= float64(len(games)) {
			// Maximum expected score per game is when dp=800.
			// PTP = max opponent rating + 800.
			maxOppRating := 0.0
			for _, r := range oppRatings {
				if r > maxOppRating {
					maxOppRating = r
				}
			}
			result[i] = chesspairing.TieBreakValue{
				PlayerID: ps.PlayerID,
				Value:    math.Round(maxOppRating + 800),
			}
			continue
		}

		// Binary search for the lowest R where sum(expectedScore(R - oppR)) >= score.
		// Search range: min opponent rating - 800 to max opponent rating + 800.
		lo := minOppRating - 800
		hi := minOppRating + 800
		for _, r := range oppRatings {
			if r+800 > hi {
				hi = r + 800
			}
		}

		// expectedTotal computes sum of expected scores at rating R.
		expectedTotal := func(r float64) float64 {
			var total float64
			for _, oppR := range oppRatings {
				dp := r - oppR
				total += expectedScore(dp)
			}
			return total
		}

		// Binary search: find lowest R where expectedTotal(R) >= score.
		for hi-lo > 0.5 {
			mid := (lo + hi) / 2
			if expectedTotal(mid) >= ps.Score {
				hi = mid
			} else {
				lo = mid
			}
		}

		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    math.Round(hi),
		}
	}
	return result, nil
}
