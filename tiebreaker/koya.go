package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("koya", func() chesspairing.TieBreaker { return &Koya{} })
}

// Koya computes the Koya system tiebreaker.
//
// The Koya system counts the number of points scored against opponents
// who have 50% or more of the maximum possible score. In a round-robin,
// this is a useful tiebreaker because it rewards consistency against
// the stronger half of the field.
//
// Implementation:
//   - Determine the "qualifying" threshold: half the number of rounds
//   - Find all opponents whose score >= threshold
//   - Sum the player's results against those opponents (1=win, 0.5=draw, 0=loss)
type Koya struct{}

func (k *Koya) ID() string   { return "koya" }
func (k *Koya) Name() string { return "Koya System" }

func (k *Koya) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	data := buildOpponentData(state, scores)

	// Qualifying threshold: 50% of the number of rounds.
	totalRounds := len(state.Rounds)
	threshold := float64(totalRounds) / 2.0

	// Build set of qualifying opponents (score >= threshold).
	qualifying := make(map[string]bool, len(scores))
	for _, ps := range scores {
		if ps.Score >= threshold {
			qualifying[ps.PlayerID] = true
		}
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		var koyaScore float64
		for _, g := range data.playerGames[ps.PlayerID] {
			if !qualifying[g.opponentID] {
				continue
			}
			switch g.result {
			case resultWin:
				koyaScore += 1.0
			case resultDraw:
				koyaScore += 0.5
			case resultLoss:
				// 0
			}
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    koyaScore,
		}
	}
	return result, nil
}
