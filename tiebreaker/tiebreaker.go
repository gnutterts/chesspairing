// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package tiebreaker implements chess tournament tiebreakers.
//
// Each tiebreaker implements the chesspairing.TieBreaker interface and computes
// a single numeric value per player. Tiebreakers are applied in order
// to resolve ties in the standings.
//
// The tiebreaker registry provides lookup by ID and library defaults per
// pairing system.
package tiebreaker

import (
	"fmt"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// registry maps tiebreaker IDs to constructor functions.
//
// Safety: all writes happen during init() (via Register calls in each
// tiebreaker file). After init completes, registry is read-only.
// This is safe without synchronization per the Go memory model:
// init functions complete before main starts, establishing a
// happens-before relationship with all subsequent reads.
var registry = map[string]func() chesspairing.TieBreaker{}

// teamRegistry maps team-only tiebreaker IDs to constructor functions.
// Team tiebreakers are kept separate from the individual registry so that
// generic listings (for example factory.TieBreakerIDs) keep describing
// individual tournaments; both registries are reachable through Get.
//
// Safety: same init-only write pattern as registry.
var teamRegistry = map[string]func() chesspairing.TieBreaker{}

// Register adds a tiebreaker constructor to the individual registry.
// Must only be called during init().
func Register(id string, fn func() chesspairing.TieBreaker) {
	registry[id] = fn
}

// RegisterTeam adds a team-only tiebreaker constructor to the team registry.
// Must only be called during init().
func RegisterTeam(id string, fn func() chesspairing.TieBreaker) {
	teamRegistry[id] = fn
}

// Get returns a tiebreaker by ID from either registry.
// Returns an error if the ID is unknown.
func Get(id string) (chesspairing.TieBreaker, error) {
	if fn, ok := registry[id]; ok {
		return fn(), nil
	}
	if fn, ok := teamRegistry[id]; ok {
		return fn(), nil
	}
	return nil, fmt.Errorf("unknown tiebreaker: %q", id)
}

// All returns the IDs of all registered individual tiebreakers.
// Team-only tiebreakers are listed by TeamAll.
func All() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	return ids
}

// TeamAll returns the IDs of all registered team-only tiebreakers.
func TeamAll() []string {
	ids := make([]string, 0, len(teamRegistry))
	for id := range teamRegistry {
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
	// Games counts the games of the round behind this record: one for a
	// single game, two for both games of a Double Swiss match (C.04.5),
	// zero for byes and unpaired rounds. Fore Buchholz (Article 8.3) draws
	// every game of the last round, so it needs the per-match draw value.
	Games int
}

type opponentTable struct {
	records        map[string][]OpponentRecord
	scores         map[string]float64
	adjustedScores map[string]float64
	totalRounds    int
	draw           float64
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
	opts := tournamentPoints(state)
	table.draw = *opts.PointDraw
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
		storeRoundGames(roundRecords, round, ratings, opts)
		for _, bye := range round.Byes {
			record := OpponentRecord{Round: round.Number}
			switch bye.Type {
			case chesspairing.ByePAB, chesspairing.ByeFullPoint:
				record.Points, record.Category = *opts.PointBye, PABOrFullPoint
			case chesspairing.ByeHalf:
				record.Points, record.Category, record.IsVUR = *opts.PointDraw, RequestedByeFinal, true
			case chesspairing.ByeZero:
				record.Points, record.Category, record.IsVUR = *opts.PointLoss, RequestedByeFinal, true
			case chesspairing.ByeAbsent:
				record.Points, record.Category, record.IsVUR = *opts.PointAbsent, RequestedByeFinal, true
			case chesspairing.ByeExcused:
				record.Points, record.Category, record.IsVUR = *opts.PointExcused, RequestedByeFinal, true
			case chesspairing.ByeClubCommitment:
				record.Points, record.Category, record.IsVUR = *opts.PointClubCommitment, RequestedByeFinal, true
			}
			roundRecords[bye.PlayerID] = record
		}
		for _, player := range state.Players {
			table.records[player.ID] = append(table.records[player.ID], roundRecords[player.ID])
		}
	}

	withdrawnAfter := make(map[string]int, len(state.Players))
	for _, player := range state.Players {
		if player.WithdrawnAfterRound != nil {
			withdrawnAfter[player.ID] = *player.WithdrawnAfterRound
		}
	}
	for playerID, records := range table.records {
		for i := range records {
			if records[i].Round > withdrawnAfter[playerID] && withdrawnAfter[playerID] != 0 {
				records[i] = OpponentRecord{Round: records[i].Round, Category: RequestedByeFinal, IsVUR: true}
			}
		}
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
				table.adjustedScores[playerID] += table.draw - record.Points
			}
		}
	}
	return table
}

