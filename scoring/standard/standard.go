// Package standard implements standard chess scoring (1-½-0).
//
// In standard scoring, each player receives a fixed number of points
// per result: 1 for a win, ½ for a draw, 0 for a loss. Byes, forfeits,
// and absences may have their own configurable point values.
//
// This is the scoring system used by FIDE for Swiss and round-robin
// tournaments, and the system used internally by the Swiss pairer for
// scoregroup formation (even when the tournament's public standings
// use a different scoring system like Keizer).
package standard

import (
	"context"
	"sort"

	"github.com/gnutterts/chesspairing"
)

// Ensure Scorer implements chesspairing.Scorer.
var _ chesspairing.Scorer = (*Scorer)(nil)

// Scorer implements the chesspairing.Scorer interface for standard scoring.
type Scorer struct {
	opts Options
}

// New creates a new standard scorer with the given options.
func New(opts Options) *Scorer {
	return &Scorer{opts: opts}
}

// NewFromMap creates a new standard scorer from a map[string]any config.
func NewFromMap(m map[string]any) *Scorer {
	return New(ParseOptions(m))
}

// Score calculates standard scores for all active players.
//
// Unlike Keizer scoring, standard scoring is simple: each result adds a
// fixed number of points. No iterative convergence is needed because
// points don't depend on rankings.
func (s *Scorer) Score(_ context.Context, state *chesspairing.TournamentState) ([]chesspairing.PlayerScore, error) {
	if len(state.Players) == 0 {
		return nil, nil
	}

	opts := s.opts.WithDefaults()

	// Identify active players.
	activePlayers := activePlayerIDs(state.Players)
	playerCount := len(activePlayers)

	// Build a lookup of player ID → player index for active players.
	playerIndex := make(map[string]int, playerCount)
	for i, id := range activePlayers {
		playerIndex[id] = i
	}

	// Initialize scores to zero.
	scores := make([]float64, playerCount)

	// Build player entries lookup for tiebreak ordering.
	playerEntries := make(map[string]chesspairing.PlayerEntry, len(state.Players))
	for _, p := range state.Players {
		playerEntries[p.ID] = p
	}

	// If there are no rounds, return zero scores ranked by rating.
	if len(state.Rounds) == 0 {
		ranking := rankByScore(activePlayers, scores, playerEntries)
		return buildPlayerScores(activePlayers, scores, ranking), nil
	}

	// Build participation map.
	playedInRound := buildParticipation(state.Rounds, playerIndex)

	// Score each round.
	for roundIdx, round := range state.Rounds {
		// Process game results.
		for _, game := range round.Games {
			whiteIdx, whiteOk := playerIndex[game.WhiteID]
			blackIdx, blackOk := playerIndex[game.BlackID]
			if !whiteOk || !blackOk {
				continue
			}

			// Double forfeit: neither player gets points. They still count
			// as having participated (avoiding absent penalty).
			if game.Result.IsDoubleForfeit() {
				continue
			}

			// Single forfeit: use forfeit-specific point values.
			if game.IsForfeit {
				switch game.Result {
				case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
					scores[whiteIdx] += *opts.PointForfeitWin
					scores[blackIdx] += *opts.PointForfeitLoss
				case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
					scores[blackIdx] += *opts.PointForfeitWin
					scores[whiteIdx] += *opts.PointForfeitLoss
				}
				continue
			}

			switch game.Result {
			case chesspairing.ResultWhiteWins:
				scores[whiteIdx] += *opts.PointWin
				scores[blackIdx] += *opts.PointLoss
			case chesspairing.ResultBlackWins:
				scores[blackIdx] += *opts.PointWin
				scores[whiteIdx] += *opts.PointLoss
			case chesspairing.ResultDraw:
				scores[whiteIdx] += *opts.PointDraw
				scores[blackIdx] += *opts.PointDraw
			case chesspairing.ResultPending:
				// Game not yet finished — no points.
			}
		}

		// Process byes. Different bye types award different points:
		//   PAB (pairing-allocated bye) = full point (PointBye)
		//   Half-bye = draw equivalent (PointDraw)
		//   Zero-bye = loss equivalent (PointLoss)
		//   Absent-bye = absent penalty (PointAbsent)
		for _, bye := range round.Byes {
			idx, ok := playerIndex[bye.PlayerID]
			if !ok {
				continue
			}
			switch bye.Type {
			case chesspairing.ByeHalf:
				scores[idx] += *opts.PointDraw
			case chesspairing.ByeZero:
				scores[idx] += *opts.PointLoss
			case chesspairing.ByeAbsent:
				scores[idx] += *opts.PointAbsent
			default: // ByePAB and any unknown type
				scores[idx] += *opts.PointBye
			}
		}

		// Process absences: players who didn't play and didn't get a bye.
		for _, id := range activePlayers {
			if !playedInRound[roundIdx][id] {
				idx := playerIndex[id]
				scores[idx] += *opts.PointAbsent
			}
		}
	}

	// Rank by score (descending), then by rating (descending).
	ranking := rankByScore(activePlayers, scores, playerEntries)
	return buildPlayerScores(activePlayers, scores, ranking), nil
}

