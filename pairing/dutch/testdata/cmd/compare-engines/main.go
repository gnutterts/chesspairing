//go:build ignore

// compare-engines compares golden data from multiple reference engines to find
// discrepancies in Dutch Swiss pairings.
//
// Usage: go run ./pairing/dutch/testdata/cmd/compare-engines/main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/gnutterts/chesspairing"
)

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Join(filepath.Dir(thisFile), "..", "..")

	engines := []struct {
		name string
		dir  string
	}{
		{"self", filepath.Join(baseDir, "golden")},
		{"javafo", filepath.Join(baseDir, "golden-javafo")},
		{"bbppairings", filepath.Join(baseDir, "golden-bbppairings")},
	}

	// Collect all scenario names across engines.
	scenarioSet := make(map[string]bool)
	for _, eng := range engines {
		entries, err := os.ReadDir(eng.dir)
		if err != nil {
			log.Printf("WARN: cannot read %s: %v", eng.dir, err)
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				scenarioSet[e.Name()] = true
			}
		}
	}
	scenarios := make([]string, 0, len(scenarioSet))
	for s := range scenarioSet {
		scenarios = append(scenarios, s)
	}
	sort.Strings(scenarios)

	totalDiffs := 0
	totalRounds := 0

	for _, scenario := range scenarios {
		// Find the maximum number of rounds across engines for this scenario.
		maxRound := 0
		for _, eng := range engines {
			rounds, _ := filepath.Glob(filepath.Join(eng.dir, scenario, "round-*.json"))
			if len(rounds) > maxRound {
				maxRound = len(rounds)
			}
		}

		for round := 1; round <= maxRound; round++ {
			results := make(map[string]*chesspairing.PairingResult)
			for _, eng := range engines {
				roundFile := filepath.Join(eng.dir, scenario, fmt.Sprintf("round-%d.json", round))
				data, err := os.ReadFile(roundFile) //nolint:gosec // tool
				if err != nil {
					continue
				}
				var pr chesspairing.PairingResult
				if err := json.Unmarshal(data, &pr); err != nil {
					log.Printf("WARN: %s/%s/round-%d.json: %v", eng.name, scenario, round, err)
					continue
				}
				results[eng.name] = &pr
			}

			if len(results) < 2 {
				continue
			}
			totalRounds++

			engineNames := make([]string, 0, len(results))
			for name := range results {
				engineNames = append(engineNames, name)
			}
			sort.Strings(engineNames)

			for i := range engineNames {
				for j := i + 1; j < len(engineNames); j++ {
					a := results[engineNames[i]]
					b := results[engineNames[j]]
					if diffs := comparePairings(a, b); len(diffs) > 0 {
						fmt.Printf("DIFF %s round %d: %s vs %s\n",
							scenario, round, engineNames[i], engineNames[j])
						for _, d := range diffs {
							fmt.Printf("  %s\n", d)
						}
						totalDiffs++
					}
				}
			}
		}
	}

	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Scenarios: %d\n", len(scenarios))
	fmt.Printf("Round comparisons: %d\n", totalRounds)
	if totalDiffs == 0 {
		fmt.Println("Result: All engines agree on all scenarios and rounds.")
	} else {
		fmt.Printf("Result: %d pairwise discrepancies found.\n", totalDiffs)
	}
}

func comparePairings(a, b *chesspairing.PairingResult) []string {
	var diffs []string

	if len(a.Pairings) != len(b.Pairings) {
		diffs = append(diffs, fmt.Sprintf("pairing count: %d vs %d", len(a.Pairings), len(b.Pairings)))
		return diffs
	}

	for i := range a.Pairings {
		ap := a.Pairings[i]
		bp := b.Pairings[i]
		if ap.Board != bp.Board || ap.WhiteID != bp.WhiteID || ap.BlackID != bp.BlackID {
			diffs = append(diffs, fmt.Sprintf("board %d: %s-%s vs %s-%s",
				ap.Board, ap.WhiteID, ap.BlackID, bp.WhiteID, bp.BlackID))
		}
	}

	// Compare byes.
	aByeIDs := byePlayerIDs(a)
	bByeIDs := byePlayerIDs(b)
	if strings.Join(aByeIDs, ",") != strings.Join(bByeIDs, ",") {
		diffs = append(diffs, fmt.Sprintf("byes: [%s] vs [%s]",
			strings.Join(aByeIDs, ","), strings.Join(bByeIDs, ",")))
	}

	return diffs
}

func byePlayerIDs(r *chesspairing.PairingResult) []string {
	ids := make([]string, len(r.Byes))
	for i, b := range r.Byes {
		ids[i] = b.PlayerID
	}
	sort.Strings(ids)
	return ids
}
