package blossom

import "testing"

// Tests ported from van Rantwijk's Python reference implementation.
// Python uses 1-based vertex indices; converted to 0-based here.
// Conversion: edge (i,j,w) → (i-1, j-1, w); expected[k] = python[k+1]-1 if python[k+1]≥0, else -1.

func assertMatching(t *testing.T, name string, edges []BlossomEdge, maxCard bool, expected []int) {
	t.Helper()
	m := MaxWeightMatching(edges, maxCard)
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

func TestBlossomEmpty(t *testing.T) {
	m := MaxWeightMatching(nil, true)
	if len(m) != 0 {
		t.Fatalf("expected empty, got %v", m)
	}
}

func TestBlossomSingleEdge(t *testing.T) {
	// Python: (0,1,1) → [1,0]
	assertMatching(t, "single", []BlossomEdge{{0, 1, 1}}, false, []int{1, 0})
}

func TestBlossom12(t *testing.T) {
	// Python: (1,2,10),(2,3,11) → [-1,-1,3,2]
	// 0-based: (0,1,10),(1,2,11) → [-1,2,1]
	assertMatching(t, "12", []BlossomEdge{{0, 1, 10}, {1, 2, 11}}, false, []int{-1, 2, 1})
}

func TestBlossom13(t *testing.T) {
	// Python: (1,2,5),(2,3,11),(3,4,5) → [-1,-1,3,2,-1]
	// 0-based: (0,1,5),(1,2,11),(2,3,5) → [-1,2,1,-1]
	assertMatching(t, "13", []BlossomEdge{{0, 1, 5}, {1, 2, 11}, {2, 3, 5}}, false, []int{-1, 2, 1, -1})
}

func TestBlossom14_MaxCard(t *testing.T) {
	// Python: (1,2,5),(2,3,11),(3,4,5) maxCard → [-1,2,1,4,3]
	// 0-based: → [1,0,3,2]
	assertMatching(t, "14", []BlossomEdge{{0, 1, 5}, {1, 2, 11}, {2, 3, 5}}, true, []int{1, 0, 3, 2})
}

func TestBlossom16_Negative(t *testing.T) {
	// Python: (1,2,2),(1,3,-2),(2,3,1),(2,4,-1),(3,4,-6) false → [-1,2,1,-1,-1]
	// 0-based: (0,1,2),(0,2,-2),(1,2,1),(1,3,-1),(2,3,-6) → [1,0,-1,-1]
	assertMatching(t, "16a", []BlossomEdge{{0, 1, 2}, {0, 2, -2}, {1, 2, 1}, {1, 3, -1}, {2, 3, -6}}, false, []int{1, 0, -1, -1})

	// Same edges, maxCard=true → [-1,3,4,1,2]
	// 0-based: → [2,3,0,1]
	assertMatching(t, "16b", []BlossomEdge{{0, 1, 2}, {0, 2, -2}, {1, 2, 1}, {1, 3, -1}, {2, 3, -6}}, true, []int{2, 3, 0, 1})
}

func TestBlossom20_SBlossom(t *testing.T) {
	// Python test20: (1,2,8),(1,3,9),(2,3,10),(3,4,7) → [-1,2,1,4,3]
	// 0-based: (0,1,8),(0,2,9),(1,2,10),(2,3,7) → [1,0,3,2]
	assertMatching(t, "20a", []BlossomEdge{{0, 1, 8}, {0, 2, 9}, {1, 2, 10}, {2, 3, 7}}, false, []int{1, 0, 3, 2})

	// Python: (1,2,8),(1,3,9),(2,3,10),(3,4,7),(1,6,5),(4,5,6) → [-1,6,3,2,5,4,1]
	// 0-based: (0,1,8),(0,2,9),(1,2,10),(2,3,7),(0,5,5),(3,4,6) → [5,2,1,4,3,0]
	assertMatching(t, "20b", []BlossomEdge{{0, 1, 8}, {0, 2, 9}, {1, 2, 10}, {2, 3, 7}, {0, 5, 5}, {3, 4, 6}}, false, []int{5, 2, 1, 4, 3, 0})
}

func TestBlossom21_TBlossom(t *testing.T) {
	// Python: (1,2,9),(1,3,8),(2,3,10),(1,4,5),(4,5,4),(1,6,3) → [-1,6,3,2,5,4,1]
	// 0-based: (0,1,9),(0,2,8),(1,2,10),(0,3,5),(3,4,4),(0,5,3) → [5,2,1,4,3,0]
	assertMatching(t, "21a", []BlossomEdge{{0, 1, 9}, {0, 2, 8}, {1, 2, 10}, {0, 3, 5}, {3, 4, 4}, {0, 5, 3}}, false, []int{5, 2, 1, 4, 3, 0})

	// Python: (1,2,9),(1,3,8),(2,3,10),(1,4,5),(4,5,3),(1,6,4) → [-1,6,3,2,5,4,1]
	// 0-based: (0,1,9),(0,2,8),(1,2,10),(0,3,5),(3,4,3),(0,5,4) → [5,2,1,4,3,0]
	assertMatching(t, "21b", []BlossomEdge{{0, 1, 9}, {0, 2, 8}, {1, 2, 10}, {0, 3, 5}, {3, 4, 3}, {0, 5, 4}}, false, []int{5, 2, 1, 4, 3, 0})

	// Python: (1,2,9),(1,3,8),(2,3,10),(1,4,5),(4,5,3),(3,6,4) → [-1,2,1,6,5,4,3]
	// 0-based: (0,1,9),(0,2,8),(1,2,10),(0,3,5),(3,4,3),(2,5,4) → [1,0,5,4,3,2]
	assertMatching(t, "21c", []BlossomEdge{{0, 1, 9}, {0, 2, 8}, {1, 2, 10}, {0, 3, 5}, {3, 4, 3}, {2, 5, 4}}, false, []int{1, 0, 5, 4, 3, 2})
}

func TestBlossom22_SNest(t *testing.T) {
	// Python: (1,2,9),(1,3,9),(2,3,10),(2,4,8),(3,5,8),(4,5,10),(5,6,6) → [-1,3,4,1,2,6,5]
	// 0-based: (0,1,9),(0,2,9),(1,2,10),(1,3,8),(2,4,8),(3,4,10),(4,5,6) → [2,3,0,1,5,4]
	assertMatching(t, "22", []BlossomEdge{{0, 1, 9}, {0, 2, 9}, {1, 2, 10}, {1, 3, 8}, {2, 4, 8}, {3, 4, 10}, {4, 5, 6}}, false, []int{2, 3, 0, 1, 5, 4})
}

func TestBlossom23_SRelabelNest(t *testing.T) {
	// Python: (1,2,10),(1,7,10),(2,3,12),(3,4,20),(3,5,20),(4,5,25),(5,6,10),(6,7,10),(7,8,8)
	//       → [-1,2,1,4,3,6,5,8,7]
	// 0-based: (0,1,10),(0,6,10),(1,2,12),(2,3,20),(2,4,20),(3,4,25),(4,5,10),(5,6,10),(6,7,8)
	//        → [1,0,3,2,5,4,7,6]
	assertMatching(t, "23", []BlossomEdge{
		{0, 1, 10}, {0, 6, 10}, {1, 2, 12}, {2, 3, 20}, {2, 4, 20},
		{3, 4, 25}, {4, 5, 10}, {5, 6, 10}, {6, 7, 8},
	}, false, []int{1, 0, 3, 2, 5, 4, 7, 6})
}

func TestBlossom24_SNestExpand(t *testing.T) {
	// Python: (1,2,8),(1,3,8),(2,3,10),(2,4,12),(3,5,12),(4,5,14),(4,6,12),(5,7,12),(6,7,14),(7,8,12)
	//       → [-1,2,1,5,6,3,4,8,7]
	// 0-based: (0,1,8),(0,2,8),(1,2,10),(1,3,12),(2,4,12),(3,4,14),(3,5,12),(4,6,12),(5,6,14),(6,7,12)
	//        → [1,0,4,5,2,3,7,6]
	assertMatching(t, "24", []BlossomEdge{
		{0, 1, 8}, {0, 2, 8}, {1, 2, 10}, {1, 3, 12}, {2, 4, 12},
		{3, 4, 14}, {3, 5, 12}, {4, 6, 12}, {5, 6, 14}, {6, 7, 12},
	}, false, []int{1, 0, 4, 5, 2, 3, 7, 6})
}

func TestBlossom25_STExpand(t *testing.T) {
	// Python: (1,2,23),(1,5,22),(1,6,15),(2,3,25),(3,4,22),(4,5,25),(4,8,14),(5,7,13)
	//       → [-1,6,3,2,8,7,1,5,4]
	// 0-based: (0,1,23),(0,4,22),(0,5,15),(1,2,25),(2,3,22),(3,4,25),(3,7,14),(4,6,13)
	//        → [5,2,1,7,6,0,4,3]
	assertMatching(t, "25", []BlossomEdge{
		{0, 1, 23}, {0, 4, 22}, {0, 5, 15}, {1, 2, 25}, {2, 3, 22},
		{3, 4, 25}, {3, 7, 14}, {4, 6, 13},
	}, false, []int{5, 2, 1, 7, 6, 0, 4, 3})
}

func TestBlossom26_SNestTExpand(t *testing.T) {
	// Python: (1,2,19),(1,3,20),(1,8,8),(2,3,25),(2,4,18),(3,5,18),(4,5,13),(4,7,7),(5,6,7)
	//       → [-1,8,3,2,7,6,5,4,1]
	// 0-based: (0,1,19),(0,2,20),(0,7,8),(1,2,25),(1,3,18),(2,4,18),(3,4,13),(3,6,7),(4,5,7)
	//        → [7,2,1,6,5,4,3,0]
	assertMatching(t, "26", []BlossomEdge{
		{0, 1, 19}, {0, 2, 20}, {0, 7, 8}, {1, 2, 25}, {1, 3, 18},
		{2, 4, 18}, {3, 4, 13}, {3, 6, 7}, {4, 5, 7},
	}, false, []int{7, 2, 1, 6, 5, 4, 3, 0})
}

func TestBlossom30_TNastyExpand(t *testing.T) {
	// Python: (1,2,45),(1,5,45),(2,3,50),(3,4,45),(4,5,50),(1,6,30),(3,9,35),(4,8,35),(5,7,26),(9,10,5)
	//       → [-1,6,3,2,8,7,1,5,4,10,9]
	// 0-based: (0,1,45),(0,4,45),(1,2,50),(2,3,45),(3,4,50),(0,5,30),(2,8,35),(3,7,35),(4,6,26),(8,9,5)
	//        → [5,2,1,7,6,0,4,3,9,8]
	assertMatching(t, "30", []BlossomEdge{
		{0, 1, 45}, {0, 4, 45}, {1, 2, 50}, {2, 3, 45}, {3, 4, 50},
		{0, 5, 30}, {2, 8, 35}, {3, 7, 35}, {4, 6, 26}, {8, 9, 5},
	}, false, []int{5, 2, 1, 7, 6, 0, 4, 3, 9, 8})
}

func TestBlossom31_TNasty2Expand(t *testing.T) {
	// Python: (1,2,45),(1,5,45),(2,3,50),(3,4,45),(4,5,50),(1,6,30),(3,9,35),(4,8,26),(5,7,40),(9,10,5)
	//       → [-1,6,3,2,8,7,1,5,4,10,9]
	// 0-based: (0,1,45),(0,4,45),(1,2,50),(2,3,45),(3,4,50),(0,5,30),(2,8,35),(3,7,26),(4,6,40),(8,9,5)
	//        → [5,2,1,7,6,0,4,3,9,8]
	assertMatching(t, "31", []BlossomEdge{
		{0, 1, 45}, {0, 4, 45}, {1, 2, 50}, {2, 3, 45}, {3, 4, 50},
		{0, 5, 30}, {2, 8, 35}, {3, 7, 26}, {4, 6, 40}, {8, 9, 5},
	}, false, []int{5, 2, 1, 7, 6, 0, 4, 3, 9, 8})
}

func TestBlossom32_TExpandLeastSlack(t *testing.T) {
	// Python: (1,2,45),(1,5,45),(2,3,50),(3,4,45),(4,5,50),(1,6,30),(3,9,35),(4,8,28),(5,7,26),(9,10,5)
	//       → [-1,6,3,2,8,7,1,5,4,10,9]
	// 0-based: (0,1,45),(0,4,45),(1,2,50),(2,3,45),(3,4,50),(0,5,30),(2,8,35),(3,7,28),(4,6,26),(8,9,5)
	//        → [5,2,1,7,6,0,4,3,9,8]
	assertMatching(t, "32", []BlossomEdge{
		{0, 1, 45}, {0, 4, 45}, {1, 2, 50}, {2, 3, 45}, {3, 4, 50},
		{0, 5, 30}, {2, 8, 35}, {3, 7, 28}, {4, 6, 26}, {8, 9, 5},
	}, false, []int{5, 2, 1, 7, 6, 0, 4, 3, 9, 8})
}

func TestBlossom33_NestTNastyExpand(t *testing.T) {
	// Python: (1,2,45),(1,7,45),(2,3,50),(3,4,45),(4,5,95),(4,6,94),(5,6,94),(6,7,50),(1,8,30),(3,11,35),(5,9,36),(7,10,26),(11,12,5)
	//       → [-1,8,3,2,6,9,4,10,1,5,7,12,11]
	// 0-based: (0,1,45),(0,6,45),(1,2,50),(2,3,45),(3,4,95),(3,5,94),(4,5,94),(5,6,50),(0,7,30),(2,10,35),(4,8,36),(6,9,26),(10,11,5)
	//        → [7,2,1,5,8,3,9,0,4,6,11,10]
	assertMatching(t, "33", []BlossomEdge{
		{0, 1, 45}, {0, 6, 45}, {1, 2, 50}, {2, 3, 45}, {3, 4, 95},
		{3, 5, 94}, {4, 5, 94}, {5, 6, 50}, {0, 7, 30}, {2, 10, 35},
		{4, 8, 36}, {6, 9, 26}, {10, 11, 5},
	}, false, []int{7, 2, 1, 5, 8, 3, 9, 0, 4, 6, 11, 10})
}

func TestBlossom34_NestRelabelExpand(t *testing.T) {
	// Python: (1,2,40),(1,3,40),(2,3,60),(2,4,55),(3,5,55),(4,5,50),(1,8,15),(5,7,30),(7,6,10),(8,10,10),(4,9,30)
	//       → [-1,2,1,5,9,3,7,6,10,4,8]
	// 0-based: (0,1,40),(0,2,40),(1,2,60),(1,3,55),(2,4,55),(3,4,50),(0,7,15),(4,6,30),(6,5,10),(7,9,10),(3,8,30)
	//        → [1,0,4,8,2,6,5,9,3,7]
	assertMatching(t, "34", []BlossomEdge{
		{0, 1, 40}, {0, 2, 40}, {1, 2, 60}, {1, 3, 55}, {2, 4, 55},
		{3, 4, 50}, {0, 7, 15}, {4, 6, 30}, {6, 5, 10}, {7, 9, 10}, {3, 8, 30},
	}, false, []int{1, 0, 4, 8, 2, 6, 5, 9, 3, 7})
}

// Additional structural tests.

func TestBlossomTriangle(t *testing.T) {
	// Triangle: only 1 pair possible from 3 vertices.
	edges := []BlossomEdge{{0, 1, 10}, {1, 2, 10}, {2, 0, 10}}
	m := MaxWeightMatching(edges, false)
	matched := 0
	for _, v := range m {
		if v != -1 {
			matched++
		}
	}
	if matched != 2 {
		t.Fatalf("expected 2 matched in triangle, got %d: %v", matched, m)
	}
}

func TestBlossomFourVerticesSquare(t *testing.T) {
	edges := []BlossomEdge{{0, 1, 10}, {1, 2, 8}, {2, 3, 10}, {3, 0, 8}}
	m := MaxWeightMatching(edges, true)
	matched := 0
	for _, v := range m {
		if v != -1 {
			matched++
		}
	}
	if matched != 4 {
		t.Fatalf("expected 4 matched in square, got %d: %v", matched, m)
	}
}

func TestBlossomDisconnected(t *testing.T) {
	assertMatching(t, "disconnected", []BlossomEdge{{0, 1, 5}, {2, 3, 8}}, true, []int{1, 0, 3, 2})
}

func TestBlossomLargeGraph(t *testing.T) {
	// 20-vertex complete graph. Must terminate quickly.
	n := 20
	var edges []BlossomEdge
	w := int64(1)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			edges = append(edges, BlossomEdge{I: i, J: j, Weight: w})
			w++
		}
	}
	m := MaxWeightMatching(edges, true)
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

