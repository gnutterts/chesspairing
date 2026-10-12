// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package lexswiss provides shared data structures and algorithms for
// lexicographic Swiss pairing systems. Both the Double-Swiss (C.04.5) and
// Team Swiss (C.04.6) engines build on this foundation.
//
// Even-sized brackets are enumerated by their Article 3.6 identifiers and
// select the first pairing satisfying all criteria. Odd-sized brackets retain
// their existing sequence-order handling.
// This is fundamentally different from Dutch/Burstein/Dubov which use
// Blossom matching or transposition-based approaches.
//
// Key shared components:
//   - ParticipantState: player/team state for pairing
//   - BuildScoreGroups: scoregroup construction
//   - AssignPAB: bye assignment (Art. 3.4)
//   - PairTopDown: top-down bracket construction and upfloater selection (Art. 3.3-3.5)
//   - PairBracket: lexicographic bracket pairing (Art. 3.6)
package lexswiss

import (
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
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
	ID          string
	DisplayName string
	// PairingNumber is retained for compatibility with callers that expose the
	// initial rank. TPN is the fixed, 1-based FIDE Tournament Pairing Number.
	PairingNumber int
	TPN           int      // fixed, 1-based FIDE Tournament Pairing Number
	Score         float64  // cumulative tournament points
	ColorHistory  []Color  // colour per round (index 0 = round 1)
	Opponents     []string // IDs of opponents faced (forfeits excluded)
	MatchesPlayed int      // matches played, excluding byes and forfeited matches
	// WasFloater reports whether the participant played a different score in
	// the preceding round; Double Swiss uses it for C7 and C8, while Team Swiss
	// uses it for C7 and C10.
	WasFloater    bool
	PABIneligible PABIneligibility
	Active        bool
	Rating        int
}

// PABIneligibility records reasons a participant cannot receive a pairing-allocated bye.
type PABIneligibility struct {
	PriorPAB          bool
	FullPointUnplayed bool
}

// Any reports whether the participant is ineligible for a pairing-allocated bye.
func (p PABIneligibility) Any() bool {
	return p.PriorPAB || p.FullPointUnplayed
}

