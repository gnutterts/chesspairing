// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("fore-buchholz", func() chesspairing.TieBreaker { return &ForeBuchholz{} })
}

// ForeBuchholz computes the Fore Buchholz tiebreaker (FIDE Art. 8.3, FB).
//
// This is Buchholz calculated before the final round, as if every paired
// final-round game ended in a draw. The last RoundData entry supplies those
// pairings whether its results are pending or already recorded. In a Double
// Swiss (C.04.5) every game of the match is drawn, so the per-match draw
// value replaces the match total.
//
// FIDE Category C tiebreaker.
type ForeBuchholz struct {
	legacy bool
}

func (fb *ForeBuchholz) ID() string   { return "fore-buchholz" }
func (fb *ForeBuchholz) Name() string { return "Fore Buchholz" }

func (fb *ForeBuchholz) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)
	if len(state.Rounds) != 0 {
		lastRound := state.Rounds[len(state.Rounds)-1].Number
		for playerID, records := range table.records {
			for i := range records {
				if records[i].Round != lastRound || records[i].OpponentID == "" {
					continue
				}
				// Every paired game of the last round is drawn (Article
				// 8.3), so a Double Swiss match contributes the draw value
				// of both of its games at the configured draw value.
				drawValue := table.draw * float64(records[i].Games)
				if records[i].Played {
					table.scores[playerID] += drawValue - records[i].Points
				} else if records[i].Category == None {
					table.scores[playerID] += drawValue
				} else {
					continue
				}
				records[i].Played = true
				records[i].Points = drawValue
				records[i].Category = None
				records[i].IsVUR = false
			}
			table.records[playerID] = records
		}
		for playerID, records := range table.records {
			table.adjustedScores[playerID] = table.scores[playerID]
			for _, record := range records {
				if record.Category == RequestedByeFinal {
					table.adjustedScores[playerID] += table.draw - record.Points
				}
			}
		}
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		var sum float64
		for _, contribution := range buchholzContributions(ps.PlayerID, table, !fb.legacy) {
			sum += contribution.value
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    sum,
		}
	}
	return result, nil
}
