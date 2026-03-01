package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("sonneborn-berger", func() chesspairing.TieBreaker { return &SonnebornBerger{} })
}

// SonnebornBerger computes the Sonneborn-Berger (SB) tiebreaker.
//
// For each game, the player gets:
//   - win:  opponent's full score
//   - draw: half of opponent's score
//   - loss: 0
//
// This rewards winning against strong opponents more than beating weak ones.
type SonnebornBerger struct{}

func (sb *SonnebornBerger) ID() string   { return "sonneborn-berger" }
func (sb *SonnebornBerger) Name() string { return "Sonneborn-Berger" }

func (sb *SonnebornBerger) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	data := buildOpponentData(state, scores)

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		var sbScore float64
		for _, g := range data.playerGames[ps.PlayerID] {
			oppScore := data.playerScoreMap[g.opponentID]
			switch g.result {
			case resultWin:
				sbScore += oppScore
			case resultDraw:
				sbScore += oppScore / 2
			case resultLoss:
				// 0
			}
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    sbScore,
		}
	}
	return result, nil
}
