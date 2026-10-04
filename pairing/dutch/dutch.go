// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// ErrTooFewPlayers is returned when there aren't enough active players.
var ErrTooFewPlayers = errors.New("swiss pairing requires at least 2 active players")

// ErrNoPairingPossible is returned when no valid pairing can be found.
var ErrNoPairingPossible = errors.New("no valid pairing exists for the remaining players")

// Pair generates pairings for the next round using the FIDE Dutch system (C.04.3).
//
// Algorithm:
//  1. Build PlayerState for all active players
//  2. Build score groups (all players enter matching pool)
//  3. Global Blossom matching (PairBracketsGlobal) — includes Stage 0.5
//     completability pre-matching for bye determination with odd player count
//  4. Order boards per FIDE A.6
//  5. Allocate colors for all paired games
//  6. Unmatched player (if any) receives PAB
//  7. Return PairingResult
func (p *Pairer) Pair(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	originalState := state
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.opts.totalRoundsInvalid {
		return nil, errors.New("dutch: total rounds must be an integer")
	}
	if p.opts.TotalRounds != nil && *p.opts.TotalRounds < 1 {
		return nil, errors.New("dutch: total rounds must be at least 1")
	}
	if p.opts.TotalRounds != nil && state.CurrentRound > *p.opts.TotalRounds {
		return nil, fmt.Errorf("dutch: current round %d exceeds total rounds %d", state.CurrentRound, *p.opts.TotalRounds)
	}
	if p.opts.Acceleration != nil && *p.opts.Acceleration == "baku" && p.opts.TotalRounds == nil {
		return nil, errors.New("dutch: Baku acceleration requires total rounds")
	}

	// Honour pre-assigned byes for the upcoming round: those players are
	// excluded from the matching pool and echoed back in result.Byes.
	state, preAssignedByes := swisslib.FilterPreAssignedByes(state)

	var notes []string

	// Baku acceleration (C.04.7): group A is fixed by the participant who was
	// last in the initial group, so late entries above that player join Group A.
	var virtualPoints func(playerID string, round int) float64
	if p.opts.Acceleration != nil && *p.opts.Acceleration == "baku" {
		if err := swisslib.BakuPremise(state.ScoringConfig); err != nil {
			return nil, fmt.Errorf("dutch: %w", err)
		}
		groupA, err := swisslib.BakuGroupA(originalState.Players)
		if err != nil {
			return nil, fmt.Errorf("dutch: %w", err)
		}
		total := *p.opts.TotalRounds
		winPoints := standard.WinPoints(state.ScoringConfig.Options)
		// The pairing scores are kept on the 1-1/2-0 scale; with the premise
		// of 1.1 a win is worth 1 there, whatever the configured points.
		virtualPoints = func(playerID string, round int) float64 {
			return swisslib.BakuVirtualPoints(1, total, round, groupA[playerID])
		}
		notes = append(notes, fmt.Sprintf("Baku acceleration: GA=%d players, VP=%.1f",
			len(groupA), swisslib.BakuVirtualPoints(winPoints, total, state.CurrentRound, true)))
	}

	// Build player states.
	players, err := swisslib.BuildPlayerStatesWithVirtualPoints(state, virtualPoints)
	if err != nil {
		return nil, err
	}
	realScores := make(map[string]float64, len(players))
	for i := range players {
		realScores[players[i].ID] = players[i].Score
	}
	if virtualPoints != nil {
		swisslib.AddVirtualPoints(players, func(id string) float64 { return virtualPoints(id, state.CurrentRound) })
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

	// Handle single player.
	if len(players) == 1 {
		if players[0].PABIneligible.Any() {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "dutch", Err: swisslib.ErrNoPABCandidate}
		}
		byes := append([]chesspairing.ByeEntry{}, preAssignedByes...)
		byes = append(byes, chesspairing.ByeEntry{PlayerID: players[0].ID, Type: chesspairing.ByePAB})
		result := &chesspairing.PairingResult{
			Byes:  byes,
			Notes: []string{players[0].ID + " receives a bye (only active player)"},
		}
		if err := validateResult(originalState, result); err != nil {
			return nil, err
		}
		return result, nil
	}

	// Build active player pointers — ALL active players enter the matching pool.
	// If odd count, the Blossom matching will leave one player unmatched;
	// that player receives the PAB. This matches the FIDE algorithm where
	// the bye emerges from bracket processing, not pre-assignment.
	activePlayers := make([]*swisslib.PlayerState, len(players))
	for i := range players {
		activePlayers[i] = &players[i]
	}

	// Build player states slice (for BuildScoreGroups which takes []PlayerState).
	playerStates := make([]swisslib.PlayerState, len(activePlayers))
	for i, ap := range activePlayers {
		playerStates[i] = *ap
	}

	totalRounds := 0
	if p.opts.TotalRounds != nil {
		totalRounds = *p.opts.TotalRounds
	}
	isLastRound := p.opts.TotalRounds != nil && state.CurrentRound == totalRounds

	// Build score groups.
	scoreGroups := swisslib.BuildScoreGroups(playerStates)

	// Build criteria context.
	playerMap := make(map[string]*swisslib.PlayerState, len(activePlayers))
	for _, ap := range activePlayers {
		playerMap[ap.ID] = ap
	}

	topScorers := map[string]bool(nil)
	if isLastRound {
		topScorers = computeTopScorers(activePlayers, state.CurrentRound-1, realScores)
	}
	critCtx := &swisslib.CriteriaContext{
		Players:        playerMap,
		TotalRounds:    totalRounds,
		CurrentRound:   state.CurrentRound,
		IsLastRound:    isLastRound,
		TopScorers:     topScorers,
		ForbiddenPairs: buildForbiddenPairSet(p.opts.ForbiddenPairs),
	}

	// Global Blossom matching — mirrors bbpPairings architecture.
	// Processes score groups top-down with a single global matching graph.
	allPairs, unmatchedPlayer, pairNotes, err := swisslib.PairBracketsGlobal(ctx, scoreGroups, critCtx, swisslib.MatchingCriteria{ApplyC9: true}, playerMap)
	if err != nil {
		return nil, err
	}
	notes = append(notes, pairNotes...)

	// Order boards: pairs with higher-scoring players come first. When two
	// pairs share the same max-player score, the pair from the higher bracket
	// (homogeneous pairing) comes before a pair from a lower bracket (floater
	// pairing). Finally, ties are broken by the stronger player's TPN ascending.
	sort.SliceStable(allPairs, func(i, j int) bool {
		pi, pj := allPairs[i], allPairs[j]

		// Primary: maximum player pairing score in each pair (descending).
		maxScoreI := pi.White.PairingScore
		if pi.Black.PairingScore > maxScoreI {
			maxScoreI = pi.Black.PairingScore
		}
		maxScoreJ := pj.White.PairingScore
		if pj.Black.PairingScore > maxScoreJ {
			maxScoreJ = pj.Black.PairingScore
		}
		if maxScoreI != maxScoreJ {
			return maxScoreI > maxScoreJ
		}

		// Secondary: originating bracket score (descending).
		// Distinguishes homogeneous pairs from floater pairs at the same
		// max player score.
		if pi.BracketScore != pj.BracketScore {
			return pi.BracketScore > pj.BracketScore
		}

		// Tertiary: stronger player = lower PairingNumber (ascending).
		minTPNi := swisslib.EffectivePairingNumber(pi.White)
		if blackTPN := swisslib.EffectivePairingNumber(pi.Black); blackTPN < minTPNi {
			minTPNi = blackTPN
		}
		minTPNj := swisslib.EffectivePairingNumber(pj.White)
		if blackTPN := swisslib.EffectivePairingNumber(pj.Black); blackTPN < minTPNj {
			minTPNj = blackTPN
		}
		return minTPNi < minTPNj
	})

	parityRank := colorParityRanks(originalState, activePlayers)

	// Allocate colors and build final pairings.
	topSeedColor, err := parseTopSeedColor(p.opts.TopSeedColor)
	if err != nil {
		return nil, fmt.Errorf("dutch: %w", err)
	}
	pairings := make([]chesspairing.GamePairing, len(allPairs))
	for i, pair := range allPairs {
		white, black := *pair.White, *pair.Black
		white.PairingNumber, black.PairingNumber = parityRank[white.ID], parityRank[black.ID]
		whiteID, blackID := swisslib.AllocateColor(&white, &black, critCtx.IsLastRound, i+1, topSeedColor, swisslib.FixedNumberParity)
		pairings[i] = chesspairing.GamePairing{
			Board:   i + 1,
			WhiteID: whiteID,
			BlackID: blackID,
		}
	}

	// Build result.
	result := &chesspairing.PairingResult{
		Pairings: pairings,
		Notes:    notes,
	}

	// Pre-assigned byes come first in result.Byes; the algorithmic PAB
	// (if any) follows.
	if len(preAssignedByes) > 0 {
		result.Byes = append(result.Byes, preAssignedByes...)
	}

	if unmatchedPlayer != nil {
		if unmatchedPlayer.PABIneligible.Any() {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "dutch", Err: swisslib.ErrNoPABCandidate}
		}
		result.Byes = append(result.Byes, chesspairing.ByeEntry{PlayerID: unmatchedPlayer.ID, Type: chesspairing.ByePAB})
		result.Notes = append(result.Notes, fmt.Sprintf("%s receives PAB (bye)", unmatchedPlayer.ID))
	}

	result.Notes = append(result.Notes, "Pairings generated by Dutch Swiss system (FIDE C.04.3)")

	if err := validateResult(originalState, result); err != nil {
		return nil, err
	}
	return result, nil
}

