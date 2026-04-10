// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"sort"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// matchContext holds the state needed for Dubov bracket matching.
type matchContext struct {
	ratings         map[string]int
	isRound1        bool
	forbidden       map[[2]string]bool
	completedRounds int
	playerMap       map[string]*swisslib.PlayerState
}

// bracketResult holds the result of matching a single bracket.
type bracketResult struct {
	pairs    []proposedPairing
	floaters []*swisslib.PlayerState
	score    DubovCandidateScore
}

// SplitG1G2 splits a score group into G1 (White-preference) and G2 (Black-preference/neutral).
//
// Round 1 (isRound1=true): G1 = first floor(n/2) players by TPN, G2 = rest.
// Later rounds: G1 = players preferring White, G2 = players preferring Black or no preference.
func SplitG1G2(players []*swisslib.PlayerState, isRound1 bool) (g1, g2 []*swisslib.PlayerState) {
	if isRound1 {
		// Round 1: first half by TPN.
		half := len(players) / 2
		g1 = make([]*swisslib.PlayerState, half)
		copy(g1, players[:half])
		g2 = make([]*swisslib.PlayerState, len(players)-half)
		copy(g2, players[half:])
		return g1, g2
	}

	// Later rounds: split by colour preference.
	for _, p := range players {
		pref := swisslib.ComputeColorPreference(p.ColorHistory)
		if pref.Color != nil && *pref.Color == swisslib.ColorWhite {
			g1 = append(g1, p)
		} else {
			g2 = append(g2, p)
		}
	}

	return g1, g2
}

// BalanceG1G2 ensures |G1| = floor(total/2) and |G2| = ceil(total/2)
// by shifting players between groups per Art. 3.2.4.
// Shifted players are chosen per Art. 4.3 (shifter sorting).
func BalanceG1G2(g1, g2 []*swisslib.PlayerState, ratings map[string]int) ([]*swisslib.PlayerState, []*swisslib.PlayerState) {
	total := len(g1) + len(g2)
	targetG1 := total / 2
	targetG2 := total - targetG1

	if len(g1) == targetG1 && len(g2) == targetG2 {
		return g1, g2
	}

	if len(g1) > targetG1 {
		// Move excess from G1 to G2.
		excess := len(g1) - targetG1
		sortShifterCandidates(g1, ratings, swisslib.ColorWhite)
		// Move last `excess` players (worst shift candidates) to G2.
		shifted := make([]*swisslib.PlayerState, excess)
		copy(shifted, g1[len(g1)-excess:])
		g1 = g1[:len(g1)-excess]
		g2 = append(shifted, g2...)
		// Re-sort G2 by TPN ascending.
		sort.SliceStable(g2, func(i, j int) bool {
			return g2[i].TPN < g2[j].TPN
		})
	} else if len(g2) > targetG2 {
		// Move excess from G2 to G1.
		excess := len(g2) - targetG2
		sortShifterCandidates(g2, ratings, swisslib.ColorBlack)
		shifted := make([]*swisslib.PlayerState, excess)
		copy(shifted, g2[len(g2)-excess:])
		g2 = g2[:len(g2)-excess]
		g1 = append(g1, shifted...)
	}

	return g1, g2
}

// sortShifterCandidates sorts players for shifter selection per Art. 4.3.
// White seekers: ascending ARO, then ascending TPN.
// Black seekers: ascending TPN.
func sortShifterCandidates(players []*swisslib.PlayerState, ratings map[string]int, seekColor swisslib.Color) {
	if seekColor == swisslib.ColorWhite {
		aros := make(map[string]float64, len(players))
		for _, p := range players {
			aros[p.ID] = ComputeARO(p, ratings)
		}
		sort.SliceStable(players, func(i, j int) bool {
			ai, aj := aros[players[i].ID], aros[players[j].ID]
			if ai != aj {
				return ai < aj
			}
			return players[i].TPN < players[j].TPN
		})
	} else {
		sort.SliceStable(players, func(i, j int) bool {
			return players[i].TPN < players[j].TPN
		})
	}
}

// MatchBracketDubov attempts to pair all players in a bracket using the Dubov algorithm.
//
// Algorithm:
//  1. Split into G1 (white-seekers) and G2 (black-seekers/neutral)
//  2. Balance G1/G2 so |G1| = floor(n/2)
//  3. Sort G1 by ascending ARO (Art. 3.2.5)
//  4. Try sequential pairing with each G2 transposition
//  5. Pick the best candidate per criteria scoring
//  6. Unmatched players become floaters
func MatchBracketDubov(bracket swisslib.Bracket, ctx *matchContext) (*bracketResult, error) {
	players := bracket.Players
	if len(players) < 2 {
		// Single player -> becomes floater.
		return &bracketResult{floaters: players}, nil
	}

	// Step 1: Split into G1 and G2.
	g1, g2 := SplitG1G2(players, ctx.isRound1)

	// Step 2: Balance the groups.
	g1, g2 = BalanceG1G2(g1, g2, ctx.ratings)

	// Step 3: Sort G1 by ascending ARO (Art. 3.2.5).
	if !ctx.isRound1 {
		SortByAROAscending(g1, ctx.ratings)
	}

	// Step 4: Generate G2 transpositions and try each.
	maxT := MaxT(ctx.completedRounds)
	var bestResult *bracketResult
	var bestScore *DubovCandidateScore

	transpositions := generateDubovTranspositions(g2, 120)

	for idx, g2perm := range transpositions {
		candidate := tryPairing(g1, g2perm, ctx, bracket.OriginalScore, maxT)
		if candidate == nil {
			continue // absolute criteria violated
		}

		candidate.score.TranspositionIdx = idx

		if bestScore == nil || candidate.score.Compare(bestScore) < 0 {
			bestResult = candidate
			bestScore = &candidate.score
		}

		if bestScore.IsPerfect() {
			break // can't do better
		}
	}

	if bestResult == nil {
		// No valid pairing found - all players float.
		return &bracketResult{floaters: players}, nil
	}

	return bestResult, nil
}

