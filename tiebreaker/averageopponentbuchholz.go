package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("avg-opponent-buchholz", func() chesspairing.TieBreaker { return &AvgOpponentBuchholz{} })
}

// AvgOpponentBuchholz computes the Average of Opponents' Buchholz
// tiebreaker (FIDE Art. 8.2, AOB).
//
// For each player, this first computes the full Buchholz of every opponent,
// then averages those values. Unplayed rounds use virtual opponent scores
// as in standard Buchholz.
//
// FIDE Category C tiebreaker.
type AvgOpponentBuchholz struct{}

func (a *AvgOpponentBuchholz) ID() string   { return "avg-opponent-buchholz" }
func (a *AvgOpponentBuchholz) Name() string { return "Avg Opponent Buchholz" }

func (a *AvgOpponentBuchholz) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	data := buildOpponentData(state, scores)
	totalRounds := len(state.Rounds)

	// Compute full Buchholz for every scored player.
	bhScores := make(map[string]float64, len(scores))
	for _, ps := range scores {
		oppScores := opponentScores(ps.PlayerID, data, totalRounds)
		var sum float64
		for _, s := range oppScores {
			sum += s
		}
		bhScores[ps.PlayerID] = sum
	}

	// For each player, average the Buchholz of their opponents.
	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		games := data.playerGames[ps.PlayerID]
		if len(games) == 0 {
			result[i] = chesspairing.TieBreakValue{PlayerID: ps.PlayerID, Value: 0}
			continue
		}

		var totalOppBH float64
		for _, g := range games {
			totalOppBH += bhScores[g.opponentID]
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    totalOppBH / float64(len(games)),
		}
	}
	return result, nil
}
