// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"fmt"

	"github.com/gnutterts/chesspairing"
)

// unratedOverride returns the rating the tournament regulations assign to
// unrated players, if any. C.07:2026 Article 10 drops ratings-based
// tie-breaks when unrated players are present, "unless detailed rules on
// the handling of unrated players are included in the tournament
// regulations"; the library models such rules with the scoring option
// "unratedRating", used as the rating of every unrated player. ok is false
// when the regulations do not define one.
func unratedOverride(state *chesspairing.TournamentState) (rating float64, ok bool) {
	rating, ok = chesspairing.GetFloat64(state.ScoringConfig.Options, "unratedRating")
	if ok && rating <= 0 {
		return 0, false
	}
	return rating, ok
}

// opponentRating returns the rating of one opponent in a ratings-based
// tie-break (C.07 Article 10). An unrated opponent (rating 0) in a game
// played over the board forces the tie-break to be dropped unless the
// tournament regulations define the handling of unrated players.
func opponentRating(state *chesspairing.TournamentState, record OpponentRecord, override float64, hasOverride bool) (float64, error) {
	if record.OppRating != 0 {
		return float64(record.OppRating), nil
	}
	if hasOverride {
		return override, nil
	}
	return 0, fmt.Errorf(
		"unrated player %s (rating 0) played a game: C.07 Art. 10 drops this tie-break unless the tournament regulations define the handling of unrated players (scoring option \"unratedRating\")",
		playerName(state, record.OpponentID),
	)
}

// playerName returns the display name of a player, or their ID when no
// display name is known.
func playerName(state *chesspairing.TournamentState, playerID string) string {
	for _, player := range state.Players {
		if player.ID == playerID {
			if player.DisplayName != "" {
				return player.DisplayName
			}
			return player.ID
		}
	}
	return playerID
}

// otbGameRecords returns one record per game the player actually played
// over the board. Ratings-based tie-breaks (C.07 Article 10) count games:
// Article 10.2 divides "the number of points achieved in games played over
// the board" by "the number of games", so in a Double Swiss (C.04.5) both
// games of a match count separately and the per-round merged record cannot
// supply the arithmetic. Forfeits stay unplayed for these tie-breaks (the
// C.04.5 Preface rule treating a forfeited single game as played "does not
// apply to rating"), and rounds after a withdrawal are excluded like in
// the per-round record model.
func otbGameRecords(state *chesspairing.TournamentState, playerID string) []OpponentRecord {
	var withdrawnAfter int
	ratings := make(map[string]int, len(state.Players))
	for _, player := range state.Players {
		ratings[player.ID] = player.Rating
		if player.ID == playerID && player.WithdrawnAfterRound != nil {
			withdrawnAfter = *player.WithdrawnAfterRound
		}
	}
	records := make([]OpponentRecord, 0, len(state.Rounds))
	for _, round := range state.Rounds {
		if withdrawnAfter != 0 && round.Number > withdrawnAfter {
			continue
		}
		for _, game := range round.Games {
			var record OpponentRecord
			switch {
			case game.WhiteID == playerID:
				record = OpponentRecord{Round: round.Number, OpponentID: game.BlackID, OppRating: ratings[game.BlackID]}
			case game.BlackID == playerID:
				record = OpponentRecord{Round: round.Number, OpponentID: game.WhiteID, OppRating: ratings[game.WhiteID]}
			default:
				continue
			}
			switch game.Result {
			case chesspairing.ResultDraw:
				record.Points, record.Played = 0.5, true
			case chesspairing.ResultWhiteWins:
				record.Played = true
				if game.WhiteID == playerID {
					record.Points = 1
				}
			case chesspairing.ResultBlackWins:
				record.Played = true
				if game.BlackID == playerID {
					record.Points = 1
				}
			default:
				// Byes, forfeits and pending games are not games played
				// over the board.
				continue
			}
			records = append(records, record)
		}
	}
	return records
}
