// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package varma

import (
	"fmt"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestAssignBasicFederationSeparation(t *testing.T) {
	// 10 players, 3 federations: NED(4), USA(3), IND(3).
	// Federations sorted by size desc, then alphabetically: IND(3), NED(4), USA(3).
	// Wait — NED has 4, so: NED(4) first, then IND(3) vs USA(3) alphabetically: IND(3), USA(3).
	// Sorted: NED(4), IND(3), USA(3).
	//
	// Group capacities for 10 players: A(3), B(3), C(2), D(2).
	//
	// NED(4): needs 4 slots. No single group has 4. Spill: A has 3 → take 3 from A, then
	// next largest available has 3 (B) → take 1 from B. NED gets A(3)+B(1)=4.
	// IND(3): needs 3. B has 2 remaining. C has 2. D has 2. No single group has 3.
	// Spill: B(2) + C(2) → take 2 from B, 1 from C. IND gets B(2)+C(1)=3.
	// USA(3): needs 3. C has 1 remaining. D has 2. Spill: D(2) + C(1) → take 2 from D, 1 from C.
	// USA gets C(1)+D(2)=3.
	//
	// Verify: all 10 numbers assigned, no duplicates, all players present.
	players := []chesspairing.PlayerEntry{
		{ID: "n1", DisplayName: "Bakker", Rating: 2400, Active: true, Federation: "NED"},
		{ID: "n2", DisplayName: "De Vries", Rating: 2350, Active: true, Federation: "NED"},
		{ID: "n3", DisplayName: "Jansen", Rating: 2300, Active: true, Federation: "NED"},
		{ID: "n4", DisplayName: "Van Dijk", Rating: 2200, Active: true, Federation: "NED"},
		{ID: "i1", DisplayName: "Anand", Rating: 2500, Active: true, Federation: "IND"},
		{ID: "i2", DisplayName: "Harikrishna", Rating: 2450, Active: true, Federation: "IND"},
		{ID: "i3", DisplayName: "Vidit", Rating: 2380, Active: true, Federation: "IND"},
		{ID: "u1", DisplayName: "Caruana", Rating: 2480, Active: true, Federation: "USA"},
		{ID: "u2", DisplayName: "Nakamura", Rating: 2460, Active: true, Federation: "USA"},
		{ID: "u3", DisplayName: "So", Rating: 2420, Active: true, Federation: "USA"},
	}

	result, err := Assign(players)
	if err != nil {
		t.Fatalf("Assign() error: %v", err)
	}

	if len(result) != 10 {
		t.Fatalf("expected 10 players, got %d", len(result))
	}

	// Verify all IDs are present (no duplicates, no missing).
	idSet := make(map[string]bool)
	for _, p := range result {
		if idSet[p.ID] {
			t.Errorf("duplicate player ID: %s", p.ID)
		}
		idSet[p.ID] = true
	}
	for _, p := range players {
		if !idSet[p.ID] {
			t.Errorf("missing player ID: %s", p.ID)
		}
	}

	// Verify federation separation: for each federation with 2+ players,
	// the assigned pairing numbers must span at least 2 different Varma groups.
	groups, err := Groups(10)
	if err != nil {
		t.Fatalf("Groups(10) error: %v", err)
	}

	// Build reverse map: pairing number (1-based) → group index.
	numToGroup := make(map[int]int)
	for gi, g := range groups {
		for _, num := range g.Numbers {
			numToGroup[num] = gi
		}
	}

	// Collect group indices per federation.
	fedGroups := make(map[string]map[int]bool)
	for i, p := range result {
		pairingNum := i + 1 // result is ordered by pairing number
		gi := numToGroup[pairingNum]
		if fedGroups[p.Federation] == nil {
			fedGroups[p.Federation] = make(map[int]bool)
		}
		fedGroups[p.Federation][gi] = true
	}

	for fed, groupSet := range fedGroups {
		if len(groupSet) < 2 {
			t.Errorf("federation %s: all players in a single Varma group (no separation)", fed)
		}
	}
}

func TestAssignSingleFederation(t *testing.T) {
	// All players same federation — they all go to group A first, spill into others.
	// The algorithm should still complete without error.
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alpha", Rating: 2000, Active: true, Federation: "NED"},
		{ID: "p2", DisplayName: "Bravo", Rating: 1900, Active: true, Federation: "NED"},
		{ID: "p3", DisplayName: "Charlie", Rating: 1800, Active: true, Federation: "NED"},
		{ID: "p4", DisplayName: "Delta", Rating: 1700, Active: true, Federation: "NED"},
		{ID: "p5", DisplayName: "Echo", Rating: 1600, Active: true, Federation: "NED"},
		{ID: "p6", DisplayName: "Foxtrot", Rating: 1500, Active: true, Federation: "NED"},
		{ID: "p7", DisplayName: "Golf", Rating: 1400, Active: true, Federation: "NED"},
		{ID: "p8", DisplayName: "Hotel", Rating: 1300, Active: true, Federation: "NED"},
		{ID: "p9", DisplayName: "India", Rating: 1200, Active: true, Federation: "NED"},
		{ID: "p10", DisplayName: "Juliet", Rating: 1100, Active: true, Federation: "NED"},
	}

	result, err := Assign(players)
	if err != nil {
		t.Fatalf("Assign() error: %v", err)
	}

	if len(result) != 10 {
		t.Fatalf("expected 10 players, got %d", len(result))
	}

	// Verify all IDs are present.
	idSet := make(map[string]bool)
	for _, p := range result {
		if idSet[p.ID] {
			t.Errorf("duplicate player ID: %s", p.ID)
		}
		idSet[p.ID] = true
	}
	for _, p := range players {
		if !idSet[p.ID] {
			t.Errorf("missing player ID: %s", p.ID)
		}
	}

	// All players share one federation, so within the result they should
	// appear in alphabetical order by DisplayName (single federation =
	// no inter-federation interleaving).
	for i := 1; i < len(result); i++ {
		if result[i].DisplayName < result[i-1].DisplayName {
			t.Errorf("players not in alphabetical order: result[%d]=%s comes after result[%d]=%s",
				i, result[i].DisplayName, i-1, result[i-1].DisplayName)
			break
		}
	}
}

