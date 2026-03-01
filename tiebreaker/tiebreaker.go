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
var registry = map[string]func() chesspairing.TieBreaker{}

// Register adds a tiebreaker constructor to the global registry.
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
	playerScoreMap map[string]float64     // player ID → total score
	playerGames    map[string][]gameEntry // player ID → all games played
	playerByes     map[string]int         // player ID → number of bye rounds
	playerAbsences map[string]int         // player ID → number of absent rounds
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
		playerByes:     make(map[string]int),
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
			case chesspairing.ResultPending:
				continue // skip unfinished games
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

		for _, byeID := range round.Byes {
			if activeSet[byeID] {
				data.playerByes[byeID]++
				played[byeID] = true
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
