// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package chesspairing

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

func titleRank(title string) int {
	switch strings.ToUpper(strings.TrimSpace(title)) {
	case "GM":
		return 1
	case "IM":
		return 2
	case "WGM":
		return 3
	case "FM":
		return 4
	case "WIM":
		return 5
	case "CM":
		return 6
	case "WFM":
		return 7
	case "WCM":
		return 8
	default:
		return 9
	}
}

// AssignPairingNumbers returns a copy of players with 1-based FIDE Tournament
// Pairing Numbers assigned according to C.04.2 article 2. Numbers are provisional
// under article 2.5: late entries renumber participants ranked below them.
// Existing complete assignments are preserved.
func AssignPairingNumbers(players []PlayerEntry) ([]PlayerEntry, error) {
	type sortable struct {
		entry PlayerEntry
		index int
	}

	less := func(a, b sortable) bool {
		if a.entry.Rating != b.entry.Rating {
			return a.entry.Rating > b.entry.Rating
		}
		ta, tb := titleRank(a.entry.Title), titleRank(b.entry.Title)
		if ta != tb {
			return ta < tb
		}
		if a.entry.DisplayName != b.entry.DisplayName {
			return a.entry.DisplayName < b.entry.DisplayName
		}
		return a.index < b.index
	}

	countSet := 0
	seen := make(map[int]bool, len(players))
	for _, p := range players {
		if p.PairingNumber < 0 {
			return nil, fmt.Errorf("pairing number must be positive: %d", p.PairingNumber)
		}
		if p.PairingNumber > 0 {
			countSet++
			if seen[p.PairingNumber] {
				return nil, fmt.Errorf("duplicate pairing number %d", p.PairingNumber)
			}
			seen[p.PairingNumber] = true
		}
	}

	out := append([]PlayerEntry(nil), players...)
	if countSet == len(players) {
		return out, nil
	}

	if countSet > 0 {
		var lateEntries []sortable
		for i, p := range players {
			if p.PairingNumber == 0 && p.JoinedRound <= 1 {
				return nil, errors.New("pairing numbers partially set")
			}
			if p.PairingNumber == 0 {
				lateEntries = append(lateEntries, sortable{entry: p, index: i})
			}
		}
		sort.SliceStable(lateEntries, func(i, j int) bool { return less(lateEntries[i], lateEntries[j]) })
		for _, late := range lateEntries {
			position := 1
			for _, p := range out {
				if p.PairingNumber == 0 {
					continue
				}
				if p.PairingNumber >= position {
					position = p.PairingNumber + 1
				}
			}
			for i, p := range out {
				if p.PairingNumber > 0 && less(late, sortable{entry: p, index: i}) && p.PairingNumber < position {
					position = p.PairingNumber
				}
			}
			for i := range out {
				if out[i].PairingNumber >= position {
					out[i].PairingNumber++
				}
			}
			out[late.index].PairingNumber = position
		}
		return out, nil
	}

	sorted := make([]sortable, len(players))
	for i, p := range players {
		sorted[i] = sortable{entry: p, index: i}
	}

	sort.SliceStable(sorted, func(i, j int) bool { return less(sorted[i], sorted[j]) })
	for i, s := range sorted {
		p := s.entry
		p.PairingNumber = i + 1
		out[s.index] = p
	}

	return out, nil
}
