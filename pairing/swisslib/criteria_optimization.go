package swisslib

// CriterionC8 implements C8 look-ahead: checks whether the downfloaters
// from this candidate allow the next bracket to be paired satisfactorily.
//
// Per FIDE C.04.3 C8: "Minimize the number of players in the round who
// receive a bye or who receive a pairing-allocated bye as a result of
// the current bracket's floaters making the next bracket impossible to pair."
//
// Algorithm:
//  1. If no LookAhead function or no RemainingBrackets, return 0 (assume OK).
//  2. Merge the candidate's floaters into the first remaining bracket.
//  3. Attempt to pair the merged bracket using C1-C7 + C10-C21 (no recursive C8).
//  4. Return 0 if a valid pairing exists, 1 if it fails.
//
// This is NOT recursive: C8 does not call itself for the simulated bracket.
// The LookAhead function is set up by the orchestrator with C8 disabled.
func CriterionC8(cand *Candidate, ctx *CriteriaContext) int {
	// No look-ahead possible: no remaining brackets or no function.
	if ctx.LookAhead == nil || len(ctx.RemainingBrackets) == 0 {
		return 0
	}

	// No floaters → nothing to check, next bracket is unaffected.
	if len(cand.Floaters) == 0 {
		return 0
	}

	// Merge floaters into the first remaining bracket.
	nextBracket := ctx.RemainingBrackets[0]
	merged := MergeIntoHeterogeneous(nextBracket, cand.Floaters)

	// Attempt to pair the merged bracket.
	if ctx.LookAhead(merged, ctx) {
		return 0 // next bracket can be paired with these floaters
	}

	return 1 // next bracket fails — penalize this candidate
}

// simulateColor determines which color each player would receive in a pair
// by calling AllocateColor and returning (assignedColorForA, assignedColorForB).
// boardNumber is set to 0 because during candidate scoring we don't know
// final board assignment — AllocateColor only uses boardNumber when both
// players have no color history (round 1 tiebreaker), which doesn't affect
// optimization criteria.
func simulateColor(a, b *PlayerState, isLastRound bool) (Color, Color) {
	whiteID, _ := AllocateColor(a, b, isLastRound, 0)
	if whiteID == a.ID {
		return ColorWhite, ColorBlack
	}
	return ColorBlack, ColorWhite
}

// colorDiffAfter computes the absolute color difference a player would have
// after receiving the given color. |whites - blacks| with the new color added.
func colorDiffAfter(p *PlayerState, assigned Color) int {
	played := filterPlayed(p.ColorHistory)
	whites, blacks := countColors(played)
	switch assigned {
	case ColorWhite:
		whites++
	case ColorBlack:
		blacks++
	}
	diff := whites - blacks
	if diff < 0 {
		diff = -diff
	}
	return diff
}

// streakAfter computes how many consecutive same-color games a player would
// have at the end of their history after receiving the given color.
func streakAfter(p *PlayerState, assigned Color) int {
	played := filterPlayed(p.ColorHistory)
	count := 1 // the new game
	for i := len(played) - 1; i >= 0; i-- {
		if played[i] == assigned {
			count++
		} else {
			break
		}
	}
	return count
}

// CriterionC10 counts topscorers and their opponents who would end up with
// |color difference| > 2 after this round. Per FIDE C.04.3 C10.
//
// Only applies when topscorers are involved (typically last round).
func CriterionC10(cand *Candidate, ctx *CriteriaContext) int {
	violations := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		colorW, colorB := simulateColor(pair.White, pair.Black, ctx.IsLastRound)

		// Check if either player is a topscorer.
		wTop := ctx.TopScorers[pair.White.ID]
		bTop := ctx.TopScorers[pair.Black.ID]
		if !wTop && !bTop {
			continue // C10 only applies when topscorers are involved
		}

		// Check each topscorer and their opponent.
		if wTop && colorDiffAfter(pair.White, colorW) > 2 {
			violations++
		}
		if bTop && colorDiffAfter(pair.Black, colorB) > 2 {
			violations++
		}
		// Also count opponent of topscorer.
		if wTop && colorDiffAfter(pair.Black, colorB) > 2 {
			violations++
		}
		if bTop && colorDiffAfter(pair.White, colorW) > 2 {
			violations++
		}
	}
	return violations
}

