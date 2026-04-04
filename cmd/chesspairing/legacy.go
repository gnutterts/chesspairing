// cmd/chesspairing/legacy.go
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/trf"
)

// parsedLegacyArgs holds the parsed state from legacy CLI arguments.
type parsedLegacyArgs struct {
	system      cp.PairingSystem
	inputFile   string
	mode        string // "pair", "check", "generate"
	outputFile  string // for -p file or -g -o file
	seed        string // for -g -s
	configFile  string // for -g config
	showVersion bool   // -r
}

// parseLegacyArgs parses bbpPairings-compatible positional arguments.
// The args slice should NOT include the program name (os.Args[0]).
func parseLegacyArgs(args []string) (*parsedLegacyArgs, error) {
	p := &parsedLegacyArgs{}
	i := 0

	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "-r":
			p.showVersion = true
			i++

		case arg == "-w":
			// JaVaFo compat: accepted, ignored
			i++

		case arg == "-q":
			// JaVaFo compat: accepted, ignored. May have optional numeric argument.
			i++
			if i < len(args) && len(args[i]) > 0 && args[i][0] >= '0' && args[i][0] <= '9' {
				i++ // skip the numeric value
			}

		case arg == "-p":
			p.mode = "pair"
			i++
			// Optional output file follows
			if i < len(args) && !isFlag(args[i]) {
				p.outputFile = args[i]
				i++
			}

		case arg == "-c":
			p.mode = "check"
			i++

		case arg == "-g":
			p.mode = "generate"
			i++
			// Optional config file follows
			if i < len(args) && !isFlag(args[i]) {
				p.configFile = args[i]
				i++
			}

		case arg == "-o":
			i++
			if i >= len(args) {
				return nil, fmt.Errorf("-o requires a filename argument")
			}
			p.outputFile = args[i]
			i++

		case arg == "-s":
			i++
			if i >= len(args) {
				return nil, fmt.Errorf("-s requires a seed argument")
			}
			p.seed = args[i]
			i++

		default:
			// Try as system flag
			if sys, ok := parseSystemFlag(arg); ok {
				p.system = sys
				i++
				continue
			}
			// Try as input file (must not start with -)
			if p.inputFile == "" && !isFlag(arg) {
				p.inputFile = arg
				i++
				continue
			}
			return nil, fmt.Errorf("unexpected argument: %s", arg)
		}
	}

	return p, nil
}

func isFlag(s string) bool {
	return len(s) > 0 && s[0] == '-'
}

func runLegacy(args []string, stdout, stderr io.Writer) int {
	parsed, err := parseLegacyArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		printUsage(stderr)
		return ExitInvalidInput
	}

	// -r with no mode: just show version
	if parsed.showVersion && parsed.mode == "" {
		return runVersion(nil, stdout, stderr)
	}

	// Print version first if -r is combined with a mode flag
	if parsed.showVersion {
		runVersion(nil, stdout, stderr)
		fmt.Fprintln(stdout) // blank line separator
	}

	if parsed.system == "" {
		fmt.Fprintln(stderr, "error: system flag required (e.g. --dutch, --burstein)")
		printUsage(stderr)
		return ExitInvalidInput
	}

	switch parsed.mode {
	case "pair":
		return execPair(parsed, stdout, stderr)
	case "check":
		return execCheck(parsed, stdout, stderr)
	case "generate":
		return runGenerate(buildGenerateArgs(parsed), stdout, stderr)
	default:
		fmt.Fprintln(stderr, "error: mode flag required (-p, -c, or -g)")
		printUsage(stderr)
		return ExitInvalidInput
	}
}

