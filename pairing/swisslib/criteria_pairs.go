package swisslib

import "math/big"

// Per-pair edge weight computation for the global Blossom matching.
//
// This replicates bbpPairings' computeEdgeWeight bit layout EXACTLY,
// using math/big.Int for multi-precision integers. bbpPairings uses
// DynamicUint; we use *big.Int which provides the same capabilities.
//
// KEY: sgBits = scoreGroupSizeBits = bitsToRepresent(maxScoreGroupSize).
// scoreGroupsShift = sum of bitsToRepresent(sgSize) for each score group
// (accumulated LOW→HIGH). scoreGroupShifts[score] = bit offset for that SG.
//
// The bit layout (HIGH → LOW) mirrors bbpPairings field ordering:
//
//   1. Bye eligibility (2 bits)
//   2. Maximize pairs in current bracket (sgBits)
//   3. Maximize scores in current bracket (scoreGroupsShift)
//   4. Maximize pairs in next bracket (sgBits)
//   5. Maximize scores in next bracket (scoreGroupsShift)
//   6. Bye assignee unplayed games (2 × sgBits)
//   7. Color preference satisfaction (4 × sgBits)
//   8. C14: downfloat repeat R-1 (sgBits)
//   9. C15: upfloat repeat R-1 (sgBits)
//  10. C16: downfloat repeat R-2 (sgBits, conditional)
//  11. C17: upfloat repeat R-2 (sgBits, conditional)
//  12. C18: downfloat score R-1 (scoreGroupsShift)
//  13. C19: upfloat opp score R-1 (scoreGroupsShift)
//  14. C20: downfloat score R-2 (scoreGroupsShift, conditional)
//  15. C21: upfloat opp score R-2 (scoreGroupsShift, conditional)
//  16. Reserve for edgeWeightComputer (3×sgBits + 1)
//
// All fields use POSITIVE logic: bit=1 means NO violation (higher = better),
// matching bbpPairings exactly.

// EdgeWeightParams holds precomputed parameters for edge weight computation.
// Computed once before the main loop in pairBracketsGlobal.
type EdgeWeightParams struct {
	// ScoreGroupSizeBits = bitsToRepresent(maxScoreGroupSize).
	// Used for boolean field widths AND reserve bits at the bottom.
	ScoreGroupSizeBits int

	// ScoreGroupsShift is the total width (in bits) of a score-indexed
	// field. Computed as the sum of bitsToRepresent(sgSize) for each
	// score group, iterated LOW→HIGH (matching bbpPairings lines 685-712).
	ScoreGroupsShift int

	// ScoreGroupShifts maps score → per-SG bit offset within a
	// scoreGroupsShift-wide field. Each SG's offset is the cumulative
	// width of all lower-scoring SGs. Matches bbpPairings' scoreGroupShifts.
	ScoreGroupShifts map[float64]int

	// PlayedRounds is the number of rounds already played.
	PlayedRounds int

	// ByeAssigneeScore is the score of the player determined to receive the
	// bye by the completability pre-matching. Used by isByeCandidate: a
	// player is a bye candidate if eligibleForBye AND score <= ByeAssigneeScore.
	// Set to -1 when even player count (no bye needed), which makes
	// isByeCandidate always false.
	ByeAssigneeScore float64

	// IsSingleDownfloaterTheByeAssignee is true when the bye assignee is
	// the single downfloater in the top bracket. When true, C9 (minimize
	// unplayed games of bye assignee) takes effect.
	IsSingleDownfloaterTheByeAssignee bool

	// UnplayedGameRanks maps playedGames count → rank (0-based, sorted by
	// most played games first). Used for C9 when IsSingleDownfloaterTheByeAssignee.
	UnplayedGameRanks map[int]int

	// ReserveBits = 3*ScoreGroupSizeBits + 1 (matches bbpPairings' reserve
	// for edgeWeightComputer addend).
	ReserveBits int

	// TotalBits is the total width of the edge weight.
	TotalBits int
}