// ByeReceived reports whether the participant is ineligible for a pairing-allocated bye.
func (p ParticipantState) ByeReceived() bool {
	return p.PABIneligible.Any()
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
// Active participants only. Sorted by score (desc), then fixed TPN (asc).
//
// Pairing scores use the tournament's point system.
// Forfeit games are excluded from opponent history (participants can be paired again).
func BuildParticipantStates(state *chesspairing.TournamentState) ([]ParticipantState, error) {
	return BuildParticipantStatesWithPAB(state, 1)
}

// BuildParticipantStatesWithPAB builds participant states with pabPoints for a
// pairing-allocated bye. Double Swiss defaults to win plus draw under C.04.5 Article 1.4.
func BuildParticipantStatesWithPAB(state *chesspairing.TournamentState, pabPoints float64) ([]ParticipantState, error) {
	// Step 1: Assign pairing numbers.
	playersWithNum, err := chesspairing.AssignPairingNumbers(state.Players)
	if err != nil {
		return nil, err
	}

	initialRanks := make(map[string]int, len(playersWithNum))
	for _, p := range playersWithNum {
		initialRanks[p.ID] = p.PairingNumber
	}

	// Step 2: Filter to active players.
	activeSet := make(map[string]bool)
	var activePlayers []chesspairing.PlayerEntry
	for _, p := range state.Players {
		if state.IsActiveInRound(p.ID, state.CurrentRound) {
			activePlayers = append(activePlayers, p)
			activeSet[p.ID] = true
		}
	}

	// Step 3: Compute scores, colour history, opponents, and bye status.
	pointsPolicy := swisslib.PairingPoints(state.ScoringConfig)
	scores := make(map[string]float64)
	colorHistories := make(map[string][]Color)
	opponents := make(map[string][]string)
	matchesPlayed := make(map[string]int)
	pabIneligibility := make(map[string]PABIneligibility)
	beforeLastScores := make(map[string]float64)
	lastOpponents := make(map[string]string)

	// Walk only completed rounds. See swisslib.BuildPlayerStates for the
	// rationale; pre-assigned byes for the upcoming round live in
	// state.PreAssignedByes, not state.Rounds[CurrentRound-1].
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}

	for ri := 0; ri < historyEnd; ri++ {
		if ri == historyEnd-1 {
			for id, score := range scores {
				beforeLastScores[id] = score
			}
		}
		round := state.Rounds[ri]
		lastOpponents = make(map[string]string)
		recordedColour := make(map[string]bool, len(activeSet))
		for _, game := range round.Games {
			white, black := pointsPolicy.Game(game)
			scores[game.WhiteID] += white
			scores[game.BlackID] += black
		}

		for _, match := range doubleSwissMatches(round.Games) {
			if match.forfeit() {
				if winner := match.forfeitWinner(); winner != "" {
					ineligible := pabIneligibility[winner]
					ineligible.FullPointUnplayed = true
					pabIneligibility[winner] = ineligible
				}
			}
			// C.04.5 Preface: a match ends by forfeit only if at least one
			// player forfeits both games; in all other cases it is regularly
			// played (C1, matches played), even when no game was played.
			// Only its colour depends on a played game (Art. 1.6, below).
			if !match.forfeit() {
				matchesPlayed[match.first]++
				matchesPlayed[match.second]++
				opponents[match.first] = append(opponents[match.first], match.second)
				opponents[match.second] = append(opponents[match.second], match.first)
				lastOpponents[match.first] = match.second
				lastOpponents[match.second] = match.first
			}

			if activeSet[match.first] {
				colour := ColorNone
				if match.played > 0 && match.firstGameWhite == match.first {
					colour = ColorWhite
				} else if match.played > 0 {
					colour = ColorBlack
				}
				colorHistories[match.first] = append(colorHistories[match.first], colour)
				recordedColour[match.first] = true
			}
			if activeSet[match.second] {
				colour := ColorNone
				if match.played > 0 && match.firstGameWhite == match.second {
					colour = ColorWhite
				} else if match.played > 0 {
					colour = ColorBlack
				}
				colorHistories[match.second] = append(colorHistories[match.second], colour)
				recordedColour[match.second] = true
			}
		}

		// Byes.
		for _, bye := range round.Byes {
			// A PAB and a requested full-point bye make a participant
			// ineligible for a future PAB under C2.
			ineligible := pabIneligibility[bye.PlayerID]
			switch bye.Type {
			case chesspairing.ByePAB:
				ineligible.PriorPAB = true
			case chesspairing.ByeFullPoint:
				ineligible.FullPointUnplayed = true
			}
			pabIneligibility[bye.PlayerID] = ineligible
			if bye.Type == chesspairing.ByePAB {
				scores[bye.PlayerID] += pabPoints
			} else {
				scores[bye.PlayerID] += pointsPolicy.Bye(bye.Type)
			}
			if activeSet[bye.PlayerID] {
				colorHistories[bye.PlayerID] = append(colorHistories[bye.PlayerID], ColorNone)
				recordedColour[bye.PlayerID] = true
			}
		}
		for _, player := range state.Players {
			if !recordedColour[player.ID] && state.IsActiveInRound(player.ID, ri+1) {
				scores[player.ID] += pointsPolicy.Bye(chesspairing.ByeAbsent)
			}
		}
		for id := range activeSet {
			if !recordedColour[id] {
				colorHistories[id] = append(colorHistories[id], ColorNone)
			}
		}
	}

	// Step 4: Build ParticipantState slice.
	participants := make([]ParticipantState, 0, len(activePlayers))
	for _, p := range activePlayers {
		ps := ParticipantState{
			ID:            p.ID,
			DisplayName:   p.DisplayName,
			PairingNumber: initialRanks[p.ID],
			TPN:           initialRanks[p.ID],
			Score:         swisslib.ScoreGroupKey(scores[p.ID]),
			ColorHistory:  colorHistories[p.ID],
			Opponents:     opponents[p.ID],
			MatchesPlayed: matchesPlayed[p.ID],
			WasFloater:    beforeLastScores[p.ID] != beforeLastScores[lastOpponents[p.ID]] && lastOpponents[p.ID] != "",
			PABIneligible: pabIneligibility[p.ID],
			Active:        true,
			Rating:        p.Rating,
		}
		participants = append(participants, ps)
	}

	return SortParticipants(participants), nil
}

