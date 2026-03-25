// Package keizer implements Keizer point scoring for chess tournaments.
//
// In Keizer scoring, each player is assigned a value number based on their
// current rank. When you win against an opponent, you receive points equal
// to their value number. Draws award a fraction of the opponent's value.
//
// This creates a self-reinforcing system: beating strong players (high value
// numbers) earns more points, which raises your rank, which increases your
// own value number. Absent players receive a penalty fraction of their own
// value number.
package keizer

import (
	"context"
	"sort"

	chesspairing "github.com/gnutterts/chesspairing"
)

// Ensure Scorer implements chesspairing.Scorer.
var _ chesspairing.Scorer = (*Scorer)(nil)

// Scorer implements the chesspairing.Scorer interface for Keizer scoring.
type Scorer struct {
	opts Options
}

// New creates a new Keizer scorer with the given options.
// Pass nil or empty Options to use all defaults.
func New(opts Options) *Scorer {
	return &Scorer{opts: opts}
}

// NewFromMap creates a new Keizer scorer from a map[string]any config.
func NewFromMap(m map[string]any) *Scorer {
	return New(ParseOptions(m))
}

// Score calculates Keizer scores for all active players.
//
// The algorithm:
// 1. Build initial ranking from ratings (or previous scores if rounds exist).
// 2. For each round, calculate points earned by each player.
// 3. Re-rank players by total Keizer points after each round.
// 4. Value numbers update each round based on current rankings.
//
// This iterative approach is important: value numbers change as rankings
// change, and all rounds use the final ranking's value numbers to compute
// the final scores. (Some Keizer variants recalculate retroactively;
// this implementation uses the standard approach where all rounds are
// scored using the final ranking.)
func (s *Scorer) Score(ctx context.Context, state *chesspairing.TournamentState) ([]chesspairing.PlayerScore, error) {
	if len(state.Players) == 0 {
		return nil, nil
	}

	activePlayers := activePlayerIDs(state.Players)
	playerCount := len(activePlayers)
	opts := s.opts.WithDefaults(playerCount)

	// Build a lookup of player ID → player index for active players.
	playerIndex := make(map[string]int, playerCount)
	for i, id := range activePlayers {
		playerIndex[id] = i
	}

	// Initialize scores to zero.
	scores := make([]float64, playerCount)

	// Determine initial ranking by rating (descending), then alphabetically.
	playerEntries := make(map[string]chesspairing.PlayerEntry, len(state.Players))
	for _, p := range state.Players {
		playerEntries[p.ID] = p
	}
	ranking := initialRanking(activePlayers, playerEntries)

	// If there are no completed rounds, return zero scores ranked by rating.
	if len(state.Rounds) == 0 {
		return buildPlayerScores(activePlayers, scores, ranking), nil
	}

	// Iterative scoring: recompute all rounds with current value numbers,
	// re-rank, repeat until rankings stabilize. Value numbers depend on
	// rank which depends on scores which depend on value numbers.
	// Typically converges in 3-5 iterations.
	//
	// Oscillation detection: when two players have very close scores,
	// the ranking can flip back and forth between iterations (e.g., two
	// players who drew each other). We detect this by remembering the
	// ranking from two iterations ago. If current == two-ago, we have
	// a 2-cycle oscillation and we average the scores from the last
	// two iterations to break the tie.
	const maxIterations = 20

	// Build which players participated in which rounds (constant across iterations).
	playedInRound := buildParticipation(state.Rounds, playerIndex)

	var prevScores []float64
	var twoAgoRanking []string

	for iter := range maxIterations {
		prevRanking := make([]string, len(ranking))
		copy(prevRanking, ranking)

		// Save previous scores for oscillation averaging.
		if iter > 0 {
			prevScores = make([]float64, playerCount)
			copy(prevScores, scores)
		}

		// Reset scores for this iteration.
		for i := range scores {
			scores[i] = 0
		}

		// Build rank lookup from current ranking.
		rankOf := make(map[string]int, playerCount)
		for rank, id := range ranking {
			rankOf[id] = rank + 1 // 1-based
		}

		// Score each round.
		for roundIdx, round := range state.Rounds {
			scoreRound(round, roundIdx, playerIndex, rankOf, opts, activePlayers, playedInRound, scores)
		}

		// Re-rank by score (descending), then by rating (descending).
		ranking = rankByScore(activePlayers, scores, playerEntries)

		// Check for convergence: ranking didn't change.
		if rankingsEqual(prevRanking, ranking) {
			break
		}

		// Check for 2-cycle oscillation: current ranking == two iterations ago.
		if twoAgoRanking != nil && rankingsEqual(twoAgoRanking, ranking) {
			// Average the scores from the last two iterations to break the cycle.
			for i := range scores {
				scores[i] = (scores[i] + prevScores[i]) / 2
			}
			ranking = rankByScore(activePlayers, scores, playerEntries)
			break
		}

		twoAgoRanking = prevRanking
	}

	return buildPlayerScores(activePlayers, scores, ranking), nil
}

