// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

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
//
// Scores are computed using ×2 integer arithmetic internally. Each game or
// absence contributes an integer "doubled" score. The final output divides
// by 2, giving scores rounded to the nearest 0.5. This eliminates float
// precision drift while preserving half-point granularity.
package keizer

import (
	"context"
	"math"
	"sort"

	"github.com/gnutterts/chesspairing"
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
//
// When SelfVictory is enabled (default), each player's own Keizer value
// is added to their total once (not per round). This is standard in every
// known Keizer implementation.
func (s *Scorer) Score(_ context.Context, state *chesspairing.TournamentState) ([]chesspairing.PlayerScore, error) {
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

	// Initialize ×2 scores to zero.
	scoresX2 := make([]int, playerCount)

	// Determine initial ranking by rating (descending), then alphabetically.
	playerEntries := make(map[string]chesspairing.PlayerEntry, len(state.Players))
	for _, p := range state.Players {
		playerEntries[p.ID] = p
	}
	ranking := initialRanking(activePlayers, playerEntries)

	// If there are no completed rounds, return zero scores ranked by rating.
	if len(state.Rounds) == 0 {
		return buildPlayerScores(activePlayers, scoresX2, ranking), nil
	}

	// Build which players participated in which rounds.
	playedInRound := buildParticipation(state.Rounds, playerIndex)

	// Build late-joiner lookup from player entries.
	joinedRound := make(map[string]int, playerCount)
	for _, p := range state.Players {
		if p.Active && p.JoinedRound > 0 {
			joinedRound[p.ID] = p.JoinedRound
		}
	}

	if *opts.Frozen {
		scoreFrozen(state.Rounds, playerIndex, opts, activePlayers, playerEntries, playedInRound, scoresX2, &ranking, joinedRound)
	} else {
		scoreIterative(state.Rounds, playerIndex, playerCount, opts, activePlayers, playerEntries, playedInRound, scoresX2, &ranking, joinedRound)
	}

	return buildPlayerScores(activePlayers, scoresX2, ranking), nil
}

// scoreFrozen implements the frozen Keizer variant. Each round is scored
// once using the ranking as it stood before that round. Points from earlier
// rounds are never retroactively recalculated when later rounds shift the
// rankings. Self-victory is added once at the end using the final ranking.
func scoreFrozen(
	rounds []chesspairing.RoundData,
	playerIndex map[string]int,
	opts Options,
	activePlayers []string,
	playerEntries map[string]chesspairing.PlayerEntry,
	playedInRound []map[string]bool,
	scoresX2 []int,
	ranking *[]string,
	joinedRound map[string]int,
) {
	absenceCounts := make([]int, len(activePlayers))

	for roundIdx, round := range rounds {
		// Value numbers for this round are locked to the current ranking.
		rankOf := make(map[string]int, len(activePlayers))
		for rank, id := range *ranking {
			rankOf[id] = rank + 1
		}

		scoreRound(round, roundIdx, playerIndex, rankOf, opts, activePlayers, playedInRound, scoresX2, absenceCounts, joinedRound)

		// Re-rank after each round so the next round uses updated value numbers.
		*ranking = rankByScore(activePlayers, scoresX2, playerEntries)
	}

	// Self-victory: add own value once, using the final ranking.
	if *opts.SelfVictory {
		rankOf := make(map[string]int, len(activePlayers))
		for rank, id := range *ranking {
			rankOf[id] = rank + 1
		}
		for _, id := range activePlayers {
			idx := playerIndex[id]
			ownValue := opts.ValueNumber(rankOf[id])
			scoresX2[idx] += ownValue * 2
		}
		*ranking = rankByScore(activePlayers, scoresX2, playerEntries)
	}
}

// scoreIterative implements the standard iterative Keizer scoring. All
// rounds are rescored with the current ranking's value numbers, then
// players are re-ranked by the new totals. This repeats until the ranking
// converges (typically 3-5 iterations). Oscillation detection breaks
// 2-cycles by averaging scores from the last two iterations.
func scoreIterative(
	rounds []chesspairing.RoundData,
	playerIndex map[string]int,
	playerCount int,
	opts Options,
	activePlayers []string,
	playerEntries map[string]chesspairing.PlayerEntry,
	playedInRound []map[string]bool,
	scoresX2 []int,
	ranking *[]string,
	joinedRound map[string]int,
) {
	const maxIterations = 20

	var prevScoresX2 []int
	var twoAgoRanking []string

	for iter := range maxIterations {
		prevRanking := make([]string, len(*ranking))
		copy(prevRanking, *ranking)

		if iter > 0 {
			prevScoresX2 = make([]int, playerCount)
			copy(prevScoresX2, scoresX2)
		}

		for i := range scoresX2 {
			scoresX2[i] = 0
		}

		rankOf := make(map[string]int, playerCount)
		for rank, id := range *ranking {
			rankOf[id] = rank + 1
		}

		absenceCounts := make([]int, playerCount)

		for roundIdx, round := range rounds {
			scoreRound(round, roundIdx, playerIndex, rankOf, opts, activePlayers, playedInRound, scoresX2, absenceCounts, joinedRound)
		}

		if *opts.SelfVictory {
			for _, id := range activePlayers {
				idx := playerIndex[id]
				rank := rankOf[id]
				ownValue := opts.ValueNumber(rank)
				scoresX2[idx] += ownValue * 2
			}
		}

		*ranking = rankByScore(activePlayers, scoresX2, playerEntries)

		if rankingsEqual(prevRanking, *ranking) {
			break
		}

		if twoAgoRanking != nil && rankingsEqual(twoAgoRanking, *ranking) {
			for i := range scoresX2 {
				scoresX2[i] = int(math.Round(float64(scoresX2[i]+prevScoresX2[i]) / 2.0))
			}
			*ranking = rankByScore(activePlayers, scoresX2, playerEntries)
			break
		}

		twoAgoRanking = prevRanking
	}
}

// PointsForResult returns the points awarded for a specific game result
// in Keizer scoring. This uses the ResultContext to access opponent/player
// value numbers. The result is rounded to 0.5 precision via ×2 arithmetic.
func (s *Scorer) PointsForResult(result chesspairing.GameResult, rctx chesspairing.ResultContext) float64 {
	playerCount := 0
	if rctx.PlayerValueNumber > 0 {
		// Estimate player count from value numbers (rough).
		playerCount = rctx.PlayerValueNumber + rctx.PlayerRank - 1
	}
	opts := s.opts.WithDefaults(playerCount)

	if rctx.IsAbsent {
		if opts.AbsentFixedValue != nil {
			return float64(*opts.AbsentFixedValue)
		}
		return float64(scoreX2(rctx.PlayerValueNumber, *opts.AbsentPenaltyFraction)) / 2.0
	}
	if rctx.IsBye {
		if opts.ByeFixedValue != nil {
			return float64(*opts.ByeFixedValue)
		}
		return float64(scoreX2(rctx.PlayerValueNumber, *opts.ByeValueFraction)) / 2.0
	}

	switch result {
	case chesspairing.ResultWhiteWins, chesspairing.ResultBlackWins:
		return float64(scoreX2(rctx.OpponentValueNumber, *opts.WinFraction)) / 2.0
	case chesspairing.ResultDraw:
		return float64(scoreX2(rctx.OpponentValueNumber, *opts.DrawFraction)) / 2.0
	case chesspairing.ResultForfeitWhiteWins, chesspairing.ResultForfeitBlackWins:
		return float64(scoreX2(rctx.OpponentValueNumber, *opts.ForfeitWinFraction)) / 2.0
	case chesspairing.ResultDoubleForfeit:
		return float64(scoreX2(rctx.OpponentValueNumber, *opts.DoubleForfeitFraction)) / 2.0
	default:
		return 0
	}
}

// scoreX2 computes the ×2 integer score for a value and fraction.
// result = round(value × fraction × 2)
func scoreX2(value int, fraction float64) int {
	return int(math.Round(float64(value) * fraction * 2))
}

// fixedX2 converts a fixed value (in real units) to ×2 representation.
func fixedX2(fixedValue int) int {
	return fixedValue * 2
}

// scoreRound processes a single round's games, byes, and absences,
// adding ×2 points to the scoresX2 slice. It also updates absenceCounts
// for absence limit/decay tracking. The joinedRound map holds each
// player's JoinedRound value; rounds before a player joined use
// LateJoinHandicap instead of absence scoring.
func scoreRound(
	round chesspairing.RoundData,
	roundIdx int,
	playerIndex map[string]int,
	rankOf map[string]int,
	opts Options,
	activePlayers []string,
	playedInRound []map[string]bool,
	scoresX2 []int,
	absenceCounts []int,
	joinedRound map[string]int,
) {
	// Process game results.
	for _, game := range round.Games {
		whiteIdx, whiteOk := playerIndex[game.WhiteID]
		blackIdx, blackOk := playerIndex[game.BlackID]
		if !whiteOk || !blackOk {
			continue
		}

		blackRank := rankOf[game.BlackID]
		whiteRank := rankOf[game.WhiteID]
		blackValue := opts.ValueNumber(blackRank)
		whiteValue := opts.ValueNumber(whiteRank)

		// Double forfeit: both players get DoubleForfeitFraction × opponent value.
		// They still count as having participated (avoiding absent penalty).
		if game.Result.IsDoubleForfeit() {
			scoresX2[whiteIdx] += scoreX2(blackValue, *opts.DoubleForfeitFraction)
			scoresX2[blackIdx] += scoreX2(whiteValue, *opts.DoubleForfeitFraction)
			continue
		}

		// Single forfeit: use forfeit-specific fractions.
		if game.IsForfeit {
			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				scoresX2[whiteIdx] += scoreX2(blackValue, *opts.ForfeitWinFraction)
				scoresX2[blackIdx] += scoreX2(whiteValue, *opts.ForfeitLossFraction)
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				scoresX2[blackIdx] += scoreX2(whiteValue, *opts.ForfeitWinFraction)
				scoresX2[whiteIdx] += scoreX2(blackValue, *opts.ForfeitLossFraction)
			}
			continue
		}

		// Regular game results.
		switch game.Result {
		case chesspairing.ResultWhiteWins:
			scoresX2[whiteIdx] += scoreX2(blackValue, *opts.WinFraction)
			scoresX2[blackIdx] += scoreX2(whiteValue, *opts.LossFraction)
		case chesspairing.ResultBlackWins:
			scoresX2[blackIdx] += scoreX2(whiteValue, *opts.WinFraction)
			scoresX2[whiteIdx] += scoreX2(blackValue, *opts.LossFraction)
		case chesspairing.ResultDraw:
			scoresX2[whiteIdx] += scoreX2(blackValue, *opts.DrawFraction)
			scoresX2[blackIdx] += scoreX2(whiteValue, *opts.DrawFraction)
		case chesspairing.ResultPending:
			// Game not yet finished — no points.
		}
	}

	// Process byes: dispatch by bye type with fixed-value override support.
	for _, bye := range round.Byes {
		idx, ok := playerIndex[bye.PlayerID]
		if !ok {
			continue
		}
		rank := rankOf[bye.PlayerID]
		ownValue := opts.ValueNumber(rank)

		scoresX2[idx] += byeScoreX2(bye.Type, ownValue, opts, idx, absenceCounts)
	}

	// Process absences: players who didn't play and didn't get a bye.
	for _, id := range activePlayers {
		if !playedInRound[roundIdx][id] {
			idx := playerIndex[id]

			// Late joiner: round is before they joined.
			if jr, ok := joinedRound[id]; ok && jr > 1 && round.Number < jr {
				scoresX2[idx] += lateJoinScoreX2(opts)
				continue
			}

			rank := rankOf[id]
			ownValue := opts.ValueNumber(rank)
			scoresX2[idx] += absenceScoreX2(ownValue, opts, idx, absenceCounts)
		}
	}
}