func TestBlossomSymmetry(t *testing.T) {
	edges := []BlossomEdge{{0, 1, 10}, {2, 3, 8}, {0, 3, 6}, {1, 2, 4}}
	m := MaxWeightMatching(edges, true)
	for v, u := range m {
		if u == -1 {
			continue
		}
		if m[u] != v {
			t.Fatalf("asymmetric: m[%d]=%d but m[%d]=%d", v, u, u, m[u])
		}
	}
}

// --- Task 1: Python reference tests 10 and 11 ---

func TestBlossom10(t *testing.T) {
	// Python test 10: empty input.
	result := MaxWeightMatching(nil, false)
	if result != nil {
		t.Errorf("test10: expected nil for empty input, got %v", result)
	}
}

func TestBlossom11(t *testing.T) {
	// Python test 11: single edge.
	edges := []BlossomEdge{{0, 1, 1}}
	assertMatching(t, "test11", edges, false, []int{1, 0})
}

// --- Task 2: Structural invariant tests ---

func TestBlossomInvariant_matchingValidity(t *testing.T) {
	testCases := []struct {
		name  string
		edges []BlossomEdge
	}{
		{"triangle", []BlossomEdge{{0, 1, 5}, {1, 2, 3}, {0, 2, 4}}},
		{"K4", []BlossomEdge{{0, 1, 1}, {0, 2, 2}, {0, 3, 3}, {1, 2, 4}, {1, 3, 5}, {2, 3, 6}}},
		{"path5", []BlossomEdge{{0, 1, 10}, {1, 2, 8}, {2, 3, 6}, {3, 4, 4}}},
		{"star", []BlossomEdge{{0, 1, 3}, {0, 2, 5}, {0, 3, 7}, {0, 4, 2}}},
		{"two_triangles", []BlossomEdge{
			{0, 1, 5}, {1, 2, 3}, {0, 2, 4},
			{3, 4, 7}, {4, 5, 2}, {3, 5, 6},
		}},
	}

	for _, tc := range testCases {
		for _, maxCard := range []bool{false, true} {
			name := tc.name
			if maxCard {
				name += "_maxCard"
			}
			t.Run(name, func(t *testing.T) {
				m := MaxWeightMatching(tc.edges, maxCard)

				// Build edge lookup.
				edgeSet := make(map[[2]int]bool)
				nvertex := 0
				for _, e := range tc.edges {
					edgeSet[[2]int{e.I, e.J}] = true
					edgeSet[[2]int{e.J, e.I}] = true
					if e.I+1 > nvertex {
						nvertex = e.I + 1
					}
					if e.J+1 > nvertex {
						nvertex = e.J + 1
					}
				}

				if len(m) != nvertex {
					t.Fatalf("matching length %d != vertex count %d", len(m), nvertex)
				}

				for v, partner := range m {
					if partner == -1 {
						continue
					}
					if partner < 0 || partner >= nvertex {
						t.Errorf("vertex %d matched to out-of-range %d", v, partner)
						continue
					}
					if m[partner] != v {
						t.Errorf("m[%d]=%d but m[%d]=%d (not symmetric)", v, partner, partner, m[partner])
					}
					if !edgeSet[[2]int{v, partner}] {
						t.Errorf("matched pair (%d,%d) is not an edge in the graph", v, partner)
					}
				}
			})
		}
	}
}

