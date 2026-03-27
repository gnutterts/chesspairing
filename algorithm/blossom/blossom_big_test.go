package blossom

import (
	"math"
	"math/big"
	"testing"
)

// Tests for MaxWeightMatchingBig — mirrors blossom_test.go but uses BigEdge.

func bigEdge(i, j int, w int64) BigEdge {
	return BigEdge{I: i, J: j, Weight: big.NewInt(w)}
}

func assertMatchingBig(t *testing.T, name string, edges []BigEdge, maxCard bool, expected []int) {
	t.Helper()
	m := MaxWeightMatchingBig(edges, maxCard)
	if len(m) != len(expected) {
		t.Fatalf("%s: length mismatch: got %v (len %d), expected %v (len %d)",
			name, m, len(m), expected, len(expected))
	}
	for i := range m {
		if m[i] != expected[i] {
			t.Fatalf("%s: mismatch at [%d]: got %v, expected %v", name, i, m, expected)
		}
	}
}

func TestBigBlossomEmpty(t *testing.T) {
	m := MaxWeightMatchingBig(nil, true)
	if len(m) != 0 {
		t.Fatalf("expected empty, got %v", m)
	}
}

func TestBigBlossomSingleEdge(t *testing.T) {
	assertMatchingBig(t, "single", []BigEdge{bigEdge(0, 1, 1)}, false, []int{1, 0})
}

func TestBigBlossom12(t *testing.T) {
	assertMatchingBig(t, "12", []BigEdge{bigEdge(0, 1, 10), bigEdge(1, 2, 11)}, false, []int{-1, 2, 1})
}

func TestBigBlossom13(t *testing.T) {
	assertMatchingBig(t, "13", []BigEdge{bigEdge(0, 1, 5), bigEdge(1, 2, 11), bigEdge(2, 3, 5)}, false, []int{-1, 2, 1, -1})
}

func TestBigBlossom14_MaxCard(t *testing.T) {
	assertMatchingBig(t, "14", []BigEdge{bigEdge(0, 1, 5), bigEdge(1, 2, 11), bigEdge(2, 3, 5)}, true, []int{1, 0, 3, 2})
}

func TestBigBlossom20_SBlossom(t *testing.T) {
	assertMatchingBig(t, "20a", []BigEdge{bigEdge(0, 1, 8), bigEdge(0, 2, 9), bigEdge(1, 2, 10), bigEdge(2, 3, 7)}, false, []int{1, 0, 3, 2})
	assertMatchingBig(t, "20b", []BigEdge{bigEdge(0, 1, 8), bigEdge(0, 2, 9), bigEdge(1, 2, 10), bigEdge(2, 3, 7), bigEdge(0, 5, 5), bigEdge(3, 4, 6)}, false, []int{5, 2, 1, 4, 3, 0})
}

func TestBigBlossom21_TBlossom(t *testing.T) {
	assertMatchingBig(t, "21a", []BigEdge{bigEdge(0, 1, 9), bigEdge(0, 2, 8), bigEdge(1, 2, 10), bigEdge(0, 3, 5), bigEdge(3, 4, 4), bigEdge(0, 5, 3)}, false, []int{5, 2, 1, 4, 3, 0})
	assertMatchingBig(t, "21b", []BigEdge{bigEdge(0, 1, 9), bigEdge(0, 2, 8), bigEdge(1, 2, 10), bigEdge(0, 3, 5), bigEdge(3, 4, 3), bigEdge(0, 5, 4)}, false, []int{5, 2, 1, 4, 3, 0})
	assertMatchingBig(t, "21c", []BigEdge{bigEdge(0, 1, 9), bigEdge(0, 2, 8), bigEdge(1, 2, 10), bigEdge(0, 3, 5), bigEdge(3, 4, 3), bigEdge(2, 5, 4)}, false, []int{1, 0, 5, 4, 3, 2})
}

func TestBigBlossom22_SNest(t *testing.T) {
	assertMatchingBig(t, "22", []BigEdge{bigEdge(0, 1, 9), bigEdge(0, 2, 9), bigEdge(1, 2, 10), bigEdge(1, 3, 8), bigEdge(2, 4, 8), bigEdge(3, 4, 10), bigEdge(4, 5, 6)}, false, []int{2, 3, 0, 1, 5, 4})
}

func TestBigBlossom25_STExpand(t *testing.T) {
	assertMatchingBig(t, "25", []BigEdge{
		bigEdge(0, 1, 23), bigEdge(0, 4, 22), bigEdge(0, 5, 15), bigEdge(1, 2, 25), bigEdge(2, 3, 22),
		bigEdge(3, 4, 25), bigEdge(3, 7, 14), bigEdge(4, 6, 13),
	}, false, []int{5, 2, 1, 7, 6, 0, 4, 3})
}

func TestBigBlossom30_TNastyExpand(t *testing.T) {
	assertMatchingBig(t, "30", []BigEdge{
		bigEdge(0, 1, 45), bigEdge(0, 4, 45), bigEdge(1, 2, 50), bigEdge(2, 3, 45), bigEdge(3, 4, 50),
		bigEdge(0, 5, 30), bigEdge(2, 8, 35), bigEdge(3, 7, 35), bigEdge(4, 6, 26), bigEdge(8, 9, 5),
	}, false, []int{5, 2, 1, 7, 6, 0, 4, 3, 9, 8})
}

func TestBigBlossomDisconnected(t *testing.T) {
	assertMatchingBig(t, "disconnected", []BigEdge{bigEdge(0, 1, 5), bigEdge(2, 3, 8)}, true, []int{1, 0, 3, 2})
}

