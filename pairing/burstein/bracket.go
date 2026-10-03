// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"context"
	"sort"

	"github.com/gnutterts/chesspairing/algorithm/blossom"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

const maxFloaterCandidates = 200000

type bracketCandidate struct {
	pairs  [][2]*swisslib.PlayerState
	floats []*swisslib.PlayerState
	c6     []float64
	c7pair int
	c7     []float64
	c8     int
	c7ok   bool
}

// pairPostSeedingBrackets implements C.04.4.2 Articles 1.9, 3.2 and 4.3.
func pairPostSeedingBrackets(ctx context.Context, groups []swisslib.ScoreGroup, indices map[string]OppositionIndex, forbidden map[[2]string]bool) ([][2]*swisslib.PlayerState, error) {
	var result [][2]*swisslib.PlayerState
	var incoming []*swisslib.PlayerState
	for gi, group := range groups {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		bracket := append(append([]*swisslib.PlayerState{}, incoming...), group.Players...)
		sortBracket(bracket, indices)
		var lower []*swisslib.PlayerState
		for _, next := range groups[gi+1:] {
			lower = append(lower, next.Players...)
		}
		var next, afterNext []*swisslib.PlayerState
		if gi+1 < len(groups) {
			next = groups[gi+1].Players
			for _, group := range groups[gi+2:] {
				afterNext = append(afterNext, group.Players...)
			}
		}
		pairs, floats, ok, err := bestBracket(ctx, bracket, lower, next, afterNext, forbidden, indices)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrNoPairingPossible
		}
		result = append(result, pairs...)
		incoming = floats
	}
	if len(incoming) != 0 {
		return nil, ErrNoPairingPossible
	}
	return result, nil
}

// sortBracket implements C.04.4.2 Articles 1.8 and 4.1.
func sortBracket(players []*swisslib.PlayerState, indices map[string]OppositionIndex) {
	sort.SliceStable(players, func(i, j int) bool {
		a, b := indices[players[i].ID], indices[players[j].ID]
		if a.Buchholz != b.Buchholz {
			return a.Buchholz > b.Buchholz
		}
		if a.SonnebornBerger != b.SonnebornBerger {
			return a.SonnebornBerger > b.SonnebornBerger
		}
		return a.TPN < b.TPN
	})
}

// compatible implements C.04.4.2 Articles 2.1.1 and 2.1.3.
func compatible(a, b *swisslib.PlayerState, forbidden map[[2]string]bool) bool {
	if swisslib.HasPlayed(a, b) || forbidden[swisslib.CanonicalPairKey(a.ID, b.ID)] {
		return false
	}
	pa, pb := swisslib.ComputeColorPreference(a.ColorHistory), swisslib.ComputeColorPreference(b.ColorHistory)
	return !(pa.AbsolutePreference && pb.AbsolutePreference && pa.Color != nil && pb.Color != nil && *pa.Color == *pb.Color)
}

// pairable implements C.04.4.2 Article 2.2.1.
func pairable(players []*swisslib.PlayerState, forbidden map[[2]string]bool) bool {
	if len(players)%2 != 0 {
		return false
	}
	if len(players) == 0 {
		return true
	}
	edges := make([]blossom.BlossomEdge, 0, len(players)*len(players)/2)
	for i := range players {
		for j := i + 1; j < len(players); j++ {
			if compatible(players[i], players[j], forbidden) {
				edges = append(edges, blossom.BlossomEdge{I: i, J: j, Weight: 1})
			}
		}
	}
	m := blossom.MaxWeightMatching(edges, true)
	return len(m) == len(players) && func() bool {
		for _, p := range m {
			if p < 0 {
				return false
			}
		}
		return true
	}()
}