func TestBlossomInvariant_weightOptimality(t *testing.T) {
	testCases := []struct {
		name     string
		edges    []BlossomEdge
		nvertex  int
		expected int64
	}{
		{
			"triangle",
			[]BlossomEdge{{0, 1, 5}, {1, 2, 3}, {0, 2, 4}},
			3, 5,
		},
		{
			"K4_best_pair",
			[]BlossomEdge{{0, 1, 1}, {0, 2, 2}, {0, 3, 3}, {1, 2, 4}, {1, 3, 5}, {2, 3, 6}},
			4, 7,
		},
		{
			"path4",
			[]BlossomEdge{{0, 1, 10}, {1, 2, 8}, {2, 3, 6}},
			4, 16,
		},
		{
			"weighted_triangle_with_pendant",
			[]BlossomEdge{{0, 1, 10}, {1, 2, 5}, {2, 3, 20}},
			4, 30,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := MaxWeightMatching(tc.edges, false)

			var totalWeight int64
			matched := make(map[int]bool)
			for v, partner := range m {
				if partner == -1 || matched[v] {
					continue
				}
				matched[v] = true
				matched[partner] = true
				for _, e := range tc.edges {
					if (e.I == v && e.J == partner) || (e.I == partner && e.J == v) {
						totalWeight += e.Weight
						break
					}
				}
			}

			if totalWeight != tc.expected {
				t.Errorf("total weight = %d, want %d (matching: %v)", totalWeight, tc.expected, m)
			}
		})
	}
}

