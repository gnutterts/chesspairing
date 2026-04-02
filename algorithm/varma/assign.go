package varma

import (
	"fmt"
	"sort"

	"github.com/gnutterts/chesspairing"
)

// Assign assigns pairing numbers to players using the Varma Tables scheme
// (FIDE C.05 Annex 2). Players from the same federation are spread across
// the 4 Varma groups to avoid same-federation clashes in early rounds.
//
// The returned slice is ordered by pairing number (index 0 = pairing number 1).
// Only active players are included. Inactive players are excluded from the result.
//
// Algorithm:
//  1. Filter to active players only.
//  2. Get the Varma group table for the active player count.
//  3. Group players by federation, sorted by federation size descending,
//     then alphabetically by federation code.
//  4. For each federation (largest first), pick the first Varma group (A→D)
//     with enough available slots. If no single group has enough, spill
//     across groups (largest-available first).
//  5. Within each federation, assign players to slots in alphabetical order
//     by DisplayName.
//
// Returns an error if the active player count is < 2 or > 24.
func Assign(players []chesspairing.PlayerEntry) ([]chesspairing.PlayerEntry, error) {
	// Step 1: filter to active players.
	var active []chesspairing.PlayerEntry
	for _, p := range players {
		if p.Active {
			active = append(active, p)
		}
	}

	n := len(active)
	if n < 2 || n > 24 {
		return nil, fmt.Errorf("varma: active player count %d out of supported range 2-24", n)
	}

	// Step 2: get group table.
	groups, err := Groups(n)
	if err != nil {
		return nil, err
	}

	// Build available slots per group (copy so we can consume them).
	available := [4][]int{}
	for i := 0; i < 4; i++ {
		available[i] = make([]int, len(groups[i].Numbers))
		copy(available[i], groups[i].Numbers)
	}

	// Step 3: group players by federation.
	type fedGroup struct {
		code    string
		players []chesspairing.PlayerEntry
	}
	fedMap := make(map[string]*fedGroup)
	for _, p := range active {
		code := p.Federation
		if code == "" {
			// Treat each player without federation as unique —
			// use a synthetic code that won't collide.
			code = "\x00_" + p.ID
		}
		fg, ok := fedMap[code]
		if !ok {
			fg = &fedGroup{code: code}
			fedMap[code] = fg
		}
		fg.players = append(fg.players, p)
	}

	// Sort federation groups: size descending, then alphabetically by code.
	feds := make([]*fedGroup, 0, len(fedMap))
	for _, fg := range fedMap {
		feds = append(feds, fg)
	}
	sort.Slice(feds, func(i, j int) bool {
		if len(feds[i].players) != len(feds[j].players) {
			return len(feds[i].players) > len(feds[j].players)
		}
		return feds[i].code < feds[j].code
	})

	// Sort players within each federation alphabetically by DisplayName.
	for _, fg := range feds {
		sort.Slice(fg.players, func(i, j int) bool {
			return fg.players[i].DisplayName < fg.players[j].DisplayName
		})
	}

	// Step 4: assign slots.
	// result maps pairing number (1-based) → player.
	result := make(map[int]chesspairing.PlayerEntry)

	for _, fg := range feds {
		needed := len(fg.players)

		// Try to find a single group with enough available slots.
		bestGroup := -1
		for gi := 0; gi < 4; gi++ {
			if len(available[gi]) >= needed {
				if bestGroup == -1 || len(available[gi]) < len(available[bestGroup]) {
					// Prefer the smallest group that still fits (best fit).
					bestGroup = gi
				}
			}
		}

		var slots []int
		if bestGroup >= 0 {
			// Take needed slots from this group.
			slots = available[bestGroup][:needed]
			available[bestGroup] = available[bestGroup][needed:]
		} else {
			// Spill: take from groups with most available slots first.
			remaining := needed
			for remaining > 0 {
				// Find group with most available slots.
				maxGroup := -1
				for gi := 0; gi < 4; gi++ {
					if len(available[gi]) > 0 {
						if maxGroup == -1 || len(available[gi]) > len(available[maxGroup]) {
							maxGroup = gi
						}
					}
				}
				if maxGroup == -1 {
					return nil, fmt.Errorf("varma: ran out of slots assigning federation %s", fg.code)
				}
				take := remaining
				if take > len(available[maxGroup]) {
					take = len(available[maxGroup])
				}
				slots = append(slots, available[maxGroup][:take]...)
				available[maxGroup] = available[maxGroup][take:]
				remaining -= take
			}
			// Sort the collected slots ascending so players are assigned
			// in pairing-number order.
			sort.Ints(slots)
		}

		// Step 5: assign players to slots in alphabetical order.
		for i, p := range fg.players {
			result[slots[i]] = p
		}
	}

	// Build the output slice ordered by pairing number.
	ordered := make([]chesspairing.PlayerEntry, 0, n)
	for num := 1; num <= n; num++ {
		p, ok := result[num]
		if !ok {
			// For odd counts, some numbers may have been filtered out by Groups().
			// But all n numbers should be assigned. This is a logic error.
			return nil, fmt.Errorf("varma: pairing number %d not assigned", num)
		}
		ordered = append(ordered, p)
	}

	return ordered, nil
}
