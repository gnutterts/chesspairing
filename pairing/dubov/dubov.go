// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/algorithm/blossom"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

var (
	// ErrTooFewPlayers is returned when there aren't enough active players.
	ErrTooFewPlayers = errors.New("dubov pairing requires at least 1 active player")
	// ErrNoPairingPossible is returned when no valid pairing can be found.
	ErrNoPairingPossible = errors.New("no valid pairing exists for the remaining players")
	// ErrBracketTooLarge prevents factorial transposition enumeration from hanging.
	ErrBracketTooLarge = errors.New("dubov bracket too large for transposition enumeration")
)

// Pair generates a C.04.4.1 Dubov pairing.
func (p *Pairer) Pair(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := swisslib.ValidatePreAssignedPAB(state); err != nil {
		return nil, fmt.Errorf("dubov: %w", err)
	}
	originalState := state
	state, preAssignedByes := swisslib.FilterPreAssignedByes(state)
	players, err := swisslib.BuildPlayerStates(state)
	if err != nil {
		return nil, err
	}
	if p.opts.TotalRounds == nil || *p.opts.TotalRounds < 1 {
		return nil, errors.New("dubov: total rounds required")
	}
	if len(players) == 0 {
		if len(preAssignedByes) == 0 {
			return nil, ErrTooFewPlayers
		}
		result := &chesspairing.PairingResult{Byes: preAssignedByes}
		if err := validateResult(originalState, result); err != nil {
			return nil, err
		}
		return result, nil
	}
	active := make([]*swisslib.PlayerState, len(players))
	for i := range players {
		active[i] = &players[i]
	}
	mctx := newMatchContext(state, p.opts)
	mctx.ctx = ctx

	var byePlayer *swisslib.PlayerState
	var pairs []proposedPairing
	if swisslib.NeedsBye(len(active)) {
		for _, candidate := range orderedByeCandidates(active) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			remaining := without(active, candidate.ID)
			candidatePairs, pairErr := processBrackets(ctx, remaining, mctx)
			if pairErr == nil {
				byePlayer, pairs = candidate, candidatePairs
				break
			}
			if errors.Is(pairErr, ErrBracketTooLarge) || ctx.Err() != nil {
				return nil, pairErr
			}
		}
		if byePlayer == nil {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "dubov", Err: swisslib.ErrNoPABCandidate}
		}
	} else {
		pairs, err = processBrackets(ctx, active, mctx)
		if err != nil {
			if errors.Is(err, ErrBracketTooLarge) || ctx.Err() != nil {
				return nil, err
			}
			return assembleResult(originalState, pairs, nil, preAssignedByes, p.opts.TopSeedColor)
		}
	}

	return assembleResult(originalState, pairs, byePlayer, preAssignedByes, p.opts.TopSeedColor)
}

func newMatchContext(state *chesspairing.TournamentState, opts Options) *matchContext {
	return &matchContext{
		ratings:     BuildRatingMapFromState(state),
		forbidden:   buildForbiddenPairSet(opts.ForbiddenPairs),
		totalRounds: *opts.TotalRounds,
		lastRound:   state.CurrentRound == *opts.TotalRounds,
	}
}

func orderedByeCandidates(players []*swisslib.PlayerState) []*swisslib.PlayerState {
	candidates := filterNoByeReceived(players)
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score < candidates[j].Score
		}
		if swisslib.GamesPlayed(candidates[i]) != swisslib.GamesPlayed(candidates[j]) {
			return swisslib.GamesPlayed(candidates[i]) > swisslib.GamesPlayed(candidates[j])
		}
		return swisslib.EffectivePairingNumber(candidates[i]) > swisslib.EffectivePairingNumber(candidates[j])
	})
	return candidates
}

func without(players []*swisslib.PlayerState, id string) []*swisslib.PlayerState {
	result := make([]*swisslib.PlayerState, 0, len(players)-1)
	for _, player := range players {
		if player.ID != id {
			result = append(result, player)
		}
	}
	return result
}

func withoutAll(players, removed []*swisslib.PlayerState) []*swisslib.PlayerState {
	removedSet := make(map[string]bool, len(removed))
	for _, player := range removed {
		removedSet[player.ID] = true
	}
	result := make([]*swisslib.PlayerState, 0, len(players))
	for _, player := range players {
		if !removedSet[player.ID] {
			result = append(result, player)
		}
	}
	return result
}

