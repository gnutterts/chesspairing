package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("wins", func() chesspairing.TieBreaker { return &Wins{} })
}

// Wins computes the number of games won over the board (FIDE Art. 7.2, WON).
//
// Only actual game wins count — byes and forfeits are excluded.
// This counts decisive results where the player won at the board.
type Wins struct{}

func (w *Wins) ID() string   { return "wins" }
func (w *Wins) Name() string { return "Games Won (OTB)" }

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