// byeScoreX2 computes the ×2 score for a bye, dispatching by bye type.
// For absence-type byes (ByeAbsent, ByeExcused), it increments the
// absence count and applies limit/decay.
func byeScoreX2(byeType chesspairing.ByeType, ownValue int, opts Options, playerIdx int, absenceCounts []int) int {
	switch byeType {
	case chesspairing.ByePAB:
		if opts.ByeFixedValue != nil {
			return fixedX2(*opts.ByeFixedValue)
		}
		return scoreX2(ownValue, *opts.ByeValueFraction)

	case chesspairing.ByeHalf:
		if opts.HalfByeFixedValue != nil {
			return fixedX2(*opts.HalfByeFixedValue)
		}
		return scoreX2(ownValue, *opts.HalfByeFraction)

	case chesspairing.ByeZero:
		if opts.ZeroByeFixedValue != nil {
			return fixedX2(*opts.ZeroByeFixedValue)
		}
		return scoreX2(ownValue, *opts.ZeroByeFraction)

	case chesspairing.ByeAbsent:
		return absenceScoreX2(ownValue, opts, playerIdx, absenceCounts)

	case chesspairing.ByeExcused:
		// Excused absences count toward the absence limit/decay.
		absenceCounts[playerIdx]++
		count := absenceCounts[playerIdx]

		if *opts.AbsenceLimit > 0 && count > *opts.AbsenceLimit {
			return 0
		}

		var s int
		if opts.ExcusedAbsentFixedValue != nil {
			s = fixedX2(*opts.ExcusedAbsentFixedValue)
		} else {
			s = scoreX2(ownValue, *opts.ExcusedAbsentFraction)
		}

		if *opts.AbsenceDecay && count > 1 {
			s >>= (count - 1)
		}
		return s

	case chesspairing.ByeClubCommitment:
		// Club commitments are NEVER subject to absence limit or decay.
		if opts.ClubCommitmentFixedValue != nil {
			return fixedX2(*opts.ClubCommitmentFixedValue)
		}
		return scoreX2(ownValue, *opts.ClubCommitmentFraction)

	default:
		// Unknown bye type: treat as unexcused absence (safe fallback).
		return absenceScoreX2(ownValue, opts, playerIdx, absenceCounts)
	}
}

