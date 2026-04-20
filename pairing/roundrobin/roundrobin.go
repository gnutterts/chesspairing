// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package roundrobin implements round-robin pairing for chess tournaments.
//
// Round-robin pairing ensures every player plays every other player exactly
// once (single round-robin) or twice with reversed colors (double round-robin).
//
// The algorithm uses the FIDE Berger tables (C.05 Annex 1):
//   - Fix the last player (or bye dummy for odd counts) at position N-1
//   - Rotate remaining N-1 players through positions 0..N-2 with stride N/2-1
//   - Each rotation produces one round of pairings
//   - For N players (or N+1 if odd, with a dummy "bye" player), there are
//     N-1 rounds per cycle
//
// Color assignment follows FIDE Berger table conventions:
//   - Board 1 (fixed player vs rotating): alternates starting color per round
//   - Other boards: the player with the lower position index gets white
//   - In cycle 2 (double RR), colors are reversed if ColorBalance is true
package roundrobin

import (
	"context"
	"fmt"

	"github.com/gnutterts/chesspairing"
)

// Pairer implements the chesspairing.Pairer interface for round-robin pairing.
type Pairer struct {
	opts Options
}

// New creates a new round-robin pairer with the given options.
func New(opts Options) *Pairer {
	return &Pairer{opts: opts.WithDefaults()}
}

// NewFromMap creates a new round-robin pairer from a map[string]any config.
func NewFromMap(m map[string]any) *Pairer {
	return New(ParseOptions(m))
}

// Pair generates pairings for the next round using the Berger table method.
func (p *Pairer) Pair(_ context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	opts := p.opts

	// Round-robin schedules are deterministic Berger tables. A pre-assigned
	// bye for a specific round would conflict with the schedule, so reject
	// the call rather than silently producing an inconsistent pairing.
	if len(state.PreAssignedByes) > 0 {
		return nil, fmt.Errorf("roundrobin pairer does not support PreAssignedByes (Berger schedule is fixed); got %d entries", len(state.PreAssignedByes))
	}

	// Get active players.
	active := state.ActivePlayerIDs(state.CurrentRound)
	if len(active) < 2 {
		result := &chesspairing.PairingResult{}
		if len(active) == 1 {
			result.Byes = []chesspairing.ByeEntry{{PlayerID: active[0], Type: chesspairing.ByePAB}}
			result.Notes = []string{active[0] + " receives a bye (only player)"}
		}
		return result, nil
	}

	n := len(active)
	// If odd, add a "BYE" dummy player. The player paired against BYE
	// receives a bye that round.
	hasBye := n%2 == 1
	if hasBye {
		n++ // table size includes dummy
	}

	roundsPerCycle := n - 1
	totalRounds := roundsPerCycle * *opts.Cycles

	// CurrentRound is 1-based. Determine which table round this is.
	roundNum := state.CurrentRound
	if roundNum < 1 || roundNum > totalRounds {
		return nil, fmt.Errorf("round %d is out of range for %d-player %d-cycle round-robin (1-%d)",
			roundNum, len(active), *opts.Cycles, totalRounds)
	}

	// Determine cycle and round within cycle (both 0-based).
	cycleIdx := (roundNum - 1) / roundsPerCycle
	roundInCycle := (roundNum - 1) % roundsPerCycle

	// FIDE recommendation (C.05 Annex 1): swap the last two rounds of
	// cycle 1 in double round-robin to avoid three consecutive games
	// with the same colour at the cycle boundary.
	if *opts.SwapLastTwoRounds && *opts.Cycles == 2 && roundsPerCycle >= 2 {
		if cycleIdx == 0 && roundInCycle == roundsPerCycle-2 {
			roundInCycle = roundsPerCycle - 1
		} else if cycleIdx == 0 && roundInCycle == roundsPerCycle-1 {
			roundInCycle = roundsPerCycle - 2
		}
	}

	// Build the Berger table for this round using the FIDE algorithm
	// (C.05 Annex 1). Fix the last player (index n-1) at the last
	// position. For odd player counts, index n-1 is the bye dummy.
	// Rotate the remaining n-1 players with stride n/2-1.
	positions := make([]int, n)
	m := n - 1 // number of rotating players
	stride := n/2 - 1

	for j := 0; j < m; j++ {
		positions[j] = ((j-roundInCycle*stride)%m + m) % m
	}
	positions[m] = m // fixed: last player (or bye dummy)

	result := &chesspairing.PairingResult{}
	board := 1

	// Generate pairings from positions.
	// Pair position 0 with position n-1, position 1 with position n-2, etc.
	for i := 0; i < n/2; i++ {
		topIdx := positions[i]
		bottomIdx := positions[n-1-i]

		// Check if either is the bye dummy.
		if hasBye && (topIdx == n-1 || bottomIdx == n-1) {
			// The real player gets a bye.
			realIdx := topIdx
			if topIdx == n-1 {
				realIdx = bottomIdx
			}
			result.Byes = append(result.Byes, chesspairing.ByeEntry{PlayerID: active[realIdx], Type: chesspairing.ByePAB})
			result.Notes = append(result.Notes,
				fmt.Sprintf("%s receives a bye (round %d)", active[realIdx], roundNum))
			continue
		}

		// Assign colors per FIDE Berger convention.
		var whiteIdx, blackIdx int

		if i == 0 {
			// Board 1: position 0 (rotating) vs position n-1 (fixed).
			// Even rounds (0-based): rotating player (topIdx) gets White.
			// Odd rounds: fixed player (bottomIdx) gets White.
			if roundInCycle%2 == 0 {
				whiteIdx, blackIdx = topIdx, bottomIdx
			} else {
				whiteIdx, blackIdx = bottomIdx, topIdx
			}
		} else {
			// Other boards: top-row player (lower slot index) gets White.
			whiteIdx, blackIdx = topIdx, bottomIdx
		}

		// In even cycles (0-based: cycle 1, 3, ...), reverse colors
		// if color balance is enabled.
		if *opts.ColorBalance && cycleIdx%2 == 1 {
			whiteIdx, blackIdx = blackIdx, whiteIdx
		}

		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   board,
			WhiteID: active[whiteIdx],
			BlackID: active[blackIdx],
		})
		board++
	}

	result.Notes = append(result.Notes,
		fmt.Sprintf("Round-robin round %d (cycle %d, round %d of %d)",
			roundNum, cycleIdx+1, roundInCycle+1, roundsPerCycle))

	return result, nil
}
