// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package standings composes a Scorer and a list of TieBreakers over a
// TournamentState into a presentation-ready standings table.
//
// The Scorer interface is unchanged. Wins / draws / losses are derived
// directly from game results in standings.Build, not from the Scorer,
// because W/D/L is orthogonal to scoring rule: Keizer and standard
// scoring both produce wins, draws, and losses from the same game
// results, they just award different point values.
//
// Two opinionated choices, both documented on Build:
//
//   - Double-forfeit games count as 0 wins, 0 draws, 0 losses for both
//     players. The game did not happen, matching what the standard scorer
//     awards (0-0) and the documented forfeit semantics in the root
//     package.
//
//   - True ties on score and all tiebreaker values share the same rank,
//     with the next distinct row's rank skipping accordingly (standard
//     "1224" competition ranking). Unique-rank stamping for ties is not
//     supported.
//
// This package depends on chesspairing/tiebreaker via BuildByID for the
// convenience of resolving tiebreaker IDs through the registry. Build
// itself takes already-resolved []TieBreaker values and has no registry
// dependency.
package standings

import (
	"context"
	"fmt"
	"sort"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/tiebreaker"
)

// Only Direct Encounter depends on the still-tied subgroup; opponent-based
// tie-breaks continue to receive the full score table.
const subgroupTieBreakerID = "direct-encounter"

// Build composes a Scorer and a list of TieBreakers over a TournamentState
// and returns a presentation-ready standings table.
//
// The output is sorted by score descending, then by each tie-breaker in
// order within the subgroups still tied. Direct Encounter is reapplied to
// any subset it leaves tied. Rows remaining tied share the same rank; the
// next distinct row's rank skips accordingly (standard "1224" ranking).
// Within a shared rank, rows retain Scorer output order.
//
// Wins / draws / losses are derived from state.Rounds[].Games:
//
//   - Decisive (1-0 or 0-1): winner +1 win, loser +1 loss.
//   - Draw: each +1 draw.
//   - Single forfeit: forfeit-winner +1 win, forfeit-loser +1 loss.
//     The game is "played" for standings purposes, reflecting what
//     happened tournament-wise. (PlayedPairs treats the same game
//     differently for pairing-history purposes — see the package
//     comment in chesspairing.go for the cross-subsystem matrix.)
//   - Double forfeit: 0 across the board for both players, including
//     no increment to GamesPlayed. The game did not happen.
//   - Bye games: not counted as wins, draws, or losses; their score
//     contribution comes from the Scorer.
//   - Pending games: skipped, contribute 0 to everything.
//
// Build returns one Standing per entry in scorer.Score's output, in the
// order Score returned them (typically active players only).
func Build(
	ctx context.Context,
	state *cp.TournamentState,
	scorer cp.Scorer,
	tieBreakers []cp.TieBreaker,
) ([]cp.Standing, error) {
	if state == nil {
		return nil, fmt.Errorf("standings: nil state")
	}
	if scorer == nil {
		return nil, fmt.Errorf("standings: nil scorer")
	}

	excluded := roundRobinExcluded(state)
	if state.PairingConfig.System == cp.PairingRoundRobin {
		state = finalRoundRobinState(state, excluded)
	}
	scoringState := state
	if state.PairingConfig.System == cp.PairingRoundRobin {
		// Include withdrawn players in the final table without erasing their
		// withdrawal dates from the tie-breakers' view of the rounds.
		copyState := *state
		copyState.Players = append([]cp.PlayerEntry(nil), state.Players...)
		for i := range copyState.Players {
			copyState.Players[i].WithdrawnAfterRound = nil
		}
		scoringState = &copyState
	}
	scores, err := scorer.Score(ctx, scoringState)
	if err != nil {
		return nil, fmt.Errorf("standings: scorer failed: %w", err)
	}

	// Partition by score, then refine only the groups still tied at each step.
	ordered := append([]cp.PlayerScore(nil), scores...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Score > ordered[j].Score })
	groups := splitScores(ordered, func(ps cp.PlayerScore) float64 { return ps.Score })
	tbValues := make([]map[string]float64, len(tieBreakers))
	for i, tb := range tieBreakers {
		values := make(map[string]float64, len(scores))
		if tb.ID() != subgroupTieBreakerID {
			// Opponent-based tie-breaks need the complete score table, even
			// when only one subgroup remains tied.
			vals, err := tb.Compute(ctx, state, scores)
			if err != nil {
				return nil, fmt.Errorf("standings: tiebreaker %q failed: %w", tb.ID(), err)
			}
			for _, v := range vals {
				values[v.PlayerID] = v.Value
			}
		}
		var next [][]cp.PlayerScore
		for _, group := range groups {
			if tb.ID() == subgroupTieBreakerID && len(group) > 1 {
				refined, err := refineDirectEncounter(ctx, state, tb, group, values)
				if err != nil {
					return nil, fmt.Errorf("standings: tiebreaker %q failed: %w", tb.ID(), err)
				}
				next = append(next, refined...)
			} else if tb.ID() == subgroupTieBreakerID {
				values[group[0].PlayerID] = 0
				next = append(next, group)
			} else {
				next = append(next, splitScores(group, func(ps cp.PlayerScore) float64 { return values[ps.PlayerID] })...)
			}
		}
		groups = next
		tbValues[i] = values
	}
	// Player lookup for DisplayName; only emitted, ranked rows occupy positions.
	playerMap := make(map[string]*cp.PlayerEntry, len(state.Players))
	for i := range state.Players {
		playerMap[state.Players[i].ID] = &state.Players[i]
	}
	positions := make(map[string]int, len(scores))
	ranks := make(map[string]int, len(scores))
	position := 0
	for _, group := range groups {
		rank := position + 1
		for _, ps := range group {
			if excluded[ps.PlayerID] || playerMap[ps.PlayerID] == nil {
				continue
			}
			positions[ps.PlayerID] = position
			ranks[ps.PlayerID] = rank
			position++
		}
	}

	// Derive W/D/L stats from game results.
	stats := deriveGameStats(state)

	out := make([]cp.Standing, 0, len(scores))
	for _, ps := range scores {
		pe := playerMap[ps.PlayerID]
		if pe == nil {
			continue
		}
		var tbs []cp.NamedValue
		for i, tb := range tieBreakers {
			val := tbValues[i][ps.PlayerID]
			tbs = append(tbs, cp.NamedValue{
				ID:    tb.ID(),
				Name:  tb.Name(),
				Value: val,
			})
		}
		s := cp.Standing{
			PlayerID:    ps.PlayerID,
			DisplayName: pe.DisplayName,
			Score:       ps.Score,
			TieBreakers: tbs,
			Team:        ps.Team,
		}
		if gs, ok := stats[ps.PlayerID]; ok {
			s.GamesPlayed = gs.played
			s.Wins = gs.wins
			s.Draws = gs.draws
			s.Losses = gs.losses
		}
		out = append(out, s)
	}

	// Players whose results do not count remain in the table, unranked at the end.
	for i := range out {
		if excluded[out[i].PlayerID] {
			out[i].Score = 0
			out[i].TieBreakers = nil
		}
	}

	// Keep subgroup order even when DE values from different groups are
	// incomparable. Excluded rows stay unranked at the end.
	sort.SliceStable(out, func(i, j int) bool {
		if excluded[out[i].PlayerID] != excluded[out[j].PlayerID] {
			return !excluded[out[i].PlayerID]
		}
		return positions[out[i].PlayerID] < positions[out[j].PlayerID]
	})
	for i := range out {
		if !excluded[out[i].PlayerID] {
			out[i].Rank = ranks[out[i].PlayerID]
		}
	}

	return out, nil
}