// --- Task 3: Edge ordering independence ---

func TestBlossomEdgeOrderIndependence(t *testing.T) {
	baseEdges := []BlossomEdge{
		{0, 1, 10}, {0, 2, 7}, {1, 2, 3},
		{2, 3, 8}, {3, 4, 5}, {4, 5, 9},
		{3, 5, 4}, {0, 5, 6},
	}

	refResult := MaxWeightMatching(baseEdges, false)
	refWeight := matchingWeight(baseEdges, refResult)

	perms := [][]int{
		{7, 6, 5, 4, 3, 2, 1, 0},
		{3, 7, 1, 5, 0, 4, 2, 6},
		{4, 0, 6, 2, 7, 3, 1, 5},
	}

	for i, perm := range perms {
		shuffled := make([]BlossomEdge, len(baseEdges))
		for j, idx := range perm {
			shuffled[j] = baseEdges[idx]
		}

		result := MaxWeightMatching(shuffled, false)
		weight := matchingWeight(shuffled, result)

		if weight != refWeight {
			t.Errorf("permutation %d: weight %d != reference %d", i, weight, refWeight)
		}
	}
}

// matchingWeight computes the total weight of a matching.
func matchingWeight(edges []BlossomEdge, m []int) int64 {
	var total int64
	seen := make(map[int]bool)
	for v, partner := range m {
		if partner == -1 || seen[v] {
			continue
		}
		seen[v] = true
		seen[partner] = true
		for _, e := range edges {
			if (e.I == v && e.J == partner) || (e.I == partner && e.J == v) {
				total += e.Weight
				break
			}
		}
	}
	return total
}

