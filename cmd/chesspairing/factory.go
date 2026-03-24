// cmd/chesspairing/factory.go
package main

import (
	"fmt"

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

// parseSystemFlag returns the PairingSystem for a CLI flag like "--dutch".
func parseSystemFlag(flag string) (cp.PairingSystem, bool) {
	sys, ok := systemFlags[flag]
	return sys, ok
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