// ComputeEdgeWeightParams builds EdgeWeightParams from sorted score groups
// (highest score first). This mirrors bbpPairings' computeMatching setup
// (lines ~685-715) where it iterates from LOWEST to HIGHEST score to
// compute scoreGroupShifts, using each SG's actual size to determine its
// bit width within score-indexed fields.
func ComputeEdgeWeightParams(scoreGroups []ScoreGroup, playedRounds int) EdgeWeightParams {
	numSGs := len(scoreGroups)
	if numSGs == 0 {
		return EdgeWeightParams{
			ScoreGroupSizeBits: 1,
			ScoreGroupsShift:   1,
			ScoreGroupShifts:   map[float64]int{},
			PlayedRounds:       playedRounds,
			ReserveBits:        4,
			TotalBits:          20,
		}
	}

	// Iterate LOW→HIGH (scoreGroups are sorted HIGH→LOW, so reverse).
	// For each SG, compute bitsToRepresent(sgSize) and accumulate into
	// scoreGroupsShift. Each SG's offset within a score-indexed field
	// is the cumulative width of all lower-scoring SGs.
	// This matches bbpPairings lines 685-712 exactly.
	maxSGSize := 0
	sgShifts := make(map[float64]int, numSGs)
	scoreGroupsShift := 0

	for i := numSGs - 1; i >= 0; i-- {
		sg := scoreGroups[i]
		sgSize := len(sg.Players)
		if sgSize > maxSGSize {
			maxSGSize = sgSize
		}
		sgShifts[sg.Score] = scoreGroupsShift
		newBits := bitsToRepresent(sgSize)
		scoreGroupsShift += newBits
	}

	sgSizeBits := bitsToRepresent(maxSGSize)
	if sgSizeBits < 1 {
		sgSizeBits = 1
	}
	if scoreGroupsShift < 1 {
		scoreGroupsShift = 1
	}

	reserveBits := 3*sgSizeBits + 1

	// Count total bits — matching bbpPairings' field widths exactly.
	// Boolean fields: sgSizeBits wide (NOT 1 bit).
	// Score-indexed fields: scoreGroupsShift wide (NOT bitsToRepresent(numSGs)).
	totalBits := reserveBits
	// C20, C21 (conditional on playedRounds > 1)
	if playedRounds > 1 {
		totalBits += scoreGroupsShift // C21: upfloat opp score 2 ago
		totalBits += scoreGroupsShift // C20: downfloat score 2 ago
	}
	// C17, C16 (conditional on playedRounds > 1)
	if playedRounds > 1 {
		totalBits += sgSizeBits // C17: upfloat repeat 2 ago
		totalBits += sgSizeBits // C16: downfloat repeat 2 ago
	}
	// C19, C18 (conditional on playedRounds > 0)
	if playedRounds > 0 {
		totalBits += scoreGroupsShift // C19: upfloat opp score R-1
		totalBits += scoreGroupsShift // C18: downfloat score R-1
	}
	// C15, C14 (conditional on playedRounds > 0)
	if playedRounds > 0 {
		totalBits += sgSizeBits // C15: upfloat repeat R-1
		totalBits += sgSizeBits // C14: downfloat repeat R-1
	}
	totalBits += 4 * sgSizeBits   // 4 color bits
	totalBits += 2 * sgSizeBits   // bye unplayed (2 × sgSizeBits)
	totalBits += scoreGroupsShift // scores in next
	totalBits += sgSizeBits       // pairs in next
	totalBits += scoreGroupsShift // scores in current
	totalBits += sgSizeBits       // pairs in current
	totalBits += 2                // bye eligibility

	return EdgeWeightParams{
		ScoreGroupSizeBits: sgSizeBits,
		ScoreGroupsShift:   scoreGroupsShift,
		ScoreGroupShifts:   sgShifts,
		PlayedRounds:       playedRounds,
		ByeAssigneeScore:   -1, // Default: no bye. Set by completability pre-matching for odd player counts.
		ReserveBits:        reserveBits,
		TotalBits:          totalBits,
	}
}

