// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package blossom

import (
	"math/big"
	"math/rand"
	"testing"
)

// bigBenchmarkSeed is the fixed seed used to build the deterministic
// benchmark graph. It must stay stable so benchmark runs are comparable.
const bigBenchmarkSeed = int64(42)

// buildBigBenchmarkGraph builds a deterministic random graph with n vertices
// and 3*n edges. Edge weights are uniform in [1, 2^80].
func buildBigBenchmarkGraph(n int) []BigEdge {
	rng := rand.New(rand.NewSource(bigBenchmarkSeed))
	maxWeight := new(big.Int).Lsh(big.NewInt(1), 80) // 2^80
	one := big.NewInt(1)

	m := 3 * n
	edges := make([]BigEdge, 0, m)
	seen := make(map[[2]int]bool)

	for len(edges) < m {
		a := rng.Intn(n)
		b := rng.Intn(n)
		if a == b {
			continue
		}
		if a > b {
			a, b = b, a
		}
		if seen[[2]int{a, b}] {
			continue
		}
		seen[[2]int{a, b}] = true

		weight := new(big.Int).Rand(rng, maxWeight)
		weight.Add(weight, one) // 1..2^80
		edges = append(edges, BigEdge{I: a, J: b, Weight: weight})
	}
	return edges
}

func BenchmarkMaxWeightMatchingBig_200(b *testing.B) {
	edges := buildBigBenchmarkGraph(200)
	b.ResetTimer()
	for b.Loop() {
		MaxWeightMatchingBig(edges, true)
	}
}
