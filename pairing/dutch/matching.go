package dutch

import (
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// SplitS1S2 splits a homogeneous bracket into S1 (top half) and S2 (bottom half).
// For a bracket of n players: S1 = floor(n/2) top-ranked, S2 = remainder.
// Players must be sorted by TPN ascending.
func SplitS1S2(players []*swisslib.PlayerState) (s1, s2 []*swisslib.PlayerState) {
	n := len(players)
	mid := n / 2
	s1 = make([]*swisslib.PlayerState, mid)
	copy(s1, players[:mid])
	s2 = make([]*swisslib.PlayerState, n-mid)
	copy(s2, players[mid:])
	return
}

// SplitS1S2Heterogeneous splits a heterogeneous bracket into S1 (downfloaters)
// and S2 (native players).
func SplitS1S2Heterogeneous(bracket swisslib.Bracket) (s1, s2 []*swisslib.PlayerState) {
	floaterSet := make(map[string]bool, len(bracket.Downfloaters))
	for _, f := range bracket.Downfloaters {
		floaterSet[f.ID] = true
	}

	s1 = make([]*swisslib.PlayerState, 0, len(bracket.Downfloaters))
	s2 = make([]*swisslib.PlayerState, 0, len(bracket.Players)-len(bracket.Downfloaters))

	for _, p := range bracket.Players {
		if floaterSet[p.ID] {
			s1 = append(s1, p)
		} else {
			s2 = append(s2, p)
		}
	}
	return
}

// buildCandidate creates a Candidate from S1/S2 by pairing S1[i] with S2[i].
// Unmatched players (excess S1 or S2) become Residuals for sub-bracket pairing.
func buildCandidate(s1, s2 []*swisslib.PlayerState, downfloaterIDs map[string]bool, bracketScore float64) *swisslib.Candidate {
	pairs := len(s1)
	if len(s2) < pairs {
		pairs = len(s2)
	}

	cand := &swisslib.Candidate{
		Pairs:          make([]swisslib.ProposedPairing, pairs),
		DownfloaterIDs: downfloaterIDs,
		BracketScore:   bracketScore,
	}

	for i := 0; i < pairs; i++ {
		cand.Pairs[i] = swisslib.ProposedPairing{
			White:        s1[i],
			Black:        s2[i],
			BracketScore: bracketScore,
		}
	}

	// Unmatched players become residuals (for sub-bracket pairing).
	if len(s2) > len(s1) {
		cand.Residuals = s2[len(s1):]
	} else if len(s1) > len(s2) {
		cand.Residuals = s1[len(s2):]
	}

	return cand
}

// MatchBracketFeasible checks whether a bracket can produce at least one valid
// pairing. This is a lightweight version of MatchBracketMulti designed for C8
// look-ahead: it uses reduced search limits and only checks absolute criteria
// (forbidden pairs, C1, C3), returning true as soon as any valid pairing is found.
//
// The reduced limits prevent the combinatorial explosion that occurs when C8
// look-ahead runs the full matching algorithm on large merged brackets:
//   - Transpositions capped at 120 (vs 5040 in full matching)
//   - Exchanges capped at 50 (vs 500 in full matching)
//   - No deferred finalization (skips sub-bracket recursion)
//   - No optimization scoring (just absolute criteria)
func MatchBracketFeasible(bracket swisslib.Bracket, ctx *swisslib.CriteriaContext) bool {
	// Try heterogeneous split first.
	if matchBracketFeasibleWithSplit(bracket, ctx) {
		return true
	}

	// Homogeneous fallback.
	if !bracket.Homogeneous && len(bracket.Players) >= 2 {
		homoBracket := swisslib.Bracket{
			Players:       bracket.Players,
			Homogeneous:   true,
			OriginalScore: bracket.OriginalScore,
		}
		return matchBracketFeasibleWithSplit(homoBracket, ctx)
	}

	return false
}

// feasibleTranspositionLimit returns a reduced transposition cap for feasibility checks.
func feasibleTranspositionLimit(s2Size int) int {
	switch {
	case s2Size <= 4:
		return 24 // full 4!
	case s2Size <= 5:
		return 120 // full 5!
	default:
		return 120 // hard cap — enough for feasibility
	}
}

// matchBracketFeasibleWithSplit is the fast feasibility checker for a single
// S1/S2 split. Returns true if any valid pairing (satisfying absolute criteria) exists.
func matchBracketFeasibleWithSplit(bracket swisslib.Bracket, ctx *swisslib.CriteriaContext) bool {
	var s1, s2 []*swisslib.PlayerState
	if bracket.Homogeneous {
		s1, s2 = SplitS1S2(bracket.Players)
	} else {
		s1, s2 = SplitS1S2Heterogeneous(bracket)
	}

	maxTrans := feasibleTranspositionLimit(len(s2))

	const feasibleMaxExchanges = 50

	// p-reduction loop (C6): try pairing all S1 first, then fewer.
	for p := len(s1); p >= 1; p-- {
		currentS1 := s1[:p]

		// Check identity + transpositions.
		transpositions := GenerateTranspositions(s2, maxTrans)
		for _, t := range transpositions {
			cand := buildCandidate(currentS1, t, nil, bracket.OriginalScore)
			if swisslib.SatisfiesAbsolute(cand, ctx) && len(cand.Residuals) == 0 {
				return true // found a complete valid pairing
			}
		}

		// Check exchanges with reduced limits.
		exchanges := GenerateExchanges(currentS1, s2)
		exchangeLimit := len(exchanges)
		if exchangeLimit > feasibleMaxExchanges {
			exchangeLimit = feasibleMaxExchanges
		}
		for _, ex := range exchanges[:exchangeLimit] {
			exTranspositions := GenerateTranspositions(ex.S2, maxTrans)
			for _, t := range exTranspositions {
				cand := buildCandidate(ex.S1, t, nil, bracket.OriginalScore)
				if swisslib.SatisfiesAbsolute(cand, ctx) && len(cand.Residuals) == 0 {
					return true
				}
			}
		}
	}

	return false
}

// GenerateTranspositions generates permutations of s2 in lexicographic order
// by TPN. Capped at maxCount to handle large brackets.
func GenerateTranspositions(s2 []*swisslib.PlayerState, maxCount int) [][]*swisslib.PlayerState {
	n := len(s2)
	if n == 0 {
		return nil
	}

	// Start with identity permutation indices.
	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}

	// Add the first permutation (original order).
	first := make([]*swisslib.PlayerState, n)
	copy(first, s2)
	results := [][]*swisslib.PlayerState{first}

	// Generate permutations in lexicographic order (Narayana Pandita's algorithm).
	for len(results) < maxCount {
		i := n - 2
		for i >= 0 && perm[i] >= perm[i+1] {
			i--
		}
		if i < 0 {
			break
		}

		j := n - 1
		for perm[j] <= perm[i] {
			j--
		}

		perm[i], perm[j] = perm[j], perm[i]

		for left, right := i+1, n-1; left < right; left, right = left+1, right-1 {
			perm[left], perm[right] = perm[right], perm[left]
		}

		t := make([]*swisslib.PlayerState, n)
		for k, idx := range perm {
			t[k] = s2[idx]
		}
		results = append(results, t)
	}

	return results
}

