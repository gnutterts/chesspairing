// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"sort"

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
// For unplayed rounds (byes, absences), a virtual opponent score is used:
// the player's own current score (FIDE C.02 recommendation).
type Buchholz struct {
	variant buchholzVariant
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
	data := buildOpponentData(state, scores)
	totalRounds := len(state.Rounds)

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		oppScores := opponentScores(ps.PlayerID, data, totalRounds)
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    b.applyVariant(oppScores),
		}
	}
	return result, nil
}

// opponentScores returns the list of opponent scores for a player.
// For rounds where the player had a bye or was absent (no opponent),
// a virtual opponent score equal to the player's own score is used.
func opponentScores(playerID string, data opponentData, totalRounds int) []float64 {
	games := data.playerGames[playerID]
	ownScore := data.playerScoreMap[playerID]

	scores := make([]float64, 0, totalRounds)
	for _, g := range games {
		scores = append(scores, data.playerScoreMap[g.opponentID])
	}

	// Add virtual opponent scores for unplayed rounds (byes and absences).
	virtualRounds := data.virtualOpponentRounds(playerID)
	for range virtualRounds {
		scores = append(scores, ownScore)
	}

	return scores
}

// applyVariant applies the Buchholz variant to a sorted list of opponent scores.
func (b *Buchholz) applyVariant(oppScores []float64) float64 {
	if len(oppScores) == 0 {
		return 0
	}

	sort.Float64s(oppScores)

	var start, end int
	switch b.variant {
	case buchholzCut1:
		start = 1 // drop lowest
		end = len(oppScores)
	case buchholzCut2:
		start = 2 // drop 2 lowest
		if start > len(oppScores) {
			start = len(oppScores)
		}
		end = len(oppScores)
	case buchholzMedian:
		start = 1                // drop lowest
		end = len(oppScores) - 1 // drop highest
		if end < start {
			end = start
		}
	case buchholzMedian2:
		start = 2 // drop 2 lowest
		if start > len(oppScores) {
			start = len(oppScores)
		}
		end = len(oppScores) - 2 // drop 2 highest
		if end < start {
			end = start
		}
	default: // full
		start = 0
		end = len(oppScores)
	}

	var sum float64
	for i := start; i < end; i++ {
		sum += oppScores[i]
	}
	return sum
}
