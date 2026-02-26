package dutch

import (
	"errors"
	"math/big"

	"github.com/gnutterts/chesspairing/blossom"
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

// DutchOptimizationCriteria returns the ordered list of optimization criterion
// functions for the Dutch system (C8-C21). Index maps to IdxC8..IdxC21.
func DutchOptimizationCriteria() []swisslib.OptimizationCriterion {
	criteria := make([]swisslib.OptimizationCriterion, swisslib.NumViolations)
	criteria[swisslib.IdxC8] = swisslib.CriterionC8
	criteria[swisslib.IdxC10] = swisslib.CriterionC10
	criteria[swisslib.IdxC11] = swisslib.CriterionC11
	criteria[swisslib.IdxC12] = swisslib.CriterionC12
	criteria[swisslib.IdxC13] = swisslib.CriterionC13
	criteria[swisslib.IdxC14] = swisslib.CriterionC14
	criteria[swisslib.IdxC15] = swisslib.CriterionC15
	criteria[swisslib.IdxC16] = swisslib.CriterionC16
	criteria[swisslib.IdxC17] = swisslib.CriterionC17
	criteria[swisslib.IdxC18] = swisslib.CriterionC18
	criteria[swisslib.IdxC19] = swisslib.CriterionC19
	criteria[swisslib.IdxC20] = swisslib.CriterionC20
	criteria[swisslib.IdxC21] = swisslib.CriterionC21
	return criteria
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

// RankedResult pairs a BracketPairing with its CandidateScore for backtracking.
type RankedResult struct {
	Pairing *swisslib.BracketPairing
	Score   swisslib.CandidateScore
}

// MatchBracket attempts to pair all players in a bracket using the candidate
// scoring model. Returns the best valid bracket pairing.
// This is a convenience wrapper around MatchBracketMulti that returns only the
// best result.
func MatchBracket(bracket swisslib.Bracket, ctx *swisslib.CriteriaContext, criteria []swisslib.OptimizationCriterion) (*swisslib.BracketPairing, error) {
	results := MatchBracketMulti(bracket, ctx, criteria, 1)
	if len(results) == 0 {
		return &swisslib.BracketPairing{
			Floaters: bracket.Players,
		}, errors.New("no valid pairing found for bracket")
	}
	return results[0].Pairing, nil
}

// MatchBracketFeasible checks whether a bracket can produce at least one valid
// pairing. This is a lightweight version of MatchBracketMulti designed for C8
// look-ahead: it uses reduced search limits and only checks absolute criteria
// (C1, C3), returning true as soon as any valid pairing is found.
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
// S1/S2 split. Returns true if any valid pairing (satisfying C1+C3) exists.
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

// MatchBracketMulti returns up to maxResults valid bracket pairings, ranked
// from best to worst by CandidateScore. Used by the backtracking orchestrator
// to try alternative floater combinations when downstream brackets fail.
//
// Performance: transpositions are capped based on bracket size (up to 5040).
// When maxResults == 1, early-exits on a perfect score.
//
// Algorithm (two-phase for candidates with residuals):
//
// Phase 1 (cheap): For each transposition/exchange, build candidate, check
// absolute criteria, and score PRE-finalization. Candidates with no residuals
// are directly finalized (trivial). Candidates with residuals are collected
// with their pre-scores.
//
// Phase 2 (expensive): Finalize only the top-N pre-scored candidates
// (recursive sub-bracket pairing for residuals), re-score post-finalization,
// and pick the best.
//
// Homogeneous fallback: For heterogeneous brackets, if the standard
// heterogeneous split (S1=downfloaters, S2=natives) fails to produce a
// complete pairing, a homogeneous split (top-half/bottom-half by TPN) is
// tried. This handles cases where bracket collapse creates an S1 larger
// than S2 (per FIDE C.04.3).
func MatchBracketMulti(bracket swisslib.Bracket, ctx *swisslib.CriteriaContext, criteria []swisslib.OptimizationCriterion, maxResults int) []RankedResult {
	// Try heterogeneous split first (standard).
	results := matchBracketWithSplit(bracket, ctx, criteria, maxResults)
	if len(results) > 0 {
		return results
	}

	// Homogeneous fallback: if this is a heterogeneous bracket and the
	// heterogeneous split failed, try treating it as homogeneous.
	if !bracket.Homogeneous && len(bracket.Players) >= 2 {
		homoBracket := swisslib.Bracket{
			Players:       bracket.Players,
			Homogeneous:   true,
			OriginalScore: bracket.OriginalScore,
			// No Downfloaters — treat all players equally.
		}
		return matchBracketWithSplit(homoBracket, ctx, criteria, maxResults)
	}

	return nil
}

// matchBracketWithSplit implements bracket matching using the Blossom algorithm.
// The S1/S2 split is no longer used for matching (Blossom handles it globally)
// but the function signature is preserved for compatibility.
func matchBracketWithSplit(bracket swisslib.Bracket, ctx *swisslib.CriteriaContext, criteria []swisslib.OptimizationCriterion, _ int) []RankedResult {
	return matchBracketBlossom(bracket, ctx, criteria)
}

// matchBracketBlossom pairs a bracket using the Blossom maximum weight
// matching algorithm. Handles both homogeneous and heterogeneous brackets
// uniformly.
//
// The edge weight encoding captures all FIDE criteria in a single Blossom run:
//  1. Criteria weight from C10-C21 (highest bits)
//  2. S1-S2 preference: edges between S1 and S2 get a bonus to prefer
//     cross-group pairings over same-group pairings
//  3. Natural pairing preference: within S1-S2 edges, prefer the "natural"
//     pairing S1[i]-S2[i], then adjacent S2 partners (C20 Minimize BSN diff)
//
// bitsToRepresent returns the number of bits needed to represent value v.
// Mirrors bbpPairings utility::typesizes::bitsToRepresent.
func bitsToRepresent(v int) int {
	if v <= 1 {
		return 1
	}
	bits := 0
	for v > 0 {
		bits++
		v >>= 1
	}
	return bits
}

// edgeKey returns the canonical (i < j) key for an edge.
func edgeKey(i, j int) [2]int {
	if i < j {
		return [2]int{i, j}
	}
	return [2]int{j, i}
}

// matchBracketBlossom implements the full bbpPairings bracket-level pairing
// algorithm (dutch.cpp lines 1087-1599). Seven phases:
//
//  1. Initial matching with edgeWeightComputer addend
//  2. (Heterogeneous only) Choose moved-down players — NOT YET IMPLEMENTED
//  3. Initialize remainder (collect unpaired score group players)
//  4. Exchange selection Phase 1 (select lower S1 for exchange)
//  5. Exchange selection Phase 2 (select higher S2 for exchange)
//  6. Finalize exchanges (zero non-participating edges)
//  7. Choose opponents (iterative finalizePair per matched S1 player)
//
// After all phases, remaining unmatched players become floaters.
func matchBracketBlossom(bracket swisslib.Bracket, ctx *swisslib.CriteriaContext, criteria []swisslib.OptimizationCriterion) []RankedResult {
	players := bracket.Players
	n := len(players)
	if n < 2 {
		return []RankedResult{{
			Pairing: &swisslib.BracketPairing{Floaters: players},
		}}
	}

	if ctx.DeadlineExceeded() {
		return nil
	}

	// Deduplicate players by ID.
	seen := make(map[string]bool, n)
	deduped := make([]*swisslib.PlayerState, 0, n)
	for _, p := range players {
		if !seen[p.ID] {
			seen[p.ID] = true
			deduped = append(deduped, p)
		}
	}
	players = deduped
	n = len(players)
	if n < 2 {
		return []RankedResult{{
			Pairing: &swisslib.BracketPairing{Floaters: players},
		}}
	}

	// Build downfloater ID set for float criteria.
	downfloaterIDs := make(map[string]bool, len(bracket.Downfloaters))
	for _, p := range bracket.Downfloaters {
		downfloaterIDs[p.ID] = true
	}

	// Compute downfloater count and S1 size.
	// dfCount = actual number of downfloaters (maps to bbpPairings' scoreGroupBegin).
	//   0 for homogeneous brackets.
	// s1Size = S1/S2 split for edgeWeightComputer.
	//   n/2 for homogeneous, dfCount for heterogeneous.
	var dfCount int
	for _, p := range players {
		if downfloaterIDs[p.ID] {
			dfCount++
		}
	}

	var s1Size int
	if bracket.Homogeneous || dfCount == 0 {
		s1Size = n / 2
	} else {
		s1Size = dfCount
	}

	// scoreGroupSizeBits: bits needed to represent the score group size.
	// Mirrors bbpPairings' scoreGroupSizeBits = bitsToRepresent(maxScoreGroupSize).
	sgSizeBits := bitsToRepresent(n)

	// =========================================================================
	// Precompute base criteria weights for all legal pairs.
	// baseWeight[i][j] stores the criteria weight (0 means illegal pair).
	// =========================================================================
	baseWeight := make([][]int64, n)
	for i := 0; i < n; i++ {
		baseWeight[i] = make([]int64, n)
	}

	for i := 0; i < n; i++ {
		if i%100 == 0 && ctx.DeadlineExceeded() {
			return nil
		}
		for j := i + 1; j < n; j++ {
			pair := &swisslib.ProposedPairing{
				White:        players[i],
				Black:        players[j],
				BracketScore: bracket.OriginalScore,
			}
			if !swisslib.C1NoRematches(pair, ctx) {
				continue
			}
			if !swisslib.C3AbsoluteColorConflict(pair, ctx) {
				continue
			}
			w := swisslib.PairEdgeWeight(pair, ctx, downfloaterIDs, bracket.OriginalScore)
			if w == 0 {
				w = 1 // Distinguish legal-zero from illegal-zero.
			}
			baseWeight[i][j] = w
			baseWeight[j][i] = w
		}
	}

	// edgeWeightComputer mirrors bbpPairings lines 1055-1085.
	// It adds exchange-preference bits on top of the base criteria weight:
	//   - isS1 bit: 1 if smallerIdx is in S1 (remainderIdx < remainderPairs)
	//   - BSN distance: shifted by 2*sgSizeBits, subtracted by remainderIdx
	//   - Reserve 1 bit for exchange optimization tweaks
	//
	// Parameters:
	//   smallerIdx, largerIdx: indices into players[] (smallerIdx < largerIdx)
	//   smallerRemIdx: remainder-relative index of the smaller player
	//   remPairs: number of S1 players in the remainder
	edgeWeightComputer := func(smallerIdx, largerIdx, smallerRemIdx, remPairs int) int64 {
		bw := baseWeight[largerIdx][smallerIdx]
		if bw == 0 {
			return 0
		}

		// Build addend matching bbpPairings' edgeWeightComputer.
		var addend int64

		// Minimize exchanges: 1 if this is an S1 player (natural pairing).
		if smallerRemIdx < remPairs {
			addend = 1
		}

		// Minimize BSN distance: shift left by 2*sgSizeBits, subtract index.
		addend <<= sgSizeBits
		addend <<= sgSizeBits
		addend -= int64(smallerRemIdx)

		// Reserve 1 bit for exchange optimization.
		addend <<= 1

		return bw + addend
	}

	// =========================================================================
	// Build mutable edge weight map. Updated by each phase.
	// edgeW[edgeKey(i,j)] holds the current edge weight. 0 = removed.
	// =========================================================================
	edgeW := make(map[[2]int]int64, n*n/2)

	// Phase 1 uses raw baseWeights (no edgeWeightComputer addend).
	// bbpPairings lines 1027-1048: set edge weights = baseEdgeWeights.
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if baseWeight[i][j] > 0 {
				edgeW[edgeKey(i, j)] = baseWeight[i][j]
			}
		}
	}

	// Helper: build edges slice from edgeW (excluding finalized vertices).
	finalized := make([]bool, n)
	matched := make([]bool, n) // Phase 2: tracks which players will be matched

	buildEdges := func() []blossom.BlossomEdge {
		edges := make([]blossom.BlossomEdge, 0, len(edgeW))
		for k, w := range edgeW {
			if w > 0 && !finalized[k[0]] && !finalized[k[1]] {
				edges = append(edges, blossom.BlossomEdge{
					I: k[0], J: k[1], Weight: w,
				})
			}
		}
		return edges
	}

	// Helper: run Blossom matching. Returns mate array.
	runBlossom := func() []int {
		edges := buildEdges()
		if len(edges) == 0 {
			return nil
		}
		return blossom.MaxWeightMatching(edges, true)
	}

	// Helper: finalizePair — lock a pair permanently.
	// Mirrors bbpPairings finalizePair (common.h): sets the edge to maxEdgeWeight
	// and zeros ALL other edges incident on both vertices.
	finalizePairFn := func(a, b int) {
		finalized[a] = true
		finalized[b] = true
		for k := range edgeW {
			if k[0] == a || k[1] == a || k[0] == b || k[1] == b {
				edgeW[k] = 0
			}
		}
	}

	// =========================================================================
	// Phase 1: Initial matching (line 1087).
	// Run Blossom with raw base weights (no exchange addend).
	// =========================================================================
	mate := runBlossom()

	// =========================================================================
	// Phase 2: Choose moved-down players (lines 1091-1255).
	// Only for heterogeneous brackets (s1Size > 0 and not homogeneous).
	//
	// Sub-phase 2a: Determine which downfloaters will pair in this bracket.
	// Sub-phase 2b: Choose opponents for each matched downfloater, finalize pairs.
	// After Phase 2, finalized downfloaters and their opponents are removed.
	// =========================================================================
	bp := &swisslib.BracketPairing{}

	isHeterogeneous := dfCount > 0 && !bracket.Homogeneous

	if isHeterogeneous && len(mate) > 0 {
		// Sub-phase 2a: Decide which downfloaters get matched.
		// bbpPairings lines 1107-1205.
		//
		// Downfloaters are in players[0..dfCount-1]. They may have different scores
		// (from collapsed brackets). Process in score-group order.
		// A downfloater is "matched" if mate[i] is a resident (>= dfCount).

		type dfScoreGroup struct {
			startIdx int     // first downfloater index in this score subgroup
			score    float64 // score of this subgroup
		}
		var dfGroups []dfScoreGroup
		for i := 0; i < dfCount; i++ {
			if len(dfGroups) == 0 || players[i].Score != dfGroups[len(dfGroups)-1].score {
				dfGroups = append(dfGroups, dfScoreGroup{startIdx: i, score: players[i].Score})
			}
		}

		for _, dfg := range dfGroups {
			// Find the end of this score subgroup.
			dfEnd := dfCount
			for k := dfg.startIdx + 1; k < dfCount; k++ {
				if players[k].Score != dfg.score {
					dfEnd = k
					break
				}
			}

			// Count remaining and matched for this subgroup.
			remainingDF := 0
			remainingMatchedDF := 0
			for k := dfg.startIdx; k < dfEnd; k++ {
				remainingDF++
				if k < len(mate) && mate[k] >= dfCount && mate[k] < n {
					remainingMatchedDF++
				}
			}

			for i := dfg.startIdx; i < dfEnd; i++ {
				if remainingMatchedDF == 0 {
					// No more can be matched in this subgroup.
					continue
				}

				if remainingDF <= remainingMatchedDF {
					// All remaining can be matched — mark optimistically.
					matched[i] = true
					if i < len(mate) && mate[i] >= dfCount && mate[i] < n {
						remainingMatchedDF--
					}
					remainingDF--
					continue
				}

				remainingDF--

				// This downfloater is NOT currently matched to a resident.
				// Try to force-match it by setting LSB=1 on its edges to residents.
				if i >= len(mate) || mate[i] < dfCount || mate[i] >= n {
					for j := dfCount; j < n; j++ {
						bw := baseWeight[j][i]
						if bw > 0 {
							edgeW[edgeKey(i, j)] = bw | 1
						}
					}
					mate = runBlossom()
				}

				if i < len(mate) && mate[i] >= dfCount && mate[i] < n {
					// Successfully matched — mark and make the match sticky.
					matched[i] = true
					remainingMatchedDF--

					// bbpPairings lines 1192-1203: set edge weight = base | (sgSize + 1)
					// to make this pair sticky in subsequent Blossom runs.
					sgSize := n - dfCount // score group size (number of residents)
					for j := dfCount; j < n; j++ {
						bw := baseWeight[j][i]
						if bw > 0 {
							edgeW[edgeKey(i, j)] = bw | int64(sgSize+1)
						}
					}
				}
			}
		}

		// Sub-phase 2b: Choose opponents for each matched downfloater.
		// bbpPairings lines 1207-1255.
		for i := 0; i < dfCount; i++ {
			if !matched[i] {
				continue
			}

			// Add addends: iterate residents from bottom (highest index) to top.
			// Skip already-matched residents. Higher-ranked (lower index) residents
			// get larger addend = more preferred.
			// bbpPairings: addend starts at playersByIndex.size() (= n).
			addend := int64(n)
			for j := n - 1; j >= dfCount; j-- {
				if matched[j] {
					continue
				}
				bw := baseWeight[j][i]
				if bw > 0 {
					edgeW[edgeKey(i, j)] = bw + addend
					addend++
				}
			}

			mate = runBlossom()

			// Finalize the pair.
			if i < len(mate) && mate[i] >= dfCount && mate[i] < n {
				partner := mate[i]
				matched[partner] = true
				bp.Pairs = append(bp.Pairs, swisslib.ProposedPairing{
					White:        players[i],
					Black:        players[partner],
					BracketScore: bracket.OriginalScore,
				})
				finalizePairFn(i, partner)
			}
		}
	}

	// =========================================================================
	// Phase 3: Initialize remainder (lines 1257-1291).
	// Collect score group players NOT matched with a downfloater.
	// Count remainderPairs: S1 players matched S1→S2 within the remainder.
	//
	// In bbpPairings, the remainder only contains scoreGroup players (residents),
	// i.e., players from scoreGroupBegin to nextScoreGroupBegin, excluding
	// those finalized (paired with a downfloater).
	// For homogeneous brackets, dfCount=0, so ALL players enter the remainder.
	// =========================================================================
	remainder := make([]int, 0, n)
	for i := dfCount; i < n; i++ {
		if finalized[i] {
			continue // Already paired with a downfloater in Phase 2
		}
		remainder = append(remainder, i)
	}

	// Count remainderPairs from the Phase 1 matching.
	// bbpPairings line 1281: count players where stableMatching[vertex] < vertex,
	// i.e., matched to a lower-indexed player. Each pair contributes one such
	// player, so remainderPairs = number of matched pairs in the remainder.
	remainderPairs := 0
	remIndexOf := make(map[int]int, len(remainder))
	for ri, pi := range remainder {
		remIndexOf[pi] = ri
	}
	for _, pi := range remainder {
		if pi < len(mate) && mate[pi] >= 0 && mate[pi] < pi {
			// mate[pi] < pi means matched to lower index = one half of a pair
			if _, inRem := remIndexOf[mate[pi]]; inRem {
				remainderPairs++
			}
		}
	}
	// Fallback: if no pairs found from matching, use s1Size (n/2 for homogeneous,
	// dfCount for heterogeneous). This handles edge cases where Phase 1 matching
	// with raw base weights didn't produce clean pairs.
	if remainderPairs == 0 {
		remainderPairs = s1Size
		// For heterogeneous brackets, s1Size = dfCount, but after Phase 2 the
		// downfloaters are removed. Clamp to half the remainder.
		if remainderPairs > len(remainder)/2 {
			remainderPairs = len(remainder) / 2
		}
	}

	// Now apply edgeWeightComputer to remainder edges (lines 1293-1317).
	// Reset ALL edges first, then set remainder edges with addend.
	for k := range edgeW {
		edgeW[k] = 0
	}
	for ri, pi := range remainder {
		for rj := ri + 1; rj < len(remainder); rj++ {
			pj := remainder[rj]
			w := edgeWeightComputer(pi, pj, ri, remainderPairs)
			if w > 0 {
				edgeW[edgeKey(pi, pj)] = w
			}
		}
	}

	// Re-run Blossom on the remainder with exchange-preference weights.
	mate = runBlossom()

	// =========================================================================
	// Phase 4: Exchange selection Phase 1 (lines 1320-1405).
	// Select lower S1 players to be exchanged.
	//
	// In the remainder, S1 = first remainderPairs players, S2 = the rest.
	// Count how many S1 players are NOT matched S1→S2 (these must exchange).
	// =========================================================================
	exchangeCount := 0
	for ri := 0; ri < remainderPairs; ri++ {
		pi := remainder[ri]
		// Matched S1→S2: mate[pi] maps to a remainder player with higher index.
		matchedToS2 := false
		if pi < len(mate) && mate[pi] >= 0 && mate[pi] < n {
			mateRI, inRem := remIndexOf[mate[pi]]
			if inRem && mateRI > ri {
				matchedToS2 = true
			}
		}
		if !matchedToS2 {
			exchangeCount++
		}
	}

	if exchangeCount > 0 && remainderPairs > 0 {
		exchangesRemaining := exchangeCount
		for ri := remainderPairs - 1; ri >= 0 && exchangesRemaining > 0; ri-- {
			if ctx.DeadlineExceeded() {
				break
			}
			pi := remainder[ri]

			// Check if this player is matched S1→S2 in remainder.
			isMatchedS1S2 := false
			if pi < len(mate) && mate[pi] >= 0 && mate[pi] < n {
				mateRI, inRem := remIndexOf[mate[pi]]
				if inRem && mateRI > ri {
					isMatchedS1S2 = true
				}
			}

			if isMatchedS1S2 {
				// Try to prevent exchange: subtract 1 from edges to opponents.
				for rj := ri + 1; rj < len(remainder); rj++ {
					pj := remainder[rj]
					w := edgeWeightComputer(pi, pj, ri, remainderPairs)
					if w > 0 {
						w--
						edgeW[edgeKey(pi, pj)] = w
					}
				}
				mate = runBlossom()
			}

			// Check if exchange happened.
			exchange := true
			if pi < len(mate) && mate[pi] >= 0 && mate[pi] < n {
				mateRI, inRem := remIndexOf[mate[pi]]
				if inRem && mateRI > ri {
					exchange = false
				}
			}

			if exchange {
				exchangesRemaining--
			}

			// Restore or zero edge weights.
			for rj := ri + 1; rj < len(remainder); rj++ {
				pj := remainder[rj]
				if exchange {
					// Zero ALL opponent edges for exchanged player.
					baseWeight[pj][pi] = 0
					baseWeight[pi][pj] = 0
				}
				w := edgeWeightComputer(pi, pj, ri, remainderPairs)
				edgeW[edgeKey(pi, pj)] = w
			}
		}
	}

	// =========================================================================
	// Phase 5: Exchange selection Phase 2 (lines 1407-1507).
	// Select higher S2 players to be exchanged.
	// =========================================================================
	if exchangeCount > 0 {
		exchangesRemaining := exchangeCount
		for ri := remainderPairs; ri < len(remainder) && exchangesRemaining > 1; ri++ {
			if ctx.DeadlineExceeded() {
				break
			}
			pi := remainder[ri]

			// Already exchanged = matched to a HIGHER remainder index.
			alreadyExchanged := false
			if pi < len(mate) && mate[pi] >= 0 && mate[pi] < n {
				mateRI, inRem := remIndexOf[mate[pi]]
				if inRem && mateRI > ri {
					alreadyExchanged = true
				}
			}

			if !alreadyExchanged {
				// Try to force exchange: add 1 to edges with higher-index opponents.
				for rj := ri + 1; rj < len(remainder); rj++ {
					pj := remainder[rj]
					w := edgeWeightComputer(pi, pj, ri, remainderPairs)
					if w > 0 {
						w++
						edgeW[edgeKey(pi, pj)] = w
					}
				}
				mate = runBlossom()
			}

			// Check if exchanged.
			exchanged := false
			if pi < len(mate) && mate[pi] >= 0 && mate[pi] < n {
				mateRI, inRem := remIndexOf[mate[pi]]
				if inRem && mateRI > ri {
					exchanged = true
				}
			}

			if exchanged {
				exchangesRemaining--
				// Zero edges to players before this player.
				for rj := 0; rj < ri; rj++ {
					pj := remainder[rj]
					baseWeight[pi][pj] = 0
					baseWeight[pj][pi] = 0
					edgeW[edgeKey(pj, pi)] = 0
				}
			}

			if !alreadyExchanged {
				// Restore original edge weights.
				for rj := ri + 1; rj < len(remainder); rj++ {
					pj := remainder[rj]
					w := edgeWeightComputer(pi, pj, ri, remainderPairs)
					edgeW[edgeKey(pi, pj)] = w
				}
			}
		}
	}

	// =========================================================================
	// Phase 6: Finalize exchanges (lines 1509-1543).
	// =========================================================================
	for ri := 0; ri < len(remainder); ri++ {
		pi := remainder[ri]
		for rj := ri + 1; rj < len(remainder); rj++ {
			pj := remainder[rj]

			iNotMatchedDown := pi >= len(mate) || mate[pi] <= pi || mate[pi] >= n
			if !iNotMatchedDown {
				// Also check mate is in remainder and has higher remainder index.
				mateRI, inRem := remIndexOf[mate[pi]]
				if !inRem || mateRI <= ri {
					iNotMatchedDown = true
				}
			}

			jHasNaturalPair := false
			if pj < len(mate) && mate[pj] > pj && mate[pj] < n {
				mateRI, inRem := remIndexOf[mate[pj]]
				if inRem && mateRI > rj {
					jHasNaturalPair = true
				}
			}

			if iNotMatchedDown || jHasNaturalPair {
				baseWeight[pj][pi] = 0
				baseWeight[pi][pj] = 0
			}
			edgeW[edgeKey(pi, pj)] = baseWeight[pj][pi]
		}
	}

	// =========================================================================
	// Phase 7: Choose opponents (lines 1545-1599).
	// For each remainder S1 player matched S1→S2 (in order), add addends,
	// run Blossom, finalize the pair.
	// =========================================================================
	for ri := 0; ri < len(remainder); ri++ {
		pi := remainder[ri]
		if ctx.DeadlineExceeded() {
			break
		}

		// Only process players matched with a higher-remainder-index partner.
		if pi >= len(mate) || mate[pi] == -1 {
			continue
		}
		mateRI, inRem := remIndexOf[mate[pi]]
		if !inRem || mateRI <= ri {
			continue
		}

		// Add addends: iterate opponents in reverse remainder order, incrementing.
		addend := int64(0)
		for rj := len(remainder) - 1; rj > ri; rj-- {
			pj := remainder[rj]
			if finalized[pj] {
				continue
			}
			bw := baseWeight[pj][pi]
			if bw == 0 {
				continue
			}
			edgeW[edgeKey(pi, pj)] = bw + addend
			addend++
		}

		mate = runBlossom()

		// Finalize the pair.
		partner := -1
		if pi < len(mate) {
			partner = mate[pi]
		}
		if partner >= 0 && partner < n && !finalized[partner] {
			bp.Pairs = append(bp.Pairs, swisslib.ProposedPairing{
				White:        players[pi],
				Black:        players[partner],
				BracketScore: bracket.OriginalScore,
			})
			finalizePairFn(pi, partner)
		}
	}

	// =========================================================================
	// Pair remaining unfinalized players via one final Blossom run.
	// =========================================================================
	remainingIdx := make([]int, 0)
	for i := 0; i < n; i++ {
		if !finalized[i] {
			remainingIdx = append(remainingIdx, i)
		}
	}
	if len(remainingIdx) >= 2 {
		remEdges := make([]blossom.BlossomEdge, 0)
		for a := 0; a < len(remainingIdx); a++ {
			for b := a + 1; b < len(remainingIdx); b++ {
				oi, oj := remainingIdx[a], remainingIdx[b]
				bw := baseWeight[oi][oj]
				if bw > 0 {
					remEdges = append(remEdges, blossom.BlossomEdge{
						I: a, J: b, Weight: bw,
					})
				}
			}
		}
		if len(remEdges) > 0 {
			remMate := blossom.MaxWeightMatching(remEdges, true)
			remMatched := make([]bool, len(remainingIdx))
			for a := 0; a < len(remainingIdx); a++ {
				if remMatched[a] {
					continue
				}
				if a >= len(remMate) || remMate[a] == -1 {
					continue
				}
				b := remMate[a]
				if b >= len(remainingIdx) || remMatched[b] {
					continue
				}
				remMatched[a] = true
				remMatched[b] = true
				bp.Pairs = append(bp.Pairs, swisslib.ProposedPairing{
					White:        players[remainingIdx[a]],
					Black:        players[remainingIdx[b]],
					BracketScore: bracket.OriginalScore,
				})
				finalized[remainingIdx[a]] = true
				finalized[remainingIdx[b]] = true
			}
		}
	}

	// Unmatched players become floaters.
	for i := 0; i < n; i++ {
		if !finalized[i] {
			bp.Floaters = append(bp.Floaters, players[i])
		}
	}

	// Score the result.
	cand := &swisslib.Candidate{
		Pairs:          bp.Pairs,
		Floaters:       bp.Floaters,
		DownfloaterIDs: downfloaterIDs,
		BracketScore:   bracket.OriginalScore,
	}
	score := swisslib.ScoreCandidate(cand, ctx, criteria)

	return []RankedResult{{Pairing: bp, Score: score}}
}

