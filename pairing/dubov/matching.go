// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"context"
	"sort"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// Enumeration safety limits. They bound Articles 4.1.3/4.4 enumeration so a
// pathological bracket returns ErrBracketTooLarge instead of hanging. The
// limits are high enough that every worked example in the FIDE text resolves
// well below them; they are package variables so the guard tests can lower
// them (see the Burstein system for the same pattern).
var (
	maxSearchNodes  = 1_000_000
	maxCombinations = 200_000
)

type searchBudget struct{ nodes int }

func (b *searchBudget) consume() error {
	if b == nil {
		return nil
	}
	b.nodes++
	if b.nodes > maxSearchNodes {
		return ErrBracketTooLarge
	}
	return nil
}

type matchContext struct {
	ratings     map[string]int
	forbidden   map[[2]string]bool
	totalRounds int
	lastRound   bool
	ctx         context.Context
}

type bracketResult struct {
	pairs   []proposedPairing
	quality candidateQuality
}

// SplitG1G2 implements Article 3.2.3. A bracket whose players have all yet to
// play is split by fixed TPN, irrespective of its round number.
func SplitG1G2(players []*swisslib.PlayerState, _ bool) (g1, g2 []*swisslib.PlayerState) {
	allUnplayed := true
	for _, player := range players {
		if swisslib.GamesPlayed(player) != 0 {
			allUnplayed = false
			break
		}
	}
	if allUnplayed {
		ordered := append([]*swisslib.PlayerState(nil), players...)
		sort.SliceStable(ordered, func(i, j int) bool {
			return swisslib.EffectivePairingNumber(ordered[i]) < swisslib.EffectivePairingNumber(ordered[j])
		})
		half := len(ordered) / 2
		return ordered[:half], ordered[half:]
	}
	for _, player := range players {
		pref := preference(player)
		if pref.Color != nil && *pref.Color == swisslib.ColorWhite {
			g1 = append(g1, player)
		} else {
			g2 = append(g2, player)
		}
	}
	return g1, g2
}

// BalanceG1G2 applies the first recomposition found in Article 4.3 order.
// It is retained for tests that need a basic recomposition without checking
// whether it yields a legal transposition.
func BalanceG1G2(g1, g2 []*swisslib.PlayerState, ratings map[string]int) ([]*swisslib.PlayerState, []*swisslib.PlayerState) {
	var first *recomposition
	_ = forEachRecomposition(g1, g2, ratings, nil, func(r recomposition) (bool, error) {
		first = &r
		return true, nil
	})
	if first == nil {
		return g1, g2
	}
	return first.g1, first.g2
}

type recomposition struct{ g1, g2 []*swisslib.PlayerState }

// forEachRecomposition enumerates Article 3.2.4 recompositions. For
// s = 0, 1, 2, ... it shifts s players out of the smaller subgroup (G1 when
// the sizes are equal) in Article 4.3 order, then equalises the subgroup sizes.
// visit returns stop=true to end the walk early.
func forEachRecomposition(g1, g2 []*swisslib.PlayerState, ratings map[string]int, budget *searchBudget, visit func(recomposition) (bool, error)) error {
	smaller, larger := g1, g2
	smallerIsG1 := true
	if len(g1) > len(g2) {
		smaller, larger = g2, g1
		smallerIsG1 = false
	}
	for s := 0; s <= len(smaller); s++ {
		stop, err := forEachFirstStageSet(smaller, s, smallerIsG1, ratings, budget, func(shifted []*swisslib.PlayerState) (bool, error) {
			newSmaller := withoutAll(smaller, shifted)
			newLarger := append(append([]*swisslib.PlayerState(nil), larger...), shifted...)
			if smallerIsG1 {
				return equaliseRecompositions(newSmaller, newLarger, ratings, budget, visit)
			}
			return equaliseRecompositions(newLarger, newSmaller, ratings, budget, visit)
		})
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}
	return nil
}

// forEachFirstStageSet enumerates the Article 3.2.4 first-stage shift sets of
// size count from the smaller subgroup in Article 4.3 order.
func forEachFirstStageSet(smaller []*swisslib.PlayerState, count int, smallerIsG1 bool, ratings map[string]int, budget *searchBudget, visit func([]*swisslib.PlayerState) (bool, error)) (bool, error) {
	if count == 0 {
		return visit(nil)
	}
	return forEachCombination(shifterOrder(smaller, ratings, smallerIsG1), count, budget, visit)
}

