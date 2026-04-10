// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package varma

import (
	"fmt"
	"testing"
)

func TestGroupsTableExhaustivenessAndUniqueness(t *testing.T) {
	// For every supported size (9-24), verify:
	// 1. All numbers 1..n appear exactly once across all 4 groups
	// 2. Group count is exactly 4
	for n := 9; n <= 24; n++ {
		groups, err := Groups(n)
		if err != nil {
			t.Fatalf("Groups(%d) error: %v", n, err)
		}

		seen := make(map[int]bool)
		total := 0
		for gi, g := range groups {
			for _, num := range g.Numbers {
				if num < 1 || num > n {
					t.Errorf("Groups(%d): group %c has out-of-range number %d", n, 'A'+gi, num)
				}
				if seen[num] {
					t.Errorf("Groups(%d): number %d appears in multiple groups", n, num)
				}
				seen[num] = true
				total++
			}
		}
		if total != n {
			t.Errorf("Groups(%d): total numbers = %d, want %d", n, total, n)
		}
		for i := 1; i <= n; i++ {
			if !seen[i] {
				t.Errorf("Groups(%d): missing number %d", n, i)
			}
		}
	}
}

func TestGroupsOddPlayerCount(t *testing.T) {
	// Odd player counts use the next-even table with highest number removed.
	// Verify the highest number (n+1) does NOT appear.
	for _, n := range []int{9, 11, 13, 15, 17, 19, 21, 23} {
		groups, err := Groups(n)
		if err != nil {
			t.Fatalf("Groups(%d) error: %v", n, err)
		}
		for gi, g := range groups {
			for _, num := range g.Numbers {
				if num == n+1 {
					t.Errorf("Groups(%d): group %c contains bye-dummy number %d", n, 'A'+gi, num)
				}
			}
		}
	}
}

func TestGroupsSpecificTable10(t *testing.T) {
	// FIDE C.05 Annex 2: 10-player table
	// A(3,4,8) B(5,7,9) C(1,6) D(2,10)
	groups, err := Groups(10)
	if err != nil {
		t.Fatal(err)
	}
	wantA := []int{3, 4, 8}
	wantB := []int{5, 7, 9}
	wantC := []int{1, 6}
	wantD := []int{2, 10}

	assertGroupNumbers(t, "10-player A", groups[0], wantA)
	assertGroupNumbers(t, "10-player B", groups[1], wantB)
	assertGroupNumbers(t, "10-player C", groups[2], wantC)
	assertGroupNumbers(t, "10-player D", groups[3], wantD)
}

func TestGroupsSpecificTable24(t *testing.T) {
	// FIDE C.05 Annex 2: 24-player table
	// A(6,7,8,9,10,11,19,20,21,22) B(1,2,3,4,13,14,15,16,17) C(12,18,23) D(5,24)
	groups, err := Groups(24)
	if err != nil {
		t.Fatal(err)
	}
	wantA := []int{6, 7, 8, 9, 10, 11, 19, 20, 21, 22}
	wantB := []int{1, 2, 3, 4, 13, 14, 15, 16, 17}
	wantC := []int{12, 18, 23}
	wantD := []int{5, 24}

	assertGroupNumbers(t, "24-player A", groups[0], wantA)
	assertGroupNumbers(t, "24-player B", groups[1], wantB)
	assertGroupNumbers(t, "24-player C", groups[2], wantC)
	assertGroupNumbers(t, "24-player D", groups[3], wantD)
}

func TestGroupsSmallCounts(t *testing.T) {
	// For 2-8 players, Groups returns trivially — each player gets
	// their own sequential number. All in one group (no separation needed).
	for _, n := range []int{2, 3, 4, 5, 6, 7, 8} {
		groups, err := Groups(n)
		if err != nil {
			t.Fatalf("Groups(%d) error: %v", n, err)
		}
		total := 0
		for _, g := range groups {
			total += len(g.Numbers)
		}
		if total != n {
			t.Errorf("Groups(%d): total = %d, want %d", n, total, n)
		}
	}
}

