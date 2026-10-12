// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestFootballTopScorerThreshold(t *testing.T) {
	players := []*swisslib.PlayerState{{ID: "draws"}, {ID: "leader"}, {ID: "half"}}
	scores := map[string]float64{"draws": 2, "leader": 4, "half": 3}
	got := computeTopScorers(players, 2, 3, scores)
	if !got["leader"] || got["draws"] || got["half"] || len(got) != 1 {
		t.Errorf("topscorers on 3-1-0 after two rounds = %v, want leader only", got)
	}
}
