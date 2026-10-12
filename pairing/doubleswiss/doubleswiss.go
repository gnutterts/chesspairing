// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/lexswiss"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// Pair implements chesspairing.Pairer for the Double-Swiss system.
func (p *Pairer) Pair(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := lexswiss.ValidatePreAssignedPAB(state); err != nil {
		return nil, fmt.Errorf("doubleswiss: %w", err)
	}
	originalState := state
	// Honour pre-assigned byes for the upcoming round.
	state, preAssignedByes := lexswiss.FilterPreAssignedByes(state)

	result := &chesspairing.PairingResult{}

	// C.04.5 Article 1.4: a PAB scores a win and a draw (1.5 points).
	pabPoints := 1.5
	if p.opts.PABPoints != nil {
		pabPoints = *p.opts.PABPoints
	}
	participants, err := lexswiss.BuildParticipantStatesWithPAB(state, pabPoints)
	if err != nil {
		return nil, err
	}
	if len(participants) <= 1 {
		if len(participants) == 1 {
			if participants[0].PABIneligible.Any() {
				return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "doubleswiss", Err: swisslib.ErrNoPABCandidate}
			}
			result.Byes = append(result.Byes, chesspairing.ByeEntry{
				PlayerID: participants[0].ID,
				Type:     chesspairing.ByePAB,
			})
		}
		if len(preAssignedByes) > 0 {
			result.Byes = append(preAssignedByes, result.Byes...)
		}
		if err := chesspairing.ValidatePairing(originalState, result); err != nil {
			if pairingErr, ok := err.(*chesspairing.PairingError); ok {
				pairingErr.System = "doubleswiss"
				if pairingErr.Kind == chesspairing.PairingIncomplete {
					pairingErr.Partial = result
				}
			}
			return nil, err
		}
		return result, nil
	}

	// Build forbidden pairs map.
	forbidden := buildForbiddenMap(p.opts.ForbiddenPairs)

	// Build participant pointer slice.
	ptrs := make([]*lexswiss.ParticipantState, len(participants))
	for i := range participants {
		ptrs[i] = &participants[i]
	}

	// Assign PAB if odd number.
	if lexswiss.NeedsBye(len(ptrs)) {
		byePlayer, err := lexswiss.AssignPAB(ctx, ptrs, forbidden)
		if err != nil {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingImpossible, System: "doubleswiss", Err: err}
		}
		if byePlayer == nil {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "doubleswiss", Err: swisslib.ErrNoPABCandidate}
		}
		result.Byes = append(result.Byes, chesspairing.ByeEntry{
			PlayerID: byePlayer.ID,
			Type:     chesspairing.ByePAB,
		})
		ptrs = removeParticipant(ptrs, byePlayer)
	}

	// Build score groups.
	participantValues := make([]lexswiss.ParticipantState, len(ptrs))
	for i, ptr := range ptrs {
		participantValues[i] = *ptr
	}
	scoreGroups := lexswiss.BuildScoreGroups(participantValues)

	// Determine if this is the last round (for criteria relaxation).
	isLastRound := p.opts.TotalRounds != nil && state.CurrentRound >= *p.opts.TotalRounds

	// Pair brackets from top to bottom with upfloater handling.
	allPairs, err := pairAllBrackets(ctx, scoreGroups, forbidden, isLastRound)
	if err != nil {
		if errors.Is(err, lexswiss.ErrNoCompletePairing) || errors.Is(err, lexswiss.ErrPairingSearchLimit) {
			partial := partialResult(result, allPairs, p.opts.TopSeedColor)
			if len(preAssignedByes) > 0 {
				partial.Byes = append(preAssignedByes, partial.Byes...)
			}
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingImpossible, System: "doubleswiss", Err: err, Partial: partial}
		}
		return nil, err
	}

	// Build participant map for lookups.
	participantMap := make(map[string]*lexswiss.ParticipantState, len(ptrs))
	for _, ptr := range ptrs {
		participantMap[ptr.ID] = ptr
	}

	// Allocate colours and build final pairings.
	for boardNum, pair := range allPairs {
		wID, bID := AllocateColor(pair[0], pair[1], p.opts.TopSeedColor)
		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   boardNum + 1,
			WhiteID: wID,
			BlackID: bID,
		})
	}

	// Sort boards under C.04.2 article 3.6.
	sortBoards(result.Pairings, participantMap)

	if len(preAssignedByes) > 0 {
		result.Byes = append(preAssignedByes, result.Byes...)
	}

	if err := chesspairing.ValidatePairing(originalState, result); err != nil {
		if pairingErr, ok := err.(*chesspairing.PairingError); ok {
			pairingErr.System = "doubleswiss"
			if pairingErr.Kind == chesspairing.PairingIncomplete {
				pairingErr.Partial = result
			}
		}
		return nil, err
	}
	return result, nil
}