// gameOutcome classifies one game's result for one of its players.
// forfeitOutcome turns a played win or loss into its forfeit counterpart;
// draws and pending games are returned unchanged.
func forfeitOutcome(o gameOutcome) gameOutcome {
	switch o {
	case outcomePlayedWin:
		return outcomeForfeitWin
	case outcomePlayedLoss:
		return outcomeForfeitLoss
	default:
		return o
	}
}

type gameOutcome int

const (
	outcomePending gameOutcome = iota
	outcomePlayedWin
	outcomePlayedDraw
	outcomePlayedLoss
	outcomeForfeitWin
	outcomeForfeitLoss
)

// gameOutcomeOf classifies a game's result for the white player (white
// true) or the black player.
func gameOutcomeOf(game chesspairing.GameData, white bool) gameOutcome {
	switch game.Result {
	case chesspairing.ResultWhiteWins:
		if white {
			return outcomePlayedWin
		}
		return outcomePlayedLoss
	case chesspairing.ResultBlackWins:
		if white {
			return outcomePlayedLoss
		}
		return outcomePlayedWin
	case chesspairing.ResultDraw:
		return outcomePlayedDraw
	case chesspairing.ResultForfeitWhiteWins:
		if white {
			return outcomeForfeitWin
		}
		return outcomeForfeitLoss
	case chesspairing.ResultForfeitBlackWins:
		if white {
			return outcomeForfeitLoss
		}
		return outcomeForfeitWin
	case chesspairing.ResultDoubleForfeit:
		return outcomeForfeitLoss
	default:
		// A known future pairing is neither a played round nor a VUR.
		return outcomePending
	}
}

// playerForfeitedGame reports whether playerID forfeited one game.
func playerForfeitedGame(game chesspairing.GameData, playerID string) bool {
	switch game.Result {
	case chesspairing.ResultDoubleForfeit:
		return true
	case chesspairing.ResultForfeitWhiteWins:
		return game.BlackID == playerID
	case chesspairing.ResultForfeitBlackWins:
		return game.WhiteID == playerID
	default:
		return false
	}
}

// playerPairKey identifies the two players of a game, order-independently.
func playerPairKey(a, b string) [2]string {
	if a <= b {
		return [2]string{a, b}
	}
	return [2]string{b, a}
}

// storeRoundGames stores the per-player records of one round's games. The
// games between the same two players form a Double Swiss match (C.04.5);
// storeMatchRecords decides the match category from both games before the
// per-player records are merged. If malformed input pairs a player with
// different opponents, the first pair in slice order supplies their record.
func storeRoundGames(roundRecords map[string]OpponentRecord, round chesspairing.RoundData, ratings map[string]int, opts standard.Options) {
	gamesByPair := make(map[[2]string][]chesspairing.GameData)
	var pairOrder [][2]string
	for _, game := range round.Games {
		key := playerPairKey(game.WhiteID, game.BlackID)
		if _, ok := gamesByPair[key]; !ok {
			pairOrder = append(pairOrder, key)
		}
		gamesByPair[key] = append(gamesByPair[key], game)
	}
	seen := make(map[string]bool)
	for _, key := range pairOrder {
		if seen[key[0]] || seen[key[1]] {
			continue
		}
		storeMatchRecords(roundRecords, round.Number, gamesByPair[key], ratings, opts)
		seen[key[0]], seen[key[1]] = true, true
	}
}

