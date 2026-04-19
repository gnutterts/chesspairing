// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package tiebreaker implements chess tournament tiebreakers.
//
// Each tiebreaker implements the chesspairing.TieBreaker interface and computes
// a single numeric value per player. Tiebreakers are applied in order
// to resolve ties in the standings.
//
// The tiebreaker registry provides lookup by ID and FIDE-recommended
// defaults per pairing system.
package tiebreaker

import (
	"fmt"

	"github.com/gnutterts/chesspairing"
)

// registry maps tiebreaker IDs to constructor functions.
//
// Safety: all writes happen during init() (via Register calls in each
// tiebreaker file). After init completes, registry is read-only.
// This is safe without synchronization per the Go memory model:
// init functions complete before main starts, establishing a
// happens-before relationship with all subsequent reads.
var registry = map[string]func() chesspairing.TieBreaker{}

// Register adds a tiebreaker constructor to the global registry.
// Must only be called during init().
func Register(id string, fn func() chesspairing.TieBreaker) {
	registry[id] = fn
}

// Get returns a tiebreaker by ID. Returns an error if the ID is unknown.
func Get(id string) (chesspairing.TieBreaker, error) {
	fn, ok := registry[id]
	if !ok {
		return nil, fmt.Errorf("unknown tiebreaker: %q", id)
	}
	return fn(), nil
}

// All returns the IDs of all registered tiebreakers.
func All() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	return ids
}

// opponentScores returns a helper that maps each player to the sum of
// their opponents' scores. This is used by Buchholz and Sonneborn-Berger.
type opponentData struct {
	playerScoreMap map[string]float64                      // player ID → total score
	playerGames    map[string][]gameEntry                  // player ID → all games played
	playerByes     map[string]map[chesspairing.ByeType]int // player ID → bye type → count
	playerAbsences map[string]int                          // player ID → number of absent rounds (no record at all)
}

// virtualOpponentRounds returns the number of rounds for which the
// player should be assigned a virtual opponent for raw Buchholz-style
// sums.
//
// Per FIDE simplified VOO (2023+), every unplayed round — regardless
// of bye type — contributes a virtual opponent equal to the player's
// own score. This includes true absences (active player with no
// record at all in the round). Tiebreakers that need per-bye-type
// policy (e.g. average-buchholz divisor, FIDE category-A "counts as
// played") should consult playerByes directly via countsAsPlayed.
func (d opponentData) virtualOpponentRounds(playerID string) int {
	return d.totalByes(playerID) + d.playerAbsences[playerID]
}

// countsAsPlayed returns the number of rounds the player effectively
// played, summed across actual games and bye types whose contract is
// "counts as played" (PAB, Half, Zero per the v0.2.0 matrix).
//
// Used by tiebreakers whose divisor must exclude rounds that do not
// count as played: ByeAbsent, ByeExcused, ByeClubCommitment, and true
// absences.
func (d opponentData) countsAsPlayed(playerID string) int {
	played := len(d.playerGames[playerID])
	byes := d.playerByes[playerID]
	played += byes[chesspairing.ByePAB] + byes[chesspairing.ByeHalf] + byes[chesspairing.ByeZero]
	return played
}

// totalByes returns the count of all bye types for a player, regardless
// of whether they count as played.
func (d opponentData) totalByes(playerID string) int {
	var n int
	for _, c := range d.playerByes[playerID] {
		n += c
	}
	return n
}

type gameEntry struct {
	opponentID string
	result     playerResult
}

type playerResult int

const (
	resultWin playerResult = iota
	resultDraw
	resultLoss
)

// buildOpponentData constructs the opponent data structure from tournament state.
func buildOpponentData(state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) opponentData {
	data := opponentData{
		playerScoreMap: make(map[string]float64, len(scores)),
		playerGames:    make(map[string][]gameEntry),
		playerByes:     make(map[string]map[chesspairing.ByeType]int),
		playerAbsences: make(map[string]int),
	}

	for _, ps := range scores {
		data.playerScoreMap[ps.PlayerID] = ps.Score
	}

	// Build set of active player IDs.
	activeSet := make(map[string]bool)
	for _, p := range state.Players {
		if p.Active {
			activeSet[p.ID] = true
		}
	}

	for _, round := range state.Rounds {
		played := make(map[string]bool)

		for _, game := range round.Games {
			if !activeSet[game.WhiteID] || !activeSet[game.BlackID] {
				continue
			}

			var whiteResult, blackResult playerResult
			switch game.Result {
			case chesspairing.ResultWhiteWins:
				whiteResult = resultWin
				blackResult = resultLoss
			case chesspairing.ResultBlackWins:
				whiteResult = resultLoss
				blackResult = resultWin
			case chesspairing.ResultDraw:
				whiteResult = resultDraw
				blackResult = resultDraw
			case chesspairing.ResultPending,
				chesspairing.ResultForfeitWhiteWins,
				chesspairing.ResultForfeitBlackWins,
				chesspairing.ResultDoubleForfeit:
				continue // skip unfinished and forfeited games
			}

			data.playerGames[game.WhiteID] = append(data.playerGames[game.WhiteID], gameEntry{
				opponentID: game.BlackID,
				result:     whiteResult,
			})
			data.playerGames[game.BlackID] = append(data.playerGames[game.BlackID], gameEntry{
				opponentID: game.WhiteID,
				result:     blackResult,
			})
			played[game.WhiteID] = true
			played[game.BlackID] = true
		}

		for _, bye := range round.Byes {
			if activeSet[bye.PlayerID] {
				if data.playerByes[bye.PlayerID] == nil {
					data.playerByes[bye.PlayerID] = make(map[chesspairing.ByeType]int)
				}
				data.playerByes[bye.PlayerID][bye.Type]++
				played[bye.PlayerID] = true
			}
		}

		// Absent players: active but didn't play or get a bye.
		for id := range activeSet {
			if !played[id] {
				data.playerAbsences[id]++
			}
		}
	}

	return data
}