func TestGroupsInvalidCounts(t *testing.T) {
	for _, n := range []int{0, 1, -1, 25, 100} {
		_, err := Groups(n)
		if err == nil {
			t.Errorf("Groups(%d) should return an error", n)
		}
	}
}

func TestGroupsAllFIDETables(t *testing.T) {
	// Test all 8 even sizes from the FIDE C.05 Annex 2 Varma tables.
	type fideTable struct {
		n      int
		groups [4][]int // A, B, C, D
	}

	tables := []fideTable{
		{n: 10, groups: [4][]int{{3, 4, 8}, {5, 7, 9}, {1, 6}, {2, 10}}},
		{n: 12, groups: [4][]int{{4, 5, 9, 10}, {1, 2, 7}, {6, 8, 12}, {3, 11}}},
		{n: 14, groups: [4][]int{{4, 5, 6, 11, 12}, {1, 2, 8, 9}, {7, 10, 13}, {3, 14}}},
		{n: 16, groups: [4][]int{{5, 6, 7, 12, 13, 14}, {1, 2, 3, 9, 10}, {8, 11, 15}, {4, 16}}},
		{n: 18, groups: [4][]int{{5, 6, 7, 8, 14, 15, 16}, {1, 2, 3, 10, 11, 12}, {9, 13, 17}, {4, 18}}},
		{n: 20, groups: [4][]int{{6, 7, 8, 9, 15, 16, 17, 18}, {1, 2, 3, 11, 12, 13, 14}, {5, 10, 19}, {4, 20}}},
		{n: 22, groups: [4][]int{{6, 7, 8, 9, 10, 17, 18, 19, 20}, {1, 2, 3, 4, 12, 13, 14, 15}, {11, 16, 21}, {5, 22}}},
		{n: 24, groups: [4][]int{{6, 7, 8, 9, 10, 11, 19, 20, 21, 22}, {1, 2, 3, 4, 13, 14, 15, 16, 17}, {12, 18, 23}, {5, 24}}},
	}

	labels := [4]string{"A", "B", "C", "D"}

	for _, tt := range tables {
		t.Run(fmt.Sprintf("N=%d", tt.n), func(t *testing.T) {
			groups, err := Groups(tt.n)
			if err != nil {
				t.Fatalf("Groups(%d) error: %v", tt.n, err)
			}

			if len(groups) != 4 {
				t.Fatalf("Groups(%d) returned %d groups, want 4", tt.n, len(groups))
			}

			for i := 0; i < 4; i++ {
				assertGroupNumbers(t, fmt.Sprintf("N=%d group %s", tt.n, labels[i]), groups[i], tt.groups[i])
			}
		})
	}
}

func assertGroupNumbers(t *testing.T, label string, g Group, want []int) {
	t.Helper()
	if len(g.Numbers) != len(want) {
		t.Errorf("%s: got %v, want %v", label, g.Numbers, want)
		return
	}
	for i, num := range g.Numbers {
		if num != want[i] {
			t.Errorf("%s: index %d = %d, want %d (got %v, want %v)", label, i, num, want[i], g.Numbers, want)
			return
		}
	}
}

func TestGroupsSizePatterns(t *testing.T) {
	for n := 10; n <= 24; n += 2 {
		groups, err := Groups(n)
		if err != nil {
			t.Fatalf("Groups(%d): %v", n, err)
		}

		// Group D always has exactly 2 members for even counts >= 10.
		if len(groups[3].Numbers) != 2 {
			t.Errorf("Groups(%d): Group D has %d members, want 2", n, len(groups[3].Numbers))
		}

		// Group C always has exactly 3 members for even counts >= 12.
		// For n == 10, Group C has 2 members.
		if n >= 12 && len(groups[2].Numbers) != 3 {
			t.Errorf("Groups(%d): Group C has %d members, want 3", n, len(groups[2].Numbers))
		}
		if n == 10 && len(groups[2].Numbers) != 2 {
			t.Errorf("Groups(%d): Group C has %d members, want 2", n, len(groups[2].Numbers))
		}

		// Total across all groups must equal n.
		total := 0
		for _, g := range groups {
			total += len(g.Numbers)
		}
		if total != n {
			t.Errorf("Groups(%d): total members %d != %d", n, total, n)
		}
	}
}