// partialResult converts brackets completed before a failed top-down search to
// a valid partial pairing result.
func partialResult(result *chesspairing.PairingResult, pairs [][2]*lexswiss.ParticipantState, topSeedColor *string) *chesspairing.PairingResult {
	for boardNum, pair := range pairs {
		whiteID, blackID := AllocateColor(pair[0], pair[1], topSeedColor)
		result.Pairings = append(result.Pairings, chesspairing.GamePairing{Board: boardNum + 1, WhiteID: whiteID, BlackID: blackID})
	}
	return result
}

// pairAllBrackets applies the top-down bracket process in C.04.5 Article 3.3.2.
func pairAllBrackets(ctx context.Context, scoreGroups []lexswiss.ScoreGroup, forbidden map[[2]string]bool, isLastRound bool) ([][2]*lexswiss.ParticipantState, error) {
	options := lexswiss.TopDownOptions{
		RelaxCriteriaForC3: true,
		ApplyC7:            !isLastRound,
	}
	if !isLastRound {
		options.CostForBracket = c8CostForBracket
	}
	return lexswiss.PairTopDown(ctx, scoreGroups, forbidden, options)
}

// c8CostForBracket counts upfloaters whose opponent was a floater in the
// preceding round, as required by Article 2.3.5.
func c8CostForBracket(upfloaters []*lexswiss.ParticipantState) lexswiss.PairCostFunc {
	upfloaterIDs := make(map[string]bool, len(upfloaters))
	for _, participant := range upfloaters {
		upfloaterIDs[participant.ID] = true
	}
	return func(a, b *lexswiss.ParticipantState) int {
		cost := 0
		if upfloaterIDs[a.ID] && b.WasFloater {
			cost++
		}
		if upfloaterIDs[b.ID] && a.WasFloater {
			cost++
		}
		return cost
	}
}

// buildForbiddenMap builds a lookup map from forbidden pair slices.
func buildForbiddenMap(pairs [][]string) map[[2]string]bool {
	if len(pairs) == 0 {
		return nil
	}
	m := make(map[[2]string]bool, len(pairs)*2)
	for _, pair := range pairs {
		if len(pair) == 2 {
			m[[2]string{pair[0], pair[1]}] = true
			m[[2]string{pair[1], pair[0]}] = true
		}
	}
	return m
}

// removeParticipant removes a specific participant from the pointer slice.
func removeParticipant(participants []*lexswiss.ParticipantState, remove *lexswiss.ParticipantState) []*lexswiss.ParticipantState {
	result := make([]*lexswiss.ParticipantState, 0, len(participants)-1)
	for _, p := range participants {
		if p.ID != remove.ID {
			result = append(result, p)
		}
	}
	return result
}

// sortBoards sorts pairings under C.04.2 article 3.6.
func sortBoards(pairings []chesspairing.GamePairing, participants map[string]*lexswiss.ParticipantState) {
	sort.SliceStable(pairings, func(i, j int) bool {
		pi1 := participants[pairings[i].WhiteID]
		pi2 := participants[pairings[i].BlackID]
		pj1 := participants[pairings[j].WhiteID]
		pj2 := participants[pairings[j].BlackID]

		return swisslib.BoardBefore(lexswiss.BoardPlayer(pi1), lexswiss.BoardPlayer(pi2),
			lexswiss.BoardPlayer(pj1), lexswiss.BoardPlayer(pj2))
	})

	for i := range pairings {
		pairings[i].Board = i + 1
	}
}
