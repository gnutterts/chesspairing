// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package lexswiss provides shared data structures and algorithms for
// lexicographic Swiss pairing systems. Both the Double-Swiss (C.04.5) and
// Team Swiss (C.04.6) engines build on this foundation.
//
// The lexicographic approach enumerates all legal pairings of a bracket in
// lexicographic order and selects the first one satisfying all criteria.
// This is fundamentally different from Dutch/Burstein/Dubov which use
// Blossom matching or transposition-based approaches.
//
// Key shared components:
//   - ParticipantState: player/team state for pairing
//   - BuildScoreGroups: scoregroup construction
//   - AssignPAB: bye assignment (Art. 3.4)
//   - SelectUpfloaters: upfloater selection (Art. 3.5)
//   - PairBracket: lexicographic bracket pairing (Art. 3.6)
package lexswiss

import (
	"sort"

	"github.com/gnutterts/chesspairing"
)

// Color represents a participant's colour assignment in a round.
type Color int

const (
	ColorNone Color = iota // bye, absent, or no game
	ColorWhite
	ColorBlack
)

// String returns the colour name for debugging.
func (c Color) String() string {
	switch c {
	case ColorWhite:
		return "White"
	case ColorBlack:
		return "Black"
	default:
		return "None"
	}
}

// Opposite returns the opposite colour (White↔Black). None returns None.
func (c Color) Opposite() Color {
	switch c {
	case ColorWhite:
		return ColorBlack
	case ColorBlack:
		return ColorWhite
	default:
		return ColorNone
	}
}

// ParticipantState holds the computed state of a single participant (player
// or team) for the lexicographic pairing algorithm. Built once per Pair()
// call from the engine's TournamentState.
//
// This is deliberately simpler than swisslib.PlayerState because the
// lexicographic systems don't need float history, Blossom criteria weights,
// or the three-tier colour preference system.
type ParticipantState struct {
	ID           string
	DisplayName  string
	InitialRank  int      // starting rank (by rating desc, then name asc), 1-based
	TPN          int      // Tournament Pairing Number, 1-based (re-ranked each round)
	Score        float64  // cumulative pairing score (standard 1-½-0)
	ColorHistory []Color  // colour per round (index 0 = round 1)
	Opponents    []string // IDs of opponents faced (forfeits excluded)
	ByeReceived  bool     // already received a PAB
	Active       bool
	Rating       int
}

// HasPlayed returns true if participant a has played against participant b
// (based on opponent history, which excludes forfeits).
func HasPlayed(a, b *ParticipantState) bool {
	for _, opp := range a.Opponents {
		if opp == b.ID {
			return true
		}
	}
	return false
}

// BuildParticipantStates converts a TournamentState into a sorted slice of
// ParticipantState values ready for the pairing algorithm.
//
// Active participants only. Sorted by score (desc), then initial rank (asc).
// TPN assigned sequentially after sorting.
//
// Pairing scores use standard 1-½-0 regardless of tournament scoring system.
// Forfeit games are excluded from opponent history (participants can be paired again).
func BuildParticipantStates(state *chesspairing.TournamentState) []ParticipantState {
	// Step 1: Assign initial ranks by rating desc, name asc.
	allPlayers := make([]chesspairing.PlayerEntry, len(state.Players))
	copy(allPlayers, state.Players)
	sort.SliceStable(allPlayers, func(i, j int) bool {
		if allPlayers[i].Rating != allPlayers[j].Rating {
			return allPlayers[i].Rating > allPlayers[j].Rating
		}
		return allPlayers[i].DisplayName < allPlayers[j].DisplayName
	})

	initialRanks := make(map[string]int, len(allPlayers))
	for i, p := range allPlayers {
		initialRanks[p.ID] = i + 1
	}

	// Step 2: Filter to active players.
	activeSet := make(map[string]bool)
	var activePlayers []chesspairing.PlayerEntry
	for _, p := range state.Players {
		if p.Active {
			activePlayers = append(activePlayers, p)
			activeSet[p.ID] = true
		}
	}

	// Step 3: Compute scores, colour history, opponents, and bye status.
	scores := make(map[string]float64)
	colorHistories := make(map[string][]Color)
	opponents := make(map[string][]string)
	byeReceived := make(map[string]bool)

	for _, round := range state.Rounds {
		for _, game := range round.Games {
			// Score: standard 1-½-0.
			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				scores[game.WhiteID] += 1.0
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				scores[game.BlackID] += 1.0
			case chesspairing.ResultDraw:
				scores[game.WhiteID] += 0.5
				scores[game.BlackID] += 0.5
			}

			// Colour history.
			if activeSet[game.WhiteID] {
				colorHistories[game.WhiteID] = append(colorHistories[game.WhiteID], ColorWhite)
			}
			if activeSet[game.BlackID] {
				colorHistories[game.BlackID] = append(colorHistories[game.BlackID], ColorBlack)
			}

			// Opponent history: exclude forfeits.
			if !game.IsForfeit {
				opponents[game.WhiteID] = append(opponents[game.WhiteID], game.BlackID)
				opponents[game.BlackID] = append(opponents[game.BlackID], game.WhiteID)
			}
		}

		// Byes.
		for _, bye := range round.Byes {
			byeReceived[bye.PlayerID] = true
			switch bye.Type {
			case chesspairing.ByePAB:
				scores[bye.PlayerID] += 1.0
			case chesspairing.ByeHalf:
				scores[bye.PlayerID] += 0.5
			}
			if activeSet[bye.PlayerID] {
				colorHistories[bye.PlayerID] = append(colorHistories[bye.PlayerID], ColorNone)
			}
		}
	}

	// Step 4: Build ParticipantState slice.
	participants := make([]ParticipantState, 0, len(activePlayers))
	for _, p := range activePlayers {
		ps := ParticipantState{
			ID:           p.ID,
			DisplayName:  p.DisplayName,
			InitialRank:  initialRanks[p.ID],
			Score:        scores[p.ID],
			ColorHistory: colorHistories[p.ID],
			Opponents:    opponents[p.ID],
			ByeReceived:  byeReceived[p.ID],
			Active:       true,
			Rating:       p.Rating,
		}
		participants = append(participants, ps)
	}

	// Step 5: Sort by score desc, then initial rank asc. Assign TPN.
	sort.SliceStable(participants, func(i, j int) bool {
		if participants[i].Score != participants[j].Score {
			return participants[i].Score > participants[j].Score
		}
		return participants[i].InitialRank < participants[j].InitialRank
	})

	for i := range participants {
		participants[i].TPN = i + 1
	}

	return participants
}