// equaliseRecompositions enumerates the Article 3.2.4 size-equalisation stage:
// shift players from the larger subgroup into the smaller one in Article 4.3
// order until both subgroups have the same size.
func equaliseRecompositions(g1, g2 []*swisslib.PlayerState, ratings map[string]int, budget *searchBudget, visit func(recomposition) (bool, error)) (bool, error) {
	if len(g1) == len(g2) {
		if err := budget.consume(); err != nil {
			return false, err
		}
		return visit(recomposition{g1: g1, g2: g2})
	}
	fromG1 := len(g1) > len(g2)
	larger, smaller := g2, g1
	if fromG1 {
		larger, smaller = g1, g2
	}
	// One same-subgroup pair is avoided for every two-player excess.
	count := (len(larger) - len(smaller)) / 2
	// Only players still seeking the colour of the subgroup they leave are
	// possible shifters (4.3.1); first-stage shifters cannot shift back.
	ordered := shifterOrder(shiftersIn(larger, fromG1), ratings, fromG1)
	return forEachCombination(ordered, count, budget, func(set []*swisslib.PlayerState) (bool, error) {
		chosen := make(map[string]bool, len(set))
		for _, player := range set {
			chosen[player.ID] = true
		}
		remaining := make([]*swisslib.PlayerState, 0, len(larger)-len(set))
		for _, player := range larger {
			if !chosen[player.ID] {
				remaining = append(remaining, player)
			}
		}
		var rec recomposition
		if fromG1 {
			rec = recomposition{g1: remaining, g2: append(append([]*swisslib.PlayerState(nil), set...), smaller...)}
		} else {
			rec = recomposition{g1: append(append([]*swisslib.PlayerState(nil), set...), smaller...), g2: remaining}
		}
		if err := budget.consume(); err != nil {
			return false, err
		}
		return visit(rec)
	})
}

func shiftersIn(players []*swisslib.PlayerState, whiteSeekers bool) []*swisslib.PlayerState {
	color := swisslib.ColorBlack
	if whiteSeekers {
		color = swisslib.ColorWhite
	}
	result := make([]*swisslib.PlayerState, 0, len(players))
	for _, player := range players {
		pref := preference(player)
		if pref.Color != nil && *pref.Color == color {
			result = append(result, player)
		}
	}
	return result
}

// shifterOrder numbers candidates as Articles 4.3.2-4.3.3 prescribe:
// White seekers by ascending ARO then ascending TPN, Black seekers by
// ascending TPN; sequence numbers are then assigned middle first.
func shifterOrder(players []*swisslib.PlayerState, ratings map[string]int, whiteSeekers bool) []*swisslib.PlayerState {
	ordered := append([]*swisslib.PlayerState(nil), players...)
	if whiteSeekers {
		SortByAROAscending(ordered, ratings)
	} else {
		sort.SliceStable(ordered, func(i, j int) bool {
			return swisslib.EffectivePairingNumber(ordered[i]) < swisslib.EffectivePairingNumber(ordered[j])
		})
	}
	sequence := make([]*swisslib.PlayerState, 0, len(ordered))
	remaining := append([]*swisslib.PlayerState(nil), ordered...)
	for len(remaining) > 0 {
		index := (len(remaining) - 1) / 2 // higher-listed of two middle players
		sequence = append(sequence, remaining[index])
		remaining = append(remaining[:index], remaining[index+1:]...)
	}
	return sequence
}

// MatchBracketDubov finds the first legal Article 3.2.4 recomposition (the
// smallest shift count, in Article 4.3 order) and its first legal G2
// transposition (Article 3.2.6).
func MatchBracketDubov(bracket swisslib.Bracket, ctx *matchContext) (*bracketResult, error) {
	return matchBracket(bracket, ctx)
}

