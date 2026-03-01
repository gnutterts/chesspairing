package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("wins", func() chesspairing.TieBreaker { return &Wins{} })
}

// Wins computes the number of wins tiebreaker.
//
// The value is simply the total number of games won by each player.
// Byes and forfeits do not count as wins for this tiebreaker.
type Wins struct{}

func (w *Wins) ID() string   { return "wins" }
func (w *Wins) Name() string { return "Number of Wins" }

func (w *Wins) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	data := buildOpponentData(state, scores)

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		var wins float64
		for _, g := range data.playerGames[ps.PlayerID] {
			if g.result == resultWin {
				wins++
			}
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    wins,
		}
	}
	return result, nil
}
