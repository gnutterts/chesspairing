// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

// Limitations:
// - It does not derive the team set from Matches; teams must be defined via PlayerEntry.TeamID.
// - It does not pass a secondary score to the colour allocation step (AllocateColor).

import (
	"context"
	"errors"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/lexswiss"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// Pair implements chesspairing.Pairer for the Team Swiss system.
func (p *Pairer) Pair(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	originalState := state
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Honour pre-assigned byes for the upcoming round.
	state, preAssignedByes := lexswiss.FilterPreAssignedByes(state)

	result := &chesspairing.PairingResult{}

	// Build participant states (each PlayerEntry represents a team). C.04.6
	// Article 1.2 makes the configured primary component the pairing score.
	participants, err := buildParticipantStates(state, *p.opts.PrimaryScore)
	if err != nil {
		return nil, err
	}
	if len(participants) <= 1 {
		if len(participants) == 1 {
			if participants[0].PABIneligible.Any() {
				return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "team", Err: swisslib.ErrNoPABCandidate}
			}
			result.TeamByes = append(result.TeamByes, chesspairing.ByeEntry{
				PlayerID: participants[0].ID,
				Type:     chesspairing.ByePAB,
			})
		}
		if len(preAssignedByes) > 0 {
			result.TeamByes = append(preAssignedByes, result.TeamByes...)
		}
		if err := chesspairing.ValidatePairing(originalState, result); err != nil {
			if pairingErr, ok := err.(*chesspairing.PairingError); ok {
				pairingErr.System = "team"
				if pairingErr.Kind == chesspairing.PairingIncomplete {
					pairingErr.Partial = result
				}
			}
			return nil, err
		}
		return result, nil
	}

	// Resolve options.
	prefType := resolveColorPrefType(*p.opts.ColorPreferenceType)

	// Build forbidden pairs map.
	forbidden := buildForbiddenMap(p.opts.ForbiddenPairs)

	// Build participant pointer slice.
	ptrs := make([]*lexswiss.ParticipantState, len(participants))
	for i := range participants {
		ptrs[i] = &participants[i]
	}

	// Assign PAB if odd number (Art. 3.4).
	if lexswiss.NeedsBye(len(ptrs)) {
		byeTeam := assignTeamPAB(ptrs)
		if byeTeam == nil {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "team", Err: swisslib.ErrNoPABCandidate}
		}
		result.TeamByes = append(result.TeamByes, chesspairing.ByeEntry{
			PlayerID: byeTeam.ID,
			Type:     chesspairing.ByePAB,
		})
		ptrs = removeParticipant(ptrs, byeTeam)
	}

	// Build score groups.
	participantValues := make([]lexswiss.ParticipantState, len(ptrs))
	for i, ptr := range ptrs {
		participantValues[i] = *ptr
	}
	scoreGroups := lexswiss.BuildScoreGroups(participantValues)

	// Determine round context.
	isLastRound := p.opts.TotalRounds != nil && state.CurrentRound >= *p.opts.TotalRounds
	isLastTwoRounds := p.opts.TotalRounds != nil && state.CurrentRound >= *p.opts.TotalRounds-1

	// Build criteria function for C8/C9.
	criteriaFn := BuildCriteriaFunc(prefType, isLastTwoRounds, isLastRound)

	// Pair brackets from top to bottom with upfloater handling.
	allPairs, err := pairAllBrackets(ctx, scoreGroups, forbidden, criteriaFn, isLastTwoRounds)
	if err != nil {
		if errors.Is(err, lexswiss.ErrNoCompletePairing) {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingImpossible, System: "team", Err: err}
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
		wID, bID := AllocateColor(pair[0], pair[1], prefType, isLastRound, p.opts.TopSeedColor, nil)
		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   boardNum + 1,
			WhiteID: wID,
			BlackID: bID,
		})
	}

	// Sort boards: max score desc, then min TPN asc.
	sortBoards(result.Pairings, participantMap)

	if len(preAssignedByes) > 0 {
		result.TeamByes = append(preAssignedByes, result.TeamByes...)
	}

	if err := chesspairing.ValidatePairing(originalState, result); err != nil {
		if pairingErr, ok := err.(*chesspairing.PairingError); ok {
			pairingErr.System = "team"
			if pairingErr.Kind == chesspairing.PairingIncomplete {
				pairingErr.Partial = result
			}
		}
		return nil, err
	}
	return result, nil
}

