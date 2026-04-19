// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package factory constructs chesspairing engines (Pairers, Scorers,
// TieBreakers) by name.
//
// Importing this package transitively imports every registered engine.
// Consumers who need a minimal binary should keep importing the engine
// packages they want directly; this package exists for callers that
// dispatch on configuration strings (CLI tools, JSON-driven services,
// test harnesses) and don't mind the larger dependency surface.
//
// Names are lowercase and match the canonical PairingSystem and
// ScoringSystem string values defined in the root package. Tiebreaker
// IDs come from the self-registering tiebreaker registry.
package factory

import (
	"fmt"
	"sort"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/burstein"
	"github.com/gnutterts/chesspairing/pairing/doubleswiss"
	"github.com/gnutterts/chesspairing/pairing/dubov"
	"github.com/gnutterts/chesspairing/pairing/dutch"
	"github.com/gnutterts/chesspairing/pairing/keizer"
	"github.com/gnutterts/chesspairing/pairing/lim"
	"github.com/gnutterts/chesspairing/pairing/roundrobin"
	"github.com/gnutterts/chesspairing/pairing/team"
	"github.com/gnutterts/chesspairing/scoring/football"
	scoringKeizer "github.com/gnutterts/chesspairing/scoring/keizer"
	"github.com/gnutterts/chesspairing/scoring/standard"
	"github.com/gnutterts/chesspairing/tiebreaker"
)

// NewPairer constructs a Pairer for the given system name. Names match
// the PairingSystem constants in the root package. opts may be nil for
// defaults.
//
// Returns an error if the name is unknown. The error includes the
// available names so callers can surface a useful diagnostic.
//
// Pairer dispatch is a closed switch on name; the closed set of pairing
// systems doesn't justify registry plumbing in every engine package. If
// the set ever crosses ~15 entries, revisit.
func NewPairer(name string, opts map[string]any) (cp.Pairer, error) {
	if opts == nil {
		opts = map[string]any{}
	}
	sys, err := cp.ParsePairingSystem(name)
	if err != nil {
		return nil, fmt.Errorf("%w; available: %v", err, PairerNames())
	}
	switch sys {
	case cp.PairingDutch:
		return dutch.NewFromMap(opts), nil
	case cp.PairingBurstein:
		return burstein.NewFromMap(opts), nil
	case cp.PairingDubov:
		return dubov.NewFromMap(opts), nil
	case cp.PairingLim:
		return lim.NewFromMap(opts), nil
	case cp.PairingDoubleSwiss:
		return doubleswiss.NewFromMap(opts), nil
	case cp.PairingTeam:
		return team.NewFromMap(opts), nil
	case cp.PairingKeizer:
		return keizer.NewFromMap(opts), nil
	case cp.PairingRoundRobin:
		return roundrobin.NewFromMap(opts), nil
	default:
		return nil, fmt.Errorf("unhandled pairing system %q (factory needs updating)", sys)
	}
}

// NewScorer constructs a Scorer for the given system name. Names match
// the ScoringSystem constants in the root package. opts may be nil for
// defaults.
//
// Same closed-switch reasoning as NewPairer.
func NewScorer(name string, opts map[string]any) (cp.Scorer, error) {
	if opts == nil {
		opts = map[string]any{}
	}
	sys, err := cp.ParseScoringSystem(name)
	if err != nil {
		return nil, fmt.Errorf("%w; available: %v", err, ScorerNames())
	}
	switch sys {
	case cp.ScoringStandard:
		return standard.NewFromMap(opts), nil
	case cp.ScoringKeizer:
		return scoringKeizer.NewFromMap(opts), nil
	case cp.ScoringFootball:
		return football.NewFromMap(opts), nil
	default:
		return nil, fmt.Errorf("unhandled scoring system %q (factory needs updating)", sys)
	}
}

// NewTieBreaker constructs a TieBreaker by ID via the self-registering
// tiebreaker registry. Returns the registry's error directly if the ID
// is unknown.
func NewTieBreaker(id string) (cp.TieBreaker, error) {
	return tiebreaker.Get(id)
}

// PairerNames returns the canonical names of every pairer this factory
// can construct, sorted alphabetically.
func PairerNames() []string {
	return []string{
		string(cp.PairingBurstein),
		string(cp.PairingDoubleSwiss),
		string(cp.PairingDubov),
		string(cp.PairingDutch),
		string(cp.PairingKeizer),
		string(cp.PairingLim),
		string(cp.PairingRoundRobin),
		string(cp.PairingTeam),
	}
}

// ScorerNames returns the canonical names of every scorer this factory
// can construct, sorted alphabetically.
func ScorerNames() []string {
	return []string{
		string(cp.ScoringFootball),
		string(cp.ScoringKeizer),
		string(cp.ScoringStandard),
	}
}

// TieBreakerIDs returns the IDs of every registered tiebreaker, sorted
// alphabetically.
func TieBreakerIDs() []string {
	ids := tiebreaker.All()
	sort.Strings(ids)
	return ids
}