func matchBracket(bracket swisslib.Bracket, ctx *matchContext) (*bracketResult, error) {
	if len(bracket.Players) == 0 || len(bracket.Players)%2 != 0 {
		return nil, nil
	}
	g1, g2 := SplitG1G2(bracket.Players, false)
	var result *bracketResult
	budget := &searchBudget{}
	err := forEachRecomposition(g1, g2, ctx.ratings, budget, func(composition recomposition) (bool, error) {
		if ctx.ctx != nil {
			if err := ctx.ctx.Err(); err != nil {
				return false, err
			}
		}
		first, err := firstLegalTransposition(composition.g1, composition.g2, ctx, bracket.OriginalScore, budget)
		if err != nil {
			return false, err
		}
		if first == nil {
			return false, nil
		}
		// Article 3.2.4 requires the first legal set in Article 4.3 order.
		result = first
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func firstLegalTransposition(g1, g2 []*swisslib.PlayerState, ctx *matchContext, score float64, budget *searchBudget) (*bracketResult, error) {
	if len(g1) != len(g2) {
		return nil, nil
	}
	SortByAROAscending(g1, ctx.ratings)
	sort.SliceStable(g2, func(i, j int) bool {
		return swisslib.EffectivePairingNumber(g2[i]) < swisslib.EffectivePairingNumber(g2[j])
	})
	used := make([]bool, len(g2))
	t2 := make([]*swisslib.PlayerState, len(g2))
	var result *bracketResult
	var search func(int) (bool, error)
	search = func(position int) (bool, error) {
		if ctx.ctx != nil {
			if err := ctx.ctx.Err(); err != nil {
				return false, err
			}
		}
		if err := budget.consume(); err != nil {
			return false, err
		}
		if position == len(g2) {
			pairs := make([]proposedPairing, 0, len(g1))
			for i := range g1 {
				pairs = append(pairs, proposedPairing{white: g1[i], black: t2[i], bracketScore: score})
			}
			result = &bracketResult{pairs: pairs}
			return true, nil
		}
		for i := 0; i < len(g2); i++ {
			if used[i] {
				continue
			}
			candidate := g2[i]
			if !SatisfiesAbsoluteCriteria(g1[position], candidate, ctx.forbidden) {
				continue
			}
			used[i] = true
			t2[position] = candidate
			found, err := search(position + 1)
			if err != nil {
				return false, err
			}
			if found {
				return true, nil
			}
			used[i] = false
			t2[position] = nil
		}
		return false, nil
	}
	_, err := search(0)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// forEachTransposition calls visit for Article 4.4 transpositions in TPN
// lexicographic order, stopping when visit returns true. It is retained for
// tests; bracket pairing uses the bounded depth-first search instead.
func forEachTransposition(g2 []*swisslib.PlayerState, visit func([]*swisslib.PlayerState) (bool, error)) error {
	if len(g2) == 0 {
		return nil
	}
	current := append([]*swisslib.PlayerState(nil), g2...)
	visited := 0
	for {
		visited++
		if visited > maxCombinations {
			return ErrBracketTooLarge
		}
		stop, err := visit(current)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
		next := nextPermutation(current)
		if next == nil {
			return nil
		}
		current = next
	}
}

// generateDubovTranspositions remains available to tests. maxCount <= 0 means
// no cap. Pairing uses firstLegalTransposition's depth-first search instead.
func generateDubovTranspositions(g2 []*swisslib.PlayerState, maxCount int) [][]*swisslib.PlayerState {
	if len(g2) == 0 {
		return nil
	}
	var result [][]*swisslib.PlayerState
	_ = forEachTransposition(g2, func(permutation []*swisslib.PlayerState) (bool, error) {
		result = append(result, append([]*swisslib.PlayerState(nil), permutation...))
		return maxCount > 0 && len(result) == maxCount, nil
	})
	return result
}

func nextPermutation(players []*swisslib.PlayerState) []*swisslib.PlayerState {
	result := append([]*swisslib.PlayerState(nil), players...)
	i := len(result) - 2
	for i >= 0 && swisslib.EffectivePairingNumber(result[i]) >= swisslib.EffectivePairingNumber(result[i+1]) {
		i--
	}
	if i < 0 {
		return nil
	}
	j := len(result) - 1
	for swisslib.EffectivePairingNumber(result[j]) <= swisslib.EffectivePairingNumber(result[i]) {
		j--
	}
	result[i], result[j] = result[j], result[i]
	for left, right := i+1, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

// forEachCombination enumerates Article 4.1.3 sets from an already numbered
// pool and calls visit for each. stop reports that visit ended the walk; err
// reports a visit or enumeration failure. When budget is non-nil, every set
// consumes one shared bracket-search node. It returns ErrBracketTooLarge once
// either limit is exceeded.
func forEachCombination(players []*swisslib.PlayerState, count int, budget *searchBudget, visit func([]*swisslib.PlayerState) (bool, error)) (bool, error) {
	if count < 0 || count > len(players) {
		return false, nil
	}
	if count == 0 {
		if err := budget.consume(); err != nil {
			return false, err
		}
		return visit(nil)
	}
	selected := make([]*swisslib.PlayerState, 0, count)
	visited := 0
	var walk func(int) (bool, error)
	walk = func(start int) (bool, error) {
		if len(selected) == count {
			visited++
			if visited > maxCombinations {
				return false, ErrBracketTooLarge
			}
			if err := budget.consume(); err != nil {
				return false, err
			}
			return visit(append([]*swisslib.PlayerState(nil), selected...))
		}
		for i := start; i <= len(players)-(count-len(selected)); i++ {
			selected = append(selected, players[i])
			stop, err := walk(i + 1)
			if err != nil {
				return false, err
			}
			if stop {
				return true, nil
			}
			selected = selected[:len(selected)-1]
		}
		return false, nil
	}
	return walk(0)
}