// roundRobinExcluded applies C.05 6.6(2) to the scheduled games in a complete
// round robin, including the second cycle when configured. Rest rounds are not games.
func roundRobinExcluded(state *cp.TournamentState) map[string]bool {
	excluded := make(map[string]bool)
	if state.PairingConfig.System != cp.PairingRoundRobin {
		return excluded
	}
	cycles := 1
	if value, ok := cp.GetInt(state.PairingConfig.Options, "cycles"); ok && value > 0 {
		cycles = value
	}
	scheduled := (len(state.Players) - 1) * cycles
	completed := make(map[string]int)
	for _, round := range state.Rounds {
		for _, game := range round.Games {
			if game.Result == cp.ResultPending || game.IsForfeit || game.Result.IsForfeit() {
				continue
			}
			completed[game.WhiteID]++
			completed[game.BlackID]++
		}
	}
	for _, player := range state.Players {
		if player.WithdrawnAfterRound != nil && completed[player.ID]*2 < scheduled {
			excluded[player.ID] = true
		}
	}
	return excluded
}

// finalRoundRobinState keeps the tournament table intact while giving scorers
// and tie-breakers a standings-only view without games against excluded players.
func finalRoundRobinState(state *cp.TournamentState, excluded map[string]bool) *cp.TournamentState {
	copyState := *state
	copyState.Players = append([]cp.PlayerEntry(nil), state.Players...)
	copyState.Rounds = append([]cp.RoundData(nil), state.Rounds...)
	for i := range copyState.Rounds {
		round := &copyState.Rounds[i]
		games := make([]cp.GameData, 0, len(round.Games))
		for _, game := range round.Games {
			if !excluded[game.WhiteID] && !excluded[game.BlackID] {
				games = append(games, game)
			}
		}
		round.Games = games
	}
	return &copyState
}