// assignTeamPAB selects the team to receive the PAB per Art. 3.4:
//  1. Leaves a legal pairing for all teams (checked by caller)
//  2. Lowest score
//  3. Most matches played (highest round count)
//  4. Largest TPN
//
// This differs from lexswiss.AssignPAB which only checks score and TPN.
// Team Swiss adds "most matches played" as a tiebreaker.
func assignTeamPAB(participants []*lexswiss.ParticipantState) *lexswiss.ParticipantState {
	var eligible []*lexswiss.ParticipantState
	for _, p := range participants {
		if !p.PABIneligible.Any() {
			eligible = append(eligible, p)
		}
	}
	if len(eligible) == 0 {
		return nil
	}

	// Sort by: lowest score, most matches (longest color history), largest TPN.
	sort.SliceStable(eligible, func(i, j int) bool {
		if eligible[i].Score != eligible[j].Score {
			return eligible[i].Score < eligible[j].Score
		}
		matchesI := len(eligible[i].ColorHistory)
		matchesJ := len(eligible[j].ColorHistory)
		if matchesI != matchesJ {
			return matchesI > matchesJ // most matches first
		}
		return eligible[i].TPN > eligible[j].TPN // largest TPN first
	})

	return eligible[0]
}

// pairAllBrackets pairs all scoregroups from top to bottom, handling
// upfloaters when a bracket has an odd number of teams.
func pairAllBrackets(ctx context.Context, scoreGroups []lexswiss.ScoreGroup, forbidden map[[2]string]bool, criteriaFn lexswiss.CriteriaFunc, isLastTwoRounds bool) ([][2]*lexswiss.ParticipantState, error) {
	if len(scoreGroups) == 0 {
		return nil, nil
	}

	// Work with mutable bracket copies.
	type bracket struct {
		participants []*lexswiss.ParticipantState
		score        float64
	}
	brackets := make([]bracket, len(scoreGroups))
	upfloaters := make(map[string]bool)
	for i, sg := range scoreGroups {
		participants := make([]*lexswiss.ParticipantState, len(sg.Participants))
		copy(participants, sg.Participants)
		brackets[i] = bracket{
			participants: participants,
			score:        sg.Score,
		}
	}

	// Handle upfloaters: if a bracket has odd participants, float the
	// lowest-ranked up to the bracket above.
	for i := len(brackets) - 1; i > 0; i-- {
		if len(brackets[i].participants)%2 == 1 {
			floater := selectTeamUpfloaterWithContext(ctx, brackets[i].participants, brackets[i-1].participants, forbidden, isLastTwoRounds)
			if floater != nil {
				brackets[i].participants = removeParticipant(brackets[i].participants, floater)
				brackets[i-1].participants = append(brackets[i-1].participants, floater)
				upfloaters[floater.ID] = true
			}
		}
	}

	// Pair each bracket. If independent brackets cannot be completed, retry
	// all participants together: a legal cross-bracket pairing is preferable
	// to silently omitting participants.
	var allPairs [][2]*lexswiss.ParticipantState
	complete := true
	for _, b := range brackets {
		if len(b.participants)%2 == 1 {
			complete = false
			break
		}
		if len(b.participants) == 0 {
			continue
		}
		pairs, err := lexswiss.PairBracket(ctx, b.participants, forbidden, c10Criteria(criteriaFn, upfloaters, isLastTwoRounds))
		if errors.Is(err, lexswiss.ErrNoCompletePairing) && !isLastTwoRounds {
			// C3 takes precedence over C10 (C.04.6 Articles 2.2.1, 2.3.7).
			pairs, err = lexswiss.PairBracket(ctx, b.participants, forbidden, criteriaFn)
		}
		if err != nil {
			if !errors.Is(err, lexswiss.ErrNoCompletePairing) {
				return nil, err
			}
			complete = false
			break
		}
		allPairs = append(allPairs, pairs...)
	}
	if complete {
		return allPairs, nil
	}
	var participants []*lexswiss.ParticipantState
	for _, b := range brackets {
		participants = append(participants, b.participants...)
	}
	return lexswiss.PairBracket(ctx, participants, forbidden, criteriaFn)
}