// bestBracket implements C.04.4.2 Articles 3.2 and 4.3. Small brackets use
// the literal Article 4.3 enumeration; larger ones use matching feasibility
// while selecting floaters and greedily fix Article 4.3 partners.
func bestBracket(ctx context.Context, bracket, lower, next, afterNext []*swisslib.PlayerState, forbidden map[[2]string]bool, indices map[string]OppositionIndex) ([][2]*swisslib.PlayerState, []*swisslib.PlayerState, bool, error) {
	for nf := len(bracket) % 2; nf <= len(bracket); nf += 2 {
		// The last bracket has nowhere to send floaters.
		if len(lower) == 0 && nf != 0 {
			break
		}
		if err := ctx.Err(); err != nil {
			return nil, nil, false, err
		}
		var best *bracketCandidate
		visit := func(floats []*swisslib.PlayerState) bool {
			if !pairable(append(append([]*swisslib.PlayerState{}, floats...), lower...), forbidden) {
				return true
			}
			pairs, ok := orderedPairs(ctx, without(bracket, floats), forbidden)
			if !ok {
				return ctx.Err() == nil
			}
			candidate := &bracketCandidate{pairs: pairs, floats: append([]*swisslib.PlayerState{}, floats...), c6: floatScores(floats), c7ok: true}
			if len(next) != 0 {
				var err error
				candidate.c7pair, candidate.c7, candidate.c7ok, err = bracketQuality(ctx, append(append([]*swisslib.PlayerState{}, floats...), next...), afterNext, forbidden)
				if err != nil || !candidate.c7ok {
					return err == nil
				}
			}
			for _, pair := range pairs {
				a, b := swisslib.ComputeColorPreference(pair[0].ColorHistory), swisslib.ComputeColorPreference(pair[1].ColorHistory)
				if a.Color != nil && b.Color != nil && *a.Color == *b.Color {
					candidate.c8++
				}
			}
			if best == nil || candidateBetter(candidate, best) {
				best = candidate
			}
			return true
		}
		if len(bracket) <= 10 {
			enumerateBracket(ctx, bracket, nf, forbidden, func(_ [][2]*swisslib.PlayerState, floats []*swisslib.PlayerState) bool { return visit(floats) })
		} else {
			// C5 can require more than the usual three floaters (for example
			// when several residents cannot meet).  This enumerates floater
			// sets, never the bracket's pairings.
			enumerateFloats(ctx, bracket, nf, visit)
		}
		if err := ctx.Err(); err != nil {
			return nil, nil, false, err
		}
		if best != nil {
			return best.pairs, best.floats, true, nil
		}
	}
	return nil, nil, false, nil
}

// bracketQuality implements C.04.4.2 criterion C7 without enumerating games.
func bracketQuality(ctx context.Context, bracket, lower []*swisslib.PlayerState, forbidden map[[2]string]bool) (int, []float64, bool, error) {
	for nf := len(bracket) % 2; nf <= len(bracket); nf += 2 {
		var best []float64
		found := false
		enumerateFloats(ctx, bracket, nf, func(floats []*swisslib.PlayerState) bool {
			if pairable(without(bracket, floats), forbidden) && pairable(append(append([]*swisslib.PlayerState{}, floats...), lower...), forbidden) {
				value := floatScores(floats)
				if !found || lexLess(value, best) {
					best, found = value, true
				}
			}
			return true
		})
		if err := ctx.Err(); err != nil {
			return 0, nil, false, err
		}
		if found {
			return (len(bracket) - nf) / 2, best, true, nil
		}
	}
	return 0, nil, false, nil
}

func enumerateFloats(ctx context.Context, players []*swisslib.PlayerState, needed int, fn func([]*swisslib.PlayerState) bool) bool {
	chosen := make([]*swisslib.PlayerState, 0, needed)
	visited := 0
	var visit func(int) bool
	visit = func(start int) bool {
		if ctx.Err() != nil {
			return false
		}
		if len(chosen) == needed {
			visited++
			if visited > maxFloaterCandidates {
				return false
			}
			return fn(chosen)
		}
		for i := start; i <= len(players)-(needed-len(chosen)); i++ {
			chosen = append(chosen, players[i])
			if !visit(i + 1) {
				return false
			}
			chosen = chosen[:len(chosen)-1]
		}
		return true
	}
	return visit(0)
}

func without(players, removed []*swisslib.PlayerState) []*swisslib.PlayerState {
	out := make([]*swisslib.PlayerState, 0, len(players)-len(removed))
	for _, player := range players {
		found := false
		for _, floater := range removed {
			if player == floater {
				found = true
				break
			}
		}
		if !found {
			out = append(out, player)
		}
	}
	return out
}

