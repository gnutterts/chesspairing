// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/algorithm/blossom"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
	"github.com/gnutterts/chesspairing/trf"
)

// The reference below is a plain reading of C.04.4.2 Articles 3.2 and 4.3: it
// enumerates every pairing of a bracket (outgoing floaters are virtual
// partners that come last), keeps those that meet C1 to C5, and takes the
// first one, in the order of Article 4.3, that is best on C6, then C7, then C8.
// It shares nothing with the production search except the player data.

const refMaxCandidates = 200000

// refMaxBracket is the largest bracket the reference enumerates (ten players
// have 945 perfect matchings, many more with floaters; the whole comparison
// takes about half a minute on eight cores). BURSTEIN_REFERENCE_MAX_BRACKET
// lowers it for a quick local run.
var refMaxBracket = func() int {
	if v, err := strconv.Atoi(os.Getenv("BURSTEIN_REFERENCE_MAX_BRACKET")); err == nil && v > 0 {
		return v
	}
	return 10
}()

type refPlayer struct {
	*swisslib.PlayerState
	index OppositionIndex
}

// refCompatible: C1 (no rematch), C3 (no two equal absolute colour
// preferences) and the forbidden pairs.
func refCompatible(a, b *refPlayer, forbidden map[[2]string]bool) bool {
	for _, opp := range a.Opponents {
		if opp == b.ID {
			return false
		}
	}
	if forbidden[swisslib.CanonicalPairKey(a.ID, b.ID)] {
		return false
	}
	pa, pb := swisslib.ComputeColorPreference(a.ColorHistory), swisslib.ComputeColorPreference(b.ColorHistory)
	return !(pa.AbsolutePreference && pb.AbsolutePreference && pa.Color != nil && pb.Color != nil && *pa.Color == *pb.Color)
}

// refPairable: can the whole set be paired under C1 and C3 (Article 2.2.1)?
func refPairable(set []*refPlayer, forbidden map[[2]string]bool) bool {
	if len(set)%2 != 0 {
		return false
	}
	var edges []blossom.BlossomEdge
	for x := range set {
		for y := x + 1; y < len(set); y++ {
			if refCompatible(set[x], set[y], forbidden) {
				edges = append(edges, blossom.BlossomEdge{I: x, J: y, Weight: 1})
			}
		}
	}
	if len(set) == 0 {
		return true
	}
	mate := blossom.MaxWeightMatching(edges, true)
	if len(mate) < len(set) {
		return false
	}
	for _, m := range mate {
		if m < 0 {
			return false
		}
	}
	return true
}

// refEnumerate visits every pairing of bracket (in BSN order) that leaves
// exactly nf players unpaired, in the order of Article 4.3: BSN 1 meets the
// highest possible BSN first, then BSN 2, and so on; a floater comes last.
func refEnumerate(bracket []*refPlayer, nf int, forbidden map[[2]string]bool, visit func(pairs [][2]*refPlayer, floats []*refPlayer) bool) bool {
	used := make([]bool, len(bracket))
	var pairs [][2]*refPlayer
	var floats []*refPlayer
	var rec func() bool
	rec = func() bool {
		first := -1
		for k := range bracket {
			if !used[k] {
				first = k
				break
			}
		}
		if first < 0 {
			if len(floats) != nf {
				return true
			}
			return visit(pairs, floats)
		}
		used[first] = true
		for partner := len(bracket) - 1; partner > first; partner-- {
			if used[partner] || !refCompatible(bracket[first], bracket[partner], forbidden) {
				continue
			}
			used[partner] = true
			pairs = append(pairs, [2]*refPlayer{bracket[first], bracket[partner]})
			if !rec() {
				return false
			}
			pairs = pairs[:len(pairs)-1]
			used[partner] = false
		}
		if len(floats) < nf {
			floats = append(floats, bracket[first])
			if !rec() {
				return false
			}
			floats = floats[:len(floats)-1]
		}
		used[first] = false
		return true
	}
	return rec()
}

