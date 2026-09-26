// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing/pairing/dutch"
)

// reproTRFPlayer builds a minimal fixed-width 001 line of 89 characters.
// Only the start number (cols 5-8), name (cols 15-47) and rating
// (cols 49-52) are populated; all other fields stay blank.
func reproTRFPlayer(sn int, name string, rating int) string {
	line := fmt.Sprintf("001 %4d      %-33s %4d", sn, name, rating)
	for len(line) < 89 {
		line += " "
	}
	return line
}

// TestD3TopSeedColorB verifies that record 152 gives the top seed Black.
func TestD3TopSeedColorB(t *testing.T) {
	trfText := strings.Join([]string{
		"012 Repro D3",
		"092 Swiss Dutch",
		"142 1",
		"152 B",
		reproTRFPlayer(1, "Alpha", 2000),
		reproTRFPlayer(2, "Beta", 1500),
		"",
	}, "\n")

	doc, err := Read(strings.NewReader(trfText))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}

	if got := state.PairingConfig.Options["topSeedColor"]; got != "black" {
		t.Fatalf("topSeedColor = %q, want %q", got, "black")
	}

	pairer := dutch.NewFromMap(state.PairingConfig.Options)
	res, err := pairer.Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair: %v", err)
	}
	if len(res.Pairings) != 1 {
		t.Fatalf("number of pairings = %d, want 1", len(res.Pairings))
	}

	whiteID := res.Pairings[0].WhiteID
	blackID := res.Pairings[0].BlackID
	if whiteID != "2" || blackID != "1" {
		t.Fatalf("round 1 = %s-%s, want 2-1", whiteID, blackID)
	}
}