// ComputeBaseEdgeWeight computes the Blossom edge weight for a pair of
// players using math/big.Int multi-precision integers. This mirrors
// bbpPairings' computeEdgeWeight EXACTLY, using the same bit widths:
// - Boolean fields: scoreGroupSizeBits wide
// - Score-indexed fields: scoreGroupsShift wide, positioned at scoreGroupShifts[score]
//
// higherPlayer = player with higher score (smaller index in sorted array).
// lowerPlayer = player with lower score (larger index).
// inCurrentBracket = lowerPlayer is in the current score group.
// inNextBracket = lowerPlayer is in the next score group.
//
// Returns a zero big.Int if the pair is incompatible (already played or
// absolute color conflict). The caller handles C1/C3 checks before calling this.
func ComputeBaseEdgeWeight(
	higherPlayer, lowerPlayer *PlayerState,
	inCurrentBracket, inNextBracket bool,
	params *EdgeWeightParams,
) *big.Int {
	sgBits := params.ScoreGroupSizeBits
	sgsShift := params.ScoreGroupsShift
	reserveBits := params.ReserveBits

	result := new(big.Int)
	shift := 0

	// Helper: set bit at position shift+offset in result.
	setBit := func(pos int) {
		result.SetBit(result, pos, 1)
	}
	// Helper: create a big.Int with value 1 shifted left by pos bits.
	one := new(big.Int).SetInt64(1)

	// === Reserve bits (bottom) ===
	// Filled by edgeWeightComputer in Phase 3, not here.
	shift += reserveBits

	// === C20/C21: downfloat/upfloat scores 2 rounds ago (conditional) ===
	if params.PlayedRounds > 1 {
		// C21: upfloat opponent score 2 ago
		if inCurrentBracket {
			lowerFloat2 := floatAtRound(lowerPlayer, params.PlayedRounds-2)
			if !(lowerFloat2 == FloatUp &&
				higherPlayer.Score > lowerPlayer.Score+0.001) {
				sgShift := params.scoreGroupShift(higherPlayer.Score)
				setBit(shift + sgShift)
			}
		}
		shift += sgsShift

		// C20: downfloat scores 2 ago
		if inCurrentBracket {
			lowerFloat2 := floatAtRound(lowerPlayer, params.PlayedRounds-2)
			higherFloat2 := floatAtRound(higherPlayer, params.PlayedRounds-2)
			if lowerFloat2 == FloatDown {
				sgShift := params.scoreGroupShift(lowerPlayer.Score)
				addend := new(big.Int).Lsh(one, uint(max(shift+sgShift, 0))) //nolint:gosec // shift values are bounded by tournament size
				result.Add(result, addend)
			}
			if higherFloat2 == FloatDown {
				sgShift := params.scoreGroupShift(higherPlayer.Score)
				addend := new(big.Int).Lsh(one, uint(max(shift+sgShift, 0))) //nolint:gosec // shift values are bounded by tournament size
				result.Add(result, addend)
			}
		}
		shift += sgsShift
	}

	// === C16/C17: downfloat/upfloat repeat 2 rounds ago (conditional) ===
	if params.PlayedRounds > 1 {
		// C17: upfloat repeat 2 ago
		if inCurrentBracket {
			lowerFloat2 := floatAtRound(lowerPlayer, params.PlayedRounds-2)
			c17 := !(higherPlayer.Score > lowerPlayer.Score+0.001 &&
				lowerFloat2 == FloatUp)
			if c17 {
				setBit(shift)
			}
		}
		shift += sgBits

		// C16: downfloat repeat 2 ago (value 0-2, occupies sgBits)
		if inCurrentBracket {
			lowerFloat2 := floatAtRound(lowerPlayer, params.PlayedRounds-2)
			higherFloat2 := floatAtRound(higherPlayer, params.PlayedRounds-2)
			// bbpPairings: result |= (lowerFloatDown); result += (higherFloatDown && sameScore)
			if lowerFloat2 == FloatDown {
				setBit(shift)
			}
			if higherPlayer.Score <= lowerPlayer.Score+0.001 &&
				higherFloat2 == FloatDown {
				addend := new(big.Int).Lsh(one, uint(max(shift, 0))) //nolint:gosec // shift values are bounded by tournament size
				result.Add(result, addend)
			}
		}
		shift += sgBits
	}

	// === C18/C19: downfloat/upfloat scores previous round (conditional) ===
	if params.PlayedRounds > 0 {
		// C19: upfloat opponent score R-1
		if inCurrentBracket {
			lowerFloat1 := floatAtRound(lowerPlayer, params.PlayedRounds-1)
			if !(lowerFloat1 == FloatUp &&
				higherPlayer.Score > lowerPlayer.Score+0.001) {
				sgShift := params.scoreGroupShift(higherPlayer.Score)
				setBit(shift + sgShift)
			}
		}
		shift += sgsShift

		// C18: downfloat score R-1
		if inCurrentBracket {
			lowerFloat1 := floatAtRound(lowerPlayer, params.PlayedRounds-1)
			higherFloat1 := floatAtRound(higherPlayer, params.PlayedRounds-1)
			if lowerFloat1 == FloatDown {
				sgShift := params.scoreGroupShift(lowerPlayer.Score)
				addend := new(big.Int).Lsh(one, uint(max(shift+sgShift, 0))) //nolint:gosec // shift values are bounded by tournament size
				result.Add(result, addend)
			}
			if higherFloat1 == FloatDown {
				sgShift := params.scoreGroupShift(higherPlayer.Score)
				addend := new(big.Int).Lsh(one, uint(max(shift+sgShift, 0))) //nolint:gosec // shift values are bounded by tournament size
				result.Add(result, addend)
			}
		}
		shift += sgsShift
	}

	// === C14/C15: downfloat/upfloat repeat previous round (conditional) ===
	if params.PlayedRounds > 0 {
		// C15: upfloat repeat R-1
		if inCurrentBracket {
			lowerFloat1 := floatAtRound(lowerPlayer, params.PlayedRounds-1)
			c15 := !(higherPlayer.Score > lowerPlayer.Score+0.001 &&
				lowerFloat1 == FloatUp)
			if c15 {
				setBit(shift)
			}
		}
		shift += sgBits

		// C14: downfloat repeat R-1 (value 0-2, occupies sgBits)
		if inCurrentBracket {
			lowerFloat1 := floatAtRound(lowerPlayer, params.PlayedRounds-1)
			higherFloat1 := floatAtRound(higherPlayer, params.PlayedRounds-1)
			if lowerFloat1 == FloatDown {
				setBit(shift)
			}
			if higherPlayer.Score <= lowerPlayer.Score+0.001 &&
				higherFloat1 == FloatDown {
				addend := new(big.Int).Lsh(one, uint(max(shift, 0))) //nolint:gosec // shift values are bounded by tournament size
				result.Add(result, addend)
			}
		}
		shift += sgBits
	}

	// === 4 color bits (each sgBits wide) ===
	// bbpPairings insertColorBits: lowerPlayer is first arg, higherPlayer
	// is second. Only set when inCurrentBracket (inCurrentScoreGroup).
	inCSG := inCurrentBracket

	prefLower := ComputeColorPreference(lowerPlayer.ColorHistory)
	prefHigher := ComputeColorPreference(higherPlayer.ColorHistory)

	lowerPrefColor := colorPrefDirection(prefLower)
	higherPrefColor := colorPrefDirection(prefHigher)

	lowerAbsImbalance := prefLower.ColorImbalance > 1 // absoluteColorImbalance(): imbalance only
	higherAbsImbalance := prefHigher.ColorImbalance > 1
	lowerAbsPref := prefLower.AbsolutePreference // absoluteColorPreference(): imbalance OR consecutive
	higherAbsPref := prefHigher.AbsolutePreference
	lowerStrongPref := prefLower.StrongPreference
	higherStrongPref := prefHigher.StrongPreference

	// Color bit 4 (lowest of the 4): strong color preference conflict
	cb4 := inCSG &&
		((!lowerStrongPref && !lowerAbsPref) ||
			(!higherStrongPref && !higherAbsPref) ||
			(lowerAbsPref && higherAbsPref) ||
			lowerPrefColor != higherPrefColor)
	if cb4 {
		setBit(shift)
	}
	shift += sgBits

	// Color bit 3: color preferences compatible
	cb3 := inCSG && colorPrefsCompatible(lowerPrefColor, higherPrefColor)
	if cb3 {
		setBit(shift)
	}
	shift += sgBits

	// Color bit 2: absolute color preference conflict (complex)
	lowerImbalance := colorImbalance(lowerPlayer)
	higherImbalance := colorImbalance(higherPlayer)
	lowerRepeated := repeatedColor(lowerPlayer)
	higherRepeated := repeatedColor(higherPlayer)

	cb2 := false
	if inCSG {
		if !lowerAbsPref || !higherAbsPref || lowerPrefColor != higherPrefColor {
			cb2 = true
		} else {
			if lowerImbalance == higherImbalance {
				cb2 = lowerRepeated == ColorNone || lowerRepeated != higherRepeated
			} else {
				var checkPlayer *PlayerState
				if lowerImbalance > higherImbalance {
					checkPlayer = higherPlayer
				} else {
					checkPlayer = lowerPlayer
				}
				checkRepeated := repeatedColor(checkPlayer)
				cb2 = checkRepeated != invertColorDir(lowerPrefColor)
			}
		}
	}
	if cb2 {
		setBit(shift)
	}
	shift += sgBits

	// Color bit 1 (highest of the 4): absolute color imbalance conflict
	cb1 := inCSG &&
		(!lowerAbsImbalance || !higherAbsImbalance || lowerPrefColor != higherPrefColor)
	if cb1 {
		setBit(shift)
	}
	shift += sgBits

	// === Bye assignee unplayed games (2 × sgBits) ===
	// Mirrors bbpPairings: for each player, if they are a bye candidate
	// (haven't received PAB AND in lowest score group), add their
	// unplayed-games count. Higher value = both players have MORE unplayed
	// games = matching them makes a bye-eligible player with FEWER unplayed
	// games more likely to be left unmatched (which is what C9 wants:
	// minimize unplayed games of the PAB assignee).
	isByeCandidateLowerForC9 := !lowerPlayer.ByeReceived &&
		lowerPlayer.Score <= params.ByeAssigneeScore+0.001
	if isByeCandidateLowerForC9 {
		unplayed := countUnplayedGames(lowerPlayer)
		if unplayed > 0 {
			addend := new(big.Int).SetInt64(int64(unplayed))
			addend.Lsh(addend, uint(max(shift, 0))) //nolint:gosec // shift values are bounded by tournament size
			result.Add(result, addend)
		}
	}
	shift += sgBits

	isByeCandidateHigherForC9 := !higherPlayer.ByeReceived &&
		higherPlayer.Score <= params.ByeAssigneeScore+0.001
	if isByeCandidateHigherForC9 {
		unplayed := countUnplayedGames(higherPlayer)
		if unplayed > 0 {
			addend := new(big.Int).SetInt64(int64(unplayed))
			addend.Lsh(addend, uint(max(shift, 0))) //nolint:gosec // shift values are bounded by tournament size
			result.Add(result, addend)
		}
	}
	shift += sgBits

	// === Maximize scores in next bracket (scoreGroupsShift) ===
	if inNextBracket {
		sgShift := params.scoreGroupShift(higherPlayer.Score)
		setBit(shift + sgShift)
	}
	shift += sgsShift

	// === Maximize pairs in next bracket (sgBits) ===
	if inNextBracket {
		setBit(shift)
	}
	shift += sgBits

	// === Maximize scores in current bracket (scoreGroupsShift) ===
	if inCurrentBracket {
		sgShift := params.scoreGroupShift(higherPlayer.Score)
		setBit(shift + sgShift)
	}
	shift += sgsShift

	// === Maximize pairs in current bracket (sgBits) ===
	if inCurrentBracket {
		setBit(shift)
	}
	shift += sgBits

	// === Bye eligibility (2 bits) ===
	// bbpPairings: 1 + !isByeCandidate(higher) + !isByeCandidate(lower)
	// isByeCandidate = player hasn't received PAB AND is in the lowest score group.
	// Higher value = neither player is a bye candidate = Blossom prefers to
	// match this pair, leaving bye candidates more likely to be unmatched.
	isByeCandidateLower := !lowerPlayer.ByeReceived &&
		lowerPlayer.Score <= params.ByeAssigneeScore+0.001
	isByeCandidateHigher := !higherPlayer.ByeReceived &&
		higherPlayer.Score <= params.ByeAssigneeScore+0.001
	byeVal := int64(1)
	if !isByeCandidateLower {
		byeVal++
	}
	if !isByeCandidateHigher {
		byeVal++
	}
	byeInt := new(big.Int).SetInt64(byeVal)
	byeInt.Lsh(byeInt, uint(max(shift, 0))) //nolint:gosec // shift values are bounded by tournament size
	result.Or(result, byeInt)

	return result
}

