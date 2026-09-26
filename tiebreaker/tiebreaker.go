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

// UnplayedCategory classifies an unplayed round under FIDE C.07 Article 16.2.
type UnplayedCategory int

const (
	None UnplayedCategory = iota
	PABOrFullPoint
	ForfeitWin
	RequestedByeFollowedByPlay
	ForfeitLoss
	RequestedByeFinal
)

// OpponentRecord is the canonical per-round input for opponent-based
// tie-breaks. OpponentID is retained for forfeits and scheduled games.
type OpponentRecord struct {
	Round      int
	Played     bool
	OpponentID string
	Points     float64
	Category   UnplayedCategory
	IsVUR      bool
	OppRating  int
}

type opponentTable struct {
	records        map[string][]OpponentRecord
	scores         map[string]float64
	adjustedScores map[string]float64
	totalRounds    int
	roundRobin     bool
}

// buildOpponentRecords is the single translation from TournamentState to
// the records consumed by opponent-based tie-breaks.
func buildOpponentRecords(state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) opponentTable {
	table := opponentTable{
		records:        make(map[string][]OpponentRecord, len(state.Players)),
		scores:         make(map[string]float64, len(state.Players)),
		adjustedScores: make(map[string]float64, len(state.Players)),
		totalRounds:    len(state.Rounds),
		roundRobin:     state.PairingConfig.System == chesspairing.PairingRoundRobin,
	}
	ratings := make(map[string]int, len(state.Players))
	for _, player := range state.Players {
		ratings[player.ID] = player.Rating
		table.records[player.ID] = make([]OpponentRecord, 0, len(state.Rounds))
	}

	for _, round := range state.Rounds {
		roundRecords := make(map[string]OpponentRecord, len(state.Players))
		for _, player := range state.Players {
			roundRecords[player.ID] = OpponentRecord{
				Round: round.Number, Category: RequestedByeFinal, IsVUR: true,
			}
		}
		for _, game := range round.Games {
			white := OpponentRecord{Round: round.Number, OpponentID: game.BlackID, OppRating: ratings[game.BlackID]}
			black := OpponentRecord{Round: round.Number, OpponentID: game.WhiteID, OppRating: ratings[game.WhiteID]}
			switch game.Result {
			case chesspairing.ResultWhiteWins:
				white.Played, white.Points = true, 1
				black.Played = true
			case chesspairing.ResultBlackWins:
				white.Played = true
				black.Played, black.Points = true, 1
			case chesspairing.ResultDraw:
				white.Played, white.Points = true, 0.5
				black.Played, black.Points = true, 0.5
			case chesspairing.ResultForfeitWhiteWins:
				white.Points, white.Category = 1, ForfeitWin
				black.Category, black.IsVUR = ForfeitLoss, true
			case chesspairing.ResultForfeitBlackWins:
				white.Category, white.IsVUR = ForfeitLoss, true
				black.Points, black.Category = 1, ForfeitWin
			case chesspairing.ResultDoubleForfeit:
				white.Category, white.IsVUR = ForfeitLoss, true
				black.Category, black.IsVUR = ForfeitLoss, true
			case chesspairing.ResultPending:
				// A known future pairing is neither a played round nor a VUR.
			}
			roundRecords[game.WhiteID] = white
			roundRecords[game.BlackID] = black
		}
		for _, bye := range round.Byes {
			record := OpponentRecord{Round: round.Number}
			switch bye.Type {
			case chesspairing.ByePAB:
				record.Points, record.Category = 1, PABOrFullPoint
			case chesspairing.ByeHalf:
				record.Points, record.Category, record.IsVUR = 0.5, RequestedByeFinal, true
			case chesspairing.ByeZero, chesspairing.ByeAbsent,
				chesspairing.ByeExcused, chesspairing.ByeClubCommitment:
				record.Category, record.IsVUR = RequestedByeFinal, true
			}
			roundRecords[bye.PlayerID] = record
		}
		for _, player := range state.Players {
			table.records[player.ID] = append(table.records[player.ID], roundRecords[player.ID])
		}
	}

	for playerID, records := range table.records {
		for i := range records {
			if records[i].Category != RequestedByeFinal {
				continue
			}
			for j := i + 1; j < len(records); j++ {
				if isNonVUR(records[j]) {
					records[i].Category = RequestedByeFollowedByPlay
					break
				}
			}
		}
		table.records[playerID] = records
	}

	for _, score := range scores {
		table.scores[score.PlayerID] = score.Score
	}
	for playerID, records := range table.records {
		if _, ok := table.scores[playerID]; !ok {
			for _, record := range records {
				table.scores[playerID] += record.Points
			}
		}
		table.adjustedScores[playerID] = table.scores[playerID]
		for _, record := range records {
			if record.Category == RequestedByeFinal {
				table.adjustedScores[playerID] += 0.5 - record.Points
			}
		}
	}
	return table
}

// isNonVUR reports whether a round is not an unplayed round (VUR) under
// C.07:2026 Article 16.1.2. Only rounds actually played over the board,
// full-point byes (PAB) and forfeit wins count as non-VUR. Requested byes,
// forfeit losses and pending rounds do not.
func isNonVUR(record OpponentRecord) bool {
	if record.Played {
		return true
	}
	switch record.Category {
	case PABOrFullPoint, ForfeitWin:
		return true
	default:
		return false
	}
}

func playedRecords(records []OpponentRecord) []OpponentRecord {
	played := make([]OpponentRecord, 0, len(records))
	for _, record := range records {
		if record.Played {
			played = append(played, record)
		}
	}
	return played
}

func scoresForAllPlayers(table opponentTable) []chesspairing.PlayerScore {
	scores := make([]chesspairing.PlayerScore, 0, len(table.records))
	for playerID := range table.records {
		scores = append(scores, chesspairing.PlayerScore{PlayerID: playerID, Score: table.scores[playerID]})
	}
	return scores
}
