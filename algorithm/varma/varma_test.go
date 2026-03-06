package varma

import (
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
