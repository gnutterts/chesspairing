package tiebreaker

import (
	"context"

	chesspairing "github.com/gnutterts/chesspairing"
)

func init() {
	Register("black-games", func() chesspairing.TieBreaker { return &BlackGames{} })
}

// BlackGames computes the number of games played over the board with the
// Black pieces (FIDE Art. 7.3, BPG).
//
// Forfeit games are excluded — only games actually played count.
// A higher value indicates the player overcame the disadvantage of
// playing Black more frequently.
//
// FIDE Category B tiebreaker.
type BlackGames struct{}

func (bg *BlackGames) ID() string   { return "black-games" }
func (bg *BlackGames) Name() string { return "Games with Black" }

func (bg *BlackGames) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	// Count Black games played over the board (forfeits excluded) per FIDE Art. 7.3.
	blackCount := make(map[string]float64, len(scores))

	for _, round := range state.Rounds {
		for _, game := range round.Games {
			if game.Result.IsForfeit() {
				continue
			}
			blackCount[game.BlackID]++
		}
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    blackCount[ps.PlayerID],
		}
	}
	return result, nil
}