// CriterionC11 counts topscorers and their opponents who would end up with
// 3 or more consecutive games of the same color. Per FIDE C.04.3 C11.
func CriterionC11(cand *Candidate, ctx *CriteriaContext) int {
	violations := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		colorW, colorB := simulateColor(pair.White, pair.Black, ctx.IsLastRound)

		wTop := ctx.TopScorers[pair.White.ID]
		bTop := ctx.TopScorers[pair.Black.ID]
		if !wTop && !bTop {
			continue
		}

		if wTop && streakAfter(pair.White, colorW) >= 3 {
			violations++
		}
		if bTop && streakAfter(pair.Black, colorB) >= 3 {
			violations++
		}
		if wTop && streakAfter(pair.Black, colorB) >= 3 {
			violations++
		}
		if bTop && streakAfter(pair.White, colorW) >= 3 {
			violations++
		}
	}
	return violations
}

// CriterionC12 counts players who would not receive their color preference
// (any strength: absolute, strong, or mild). Per FIDE C.04.3 C12.
func CriterionC12(cand *Candidate, ctx *CriteriaContext) int {
	violations := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		colorW, colorB := simulateColor(pair.White, pair.Black, ctx.IsLastRound)

		prefW := ComputeColorPreference(pair.White.ColorHistory)
		prefB := ComputeColorPreference(pair.Black.ColorHistory)

		wantW := prefW.PreferredColor()
		wantB := prefB.PreferredColor()

		if wantW != nil && *wantW != colorW {
			violations++
		}
		if wantB != nil && *wantB != colorB {
			violations++
		}
	}
	return violations
}

// CriterionC13 counts players who would not receive their strong color
// preference (strong or absolute). Per FIDE C.04.3 C13.
func CriterionC13(cand *Candidate, ctx *CriteriaContext) int {
	violations := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		colorW, colorB := simulateColor(pair.White, pair.Black, ctx.IsLastRound)

		prefW := ComputeColorPreference(pair.White.ColorHistory)
		prefB := ComputeColorPreference(pair.Black.ColorHistory)

		// Check strong or absolute preference denial.
		if prefW.AbsolutePreference && *prefW.Color != colorW {
			violations++
		} else if prefW.StrongPreference && *prefW.Color != colorW {
			violations++
		}

		if prefB.AbsolutePreference && *prefB.Color != colorB {
			violations++
		} else if prefB.StrongPreference && *prefB.Color != colorB {
			violations++
		}
	}
	return violations
}

// floatAtRound returns the float direction a player had at a specific round
// (0-indexed into FloatHistory). Returns FloatNone if history is too short.
func floatAtRound(p *PlayerState, roundIdx int) Float {
	if roundIdx < 0 || roundIdx >= len(p.FloatHistory) {
		return FloatNone
	}
	return p.FloatHistory[roundIdx]
}

// CriterionC14 counts resident downfloaters in this bracket who also
// downfloated in the previous round. Per FIDE C.04.3 C14.
//
// "Resident downfloater" = player in DownfloaterIDs (S1 player floating
// down from higher bracket).
func CriterionC14(cand *Candidate, ctx *CriteriaContext) int {
	prevRoundIdx := ctx.CurrentRound - 2 // 0-indexed: round N-1
	violations := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		for _, p := range []*PlayerState{pair.White, pair.Black} {
			if cand.DownfloaterIDs[p.ID] && floatAtRound(p, prevRoundIdx) == FloatDown {
				violations++
			}
		}
	}
	return violations
}

// CriterionC15 counts MDP opponents (S2 players paired with downfloaters)
// who upfloated in the previous round. Per FIDE C.04.3 C15.
func CriterionC15(cand *Candidate, ctx *CriteriaContext) int {
	prevRoundIdx := ctx.CurrentRound - 2
	violations := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		// For each pair where one player is a downfloater, check the opponent.
		wDown := cand.DownfloaterIDs[pair.White.ID]
		bDown := cand.DownfloaterIDs[pair.Black.ID]

		if wDown && !bDown && floatAtRound(pair.Black, prevRoundIdx) == FloatUp {
			violations++
		}
		if bDown && !wDown && floatAtRound(pair.White, prevRoundIdx) == FloatUp {
			violations++
		}
	}
	return violations
}

// CriterionC16 counts resident downfloaters who downfloated 2 rounds ago.
// Per FIDE C.04.3 C16.
func CriterionC16(cand *Candidate, ctx *CriteriaContext) int {
	twoAgoIdx := ctx.CurrentRound - 3 // 0-indexed: round N-2
	violations := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		for _, p := range []*PlayerState{pair.White, pair.Black} {
			if cand.DownfloaterIDs[p.ID] && floatAtRound(p, twoAgoIdx) == FloatDown {
				violations++
			}
		}
	}
	return violations
}