// scoreGroupShift returns the per-SG bit offset for a given score.
// Returns 0 if the score is not found (safe default).
func (p *EdgeWeightParams) scoreGroupShift(score float64) int {
	if shift, ok := p.ScoreGroupShifts[score]; ok {
		return shift
	}
	return 0
}

// colorPrefDirection returns the preferred color direction:
// ColorWhite, ColorBlack, or ColorNone (no preference).
func colorPrefDirection(pref ColorPreference) Color {
	if pref.Color != nil {
		return *pref.Color
	}
	return ColorNone
}

// countUnplayedGames returns the number of rounds where a player did not play
// (no color assigned = bye, absent, or withdrawn round).
func countUnplayedGames(p *PlayerState) int {
	unplayed := 0
	for _, c := range p.ColorHistory {
		if c == ColorNone {
			unplayed++
		}
	}
	return unplayed
}

// colorPrefsCompatible returns true if two color preferences are compatible
// (different or at least one is NONE). Mirrors bbpPairings'
// colorPreferencesAreCompatible.
func colorPrefsCompatible(a, b Color) bool {
	if a == ColorNone || b == ColorNone {
		return true
	}
	return a != b
}

// colorImbalance returns the signed color imbalance (whites - blacks)
// from played games. Matches bbpPairings' player.colorImbalance.
func colorImbalance(p *PlayerState) int {
	w, b := countColors(filterPlayed(p.ColorHistory))
	return w - b
}

