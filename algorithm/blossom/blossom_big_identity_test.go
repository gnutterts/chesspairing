// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package blossom

import (
	"encoding/json"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

// bigIdentitySeed is the fixed seed for the deterministic random graphs used
// by the big.Int matching identity test. It must never change, otherwise the
// recorded golden file no longer corresponds to the generated graphs. The
// golden is a fixed reference produced by the code before the buffer reuse
// change and must never be regenerated.
const bigIdentitySeed = int64(20260926)

// bigIdentityCase records the expected matching for one generated graph.
type bigIdentityCase struct {
	N        int   `json:"n"`
	Matching []int `json:"matching"`
}

// generateBigIdentityCases builds 30 deterministic random graphs with 20 to
// 200 vertices and edge weights up to 2^80, then computes the matching for
// each graph with MaxWeightMatchingBig.
func generateBigIdentityCases() []bigIdentityCase {
	rng := rand.New(rand.NewSource(bigIdentitySeed))
	maxWeight := new(big.Int).Lsh(big.NewInt(1), 80) // 2^80
	one := big.NewInt(1)

	cases := make([]bigIdentityCase, 30)
	for i := range cases {
		n := 20 + rng.Intn(181) // 20..200
		m := n + rng.Intn(2*n+1)
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

		cases[i] = bigIdentityCase{
			N:        n,
			Matching: MaxWeightMatchingBig(edges, true),
		}
	}
	return cases
}

func TestBigBlossomIdentityGolden(t *testing.T) {
	goldenPath := filepath.Join("testdata", "big_matching_identity.golden")

	data, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var want []bigIdentityCase
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatalf("unmarshal golden: %v", err)
	}

	got := generateBigIdentityCases()
	if len(got) != len(want) {
		t.Fatalf("case count mismatch: got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].N != want[i].N {
			t.Fatalf("case %d: node count mismatch: got %d, want %d", i, got[i].N, want[i].N)
		}
		// A graph whose highest vertex has no edges yields a shorter matching;
		// recorded as such.
		if len(got[i].Matching) != len(want[i].Matching) {
			t.Fatalf("case %d: matching length mismatch: got %d, want %d",
				i, len(got[i].Matching), len(want[i].Matching))
		}
		for j := range got[i].Matching {
			if got[i].Matching[j] != want[i].Matching[j] {
				t.Fatalf("case %d: matching[%d] mismatch: got %d, want %d",
					i, j, got[i].Matching[j], want[i].Matching[j])
			}
		}
	}
}