// parseTopSeedColor converts the TopSeedColor option to a *swisslib.Color.
// It accepts TRF color synonyms case-insensitively.
func validateResult(state *chesspairing.TournamentState, result *chesspairing.PairingResult) error {
	if err := chesspairing.ValidatePairing(state, result); err != nil {
		if pairingErr, ok := err.(*chesspairing.PairingError); ok {
			pairingErr.System = "dutch"
			if pairingErr.Kind == chesspairing.PairingIncomplete {
				pairingErr.Partial = result
			}
			if pairingErr.Kind == chesspairing.PairingIncomplete && len(result.Pairings) == 0 {
				pairingErr.Kind = chesspairing.PairingImpossible
				pairingErr.Missing = nil
				pairingErr.Partial = nil
			}
		}
		return err
	}
	return nil
}

func parseTopSeedColor(opt *string) (*swisslib.Color, error) {
	if opt == nil {
		return nil, nil
	}

	switch strings.ToLower(*opt) {
	case "auto", "w", "white", "white1":
		return nil, nil
	case "b", "black", "black1":
		c := swisslib.ColorBlack
		return &c, nil
	default:
		return nil, fmt.Errorf("invalid top seed color %q", *opt)
	}
}

// buildForbiddenPairSet converts the options ForbiddenPairs slice into
// the canonicalized map format used by CriteriaContext.
func buildForbiddenPairSet(pairs [][]string) map[[2]string]bool {
	if len(pairs) == 0 {
		return nil
	}
	m := make(map[[2]string]bool, len(pairs))
	for _, pair := range pairs {
		if len(pair) == 2 {
			m[swisslib.CanonicalPairKey(pair[0], pair[1])] = true
		}
	}
	return m
}