// repeatedColor returns the last color that was played consecutively
// (2+ times in a row), or ColorNone. Matches bbpPairings' player.repeatedColor.
func repeatedColor(p *PlayerState) Color {
	played := filterPlayed(p.ColorHistory)
	if len(played) < 2 {
		return ColorNone
	}
	last := played[len(played)-1]
	secondLast := played[len(played)-2]
	if last == secondLast {
		return last
	}
	return ColorNone
}

// invertColorDir inverts a color direction (White↔Black, None→None).
func invertColorDir(c Color) Color {
	return c.Opposite()
}

// bitsToRepresent returns the number of bits needed to represent n.
// bitsToRepresent(0) = 0, bitsToRepresent(1) = 1, bitsToRepresent(2) = 2,
// bitsToRepresent(3) = 2, bitsToRepresent(4) = 3, etc.
func bitsToRepresent(n int) int {
	if n <= 0 {
		return 0
	}
	bits := 0
	v := n
	for v > 0 {
		bits++
		v >>= 1
	}
	return bits
}

// ---- Legacy functions kept for backward compatibility ----
// These are used by the final-pass matching in pairBracketsGlobal.

// EdgeWeightTieBreakBits is the number of low-order bits reserved in the
// edge weight for a tie-breaker value (natural S1-S2 ordering). The
// criteria weight occupies bits above this.
const EdgeWeightTieBreakBits = 10

