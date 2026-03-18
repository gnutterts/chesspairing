package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("player-rating", func() chesspairing.TieBreaker { return &PlayerRating{} })
}

// PlayerRating computes the player's own rating tiebreaker
// (FIDE Art. 10.6, RTNG).
//
// The value is the player's rating. Higher rating ranks higher.
//
// FIDE Category D tiebreaker.
type PlayerRating struct{}

func (pr *PlayerRating) ID() string   { return "player-rating" }
func (pr *PlayerRating) Name() string { return "Player Rating" }

func (pr *PlayerRating) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	// Build rating lookup.
	ratings := make(map[string]int, len(state.Players))
	for _, p := range state.Players {
		ratings[p.ID] = p.Rating
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    float64(ratings[ps.PlayerID]),
		}
	}
	return result, nil
}