func refFloatScores(floats []*refPlayer) []float64 {
	scores := make([]float64, len(floats))
	for i, f := range floats {
		scores[i] = f.PairingScore
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(scores)))
	return scores
}

// refLess compares two score lists lexicographically, smaller is better; a
// shorter list that is a prefix of a longer one is better.
func refLess(a, b []float64) bool {
	for i := range a {
		if i >= len(b) {
			return false
		}
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return len(a) < len(b)
}

// refMaxPairs returns the largest number of pairs the bracket can make while
// the players it sends down can still be paired with the lower players (C5),
// and the best C6 value among the pairings with that number of pairs.
func refMaxPairs(bracket, lower []*refPlayer, forbidden map[[2]string]bool, budget *int) (int, []float64, bool) {
	for nf := len(bracket) % 2; nf <= len(bracket); nf += 2 {
		found := false
		var best []float64
		complete := refEnumerate(bracket, nf, forbidden, func(_ [][2]*refPlayer, floats []*refPlayer) bool {
			if *budget--; *budget < 0 {
				return false
			}
			if !refPairable(append(append([]*refPlayer{}, floats...), lower...), forbidden) {
				return true
			}
			if c6 := refFloatScores(floats); !found || refLess(c6, best) {
				best = c6
			}
			found = true
			return true
		})
		if !complete {
			return 0, nil, false
		}
		if found {
			return (len(bracket) - nf) / 2, best, true
		}
	}
	return 0, nil, false
}

type refCandidate struct {
	pairs   [][2]*refPlayer
	floats  []*refPlayer
	c6, c7  []float64
	c7pairs int
	c8      int
}

// refPair pairs the players (already without the bye) bracket by bracket. It
// reports false when a bracket is too large to enumerate.
func refPair(groups [][]*refPlayer, forbidden map[[2]string]bool) ([][2]*refPlayer, bool) {
	var result [][2]*refPlayer
	var incoming []*refPlayer
	for gi, group := range groups {
		bracket := append(append([]*refPlayer{}, incoming...), group...)
		sort.SliceStable(bracket, func(i, j int) bool {
			a, b := bracket[i].index, bracket[j].index
			if a.Buchholz != b.Buchholz {
				return a.Buchholz > b.Buchholz
			}
			if a.SonnebornBerger != b.SonnebornBerger {
				return a.SonnebornBerger > b.SonnebornBerger
			}
			return a.TPN < b.TPN
		})
		if len(bracket) > refMaxBracket {
			return nil, false
		}
		var lower, afterNext, next []*refPlayer
		if gi+1 < len(groups) {
			next = groups[gi+1]
		}
		for _, g := range groups[gi+1:] {
			lower = append(lower, g...)
		}
		for _, g := range groups[min(gi+2, len(groups)):] {
			afterNext = append(afterNext, g...)
		}
		budget := refMaxCandidates
		pairsNeeded, _, ok := refMaxPairs(bracket, lower, forbidden, &budget)
		if !ok {
			return nil, false
		}
		nf := len(bracket) - 2*pairsNeeded
		var best *refCandidate
		budget = refMaxCandidates
		// C5 and C7 depend on the outgoing floaters only, not on how the
		// others are paired, so they are evaluated once per floater set.
		type floaterOutcome struct {
			feasible bool
			c7pairs  int
			c7       []float64
		}
		outcomes := make(map[string]floaterOutcome)
		complete := refEnumerate(bracket, nf, forbidden, func(pairs [][2]*refPlayer, floats []*refPlayer) bool {
			if budget--; budget < 0 {
				return false
			}
			key := ""
			for _, f := range floats {
				key += f.ID + ","
			}
			outcome, seen := outcomes[key]
			if !seen {
				outcome.feasible = refPairable(append(append([]*refPlayer{}, floats...), lower...), forbidden)
				if outcome.feasible && next != nil {
					nextBracket := append(append([]*refPlayer{}, floats...), next...)
					if len(nextBracket) > refMaxBracket {
						budget = -1
						return false
					}
					inner := refMaxCandidates
					outcome.c7pairs, outcome.c7, _ = refMaxPairs(nextBracket, afterNext, forbidden, &inner)
				}
				outcomes[key] = outcome
			}
			if !outcome.feasible {
				return true
			}
			cand := &refCandidate{pairs: append([][2]*refPlayer{}, pairs...), floats: append([]*refPlayer{}, floats...), c6: refFloatScores(floats), c7pairs: outcome.c7pairs, c7: outcome.c7}
			for _, p := range pairs {
				a, b := swisslib.ComputeColorPreference(p[0].ColorHistory), swisslib.ComputeColorPreference(p[1].ColorHistory)
				if a.Color != nil && b.Color != nil && *a.Color == *b.Color {
					cand.c8++
				}
			}
			if best == nil || refBetter(cand, best) {
				best = cand
			}
			return true
		})
		if !complete || best == nil {
			return nil, false
		}
		result = append(result, best.pairs...)
		incoming = best.floats
	}
	return result, len(incoming) == 0
}

// refBetter is true when a beats b on C6, then C7 (number of pairs, then the
// floater scores of the next bracket), then C8. Ties keep the earlier one.
func refBetter(a, b *refCandidate) bool {
	if refLess(a.c6, b.c6) || refLess(b.c6, a.c6) {
		return refLess(a.c6, b.c6)
	}
	if a.c7pairs != b.c7pairs {
		return a.c7pairs > b.c7pairs
	}
	if refLess(a.c7, b.c7) || refLess(b.c7, a.c7) {
		return refLess(a.c7, b.c7)
	}
	return a.c8 < b.c8
}

func pairKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "-" + b
}