// ComputeEdgeWeight converts a 12-element violation array (C10-C21) into
// a single int64 weight. Higher weight = fewer violations = better.
//
// Each criterion gets a 3-bit field (supporting violation counts 0-7).
// The field stores (maxVal - min(violations, maxVal)) so that fewer
// violations produce a higher value. Criteria are ordered by FIDE
// priority: C10 at bits 43-45 (highest), C21 at bits 10-12 (lowest),
// with bits 0-9 reserved for tie-breaking (see EdgeWeightTieBreakBits).
//
// violations[0] = C10 violation count, violations[11] = C21 violation count.
func ComputeEdgeWeight(violations [12]int) int64 {
	const bitsPerCriterion = 3
	const maxVal = 7 // (1 << bitsPerCriterion) - 1

	var w int64
	for i := 0; i < 12; i++ {
		v := violations[i]
		if v > maxVal {
			v = maxVal
		}
		inverted := int64(maxVal - v)
		// C10 (i=0) gets the highest bit position, C21 (i=11) the lowest.
		shift := EdgeWeightTieBreakBits + (11-i)*bitsPerCriterion
		w |= inverted << shift
	}
	return w
}

// PairEdgeWeight computes the Blossom edge weight for a single proposed
// pairing by evaluating all per-pair criteria (C10-C21).
// Used only by the final-pass matching at the end of pairBracketsGlobal.
func PairEdgeWeight(pair *ProposedPairing, ctx *CriteriaContext, isDownfloater map[string]bool, bracketScore float64) int64 {
	return PairEdgeWeightGlobal(pair, ctx, isDownfloater, bracketScore, true, false)
}