// computeTopScorers identifies players with over 50% of the maximum possible
// score when pairing the final round (C.04.3 1.8). That maximum is the number
// of rounds played so far, so in a nine-round tournament players with 4.5 points
// or more are topscorers. Only relevant in the final round.
func computeTopScorers(players []*swisslib.PlayerState, playedRounds int, realScores map[string]float64) map[string]bool {
	threshold := float64(playedRounds) / 2.0

	topScorers := make(map[string]bool)
	for _, pl := range players {
		if realScores[pl.ID] > threshold {
			topScorers[pl.ID] = true
		}
	}
	return topScorers
}

// colorParityRanks numbers the players who have entered the tournament 1, 2, 3,
// ... in order of their fixed pairing number. When both players of a pair have
// no colour preference, the colour rule of C.04.3 5.2.5 looks at the parity of
// this number. A player has entered when taking part in this round or having
// taken part in an earlier one (a game, including a forfeit, or a pairing-
// allocated bye). A player who has not (a requested bye or an absence in every
// round so far, and in this one) is a late entry who is only given a TPN when
// they actually arrive (C.04.2 2.4, Annotated Dutch Rules 2.4 and 5.2.5); this
// is how bbpPairings numbers them in every round. Players who withdrew later
// keep their number.
func colorParityRanks(state *chesspairing.TournamentState, active []*swisslib.PlayerState) map[string]int {
	numbers := make(map[string]int, len(state.Players))
	if numbered, err := chesspairing.AssignPairingNumbers(state.Players); err == nil {
		for _, p := range numbered {
			numbers[p.ID] = p.PairingNumber
		}
	}
	entered := make(map[string]bool, len(active))
	for _, player := range active {
		entered[player.ID] = true
		if _, ok := numbers[player.ID]; !ok {
			numbers[player.ID] = swisslib.EffectivePairingNumber(player)
		}
	}
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}
	for _, round := range state.Rounds[:historyEnd] {
		for _, game := range round.Games {
			entered[game.WhiteID] = true
			entered[game.BlackID] = true
		}
		for _, bye := range round.Byes {
			if bye.Type == chesspairing.ByePAB {
				entered[bye.PlayerID] = true
			}
		}
	}
	ids := make([]string, 0, len(entered))
	for id := range entered {
		if _, ok := numbers[id]; ok {
			ids = append(ids, id)
		}
	}
	sort.SliceStable(ids, func(i, j int) bool {
		if numbers[ids[i]] != numbers[ids[j]] {
			return numbers[ids[i]] < numbers[ids[j]]
		}
		return ids[i] < ids[j]
	})
	ranks := make(map[string]int, len(ids))
	for i, id := range ids {
		ranks[id] = i + 1
	}
	return ranks
}
