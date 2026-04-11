// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/factory.go
package main

import (
	"fmt"
	"strings"

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
)

// systemFlags maps CLI flag strings to PairingSystem constants.
var systemFlags = map[string]cp.PairingSystem{
	"--dutch":        cp.PairingDutch,
	"--burstein":     cp.PairingBurstein,
	"--dubov":        cp.PairingDubov,
	"--lim":          cp.PairingLim,
	"--double-swiss": cp.PairingDoubleSwiss,
	"--team":         cp.PairingTeam,
	"--keizer":       cp.PairingKeizer,
	"--roundrobin":   cp.PairingRoundRobin,
}

// systemAliases maps alternative flag names (bbpPairings / JaVaFo style) to
// their canonical form. All keys are stored lowercase for case-insensitive
// matching.
var systemAliases = map[string]string{
	"--fide-dutch":    "--dutch",
	"--fide-burstein": "--burstein",
	"--fide-dubov":    "--dubov",
	"--fide-lim":      "--lim",
	"--round-robin":   "--roundrobin",
	"--rr":            "--roundrobin",
	"--doubleswiss":   "--double-swiss",
}

// parseSystemFlag returns the PairingSystem for a CLI flag like "--dutch".
// Matching is case-insensitive and recognizes bbpPairings-style aliases
// (e.g. "--FIDE-Dutch", "--round-robin").
func parseSystemFlag(flag string) (cp.PairingSystem, bool) {
	// Try exact match first (fast path for the common case)
	if sys, ok := systemFlags[flag]; ok {
		return sys, true
	}

	lower := strings.ToLower(flag)

	// Try canonical flags case-insensitively
	if sys, ok := systemFlags[lower]; ok {
		return sys, true
	}

	// Try aliases
	if canonical, ok := systemAliases[lower]; ok {
		return systemFlags[canonical], true
	}
	return "", false
}

// newPairer creates a Pairer for the given system. opts may be nil for defaults.
func newPairer(system cp.PairingSystem, opts map[string]any) (cp.Pairer, error) {
	if opts == nil {
		opts = map[string]any{}
	}
	switch system {
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
		return nil, fmt.Errorf("unknown pairing system: %q", system)
	}
}

// newScorer creates a Scorer for the given system. opts may be nil for defaults.
func newScorer(system cp.ScoringSystem, opts map[string]any) (cp.Scorer, error) {
	if opts == nil {
		opts = map[string]any{}
	}
	switch system {
	case cp.ScoringStandard:
		return standard.NewFromMap(opts), nil
	case cp.ScoringKeizer:
		return scoringKeizer.NewFromMap(opts), nil
	case cp.ScoringFootball:
		return football.NewFromMap(opts), nil
	default:
		return nil, fmt.Errorf("unknown scoring system: %q", system)
	}
}