// PointsForResult returns the points awarded for a specific game result
// in standard scoring. For standard scoring, points are fixed regardless
// of opponent — the ResultContext is only used for bye/absent/forfeit flags.
func (s *Scorer) PointsForResult(result chesspairing.GameResult, rctx chesspairing.ResultContext) float64 {
	opts := s.opts.WithDefaults()

	if rctx.IsAbsent {
		return *opts.PointAbsent
	}
	if rctx.IsBye {
		return *opts.PointBye
	}
	if rctx.IsForfeit {
		switch result {
		case chesspairing.ResultWhiteWins, chesspairing.ResultBlackWins:
			return *opts.PointForfeitWin
		default:
			return *opts.PointForfeitLoss
		}
	}

	switch result {
	case chesspairing.ResultWhiteWins, chesspairing.ResultBlackWins:
		return *opts.PointWin
	case chesspairing.ResultDraw:
		return *opts.PointDraw
	default:
		return 0
	}
}

// activePlayerIDs returns IDs of active players in their original order.
func activePlayerIDs(players []chesspairing.PlayerEntry) []string {
	ids := make([]string, 0, len(players))
	for _, p := range players {
		if p.Active {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// rankByScore returns player IDs sorted by score (descending),
// then by rating (descending) as secondary tiebreak.
func rankByScore(ids []string, scores []float64, entries map[string]chesspairing.PlayerEntry) []string {
	idIndex := make(map[string]int, len(ids))
	for i, id := range ids {
		idIndex[id] = i
	}

	ranked := make([]string, len(ids))
	copy(ranked, ids)
	sort.Slice(ranked, func(i, j int) bool {
		si := scores[idIndex[ranked[i]]]
		sj := scores[idIndex[ranked[j]]]
		if si != sj {
			return si > sj
		}
		ri := entries[ranked[i]].Rating
		rj := entries[ranked[j]].Rating
		if ri != rj {
			return ri > rj
		}
		return entries[ranked[i]].DisplayName < entries[ranked[j]].DisplayName
	})
	return ranked
}

// buildParticipation returns, for each round, which players participated
// (either played a game or received a bye).
func buildParticipation(rounds []chesspairing.RoundData, playerIndex map[string]int) []map[string]bool {
	result := make([]map[string]bool, len(rounds))
	for i, round := range rounds {
		participated := make(map[string]bool)
		for _, game := range round.Games {
			if _, ok := playerIndex[game.WhiteID]; ok {
				participated[game.WhiteID] = true
			}
			if _, ok := playerIndex[game.BlackID]; ok {
				participated[game.BlackID] = true
			}
		}
		for _, bye := range round.Byes {
			if _, ok := playerIndex[bye.PlayerID]; ok {
				participated[bye.PlayerID] = true
			}
		}
		result[i] = participated
	}
	return result
}

// buildPlayerScores converts internal scores + ranking into chesspairing.PlayerScore.
func buildPlayerScores(ids []string, scores []float64, ranking []string) []chesspairing.PlayerScore {
	idIndex := make(map[string]int, len(ids))
	for i, id := range ids {
		idIndex[id] = i
	}

	result := make([]chesspairing.PlayerScore, len(ranking))
	for i, id := range ranking {
		result[i] = chesspairing.PlayerScore{
			PlayerID: id,
			Score:    scores[idIndex[id]],
			Rank:     i + 1,
		}
	}
	return result
}
