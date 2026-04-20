// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package keizer implements Keizer-style pairing for chess tournaments.
//
// Keizer pairing works by ranking players by their current Keizer score
// (computed by the Keizer scorer, or by rating if no rounds have been played),
// then pairing top-down: rank 1 vs rank 2, rank 3 vs rank 4, and so on.
// The lowest-ranked player gets a bye if there's an odd number of players.
//
// Repeat avoidance: by default, players must wait at least 3 rounds before
// being paired against the same opponent again. When a conflict occurs,
// the partner is swapped with the nearest available lower-ranked player.
//
// Color assignment uses a priority cascade: absolute preferences (imbalance
// or consecutive same color) take priority, followed by strong preferences,
// color history differences, rank tiebreak, and board alternation.
package keizer

import (
	"context"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
	keizerscoring "github.com/gnutterts/chesspairing/scoring/keizer"
)

// Pairer implements the chesspairing.Pairer interface for Keizer pairing.
type Pairer struct {
	opts Options
}

// New creates a new Keizer pairer with the given options.
func New(opts Options) *Pairer {
	return &Pairer{opts: opts.WithDefaults()}
}

// NewFromMap creates a new Keizer pairer from a map[string]any config.
func NewFromMap(m map[string]any) *Pairer {
	return New(ParseOptions(m))
}

// Pair generates pairings for the next round using the Keizer method.
func (p *Pairer) Pair(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	opts := p.opts

	// Honour pre-assigned byes for the upcoming round: those players are
	// excluded from the matching pool and echoed back in result.Byes.
	preAssigned := make(map[string]bool, len(state.PreAssignedByes))
	for _, b := range state.PreAssignedByes {
		preAssigned[b.PlayerID] = true
	}
	preAssignedByes := append([]chesspairing.ByeEntry(nil), state.PreAssignedByes...)

	// Get active players.
	allActive := activePlayerIDs(state.Players)
	active := make([]string, 0, len(allActive))
	for _, id := range allActive {
		if !preAssigned[id] {
			active = append(active, id)
		}
	}
	if len(active) < 2 {
		// Not enough players to pair.
		result := &chesspairing.PairingResult{}
		if len(active) == 1 {
			result.Byes = []chesspairing.ByeEntry{{PlayerID: active[0], Type: chesspairing.ByePAB}}
			result.Notes = []string{active[0] + " receives a bye (only player)"}
		}
		if len(preAssignedByes) > 0 {
			result.Byes = append(preAssignedByes, result.Byes...)
		}
		return result, nil
	}

	// Build player entries lookup.
	entries := make(map[string]chesspairing.PlayerEntry, len(state.Players))
	for _, pl := range state.Players {
		entries[pl.ID] = pl
	}

	// Rank players: by Keizer score (if rounds exist) or by rating.
	ranked := rankPlayers(ctx, active, state, entries, opts.ScoringOptions)

	// Build pairing history for repeat avoidance.
	history := buildHistory(state.Rounds)

	// Build color histories for color allocation.
	colorHistories := buildColorHistories(state.Rounds)

	// Pair top-down.
	result := pairRanked(ranked, opts, history, colorHistories, state.CurrentRound)
	if len(preAssignedByes) > 0 {
		result.Byes = append(preAssignedByes, result.Byes...)
	}
	return result, nil
}

