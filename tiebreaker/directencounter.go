// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

func init() {
	Register("direct-encounter", func() chesspairing.TieBreaker { return &DirectEncounter{} })
}

// DirectEncounter computes the direct encounter (head-to-head) tiebreaker.
//
// For each group of tied players in scores, the value is the score against
// other members of that group. Callers can pass a subgroup after earlier
// tie-breaks have separated players. Games outside the group are ignored.
//
// If a player is not tied with anyone, their direct encounter value is 0
// (it doesn't matter since there's nothing to break).
type DirectEncounter struct{}

// SwissDirectEncounterLeader reports whether every pair has met, or the sole
// player guaranteed to lead the separate standings even if every missing game
// goes against them. An empty leader in an incomplete group leaves it tied.
func SwissDirectEncounterLeader(state *chesspairing.TournamentState, group []chesspairing.PlayerScore, values map[string]float64) (string, bool) {
	table := buildOpponentRecords(state, group)
	met := make(map[string]map[string]bool, len(group))
	for _, ps := range group {
		met[ps.PlayerID] = make(map[string]bool)
		for _, record := range table.records[ps.PlayerID] {
			if record.Played {
				met[ps.PlayerID][record.OpponentID] = true
			}
		}
	}
	opts := standard.ParseOptions(state.ScoringConfig.Options).WithDefaults()
	maxAward := *opts.PointWin
	if *opts.PointForfeitWin > maxAward {
		maxAward = *opts.PointForfeitWin
	}
	if *opts.PointBye > maxAward {
		maxAward = *opts.PointBye
	}
	complete := true
	upper := make(map[string]float64, len(group))
	for _, ps := range group {
		upper[ps.PlayerID] = values[ps.PlayerID]
		for _, other := range group {
			if ps.PlayerID != other.PlayerID && !met[ps.PlayerID][other.PlayerID] {
				complete = false
				upper[ps.PlayerID] += maxAward
			}
		}
	}
	if complete {
		return "", true
	}
	for _, ps := range group {
		guaranteed := true
		for _, other := range group {
			if ps.PlayerID != other.PlayerID && values[ps.PlayerID] <= upper[other.PlayerID] {
				guaranteed = false
				break
			}
		}
		if guaranteed {
			return ps.PlayerID, false
		}
	}
	return "", false
}

func (de *DirectEncounter) ID() string   { return "direct-encounter" }
func (de *DirectEncounter) Name() string { return "Direct Encounter" }

func (de *DirectEncounter) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildOpponentRecords(state, scores)

	// Group players by their primary score to identify tied groups.
	groups := make(map[float64]map[string]bool)
	for _, ps := range scores {
		if groups[ps.Score] == nil {
			groups[ps.Score] = make(map[string]bool)
		}
		groups[ps.Score][ps.PlayerID] = true
	}

	// For each player, compute their score from games against tied opponents.
	deScores := make(map[string]float64, len(scores))
	for _, ps := range scores {
		tiedGroup := groups[ps.Score]
		if len(tiedGroup) <= 1 {
			continue // not tied with anyone
		}

		// Repeated encounters contribute their mean, not their sum.
		byOpponent := make(map[string][]float64)
		for _, record := range table.records[ps.PlayerID] {
			if !tiedGroup[record.OpponentID] {
				continue // skip games against non-tied players
			}
			// C.07:2026 Article 15.2: a forfeit only counts as a regular game
			// with pre-determined pairings (round robin). In Swiss tournaments
			// forfeits are not head-to-head results.
			forfeitAsGame := table.roundRobin && (record.Category == ForfeitWin || record.Category == ForfeitLoss)
			if !record.Played && !forfeitAsGame {
				continue
			}
			byOpponent[record.OpponentID] = append(byOpponent[record.OpponentID], record.Points)
		}
		for _, points := range byOpponent {
			var total float64
			for _, point := range points {
				total += point
			}
			deScores[ps.PlayerID] += total / float64(len(points))
		}
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    deScores[ps.PlayerID],
		}
	}
	return result, nil
}
