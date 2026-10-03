// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"context"
	"sort"

	"github.com/gnutterts/chesspairing/algorithm/blossom"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

type bracketCandidate struct {
	pairs  [][2]*swisslib.PlayerState
	floats []*swisslib.PlayerState
	c6     []float64
	c7pair int
	c7     []float64
	c8     int
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
		pairs, floats, ok := bestBracket(ctx, bracket, lower, next, afterNext, forbidden, indices)
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

// bestBracket implements C.04.4.2 Articles 3.2 and 4.3 by enumerating the
// Article 4.3 order. It is deliberately local to one bracket.
func bestBracket(ctx context.Context, bracket, lower, next, afterNext []*swisslib.PlayerState, forbidden map[[2]string]bool, indices map[string]OppositionIndex) ([][2]*swisslib.PlayerState, []*swisslib.PlayerState, bool) {
	for nf := len(bracket) % 2; nf <= len(bracket); nf += 2 {
		var best *bracketCandidate
		enumerateBracket(ctx, bracket, nf, forbidden, func(pairs [][2]*swisslib.PlayerState, floats []*swisslib.PlayerState) bool {
			rest := append(append([]*swisslib.PlayerState{}, floats...), lower...)
			if !pairable(rest, forbidden) {
				return true
			}
			candidate := &bracketCandidate{pairs: append([][2]*swisslib.PlayerState{}, pairs...), floats: append([]*swisslib.PlayerState{}, floats...), c6: floatScores(floats)}
			if len(next) != 0 {
				nextBracket := append(append([]*swisslib.PlayerState{}, floats...), next...)
				candidate.c7pair, candidate.c7, _ = bracketQuality(nextBracket, afterNext, forbidden)
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
		})
		if best != nil {
			return best.pairs, best.floats, true
		}
	}
	return nil, nil, false
}

func bracketQuality(bracket, lower []*swisslib.PlayerState, forbidden map[[2]string]bool) (int, []float64, bool) {
	for nf := len(bracket) % 2; nf <= len(bracket); nf += 2 {
		found := false
		var best []float64
		enumerateBracket(context.Background(), bracket, nf, forbidden, func(_ [][2]*swisslib.PlayerState, floats []*swisslib.PlayerState) bool {
			if !pairable(append(append([]*swisslib.PlayerState{}, floats...), lower...), forbidden) {
				return true
			}
			value := floatScores(floats)
			if !found || lexLess(value, best) {
				best, found = value, true
			}
			return true
		})
		if found {
			return (len(bracket) - nf) / 2, best, true
		}
	}
	return 0, nil, false
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
	if a.c7pair != b.c7pair {
		return a.c7pair > b.c7pair
	}
	if lexLess(a.c7, b.c7) != lexLess(b.c7, a.c7) {
		return lexLess(a.c7, b.c7)
	}
	return a.c8 < b.c8
}
