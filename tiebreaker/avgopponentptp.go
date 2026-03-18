package tiebreaker

import (
	"context"
	"math"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("avg-opponent-ptp", func() chesspairing.TieBreaker { return &AvgOpponentPTP{} })
}

// AvgOpponentPTP computes the average of opponents' Performance with
// Tournament Points (FIDE Art. 10.5, APPO).
//
// For each player, this first computes PTP for every opponent, then averages
// those values. Result is rounded to the nearest whole number.
//
// FIDE Category D tiebreaker.
type AvgOpponentPTP struct{}

func (a *AvgOpponentPTP) ID() string   { return "avg-opponent-ptp" }
func (a *AvgOpponentPTP) Name() string { return "Avg Opponent PTP" }

func (a *AvgOpponentPTP) Compute(ctx context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	data := buildOpponentData(state, scores)

	// Compute PTP for every player first.
	ptp := &PerformancePoints{}
	ptpValues, err := ptp.Compute(ctx, state, scores)
	if err != nil {
		return nil, err
	}
	ptpMap := make(map[string]float64, len(ptpValues))
	for _, v := range ptpValues {
		ptpMap[v.PlayerID] = v.Value
	}

	// Average opponents' PTP.
	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		games := data.playerGames[ps.PlayerID]
		if len(games) == 0 {
			result[i] = chesspairing.TieBreakValue{PlayerID: ps.PlayerID, Value: 0}
			continue
		}

		var totalOppPTP float64
		for _, g := range games {
			totalOppPTP += ptpMap[g.opponentID]
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    math.Round(totalOppPTP / float64(len(games))),
		}
	}
	return result, nil
}