// --- Task 7: All-negative weights and empty edges ---

func TestBlossomAllNegativeWeights(t *testing.T) {
	edges := []BlossomEdge{
		{0, 1, -5},
		{1, 2, -3},
		{0, 2, -8},
		{2, 3, -1},
	}

	t.Run("maxWeight_false", func(t *testing.T) {
		m := MaxWeightMatching(edges, false)
		for v, partner := range m {
			if partner != -1 {
				t.Errorf("vertex %d matched to %d, expected all unmatched", v, partner)
			}
		}
	})

	t.Run("maxCard_true", func(t *testing.T) {
		m := MaxWeightMatching(edges, true)
		matchedCount := 0
		for _, partner := range m {
			if partner != -1 {
				matchedCount++
			}
		}
		if matchedCount != 4 {
			t.Errorf("maxCardinality: %d matched vertices, want 4", matchedCount)
		}
	})
}

func TestBlossomSingleVertexNoEdges(t *testing.T) {
	edges := []BlossomEdge{}
	m := MaxWeightMatching(edges, false)
	if m != nil {
		t.Errorf("expected nil for empty edges, got %v", m)
	}

	m = MaxWeightMatching(edges, true)
	if m != nil {
		t.Errorf("expected nil for empty edges (maxCard), got %v", m)
	}
}