func TestBigBlossomLargeWeights(t *testing.T) {
	// Test with weights > 64 bits to verify big.Int works correctly.
	w1 := new(big.Int).Lsh(big.NewInt(1), 100) // 2^100
	w2 := new(big.Int).Lsh(big.NewInt(1), 80)  // 2^80
	w3 := new(big.Int).Lsh(big.NewInt(1), 60)  // 2^60

	edges := []BigEdge{
		{I: 0, J: 1, Weight: w1},
		{I: 1, J: 2, Weight: w2},
		{I: 2, J: 3, Weight: w3},
	}
	// Should match 0-1 (highest weight) and 2-3 (only option left).
	assertMatchingBig(t, "largeWeights", edges, true, []int{1, 0, 3, 2})
}

func TestBigBlossomLargeGraph(t *testing.T) {
	// 20-vertex complete graph. Must terminate quickly.
	n := 20
	var edges []BigEdge
	w := int64(1)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			edges = append(edges, BigEdge{I: i, J: j, Weight: big.NewInt(w)})
			w++
		}
	}
	m := MaxWeightMatchingBig(edges, true)
	matched := 0
	for _, v := range m {
		if v != -1 {
			matched++
		}
	}
	if matched != n {
		t.Fatalf("expected %d matched, got %d: %v", n, matched, m)
	}
}

// TestBigBlossomMatchesInt64 verifies that MaxWeightMatchingBig produces
// identical results to MaxWeightMatching for the same edges when weights
// fit in int64.
func TestBigBlossomMatchesInt64(t *testing.T) {
	allTests := []struct {
		name    string
		edges   []BlossomEdge
		maxCard bool
	}{
		{"10_empty", nil, false},
		{"11_single", []BlossomEdge{{0, 1, 1}}, false},
		{"12", []BlossomEdge{{0, 1, 10}, {1, 2, 11}}, false},
		{"13", []BlossomEdge{{0, 1, 5}, {1, 2, 11}, {2, 3, 5}}, false},
		{"14_maxCard", []BlossomEdge{{0, 1, 5}, {1, 2, 11}, {2, 3, 5}}, true},
		{"16_neg", []BlossomEdge{{0, 1, 2}, {0, 2, -2}, {1, 2, 1}, {1, 3, -1}, {2, 3, -6}}, false},
		{"16_neg_maxCard", []BlossomEdge{{0, 1, 2}, {0, 2, -2}, {1, 2, 1}, {1, 3, -1}, {2, 3, -6}}, true},
		{"20a", []BlossomEdge{{0, 1, 8}, {0, 2, 9}, {1, 2, 10}, {2, 3, 7}}, false},
		{"20b", []BlossomEdge{{0, 1, 8}, {0, 2, 9}, {1, 2, 10}, {2, 3, 7}, {0, 5, 5}, {3, 4, 6}}, false},
		{"21a", []BlossomEdge{{0, 1, 9}, {0, 2, 8}, {1, 2, 10}, {0, 3, 5}, {3, 4, 4}, {0, 5, 3}}, false},
		{"21b", []BlossomEdge{{0, 1, 9}, {0, 2, 8}, {1, 2, 10}, {0, 3, 5}, {3, 4, 3}, {0, 5, 4}}, false},
		{"21c", []BlossomEdge{{0, 1, 9}, {0, 2, 8}, {1, 2, 10}, {0, 3, 5}, {3, 4, 3}, {2, 5, 4}}, false},
		{"22", []BlossomEdge{
			{0, 1, 9}, {0, 2, 9}, {1, 2, 10}, {1, 3, 8}, {2, 4, 8}, {3, 4, 10}, {4, 5, 6},
		}, false},
		{"triangle", []BlossomEdge{{0, 1, 5}, {1, 2, 3}, {0, 2, 4}}, false},
		{"K4", []BlossomEdge{{0, 1, 1}, {0, 2, 2}, {0, 3, 3}, {1, 2, 4}, {1, 3, 5}, {2, 3, 6}}, false},
		{"disconnected", []BlossomEdge{{0, 1, 5}, {2, 3, 8}}, true},
	}

	for _, tc := range allTests {
		t.Run(tc.name, func(t *testing.T) {
			expected := MaxWeightMatching(tc.edges, tc.maxCard)

			bigEdges := make([]BigEdge, len(tc.edges))
			for i, e := range tc.edges {
				bigEdges[i] = BigEdge{I: e.I, J: e.J, Weight: big.NewInt(e.Weight)}
			}

			got := MaxWeightMatchingBig(bigEdges, tc.maxCard)

			if len(got) != len(expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(got), len(expected))
			}
			for i := range got {
				if got[i] != expected[i] {
					t.Errorf("m[%d] = %d, want %d", i, got[i], expected[i])
				}
			}
		})
	}
}

func TestBigBlossom_int64Boundary(t *testing.T) {
	maxWeight := new(big.Int).SetInt64(math.MaxInt64)
	halfMax := new(big.Int).Rsh(maxWeight, 1) // 2^62

	edges := []BigEdge{
		{0, 1, new(big.Int).Set(maxWeight)},
		{1, 2, new(big.Int).Set(halfMax)},
		{0, 2, big.NewInt(1)},
	}

	result := MaxWeightMatchingBig(edges, false)
	if result[0] != 1 || result[1] != 0 {
		t.Errorf("expected (0,1) matched, got %v", result)
	}
}
