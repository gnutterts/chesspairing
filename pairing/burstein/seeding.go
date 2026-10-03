// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// pairDutchSeeding implements C.04.4.2 Article 1.6 by applying the Dutch
// pairing procedure. It is kept local to avoid a package cycle with Dutch's
// compatibility tests, which also exercise Burstein.
func (p *Pairer) pairDutchSeeding(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	original := state
	if p.opts.TotalRounds != nil && *p.opts.TotalRounds < 1 {
		return nil, fmt.Errorf("dutch: total rounds must be at least 1")
	}
	if p.opts.TotalRounds != nil && state.CurrentRound > *p.opts.TotalRounds {
		return nil, fmt.Errorf("dutch: current round %d exceeds total rounds %d", state.CurrentRound, *p.opts.TotalRounds)
	}
	if p.opts.Acceleration != nil && *p.opts.Acceleration == "baku" && p.opts.TotalRounds == nil {
		return nil, fmt.Errorf("dutch: Baku acceleration requires total rounds")
	}
	totalRounds := 0
	if p.opts.TotalRounds != nil {
		totalRounds = *p.opts.TotalRounds
	}
	state, preAssigned := swisslib.FilterPreAssignedByes(state)
	var notes []string
	var virtualPoints func(string, int) float64
	if p.opts.Acceleration != nil && *p.opts.Acceleration == "baku" {
		groupA := swisslib.BakuGASize(len(original.Players))
		winPoints := standard.WinPoints(state.ScoringConfig.Options)
		numbers := make(map[string]int, len(original.Players))
		if numbered, err := chesspairing.AssignPairingNumbers(original.Players); err == nil {
			for _, player := range numbered {
				numbers[player.ID] = player.PairingNumber
			}
		}
		virtualPoints = func(id string, round int) float64 {
			number, ok := numbers[id]
			return swisslib.BakuVirtualPoints(winPoints, totalRounds, round, ok && number <= groupA)
		}
		notes = append(notes, fmt.Sprintf("Baku acceleration: GA=%d players, VP=%.1f", groupA, swisslib.BakuVirtualPoints(winPoints, totalRounds, state.CurrentRound, true)))
	}
	players, err := swisslib.BuildPlayerStatesWithVirtualPoints(state, virtualPoints)
	if err != nil {
		return nil, err
	}
	if virtualPoints != nil {
		swisslib.AddVirtualPoints(players, func(id string) float64 { return virtualPoints(id, state.CurrentRound) })
	}
	if len(players) == 0 {
		if len(preAssigned) == 0 {
			return nil, ErrTooFewPlayers
		}
		return &chesspairing.PairingResult{Byes: preAssigned}, nil
	}
	if len(players) == 1 {
		if players[0].PABIneligible.Any() {
			return nil, noPABError()
		}
		return &chesspairing.PairingResult{Byes: append(preAssigned, chesspairing.ByeEntry{PlayerID: players[0].ID, Type: chesspairing.ByePAB}), Notes: []string{players[0].ID + " receives a bye (only active player)"}}, nil
	}
	active := make([]*swisslib.PlayerState, len(players))
	states := make([]swisslib.PlayerState, len(players))
	playerMap := make(map[string]*swisslib.PlayerState, len(players))
	for i := range players {
		active[i] = &players[i]
		states[i] = players[i]
		playerMap[players[i].ID] = &players[i]
	}
	criteria := &swisslib.CriteriaContext{
		Players: playerMap, TotalRounds: totalRounds, CurrentRound: state.CurrentRound,
		IsLastRound: p.opts.TotalRounds != nil && state.CurrentRound == totalRounds,
		TopScorers:  nil, ForbiddenPairs: buildForbiddenPairSet(p.opts.ForbiddenPairs),
	}
	if criteria.IsLastRound {
		realScores := make(map[string]float64, len(players))
		for _, player := range players {
			realScores[player.ID] = player.Score
		}
		criteria.TopScorers = seedingTopScorers(active, state.CurrentRound-1, realScores)
	}
	pairs, unmatched, pairNotes, err := swisslib.PairBracketsGlobal(ctx, swisslib.BuildScoreGroups(states), criteria, swisslib.MatchingCriteria{ApplyC9: true}, playerMap)
	if err != nil {
		return nil, err
	}
	notes = append(notes, pairNotes...)
	sort.SliceStable(pairs, func(i, j int) bool {
		return seedingPairBefore(pairs[i], pairs[j])
	})
	parity := colorParityRanks(original, active)
	topSeed, err := seedingTopSeedColor(p.opts.TopSeedColor)
	if err != nil {
		return nil, fmt.Errorf("dutch: %w", err)
	}
	result := &chesspairing.PairingResult{Notes: notes}
	for i, pair := range pairs {
		white, black := *pair.White, *pair.Black
		white.PairingNumber, black.PairingNumber = parity[white.ID], parity[black.ID]
		whiteID, blackID := swisslib.AllocateColor(&white, &black, criteria.IsLastRound, i+1, topSeed, swisslib.FixedNumberParity)
		result.Pairings = append(result.Pairings, chesspairing.GamePairing{Board: i + 1, WhiteID: whiteID, BlackID: blackID})
	}
	result.Byes = append(result.Byes, preAssigned...)
	if unmatched != nil {
		if unmatched.PABIneligible.Any() {
			return nil, noPABError()
		}
		result.Byes = append(result.Byes, chesspairing.ByeEntry{PlayerID: unmatched.ID, Type: chesspairing.ByePAB})
		result.Notes = append(result.Notes, fmt.Sprintf("%s receives PAB (bye)", unmatched.ID))
	}
	result.Notes = append(result.Notes, "Pairings generated by Dutch Swiss system (FIDE C.04.3)")
	return result, nil
}