type scoreGroups struct {
	scores  []float64
	byScore map[float64][]*swisslib.PlayerState
}

func groupPlayers(players []*swisslib.PlayerState) *scoreGroups {
	groups := &scoreGroups{byScore: make(map[float64][]*swisslib.PlayerState)}
	seen := make(map[float64]bool)
	for _, player := range players {
		score := swisslib.ScoreGroupKey(player.Score)
		if !seen[score] {
			seen[score] = true
			groups.scores = append(groups.scores, score)
		}
		groups.byScore[score] = append(groups.byScore[score], player)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(groups.scores)))
	for _, score := range groups.scores {
		sort.SliceStable(groups.byScore[score], func(i, j int) bool {
			return swisslib.EffectivePairingNumber(groups.byScore[score][i]) < swisslib.EffectivePairingNumber(groups.byScore[score][j])
		})
	}
	return groups
}

// lowerPlayers returns every player below topScore in Article 4.2 order
// (descending score, then ascending TPN).
func lowerPlayers(groups *scoreGroups, topScore float64) []*swisslib.PlayerState {
	var result []*swisslib.PlayerState
	for _, score := range groups.scores {
		if score == topScore {
			continue
		}
		result = append(result, groups.byScore[score]...)
	}
	return result
}

// processBrackets pairs the given even set of players under 3.2, pulling
// upfloaters from lower scoregroups. On failure it returns the pairs already
// committed for higher scoregroups.
func processBrackets(ctx context.Context, players []*swisslib.PlayerState, mctx *matchContext) ([]proposedPairing, error) {
	if len(players)%2 != 0 {
		return nil, ErrNoPairingPossible
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	remaining := append([]*swisslib.PlayerState(nil), players...)
	var allPairs []proposedPairing
	for len(remaining) > 0 {
		if err := ctx.Err(); err != nil {
			return allPairs, err
		}
		groups := groupPlayers(remaining)
		score := groups.scores[0]
		residents := groups.byScore[score]
		lower := lowerPlayers(groups, score)
		result, selected, err := selectBracket(ctx, residents, lower, score, mctx)
		if err != nil {
			return allPairs, err
		}
		allPairs = append(allPairs, result.pairs...)
		used := append(append([]*swisslib.PlayerState(nil), residents...), selected...)
		remaining = withoutAll(remaining, used)
	}
	return allPairs, nil
}

// completable reports whether the given players admit a complete pairing that
// satisfies the absolute criteria C1 and C3 (plus any forbidden pairs). This
// is the C4 (Article 2.2.1) existence check, independent of the Dubov bracket
// procedure: the remaining players only need a legal matching to exist.
func completable(ctx context.Context, players []*swisslib.PlayerState, forbidden map[[2]string]bool) (bool, error) {
	if len(players) == 0 {
		return true, nil
	}
	var edges []blossom.BlossomEdge
	for i := 0; i < len(players); i++ {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		for j := i + 1; j < len(players); j++ {
			if SatisfiesAbsoluteCriteria(players[i], players[j], forbidden) {
				edges = append(edges, blossom.BlossomEdge{I: i, J: j, Weight: 1})
			}
		}
	}
	if len(edges) == 0 {
		return false, nil
	}
	match := blossom.MaxWeightMatching(edges, true)
	if err := ctx.Err(); err != nil {
		return false, err
	}
	for i := range players {
		if i >= len(match) || match[i] == -1 {
			return false, nil
		}
	}
	return true, nil
}

func selectBracket(ctx context.Context, residents, lower []*swisslib.PlayerState, score float64, mctx *matchContext) (*bracketResult, []*swisslib.PlayerState, error) {
	for count := len(residents) % 2; count <= len(lower); count += 2 {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		// Article 4.2.3: C6 fixes the upfloaters' scores before 4.2 order
		// decides between sets. Sets with equal score multisets are not
		// contiguous in 4.2 order and multisets do not arrive in C6 order, so
		// list the multisets of this count, sort them by C6, and visit only
		// the sets of the best multiset first, still in 4.2 order.
		multisets, err := scoreMultisets(lower, count)
		if err != nil {
			return nil, nil, err
		}
		visited := 0
		for _, scores := range multisets {
			var sets [][]*swisslib.PlayerState
			_, err := forEachSetWithScores(lower, scores, func(set []*swisslib.PlayerState) (bool, error) {
				if err := ctx.Err(); err != nil {
					return false, err
				}
				visited++
				if visited > maxCombinations {
					return false, ErrBracketTooLarge
				}
				sets = append(sets, set)
				return false, nil
			})
			if err != nil {
				return nil, nil, err
			}
			best, bestSet, err := evaluateUpfloaterSets(ctx, sets, residents, lower, score, mctx)
			if err != nil {
				return nil, nil, err
			}
			if best != nil {
				return best, bestSet, nil
			}
		}
	}
	return nil, nil, ErrNoPairingPossible
}

// scoreMultisets lists every score multiset (ascending) of count players from
// lower, best C6 multiset first. It returns ErrBracketTooLarge when there are
// more than maxCombinations multisets.
func scoreMultisets(lower []*swisslib.PlayerState, count int) ([][]float64, error) {
	var levels []float64
	available := make(map[float64]int)
	for _, player := range lower {
		if available[player.Score] == 0 {
			levels = append(levels, player.Score)
		}
		available[player.Score]++
	}
	sort.Float64s(levels)
	var result [][]float64
	chosen := make([]float64, 0, count)
	var walk func(level int) bool
	walk = func(level int) bool {
		if len(chosen) == count {
			result = append(result, append([]float64(nil), chosen...))
			return len(result) <= maxCombinations
		}
		if level == len(levels) {
			return true
		}
		for take := 0; take <= available[levels[level]] && len(chosen)+take <= count; take++ {
			for range take {
				chosen = append(chosen, levels[level])
			}
			ok := walk(level + 1)
			chosen = chosen[:len(chosen)-take]
			if !ok {
				return false
			}
		}
		return true
	}
	if !walk(0) {
		return nil, ErrBracketTooLarge
	}
	sort.SliceStable(result, func(i, j int) bool {
		return betterScoreMultiset(result[i], result[j])
	})
	return result, nil
}

// forEachSetWithScores visits, in combination (4.2) order, the sets of lower
// whose ascending score multiset equals scores.
func forEachSetWithScores(lower []*swisslib.PlayerState, scores []float64, visit func([]*swisslib.PlayerState) (bool, error)) (bool, error) {
	need := make(map[float64]int)
	for _, score := range scores {
		need[score]++
	}
	// remaining[i][score] counts the players from index i on with that score.
	remaining := make([]map[float64]int, len(lower)+1)
	remaining[len(lower)] = map[float64]int{}
	for i := len(lower) - 1; i >= 0; i-- {
		counts := make(map[float64]int, len(need))
		for score := range need {
			counts[score] = remaining[i+1][score]
		}
		if _, ok := need[lower[i].Score]; ok {
			counts[lower[i].Score]++
		}
		remaining[i] = counts
	}
	selected := make([]*swisslib.PlayerState, 0, len(scores))
	var walk func(int) (bool, error)
	walk = func(start int) (bool, error) {
		if len(selected) == len(scores) {
			return visit(append([]*swisslib.PlayerState(nil), selected...))
		}
		for score, n := range need {
			if remaining[start][score] < n {
				return false, nil
			}
		}
		for i := start; i < len(lower); i++ {
			if need[lower[i].Score] == 0 {
				continue
			}
			need[lower[i].Score]--
			selected = append(selected, lower[i])
			stop, err := walk(i + 1)
			selected = selected[:len(selected)-1]
			need[lower[i].Score]++
			if err != nil || stop {
				return stop, err
			}
		}
		return false, nil
	}
	return walk(0)
}

// betterScoreMultiset reports whether left is a better C6 score multiset than
// right: the ascending score vectors are compared lexicographically, larger
// first (Article 2.3.2: "maximise the scores (taken in ascending order) of
// the upfloaters").
func betterScoreMultiset(left, right []float64) bool {
	for i := range left {
		if left[i] != right[i] {
			return left[i] > right[i]
		}
	}
	return false
}

func evaluateUpfloaterSets(ctx context.Context, sets [][]*swisslib.PlayerState, residents, lower []*swisslib.PlayerState, score float64, mctx *matchContext) (*bracketResult, []*swisslib.PlayerState, error) {
	var best *bracketResult
	var bestSet []*swisslib.PlayerState
	for _, set := range sets {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		bracketPlayers := append(append([]*swisslib.PlayerState(nil), residents...), set...)
		upfloaterIDs := make(map[string]bool, len(set))
		for _, player := range set {
			upfloaterIDs[player.ID] = true
		}
		result, err := matchBracket(swisslib.Bracket{Players: bracketPlayers, OriginalScore: score}, mctx)
		if err != nil {
			return nil, nil, err
		}
		if result == nil {
			continue
		}
		complete, err := completable(ctx, withoutAll(lower, set), mctx.forbidden)
		if err != nil {
			return nil, nil, err
		}
		if !complete {
			continue
		}
		result.quality = qualityFor(result.pairs, upfloaterIDs, mctx.lastRound, MaxT(mctx.totalRounds))
		if best == nil || betterC6ToC10(result.quality, best.quality) {
			best, bestSet = result, set
		}
	}
	return best, bestSet, nil
}

func betterC6ToC10(left, right candidateQuality) bool {
	// Equal score multisets can still yield different C6 vectors when
	// upfloaters pair each other, so retain C6 before C7-C10.
	n := len(left.differences)
	if len(right.differences) < n {
		n = len(right.differences)
	}
	for i := 0; i < n; i++ {
		if left.differences[i] != right.differences[i] {
			return left.differences[i] < right.differences[i]
		}
	}
	if len(left.differences) != len(right.differences) {
		return len(left.differences) < len(right.differences)
	}
	if left.colorViolations != right.colorViolations {
		return left.colorViolations < right.colorViolations
	}
	if left.maximumUpfloaters != right.maximumUpfloaters {
		return left.maximumUpfloaters < right.maximumUpfloaters
	}
	if left.maximumTimes != right.maximumTimes {
		return left.maximumTimes < right.maximumTimes
	}
	return left.previousUpfloaters < right.previousUpfloaters
}

func assembleResult(originalState *chesspairing.TournamentState, pairs []proposedPairing, byePlayer *swisslib.PlayerState, preAssignedByes []chesspairing.ByeEntry, topSeedColorOpt *string) (*chesspairing.PairingResult, error) {
	sortBoardOrder(pairs)
	topSeedColor := parseTopSeedColor(topSeedColorOpt)
	pairings := make([]chesspairing.GamePairing, len(pairs))
	for i, pair := range pairs {
		whiteID, blackID := AllocateColor(pair.white, pair.black, i+1, topSeedColor)
		pairings[i] = chesspairing.GamePairing{Board: i + 1, WhiteID: whiteID, BlackID: blackID}
	}
	result := &chesspairing.PairingResult{Pairings: pairings, Byes: append([]chesspairing.ByeEntry(nil), preAssignedByes...)}
	if byePlayer != nil {
		result.Byes = append(result.Byes, chesspairing.ByeEntry{PlayerID: byePlayer.ID, Type: chesspairing.ByePAB})
	}
	result.Notes = append(result.Notes, "Pairings generated by Dubov Swiss system (C.04.4.1)")
	if err := validateResult(originalState, result); err != nil {
		return nil, err
	}
	return result, nil
}

func validateResult(state *chesspairing.TournamentState, result *chesspairing.PairingResult) error {
	if err := chesspairing.ValidatePairing(state, result); err != nil {
		if pairingErr, ok := err.(*chesspairing.PairingError); ok {
			pairingErr.System = "dubov"
			if pairingErr.Kind == chesspairing.PairingIncomplete {
				pairingErr.Partial = result
			}
		}
		return err
	}
	return nil
}

func parseTopSeedColor(opt *string) *swisslib.Color {
	if opt == nil || *opt == "auto" || *opt == "white" {
		return nil
	}
	if *opt == "black" {
		color := swisslib.ColorBlack
		return &color
	}
	return nil
}

func buildForbiddenPairSet(pairs [][]string) map[[2]string]bool {
	result := make(map[[2]string]bool, len(pairs))
	for _, pair := range pairs {
		if len(pair) == 2 {
			result[swisslib.CanonicalPairKey(pair[0], pair[1])] = true
		}
	}
	return result
}

// sortBoardOrder sorts pairs under C.04.2 article 3.6.
func sortBoardOrder(pairs []proposedPairing) {
	sort.SliceStable(pairs, func(i, j int) bool {
		a, b := pairs[i], pairs[j]
		return swisslib.BoardBefore(swisslib.BoardPlayer(a.white), swisslib.BoardPlayer(a.black),
			swisslib.BoardPlayer(b.white), swisslib.BoardPlayer(b.black))
	})
}
