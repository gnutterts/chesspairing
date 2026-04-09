// cmd/chesspairing/standings.go
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/tiebreaker"
	"github.com/gnutterts/chesspairing/trf"
)

const standingsUsage = `Usage: chesspairing standings SYSTEM input-file [options]

Compute and display tournament standings.

Arguments:
  SYSTEM       Pairing system flag (required, for default tiebreaker selection):
               --dutch, --burstein, --dubov, --lim,
               --double-swiss, --team, --keizer, --roundrobin
  input-file   TRF16 tournament file, or "-" for stdin

Options:
  --scoring SYSTEM   Scoring system: standard, keizer, football (default: standard)
  --tiebreakers IDS  Comma-separated tiebreaker IDs (default: system-specific)
  --win N            Points for a win (overrides default)
  --draw N           Points for a draw
  --loss N           Points for a loss
  --forfeit-win N    Points for a forfeit win
  --bye N            Points for a bye
  --forfeit-loss N   Points for a forfeit loss
  --json             Output as JSON
  --help             Show this help

Exit codes:
  0  Success
  3  Invalid input
  5  File access error

Examples:
  chesspairing standings --dutch tournament.trf
  chesspairing standings --dutch tournament.trf --tiebreakers buchholz,wins
  chesspairing standings --dutch tournament.trf --json
  chesspairing standings --dutch - < tournament.trf
`

func runStandings(args []string, stdout, stderr io.Writer) int {
	// Check for --help before any parsing
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Fprint(stdout, standingsUsage)
			return ExitSuccess
		}
	}

	// First pass: extract system flag (before flag parsing, since --dutch etc. aren't flag-package flags)
	var system cp.PairingSystem
	var remaining []string
	for _, arg := range args {
		if sys, ok := parseSystemFlag(arg); ok {
			system = sys
		} else {
			remaining = append(remaining, arg)
		}
	}

	if system == "" {
		fmt.Fprintln(stderr, "error: system flag required (e.g. --dutch)")
		fmt.Fprintln(stderr, "usage: chesspairing standings SYSTEM input-file [options]")
		return ExitInvalidInput
	}

	flags, positional := separateFlags(remaining, map[string]bool{
		"--scoring": true, "--tiebreakers": true,
		"--win": true, "--draw": true, "--loss": true,
		"--forfeit-win": true, "--bye": true, "--forfeit-loss": true,
	})

	fs := flag.NewFlagSet("standings", flag.ContinueOnError)
	fs.SetOutput(stderr)
	scoring := fs.String("scoring", "standard", "scoring system: standard, keizer, football")
	tbFlag := fs.String("tiebreakers", "", "comma-separated tiebreaker IDs (default: system-specific)")
	jsonOut := fs.Bool("json", false, "output as JSON")
	win := fs.Float64("win", -1, "points for a win")
	draw := fs.Float64("draw", -1, "points for a draw")
	loss := fs.Float64("loss", -1, "points for a loss")
	forfeitWin := fs.Float64("forfeit-win", -1, "points for a forfeit win")
	bye := fs.Float64("bye", -1, "points for a bye")
	forfeitLoss := fs.Float64("forfeit-loss", -1, "points for a forfeit loss")

	if err := fs.Parse(flags); err != nil {
		return ExitInvalidInput
	}

	if len(positional) < 1 {
		fmt.Fprintln(stderr, "error: input file required")
		return ExitInvalidInput
	}

	inputFile := positional[0]

	rc, err := openInput(inputFile)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		if inputFile == "" {
			return ExitInvalidInput
		}
		return ExitFileAccess
	}
	defer func() { _ = rc.Close() }()

	doc, err := trf.Read(rc)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot parse TRF: %v\n", err)
		return ExitInvalidInput
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitInvalidInput
	}

	// Build scoring options from CLI flags (override TRF / defaults)
	scoringOpts := state.ScoringConfig.Options
	if scoringOpts == nil {
		scoringOpts = map[string]any{}
	}
	if *win >= 0 {
		scoringOpts["pointWin"] = *win
	}
	if *draw >= 0 {
		scoringOpts["pointDraw"] = *draw
	}
	if *loss >= 0 {
		scoringOpts["pointLoss"] = *loss
	}
	if *forfeitWin >= 0 {
		scoringOpts["pointForfeitWin"] = *forfeitWin
	}
	if *bye >= 0 {
		scoringOpts["pointBye"] = *bye
	}
	if *forfeitLoss >= 0 {
		scoringOpts["pointForfeitLoss"] = *forfeitLoss
	}

	scoringSystem := cp.ScoringSystem(*scoring)
	scorer, err := newScorer(scoringSystem, scoringOpts)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitInvalidInput
	}

	ctx := context.Background()
	scores, err := scorer.Score(ctx, state)
	if err != nil {
		fmt.Fprintf(stderr, "error: scoring failed: %v\n", err)
		return ExitUnexpected
	}

	// Determine tiebreakers
	var tbIDs []string
	if *tbFlag != "" {
		tbIDs = strings.Split(*tbFlag, ",")
	} else {
		tbIDs = cp.DefaultTiebreakers(system)
	}

	// Compute tiebreakers
	tbValues := make(map[string]map[string]float64) // tbID -> playerID -> value
	for _, tbID := range tbIDs {
		tb, err := tiebreaker.Get(tbID)
		if err != nil {
			fmt.Fprintf(stderr, "warning: unknown tiebreaker %q, skipping\n", tbID)
			continue
		}
		vals, err := tb.Compute(ctx, state, scores)
		if err != nil {
			fmt.Fprintf(stderr, "warning: tiebreaker %q failed: %v\n", tbID, err)
			continue
		}
		m := make(map[string]float64, len(vals))
		for _, v := range vals {
			m[v.PlayerID] = v.Value
		}
		tbValues[tbID] = m
	}

	// Build standings
	standings := buildStandings(state, scores, tbIDs, tbValues)

	if *jsonOut {
		if err := formatStandingsJSON(stdout, standings, *scoring, tbIDs); err != nil {
			fmt.Fprintf(stderr, "error: encoding JSON: %v\n", err)
			return ExitUnexpected
		}
	} else {
		formatStandingsText(stdout, standings)
	}

	return ExitSuccess
}

