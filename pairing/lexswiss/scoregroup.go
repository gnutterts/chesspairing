// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import "sort"

// ScoreGroup holds all participants with the same pairing score.
// Participants are ordered by TPN ascending within the group.
type ScoreGroup struct {
	Score        float64
	Participants []*ParticipantState // ordered by TPN ascending
}

// BuildScoreGroups creates score groups from participant states.
// Participants are grouped by Score and each group is sorted by TPN ascending.
// Returns groups in descending score order. Input need not be pre-sorted.
func BuildScoreGroups(participants []ParticipantState) []ScoreGroup {
	if len(participants) == 0 {
		return nil
	}

	// Group by score.
	groups := make(map[float64][]*ParticipantState)
	var scores []float64
	seen := make(map[float64]bool)

	for i := range participants {
		p := &participants[i]
		if !seen[p.Score] {
			seen[p.Score] = true
			scores = append(scores, p.Score)
		}
		groups[p.Score] = append(groups[p.Score], p)
	}

	// Sort scores descending.
	sort.Float64s(scores)
	for i, j := 0, len(scores)-1; i < j; i, j = i+1, j-1 {
		scores[i], scores[j] = scores[j], scores[i]
	}

	// Build result with participants sorted by TPN within each group.
	result := make([]ScoreGroup, 0, len(scores))
	for _, score := range scores {
		participantList := groups[score]
		sort.Slice(participantList, func(i, j int) bool {
			return participantList[i].TPN < participantList[j].TPN
		})
		result = append(result, ScoreGroup{
			Score:        score,
			Participants: participantList,
		})
	}

	return result
}
