// Package varma implements the Varma Tables pre-processing number assignment
// scheme for round-robin chess tournaments (FIDE C.05 Annex 2).
//
// Varma Tables assign pairing numbers to players so that players from the
// same federation (country) are spread across the tournament schedule,
// avoiding same-federation clashes in early rounds.
//
// The scheme covers tournaments of 2-24 players. For 2-8 players the
// tables are trivial (sequential assignment). For 9-24 players, FIDE
// provides specific group assignments (A, B, C, D) that achieve optimal
// federation separation in the Berger table rotation.
//
// Usage: call Groups(n) to get the 4 group assignments for n players,
// then call Assign(players) to assign pairing numbers based on federation.
package varma

import "fmt"

// Group represents one of the four Varma groups (A, B, C, D).
// Numbers are the pairing number slots available in this group,
// sorted in ascending order.
type Group struct {
	Label   byte  // 'A', 'B', 'C', or 'D'
	Numbers []int // pairing number slots, 1-based, sorted ascending
}

// varmaTable holds the raw FIDE lookup table for even player counts 10-24.
// Key: even player count. Value: 4 slices of pairing numbers (A, B, C, D).
var varmaTable = map[int][4][]int{
	10: {
		{3, 4, 8},
		{5, 7, 9},
		{1, 6},
		{2, 10},
	},
	12: {
		{4, 5, 9, 10},
		{1, 2, 7},
		{6, 8, 12},
		{3, 11},
	},
	14: {
		{4, 5, 6, 11, 12},
		{1, 2, 8, 9},
		{7, 10, 13},
		{3, 14},
	},
	16: {
		{5, 6, 7, 12, 13, 14},
		{1, 2, 3, 9, 10},
		{8, 11, 15},
		{4, 16},
	},
	18: {
		{5, 6, 7, 8, 14, 15, 16},
		{1, 2, 3, 10, 11, 12},
		{9, 13, 17},
		{4, 18},
	},
	20: {
		{6, 7, 8, 9, 15, 16, 17, 18},
		{1, 2, 3, 11, 12, 13, 14},
		{5, 10, 19},
		{4, 20},
	},
	22: {
		{6, 7, 8, 9, 10, 17, 18, 19, 20},
		{1, 2, 3, 4, 12, 13, 14, 15},
		{11, 16, 21},
		{5, 22},
	},
	24: {
		{6, 7, 8, 9, 10, 11, 19, 20, 21, 22},
		{1, 2, 3, 4, 13, 14, 15, 16, 17},
		{12, 18, 23},
		{5, 24},
	},
}

// Groups returns the four Varma group assignments for n players.
//
// For 2-8 players, all numbers are placed in group A (trivial — no
// federation separation benefit with fewer than 9 players).
//
// For 9-24 players, the FIDE C.05 Annex 2 lookup tables are used.
// Odd counts use the next-even table with the highest number removed
// (that number would be the bye dummy in round-robin).
//
// Returns an error for n < 2 or n > 24.
func Groups(n int) ([4]Group, error) {
	if n < 2 || n > 24 {
		return [4]Group{}, fmt.Errorf("varma: player count %d out of supported range 2-24", n)
	}

	labels := [4]byte{'A', 'B', 'C', 'D'}

	// Trivial range: 2-8 players — all in group A.
	if n <= 8 {
		nums := make([]int, n)
		for i := range nums {
			nums[i] = i + 1
		}
		var groups [4]Group
		groups[0] = Group{Label: labels[0], Numbers: nums}
		for i := 1; i < 4; i++ {
			groups[i] = Group{Label: labels[i], Numbers: nil}
		}
		return groups, nil
	}

	// Determine table key: round up to even.
	tableKey := n
	if n%2 == 1 {
		tableKey = n + 1
	}

	raw, ok := varmaTable[tableKey]
	if !ok {
		return [4]Group{}, fmt.Errorf("varma: no table for %d players (tableKey=%d)", n, tableKey)
	}

	// For odd counts, filter out the highest number (bye dummy = tableKey).
	exclude := -1
	if n%2 == 1 {
		exclude = tableKey
	}

	var groups [4]Group
	for i := 0; i < 4; i++ {
		var filtered []int
		for _, num := range raw[i] {
			if num != exclude {
				filtered = append(filtered, num)
			}
		}
		groups[i] = Group{Label: labels[i], Numbers: filtered}
	}

	return groups, nil
}
