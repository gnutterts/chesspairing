// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"math/big"

	"github.com/gnutterts/chesspairing/algorithm/blossom"
)

// PairBracketsGlobal performs global Blossom matching across all score groups.
// This mirrors bbpPairings' computeMatching architecture: a single global
// matching graph is built with all players, and brackets are processed
// top-down using a 7-phase loop that incrementally updates edge weights.
//
// For odd player counts, a completability pre-matching (Stage 0.5) runs first
// to determine which player will receive the bye. The unmatched player's score
// becomes ByeAssigneeScore in EdgeWeightParams, which influences the real
// edge weights via isByeCandidate logic.
//
// Used by Dutch (C.04.3) and Burstein (C.04.4.2) Swiss pairing systems.
// The behavior is controlled through the CriteriaContext (TopScorers,
// LookAhead, ForbiddenPairs) and the edge weight parameters which encode
// system-specific optimization criteria.
//
// Returns the committed pairings, the unmatched player (bye recipient for
// odd player counts, nil for even), and diagnostic notes.
func PairBracketsGlobal(
	scoreGroups []ScoreGroup,
	ctx *CriteriaContext,
	playerMap map[string]*PlayerState,
) ([]ProposedPairing, *PlayerState, []string) {
	if len(scoreGroups) == 0 {
		return nil, nil, nil
	}

	var notes []string
	var allCommitted []ProposedPairing

	// Precompute edge weight parameters (mirrors bbpPairings' computeMatching
	// setup at lines 685-715).
	ewParams := ComputeEdgeWeightParams(scoreGroups, ctx.CurrentRound-1)
	sgSizeBits := ewParams.ScoreGroupSizeBits

	// =====================================================================
	// Stage 0: Build the GLOBAL player list from ALL score groups.
	// bbpPairings creates matchingComputer with ALL players upfront
	// (lines 753-930). We do the same.
	// =====================================================================
	var allPlayers []*PlayerState
	// sgBoundaries[i] = start index of score group i in allPlayers.
	sgBoundaries := make([]int, len(scoreGroups)+1)
	for si, sg := range scoreGroups {
		sgBoundaries[si] = len(allPlayers)
		allPlayers = append(allPlayers, sg.Players...)
	}
	sgBoundaries[len(scoreGroups)] = len(allPlayers)
	totalN := len(allPlayers)

	if totalN < 2 {
		// Single player: return them as unmatched (PAB candidate).
		if totalN == 1 {
			return nil, allPlayers[0], nil
		}
		return nil, nil, nil
	}

	// =====================================================================
	// Stage 0.5: Completability pre-matching (odd player count only).
	//
	// Mirrors bbpPairings lines 766-930: run a simplified Blossom matching
	// to determine which player will receive the bye. The unmatched player's
	// score becomes byeAssigneeScore, which is used by isByeCandidate in
	// the real edge weights.
	//
	// Simplified edge weight (per bbpPairings):
	//   bit 0..1: 1 + !eligibleForBye(i) + !eligibleForBye(j)
	//     → Prefer matching bye-ineligible players (leave eligible unmatched)
	//   bit 2..N: scoreGroupShifts[score_i] + scoreGroupShifts[score_j]
	//     → C5: maximize sum of matched scores (leave lowest score unmatched)
	//   bit N+1: (score_i >= topScore ? 1 : 0) + (score_j >= topScore ? 1 : 0)
	//     → Protect top-score players from getting the bye
	//
	// For even player count: byeAssigneeScore stays -1 (no bye candidate).
	// =====================================================================
	if NeedsBye(totalN) {
		topScore := scoreGroups[0].Score
		sgsShift := ewParams.ScoreGroupsShift

		var preEdges []blossom.BigEdge
		for i := 0; i < totalN; i++ {
			for j := i + 1; j < totalN; j++ {
				pi, pj := allPlayers[i], allPlayers[j]

				// Only check C1 (already played) — not C3 (color) since the
				// completability pre-matching ignores color constraints.
				if HasPlayed(pi, pj) {
					continue
				}
				if IsPairForbiddenByID(pi.ID, pj.ID, ctx) {
					continue
				}

				// bbpPairings completability edge weight (dutch.cpp lines 779-800):
				// Built bottom-up with shifts, final layout (HIGH→LOW):
				//   bye eligibility (2 bits) | score sum (sgsShift bits) | top score (sgSizeBits bits)
				w := new(big.Int)

				// 1. Bye eligibility: 1 + !eligibleForBye(i) + !eligibleForBye(j)
				byeVal := int64(1)
				if pi.ByeReceived {
					byeVal++
				}
				if pj.ByeReceived {
					byeVal++
				}
				w.SetInt64(byeVal)

				// 2. Shift left by scoreGroupsShift, OR in score sum
				w.Lsh(w, uint(sgsShift))
				scoreSumVal := int64(0)
				if shift, ok := ewParams.ScoreGroupShifts[pi.Score]; ok {
					scoreSumVal += int64(shift)
				}
				if shift, ok := ewParams.ScoreGroupShifts[pj.Score]; ok {
					scoreSumVal += int64(shift)
				}
				if scoreSumVal > 0 {
					w.Or(w, new(big.Int).SetInt64(scoreSumVal))
				}

				// 3. Shift left by scoreGroupSizeBits, OR in top score bit
				w.Lsh(w, uint(ewParams.ScoreGroupSizeBits))
				topVal := int64(0)
				if pi.Score >= topScore-0.001 {
					topVal = 1
				}
				if topVal > 0 {
					w.Or(w, new(big.Int).SetInt64(topVal))
				}

				preEdges = append(preEdges, blossom.BigEdge{
					I: i, J: j, Weight: w,
				})
			}
		}

		if len(preEdges) > 0 {
			preMatch := blossom.MaxWeightMatchingBig(preEdges, true)
			// Find the unmatched player — their score is byeAssigneeScore.
			for idx, partner := range preMatch {
				if partner == -1 && idx < totalN {
					ewParams.ByeAssigneeScore = allPlayers[idx].Score
					break
				}
			}
		}
	}

	// Global base weights: baseWeight[i][j] for all i < j in allPlayers.
	// Indexed as map[(i,j)] where i < j, to avoid O(n^2) memory for slices.
	// Uses *big.Int because bbpPairings edge weights exceed 64 bits.
	bigOne := big.NewInt(1)
	globalBase := make(map[[2]int]*big.Int, totalN*totalN/4)

	// =====================================================================
	// Stage 1: Pre-populate ALL edges with ComputeBaseEdgeWeight(false, false).
	// This mirrors bbpPairings lines 766-827 where it sets edge weights
	// for ALL pairs before the bracket loop starts.
	// =====================================================================
	for i := 0; i < totalN; i++ {
		if i%100 == 0 && ctx.DeadlineExceeded() {
			break
		}
		for j := i + 1; j < totalN; j++ {
			pi, pj := allPlayers[i], allPlayers[j]
			if HasPlayed(pi, pj) {
				continue
			}
			if IsPairForbiddenByID(pi.ID, pj.ID, ctx) {
				continue
			}
			// Use lowest score as bracketScore for C3 check during init.
			bs := pj.Score
			if pi.Score < bs {
				bs = pi.Score
			}
			if !C3AbsoluteColorConflict(&ProposedPairing{
				White: pi, Black: pj, BracketScore: bs,
			}, ctx) {
				continue
			}
			w := ComputeBaseEdgeWeight(pi, pj, false, false, &ewParams)
			if w.Sign() == 0 {
				w = new(big.Int).Set(bigOne)
			}
			globalBase[edgeKey(i, j)] = w
		}
	}

	// Global mutable edge weight map. Initialized from globalBase.
	globalEdgeW := make(map[[2]int]*big.Int, len(globalBase))
	for k, w := range globalBase {
		globalEdgeW[k] = new(big.Int).Set(w)
	}

	// Global finalized flags.
	globalFinalized := make([]bool, totalN)

	// Track committed player IDs.
	committed := make(map[string]bool)

	// Helper: build edges for Blossom from the global edge map.
	// Unlike the previous approach that filtered to "active vertices", we now
	// include ALL non-finalized vertices — matching bbpPairings' architecture
	// where the matching computer always sees all players. Future-bracket
	// players have lower weights from Stage 1, so Blossom prefers current-bracket
	// matches naturally.
	buildGlobalEdges := func() []blossom.BigEdge {
		edges := make([]blossom.BigEdge, 0, len(globalEdgeW))
		for k, w := range globalEdgeW {
			if w != nil && w.Sign() > 0 && !globalFinalized[k[0]] && !globalFinalized[k[1]] {
				edges = append(edges, blossom.BigEdge{
					I: k[0], J: k[1], Weight: new(big.Int).Set(w),
				})
			}
		}
		return edges
	}

	runGlobalBlossom := func() []int {
		edges := buildGlobalEdges()
		if len(edges) == 0 {
			return nil
		}
		return blossom.MaxWeightMatchingBig(edges, true)
	}

	finalizePairGlobal := func(a, b int) {
		globalFinalized[a] = true
		globalFinalized[b] = true
		// Mirror bbpPairings finalizePair: zero all edges from a and b,
		// EXCEPT keep the mutual edge at weight 1 so Blossom continues
		// to match them in subsequent computeMatching() calls.
		pairKey := edgeKey(a, b)
		for k := range globalEdgeW {
			if k[0] == a || k[1] == a || k[0] == b || k[1] == b {
				if k == pairKey {
					globalEdgeW[k] = new(big.Int).Set(bigOne)
				} else {
					globalEdgeW[k] = new(big.Int)
				}
			}
		}
	}

	// =====================================================================
	// Stage 2: Bracket loop.
	// playersByIndex is a LOCAL view: downfloaters + current SG + next SG.
	// vertexIdx maps local index → global index in allPlayers.
	// =====================================================================
	var playersByIndex []*PlayerState
	var vertexIdx []int
	scoreGroupBegin := 0
	sgIter := 0

	// Bootstrap: seed with the first score group.
	for gi := sgBoundaries[0]; gi < sgBoundaries[1]; gi++ {
		playersByIndex = append(playersByIndex, allPlayers[gi])
		vertexIdx = append(vertexIdx, gi)
	}
	sgIter = 1
	maxIter := 2*len(scoreGroups) + 2 // safety limit

	for iter := 0; (len(playersByIndex) > 1 || sgIter < len(scoreGroups)) && iter < maxIter; iter++ {
		if ctx.DeadlineExceeded() {
			break
		}

		nextScoreGroupBegin := len(playersByIndex)

		// Append the next score group's players.
		if sgIter < len(scoreGroups) {
			for gi := sgBoundaries[sgIter]; gi < sgBoundaries[sgIter+1]; gi++ {
				if !committed[allPlayers[gi].ID] {
					playersByIndex = append(playersByIndex, allPlayers[gi])
					vertexIdx = append(vertexIdx, gi)
				}
			}
			sgIter++
		}

		n := len(playersByIndex)
		if n < 2 {
			break
		}

		// Bracket score for the current bracket.
		var bracketScore float64
		if scoreGroupBegin < nextScoreGroupBegin && scoreGroupBegin < n {
			bracketScore = playersByIndex[scoreGroupBegin].Score
		} else if n > 0 {
			bracketScore = playersByIndex[0].Score
		}

		dfCount := scoreGroupBegin

		// =====================================================================
		// Update global edges for current bracket + next SG.
		// Mirrors bbpPairings computeBaseEdgeWeights: only recompute edges
		// where largerPlayerIndex >= scoreGroupBegin (at least one player
		// is from current bracket or next SG). This UPDATES the global
		// edge map — other edges remain at their Stage 1 values.
		// =====================================================================
		// Also rebuild a local baseWeight view for Phases 2-7.
		baseWeight := make([][]*big.Int, n)
		for i := 0; i < n; i++ {
			baseWeight[i] = make([]*big.Int, n)
		}

		for li := 0; li < n; li++ {
			if li%100 == 0 && ctx.DeadlineExceeded() {
				break
			}
			for lj := li + 1; lj < n; lj++ {
				// Only update where larger local index >= scoreGroupBegin.
				if lj < scoreGroupBegin {
					continue
				}

				gi, gj := vertexIdx[li], vertexIdx[lj]
				pi, pj := playersByIndex[li], playersByIndex[lj]

				if HasPlayed(pi, pj) {
					continue
				}
				if IsPairForbiddenByID(pi.ID, pj.ID, ctx) {
					continue
				}
				if !C3AbsoluteColorConflict(&ProposedPairing{
					White: pi, Black: pj, BracketScore: bracketScore,
				}, ctx) {
					continue
				}

				inCurrentBracket := lj < nextScoreGroupBegin
				inNextBracket := lj >= nextScoreGroupBegin

				w := ComputeBaseEdgeWeight(pi, pj, inCurrentBracket, inNextBracket, &ewParams)
				if w.Sign() == 0 {
					w = new(big.Int).Set(bigOne)
				}

				baseWeight[li][lj] = w
				baseWeight[lj][li] = new(big.Int).Set(w)

				// Update global edge map.
				key := edgeKey(gi, gj)
				globalBase[key] = new(big.Int).Set(w)
				globalEdgeW[key] = new(big.Int).Set(w)
			}
		}

		// Also populate baseWeight for downfloater-downfloater pairs
		// (local indices below scoreGroupBegin) from globalBase.
		for li := 0; li < scoreGroupBegin; li++ {
			for lj := li + 1; lj < scoreGroupBegin; lj++ {
				gi, gj := vertexIdx[li], vertexIdx[lj]
				key := edgeKey(gi, gj)
				if w, ok := globalBase[key]; ok && w != nil && w.Sign() > 0 {
					baseWeight[li][lj] = new(big.Int).Set(w)
					baseWeight[lj][li] = new(big.Int).Set(w)
				}
			}
		}

		// edgeWeightComputer (mirrors bbpPairings lines 1055-1085).
		// Operates on LOCAL indices but writes to GLOBAL edgeW via vertexIdx.
		edgeWeightComputer := func(smallerLI, largerLI, smallerRemIdx, remPairs int) *big.Int {
			bw := baseWeight[largerLI][smallerLI]
			if bw == nil || bw.Sign() == 0 {
				return new(big.Int)
			}

			// Build addend matching bbpPairings' edgeWeightComputer.
			addend := new(big.Int)

			// Minimize exchanges: 1 if this is an S1 player (natural pairing).
			if smallerRemIdx < remPairs {
				addend.SetInt64(1)
			}

			// Minimize BSN distance: shift left by 2*sgSizeBits, subtract index.
			addend.Lsh(addend, uint(max(sgSizeBits, 0))) //nolint:gosec // sgSizeBits is bounded by tournament size
			addend.Lsh(addend, uint(max(sgSizeBits, 0))) //nolint:gosec // sgSizeBits is bounded by tournament size
			addend.Sub(addend, big.NewInt(int64(smallerRemIdx)))

			// Reserve 1 bit for exchange optimization.
			addend.Lsh(addend, 1)

			return new(big.Int).Add(bw, addend)
		}

		// Local helpers that operate on local indices but use global edgeW.
		setEdge := func(li, lj int, w *big.Int) {
			gi, gj := vertexIdx[li], vertexIdx[lj]
			globalEdgeW[edgeKey(gi, gj)] = w
		}
		getEdge := func(li, lj int) *big.Int {
			gi, gj := vertexIdx[li], vertexIdx[lj]
			w := globalEdgeW[edgeKey(gi, gj)]
			if w == nil {
				return new(big.Int)
			}
			return w
		}
		_ = getEdge // may not be used in all paths

		localFinalized := make([]bool, n)
		matchedPhase2 := make([]bool, n)

		finalizePairLocal := func(a, b int) {
			localFinalized[a] = true
			localFinalized[b] = true
			finalizePairGlobal(vertexIdx[a], vertexIdx[b])
		}

		// =====================================================================
		// Phase 1: Initial matching (global Blossom).
		// =====================================================================
		mate := runGlobalBlossom()

		// Translate global matching to local mate array.
		// mate[globalIdx] = matched globalIdx. We need localMate[localIdx].
		globalToLocal := make(map[int]int, n)
		for li, gi := range vertexIdx {
			globalToLocal[gi] = li
		}

		toLocalMate := func(globalMate []int) []int {
			localMate := make([]int, n)
			for i := range localMate {
				localMate[i] = -1
			}
			for li, gi := range vertexIdx {
				if gi < len(globalMate) && globalMate[gi] >= 0 {
					if mli, ok := globalToLocal[globalMate[gi]]; ok {
						localMate[li] = mli
					}
					// If mate is outside our local view, leave as -1.
				}
			}
			return localMate
		}

		localMate := toLocalMate(mate)

		// =====================================================================
		// Phase 2: Choose moved-down players (heterogeneous only).
		// =====================================================================
		var iterPairs []ProposedPairing
		isHeterogeneous := dfCount > 0

		if isHeterogeneous && len(mate) > 0 {
			type dfScoreGroup struct {
				startIdx int
				score    float64
			}
			var dfGroups []dfScoreGroup
			for i := 0; i < dfCount; i++ {
				if len(dfGroups) == 0 || playersByIndex[i].Score != dfGroups[len(dfGroups)-1].score {
					dfGroups = append(dfGroups, dfScoreGroup{startIdx: i, score: playersByIndex[i].Score})
				}
			}

			for _, dfg := range dfGroups {
				dfEnd := dfCount
				for k := dfg.startIdx + 1; k < dfCount; k++ {
					if playersByIndex[k].Score != dfg.score {
						dfEnd = k
						break
					}
				}

				remainingDF := 0
				remainingMatchedDF := 0
				for k := dfg.startIdx; k < dfEnd; k++ {
					remainingDF++
					if localMate[k] >= dfCount && localMate[k] < nextScoreGroupBegin {
						remainingMatchedDF++
					}
				}

				for i := dfg.startIdx; i < dfEnd; i++ {
					if remainingMatchedDF == 0 {
						continue
					}
					if remainingDF <= remainingMatchedDF {
						// Auto-accept: all remaining DFs can be matched.
						// bbpPairings does NOT decrement either counter here.
						matchedPhase2[i] = true
						continue
					}
					remainingDF--

					if localMate[i] < dfCount || localMate[i] >= nextScoreGroupBegin {
						for j := dfCount; j < nextScoreGroupBegin; j++ {
							bw := baseWeight[j][i]
							if bw != nil && bw.Sign() > 0 {
								setEdge(i, j, new(big.Int).Or(bw, bigOne))
							}
						}
						mate = runGlobalBlossom()
						localMate = toLocalMate(mate)
					}

					if localMate[i] >= dfCount && localMate[i] < nextScoreGroupBegin {
						matchedPhase2[i] = true
						remainingMatchedDF--
						sgSize := nextScoreGroupBegin - dfCount
						stickyVal := big.NewInt(int64(sgSize + 1))
						for j := dfCount; j < nextScoreGroupBegin; j++ {
							bw := baseWeight[j][i]
							if bw != nil && bw.Sign() > 0 {
								setEdge(i, j, new(big.Int).Or(bw, stickyVal))
							}
						}
					}
				}
			}

			// Sub-phase 2b: Choose opponents for each matched downfloater.
			for i := 0; i < dfCount; i++ {
				if !matchedPhase2[i] {
					continue
				}
				addend := big.NewInt(int64(n))
				for j := nextScoreGroupBegin - 1; j >= dfCount; j-- {
					if matchedPhase2[j] {
						continue
					}
					bw := baseWeight[j][i]
					if bw != nil && bw.Sign() > 0 {
						setEdge(i, j, new(big.Int).Add(bw, addend))
						addend.Add(addend, bigOne)
					}
				}
				mate = runGlobalBlossom()
				localMate = toLocalMate(mate)
				if localMate[i] >= dfCount && localMate[i] < nextScoreGroupBegin {
					partner := localMate[i]
					matchedPhase2[partner] = true
					iterPairs = append(iterPairs, ProposedPairing{
						White:        playersByIndex[i],
						Black:        playersByIndex[partner],
						BracketScore: bracketScore,
					})
					finalizePairLocal(i, partner)
				}
			}
		}

		// =====================================================================
		// Phase 3: Initialize remainder.
		// Mirrors bbpPairings lines 1270-1285: skip SG players whose Blossom
		// mate is a downfloater (index < scoreGroupBegin/dfCount), even if
		// that downfloater wasn't selected by Phase 2.
		// =====================================================================
		remainder := make([]int, 0, n)
		for i := dfCount; i < nextScoreGroupBegin; i++ {
			if localFinalized[i] {
				continue
			}
			// bbpPairings: skip if stableMatching < scoreGroupBeginVertex
			if localMate[i] >= 0 && localMate[i] < dfCount {
				continue
			}
			remainder = append(remainder, i)
		}

		// Count remainderPairs from latest matching.
		// bbpPairings lines 1281-1284: count players whose mate has lower vertex.
		remainderPairs := 0
		remIndexOf := make(map[int]int, len(remainder))
		for ri, li := range remainder {
			remIndexOf[li] = ri
		}
		for _, li := range remainder {
			if localMate[li] >= 0 && localMate[li] < li {
				if _, inRem := remIndexOf[localMate[li]]; inRem {
					remainderPairs++
				}
			}
		}
		if remainderPairs == 0 {
			var remS1 int
			if dfCount > 0 {
				remS1 = (nextScoreGroupBegin - dfCount) / 2
			} else {
				remS1 = len(remainder) / 2
			}
			remainderPairs = remS1
			if remainderPairs > len(remainder)/2 {
				remainderPairs = len(remainder) / 2
			}
		}

		// Apply edgeWeightComputer to remainder edges.
		remSet := make(map[int]bool, len(remainder))
		for _, li := range remainder {
			remSet[li] = true
		}
		// Zero out remainder×remainder edges in global map.
		for _, li := range remainder {
			for _, lj := range remainder {
				if li < lj {
					setEdge(li, lj, new(big.Int))
				}
			}
		}
		for ri, li := range remainder {
			for rj := ri + 1; rj < len(remainder); rj++ {
				lj := remainder[rj]
				w := edgeWeightComputer(li, lj, ri, remainderPairs)
				if w != nil && w.Sign() > 0 {
					setEdge(li, lj, w)
				}
			}
		}
		mate = runGlobalBlossom()
		localMate = toLocalMate(mate)

		// =====================================================================
		// Phase 4: Exchange selection Phase 1.
		// =====================================================================
		exchangeCount := 0
		for ri := 0; ri < remainderPairs; ri++ {
			li := remainder[ri]
			matchedToS2 := false
			if localMate[li] >= 0 && localMate[li] < n {
				mateRI, inRem := remIndexOf[localMate[li]]
				if inRem && mateRI > ri {
					matchedToS2 = true
				}
			}
			if !matchedToS2 {
				exchangeCount++
			}
		}

		if exchangeCount > 0 && remainderPairs > 0 {
			exchangesRemaining := exchangeCount
			for ri := remainderPairs - 1; ri >= 0 && exchangesRemaining > 0; ri-- {
				if ctx.DeadlineExceeded() {
					break
				}
				li := remainder[ri]
				isMatchedS1S2 := false
				if localMate[li] >= 0 && localMate[li] < n {
					mateRI, inRem := remIndexOf[localMate[li]]
					if inRem && mateRI > ri {
						isMatchedS1S2 = true
					}
				}
				if isMatchedS1S2 {
					for rj := ri + 1; rj < len(remainder); rj++ {
						lj := remainder[rj]
						w := edgeWeightComputer(li, lj, ri, remainderPairs)
						if w != nil && w.Sign() > 0 {
							w = new(big.Int).Sub(w, bigOne)
							setEdge(li, lj, w)
						}
					}
					mate = runGlobalBlossom()
					localMate = toLocalMate(mate)
				}
				exchange := true
				if localMate[li] >= 0 && localMate[li] < n {
					mateRI, inRem := remIndexOf[localMate[li]]
					if inRem && mateRI > ri {
						exchange = false
					}
				}
				if exchange {
					exchangesRemaining--
				}
				for rj := ri + 1; rj < len(remainder); rj++ {
					lj := remainder[rj]
					if exchange {
						baseWeight[lj][li] = nil
						baseWeight[li][lj] = nil
					}
					w := edgeWeightComputer(li, lj, ri, remainderPairs)
					setEdge(li, lj, w)
				}
			}
		}

		// =====================================================================
		// Phase 5: Exchange selection Phase 2.
		// =====================================================================
		if exchangeCount > 0 {
			exchangesRemaining := exchangeCount
			for ri := remainderPairs; ri < len(remainder) && exchangesRemaining > 1; ri++ {
				if ctx.DeadlineExceeded() {
					break
				}
				li := remainder[ri]
				if li >= nextScoreGroupBegin {
					continue
				}

				alreadyExchanged := false
				if localMate[li] >= 0 && localMate[li] < n {
					mateRI, inRem := remIndexOf[localMate[li]]
					if inRem && mateRI > ri {
						alreadyExchanged = true
					}
				}
				if !alreadyExchanged {
					for rj := ri + 1; rj < len(remainder); rj++ {
						lj := remainder[rj]
						w := edgeWeightComputer(li, lj, ri, remainderPairs)
						if w != nil && w.Sign() > 0 {
							w = new(big.Int).Add(w, bigOne)
							setEdge(li, lj, w)
						}
					}
					mate = runGlobalBlossom()
					localMate = toLocalMate(mate)
				}
				exchanged := false
				if localMate[li] >= 0 && localMate[li] < n {
					mateRI, inRem := remIndexOf[localMate[li]]
					if inRem && mateRI > ri {
						exchanged = true
					}
				}
				if exchanged {
					exchangesRemaining--
					for rj := 0; rj < ri; rj++ {
						lj := remainder[rj]
						baseWeight[li][lj] = nil
						baseWeight[lj][li] = nil
						setEdge(lj, li, new(big.Int))
					}
				}
				if !alreadyExchanged {
					for rj := ri + 1; rj < len(remainder); rj++ {
						lj := remainder[rj]
						w := edgeWeightComputer(li, lj, ri, remainderPairs)
						setEdge(li, lj, w)
					}
				}
			}
		}

		// =====================================================================
		// Phase 6: Finalize exchanges.
		// =====================================================================
		for ri := 0; ri < len(remainder); ri++ {
			li := remainder[ri]
			for rj := ri + 1; rj < len(remainder); rj++ {
				lj := remainder[rj]

				iNotMatchedDown := localMate[li] <= li || localMate[li] >= nextScoreGroupBegin
				if !iNotMatchedDown {
					mateRI, inRem := remIndexOf[localMate[li]]
					if !inRem || mateRI <= ri {
						iNotMatchedDown = true
					}
				}

				jHasNaturalPair := false
				if localMate[lj] > lj && localMate[lj] < n {
					mateRI, inRem := remIndexOf[localMate[lj]]
					if inRem && mateRI > rj && localMate[lj] < nextScoreGroupBegin {
						jHasNaturalPair = true
					}
				}

				if iNotMatchedDown || jHasNaturalPair {
					baseWeight[lj][li] = nil
					baseWeight[li][lj] = nil
				}
				bwVal := baseWeight[lj][li]
				if bwVal == nil {
					bwVal = new(big.Int)
				}
				setEdge(li, lj, bwVal)
			}
		}

		// =====================================================================
		// Phase 7: Choose opponents (within remainder only).
		// =====================================================================
		// Re-run Blossom after Phase 6 modifications.
		mate = runGlobalBlossom()
		localMate = toLocalMate(mate)

		for ri := 0; ri < len(remainder); ri++ {
			li := remainder[ri]
			if ctx.DeadlineExceeded() {
				break
			}

			if localFinalized[li] {
				continue
			}

			if localMate[li] < 0 {
				continue
			}
			mateRI, inRem := remIndexOf[localMate[li]]
			if !inRem || mateRI <= ri {
				continue
			}
			// Only match within current SG (remainder pairs).
			if localMate[li] >= nextScoreGroupBegin {
				continue
			}

			addend := new(big.Int)
			for rj := len(remainder) - 1; rj > ri; rj-- {
				lj := remainder[rj]
				if localFinalized[lj] {
					continue
				}
				bw := baseWeight[lj][li]
				if bw == nil || bw.Sign() == 0 {
					continue
				}
				setEdge(li, lj, new(big.Int).Add(bw, addend))
				addend.Add(addend, bigOne)
			}
			mate = runGlobalBlossom()
			localMate = toLocalMate(mate)

			p7mate := localMate[li]
			if p7mate >= 0 && p7mate < nextScoreGroupBegin && !localFinalized[p7mate] {
				iterPairs = append(iterPairs, ProposedPairing{
					White:        playersByIndex[li],
					Black:        playersByIndex[p7mate],
					BracketScore: bracketScore,
				})
				finalizePairLocal(li, p7mate)
			}
		}

		// =====================================================================
		// Carry forward: commit finalized players, keep unmatched.
		// Mirrors bbpPairings lines 1601-1648. Players matched within the
		// current bracket are saved; all others become downfloaters in the
		// next iteration (including those matched to next-SG players by the
		// Blossom but not committed by Phase 7).
		// =====================================================================
		var newPlayersByIndex []*PlayerState
		var newVertexIdx []int
		newScoreGroupBegin := 0

		for i := 0; i < n; i++ {
			p := playersByIndex[i]
			if localFinalized[i] {
				committed[p.ID] = true
			} else {
				if i < nextScoreGroupBegin {
					newScoreGroupBegin++
				}
				newPlayersByIndex = append(newPlayersByIndex, p)
				newVertexIdx = append(newVertexIdx, vertexIdx[i])
			}
		}

		// Record float directions.
		for _, pair := range iterPairs {
			for _, p := range []*PlayerState{pair.White, pair.Black} {
				if mp, ok := playerMap[p.ID]; ok {
					switch {
					case p.Score > bracketScore+0.001:
						mp.FloatHistory = append(mp.FloatHistory, FloatDown)
					case p.Score < bracketScore-0.001:
						mp.FloatHistory = append(mp.FloatHistory, FloatUp)
					default:
						mp.FloatHistory = append(mp.FloatHistory, FloatNone)
					}
				}
			}
		}
		allCommitted = append(allCommitted, iterPairs...)

		playersByIndex = newPlayersByIndex
		vertexIdx = newVertexIdx
		scoreGroupBegin = newScoreGroupBegin

		if len(playersByIndex) <= 1 && sgIter >= len(scoreGroups) {
			break
		}
		// Safety: prevent infinite loops. bbpPairings loops until
		// playersByIndex <= 1. In the worst case, each SG is visited once
		// as current bracket, and then once more as downfloaters are
		// re-processed. Max iterations = 2 * numScoreGroups + some margin.
	}

	// Identify unmatched player (PAB candidate) — the one player not committed.
	// With odd player count, exactly one player will be left unmatched by the
	// Blossom matching. The bye eligibility bits in edge weights ensure that
	// bye-ineligible players (already received PAB) get higher edge weights,
	// so the Blossom prefers to match them and leave a bye-eligible player
	// unmatched.
	var unmatchedPlayer *PlayerState
	for _, p := range allPlayers {
		if !committed[p.ID] {
			unmatchedPlayer = p
			break
		}
	}

	return allCommitted, unmatchedPlayer, notes
}

// edgeKey returns the canonical (i < j) key for an edge.
func edgeKey(i, j int) [2]int {
	if i < j {
		return [2]int{i, j}
	}
	return [2]int{j, i}
}