// PairEdgeWeightGlobal is the legacy edge weight function.
// Kept for the final-pass matching only. The main loop now uses
// ComputeBaseEdgeWeight instead.
func PairEdgeWeightGlobal(pair *ProposedPairing, ctx *CriteriaContext, isDownfloater map[string]bool, bracketScore float64, inCurrentBracket, inNextBracket bool) int64 {
	var violations [12]int
	if inCurrentBracket {
		violations[0] = PairCriterionC10(pair, ctx)
		violations[1] = PairCriterionC11(pair, ctx)
		violations[2] = PairCriterionC12(pair, ctx)
		violations[3] = PairCriterionC13(pair, ctx)
		violations[4] = PairCriterionC14(pair, ctx, isDownfloater)
		violations[5] = PairCriterionC15(pair, ctx, isDownfloater)
		violations[6] = PairCriterionC16(pair, ctx, isDownfloater)
		violations[7] = PairCriterionC17(pair, ctx, isDownfloater)
		violations[8] = PairCriterionC18(pair, ctx, isDownfloater, bracketScore)
		violations[9] = PairCriterionC19(pair, ctx, isDownfloater, bracketScore)
		violations[10] = PairCriterionC20(pair, ctx, isDownfloater, bracketScore)
		violations[11] = PairCriterionC21(pair, ctx, isDownfloater, bracketScore)
	}
	return ComputeEdgeWeight(violations)
}

// PairCriterionC10 counts topscorers and opponents who would end up with
// |color difference| > 2 after this round. Returns 0, 1, or 2.
func PairCriterionC10(pair *ProposedPairing, ctx *CriteriaContext) int {
	wTop := ctx.TopScorers[pair.White.ID]
	bTop := ctx.TopScorers[pair.Black.ID]
	if !wTop && !bTop {
		return 0
	}

	colorW, colorB := simulateColor(pair.White, pair.Black, ctx.IsLastRound)
	violations := 0
	if wTop && colorDiffAfter(pair.White, colorW) > 2 {
		violations++
	}
	if bTop && colorDiffAfter(pair.Black, colorB) > 2 {
		violations++
	}
	if wTop && colorDiffAfter(pair.Black, colorB) > 2 {
		violations++
	}
	if bTop && colorDiffAfter(pair.White, colorW) > 2 {
		violations++
	}
	return violations
}

// PairCriterionC11 counts topscorers and opponents who would end up with
// 3+ consecutive same-color games. Returns 0, 1, or 2.
func PairCriterionC11(pair *ProposedPairing, ctx *CriteriaContext) int {
	wTop := ctx.TopScorers[pair.White.ID]
	bTop := ctx.TopScorers[pair.Black.ID]
	if !wTop && !bTop {
		return 0
	}

	colorW, colorB := simulateColor(pair.White, pair.Black, ctx.IsLastRound)
	violations := 0
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
	return violations
}

// PairCriterionC12 counts players who would not receive their color
// preference. Returns 0, 1, or 2.
func PairCriterionC12(pair *ProposedPairing, ctx *CriteriaContext) int {
	colorW, colorB := simulateColor(pair.White, pair.Black, ctx.IsLastRound)
	prefW := ComputeColorPreference(pair.White.ColorHistory)
	prefB := ComputeColorPreference(pair.Black.ColorHistory)

	violations := 0
	wantW := prefW.PreferredColor()
	wantB := prefB.PreferredColor()
	if wantW != nil && *wantW != colorW {
		violations++
	}
	if wantB != nil && *wantB != colorB {
		violations++
	}
	return violations
}

// PairCriterionC13 counts players who would not receive their strong (or
// absolute) color preference. Returns 0, 1, or 2.
func PairCriterionC13(pair *ProposedPairing, ctx *CriteriaContext) int {
	colorW, colorB := simulateColor(pair.White, pair.Black, ctx.IsLastRound)
	prefW := ComputeColorPreference(pair.White.ColorHistory)
	prefB := ComputeColorPreference(pair.Black.ColorHistory)

	violations := 0
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
	return violations
}

// PairCriterionC14 counts downfloaters in this pair who also downfloated
// in the previous round. Returns 0, 1, or 2.
func PairCriterionC14(pair *ProposedPairing, ctx *CriteriaContext, isDownfloater map[string]bool) int {
	prevRoundIdx := ctx.CurrentRound - 2
	violations := 0
	for _, p := range []*PlayerState{pair.White, pair.Black} {
		if isDownfloater[p.ID] && floatAtRound(p, prevRoundIdx) == FloatDown {
			violations++
		}
	}
	return violations
}

// PairCriterionC15 counts MDP opponents who upfloated in the previous
// round. Returns 0 or 1.
func PairCriterionC15(pair *ProposedPairing, ctx *CriteriaContext, isDownfloater map[string]bool) int {
	prevRoundIdx := ctx.CurrentRound - 2
	violations := 0
	wDown := isDownfloater[pair.White.ID]
	bDown := isDownfloater[pair.Black.ID]
	if wDown && !bDown && floatAtRound(pair.Black, prevRoundIdx) == FloatUp {
		violations++
	}
	if bDown && !wDown && floatAtRound(pair.White, prevRoundIdx) == FloatUp {
		violations++
	}
	return violations
}