// storeMatchRecords stores the round record of each player of one pair's
// games in the current round. A single game stores its own record; the
// games of a Double Swiss match are merged by matchRecord.
func storeMatchRecords(roundRecords map[string]OpponentRecord, roundNumber int, games []chesspairing.GameData, ratings map[string]int, opts standard.Options) {
	whiteID, blackID := games[0].WhiteID, games[0].BlackID
	whiteOutcomes := make([]gameOutcome, 0, len(games))
	blackOutcomes := make([]gameOutcome, 0, len(games))
	whiteForfeits, blackForfeits := 0, 0
	for _, game := range games {
		// The colours swap between the games of a match, so each game is
		// read for the side the pair's players actually took in it.
		whiteIsWhite := game.WhiteID == whiteID
		whiteOutcome := gameOutcomeOf(game, whiteIsWhite)
		blackOutcome := gameOutcomeOf(game, !whiteIsWhite)
		if game.IsForfeit {
			whiteOutcome = forfeitOutcome(whiteOutcome)
			blackOutcome = forfeitOutcome(blackOutcome)
		}
		whiteOutcomes = append(whiteOutcomes, whiteOutcome)
		blackOutcomes = append(blackOutcomes, blackOutcome)
		if whiteOutcome == outcomeForfeitLoss || playerForfeitedGame(game, whiteID) {
			whiteForfeits++
		}
		if blackOutcome == outcomeForfeitLoss || playerForfeitedGame(game, blackID) {
			blackForfeits++
		}
	}
	roundRecords[whiteID] = matchRecord(roundNumber, blackID, ratings[blackID],
		whiteForfeits == len(games), blackForfeits == len(games), whiteOutcomes, opts)
	roundRecords[blackID] = matchRecord(roundNumber, whiteID, ratings[whiteID],
		blackForfeits == len(games), whiteForfeits == len(games), blackOutcomes, opts)
}

// matchRecord merges one player's game outcomes in a Double Swiss match
// (C.04.5) into the per-round record. The match ends by forfeit only when
// one player forfeited every game: the forfeiting player keeps a forfeit
// loss, and the opponent is awarded the forfeit win of the whole match.
// Otherwise the match is regularly played and a single forfeited game is
// treated as played for tie-breaks and standings (ratings keep forfeits
// unplayed, see otbGameRecords); the pending games of an unfinished match
// stay pending.
func matchRecord(roundNumber int, opponentID string, oppRating int, forfeitedAll, opponentForfeitedAll bool, outcomes []gameOutcome, opts standard.Options) OpponentRecord {
	record := OpponentRecord{
		Round:      roundNumber,
		OpponentID: opponentID,
		OppRating:  oppRating,
		Games:      len(outcomes),
	}
	forfeitWin, forfeitLoss := false, false
	for _, outcome := range outcomes {
		switch outcome {
		case outcomePlayedWin:
			record.Played, record.Points = true, record.Points+*opts.PointWin
		case outcomePlayedDraw:
			record.Played, record.Points = true, record.Points+*opts.PointDraw
		case outcomePlayedLoss:
			record.Played, record.Points = true, record.Points+*opts.PointLoss
		case outcomeForfeitWin:
			forfeitWin, record.Points = true, record.Points+*opts.PointForfeitWin
		case outcomeForfeitLoss:
			forfeitLoss, record.Points = true, record.Points+*opts.PointForfeitLoss
		case outcomePending:
		}
	}
	switch {
	case forfeitedAll:
		// This player forfeited every game of the match: it ends by
		// forfeit and is lost.
		record.Category, record.IsVUR, record.Played = ForfeitLoss, true, false
	case opponentForfeitedAll:
		// The opponent forfeited every game of the match: it ends by
		// forfeit and is won.
		record.Category, record.IsVUR, record.Played = ForfeitWin, false, false
	case record.Played || (forfeitWin && forfeitLoss):
		// A game was played over the board, or each player forfeited one
		// game: the match is regularly played and every forfeited single
		// game is treated as played.
		record.Played, record.Category, record.IsVUR = true, None, false
	case forfeitWin:
		// Only part of the match is decided by forfeit so far; the pending
		// games stay pending.
		record.Category, record.IsVUR, record.Played = ForfeitWin, false, false
	case forfeitLoss:
		record.Category, record.IsVUR, record.Played = ForfeitLoss, true, false
	}
	return record
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