func readCorpusTournament(t *testing.T, file string) (*trf.Document, int) {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := trf.Read(bytes.NewReader(bytes.ReplaceAll(data, []byte{'\r'}, []byte{'\n'})))
	if err != nil {
		t.Fatalf("%s: %v", file, err)
	}
	rounds := 0
	for _, p := range doc.Players {
		rounds = max(rounds, len(p.Rounds))
	}
	return doc, rounds
}

// corpusRoundState builds the input for pairing round r of a corpus
// tournament: the history before r, and the requested byes of round r.
func corpusRoundState(t *testing.T, doc *trf.Document, total, r int) *chesspairing.TournamentState {
	t.Helper()
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatal(err)
	}
	state.Rounds = state.Rounds[:min(r-1, len(state.Rounds))]
	state.CurrentRound = r
	state.PreAssignedByes = nil
	for _, p := range doc.Players {
		if r > len(p.Rounds) {
			continue
		}
		rr := p.Rounds[r-1]
		if rr.Opponent != 0 || rr.Result == trf.ResultUnpaired {
			continue
		}
		kind := chesspairing.ByeAbsent
		switch rr.Result {
		case trf.ResultHalfBye:
			kind = chesspairing.ByeHalf
		case trf.ResultZeroBye:
			kind = chesspairing.ByeZero
		case trf.ResultFullBye:
			kind = chesspairing.ByePAB
		}
		state.PreAssignedByes = append(state.PreAssignedByes, chesspairing.ByeEntry{PlayerID: strconv.Itoa(p.StartNumber), Type: kind})
	}
	return state
}

// TestPostSeedingMatchesLiteralReference compares the production pairer with
// the enumeration above on every post-seeding round of the Dutch harness
// corpus whose brackets all have at most ten players. Larger rounds are
// checked for legality and determinism only.
func TestPostSeedingMatchesLiteralReference(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "..", "internal", "harness", "testdata", "corpus", "*.trf"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no corpus files: %v", err)
	}
	sort.Strings(files)
	var compared, legalOnly, unpairable atomic.Int64
	t.Run("corpus", func(t *testing.T) {
		for _, file := range files {
			t.Run(strings.TrimSuffix(filepath.Base(file), ".trf"), func(t *testing.T) {
				t.Parallel()
				doc, total := readCorpusTournament(t, file)
				for r := SeedingRounds(total) + 1; r <= total; r++ {
					switch compareRound(t, file, doc, total, r) {
					case roundCompared:
						compared.Add(1)
					case roundLegalOnly:
						legalOnly.Add(1)
					case roundNoPairing:
						unpairable.Add(1)
					}
				}
			})
		}
	})
	t.Logf("compared with the literal reference: %d rounds; legality only (bracket over %d players): %d; no pairing (error): %d", compared.Load(), refMaxBracket, legalOnly.Load(), unpairable.Load())
	if compared.Load() == 0 {
		t.Fatal("no round was compared")
	}
}

