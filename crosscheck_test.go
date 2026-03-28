package chesspairing_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/doubleswiss"
	"github.com/gnutterts/chesspairing/pairing/dubov"
	"github.com/gnutterts/chesspairing/pairing/dutch"
	"github.com/gnutterts/chesspairing/pairing/lim"
	"github.com/gnutterts/chesspairing/pairing/team"
)

// normalizedPair is a pair of player IDs with the lower ID first.
type normalizedPair struct {
	a, b string
}

func (p normalizedPair) String() string {
	return fmt.Sprintf("{%s, %s}", p.a, p.b)
}

// normalizePairings extracts pairs from a PairingResult, normalizes each pair
// so the lower ID comes first, then sorts pairs lexicographically.
func normalizePairings(result *chesspairing.PairingResult) []normalizedPair {
	pairs := make([]normalizedPair, len(result.Pairings))
	for i, gp := range result.Pairings {
		a, b := gp.WhiteID, gp.BlackID
		if a > b {
			a, b = b, a
		}
		pairs[i] = normalizedPair{a, b}
	}
	slices.SortFunc(pairs, func(x, y normalizedPair) int {
		if c := strings.Compare(x.a, y.a); c != 0 {
			return c
		}
		return strings.Compare(x.b, y.b)
	})
	return pairs
}

// pairsEqual returns true if two normalized pair slices are identical.
func pairsEqual(a, b []normalizedPair) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// formatPairs produces a readable string for logging.
func formatPairs(pairs []normalizedPair) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = p.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// TestCrossSystem_Round1Consistency verifies that Swiss-type pairers produce
// consistent round-1 pairings for 6 evenly-spaced players.
//
// Swiss pairers fall into two architectural families:
//
//   - Fold-based (Dutch, Dubov, Lim): S1/S2 half-split where the top half
//     plays the bottom half. Expected: {p1,p4}, {p2,p5}, {p3,p6}.
//
//   - Lexicographic (Double-Swiss, Team): Art. 3.6 lexicographic enumeration
//     where the lowest-TPN unused participant pairs with the next available.
//     Expected: {p1,p2}, {p3,p4}, {p5,p6}.
//
// The test verifies consistency within each family and documents the known
// divergence between families.
func TestCrossSystem_Round1Consistency(t *testing.T) {
	// Build 6 players: p1=2500, p2=2400, ..., p6=2000.
	players := make([]chesspairing.PlayerEntry, 6)
	for i := range 6 {
		players[i] = chesspairing.PlayerEntry{
			ID:     fmt.Sprintf("p%d", i+1),
			Rating: 2500 - i*100,
			Active: true,
		}
	}

	state := &chesspairing.TournamentState{
		Players:      players,
		Rounds:       nil, // No previous rounds.
		CurrentRound: 1,
	}

	totalRounds := 5
	colorPrefA := "A"

	type namedPairer struct {
		name   string
		family string // "fold" or "lex"
		pairer chesspairing.Pairer
	}

	pairers := []namedPairer{
		{"Dutch", "fold", dutch.New(dutch.Options{})},
		{"Dubov", "fold", dubov.New(dubov.Options{})},
		{"Lim", "fold", lim.New(lim.Options{})},
		{"DoubleSwiss", "lex", doubleswiss.New(doubleswiss.Options{TotalRounds: &totalRounds})},
		{"Team", "lex", team.New(team.Options{TotalRounds: &totalRounds, ColorPreferenceType: &colorPrefA})},
	}

	ctx := context.Background()

	type systemResult struct {
		name   string
		family string
		pairs  []normalizedPair
		byes   []chesspairing.ByeEntry
		err    error
	}

	results := make([]systemResult, len(pairers))
	for i, p := range pairers {
		result, err := p.pairer.Pair(ctx, state)
		if err != nil {
			results[i] = systemResult{name: p.name, family: p.family, err: err}
			continue
		}
		results[i] = systemResult{
			name:   p.name,
			family: p.family,
			pairs:  normalizePairings(result),
			byes:   result.Byes,
		}
	}

	// Log all results.
	t.Log("Round-1 pairings for 6 players (p1=2500 .. p6=2000):")
	for _, r := range results {
		if r.err != nil {
			t.Logf("  %-12s [%-4s] ERROR: %v", r.name, r.family, r.err)
			continue
		}
		t.Logf("  %-12s [%-4s] pairs=%s  byes=%v", r.name, r.family, formatPairs(r.pairs), r.byes)
	}

	// Expected pairings per family.
	expectedFold := []normalizedPair{
		{"p1", "p4"},
		{"p2", "p5"},
		{"p3", "p6"},
	}
	expectedLex := []normalizedPair{
		{"p1", "p2"},
		{"p3", "p4"},
		{"p5", "p6"},
	}

	// Verify each system against its family's expected pairings.
	for _, r := range results {
		if r.err != nil {
			t.Errorf("%s failed to pair: %v", r.name, r.err)
			continue
		}

		var expected []normalizedPair
		switch r.family {
		case "fold":
			expected = expectedFold
		case "lex":
			expected = expectedLex
		}

		if !pairsEqual(r.pairs, expected) {
			t.Errorf("%s [%s] produced unexpected pairs: got %s, want %s",
				r.name, r.family, formatPairs(r.pairs), formatPairs(expected))
		}
	}

	// Cross-family consistency: verify fold and lex families differ as expected.
	t.Log("Cross-family note: fold-based and lexicographic systems use different " +
		"matching algorithms (S1/S2 half-split vs Art. 3.6 lexicographic enumeration), " +
		"producing structurally different round-1 pairings. This is correct per FIDE rules.")

	// Within-family consistency: verify all fold-based agree with each other.
	var foldRef *systemResult
	for i := range results {
		if results[i].family == "fold" && results[i].err == nil {
			foldRef = &results[i]
			break
		}
	}
	if foldRef != nil {
		for _, r := range results {
			if r.family != "fold" || r.err != nil || r.name == foldRef.name {
				continue
			}
			if !pairsEqual(r.pairs, foldRef.pairs) {
				t.Errorf("fold-based inconsistency: %s got %s, but %s got %s",
					r.name, formatPairs(r.pairs), foldRef.name, formatPairs(foldRef.pairs))
			}
		}
	}

	// Within-family consistency: verify all lex-based agree with each other.
	var lexRef *systemResult
	for i := range results {
		if results[i].family == "lex" && results[i].err == nil {
			lexRef = &results[i]
			break
		}
	}
	if lexRef != nil {
		for _, r := range results {
			if r.family != "lex" || r.err != nil || r.name == lexRef.name {
				continue
			}
			if !pairsEqual(r.pairs, lexRef.pairs) {
				t.Errorf("lex-based inconsistency: %s got %s, but %s got %s",
					r.name, formatPairs(r.pairs), lexRef.name, formatPairs(lexRef.pairs))
			}
		}
	}
}