// BuildByID is a convenience wrapper that resolves tiebreaker IDs through
// the chesspairing/tiebreaker registry before calling Build. Unknown IDs
// return an error rather than being silently skipped.
func BuildByID(
	ctx context.Context,
	state *cp.TournamentState,
	scorer cp.Scorer,
	tbIDs []string,
) ([]cp.Standing, error) {
	tbs := make([]cp.TieBreaker, 0, len(tbIDs))
	for _, id := range tbIDs {
		tb, err := tiebreaker.Get(id)
		if err != nil {
			return nil, fmt.Errorf("standings: %w", err)
		}
		tbs = append(tbs, tb)
	}
	return Build(ctx, state, scorer, tbs)
}

type gameStats struct {
	played, wins, draws, losses int
}

func deriveGameStats(state *cp.TournamentState) map[string]*gameStats {
	stats := make(map[string]*gameStats, len(state.Players))
	for _, p := range state.Players {
		stats[p.ID] = &gameStats{}
	}
	for _, rd := range state.Rounds {
		for _, g := range rd.Games {
			// Pending games and double forfeits don't count for W/D/L
			// or GamesPlayed. Double forfeit: the game did not happen.
			if g.Result == cp.ResultPending || g.Result.IsDoubleForfeit() {
				continue
			}
			if s, ok := stats[g.WhiteID]; ok {
				s.played++
			}
			if s, ok := stats[g.BlackID]; ok {
				s.played++
			}
			switch g.Result {
			case cp.ResultWhiteWins, cp.ResultForfeitWhiteWins:
				if s, ok := stats[g.WhiteID]; ok {
					s.wins++
				}
				if s, ok := stats[g.BlackID]; ok {
					s.losses++
				}
			case cp.ResultBlackWins, cp.ResultForfeitBlackWins:
				if s, ok := stats[g.BlackID]; ok {
					s.wins++
				}
				if s, ok := stats[g.WhiteID]; ok {
					s.losses++
				}
			case cp.ResultDraw:
				if s, ok := stats[g.WhiteID]; ok {
					s.draws++
				}
				if s, ok := stats[g.BlackID]; ok {
					s.draws++
				}
			}
		}
	}
	return stats
}

// splitScores keeps equal-valued players together in stable descending order.
func splitScores(group []cp.PlayerScore, value func(cp.PlayerScore) float64) [][]cp.PlayerScore {
	sorted := append([]cp.PlayerScore(nil), group...)
	sort.SliceStable(sorted, func(i, j int) bool { return value(sorted[i]) > value(sorted[j]) })
	var result [][]cp.PlayerScore
	for _, ps := range sorted {
		if len(result) == 0 || value(result[len(result)-1][0]) != value(ps) {
			result = append(result, []cp.PlayerScore{})
		}
		result[len(result)-1] = append(result[len(result)-1], ps)
	}
	return result
}

func refineDirectEncounter(ctx context.Context, state *cp.TournamentState, tb cp.TieBreaker, group []cp.PlayerScore, values map[string]float64) ([][]cp.PlayerScore, error) {
	computed, err := tb.Compute(ctx, state, group)
	if err != nil {
		return nil, err
	}
	for _, v := range computed {
		values[v.PlayerID] = v.Value
	}
	if state.PairingConfig.System != cp.PairingRoundRobin {
		leader, complete := tiebreaker.SwissDirectEncounterLeader(state, group, values)
		if !complete {
			if leader == "" {
				return [][]cp.PlayerScore{group}, nil
			}
			remaining := make([]cp.PlayerScore, 0, len(group)-1)
			result := [][]cp.PlayerScore{{}}
			for _, ps := range group {
				if ps.PlayerID == leader {
					result[0] = append(result[0], ps)
				} else {
					remaining = append(remaining, ps)
				}
			}
			if len(remaining) > 1 {
				refined, err := refineDirectEncounter(ctx, state, tb, remaining, values)
				if err != nil {
					return nil, err
				}
				return append(result, refined...), nil
			}
			return append(result, remaining), nil
		}
	}
	parts := splitScores(group, func(ps cp.PlayerScore) float64 { return values[ps.PlayerID] })
	if len(parts) == 1 {
		return parts, nil
	}
	var result [][]cp.PlayerScore
	for _, part := range parts {
		if len(part) > 1 {
			refined, err := refineDirectEncounter(ctx, state, tb, part, values)
			if err != nil {
				return nil, err
			}
			result = append(result, refined...)
		} else {
			result = append(result, part)
		}
	}
	return result, nil
}
