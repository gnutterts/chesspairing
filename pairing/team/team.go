// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

// Pair implements chesspairing.Pairer for the Team Swiss system.
func (p *Pairer) Pair(_ context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	result := &chesspairing.PairingResult{}

	// Build participant states (each PlayerEntry represents a team).
	participants := lexswiss.BuildParticipantStates(state)
	if len(participants) <= 1 {
		if len(participants) == 1 {
			result.Byes = append(result.Byes, chesspairing.ByeEntry{
				PlayerID: participants[0].ID,
				Type:     chesspairing.ByePAB,
			})
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
		if byeTeam != nil {
			result.Byes = append(result.Byes, chesspairing.ByeEntry{
				PlayerID: byeTeam.ID,
				Type:     chesspairing.ByePAB,
			})
			ptrs = removeParticipant(ptrs, byeTeam)
		}
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
	allPairs := pairAllBrackets(scoreGroups, forbidden, criteriaFn)

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
		if !p.ByeReceived {
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
func pairAllBrackets(scoreGroups []lexswiss.ScoreGroup, forbidden map[[2]string]bool, criteriaFn lexswiss.CriteriaFunc) [][2]*lexswiss.ParticipantState {
	if len(scoreGroups) == 0 {
		return nil
	}

	// Work with mutable bracket copies.
	type bracket struct {
		participants []*lexswiss.ParticipantState
		score        float64
	}
	brackets := make([]bracket, len(scoreGroups))
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
			floater := lexswiss.SelectUpfloater(brackets[i].participants, brackets[i-1].participants, forbidden)
			if floater != nil {
				brackets[i].participants = removeParticipant(brackets[i].participants, floater)
				brackets[i-1].participants = append(brackets[i-1].participants, floater)
			}
		}
	}

	// Pair each bracket.
	var allPairs [][2]*lexswiss.ParticipantState
	for _, b := range brackets {
		if len(b.participants) < 2 {
			continue
		}
		pairs := lexswiss.PairBracket(b.participants, forbidden, criteriaFn)
		allPairs = append(allPairs, pairs...)
	}

	return allPairs
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
