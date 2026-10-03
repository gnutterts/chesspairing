// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package chesspairing

import (
	"fmt"
	"sort"
	"strings"
)

// PairingErrorKind classifies failures while producing or validating a pairing.
type PairingErrorKind int

const (
	PairingIncomplete PairingErrorKind = iota + 1
	PairingNoPABCandidate
	PairingImpossible
	PairingInvalidInput
)

// String returns the human-readable description of a pairing failure.
func (k PairingErrorKind) String() string {
	switch k {
	case PairingIncomplete:
		return "pairing is incomplete"
	case PairingNoPABCandidate:
		return "no player is eligible for the pairing-allocated bye"
	case PairingImpossible:
		return "no pairing satisfies the absolute criteria"
	case PairingInvalidInput:
		return "invalid pairing input"
	default:
		return "unknown pairing error"
	}
}

// PairingError is returned when a pairer cannot produce a complete valid pairing.
type PairingError struct {
	Kind    PairingErrorKind
	System  string
	Missing []string
	Partial *PairingResult
	Err     error
}

// Error returns a descriptive pairing error.
func (e *PairingError) Error() string {
	parts := []string{e.Kind.String()}
	if len(e.Missing) > 0 {
		missing := append([]string(nil), e.Missing...)
		sort.Strings(missing)
		parts = append(parts, "missing "+strings.Join(missing, ", "))
	}
	if e.Err != nil {
		parts = append(parts, e.Err.Error())
	}
	message := strings.Join(parts, ": ")
	if e.System == "" {
		return message
	}
	return fmt.Sprintf("%s: %s", e.System, message)
}

// Unwrap returns the underlying failure.
func (e *PairingError) Unwrap() error { return e.Err }

// ValidatePairing verifies that result accounts for every player eligible for
// the current round and obeys the pairing system's repeat-pairing rule.
func ValidatePairing(state *TournamentState, result *PairingResult) error {
	if state == nil || result == nil {
		return &PairingError{Kind: PairingInvalidInput, Err: fmt.Errorf("state and result must not be nil")}
	}
	players := make(map[string]bool, len(state.Players))
	preAssigned := make(map[string]bool, len(state.PreAssignedByes))
	for _, bye := range state.PreAssignedByes {
		preAssigned[bye.PlayerID] = true
	}
	active := make(map[string]bool, len(state.Players))
	isTeam := state.PairingConfig.System == PairingTeam
	for _, player := range state.Players {
		id := player.ID
		if isTeam && player.TeamID != "" {
			id = player.TeamID
		}
		players[id] = true
		if state.IsActiveInRound(player.ID, state.CurrentRound) && !preAssigned[id] {
			active[id] = true
		}
	}
	seen := make(map[string]bool, len(active)+len(preAssigned))
	pabs := 0
	for _, pair := range result.Pairings {
		for _, id := range []string{pair.WhiteID, pair.BlackID} {
			if !active[id] {
				return invalidPairing(fmt.Errorf("unknown or inactive player %s", id))
			}
			if seen[id] {
				return invalidPairing(fmt.Errorf("player %s appears more than once", id))
			}
			seen[id] = true
		}
		if repeatsForbidden(state) && previouslyMet(state, pair.WhiteID, pair.BlackID) {
			return invalidPairing(fmt.Errorf("players %s and %s already met", pair.WhiteID, pair.BlackID))
		}
	}
	for _, bye := range result.Byes {
		if bye.Type == ByePAB {
			pabs++
		}
		if !players[bye.PlayerID] || (!active[bye.PlayerID] && !preAssigned[bye.PlayerID]) {
			return invalidPairing(fmt.Errorf("unknown or inactive player %s", bye.PlayerID))
		}
		if seen[bye.PlayerID] {
			return invalidPairing(fmt.Errorf("player %s appears more than once", bye.PlayerID))
		}
		seen[bye.PlayerID] = true
	}
	for _, bye := range result.TeamByes {
		if bye.Type == ByePAB {
			pabs++
		}
		if !players[bye.PlayerID] || (!active[bye.PlayerID] && !preAssigned[bye.PlayerID]) {
			return invalidPairing(fmt.Errorf("unknown or inactive player %s", bye.PlayerID))
		}
		if seen[bye.PlayerID] {
			return invalidPairing(fmt.Errorf("player %s appears more than once", bye.PlayerID))
		}
		seen[bye.PlayerID] = true
	}
	if pabs > 1 {
		return invalidPairing(fmt.Errorf("more than one pairing-allocated bye"))
	}
	missing := make([]string, 0)
	for id := range active {
		if !seen[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return &PairingError{Kind: PairingIncomplete, Missing: missing}
	}
	return nil
}

func invalidPairing(err error) error { return &PairingError{Kind: PairingInvalidInput, Err: err} }

func repeatsForbidden(state *TournamentState) bool {
	switch state.PairingConfig.System {
	case "", PairingRoundRobin:
		return false
	case PairingKeizer:
		allowed, ok := state.PairingConfig.Options["allowRepeatPairings"].(bool)
		return ok && !allowed
	default:
		return true
	}
}

func previouslyMet(state *TournamentState, first, second string) bool {
	for _, round := range state.Rounds {
		if round.Number >= state.CurrentRound {
			continue
		}
		for _, game := range round.Games {
			if game.IsForfeit {
				continue
			}
			if (game.WhiteID == first && game.BlackID == second) || (game.WhiteID == second && game.BlackID == first) {
				return true
			}
		}
	}
	return false
}