// activePlayerIDs returns IDs of active players.
func activePlayerIDs(players []chesspairing.PlayerEntry) []string {
	ids := make([]string, 0, len(players))
	for _, p := range players {
		if p.Active {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// rankPlayers returns player IDs sorted by Keizer score if rounds exist,
// otherwise by rating (descending). Uses the Keizer scorer internally
// because Keizer pairing rank = Keizer scoring rank.
func rankPlayers(ctx context.Context, ids []string, state *chesspairing.TournamentState, entries map[string]chesspairing.PlayerEntry, scoringOpts *keizerscoring.Options) []string {
	ranked := make([]string, len(ids))
	copy(ranked, ids)

	if len(state.Rounds) == 0 {
		// No rounds: sort by rating descending.
		sortByRating(ranked, entries)
		return ranked
	}

	// Use the Keizer scorer to compute scores for ranking.
	var opts keizerscoring.Options
	if scoringOpts != nil {
		opts = *scoringOpts
	}
	scorer := keizerscoring.New(opts)
	scores, err := scorer.Score(ctx, state)
	if err != nil {
		// Fall back to rating if scoring fails.
		sortByRating(ranked, entries)
		return ranked
	}

	// Build score lookup.
	scoreOf := make(map[string]float64, len(scores))
	for _, ps := range scores {
		scoreOf[ps.PlayerID] = ps.Score
	}

	sort.Slice(ranked, func(i, j int) bool {
		si := scoreOf[ranked[i]]
		sj := scoreOf[ranked[j]]
		if si != sj {
			return si > sj
		}
		ri := entries[ranked[i]].Rating
		rj := entries[ranked[j]].Rating
		if ri != rj {
			return ri > rj
		}
		return entries[ranked[i]].DisplayName < entries[ranked[j]].DisplayName
	})
	return ranked
}

// sortByRating sorts player IDs by rating descending, with display name
// as alphabetical tiebreak for deterministic ordering.
func sortByRating(ranked []string, entries map[string]chesspairing.PlayerEntry) {
	sort.Slice(ranked, func(i, j int) bool {
		ri := entries[ranked[i]].Rating
		rj := entries[ranked[j]].Rating
		if ri != rj {
			return ri > rj
		}
		return entries[ranked[i]].DisplayName < entries[ranked[j]].DisplayName
	})
}

// pairingHistory tracks which round each pair of players last played.
type pairingHistory map[string]map[string]int // playerA → playerB → round number

// buildHistory builds the pairing history from completed rounds.
func buildHistory(rounds []chesspairing.RoundData) pairingHistory {
	h := make(pairingHistory)
	for _, round := range rounds {
		for _, game := range round.Games {
			// Skip forfeits — per project convention, forfeits are
			// excluded from pairing history (players can re-pair).
			if game.IsForfeit {
				continue
			}
			if h[game.WhiteID] == nil {
				h[game.WhiteID] = make(map[string]int)
			}
			if h[game.BlackID] == nil {
				h[game.BlackID] = make(map[string]int)
			}
			h[game.WhiteID][game.BlackID] = round.Number
			h[game.BlackID][game.WhiteID] = round.Number
		}
	}
	return h
}

// canPair checks if two players can be paired given the repeat rules.
func canPair(a, b string, opts Options, history pairingHistory, currentRound int) bool {
	if !*opts.AllowRepeatPairings {
		// Check if they've ever played.
		if _, played := history[a][b]; played {
			return false
		}
		return true
	}

	// Allow repeats but with minimum rounds between.
	lastRound, played := history[a][b]
	if !played {
		return true
	}
	return currentRound-lastRound >= *opts.MinRoundsBetweenRepeats
}

// buildColorHistories returns the full color history for each player
// across all completed rounds. Forfeits are excluded (no color assigned).
// Byes produce ColorNone (filtered out by ComputeColorPreference).
func buildColorHistories(rounds []chesspairing.RoundData) map[string][]swisslib.Color {
	histories := make(map[string][]swisslib.Color)
	for _, round := range rounds {
		for _, game := range round.Games {
			if game.IsForfeit {
				continue
			}
			histories[game.WhiteID] = append(histories[game.WhiteID], swisslib.ColorWhite)
			histories[game.BlackID] = append(histories[game.BlackID], swisslib.ColorBlack)
		}
		for _, bye := range round.Byes {
			histories[bye.PlayerID] = append(histories[bye.PlayerID], swisslib.ColorNone)
		}
	}
	return histories
}

// pairRanked creates pairings from a ranked list of players.
// It pairs top-down: rank 1 vs rank 2, rank 3 vs rank 4, etc.
// If odd number of players, the lowest-ranked player gets a bye.
func pairRanked(ranked []string, opts Options, history pairingHistory, colorHistories map[string][]swisslib.Color, currentRound int) *chesspairing.PairingResult {
	n := len(ranked)
	result := &chesspairing.PairingResult{}

	paired := make(map[string]bool, n)

	// Build TPN lookup from ranked position (index + 1).
	tpnOf := make(map[string]int, n)
	for i, id := range ranked {
		tpnOf[id] = i + 1
	}

	// If odd, the lowest-ranked player gets a bye.
	if n%2 == 1 {
		byePlayer := ranked[n-1]
		paired[byePlayer] = true
		result.Byes = []chesspairing.ByeEntry{{PlayerID: byePlayer, Type: chesspairing.ByePAB}}
		result.Notes = append(result.Notes, byePlayer+" receives a bye (lowest ranked)")
	}

	// Pair top-down: rank 1 vs rank 2, rank 3 vs rank 4, etc.
	board := 1
	for i := 0; i < n-1; i += 2 {
		if paired[ranked[i]] {
			// This player already has a bye — skip, adjust iteration.
			i--
			continue
		}

		topPlayer := ranked[i]
		partner := ranked[i+1]

		// Check repeat avoidance.
		if !canPair(topPlayer, partner, opts, history, currentRound) {
			// Try swapping the partner with the next available player.
			swapped := false
			for alt := i + 2; alt < n; alt++ {
				if paired[ranked[alt]] {
					continue
				}
				if canPair(topPlayer, ranked[alt], opts, history, currentRound) {
					oldPartner := ranked[i+1]
					newPartner := ranked[alt]
					ranked[i+1], ranked[alt] = ranked[alt], ranked[i+1]
					partner = ranked[i+1]
					swapped = true
					result.Notes = append(result.Notes,
						"Swapped "+newPartner+" for "+oldPartner+" to avoid repeat pairing with "+topPlayer)
					break
				}
			}
			if !swapped {
				result.Notes = append(result.Notes,
					"Could not avoid repeat pairing: "+topPlayer+" vs "+partner)
			}
		}

		whiteID, blackID := allocateColor(topPlayer, partner, colorHistories, tpnOf, board)

		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   board,
			WhiteID: whiteID,
			BlackID: blackID,
		})

		paired[topPlayer] = true
		paired[partner] = true
		board++
	}

	return result
}

// allocateColor assigns white/black using the full swisslib color preference
// cascade: absolute > strong > color-history difference > rank > board alternation.
func allocateColor(a, b string, colorHistories map[string][]swisslib.Color, tpnOf map[string]int, board int) (string, string) {
	pa := &swisslib.PlayerState{
		ID:           a,
		TPN:          tpnOf[a],
		ColorHistory: colorHistories[a],
	}
	pb := &swisslib.PlayerState{
		ID:           b,
		TPN:          tpnOf[b],
		ColorHistory: colorHistories[b],
	}
	return swisslib.AllocateColor(pa, pb, false, board, nil)
}