func TestAssignNoFederation(t *testing.T) {
	// Players with empty federation — each treated as unique, all assigned sequentially.
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alpha", Rating: 2000, Active: true},
		{ID: "p2", DisplayName: "Bravo", Rating: 1900, Active: true},
		{ID: "p3", DisplayName: "Charlie", Rating: 1800, Active: true},
		{ID: "p4", DisplayName: "Delta", Rating: 1700, Active: true},
		{ID: "p5", DisplayName: "Echo", Rating: 1600, Active: true},
		{ID: "p6", DisplayName: "Foxtrot", Rating: 1500, Active: true},
		{ID: "p7", DisplayName: "Golf", Rating: 1400, Active: true},
		{ID: "p8", DisplayName: "Hotel", Rating: 1300, Active: true},
		{ID: "p9", DisplayName: "India", Rating: 1200, Active: true},
		{ID: "p10", DisplayName: "Juliet", Rating: 1100, Active: true},
	}

	result, err := Assign(players)
	if err != nil {
		t.Fatalf("Assign() error: %v", err)
	}

	if len(result) != 10 {
		t.Fatalf("expected 10 players, got %d", len(result))
	}

	// Verify all IDs are present.
	idSet := make(map[string]bool)
	for _, p := range result {
		if idSet[p.ID] {
			t.Errorf("duplicate player ID: %s", p.ID)
		}
		idSet[p.ID] = true
	}
	for _, p := range players {
		if !idSet[p.ID] {
			t.Errorf("missing player ID: %s", p.ID)
		}
	}
}

func TestAssignOddPlayerCount(t *testing.T) {
	// 9 players: uses 10-player table with number 10 removed.
	players := make([]chesspairing.PlayerEntry, 9)
	feds := []string{"NED", "NED", "NED", "USA", "USA", "USA", "IND", "IND", "IND"}
	for i := range players {
		players[i] = chesspairing.PlayerEntry{
			ID:          fmt.Sprintf("p%d", i+1),
			DisplayName: fmt.Sprintf("Player%d", i+1),
			Rating:      2000 - i*100,
			Active:      true,
			Federation:  feds[i],
		}
	}

	result, err := Assign(players)
	if err != nil {
		t.Fatalf("Assign() error: %v", err)
	}

	if len(result) != 9 {
		t.Fatalf("expected 9 players, got %d", len(result))
	}
}