// ComputePABIneligibility returns the map of C2 ineligibility for all participants
// based on completed rounds (C.04.5 2.1.2 / C.04.6 2.1.2 [C2]).
func ComputePABIneligibility(state *chesspairing.TournamentState) map[string]PABIneligibility {
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}
	ineligibility := make(map[string]PABIneligibility)
	for ri := 0; ri < historyEnd; ri++ {
		round := state.Rounds[ri]
		for _, match := range doubleSwissMatches(round.Games) {
			if match.forfeit() {
				if winner := match.forfeitWinner(); winner != "" {
					inel := ineligibility[winner]
					inel.FullPointUnplayed = true
					ineligibility[winner] = inel
				}
			}
		}
		for _, bye := range round.Byes {
			inel := ineligibility[bye.PlayerID]
			switch bye.Type {
			case chesspairing.ByePAB:
				inel.PriorPAB = true
			case chesspairing.ByeFullPoint:
				inel.FullPointUnplayed = true
			}
			ineligibility[bye.PlayerID] = inel
		}
	}
	return ineligibility
}

type doubleSwissMatch struct {
	first, second                            string
	firstGameWhite                           string
	games, played, firstForfeit, lastForfeit int
}

func (m doubleSwissMatch) forfeit() bool {
	return m.games > 0 && (m.firstForfeit == m.games || m.lastForfeit == m.games)
}

func (m doubleSwissMatch) forfeitWinner() string {
	if m.firstForfeit == m.games && m.lastForfeit != m.games {
		return m.second
	}
	if m.lastForfeit == m.games && m.firstForfeit != m.games {
		return m.first
	}
	return ""
}

// doubleSwissMatches groups the two games of a Double-Swiss match. The
// first entry seen is game 1. Under C.04.5's Preface, it ends by forfeit
// only when one player forfeits both.
func doubleSwissMatches(games []chesspairing.GameData) []doubleSwissMatch {
	matches := make(map[[2]string]doubleSwissMatch)
	order := make([][2]string, 0, len(games))
	for _, game := range games {
		key := [2]string{game.WhiteID, game.BlackID}
		if key[1] < key[0] {
			key[0], key[1] = key[1], key[0]
		}
		match, found := matches[key]
		if !found {
			match = doubleSwissMatch{first: key[0], second: key[1], firstGameWhite: game.WhiteID}
			order = append(order, key)
		}
		match.games++
		// A game is a forfeit when game.IsForfeit || game.Result.IsForfeit().
		// trf/convert.go encodes unrated short games (TRF W/D/L) as plain
		// results with IsForfeit: true, so both signals are honoured.
		if game.IsForfeit || game.Result.IsForfeit() {
			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				if game.BlackID == match.first {
					match.firstForfeit++
				} else {
					match.lastForfeit++
				}
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				if game.WhiteID == match.first {
					match.firstForfeit++
				} else {
					match.lastForfeit++
				}
			case chesspairing.ResultDoubleForfeit:
				match.firstForfeit++
				match.lastForfeit++
			default:
				match.firstForfeit++
				match.lastForfeit++
			}
		} else if game.Result != chesspairing.ResultPending && game.Result != "" {
			match.played++
		}
		matches[key] = match
	}
	result := make([]doubleSwissMatch, 0, len(order))
	for _, key := range order {
		result = append(result, matches[key])
	}
	return result
}

// SortParticipants sorts by score and fixed TPN. C.04.5/C.04.6 Article 1.2
// defines the pairing order this way; Article 1.1 keeps the TPN fixed.
func SortParticipants(participants []ParticipantState) []ParticipantState {
	sort.SliceStable(participants, func(i, j int) bool {
		if participants[i].Score != participants[j].Score {
			return participants[i].Score > participants[j].Score
		}
		return participants[i].TPN < participants[j].TPN
	})
	return participants
}
