// Package roundrobin implements round-robin pairing for chess tournaments.
//
// Round-robin pairing ensures every player plays every other player exactly
// once (single round-robin) or twice with reversed colors (double round-robin).
//
// The algorithm uses the Berger tables / circle method:
//   - Fix player 1 in position 0
//   - Rotate remaining players through positions 1..N-1
//   - Each rotation produces one round of pairings
//   - For N players (or N+1 if odd, with a dummy "bye" player), there are
//     N-1 rounds per cycle
//
// Color assignment follows standard Berger table conventions:
//   - In round r, player at position 0 alternates white/black
//   - Other pairings: the player in the lower position gets white
//   - In cycle 2 (double RR), colors are reversed if ColorBalance is true
package roundrobin

import (
	"context"
	"fmt"

	chesspairing "github.com/gnutterts/chesspairing"
)

// Pairer implements the chesspairing.Pairer interface for round-robin pairing.
type Pairer struct {
	opts Options
}

// New creates a new round-robin pairer with the given options.
func New(opts Options) *Pairer {
	return &Pairer{opts: opts}
}

// NewFromMap creates a new round-robin pairer from a map[string]any config.
func NewFromMap(m map[string]any) *Pairer {
	return New(ParseOptions(m))
}

// Pair generates pairings for the next round using the Berger table method.
func (p *Pairer) Pair(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	opts := p.opts.WithDefaults()

	// Get active players.
	active := activePlayerIDs(state.Players)
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

	// Build the Berger table for this round.
	// Positions: fix active[0] at position 0, rotate active[1..n-1].
	// For a bye player, use the sentinel index n-1.
	positions := make([]int, n)
	positions[0] = 0 // fixed

	// Fill positions 1..n-1 with a rotated sequence.
	// For round r (0-based), position j (1-based) maps to:
	//   ((j - 1 + r) % (n-1)) + 1
	// We use player indices into the active array, with index n-1
	// representing the bye dummy (if odd).
	for j := 1; j < n; j++ {
		positions[j] = ((j - 1 + roundInCycle) % (n - 1)) + 1
	}

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

		// Assign colors per Berger convention.
		whiteIdx, blackIdx := topIdx, bottomIdx

		// Position 0 player: alternates starting color.
		// In odd-numbered rounds (0-based), the fixed player gets black.
		if i == 0 {
			if roundInCycle%2 == 1 {
				whiteIdx, blackIdx = bottomIdx, topIdx
			}
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

// activePlayerIDs returns IDs of active players in their original order.
func activePlayerIDs(players []chesspairing.PlayerEntry) []string {
	ids := make([]string, 0, len(players))
	for _, p := range players {
		if p.Active {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// Ensure Pairer implements chesspairing.Pairer.
var _ chesspairing.Pairer = (*Pairer)(nil)