// pairBracketsGlobal implements the bbpPairings global matching architecture
// (dutch.cpp lines 936-1648). Instead of processing brackets independently
// and passing floaters down, it maintains a SINGLE global Blossom graph and
// processes score groups incrementally:
//
//  1. Bootstrap with the first (highest) score group.
//  2. Main loop: append the next score group, compute edges, run 7 phases
//     on the global graph.
//  3. After each iteration: commit current-bracket matches, carry forward
//     unmatched players + next-score-group players.
//
// This mirrors bbpPairings exactly: the matching graph always contains
// [downfloaters from previous iterations] + [current score group] + [next score group].
// Only current-bracket player matches are committed; next-score-group matches
// serve as "lookahead" for optimization but are not finalized.
func pairBracketsGlobal(
	scoreGroups []swisslib.ScoreGroup,
	ctx *swisslib.CriteriaContext,
	playerMap map[string]*swisslib.PlayerState,
) ([]swisslib.ProposedPairing, []string) {
	if len(scoreGroups) == 0 {
		return nil, nil
	}

	var notes []string
	var allCommitted []swisslib.ProposedPairing

	// Precompute edge weight parameters (mirrors bbpPairings' computeMatching
	// setup at lines 685-715).
	ewParams := swisslib.ComputeEdgeWeightParams(scoreGroups, ctx.CurrentRound-1)
	sgSizeBits := ewParams.ScoreGroupSizeBits

	// =====================================================================
	// Stage 0: Build the GLOBAL player list from ALL score groups.
	// bbpPairings creates matchingComputer with ALL players upfront
	// (lines 753-930). We do the same.
	// =====================================================================
	var allPlayers []*swisslib.PlayerState
	// sgBoundaries[i] = start index of score group i in allPlayers.
	sgBoundaries := make([]int, len(scoreGroups)+1)
	for si, sg := range scoreGroups {
		sgBoundaries[si] = len(allPlayers)
		allPlayers = append(allPlayers, sg.Players...)
	}
	sgBoundaries[len(scoreGroups)] = len(allPlayers)
	totalN := len(allPlayers)

	if totalN < 2 {
		return nil, nil
	}

	// Global base weights: baseWeight[i][j] for all i < j in allPlayers.
	// Indexed as map[(i,j)] where i < j, to avoid O(n^2) memory for slices.
	// Uses *big.Int because bbpPairings edge weights exceed 64 bits.
	bigOne := big.NewInt(1)
	globalBase := make(map[[2]int]*big.Int, totalN*totalN/4)

	// =====================================================================
	// Stage 1: Pre-populate ALL edges with ComputeBaseEdgeWeight(false, false).
	// This mirrors bbpPairings lines 766-827 where it sets edge weights
	// for ALL pairs before the bracket loop starts.
	// =====================================================================
	for i := 0; i < totalN; i++ {
		if i%100 == 0 && ctx.DeadlineExceeded() {
			break
		}
		for j := i + 1; j < totalN; j++ {
			pi, pj := allPlayers[i], allPlayers[j]
			if swisslib.HasPlayed(pi, pj) {
				continue
			}
			// Use lowest score as bracketScore for C3 check during init.
			bs := pj.Score
			if pi.Score < bs {
				bs = pi.Score
			}
			if !swisslib.C3AbsoluteColorConflict(&swisslib.ProposedPairing{
				White: pi, Black: pj, BracketScore: bs,
			}, ctx) {
				continue
			}
			w := swisslib.ComputeBaseEdgeWeight(pi, pj, false, false, &ewParams)
			if w.Sign() == 0 {
				w = new(big.Int).Set(bigOne)
			}
			globalBase[edgeKey(i, j)] = w
		}
	}

	// Global mutable edge weight map. Initialized from globalBase.
	globalEdgeW := make(map[[2]int]*big.Int, len(globalBase))
	for k, w := range globalBase {
		globalEdgeW[k] = new(big.Int).Set(w)
	}

	// Global finalized flags.
	globalFinalized := make([]bool, totalN)

	// Track committed player IDs.
	committed := make(map[string]bool)

	// Helper: build edges for Blossom from the global edge map.
	// Unlike the previous approach that filtered to "active vertices", we now
	// include ALL non-finalized vertices — matching bbpPairings' architecture
	// where the matching computer always sees all players. Future-bracket
	// players have lower weights from Stage 1, so Blossom prefers current-bracket
	// matches naturally.
	buildGlobalEdges := func() []blossom.BigEdge {
		edges := make([]blossom.BigEdge, 0, len(globalEdgeW))
		for k, w := range globalEdgeW {
			if w != nil && w.Sign() > 0 && !globalFinalized[k[0]] && !globalFinalized[k[1]] {
				edges = append(edges, blossom.BigEdge{
					I: k[0], J: k[1], Weight: new(big.Int).Set(w),
				})
			}
		}
		return edges
	}

	runGlobalBlossom := func() []int {
		edges := buildGlobalEdges()
		if len(edges) == 0 {
			return nil
		}
		return blossom.MaxWeightMatchingBig(edges, true)
	}

	finalizePairGlobal := func(a, b int) {
		globalFinalized[a] = true
		globalFinalized[b] = true
		// Mirror bbpPairings finalizePair: zero all edges from a and b,
		// EXCEPT keep the mutual edge at weight 1 so Blossom continues
		// to match them in subsequent computeMatching() calls.
		pairKey := edgeKey(a, b)
		for k := range globalEdgeW {
			if k[0] == a || k[1] == a || k[0] == b || k[1] == b {
				if k == pairKey {
					globalEdgeW[k] = new(big.Int).Set(bigOne)
				} else {
					globalEdgeW[k] = new(big.Int)
				}
			}
		}
	}

	// =====================================================================
	// Stage 2: Bracket loop.
	// playersByIndex is a LOCAL view: downfloaters + current SG + next SG.
	// vertexIdx maps local index → global index in allPlayers.
	// =====================================================================
	var playersByIndex []*swisslib.PlayerState
	var vertexIdx []int
	scoreGroupBegin := 0
	sgIter := 0

	// Bootstrap: seed with the first score group.
	for gi := sgBoundaries[0]; gi < sgBoundaries[1]; gi++ {
		playersByIndex = append(playersByIndex, allPlayers[gi])
		vertexIdx = append(vertexIdx, gi)
	}
	sgIter = 1
	maxIter := 2*len(scoreGroups) + 2 // safety limit

	for iter := 0; (len(playersByIndex) > 1 || sgIter < len(scoreGroups)) && iter < maxIter; iter++ {
		if ctx.DeadlineExceeded() {
			break
		}

		nextScoreGroupBegin := len(playersByIndex)

		// Append the next score group's players.
		if sgIter < len(scoreGroups) {
			for gi := sgBoundaries[sgIter]; gi < sgBoundaries[sgIter+1]; gi++ {
				if !committed[allPlayers[gi].ID] {
					playersByIndex = append(playersByIndex, allPlayers[gi])
					vertexIdx = append(vertexIdx, gi)
				}
			}
			sgIter++
		}

		n := len(playersByIndex)
		if n < 2 {
			break
		}

		// Bracket score for the current bracket.
		var bracketScore float64
		if scoreGroupBegin < nextScoreGroupBegin && scoreGroupBegin < n {
			bracketScore = playersByIndex[scoreGroupBegin].Score
		} else if n > 0 {
			bracketScore = playersByIndex[0].Score
		}

		dfCount := scoreGroupBegin

		// =====================================================================
		// Update global edges for current bracket + next SG.
		// Mirrors bbpPairings computeBaseEdgeWeights: only recompute edges
		// where largerPlayerIndex >= scoreGroupBegin (at least one player
		// is from current bracket or next SG). This UPDATES the global
		// edge map — other edges remain at their Stage 1 values.
		// =====================================================================
		// Also rebuild a local baseWeight view for Phases 2-7.
		baseWeight := make([][]*big.Int, n)
		for i := 0; i < n; i++ {
			baseWeight[i] = make([]*big.Int, n)
		}

		for li := 0; li < n; li++ {
			if li%100 == 0 && ctx.DeadlineExceeded() {
				break
			}
			for lj := li + 1; lj < n; lj++ {
				// Only update where larger local index >= scoreGroupBegin.
				if lj < scoreGroupBegin {
					continue
				}

				gi, gj := vertexIdx[li], vertexIdx[lj]
				pi, pj := playersByIndex[li], playersByIndex[lj]

				if swisslib.HasPlayed(pi, pj) {
					continue
				}
				if !swisslib.C3AbsoluteColorConflict(&swisslib.ProposedPairing{
					White: pi, Black: pj, BracketScore: bracketScore,
				}, ctx) {
					continue
				}

				inCurrentBracket := lj < nextScoreGroupBegin
				inNextBracket := lj >= nextScoreGroupBegin

				w := swisslib.ComputeBaseEdgeWeight(pi, pj, inCurrentBracket, inNextBracket, &ewParams)
				if w.Sign() == 0 {
					w = new(big.Int).Set(bigOne)
				}

				baseWeight[li][lj] = w
				baseWeight[lj][li] = new(big.Int).Set(w)

				// Update global edge map.
				key := edgeKey(gi, gj)
				globalBase[key] = new(big.Int).Set(w)
				globalEdgeW[key] = new(big.Int).Set(w)
			}
		}

		// Also populate baseWeight for downfloater-downfloater pairs
		// (local indices below scoreGroupBegin) from globalBase.
		for li := 0; li < scoreGroupBegin; li++ {
			for lj := li + 1; lj < scoreGroupBegin; lj++ {
				gi, gj := vertexIdx[li], vertexIdx[lj]
				key := edgeKey(gi, gj)
				if w, ok := globalBase[key]; ok && w != nil && w.Sign() > 0 {
					baseWeight[li][lj] = new(big.Int).Set(w)
					baseWeight[lj][li] = new(big.Int).Set(w)
				}
			}
		}

		// edgeWeightComputer (mirrors bbpPairings lines 1055-1085).
		// Operates on LOCAL indices but writes to GLOBAL edgeW via vertexIdx.
		edgeWeightComputer := func(smallerLI, largerLI, smallerRemIdx, remPairs int) *big.Int {
			bw := baseWeight[largerLI][smallerLI]
			if bw == nil || bw.Sign() == 0 {
				return new(big.Int)
			}

			// Build addend matching bbpPairings' edgeWeightComputer.
			addend := new(big.Int)

			// Minimize exchanges: 1 if this is an S1 player (natural pairing).
			if smallerRemIdx < remPairs {
				addend.SetInt64(1)
			}

			// Minimize BSN distance: shift left by 2*sgSizeBits, subtract index.
			addend.Lsh(addend, uint(max(sgSizeBits, 0))) //nolint:gosec // sgSizeBits is bounded by tournament size
			addend.Lsh(addend, uint(max(sgSizeBits, 0))) //nolint:gosec // sgSizeBits is bounded by tournament size
			addend.Sub(addend, big.NewInt(int64(smallerRemIdx)))

			// Reserve 1 bit for exchange optimization.
			addend.Lsh(addend, 1)

			return new(big.Int).Add(bw, addend)
		}

		// Local helpers that operate on local indices but use global edgeW.
		setEdge := func(li, lj int, w *big.Int) {
			gi, gj := vertexIdx[li], vertexIdx[lj]
			globalEdgeW[edgeKey(gi, gj)] = w
		}
		getEdge := func(li, lj int) *big.Int {
			gi, gj := vertexIdx[li], vertexIdx[lj]
			w := globalEdgeW[edgeKey(gi, gj)]
			if w == nil {
				return new(big.Int)
			}
			return w
		}
		_ = getEdge // may not be used in all paths

		localFinalized := make([]bool, n)
		matchedPhase2 := make([]bool, n)

		finalizePairLocal := func(a, b int) {
			localFinalized[a] = true
			localFinalized[b] = true
			finalizePairGlobal(vertexIdx[a], vertexIdx[b])
		}

		// =====================================================================
		// Phase 1: Initial matching (global Blossom).
		// =====================================================================
		mate := runGlobalBlossom()

		// Translate global matching to local mate array.
		// mate[globalIdx] = matched globalIdx. We need localMate[localIdx].
		globalToLocal := make(map[int]int, n)
		for li, gi := range vertexIdx {
			globalToLocal[gi] = li
		}

		toLocalMate := func(globalMate []int) []int {
			localMate := make([]int, n)
			for i := range localMate {
				localMate[i] = -1
			}
			for li, gi := range vertexIdx {
				if gi < len(globalMate) && globalMate[gi] >= 0 {
					if mli, ok := globalToLocal[globalMate[gi]]; ok {
						localMate[li] = mli
					}
					// If mate is outside our local view, leave as -1.
				}
			}
			return localMate
		}

		localMate := toLocalMate(mate)

		// =====================================================================
		// Phase 2: Choose moved-down players (heterogeneous only).
		// =====================================================================
		var iterPairs []swisslib.ProposedPairing
		isHeterogeneous := dfCount > 0

		if isHeterogeneous && len(mate) > 0 {
			type dfScoreGroup struct {
				startIdx int
				score    float64
			}
			var dfGroups []dfScoreGroup
			for i := 0; i < dfCount; i++ {
				if len(dfGroups) == 0 || playersByIndex[i].Score != dfGroups[len(dfGroups)-1].score {
					dfGroups = append(dfGroups, dfScoreGroup{startIdx: i, score: playersByIndex[i].Score})
				}
			}

			for _, dfg := range dfGroups {
				dfEnd := dfCount
				for k := dfg.startIdx + 1; k < dfCount; k++ {
					if playersByIndex[k].Score != dfg.score {
						dfEnd = k
						break
					}
				}

				remainingDF := 0
				remainingMatchedDF := 0
				for k := dfg.startIdx; k < dfEnd; k++ {
					remainingDF++
					if localMate[k] >= dfCount && localMate[k] < nextScoreGroupBegin {
						remainingMatchedDF++
					}
				}

				for i := dfg.startIdx; i < dfEnd; i++ {
					if remainingMatchedDF == 0 {
						continue
					}
					if remainingDF <= remainingMatchedDF {
						// Auto-accept: all remaining DFs can be matched.
						// bbpPairings does NOT decrement either counter here.
						matchedPhase2[i] = true
						continue
					}
					remainingDF--

					if localMate[i] < dfCount || localMate[i] >= nextScoreGroupBegin {
						for j := dfCount; j < nextScoreGroupBegin; j++ {
							bw := baseWeight[j][i]
							if bw != nil && bw.Sign() > 0 {
								setEdge(i, j, new(big.Int).Or(bw, bigOne))
							}
						}
						mate = runGlobalBlossom()
						localMate = toLocalMate(mate)
					}

					if localMate[i] >= dfCount && localMate[i] < nextScoreGroupBegin {
						matchedPhase2[i] = true
						remainingMatchedDF--
						sgSize := nextScoreGroupBegin - dfCount
						stickyVal := big.NewInt(int64(sgSize + 1))
						for j := dfCount; j < nextScoreGroupBegin; j++ {
							bw := baseWeight[j][i]
							if bw != nil && bw.Sign() > 0 {
								setEdge(i, j, new(big.Int).Or(bw, stickyVal))
							}
						}
					}
				}
			}

			// Sub-phase 2b: Choose opponents for each matched downfloater.
			for i := 0; i < dfCount; i++ {
				if !matchedPhase2[i] {
					continue
				}
				addend := big.NewInt(int64(n))
				for j := nextScoreGroupBegin - 1; j >= dfCount; j-- {
					if matchedPhase2[j] {
						continue
					}
					bw := baseWeight[j][i]
					if bw != nil && bw.Sign() > 0 {
						setEdge(i, j, new(big.Int).Add(bw, addend))
						addend.Add(addend, bigOne)
					}
				}
				mate = runGlobalBlossom()
				localMate = toLocalMate(mate)
				if localMate[i] >= dfCount && localMate[i] < nextScoreGroupBegin {
					partner := localMate[i]
					matchedPhase2[partner] = true
					iterPairs = append(iterPairs, swisslib.ProposedPairing{
						White:        playersByIndex[i],
						Black:        playersByIndex[partner],
						BracketScore: bracketScore,
					})
					finalizePairLocal(i, partner)
				}
			}
		}

		// =====================================================================
		// Phase 3: Initialize remainder.
		// Mirrors bbpPairings lines 1270-1285: skip SG players whose Blossom
		// mate is a downfloater (index < scoreGroupBegin/dfCount), even if
		// that downfloater wasn't selected by Phase 2.
		// =====================================================================
		remainder := make([]int, 0, n)
		for i := dfCount; i < nextScoreGroupBegin; i++ {
			if localFinalized[i] {
				continue
			}
			// bbpPairings: skip if stableMatching < scoreGroupBeginVertex
			if localMate[i] >= 0 && localMate[i] < dfCount {
				continue
			}
			remainder = append(remainder, i)
		}

		// Count remainderPairs from latest matching.
		// bbpPairings lines 1281-1284: count players whose mate has lower vertex.
		remainderPairs := 0
		remIndexOf := make(map[int]int, len(remainder))
		for ri, li := range remainder {
			remIndexOf[li] = ri
		}
		for _, li := range remainder {
			if localMate[li] >= 0 && localMate[li] < li {
				if _, inRem := remIndexOf[localMate[li]]; inRem {
					remainderPairs++
				}
			}
		}
		if remainderPairs == 0 {
			var remS1 int
			if dfCount > 0 {
				remS1 = (nextScoreGroupBegin - dfCount) / 2
			} else {
				remS1 = len(remainder) / 2
			}
			remainderPairs = remS1
			if remainderPairs > len(remainder)/2 {
				remainderPairs = len(remainder) / 2
			}
		}

		// Apply edgeWeightComputer to remainder edges.
		remSet := make(map[int]bool, len(remainder))
		for _, li := range remainder {
			remSet[li] = true
		}
		// Zero out remainder×remainder edges in global map.
		for _, li := range remainder {
			for _, lj := range remainder {
				if li < lj {
					setEdge(li, lj, new(big.Int))
				}
			}
		}
		for ri, li := range remainder {
			for rj := ri + 1; rj < len(remainder); rj++ {
				lj := remainder[rj]
				w := edgeWeightComputer(li, lj, ri, remainderPairs)
				if w != nil && w.Sign() > 0 {
					setEdge(li, lj, w)
				}
			}
		}
		mate = runGlobalBlossom()
		localMate = toLocalMate(mate)

		// =====================================================================
		// Phase 4: Exchange selection Phase 1.
		// =====================================================================
		exchangeCount := 0
		for ri := 0; ri < remainderPairs; ri++ {
			li := remainder[ri]
			matchedToS2 := false
			if localMate[li] >= 0 && localMate[li] < n {
				mateRI, inRem := remIndexOf[localMate[li]]
				if inRem && mateRI > ri {
					matchedToS2 = true
				}
			}
			if !matchedToS2 {
				exchangeCount++
			}
		}

		if exchangeCount > 0 && remainderPairs > 0 {
			exchangesRemaining := exchangeCount
			for ri := remainderPairs - 1; ri >= 0 && exchangesRemaining > 0; ri-- {
				if ctx.DeadlineExceeded() {
					break
				}
				li := remainder[ri]
				isMatchedS1S2 := false
				if localMate[li] >= 0 && localMate[li] < n {
					mateRI, inRem := remIndexOf[localMate[li]]
					if inRem && mateRI > ri {
						isMatchedS1S2 = true
					}
				}
				if isMatchedS1S2 {
					for rj := ri + 1; rj < len(remainder); rj++ {
						lj := remainder[rj]
						w := edgeWeightComputer(li, lj, ri, remainderPairs)
						if w != nil && w.Sign() > 0 {
							w = new(big.Int).Sub(w, bigOne)
							setEdge(li, lj, w)
						}
					}
					mate = runGlobalBlossom()
					localMate = toLocalMate(mate)
				}
				exchange := true
				if localMate[li] >= 0 && localMate[li] < n {
					mateRI, inRem := remIndexOf[localMate[li]]
					if inRem && mateRI > ri {
						exchange = false
					}
				}
				if exchange {
					exchangesRemaining--
				}
				for rj := ri + 1; rj < len(remainder); rj++ {
					lj := remainder[rj]
					if exchange {
						baseWeight[lj][li] = nil
						baseWeight[li][lj] = nil
					}
					w := edgeWeightComputer(li, lj, ri, remainderPairs)
					setEdge(li, lj, w)
				}
			}
		}

		// =====================================================================
		// Phase 5: Exchange selection Phase 2.
		// =====================================================================
		if exchangeCount > 0 {
			exchangesRemaining := exchangeCount
			for ri := remainderPairs; ri < len(remainder) && exchangesRemaining > 1; ri++ {
				if ctx.DeadlineExceeded() {
					break
				}
				li := remainder[ri]
				if li >= nextScoreGroupBegin {
					continue
				}

				alreadyExchanged := false
				if localMate[li] >= 0 && localMate[li] < n {
					mateRI, inRem := remIndexOf[localMate[li]]
					if inRem && mateRI > ri {
						alreadyExchanged = true
					}
				}
				if !alreadyExchanged {
					for rj := ri + 1; rj < len(remainder); rj++ {
						lj := remainder[rj]
						w := edgeWeightComputer(li, lj, ri, remainderPairs)
						if w != nil && w.Sign() > 0 {
							w = new(big.Int).Add(w, bigOne)
							setEdge(li, lj, w)
						}
					}
					mate = runGlobalBlossom()
					localMate = toLocalMate(mate)
				}
				exchanged := false
				if localMate[li] >= 0 && localMate[li] < n {
					mateRI, inRem := remIndexOf[localMate[li]]
					if inRem && mateRI > ri {
						exchanged = true
					}
				}
				if exchanged {
					exchangesRemaining--
					for rj := 0; rj < ri; rj++ {
						lj := remainder[rj]
						baseWeight[li][lj] = nil
						baseWeight[lj][li] = nil
						setEdge(lj, li, new(big.Int))
					}
				}
				if !alreadyExchanged {
					for rj := ri + 1; rj < len(remainder); rj++ {
						lj := remainder[rj]
						w := edgeWeightComputer(li, lj, ri, remainderPairs)
						setEdge(li, lj, w)
					}
				}
			}
		}

		// =====================================================================
		// Phase 6: Finalize exchanges.
		// =====================================================================
		for ri := 0; ri < len(remainder); ri++ {
			li := remainder[ri]
			for rj := ri + 1; rj < len(remainder); rj++ {
				lj := remainder[rj]

				iNotMatchedDown := localMate[li] <= li || localMate[li] >= nextScoreGroupBegin
				if !iNotMatchedDown {
					mateRI, inRem := remIndexOf[localMate[li]]
					if !inRem || mateRI <= ri {
						iNotMatchedDown = true
					}
				}

				jHasNaturalPair := false
				if localMate[lj] > lj && localMate[lj] < n {
					mateRI, inRem := remIndexOf[localMate[lj]]
					if inRem && mateRI > rj && localMate[lj] < nextScoreGroupBegin {
						jHasNaturalPair = true
					}
				}

				if iNotMatchedDown || jHasNaturalPair {
					baseWeight[lj][li] = nil
					baseWeight[li][lj] = nil
				}
				bwVal := baseWeight[lj][li]
				if bwVal == nil {
					bwVal = new(big.Int)
				}
				setEdge(li, lj, bwVal)
			}
		}

		// =====================================================================
		// Phase 7: Choose opponents (within remainder only).
		// =====================================================================
		// Re-run Blossom after Phase 6 modifications.
		mate = runGlobalBlossom()
		localMate = toLocalMate(mate)

		for ri := 0; ri < len(remainder); ri++ {
			li := remainder[ri]
			if ctx.DeadlineExceeded() {
				break
			}

			if localFinalized[li] {
				continue
			}

			if localMate[li] < 0 {
				continue
			}
			mateRI, inRem := remIndexOf[localMate[li]]
			if !inRem || mateRI <= ri {
				continue
			}
			// Only match within current SG (remainder pairs).
			if localMate[li] >= nextScoreGroupBegin {
				continue
			}

			addend := new(big.Int)
			for rj := len(remainder) - 1; rj > ri; rj-- {
				lj := remainder[rj]
				if localFinalized[lj] {
					continue
				}
				bw := baseWeight[lj][li]
				if bw == nil || bw.Sign() == 0 {
					continue
				}
				setEdge(li, lj, new(big.Int).Add(bw, addend))
				addend.Add(addend, bigOne)
			}
			mate = runGlobalBlossom()
			localMate = toLocalMate(mate)

			p7mate := localMate[li]
			if p7mate >= 0 && p7mate < nextScoreGroupBegin && !localFinalized[p7mate] {
				iterPairs = append(iterPairs, swisslib.ProposedPairing{
					White:        playersByIndex[li],
					Black:        playersByIndex[p7mate],
					BracketScore: bracketScore,
				})
				finalizePairLocal(li, p7mate)
			}
		}

		// =====================================================================
		// Carry forward: commit finalized players, keep unmatched.
		// Mirrors bbpPairings lines 1601-1648. Players matched within the
		// current bracket are saved; all others become downfloaters in the
		// next iteration (including those matched to next-SG players by the
		// Blossom but not committed by Phase 7).
		// =====================================================================
		var newPlayersByIndex []*swisslib.PlayerState
		var newVertexIdx []int
		newScoreGroupBegin := 0

		for i := 0; i < n; i++ {
			p := playersByIndex[i]
			if localFinalized[i] {
				committed[p.ID] = true
			} else {
				if i < nextScoreGroupBegin {
					newScoreGroupBegin++
				}
				newPlayersByIndex = append(newPlayersByIndex, p)
				newVertexIdx = append(newVertexIdx, vertexIdx[i])
			}
		}

		// Record float directions.
		for _, pair := range iterPairs {
			for _, p := range []*swisslib.PlayerState{pair.White, pair.Black} {
				if mp, ok := playerMap[p.ID]; ok {
					switch {
					case p.Score > bracketScore+0.001:
						mp.FloatHistory = append(mp.FloatHistory, swisslib.FloatDown)
					case p.Score < bracketScore-0.001:
						mp.FloatHistory = append(mp.FloatHistory, swisslib.FloatUp)
					default:
						mp.FloatHistory = append(mp.FloatHistory, swisslib.FloatNone)
					}
				}
			}
		}
		allCommitted = append(allCommitted, iterPairs...)

		playersByIndex = newPlayersByIndex
		vertexIdx = newVertexIdx
		scoreGroupBegin = newScoreGroupBegin

		if len(playersByIndex) <= 1 && sgIter >= len(scoreGroups) {
			break
		}
		// Safety: prevent infinite loops. bbpPairings loops until
		// playersByIndex <= 1. In the worst case, each SG is visited once
		// as current bracket, and then once more as downfloaters are
		// re-processed. Max iterations = 2 * numScoreGroups + some margin.
	}

	return allCommitted, notes
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