// PairCriterionC16 counts downfloaters who also downfloated 2 rounds ago.
// Returns 0, 1, or 2.
func PairCriterionC16(pair *ProposedPairing, ctx *CriteriaContext, isDownfloater map[string]bool) int {
	twoAgoIdx := ctx.CurrentRound - 3
	violations := 0
	for _, p := range []*PlayerState{pair.White, pair.Black} {
		if isDownfloater[p.ID] && floatAtRound(p, twoAgoIdx) == FloatDown {
			violations++
		}
	}
	return violations
}

// PairCriterionC17 counts MDP opponents who upfloated 2 rounds ago.
// Returns 0 or 1.
func PairCriterionC17(pair *ProposedPairing, ctx *CriteriaContext, isDownfloater map[string]bool) int {
	twoAgoIdx := ctx.CurrentRound - 3
	violations := 0
	wDown := isDownfloater[pair.White.ID]
	bDown := isDownfloater[pair.Black.ID]
	if wDown && !bDown && floatAtRound(pair.Black, twoAgoIdx) == FloatUp {
		violations++
	}
	if bDown && !wDown && floatAtRound(pair.White, twoAgoIdx) == FloatUp {
		violations++
	}
	return violations
}

// PairCriterionC18 returns 1 if any downfloater in this pair also
// downfloated in the previous round and has score diff > 0.
func PairCriterionC18(pair *ProposedPairing, _ *CriteriaContext, isDownfloater map[string]bool, bracketScore float64) int {
	for _, p := range []*PlayerState{pair.White, pair.Black} {
		if isDownfloater[p.ID] && scoreDiffInt(p.Score, bracketScore) > 0 {
			return 1
		}
	}
	return 0
}

// PairCriterionC19 returns 1 if any MDP opponent in this pair upfloated
// in the previous round and has score diff > 0.
func PairCriterionC19(pair *ProposedPairing, ctx *CriteriaContext, isDownfloater map[string]bool, bracketScore float64) int {
	prevRoundIdx := ctx.CurrentRound - 2
	wDown := isDownfloater[pair.White.ID]
	bDown := isDownfloater[pair.Black.ID]
	if wDown && !bDown && floatAtRound(pair.Black, prevRoundIdx) == FloatUp {
		if scoreDiffInt(pair.Black.Score, bracketScore) > 0 {
			return 1
		}
	}
	if bDown && !wDown && floatAtRound(pair.White, prevRoundIdx) == FloatUp {
		if scoreDiffInt(pair.White.Score, bracketScore) > 0 {
			return 1
		}
	}
	return 0
}

// PairCriterionC20 returns 1 if any downfloater also downfloated 2 rounds ago
// and has score diff > 0.
func PairCriterionC20(pair *ProposedPairing, ctx *CriteriaContext, isDownfloater map[string]bool, bracketScore float64) int {
	twoAgoIdx := ctx.CurrentRound - 3
	for _, p := range []*PlayerState{pair.White, pair.Black} {
		if isDownfloater[p.ID] && floatAtRound(p, twoAgoIdx) == FloatDown {
			if scoreDiffInt(p.Score, bracketScore) > 0 {
				return 1
			}
		}
	}
	return 0
}

// PairCriterionC21 returns 1 if any MDP opponent upfloated 2 rounds ago
// and has score diff > 0.
func PairCriterionC21(pair *ProposedPairing, ctx *CriteriaContext, isDownfloater map[string]bool, bracketScore float64) int {
	twoAgoIdx := ctx.CurrentRound - 3
	wDown := isDownfloater[pair.White.ID]
	bDown := isDownfloater[pair.Black.ID]
	if wDown && !bDown && floatAtRound(pair.Black, twoAgoIdx) == FloatUp {
		if scoreDiffInt(pair.Black.Score, bracketScore) > 0 {
			return 1
		}
	}
	if bDown && !wDown && floatAtRound(pair.White, twoAgoIdx) == FloatUp {
		if scoreDiffInt(pair.White.Score, bracketScore) > 0 {
			return 1
		}
	}
	return 0
}