// tryPairing attempts to pair G1[i] with G2[i] sequentially.
// Returns nil if any absolute criterion is violated.
func tryPairing(g1, g2 []*swisslib.PlayerState, ctx *matchContext, bracketScore float64, maxT int) *bracketResult {
	n := len(g1)
	if n > len(g2) {
		n = len(g2)
	}

	pairs := make([]proposedPairing, 0, n)

	for i := 0; i < n; i++ {
		a, b := g1[i], g2[i]

		// Check absolute criteria (C1, C3, forbidden).
		if !SatisfiesAbsoluteCriteria(a, b, ctx.forbidden) {
			return nil // absolute violation -> reject entire transposition
		}

		pairs = append(pairs, proposedPairing{
			white:        a,
			black:        b,
			bracketScore: bracketScore,
		})
	}

	// Collect floaters: any remaining players in the larger group.
	var floaters []*swisslib.PlayerState
	if len(g1) > n {
		floaters = append(floaters, g1[n:]...)
	}
	if len(g2) > n {
		floaters = append(floaters, g2[n:]...)
	}

	// Build floater ID set for criteria evaluation.
	floaterIDs := make(map[string]bool, len(floaters))
	for _, f := range floaters {
		floaterIDs[f.ID] = true
	}

	// Score the candidate.
	score := DubovCandidateScore{
		UpfloaterCount:    CriterionC4(floaters),
		UpfloaterScoreSum: CriterionC5(floaters),
	}
	score.Violations[IdxC4] = score.UpfloaterCount
	score.Violations[IdxC5] = 0 // C5 handled separately (score sum, not violation count)
	score.Violations[IdxC6] = CriterionC6(pairs)
	score.Violations[IdxC7] = CriterionC7(floaters, maxT)
	score.Violations[IdxC8] = CriterionC8(floaters)
	score.Violations[IdxC9] = CriterionC9(pairs, floaterIDs, maxT)
	score.Violations[IdxC10] = CriterionC10(floaters, maxT)

	return &bracketResult{
		pairs:    pairs,
		floaters: floaters,
		score:    score,
	}
}

// generateDubovTranspositions generates permutations of G2 for transposition-based
// matching. Per Art. 3.2.6/4.4: G2 permutations sorted by ascending TPN sequence.
// Uses Narayana Pandita's algorithm, capped at maxCount.
func generateDubovTranspositions(g2 []*swisslib.PlayerState, maxCount int) [][]*swisslib.PlayerState {
	if len(g2) == 0 {
		return nil
	}

	results := make([][]*swisslib.PlayerState, 0, maxCount)

	// Copy initial permutation (identity - already sorted by TPN ascending).
	current := make([]*swisslib.PlayerState, len(g2))
	copy(current, g2)

	first := make([]*swisslib.PlayerState, len(g2))
	copy(first, current)
	results = append(results, first)

	for len(results) < maxCount {
		next := nextPermutation(current)
		if next == nil {
			break
		}
		perm := make([]*swisslib.PlayerState, len(next))
		copy(perm, next)
		results = append(results, perm)
		current = next
	}

	return results
}

// nextPermutation generates the next lexicographic permutation by TPN.
// Returns nil if no more permutations exist.
func nextPermutation(players []*swisslib.PlayerState) []*swisslib.PlayerState {
	n := len(players)
	if n < 2 {
		return nil
	}

	result := make([]*swisslib.PlayerState, n)
	copy(result, players)

	// Narayana Pandita's algorithm:
	// 1. Find largest i such that result[i].TPN < result[i+1].TPN.
	i := n - 2
	for i >= 0 && result[i].TPN >= result[i+1].TPN {
		i--
	}
	if i < 0 {
		return nil // last permutation
	}

	// 2. Find largest j such that result[i].TPN < result[j].TPN.
	j := n - 1
	for result[j].TPN <= result[i].TPN {
		j--
	}

	// 3. Swap result[i] and result[j].
	result[i], result[j] = result[j], result[i]

	// 4. Reverse result[i+1:].
	for left, right := i+1, n-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}

	return result
}
