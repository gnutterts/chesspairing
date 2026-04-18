// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package chesspairing

import "sort"

// HistoryOptions controls how PlayedPairs interprets tournament history.
//
// The zero value matches the FIDE / library-canonical defaults: forfeited
// games do not count as played for repeat-pair detection. Add this struct
// to a function call as HistoryOptions{} when the defaults are correct.
type HistoryOptions struct {
	// IncludeForfeits controls whether single-forfeit games count as
	// "played" for the purpose of repeat-pair detection.
	//
	// false (FIDE / library-canonical, default): a forfeited game is not
	// in the pairing history; the two players may be paired again. This
	// matches the documented forfeit semantics in this package — a single
	// forfeit awards the point but the game did not happen for pairing
	// purposes.
	//
	// true (house-rule territory): forfeited games count as played; the
	// two players will not be re-paired. Some local regulations require
	// this; pass true only when local rules call for it.
	//
	// Double-forfeit games are never included regardless of this option,
	// matching ResultDoubleForfeit's documented "the game never happened"
	// semantics.
	IncludeForfeits bool
}

// PlayedPairs returns the set of player pairs that have already been
// paired in the tournament, suitable for use as a forbidden-pair
// constraint when computing the next round.
//
// Each entry in the returned slice is a two-element slice of player IDs
// sorted lexicographically. The outer slice is sorted lexicographically
// by first then second element so the output is deterministic.
//
// Pending games (ResultPending) are skipped — they have not been played
// yet. Bye entries are not pairs and are skipped entirely. Forfeit
// handling is controlled by opts.IncludeForfeits; double forfeits are
// never included.
//
// Returns nil for an empty result.
//
// Forfeit handling across this package is documented at the package
// level — see the package comment in chesspairing.go.
func PlayedPairs(state *TournamentState, opts HistoryOptions) [][]string {
	if state == nil {
		return nil
	}
	seen := make(map[[2]string]bool)
	for _, round := range state.Rounds {
		for _, game := range round.Games {
			if game.WhiteID == "" || game.BlackID == "" {
				continue
			}
			switch game.Result {
			case ResultPending:
				continue
			case ResultDoubleForfeit:
				continue
			case ResultForfeitWhiteWins, ResultForfeitBlackWins:
				if !opts.IncludeForfeits {
					continue
				}
			}
			a, b := game.WhiteID, game.BlackID
			if a > b {
				a, b = b, a
			}
			seen[[2]string{a, b}] = true
		}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([][]string, 0, len(seen))
	for k := range seen {
		out = append(out, []string{k[0], k[1]})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i][0] != out[j][0] {
			return out[i][0] < out[j][0]
		}
		return out[i][1] < out[j][1]
	})
	return out
}