func seedingTopScorers(players []*swisslib.PlayerState, playedRounds int, scores map[string]float64) map[string]bool {
	result := make(map[string]bool)
	for _, player := range players {
		if scores[player.ID] > float64(playedRounds)/2 {
			result[player.ID] = true
		}
	}
	return result
}

func seedingPairBefore(a, b swisslib.ProposedPairing) bool {
	maxA, maxB := a.White.PairingScore, b.White.PairingScore
	if a.Black.PairingScore > maxA {
		maxA = a.Black.PairingScore
	}
	if b.Black.PairingScore > maxB {
		maxB = b.Black.PairingScore
	}
	if maxA != maxB {
		return maxA > maxB
	}
	if a.BracketScore != b.BracketScore {
		return a.BracketScore > b.BracketScore
	}
	minA, minB := swisslib.EffectivePairingNumber(a.White), swisslib.EffectivePairingNumber(b.White)
	if value := swisslib.EffectivePairingNumber(a.Black); value < minA {
		minA = value
	}
	if value := swisslib.EffectivePairingNumber(b.Black); value < minB {
		minB = value
	}
	return minA < minB
}

func seedingTopSeedColor(option *string) (*swisslib.Color, error) {
	if option == nil {
		return nil, nil
	}
	switch strings.ToLower(*option) {
	case "auto", "w", "white", "white1":
		return nil, nil
	case "b", "black", "black1":
		color := swisslib.ColorBlack
		return &color, nil
	default:
		return nil, fmt.Errorf("invalid top seed color %q", *option)
	}
}

// colorParityRanks follows the Dutch entered-player TPN numbering used by
// C.04.4.2 Article 5.2.1.
func colorParityRanks(state *chesspairing.TournamentState, active []*swisslib.PlayerState) map[string]int {
	numbers := make(map[string]int, len(state.Players))
	if numbered, err := chesspairing.AssignPairingNumbers(state.Players); err == nil {
		for _, player := range numbered {
			numbers[player.ID] = player.PairingNumber
		}
	}
	entered := make(map[string]bool, len(active))
	for _, player := range active {
		entered[player.ID] = true
		if _, ok := numbers[player.ID]; !ok {
			numbers[player.ID] = swisslib.EffectivePairingNumber(player)
		}
	}
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}
	for _, round := range state.Rounds[:historyEnd] {
		for _, game := range round.Games {
			entered[game.WhiteID], entered[game.BlackID] = true, true
		}
		for _, bye := range round.Byes {
			if bye.Type == chesspairing.ByePAB {
				entered[bye.PlayerID] = true
			}
		}
	}
	ids := make([]string, 0, len(entered))
	for id := range entered {
		if _, ok := numbers[id]; ok {
			ids = append(ids, id)
		}
	}
	sort.SliceStable(ids, func(i, j int) bool {
		if numbers[ids[i]] != numbers[ids[j]] {
			return numbers[ids[i]] < numbers[ids[j]]
		}
		return ids[i] < ids[j]
	})
	ranks := make(map[string]int, len(ids))
	for i, id := range ids {
		ranks[id] = i + 1
	}
	return ranks
}
