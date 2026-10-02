// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// TestPair_TopscorerThresholdUsesPlayedRounds covers C.04.3 1.8: before the
// last round of a five-round tournament four rounds are played, so a player with
// 2.5 points has over 50% of the maximum possible score (4) and is a topscorer.
func TestPair_TopscorerThresholdUsesPlayedRounds(t *testing.T) {
	game := func(w, b string, r chesspairing.GameResult) chesspairing.GameData {
		return chesspairing.GameData{WhiteID: w, BlackID: b, Result: r}
	}
	bye := func(id string, kind chesspairing.ByeType) []chesspairing.ByeEntry {
		return []chesspairing.ByeEntry{{PlayerID: id, Type: kind}}
	}
	ratings := []int{2400, 2351, 2241, 2207, 2106}
	var players []chesspairing.PlayerEntry
	for i, rating := range ratings {
		players = append(players, chesspairing.PlayerEntry{ID: fmt.Sprintf("p%d", i+1), Rating: rating})
	}
	totalRounds := 5
	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					game("p1", "p3", chesspairing.ResultBlackWins),
					game("p2", "p4", chesspairing.ResultBlackWins),
				},
				Byes: bye("p5", chesspairing.ByeZero),
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					game("p1", "p5", chesspairing.ResultDraw),
					game("p4", "p3", chesspairing.ResultBlackWins),
				},
				Byes: bye("p2", chesspairing.ByePAB),
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					game("p3", "p1", chesspairing.ResultDraw),
				},
				Byes: []chesspairing.ByeEntry{
					{PlayerID: "p4", Type: chesspairing.ByeFullPoint},
					{PlayerID: "p2", Type: chesspairing.ByeZero},
					{PlayerID: "p5", Type: chesspairing.ByeZero},
				},
			},
			{
				Number: 4,
				Games: []chesspairing.GameData{
					game("p2", "p1", chesspairing.ResultWhiteWins),
					game("p3", "p4", chesspairing.ResultBlackWins),
				},
				Byes: bye("p5", chesspairing.ByeZero),
			},
		},
		CurrentRound: 5,
	}

	// p2 and p3 both prefer black, which only the topscorer exception (p3 has
	// 2.5 points) allows them to meet. A pairing without it would give the bye
	// to p3. Reference: bbpPairings 6.0 pairs 1-4 and 3-2 with the bye for p5.
	result, err := New(Options{TotalRounds: &totalRounds}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error: %v", err)
	}
	var pairs []string
	for _, p := range result.Pairings {
		ids := []string{p.WhiteID, p.BlackID}
		sort.Strings(ids)
		pairs = append(pairs, ids[0]+"-"+ids[1])
	}
	sort.Strings(pairs)
	if len(pairs) != 2 || pairs[0] != "p1-p4" || pairs[1] != "p2-p3" {
		t.Errorf("pairings = %v, want [p1-p4 p2-p3]", pairs)
	}
	if len(result.Byes) != 1 || result.Byes[0].PlayerID != "p5" {
		t.Errorf("byes = %v, want one bye for p5", result.Byes)
	}
}