// CriterionC17 counts MDP opponents who upfloated 2 rounds ago.
// Per FIDE C.04.3 C17.
func CriterionC17(cand *Candidate, ctx *CriteriaContext) int {
	twoAgoIdx := ctx.CurrentRound - 3
	violations := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		wDown := cand.DownfloaterIDs[pair.White.ID]
		bDown := cand.DownfloaterIDs[pair.Black.ID]

		if wDown && !bDown && floatAtRound(pair.Black, twoAgoIdx) == FloatUp {
			violations++
		}
		if bDown && !wDown && floatAtRound(pair.White, twoAgoIdx) == FloatUp {
			violations++
		}
	}
	return violations
}

// scoreDiffInt computes |playerScore - bracketScore| × 2 as an integer.
// Multiplied by 2 to avoid float comparison (all scores are multiples of 0.5).
func scoreDiffInt(playerScore, bracketScore float64) int {
	diff := playerScore - bracketScore
	if diff < 0 {
		diff = -diff
	}
	return int(diff * 2)
}

// CriterionC18 returns the maximum score difference (×2) of downfloaters
// who also downfloated in the previous round. Per FIDE C.04.3 C18.
//
// Returns 0 if no such downfloaters exist.
func CriterionC18(cand *Candidate, ctx *CriteriaContext) int {
	prevRoundIdx := ctx.CurrentRound - 2
	maxDiff := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		for _, p := range []*PlayerState{pair.White, pair.Black} {
			if cand.DownfloaterIDs[p.ID] && floatAtRound(p, prevRoundIdx) == FloatDown {
				diff := scoreDiffInt(p.Score, cand.BracketScore)
				if diff > maxDiff {
					maxDiff = diff
				}
			}
		}
	}
	return maxDiff
}

// CriterionC19 returns the maximum score difference (×2) of MDP opponents
// who upfloated in the previous round. Per FIDE C.04.3 C19.
func CriterionC19(cand *Candidate, ctx *CriteriaContext) int {
	prevRoundIdx := ctx.CurrentRound - 2
	maxDiff := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		wDown := cand.DownfloaterIDs[pair.White.ID]
		bDown := cand.DownfloaterIDs[pair.Black.ID]

		if wDown && !bDown && floatAtRound(pair.Black, prevRoundIdx) == FloatUp {
			diff := scoreDiffInt(pair.Black.Score, cand.BracketScore)
			if diff > maxDiff {
				maxDiff = diff
			}
		}
		if bDown && !wDown && floatAtRound(pair.White, prevRoundIdx) == FloatUp {
			diff := scoreDiffInt(pair.White.Score, cand.BracketScore)
			if diff > maxDiff {
				maxDiff = diff
			}
		}
	}
	return maxDiff
}

// CriterionC20 returns the maximum score difference (×2) of downfloaters
// who downfloated 2 rounds ago. Per FIDE C.04.3 C20.
func CriterionC20(cand *Candidate, ctx *CriteriaContext) int {
	twoAgoIdx := ctx.CurrentRound - 3
	maxDiff := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		for _, p := range []*PlayerState{pair.White, pair.Black} {
			if cand.DownfloaterIDs[p.ID] && floatAtRound(p, twoAgoIdx) == FloatDown {
				diff := scoreDiffInt(p.Score, cand.BracketScore)
				if diff > maxDiff {
					maxDiff = diff
				}
			}
		}
	}
	return maxDiff
}

// CriterionC21 returns the maximum score difference (×2) of MDP opponents
// who upfloated 2 rounds ago. Per FIDE C.04.3 C21.
func CriterionC21(cand *Candidate, ctx *CriteriaContext) int {
	twoAgoIdx := ctx.CurrentRound - 3
	maxDiff := 0
	for i := range cand.Pairs {
		pair := &cand.Pairs[i]
		wDown := cand.DownfloaterIDs[pair.White.ID]
		bDown := cand.DownfloaterIDs[pair.Black.ID]

		if wDown && !bDown && floatAtRound(pair.Black, twoAgoIdx) == FloatUp {
			diff := scoreDiffInt(pair.Black.Score, cand.BracketScore)
			if diff > maxDiff {
				maxDiff = diff
			}
		}
		if bDown && !wDown && floatAtRound(pair.White, twoAgoIdx) == FloatUp {
			diff := scoreDiffInt(pair.White.Score, cand.BracketScore)
			if diff > maxDiff {
				maxDiff = diff
			}
		}
	}
	return maxDiff
}
