// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import "sort"

// ByeSelector selects a player to receive the pairing-allocated bye (PAB).
type ByeSelector interface {
	SelectBye(players []*PlayerState) *PlayerState
}

// NeedsBye returns true if the player count is odd.
func NeedsBye(playerCount int) bool {
	return playerCount%2 == 1
}

// GamesPlayed returns the number of games a player has played
// (rounds with a color, excluding byes/absences).
func GamesPlayed(p *PlayerState) int {
	count := 0
	for _, c := range p.ColorHistory {
		if c != ColorNone {
			count++
		}
	}
	return count
}

// DutchByeSelector selects the bye player per Dutch system rules:
// lowest-ranked player (highest TPN) in the lowest score group
// who has not already received a PAB.
type DutchByeSelector struct{}

// SelectBye returns the player to receive the bye, or nil if all have
// already received one (shouldn't happen in valid tournaments).
func (s DutchByeSelector) SelectBye(players []*PlayerState) *PlayerState {
	// Filter to players who haven't received a bye.
	eligible := filterNoByeReceived(players)
	if len(eligible) == 0 {
		return nil
	}

	// Sort by score ascending (C5), then games played descending (C9: fewer
	// unplayed games = more games played), then TPN descending (lowest rank).
	sort.SliceStable(eligible, func(i, j int) bool {
		if eligible[i].Score != eligible[j].Score {
			return eligible[i].Score < eligible[j].Score
		}
		gi := GamesPlayed(eligible[i])
		gj := GamesPlayed(eligible[j])
		if gi != gj {
			return gi > gj // more games played = fewer unplayed → preferred
		}
		return eligible[i].TPN > eligible[j].TPN
	})

	return eligible[0]
}

// BursteinByeSelector selects the bye player per Burstein system rules:
// 1. Lowest score
// 2. Among ties: most games played
// 3. Among ties: lowest ranking (highest TPN)
type BursteinByeSelector struct{}

// SelectBye returns the player to receive the bye per Burstein rules.
func (s BursteinByeSelector) SelectBye(players []*PlayerState) *PlayerState {
	eligible := filterNoByeReceived(players)
	if len(eligible) == 0 {
		return nil
	}

	sort.SliceStable(eligible, func(i, j int) bool {
		// 1. Lowest score first.
		if eligible[i].Score != eligible[j].Score {
			return eligible[i].Score < eligible[j].Score
		}
		// 2. Most games played first.
		gi := GamesPlayed(eligible[i])
		gj := GamesPlayed(eligible[j])
		if gi != gj {
			return gi > gj
		}
		// 3. Highest TPN (lowest ranking) first.
		return eligible[i].TPN > eligible[j].TPN
	})

	return eligible[0]
}

// filterNoByeReceived returns players who have not yet received a PAB.
func filterNoByeReceived(players []*PlayerState) []*PlayerState {
	var eligible []*PlayerState
	for _, p := range players {
		if !p.ByeReceived {
			eligible = append(eligible, p)
		}
	}
	return eligible
}

// Compile-time checks.
var (
	_ ByeSelector = DutchByeSelector{}
	_ ByeSelector = BursteinByeSelector{}
)