func TestAssignAlphabeticalOrderWithinFederation(t *testing.T) {
	// Players within a federation should be assigned in alphabetical order by DisplayName.
	// Build 10 players with 2 federations to keep it simple.
	players := []chesspairing.PlayerEntry{
		{ID: "z1", DisplayName: "Zebra", Rating: 2000, Active: true, Federation: "NED"},
		{ID: "a1", DisplayName: "Alpha", Rating: 2100, Active: true, Federation: "NED"},
		{ID: "m1", DisplayName: "Mike", Rating: 1900, Active: true, Federation: "NED"},
		{ID: "b1", DisplayName: "Bravo", Rating: 2050, Active: true, Federation: "NED"},
		{ID: "c1", DisplayName: "Charlie", Rating: 1950, Active: true, Federation: "NED"},
		{ID: "x1", DisplayName: "Zeta", Rating: 2200, Active: true, Federation: "USA"},
		{ID: "d1", DisplayName: "Delta", Rating: 2150, Active: true, Federation: "USA"},
		{ID: "e1", DisplayName: "Echo", Rating: 2050, Active: true, Federation: "USA"},
		{ID: "f1", DisplayName: "Foxtrot", Rating: 1950, Active: true, Federation: "USA"},
		{ID: "g1", DisplayName: "Golf", Rating: 1850, Active: true, Federation: "USA"},
	}

	result, err := Assign(players)
	if err != nil {
		t.Fatalf("Assign() error: %v", err)
	}

	// Within each federation, players should appear in alphabetical order
	// by DisplayName within their assigned group slots.
	// NED has 5 players: Alpha, Bravo, Charlie, Mike, Zebra (alphabetical).
	// USA has 5 players: Delta, Echo, Foxtrot, Golf, Zeta (alphabetical).
	//
	// We verify that for each federation, the relative order in the result
	// slice (by index position = pairing number) is alphabetical by DisplayName.
	fedOrder := make(map[string][]string)
	for _, p := range result {
		fedOrder[p.Federation] = append(fedOrder[p.Federation], p.DisplayName)
	}

	for fed, names := range fedOrder {
		for i := 1; i < len(names); i++ {
			if names[i] < names[i-1] {
				t.Errorf("federation %s: players not in alphabetical order: %v", fed, names)
				break
			}
		}
	}
}

func TestAssignTooFewPlayers(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true, Federation: "NED"},
	}
	_, err := Assign(players)
	if err == nil {
		t.Error("Assign() should return error for < 2 players")
	}
}

func TestAssignTooManyPlayers(t *testing.T) {
	players := make([]chesspairing.PlayerEntry, 25)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{
			ID:          fmt.Sprintf("p%d", i+1),
			DisplayName: fmt.Sprintf("Player%d", i+1),
			Rating:      2000,
			Active:      true,
			Federation:  "NED",
		}
	}
	_, err := Assign(players)
	if err == nil {
		t.Error("Assign() should return error for > 24 players")
	}
}

func TestAssignInactivePlayersExcluded(t *testing.T) {
	// 10 entries but 2 are inactive → 8 active players (trivial range).
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alpha", Rating: 2000, Active: true, Federation: "NED"},
		{ID: "p2", DisplayName: "Bravo", Rating: 1900, Active: true, Federation: "NED"},
		{ID: "p3", DisplayName: "Charlie", Rating: 1800, Active: false, Federation: "NED"},
		{ID: "p4", DisplayName: "Delta", Rating: 1700, Active: true, Federation: "USA"},
		{ID: "p5", DisplayName: "Echo", Rating: 1600, Active: true, Federation: "USA"},
		{ID: "p6", DisplayName: "Foxtrot", Rating: 1500, Active: false, Federation: "USA"},
		{ID: "p7", DisplayName: "Golf", Rating: 1400, Active: true, Federation: "IND"},
		{ID: "p8", DisplayName: "Hotel", Rating: 1300, Active: true, Federation: "IND"},
		{ID: "p9", DisplayName: "India", Rating: 1200, Active: true, Federation: "IND"},
		{ID: "p10", DisplayName: "Juliet", Rating: 1100, Active: true, Federation: "IND"},
	}

	result, err := Assign(players)
	if err != nil {
		t.Fatalf("Assign() error: %v", err)
	}

	// Only active players should be in the result.
	if len(result) != 8 {
		t.Fatalf("expected 8 active players, got %d", len(result))
	}

	for _, p := range result {
		if !p.Active {
			t.Errorf("inactive player %s should not be in result", p.ID)
		}
	}
}

func TestAssignZeroActivePlayers(t *testing.T) {
	// All players inactive — should return an error.
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alpha", Rating: 2000, Active: false, Federation: "NED"},
		{ID: "p2", DisplayName: "Bravo", Rating: 1900, Active: false, Federation: "NED"},
		{ID: "p3", DisplayName: "Charlie", Rating: 1800, Active: false, Federation: "USA"},
	}

	_, err := Assign(players)
	if err == nil {
		t.Error("Assign() should return error when all players are inactive")
	}
}

