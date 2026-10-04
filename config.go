// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package chesspairing

// ScoringSystem identifies which scoring algorithm to use.
type ScoringSystem string

const (
	ScoringStandard ScoringSystem = "standard"
	ScoringKeizer   ScoringSystem = "keizer"
	ScoringFootball ScoringSystem = "football"
	ScoringTeam     ScoringSystem = "team"
)

// IsValid returns true if the scoring system is a recognized value.
func (s ScoringSystem) IsValid() bool {
	switch s {
	case ScoringStandard, ScoringKeizer, ScoringFootball, ScoringTeam:
		return true
	}
	return false
}

// PairingSystem identifies which pairing algorithm to use.
type PairingSystem string

const (
	PairingDutch       PairingSystem = "dutch"
	PairingBurstein    PairingSystem = "burstein"
	PairingDubov       PairingSystem = "dubov"
	PairingLim         PairingSystem = "lim"
	PairingDoubleSwiss PairingSystem = "doubleswiss"
	PairingTeam        PairingSystem = "team"
	PairingKeizer      PairingSystem = "keizer"
	PairingRoundRobin  PairingSystem = "roundrobin"
)

// IsValid returns true if the pairing system is a recognized value.
func (p PairingSystem) IsValid() bool {
	switch p {
	case PairingDutch, PairingBurstein, PairingDubov, PairingLim, PairingDoubleSwiss, PairingTeam, PairingKeizer, PairingRoundRobin:
		return true
	}
	return false
}

// ScoringConfig holds tournament-wide scoring settings.
type ScoringConfig struct {
	System      ScoringSystem
	Tiebreakers []string
	Options     map[string]any
}

// PairingConfig holds per-period pairing settings.
type PairingConfig struct {
	System  PairingSystem
	Options map[string]any
}

// DefaultTiebreakers returns the library default tiebreaker order for the
// given pairing system. Tournament regulations select the actual order. For
// team scoring the defaults assume match points are the primary score (FIDE
// C.07 Article 13); an event with game points as primary score needs an
// explicit list.
func DefaultTiebreakers(system PairingSystem) []string {
	switch system {
	case PairingDutch, PairingBurstein, PairingDubov, PairingLim, PairingDoubleSwiss:
		return []string{"buchholz-cut1", "buchholz", "sonneborn-berger", "direct-encounter"}
	case PairingTeam:
		return []string{"buchholz-mp-cut1", "buchholz-mp", "emmsb", "mpvgp"}
	case PairingRoundRobin:
		return []string{"sonneborn-berger", "direct-encounter", "wins", "koya"}
	case PairingKeizer:
		return []string{"games-played", "direct-encounter", "wins"}
	default:
		return []string{"direct-encounter", "wins"}
	}
}