// Exchange represents a swap of players between S1 and S2.
type Exchange struct {
	S1 []*swisslib.PlayerState
	S2 []*swisslib.PlayerState
}

// GenerateExchanges generates all possible exchanges between S1 and S2,
// ordered by the number of players swapped (fewer swaps first).
func GenerateExchanges(s1, s2 []*swisslib.PlayerState) []Exchange {
	n1 := len(s1)
	n2 := len(s2)

	var exchanges []Exchange

	maxSwap := n1
	if n2 < maxSwap {
		maxSwap = n2
	}

	for swapCount := 1; swapCount <= maxSwap; swapCount++ {
		s1Combos := combinations(s1, swapCount)
		s2Combos := combinations(s2, swapCount)

		for _, s1Swap := range s1Combos {
			for _, s2Swap := range s2Combos {
				newS1 := swapPlayers(s1, s1Swap, s2Swap)
				newS2 := swapPlayers(s2, s2Swap, s1Swap)
				exchanges = append(exchanges, Exchange{S1: newS1, S2: newS2})
			}
		}
	}

	return exchanges
}

// combinations generates all k-element subsets of the input slice.
func combinations(players []*swisslib.PlayerState, k int) [][]*swisslib.PlayerState {
	var result [][]*swisslib.PlayerState
	var combo []*swisslib.PlayerState
	var generate func(start int)
	generate = func(start int) {
		if len(combo) == k {
			c := make([]*swisslib.PlayerState, k)
			copy(c, combo)
			result = append(result, c)
			return
		}
		for i := start; i < len(players); i++ {
			combo = append(combo, players[i])
			generate(i + 1)
			combo = combo[:len(combo)-1]
		}
	}
	generate(0)
	return result
}

// swapPlayers creates a new slice with some players replaced.
func swapPlayers(original, out, in []*swisslib.PlayerState) []*swisslib.PlayerState {
	outSet := make(map[string]bool, len(out))
	for _, p := range out {
		outSet[p.ID] = true
	}

	result := make([]*swisslib.PlayerState, 0, len(original))
	for _, p := range original {
		if !outSet[p.ID] {
			result = append(result, p)
		}
	}
	result = append(result, in...)

	for i := 1; i < len(result); i++ {
		for j := i; j > 0 && result[j].TPN < result[j-1].TPN; j-- {
			result[j], result[j-1] = result[j-1], result[j]
		}
	}

	return result
}
