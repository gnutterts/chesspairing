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

// Build composes a Scorer and a list of TieBreakers over a TournamentState
// and returns a presentation-ready standings table.
//
// The output is sorted by score descending, then by tiebreaker values in
// the order tieBreakers was supplied, descending. Rows that tie on score
// and all tiebreaker values share the same rank; the next distinct row's
// rank skips accordingly (standard "1224" ranking). Within a shared rank,
// rows are stable-sorted by Scorer output order.
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

	scores, err := scorer.Score(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("standings: scorer failed: %w", err)
	}

	// Compute tiebreakers up front. tbValues[tbIdx][playerID] = value.
	tbValues := make([]map[string]float64, len(tieBreakers))
	for i, tb := range tieBreakers {
		vals, err := tb.Compute(ctx, state, scores)
		if err != nil {
			return nil, fmt.Errorf("standings: tiebreaker %q failed: %w", tb.ID(), err)
		}
		m := make(map[string]float64, len(vals))
		for _, v := range vals {
			m[v.PlayerID] = v.Value
		}
		tbValues[i] = m
	}

	// Player lookup for DisplayName.
	playerMap := make(map[string]*cp.PlayerEntry, len(state.Players))
	for i := range state.Players {
		playerMap[state.Players[i].ID] = &state.Players[i]
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
		}
		if gs, ok := stats[ps.PlayerID]; ok {
			s.GamesPlayed = gs.played
			s.Wins = gs.wins
			s.Draws = gs.draws
			s.Losses = gs.losses
		}
		out = append(out, s)
	}

	// Sort by score desc, then tiebreakers in order desc. Stable so that
	// rows tied on everything keep the Scorer's output order.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		for k := range out[i].TieBreakers {
			if k >= len(out[j].TieBreakers) {
				break
			}
			if out[i].TieBreakers[k].Value != out[j].TieBreakers[k].Value {
				return out[i].TieBreakers[k].Value > out[j].TieBreakers[k].Value
			}
		}
		return false
	})

	// Assign shared ranks ("1224" style).
	if len(out) > 0 {
		out[0].Rank = 1
		for i := 1; i < len(out); i++ {
			if out[i].Score == out[i-1].Score && tieBreakersEqual(out[i].TieBreakers, out[i-1].TieBreakers) {
				out[i].Rank = out[i-1].Rank
			} else {
				out[i].Rank = i + 1
			}
		}
	}

	return out, nil
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

func tieBreakersEqual(a, b []cp.NamedValue) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}
