// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"context"
	"errors"
	"fmt"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// ErrTooFewPlayers is returned when there aren't enough active players.
var ErrTooFewPlayers = errors.New("burstein pairing requires at least 2 active players")

// ErrNoPairingPossible is returned when no valid pairing can be found.
var ErrNoPairingPossible = errors.New("no valid pairing exists for the remaining players")

// Pair generates pairings for the next round using the Burstein system (C.04.4.2).
//
// Algorithm:
//  1. Build PlayerState for all active players
//  2. Determine if this is a seeding round or post-seeding round
//  3. Seeding rounds: use TPN-based ranking
//  4. Post-seeding rounds: re-rank by opposition index
//  5. Build score groups (all players enter matching pool)
//  6. Global Blossom matching (PairBracketsGlobal) — includes Stage 0.5
//     completability pre-matching for bye determination with odd player count
//  7. AllocateColor with topScorerRules=false; unmatched player receives PAB
func (p *Pairer) Pair(_ context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	// Honour pre-assigned byes for the upcoming round.
	state, preAssignedByes := swisslib.FilterPreAssignedByes(state)

	// Build player states.
	players := swisslib.BuildPlayerStates(state)

	if len(players) == 0 {
		if len(preAssignedByes) > 0 {
			return &chesspairing.PairingResult{Byes: preAssignedByes}, nil
		}
		return nil, ErrTooFewPlayers
	}

	var notes []string

	// Handle single player.
	if len(players) == 1 {
		byes := append([]chesspairing.ByeEntry{}, preAssignedByes...)
		byes = append(byes, chesspairing.ByeEntry{PlayerID: players[0].ID, Type: chesspairing.ByePAB})
		return &chesspairing.PairingResult{
			Byes:  byes,
			Notes: []string{players[0].ID + " receives a bye (only active player)"},
		}, nil
	}

	// Determine total rounds for seeding calculation.
	totalRounds := p.totalRounds(state)
	isSeeding := IsSeedingRound(state.CurrentRound, totalRounds)

	if isSeeding {
		notes = append(notes, fmt.Sprintf("Seeding round %d of %d", state.CurrentRound, SeedingRounds(totalRounds)))
	} else {
		notes = append(notes, fmt.Sprintf("Post-seeding round %d (opposition index ranking)", state.CurrentRound))
		// Re-rank players by opposition index.
		players = RankByOppositionIndex(players, state)
	}

	// Build active player pointers — ALL active players enter the matching pool.
	// If odd count, the Blossom matching will leave one player unmatched;
	// that player receives the PAB. This matches the FIDE algorithm where
	// the bye emerges from bracket processing, not pre-assignment.
	activePlayers := make([]*swisslib.PlayerState, len(players))
	for i := range players {
		activePlayers[i] = &players[i]
	}

	// Build player states slice for BuildScoreGroups.
	playerStates := make([]swisslib.PlayerState, len(activePlayers))
	for i, ap := range activePlayers {
		playerStates[i] = *ap
	}

	// Apply Baku acceleration if configured.
	if p.opts.Acceleration != nil && *p.opts.Acceleration == "baku" {
		gaSize := swisslib.BakuGASize(len(state.Players))
		swisslib.ApplyBakuAcceleration(playerStates, state.CurrentRound, totalRounds, gaSize)
		// Also update the pointer-based activePlayers to reflect PairingScore.
		for i := range activePlayers {
			activePlayers[i].PairingScore = playerStates[i].PairingScore
		}
		notes = append(notes, fmt.Sprintf("Baku acceleration: GA=%d players, VP=%.1f",
			gaSize, swisslib.BakuVirtualPoints(totalRounds, state.CurrentRound, true)))
	}

	// Build score groups.
	scoreGroups := swisslib.BuildScoreGroups(playerStates)

	// Build criteria context.
	playerMap := make(map[string]*swisslib.PlayerState, len(activePlayers))
	for _, ap := range activePlayers {
		playerMap[ap.ID] = ap
	}

	critCtx := &swisslib.CriteriaContext{
		Players:        playerMap,
		TotalRounds:    totalRounds,
		CurrentRound:   state.CurrentRound,
		IsLastRound:    state.CurrentRound == totalRounds,
		TopScorers:     map[string]bool{}, // Burstein: no topscorer rules
		ForbiddenPairs: buildForbiddenPairSet(p.opts.ForbiddenPairs),
	}

	// Burstein uses only color criteria C10-C13.
	// No look-ahead (C8) and no float criteria (C14-C21).
	// Note: PairBracketsGlobal currently uses all criteria via ComputeBaseEdgeWeight.
	// The float criteria provide additional optimization but don't change correctness.

	// Global Blossom matching — same architecture as Dutch.
	// Replaces the broken bracket-by-bracket approach with global matching
	// that considers all players simultaneously.
	allPairs, unmatchedPlayer, pairNotes := swisslib.PairBracketsGlobal(scoreGroups, critCtx, playerMap)
	notes = append(notes, pairNotes...)

	// Allocate colors and build final pairings.
	// Burstein: topScorerRules=false.
	topSeedColor := parseTopSeedColor(p.opts.TopSeedColor)
	pairings := make([]chesspairing.GamePairing, len(allPairs))
	for i, pair := range allPairs {
		whiteID, blackID := swisslib.AllocateColor(pair.White, pair.Black, false, i+1, topSeedColor)
		pairings[i] = chesspairing.GamePairing{
			Board:   i + 1,
			WhiteID: whiteID,
			BlackID: blackID,
		}
	}

	// Build result.
	result := &chesspairing.PairingResult{
		Pairings: pairings,
		Notes:    notes,
	}

	if len(preAssignedByes) > 0 {
		result.Byes = append(result.Byes, preAssignedByes...)
	}

	if unmatchedPlayer != nil {
		result.Byes = append(result.Byes, chesspairing.ByeEntry{PlayerID: unmatchedPlayer.ID, Type: chesspairing.ByePAB})
		result.Notes = append(result.Notes, fmt.Sprintf("%s receives PAB (bye)", unmatchedPlayer.ID))
	}

	result.Notes = append(result.Notes, "Pairings generated by Burstein Swiss system (C.04.4.2)")

	return result, nil
}

// parseTopSeedColor converts the TopSeedColor string option to a *swisslib.Color.
// Returns nil for "auto" or "white" (default behavior), and &ColorBlack for "black".
func parseTopSeedColor(opt *string) *swisslib.Color {
	if opt == nil || *opt == "auto" || *opt == "white" {
		return nil
	}
	if *opt == "black" {
		c := swisslib.ColorBlack
		return &c
	}
	return nil
}

// buildForbiddenPairSet converts the options ForbiddenPairs slice into
// the canonicalized map format used by CriteriaContext.
func buildForbiddenPairSet(pairs [][]string) map[[2]string]bool {
	if len(pairs) == 0 {
		return nil
	}
	m := make(map[[2]string]bool, len(pairs))
	for _, pair := range pairs {
		if len(pair) == 2 {
			m[swisslib.CanonicalPairKey(pair[0], pair[1])] = true
		}
	}
	return m
}

// totalRounds returns the total number of rounds for seeding calculation.
// Uses options override if set, otherwise derives from state.
func (p *Pairer) totalRounds(state *chesspairing.TournamentState) int {
	if p.opts.TotalRounds != nil {
		return *p.opts.TotalRounds
	}

	// Derive from state: use CurrentRound as best estimate if larger
	// than completed rounds.
	total := state.CurrentRound
	if total < len(state.Rounds)+1 {
		total = len(state.Rounds) + 1
	}
	return total
}