// buildStandings assembles Standing structs from scores and tiebreaker values,
// sorts them, and assigns shared ranks.
func buildStandings(state *cp.TournamentState, scores []cp.PlayerScore, tbIDs []string, tbValues map[string]map[string]float64) []cp.Standing {
	// Build player lookup
	playerMap := make(map[string]*cp.PlayerEntry, len(state.Players))
	for i := range state.Players {
		playerMap[state.Players[i].ID] = &state.Players[i]
	}

	// Build game stats per player
	type stats struct {
		played, wins, draws, losses int
	}
	gameStats := make(map[string]*stats)
	for _, pe := range state.Players {
		gameStats[pe.ID] = &stats{}
	}
	for _, rd := range state.Rounds {
		for _, g := range rd.Games {
			if g.Result == cp.ResultPending {
				continue
			}
			for _, pid := range []string{g.WhiteID, g.BlackID} {
				if s, ok := gameStats[pid]; ok {
					s.played++
				}
			}
			switch g.Result {
			case cp.ResultWhiteWins, cp.ResultForfeitWhiteWins:
				if s, ok := gameStats[g.WhiteID]; ok {
					s.wins++
				}
				if s, ok := gameStats[g.BlackID]; ok {
					s.losses++
				}
			case cp.ResultBlackWins, cp.ResultForfeitBlackWins:
				if s, ok := gameStats[g.BlackID]; ok {
					s.wins++
				}
				if s, ok := gameStats[g.WhiteID]; ok {
					s.losses++
				}
			case cp.ResultDraw:
				if s, ok := gameStats[g.WhiteID]; ok {
					s.draws++
				}
				if s, ok := gameStats[g.BlackID]; ok {
					s.draws++
				}
			}
		}
	}

	// Build standings
	standings := make([]cp.Standing, 0, len(scores))
	for _, ps := range scores {
		pe := playerMap[ps.PlayerID]
		if pe == nil {
			continue
		}

		var tbs []cp.NamedValue
		for _, tbID := range tbIDs {
			tb, err := tiebreaker.Get(tbID)
			if err != nil {
				continue
			}
			val := 0.0
			if m, ok := tbValues[tbID]; ok {
				val = m[ps.PlayerID]
			}
			tbs = append(tbs, cp.NamedValue{ID: tbID, Name: tb.Name(), Value: val})
		}

		gs := gameStats[ps.PlayerID]
		s := cp.Standing{
			PlayerID:    ps.PlayerID,
			DisplayName: pe.DisplayName,
			Score:       ps.Score,
			TieBreakers: tbs,
		}
		if gs != nil {
			s.GamesPlayed = gs.played
			s.Wins = gs.wins
			s.Draws = gs.draws
			s.Losses = gs.losses
		}
		standings = append(standings, s)
	}

	// Sort by score desc, then tiebreakers in order
	sort.SliceStable(standings, func(i, j int) bool {
		if standings[i].Score != standings[j].Score {
			return standings[i].Score > standings[j].Score
		}
		// Compare tiebreakers in order
		for k := range standings[i].TieBreakers {
			if k >= len(standings[j].TieBreakers) {
				break
			}
			if standings[i].TieBreakers[k].Value != standings[j].TieBreakers[k].Value {
				return standings[i].TieBreakers[k].Value > standings[j].TieBreakers[k].Value
			}
		}
		return false
	})

	// Assign shared ranks
	if len(standings) > 0 {
		standings[0].Rank = 1
		for i := 1; i < len(standings); i++ {
			if standings[i].Score == standings[i-1].Score && tiebreakersSame(standings[i], standings[i-1]) {
				standings[i].Rank = standings[i-1].Rank
			} else {
				standings[i].Rank = i + 1
			}
		}
	}

	return standings
}

func tiebreakersSame(a, b cp.Standing) bool {
	if len(a.TieBreakers) != len(b.TieBreakers) {
		return false
	}
	for i := range a.TieBreakers {
		if a.TieBreakers[i].Value != b.TieBreakers[i].Value {
			return false
		}
	}
	return true
}
