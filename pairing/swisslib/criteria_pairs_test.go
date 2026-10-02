// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"context"
	"math/big"
	"testing"
)

func TestC9CountsForfeitsAsUnplayed(t *testing.T) {
	groups := []ScoreGroup{{Score: 0, Players: []*PlayerState{{ID: "forfeit"}, {ID: "played"}, {ID: "floater"}}}}
	params := ComputeEdgeWeightParams(groups, 1)
	params.C9Candidates = map[string]bool{"forfeit": true, "played": true}
	forfeit := &PlayerState{ID: "forfeit"}
	played := &PlayerState{ID: "played", ColorHistory: []Color{ColorWhite}}
	floater := &PlayerState{ID: "floater"}

	forfeitWeight := ComputeBaseEdgeWeight(floater, forfeit, true, false, true, &params)
	playedWeight := ComputeBaseEdgeWeight(floater, played, true, false, true, &params)
	if forfeitWeight.Cmp(playedWeight) <= 0 {
		t.Fatalf("C9 should pair the player with the forfeit (%s) and leave the player with a played game (%s) for the PAB", forfeitWeight, playedWeight)
	}

	candidateWeight := ComputeBaseEdgeWeight(forfeit, played, true, false, true, &params)
	if field := c9Field(candidateWeight, &params); field.Int64() != 1 {
		t.Errorf("candidate-candidate C9 field = %d, want 1 (forfeit 1 + played 0)", field.Int64())
	}
}

func TestLegacyByeGamesCountsByesNotForfeits(t *testing.T) {
	groups := []ScoreGroup{{Score: 0, Players: []*PlayerState{{ID: "absent"}, {ID: "forfeit"}, {ID: "floater"}}}}
	params := ComputeEdgeWeightParams(groups, 1)
	params.LegacyByeGames = true
	params.C9Bits = params.ScoreGroupSizeBits
	params.PABEligible = map[string]bool{"absent": true, "forfeit": true}
	absent := &PlayerState{ID: "absent", ColorHistory: []Color{ColorNone}}
	forfeit := &PlayerState{ID: "forfeit"}
	floater := &PlayerState{ID: "floater"}

	absentWeight := ComputeBaseEdgeWeight(floater, absent, true, false, false, &params)
	forfeitWeight := ComputeBaseEdgeWeight(floater, forfeit, true, false, false, &params)
	if absentWeight.Cmp(forfeitWeight) <= 0 {
		t.Fatalf("legacy field should count the absent round (%s) and not the forfeit (%s)", absentWeight, forfeitWeight)
	}
}

func TestC9AppliesWhenLowestBracketTakesOneDownfloater(t *testing.T) {
	players := []*PlayerState{
		{ID: "p1", Score: 1, PairingScore: 1, ColorHistory: []Color{ColorWhite, ColorBlack}},
		{ID: "p2", Score: 1, PairingScore: 1, ColorHistory: []Color{ColorBlack, ColorWhite}},
		{ID: "p3", Score: 1, PairingScore: 1, ColorHistory: []Color{ColorBlack, ColorWhite}},
		{ID: "p4", Score: 1, PairingScore: 1, ColorHistory: []Color{ColorWhite, ColorNone}},
		{ID: "l1", Score: 0, PairingScore: 0, PABIneligible: PABIneligibility{PriorPAB: true}},
	}
	groups := []ScoreGroup{
		{Score: 1, Players: players[:4]},
		{Score: 0, Players: players[4:]},
	}
	playerMap := make(map[string]*PlayerState, len(players))
	for _, player := range players {
		playerMap[player.ID] = player
	}
	cctx := &CriteriaContext{Players: playerMap, CurrentRound: 3, TotalRounds: 4}
	pairs, bye, _, err := PairBracketsGlobal(context.Background(), groups, cctx, MatchingCriteria{ApplyC9: true}, playerMap)
	if err != nil {
		t.Fatalf("PairBracketsGlobal: %v", err)
	}
	// p4 has one unplayed round and p3 has none, so C9 leaves p3 for the PAB
	// and sends p4 down to l1.
	if bye == nil || bye.ID != "p3" {
		t.Fatalf("PAB = %v, want p3 (no unplayed rounds); pairs = %v", bye, pairs)
	}
}

func TestC9SelectsPABInMiddleBracket(t *testing.T) {
	players := []*PlayerState{
		{ID: "top-1", Score: 2, PairingScore: 2},
		{ID: "top-2", Score: 2, PairingScore: 2},
		{ID: "middle-unplayed", Score: 1, PairingScore: 1},
		{ID: "middle-played", Score: 1, PairingScore: 1, ColorHistory: []Color{ColorWhite}},
		{ID: "middle-other", Score: 1, PairingScore: 1},
		{ID: "low-1", Score: 0, PairingScore: 0, PABIneligible: PABIneligibility{PriorPAB: true}},
		{ID: "low-2", Score: 0, PairingScore: 0, PABIneligible: PABIneligibility{PriorPAB: true}},
	}
	groups := []ScoreGroup{
		{Score: 2, Players: players[:2]},
		{Score: 1, Players: players[2:5]},
		{Score: 0, Players: players[5:]},
	}
	playerMap := make(map[string]*PlayerState, len(players))
	for _, player := range players {
		playerMap[player.ID] = player
	}
	cctx := &CriteriaContext{Players: playerMap, CurrentRound: 2, TotalRounds: 3}
	_, bye, _, err := PairBracketsGlobal(context.Background(), groups, cctx, MatchingCriteria{ApplyC9: true}, playerMap)
	if err != nil {
		t.Fatalf("PairBracketsGlobal: %v", err)
	}
	if bye == nil || bye.ID != "middle-played" {
		t.Fatalf("PAB = %v, want middle-played", bye)
	}
}

func TestC9FieldFitsNineRounds(t *testing.T) {
	groups := []ScoreGroup{{Score: 0, Players: []*PlayerState{{ID: "candidate"}, {ID: "opponent"}}}}
	params := ComputeEdgeWeightParams(groups, 9)
	params.C9Candidates = map[string]bool{"candidate": true}
	candidate := &PlayerState{ID: "candidate", ColorHistory: []Color{ColorWhite}}
	opponent := &PlayerState{ID: "opponent"}

	weight := ComputeBaseEdgeWeight(opponent, candidate, true, false, true, &params)
	field := c9Field(weight, &params)
	// Nine completed rounds minus one over-the-board game is eight unplayed rounds.
	if field.Int64() != 8 {
		t.Errorf("C9 field = %d, want 8", field.Int64())
	}
	if params.C9Bits != 5 {
		t.Errorf("C9 field width = %d, want 5", params.C9Bits)
	}
}

func c9Field(weight *big.Int, params *EdgeWeightParams) *big.Int {
	field := new(big.Int).Rsh(weight, uint(c9Shift(params)))
	return field.And(field, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(2*params.C9Bits)), big.NewInt(1)))
}

func c9Shift(params *EdgeWeightParams) int {
	c9Shift := params.ReserveBits + 4*params.ScoreGroupSizeBits
	if params.PlayedRounds > 0 {
		c9Shift += 2*params.ScoreGroupsShift + 2*params.ScoreGroupSizeBits
	}
	if params.PlayedRounds > 1 {
		c9Shift += 2*params.ScoreGroupsShift + 2*params.ScoreGroupSizeBits
	}
	return c9Shift
}