// absenceScoreX2 computes the ×2 score for an unexcused absence.
// Increments the absence count and applies limit/decay.
func absenceScoreX2(ownValue int, opts Options, playerIdx int, absenceCounts []int) int {
	absenceCounts[playerIdx]++
	count := absenceCounts[playerIdx]

	// Check absence limit.
	if *opts.AbsenceLimit > 0 && count > *opts.AbsenceLimit {
		return 0
	}

	var s int
	if opts.AbsentFixedValue != nil {
		s = fixedX2(*opts.AbsentFixedValue)
	} else {
		s = scoreX2(ownValue, *opts.AbsentPenaltyFraction)
	}

	// Apply decay: halve per successive absence.
	if *opts.AbsenceDecay && count > 1 {
		s >>= (count - 1)
	}
	return s
}

// lateJoinScoreX2 computes the ×2 score for a round before the player
// joined the tournament. Uses LateJoinHandicap as a fixed value, not
// subject to absence limits or decay.
func lateJoinScoreX2(opts Options) int {
	return int(math.Round(*opts.LateJoinHandicap * 2))
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

// rankByScore returns player IDs sorted by ×2 score (descending),
// then by rating (descending) as secondary tiebreak for ranking purposes.
func rankByScore(ids []string, scoresX2 []int, entries map[string]chesspairing.PlayerEntry) []string {
	// Build index lookup.
	idIndex := make(map[string]int, len(ids))
	for i, id := range ids {
		idIndex[id] = i
	}

	ranked := make([]string, len(ids))
	copy(ranked, ids)
	sort.Slice(ranked, func(i, j int) bool {
		si := scoresX2[idIndex[ranked[i]]]
		sj := scoresX2[idIndex[ranked[j]]]
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

// buildPlayerScores converts internal ×2 scores + ranking into chesspairing.PlayerScore.
// Scores are divided by 2 to produce the final value.
func buildPlayerScores(ids []string, scoresX2 []int, ranking []string) []chesspairing.PlayerScore {
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
			Score:    float64(scoresX2[idIndex[id]]) / 2.0,
			Rank:     i + 1,
		}
	}
	return result
}