// PointsForResult returns the points awarded for a specific game result
// in Keizer scoring. This uses the ResultContext to access opponent/player
// value numbers.
func (s *Scorer) PointsForResult(result chesspairing.GameResult, rctx chesspairing.ResultContext) float64 {
	playerCount := 0
	if rctx.PlayerValueNumber > 0 {
		// Estimate player count from value numbers (rough).
		playerCount = rctx.PlayerValueNumber + rctx.PlayerRank - 1
	}
	opts := s.opts.WithDefaults(playerCount)

	if rctx.IsAbsent {
		return float64(rctx.PlayerValueNumber) * *opts.AbsentPenaltyFraction
	}
	if rctx.IsBye {
		return float64(rctx.PlayerValueNumber) * *opts.ByeValueFraction
	}

	switch result {
	case chesspairing.ResultWhiteWins, chesspairing.ResultBlackWins:
		// Caller determines if this player won or lost.
		// For PointsForResult, we assume the player is the winner
		// if the result matches their color. The caller should handle
		// the perspective. Here we treat it as: the player won.
		return float64(rctx.OpponentValueNumber) * *opts.WinFraction
	case chesspairing.ResultDraw:
		return float64(rctx.OpponentValueNumber) * *opts.DrawFraction
	default:
		return 0
	}
}

// scoreRound processes a single round's games, byes, and absences,
// adding points to the scores slice.
func scoreRound(
	round chesspairing.RoundData,
	roundIdx int,
	playerIndex map[string]int,
	rankOf map[string]int,
	opts Options,
	activePlayers []string,
	playedInRound []map[string]bool,
	scores []float64,
) {
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

		blackRank := rankOf[game.BlackID]
		whiteRank := rankOf[game.WhiteID]
		blackValue := opts.ValueNumber(blackRank)
		whiteValue := opts.ValueNumber(whiteRank)

		// Single forfeit: winner gets forfeit-win fraction, loser gets loss fraction.
		if game.IsForfeit {
			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				scores[whiteIdx] += float64(blackValue) * *opts.WinFraction
				scores[blackIdx] += float64(whiteValue) * *opts.LossFraction
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				scores[blackIdx] += float64(whiteValue) * *opts.WinFraction
				scores[whiteIdx] += float64(blackValue) * *opts.LossFraction
			}
			continue
		}

		switch game.Result {
		case chesspairing.ResultWhiteWins:
			scores[whiteIdx] += float64(blackValue) * *opts.WinFraction
			scores[blackIdx] += float64(whiteValue) * *opts.LossFraction
		case chesspairing.ResultBlackWins:
			scores[blackIdx] += float64(whiteValue) * *opts.WinFraction
			scores[whiteIdx] += float64(blackValue) * *opts.LossFraction
		case chesspairing.ResultDraw:
			scores[whiteIdx] += float64(blackValue) * *opts.DrawFraction
			scores[blackIdx] += float64(whiteValue) * *opts.DrawFraction
		case chesspairing.ResultPending:
			// Game not yet finished — no points.
		}
	}

	// Process byes.
	for _, bye := range round.Byes {
		idx, ok := playerIndex[bye.PlayerID]
		if !ok {
			continue
		}
		rank := rankOf[bye.PlayerID]
		ownValue := opts.ValueNumber(rank)
		scores[idx] += float64(ownValue) * *opts.ByeValueFraction
	}

	// Process absences: players who didn't play and didn't get a bye.
	for _, id := range activePlayers {
		if !playedInRound[roundIdx][id] {
			idx := playerIndex[id]
			rank := rankOf[id]
			ownValue := opts.ValueNumber(rank)
			scores[idx] += float64(ownValue) * *opts.AbsentPenaltyFraction
		}
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

// initialRanking returns player IDs sorted by rating (descending),
// then alphabetically by display name (for deterministic ordering).
func initialRanking(ids []string, entries map[string]chesspairing.PlayerEntry) []string {
	ranked := make([]string, len(ids))
	copy(ranked, ids)
	sort.Slice(ranked, func(i, j int) bool {
		ri := entries[ranked[i]].Rating
		rj := entries[ranked[j]].Rating
		if ri != rj {
			return ri > rj
		}
		return entries[ranked[i]].DisplayName < entries[ranked[j]].DisplayName
	})
	return ranked
}

// rankByScore returns player IDs sorted by score (descending),
// then by rating (descending) as secondary tiebreak for ranking purposes.
func rankByScore(ids []string, scores []float64, entries map[string]chesspairing.PlayerEntry) []string {
	// Build index lookup.
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
		// Tiebreak by rating.
		ri := entries[ranked[i]].Rating
		rj := entries[ranked[j]].Rating
		if ri != rj {
			return ri > rj
		}
		return entries[ranked[i]].DisplayName < entries[ranked[j]].DisplayName
	})
	return ranked
}

// rankingsEqual checks if two ranking slices are identical.
func rankingsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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

	rankOf := make(map[string]int, len(ranking))
	for rank, id := range ranking {
		rankOf[id] = rank + 1
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