// c10Criteria applies C.04.6 Article 2.3.7 (C10): except in the last two
// rounds, avoid pairing an upfloater with a previous-round floater. The caller
// relaxes this quality criterion when needed for C3 completion.
func c10Criteria(criteria lexswiss.CriteriaFunc, upfloaters map[string]bool, isLastTwoRounds bool) lexswiss.CriteriaFunc {
	if isLastTwoRounds || len(upfloaters) == 0 {
		return criteria
	}
	return func(first, second *lexswiss.ParticipantState) bool {
		if criteria != nil && !criteria(first, second) {
			return false
		}
		return (!upfloaters[first.ID] || !second.WasFloater) && (!upfloaters[second.ID] || !first.WasFloater)
	}
}

// selectTeamUpfloater applies C.04.6 Article 2.3.4 (C7): except in the
// last two rounds, an upfloater that was a floater in the previous round is
// avoided when C1 and C3 leave a non-floater candidate. C4-C6 are preserved
// by the bracket construction and compatible-candidate fallback.
func selectTeamUpfloater(bracket, target []*lexswiss.ParticipantState, forbidden map[[2]string]bool, isLastTwoRounds bool) *lexswiss.ParticipantState {
	return selectTeamUpfloaterWithContext(context.Background(), bracket, target, forbidden, isLastTwoRounds)
}

func selectTeamUpfloaterWithContext(ctx context.Context, bracket, target []*lexswiss.ParticipantState, forbidden map[[2]string]bool, isLastTwoRounds bool) *lexswiss.ParticipantState {
	if isLastTwoRounds {
		return lexswiss.SelectUpfloaterWithContext(ctx, bracket, target, forbidden)
	}
	return lexswiss.SelectUpfloaterWithPreferenceAndContext(ctx, bracket, target, forbidden, func(participant *lexswiss.ParticipantState) bool {
		return !participant.WasFloater
	})
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

// sortBoards sorts pairings for board ordering:
// max score of pair (desc), then min TPN of pair (asc).
func sortBoards(pairings []chesspairing.GamePairing, participants map[string]*lexswiss.ParticipantState) {
	sort.SliceStable(pairings, func(i, j int) bool {
		pi1 := participants[pairings[i].WhiteID]
		pi2 := participants[pairings[i].BlackID]
		pj1 := participants[pairings[j].WhiteID]
		pj2 := participants[pairings[j].BlackID]

		maxI := pi1.Score
		if pi2 != nil && pi2.Score > maxI {
			maxI = pi2.Score
		}
		maxJ := pj1.Score
		if pj2 != nil && pj2.Score > maxJ {
			maxJ = pj2.Score
		}
		if maxI != maxJ {
			return maxI > maxJ
		}

		minI := pi1.TPN
		if pi2 != nil && pi2.TPN < minI {
			minI = pi2.TPN
		}
		minJ := pj1.TPN
		if pj2 != nil && pj2.TPN < minJ {
			minJ = pj2.TPN
		}
		return minI < minJ
	})

	for i := range pairings {
		pairings[i].Board = i + 1
	}
}