// orderedPairs implements Article 4.3. It fixes the highest available BSN
// partner first, subject to retaining the best C8 value.
func orderedPairs(ctx context.Context, players []*swisslib.PlayerState, forbidden map[[2]string]bool) ([][2]*swisslib.PlayerState, bool) {
	remaining := append([]*swisslib.PlayerState{}, players...)
	target, ok := minimumC8(remaining, forbidden)
	if !ok {
		return nil, false
	}
	pairs := make([][2]*swisslib.PlayerState, 0, len(players)/2)
	for len(remaining) != 0 {
		if ctx.Err() != nil {
			return nil, false
		}
		a := remaining[0]
		found := false
		for j := len(remaining) - 1; j > 0; j-- {
			rest := append([]*swisslib.PlayerState{}, remaining[1:j]...)
			rest = append(rest, remaining[j+1:]...)
			cost, possible := minimumC8(rest, forbidden)
			if compatible(a, remaining[j], forbidden) && possible && c8Cost(a, remaining[j])+cost == target {
				pairs = append(pairs, [2]*swisslib.PlayerState{a, remaining[j]})
				remaining, target, found = rest, cost, true
				break
			}
		}
		if !found {
			return nil, false
		}
	}
	return pairs, true
}

func c8Cost(a, b *swisslib.PlayerState) int {
	pa, pb := swisslib.ComputeColorPreference(a.ColorHistory), swisslib.ComputeColorPreference(b.ColorHistory)
	if pa.Color != nil && pb.Color != nil && *pa.Color == *pb.Color {
		return 1
	}
	return 0
}

func minimumC8(players []*swisslib.PlayerState, forbidden map[[2]string]bool) (int, bool) {
	if len(players) == 0 {
		return 0, true
	}
	edges := make([]blossom.BlossomEdge, 0, len(players)*len(players)/2)
	for i := range players {
		for j := i + 1; j < len(players); j++ {
			if compatible(players[i], players[j], forbidden) {
				edges = append(edges, blossom.BlossomEdge{I: i, J: j, Weight: int64(100 - c8Cost(players[i], players[j]))})
			}
		}
	}
	mate := blossom.MaxWeightMatching(edges, true)
	cost := 0
	for i, j := range mate {
		if j < 0 {
			return 0, false
		}
		if i < j {
			cost += c8Cost(players[i], players[j])
		}
	}
	return cost, true
}

func enumerateBracket(ctx context.Context, players []*swisslib.PlayerState, floatsNeeded int, forbidden map[[2]string]bool, fn func([][2]*swisslib.PlayerState, []*swisslib.PlayerState) bool) bool {
	used := make([]bool, len(players))
	var pairs [][2]*swisslib.PlayerState
	var floats []*swisslib.PlayerState
	var visit func() bool
	visit = func() bool {
		if ctx.Err() != nil {
			return false
		}
		i := -1
		for n := range players {
			if !used[n] {
				i = n
				break
			}
		}
		if i < 0 {
			if len(floats) != floatsNeeded {
				return true
			}
			return fn(pairs, floats)
		}
		used[i] = true
		for j := len(players) - 1; j > i; j-- {
			if !used[j] && compatible(players[i], players[j], forbidden) {
				used[j] = true
				pairs = append(pairs, [2]*swisslib.PlayerState{players[i], players[j]})
				if !visit() {
					return false
				}
				pairs = pairs[:len(pairs)-1]
				used[j] = false
			}
		}
		if len(floats) < floatsNeeded {
			floats = append(floats, players[i])
			if !visit() {
				return false
			}
			floats = floats[:len(floats)-1]
		}
		used[i] = false
		return true
	}
	return visit()
}

func floatScores(players []*swisslib.PlayerState) []float64 {
	scores := make([]float64, len(players))
	for i, p := range players {
		scores[i] = p.PairingScore
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(scores)))
	return scores
}
func lexLess(a, b []float64) bool {
	for i := range a {
		if i >= len(b) || a[i] != b[i] {
			return i < len(b) && a[i] < b[i]
		}
	}
	return len(a) < len(b)
}
func candidateBetter(a, b *bracketCandidate) bool {
	if lexLess(a.c6, b.c6) != lexLess(b.c6, a.c6) {
		return lexLess(a.c6, b.c6)
	}
	if a.c7ok != b.c7ok {
		return a.c7ok
	}
	if a.c7pair != b.c7pair {
		return a.c7pair > b.c7pair
	}
	if lexLess(a.c7, b.c7) != lexLess(b.c7, a.c7) {
		return lexLess(a.c7, b.c7)
	}
	return a.c8 < b.c8
}
