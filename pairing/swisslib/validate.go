package swisslib

import (
	"fmt"

	"github.com/gnutterts/chesspairing"
)

// ValidatePairing checks that a PairingResult is structurally valid:
// - Every active player is either paired exactly once or has a bye.
// - No player appears in more than one pairing.
// - No unknown player IDs in pairings or byes.
// - Board numbers are sequential starting from 1.
func ValidatePairing(players []PlayerState, result *chesspairing.PairingResult) error {
	activeIDs := make(map[string]bool, len(players))
	for _, p := range players {
		if p.Active {
			activeIDs[p.ID] = true
		}
	}

	seen := make(map[string]bool)

	// Check pairings.
	for i, pair := range result.Pairings {
		if pair.Board != i+1 {
			return fmt.Errorf("board number mismatch: expected %d, got %d", i+1, pair.Board)
		}

		for _, id := range []string{pair.WhiteID, pair.BlackID} {
			if !activeIDs[id] {
				return fmt.Errorf("unknown or inactive player in pairing: %s", id)
			}
			if seen[id] {
				return fmt.Errorf("player %s appears in multiple pairings", id)
			}
			seen[id] = true
		}
	}

	// Check byes.
	for _, id := range result.Byes {
		if !activeIDs[id] {
			return fmt.Errorf("unknown or inactive player in bye: %s", id)
		}
		if seen[id] {
			return fmt.Errorf("player %s appears in both pairing and bye", id)
		}
		seen[id] = true
	}

	// Check all active players accounted for.
	for id := range activeIDs {
		if !seen[id] {
			return fmt.Errorf("active player %s not paired or given bye", id)
		}
	}

	return nil
}