func execPair(parsed *parsedLegacyArgs, stdout, stderr io.Writer) int {
	if parsed.inputFile == "" {
		fmt.Fprintln(stderr, "error: input file required")
		return ExitInvalidInput
	}

	// Open and parse TRF
	f, err := os.Open(parsed.inputFile)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot open %s: %v\n", parsed.inputFile, err)
		return ExitFileAccess
	}
	defer f.Close()

	doc, err := trf.Read(f)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot parse TRF: %v\n", err)
		return ExitInvalidInput
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot convert TRF to tournament state: %v\n", err)
		return ExitInvalidInput
	}

	// Override pairing system from CLI
	state.PairingConfig.System = parsed.system

	pairer, err := newPairer(parsed.system, state.PairingConfig.Options)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitInvalidInput
	}

	ctx := context.Background()
	result, err := pairer.Pair(ctx, state)
	if err != nil {
		fmt.Fprintf(stderr, "error: pairing failed: %v\n", err)
		return ExitNoPairing
	}

	// Build player ID → start number map for output
	playerNumbers := make(map[string]int, len(doc.Players))
	for _, pl := range doc.Players {
		playerNumbers[fmt.Sprintf("%d", pl.StartNumber)] = pl.StartNumber
	}

	// Determine output destination
	var out io.Writer = stdout
	if parsed.outputFile != "" {
		outF, err := os.Create(parsed.outputFile)
		if err != nil {
			fmt.Fprintf(stderr, "error: cannot create %s: %v\n", parsed.outputFile, err)
			return ExitFileAccess
		}
		defer outF.Close()
		out = outF
	}

	formatPairList(out, result, playerNumbers)
	return ExitSuccess
}

func execCheck(parsed *parsedLegacyArgs, stdout, stderr io.Writer) int {
	if parsed.inputFile == "" {
		fmt.Fprintln(stderr, "error: input file required")
		return ExitInvalidInput
	}

	f, err := os.Open(parsed.inputFile)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot open %s: %v\n", parsed.inputFile, err)
		return ExitFileAccess
	}
	defer f.Close()

	doc, err := trf.Read(f)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot parse TRF: %v\n", err)
		return ExitInvalidInput
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitInvalidInput
	}

	if len(state.Rounds) == 0 {
		fmt.Fprintln(stderr, "error: no rounds in tournament to check")
		return ExitInvalidInput
	}

	// Remove the last round and re-pair to check if it matches
	lastRound := state.Rounds[len(state.Rounds)-1]
	state.Rounds = state.Rounds[:len(state.Rounds)-1]
	state.CurrentRound = len(state.Rounds)

	state.PairingConfig.System = parsed.system

	pairer, err := newPairer(parsed.system, state.PairingConfig.Options)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitInvalidInput
	}

	ctx := context.Background()
	result, err := pairer.Pair(ctx, state)
	if err != nil {
		fmt.Fprintf(stderr, "error: pairing failed: %v\n", err)
		return ExitNoPairing
	}

	// Compare result pairings with the last round's games
	if pairingsMatch(result, &lastRound) {
		fmt.Fprintln(stdout, "OK: pairings match")
		return ExitSuccess
	}

	fmt.Fprintln(stdout, "MISMATCH: generated pairings differ from existing round")
	return ExitNoPairing
}

// pairingsMatch checks if a PairingResult matches an existing round's games.
// Comparison is by player IDs per board, ignoring board numbering differences.
func pairingsMatch(result *cp.PairingResult, round *cp.RoundData) bool {
	if len(result.Pairings) != len(round.Games) {
		return false
	}

	// Build set of pairs from result (order-independent)
	type pair struct{ w, b string }
	resultPairs := make(map[pair]bool, len(result.Pairings))
	for _, p := range result.Pairings {
		resultPairs[pair{p.WhiteID, p.BlackID}] = true
	}

	// Check all existing games are in result
	for _, g := range round.Games {
		if !resultPairs[pair{g.WhiteID, g.BlackID}] {
			return false
		}
	}

	// Check byes match
	if len(result.Byes) != len(round.Byes) {
		return false
	}
	resultByes := make(map[string]bool, len(result.Byes))
	for _, b := range result.Byes {
		resultByes[b.PlayerID] = true
	}
	for _, b := range round.Byes {
		if !resultByes[b.PlayerID] {
			return false
		}
	}

	return true
}

// buildGenerateArgs reconstructs args from parsed legacy state for runGenerate.
func buildGenerateArgs(p *parsedLegacyArgs) []string {
	var args []string
	if p.system != "" {
		// Find the flag string for this system.
		for flag, sys := range systemFlags {
			if sys == p.system {
				args = append(args, flag)
				break
			}
		}
	}
	if p.configFile != "" {
		args = append(args, "--config", p.configFile)
	}
	if p.outputFile != "" {
		args = append(args, "-o", p.outputFile)
	}
	if p.seed != "" {
		args = append(args, "-s", p.seed)
	}
	return args
}