func TestAssignMinimumTwoPlayers(t *testing.T) {
	// Exactly 2 active players — minimum valid count, should succeed.
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alpha", Rating: 2000, Active: true, Federation: "NED"},
		{ID: "p2", DisplayName: "Bravo", Rating: 1900, Active: true, Federation: "USA"},
	}

	result, err := Assign(players)
	if err != nil {
		t.Fatalf("Assign() error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 players, got %d", len(result))
	}

	// Verify both IDs are present.
	idSet := make(map[string]bool)
	for _, p := range result {
		idSet[p.ID] = true
	}
	for _, p := range players {
		if !idSet[p.ID] {
			t.Errorf("missing player ID: %s", p.ID)
		}
	}
}

func TestAssignFederationSeparationWithBergerTable(t *testing.T) {
	// 10 players, 5 federations of 2 each.
	players := []chesspairing.PlayerEntry{
		{ID: "a1", DisplayName: "Alpha One", Rating: 2500, Federation: "NED", Active: true},
		{ID: "a2", DisplayName: "Alpha Two", Rating: 2400, Federation: "NED", Active: true},
		{ID: "b1", DisplayName: "Beta One", Rating: 2300, Federation: "USA", Active: true},
		{ID: "b2", DisplayName: "Beta Two", Rating: 2200, Federation: "USA", Active: true},
		{ID: "c1", DisplayName: "Gamma One", Rating: 2100, Federation: "GER", Active: true},
		{ID: "c2", DisplayName: "Gamma Two", Rating: 2000, Federation: "GER", Active: true},
		{ID: "d1", DisplayName: "Delta One", Rating: 1900, Federation: "FRA", Active: true},
		{ID: "d2", DisplayName: "Delta Two", Rating: 1800, Federation: "FRA", Active: true},
		{ID: "e1", DisplayName: "Epsilon One", Rating: 1700, Federation: "ESP", Active: true},
		{ID: "e2", DisplayName: "Epsilon Two", Rating: 1600, Federation: "ESP", Active: true},
	}

	assigned, err := Assign(players)
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}

	// Build federation-by-pairing-number lookup.
	pnToFed := make(map[int]string)
	for i, p := range assigned {
		pnToFed[i+1] = p.Federation
	}

	// Standard Berger table: round 1 pairings for n=10 are:
	// (1,10), (2,9), (3,8), (4,7), (5,6)
	n := 10
	round1Pairs := make([][2]int, n/2)
	for i := 0; i < n/2; i++ {
		round1Pairs[i] = [2]int{i + 1, n - i}
	}

	// Check that no round-1 pair has same federation.
	for _, pair := range round1Pairs {
		fed1, fed2 := pnToFed[pair[0]], pnToFed[pair[1]]
		if fed1 == fed2 {
			t.Errorf("round 1: pairing (%d,%d) has same federation %q", pair[0], pair[1], fed1)
		}
	}

	for i, p := range assigned {
		t.Logf("PN %d: %s (fed=%s)", i+1, p.DisplayName, p.Federation)
	}
}

func TestAssignLargeFederationSpill(t *testing.T) {
	// 12 players: 1 federation with 7 players, 1 federation with 5.
	players := make([]chesspairing.PlayerEntry, 12)
	for i := 0; i < 7; i++ {
		players[i] = chesspairing.PlayerEntry{
			ID:          fmt.Sprintf("big%d", i+1),
			DisplayName: fmt.Sprintf("Big %02d", i+1),
			Rating:      2000 - i*50,
			Federation:  "BIG",
			Active:      true,
		}
	}
	for i := 0; i < 5; i++ {
		players[7+i] = chesspairing.PlayerEntry{
			ID:          fmt.Sprintf("small%d", i+1),
			DisplayName: fmt.Sprintf("Small %02d", i+1),
			Rating:      1700 - i*50,
			Federation:  "SML",
			Active:      true,
		}
	}

	assigned, err := Assign(players)
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}

	// Verify all 12 unique IDs present.
	idSet := make(map[string]bool)
	for _, p := range assigned {
		idSet[p.ID] = true
	}
	if len(idSet) != 12 {
		t.Errorf("got %d unique IDs, want 12", len(idSet))
	}

	// Verify the BIG federation spans at least 2 groups.
	bigGroups := make(map[byte]bool)
	groups, _ := Groups(12)
	pnToGroup := make(map[int]byte)
	for _, g := range groups {
		for _, num := range g.Numbers {
			pnToGroup[num] = g.Label
		}
	}
	for i, p := range assigned {
		if p.Federation == "BIG" {
			bigGroups[pnToGroup[i+1]] = true
		}
	}
	if len(bigGroups) < 2 {
		t.Errorf("BIG federation only in %d group(s), expected at least 2", len(bigGroups))
	}

	t.Logf("BIG federation spans groups: %v", bigGroups)
}
