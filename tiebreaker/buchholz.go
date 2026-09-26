// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("buchholz", func() chesspairing.TieBreaker { return &Buchholz{variant: buchholzFull} })
	Register("buchholz-cut1", func() chesspairing.TieBreaker { return &Buchholz{variant: buchholzCut1} })
	Register("buchholz-cut2", func() chesspairing.TieBreaker { return &Buchholz{variant: buchholzCut2} })
	Register("buchholz-median", func() chesspairing.TieBreaker { return &Buchholz{variant: buchholzMedian} })
	Register("buchholz-median2", func() chesspairing.TieBreaker { return &Buchholz{variant: buchholzMedian2} })
}

type buchholzVariant int

const (
	buchholzFull    buchholzVariant = iota // Sum of all opponents' scores
	buchholzCut1                           // Drop lowest opponent score
	buchholzCut2                           // Drop 2 lowest opponent scores
	buchholzMedian                         // Drop highest and lowest
	buchholzMedian2                        // Drop 2 highest and 2 lowest
)

// Buchholz computes the Buchholz tiebreaker: the sum of all opponents' scores.
// Variants drop the lowest, two lowest, or highest+lowest opponent scores.
//
// Unplayed rounds and adjusted opponent scores follow FIDE C.07 Articles
// 16.3-16.5.
type Buchholz struct {
	variant buchholzVariant
	legacy  bool
}

func (b *Buchholz) ID() string {
	switch b.variant {
	case buchholzCut1:
		return "buchholz-cut1"
	case buchholzCut2:
		return "buchholz-cut2"
	case buchholzMedian:
		return "buchholz-median"
	case buchholzMedian2:
		return "buchholz-median2"
	default:
		return "buchholz"
	}
}

func (b *Buchholz) Name() string {
	switch b.variant {
	case buchholzCut1:
		return "Buchholz Cut-1"
	case buchholzCut2:
		return "Buchholz Cut-2"
	case buchholzMedian:
		return "Buchholz Median"
	case buchholzMedian2:
		return "Buchholz Median-2"
	default:
		return "Buchholz"
	}
}

func (b *Buchholz) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		contributions := buchholzContributions(ps.PlayerID, table, !b.legacy)
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    b.applyVariant(contributions),
		}
	}
	return result, nil
}

type tieBreakContribution struct {
	value        float64
	significance float64
	result       float64
	vur          bool
}

func buchholzContributions(playerID string, table opponentTable, capped bool) []tieBreakContribution {
	contributions := make([]tieBreakContribution, 0, table.totalRounds)
	for _, record := range table.records[playerID] {
		switch {
		case record.Played:
			opponentScore := table.adjustedScores[record.OpponentID]
			contributions = append(contributions, tieBreakContribution{value: opponentScore, significance: opponentScore})
		case record.Category == ForfeitWin || record.Category == ForfeitLoss:
			if table.roundRobin {
				// Article 15.2: in pre-determined pairings forfeits are
				// regular games, so the scheduled opponent's score counts.
				opponentScore := table.adjustedScores[record.OpponentID]
				contributions = append(contributions, tieBreakContribution{value: opponentScore, significance: opponentScore})
				continue
			}
			opponentScore := dummyScore(playerID, record, table, capped)
			contributions = append(contributions, tieBreakContribution{
				value: opponentScore, significance: opponentScore, vur: record.IsVUR,
			})
		case record.Category != None:
			opponentScore := dummyScore(playerID, record, table, capped)
			contributions = append(contributions, tieBreakContribution{
				value: opponentScore, significance: opponentScore, vur: record.IsVUR,
			})
		}
	}
	return contributions
}

func dummyScore(playerID string, record OpponentRecord, table opponentTable, capped bool) float64 {
	dummy := table.scores[playerID]
	if !capped {
		return dummy
	}
	var ceiling float64
	if record.Category == ForfeitWin || record.Category == ForfeitLoss {
		ceiling = table.adjustedScores[record.OpponentID]
	} else {
		ceiling = 0.5 * float64(table.totalRounds)
	}
	if dummy > ceiling {
		return ceiling
	}
	return dummy
}

func (b *Buchholz) applyVariant(contributions []tieBreakContribution) float64 {
	values := append([]tieBreakContribution(nil), contributions...)
	switch b.variant {
	case buchholzCut1:
		values = cutLeast(values)
	case buchholzCut2:
		values = cutLeast(cutLeast(values))
	case buchholzMedian:
		values = cutGreatest(cutLeast(values))
	case buchholzMedian2:
		values = cutLeast(cutLeast(values))
		values = cutGreatest(cutGreatest(values))
	}

	var sum float64
	for _, contribution := range values {
		sum += contribution.value
	}
	return sum
}

// cutLeast applies the Article 16.5 exception: while a VUR contribution is
// present, the lowest such contribution is cut instead of a lower non-VUR.
func cutLeast(values []tieBreakContribution) []tieBreakContribution {
	if len(values) == 0 {
		return values
	}
	index := -1
	for i, value := range values {
		if value.vur && (index < 0 || value.value < values[index].value) {
			index = i
		}
	}
	if index < 0 {
		index = 0
		for i := 1; i < len(values); i++ {
			if values[i].value < values[index].value {
				index = i
			}
		}
	}
	return append(values[:index], values[index+1:]...)
}

func cutLeastSonnebornBerger(values []tieBreakContribution) []tieBreakContribution {
	if len(values) == 0 {
		return values
	}
	least := 0
	lowestVUR := -1
	for i, value := range values {
		if value.significance < values[least].significance ||
			(value.significance == values[least].significance && value.result < values[least].result) {
			least = i
		}
		if value.vur && (lowestVUR < 0 || value.value < values[lowestVUR].value) {
			lowestVUR = i
		}
	}
	if lowestVUR >= 0 && values[lowestVUR].value >= values[least].value {
		least = lowestVUR
	}
	return append(values[:least], values[least+1:]...)
}

func cutGreatest(values []tieBreakContribution) []tieBreakContribution {
	if len(values) == 0 {
		return values
	}
	index := 0
	for i := 1; i < len(values); i++ {
		if values[i].value > values[index].value {
			index = i
		}
	}
	return append(values[:index], values[index+1:]...)
}
