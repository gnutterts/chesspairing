// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/factory.go
package main

import (
	"strings"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/factory"
)

// systemFlags maps CLI flag strings to PairingSystem constants. The CLI
// uses double-dash flags ("--dutch") for system selection; the public
// chesspairing/factory package operates on bare names. This map is the
// shim between the two and is also used by the legacy mode to render
// system constants back into their preferred CLI flag.
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

// parseSystemFlag returns the PairingSystem for a CLI flag like "--dutch".
// Matching is case-insensitive and recognizes bbpPairings-style aliases
// (e.g. "--FIDE-Dutch", "--round-robin") via cp.ParsePairingSystem.
func parseSystemFlag(flag string) (cp.PairingSystem, bool) {
	if !strings.HasPrefix(flag, "--") {
		return "", false
	}
	sys, err := cp.ParsePairingSystem(strings.TrimPrefix(flag, "--"))
	if err != nil {
		return "", false
	}
	return sys, true
}

// newPairer creates a Pairer for the given system. opts may be nil for defaults.
func newPairer(system cp.PairingSystem, opts map[string]any) (cp.Pairer, error) {
	return factory.NewPairer(string(system), opts)
}

// newScorer creates a Scorer for the given system. opts may be nil for defaults.
func newScorer(system cp.ScoringSystem, opts map[string]any) (cp.Scorer, error) {
	return factory.NewScorer(string(system), opts)
}
