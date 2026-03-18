package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("black-wins", func() chesspairing.TieBreaker { return &BlackWins{} })
}

// BlackWins computes the number of games won over the board with the
// Black pieces (FIDE Art. 7.4, BWG).
//
// Only OTB wins count — forfeit wins are excluded.
//
// FIDE Category B tiebreaker.
type BlackWins struct{}

func (bw *BlackWins) ID() string   { return "black-wins" }
func (bw *BlackWins) Name() string { return "Black Wins" }

func (bw *BlackWins) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	blackWinCount := make(map[string]float64, len(scores))

	for _, round := range state.Rounds {
		for _, game := range round.Games {
			if game.Result == chesspairing.ResultBlackWins {
				blackWinCount[game.BlackID]++
			}
			// Forfeit wins, draws, white wins: don't count.
		}
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    blackWinCount[ps.PlayerID],
		}
	}
	return result, nil
}
