package tiebreaker

import (
	"context"

	chesspairing "github.com/gnutterts/chesspairing"
)

func init() {
	Register("games-played", func() chesspairing.TieBreaker { return &GamesPlayed{} })
}

// GamesPlayed computes the number of games actually played.
//
// This is primarily useful for Keizer tournaments where players can
// miss rounds. A player who attended more evenings and played more
// games should rank higher than one with the same score but fewer
// games (suggesting they scored from absent penalties or byes).
//
// Byes and absences do NOT count as games played.
type GamesPlayed struct{}

func (gp *GamesPlayed) ID() string   { return "games-played" }
func (gp *GamesPlayed) Name() string { return "Games Played" }

func (gp *GamesPlayed) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	// Count actual games per player.
	gameCount := make(map[string]float64, len(scores))

	for _, round := range state.Rounds {
		for _, game := range round.Games {
			gameCount[game.WhiteID]++
			gameCount[game.BlackID]++
		}
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    gameCount[ps.PlayerID],
		}
	}
	return result, nil
}
