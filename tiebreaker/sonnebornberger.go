// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("sonneborn-berger", func() chesspairing.TieBreaker { return &SonnebornBerger{} })
	Register("sonneborn-berger-cut1", func() chesspairing.TieBreaker { return &SonnebornBerger{cut1: true} })
}

// SonnebornBerger computes the Sonneborn-Berger (SB) tiebreaker.
//
// For each game, the player gets:
//   - win:  opponent's full score
//   - draw: half of opponent's score
//   - loss: 0
//
// This rewards winning against strong opponents more than beating weak ones.
type SonnebornBerger struct {
	cut1   bool
	legacy bool
}

func (sb *SonnebornBerger) ID() string {
	if sb.cut1 {
		return "sonneborn-berger-cut1"
	}
	return "sonneborn-berger"
}

func (sb *SonnebornBerger) Name() string {
	if sb.cut1 {
		return "Sonneborn-Berger Cut-1"
	}
	return "Sonneborn-Berger"
}

func (sb *SonnebornBerger) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)
	roundRobin := state.PairingConfig.System == chesspairing.PairingRoundRobin

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		contributions := make([]tieBreakContribution, 0, table.totalRounds)
		for _, record := range table.records[ps.PlayerID] {
			var opponentScore float64
			switch {
			case record.Played:
				opponentScore = table.adjustedScores[record.OpponentID]
			case record.Category == ForfeitWin || record.Category == ForfeitLoss:
				if roundRobin {
					// Article 15.2: forfeits are regular games in
					// pre-determined pairings and therefore not VURs for
					// SB-C1. The 2023 rules still evaluate a forfeit win
					// as a draw for SB.
					opponentScore = table.adjustedScores[record.OpponentID]
					points := record.Points
					if sb.legacy && record.Category == ForfeitWin {
						points = 0.5
					}
					contributions = append(contributions, tieBreakContribution{
						value:        opponentScore * points,
						significance: opponentScore,
						result:       points,
						vur:          false,
					})
					continue
				}
				opponentScore = dummyScore(ps.PlayerID, record, table, !sb.legacy)
			case record.Category != None:
				opponentScore = dummyScore(ps.PlayerID, record, table, !sb.legacy)
			default:
				continue
			}
			contributions = append(contributions, tieBreakContribution{
				value:        opponentScore * record.Points,
				significance: opponentScore,
				result:       record.Points,
				vur:          record.IsVUR,
			})
		}
		if sb.cut1 {
			contributions = cutLeastSonnebornBerger(contributions)
		}
		var sbScore float64
		for _, contribution := range contributions {
			sbScore += contribution.value
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    sbScore,
		}
	}
	return result, nil
}
