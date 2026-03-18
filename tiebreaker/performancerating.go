package tiebreaker

import (
	"context"
	"math"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("performance-rating", func() chesspairing.TieBreaker { return &PerformanceRating{} })
}

// PerformanceRating computes the tournament performance rating
// (FIDE Art. 10.2, TPR).
//
// TPR = ARO + dp(p), where:
//   - ARO is the average rating of opponents
//   - p = score / games (fractional score)
//   - dp(p) is the rating difference from the FIDE B.02 table
//
// For players with no games, TPR = 0.
// Result is rounded to the nearest whole number (0.5 rounds up).
//
// FIDE Category D tiebreaker.
type PerformanceRating struct{}

func (tpr *PerformanceRating) ID() string   { return "performance-rating" }
func (tpr *PerformanceRating) Name() string { return "Performance Rating" }

func (tpr *PerformanceRating) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
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

		// ARO: average rating of opponents.
		var totalRating float64
		for _, g := range games {
			totalRating += float64(ratings[g.opponentID])
		}
		aro := totalRating / float64(len(games))

		// Fractional score: score / games.
		p := ps.Score / float64(len(games))
		if p > 1.0 {
			p = 1.0
		}
		if p < 0.0 {
			p = 0.0
		}

		dp := dpFromP(p)
		tprValue := math.Round(aro + dp)

		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    tprValue,
		}
	}
	return result, nil
}