type roundOutcome int

const (
	roundCompared roundOutcome = iota
	roundLegalOnly
	roundNoPairing
)

func compareRound(t *testing.T, file string, doc *trf.Document, total, r int) roundOutcome {
	t.Helper()
	state := corpusRoundState(t, doc, total, r)
	pairer := New(Options{TotalRounds: &total})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Errorf("%s round %d: %v", filepath.Base(file), r, err)
		return roundNoPairing
	}
	again, err := pairer.Pair(context.Background(), state)
	if err != nil || formatPairings(result) != formatPairings(again) {
		t.Fatalf("%s round %d: not deterministic", filepath.Base(file), r)
	}
	assertLegal(t, file, r, state, result)

	filtered, _ := swisslib.FilterPreAssignedByes(state)
	players, err := swisslib.BuildPlayerStates(filtered)
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]*refPlayer, len(players))
	for i := range players {
		byID[players[i].ID] = &refPlayer{PlayerState: &players[i], index: ComputeOppositionIndex(&players[i], filtered)}
	}
	// The bye (Article 3.1) is taken as the pairer chose it.
	byeID := ""
	for _, b := range result.Byes {
		if b.Type == chesspairing.ByePAB {
			byeID = b.PlayerID
		}
	}
	var rest []swisslib.PlayerState
	for i := range players {
		if players[i].ID != byeID {
			rest = append(rest, players[i])
		}
	}
	var groups [][]*refPlayer
	for _, g := range swisslib.BuildScoreGroups(rest) {
		var members []*refPlayer
		for _, m := range g.Players {
			members = append(members, byID[m.ID])
		}
		groups = append(groups, members)
	}
	want, ok := refPair(groups, nil)
	if !ok {
		return roundLegalOnly
	}
	got := make(map[string]bool)
	for _, g := range result.Pairings {
		got[pairKey(g.WhiteID, g.BlackID)] = true
	}
	wantSet := make(map[string]bool)
	for _, p := range want {
		wantSet[pairKey(p[0].ID, p[1].ID)] = true
	}
	if len(got) != len(wantSet) {
		t.Errorf("%s round %d: %d pairs, reference %d", filepath.Base(file), r, len(got), len(wantSet))
		return roundCompared
	}
	for k := range wantSet {
		if !got[k] {
			t.Errorf("%s round %d: reference pairs %s, production does not", filepath.Base(file), r, k)
		}
	}
	return roundCompared
}

func assertLegal(t *testing.T, file string, r int, state *chesspairing.TournamentState, result *chesspairing.PairingResult) {
	t.Helper()
	filtered, pre := swisslib.FilterPreAssignedByes(state)
	players, err := swisslib.BuildPlayerStates(filtered)
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]*swisslib.PlayerState, len(players))
	for i := range players {
		byID[players[i].ID] = &players[i]
	}
	seen := make(map[string]bool)
	for _, g := range result.Pairings {
		a, b := byID[g.WhiteID], byID[g.BlackID]
		if a == nil || b == nil || seen[g.WhiteID] || seen[g.BlackID] || swisslib.HasPlayed(a, b) {
			t.Fatalf("%s round %d: illegal pairing %s-%s", filepath.Base(file), r, g.WhiteID, g.BlackID)
		}
		seen[g.WhiteID], seen[g.BlackID] = true, true
	}
	for _, b := range result.Byes {
		seen[b.PlayerID] = true
	}
	for _, b := range pre {
		seen[b.PlayerID] = true
	}
	if len(seen) != len(players)+len(pre) {
		t.Fatalf("%s round %d: %d players placed, want %d", filepath.Base(file), r, len(seen), len(players)+len(pre))
	}
}
