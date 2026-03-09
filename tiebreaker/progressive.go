package tiebreaker

import (
	"context"

	chesspairing "github.com/gnutterts/chesspairing"
)

func init() {
	Register("progressive", func() chesspairing.TieBreaker { return &Progressive{} })
}

// Progressive computes the progressive score tiebreaker.
//
// The progressive score (also called cumulative score) is the sum of
// cumulative round-by-round scores. A player who scores well in early
// rounds accumulates a higher progressive score than one who scores
// the same total but in later rounds.
//
// Example: a player scoring 1, 0, 1, 1 has cumulative scores
// [1, 1, 2, 3] and progressive = 1 + 1 + 2 + 3 = 7.
type Progressive struct{}

func (p *Progressive) ID() string   { return "progressive" }
func (p *Progressive) Name() string { return "Progressive Score" }

func (p *Progressive) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	// Build round-by-round scores for each player.
	roundScores := buildRoundScores(state)

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		var progressive float64
		var cumulative float64
		for _, roundScore := range roundScores[ps.PlayerID] {
			cumulative += roundScore
			progressive += cumulative
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    progressive,
		}
	}
	return result, nil
}

// buildRoundScores returns a per-player, per-round score array.
// Each entry is the points scored in that round (1=win, 0.5=draw, 0=loss).
func buildRoundScores(state *chesspairing.TournamentState) map[string][]float64 {
	scores := make(map[string][]float64)

	// Initialize all active players.
	for _, p := range state.Players {
		if p.Active {
			scores[p.ID] = make([]float64, len(state.Rounds))
		}
	}

	for roundIdx, round := range state.Rounds {
		for _, game := range round.Games {
			switch game.Result {
			case chesspairing.ResultWhiteWins:
				if _, ok := scores[game.WhiteID]; ok {
					scores[game.WhiteID][roundIdx] = 1.0
				}
				if _, ok := scores[game.BlackID]; ok {
					scores[game.BlackID][roundIdx] = 0.0
				}
			case chesspairing.ResultBlackWins:
				if _, ok := scores[game.WhiteID]; ok {
					scores[game.WhiteID][roundIdx] = 0.0
				}
				if _, ok := scores[game.BlackID]; ok {
					scores[game.BlackID][roundIdx] = 1.0
				}
			case chesspairing.ResultDraw:
				if _, ok := scores[game.WhiteID]; ok {
					scores[game.WhiteID][roundIdx] = 0.5
				}
				if _, ok := scores[game.BlackID]; ok {
					scores[game.BlackID][roundIdx] = 0.5
				}
			}
		}

		// Byes count as wins for progressive purposes.
		for _, bye := range round.Byes {
			if _, ok := scores[bye.PlayerID]; ok {
				scores[bye.PlayerID][roundIdx] = 1.0
			}
		}

		// Absent players get 0 for the round (already initialized to 0).
	}

	return scores
}
