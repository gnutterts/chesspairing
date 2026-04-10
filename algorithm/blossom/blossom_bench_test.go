// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package blossom

import (
	"testing"
)

// buildCompleteGraph creates edges for a complete graph with n nodes.
// Edge weights are based on node indices to create a deterministic pattern.
func buildCompleteGraph(n int) []BlossomEdge {
	edges := make([]BlossomEdge, 0, n*(n-1)/2)
	for i := range n {
		for j := i + 1; j < n; j++ {
			// Weight based on difference to create non-trivial matching.
			w := int64((n - (j - i)) * 100)
			edges = append(edges, BlossomEdge{I: i, J: j, Weight: w})
		}
	}
	return edges
}

func BenchmarkMaxWeightMatching_10(b *testing.B) {
	edges := buildCompleteGraph(10)
	b.ResetTimer()
	for b.Loop() {
		MaxWeightMatching(edges, true)
	}
}

func BenchmarkMaxWeightMatching_20(b *testing.B) {
	edges := buildCompleteGraph(20)
	b.ResetTimer()
	for b.Loop() {
		MaxWeightMatching(edges, true)
	}
}

func BenchmarkMaxWeightMatching_50(b *testing.B) {
	edges := buildCompleteGraph(50)
	b.ResetTimer()
	for b.Loop() {
		MaxWeightMatching(edges, true)
	}
}

func BenchmarkMaxWeightMatching_100(b *testing.B) {
	edges := buildCompleteGraph(100)
	b.ResetTimer()
	for b.Loop() {
		MaxWeightMatching(edges, true)
	}
}
