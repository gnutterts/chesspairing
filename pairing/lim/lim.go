package lim

import (
	"context"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// Pair implements chesspairing.Pairer for the Lim Swiss system.
func (p *Pairer) Pair(_ context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	result := &chesspairing.PairingResult{}

	// Build player states.
	players := swisslib.BuildPlayerStates(state)
	if len(players) <= 1 {
		// 0 or 1 player: just assign bye if needed.
		if len(players) == 1 {
			result.Byes = append(result.Byes, chesspairing.ByeEntry{
				PlayerID: players[0].ID,
				Type:     chesspairing.ByePAB,
			})
		}
		return result, nil
	}

	// Build forbidden pairs map.
	forbidden := buildForbiddenMap(p.opts.ForbiddenPairs)

	// Assign PAB if odd number of players.
	playerPtrs := make([]*swisslib.PlayerState, len(players))
	for i := range players {
		playerPtrs[i] = &players[i]
	}

	if swisslib.NeedsBye(len(playerPtrs)) {
		byeSelector := LimByeSelector{}
		byePlayer := byeSelector.SelectBye(playerPtrs)
		if byePlayer != nil {
			result.Byes = append(result.Byes, chesspairing.ByeEntry{
				PlayerID: byePlayer.ID,
				Type:     chesspairing.ByePAB,
			})
			// Remove bye player from pairing pool.
			playerPtrs = removePtrs(playerPtrs, byePlayer)
		}
	}

	// Build score groups.
	scoreGroups := swisslib.BuildScoreGroups(derefPlayers(playerPtrs))

	// Determine median score.
	roundsPlayed := len(state.Rounds)
	medianScore := float64(roundsPlayed) / 2.0

	// Determine processing order (Art. 2.2).
	aboveMedian, belowMedian, medianGroup := splitByMedian(scoreGroups, medianScore)

	// Process scoregroups in Lim order, collecting floaters (unpaired players)
	// from each group. When a scoregroup has an odd number of players, one
	// player is selected as a floater using the Lim floater selection rules
	// (Art. 3.2-3.9) which consider compatibility with adjacent groups.
	// After all groups are processed, floaters are paired together across
	// scoregroup boundaries.

	// Build the ordered sequence of groups for processing.
	var ordered []swisslib.ScoreGroup

	// Phase 1: Highest → just above median (downward).
	ordered = append(ordered, aboveMedian...)

	// Phase 2: Lowest → just below median (upward, reversed).
	for i := len(belowMedian) - 1; i >= 0; i-- {
		ordered = append(ordered, belowMedian[i])
	}

	// Phase 3: Median group last (paired downward, Art. 2.2).
	if medianGroup != nil {
		ordered = append(ordered, *medianGroup)
	}

	var allPairs [][2]*swisslib.PlayerState
	var pendingFloaters []FloaterEntry
	floatedIDs := make(map[string]bool)
	isMaxi := *p.opts.MaxiTournament

	for idx, sg := range ordered {
		groupPlayers := make([]*swisslib.PlayerState, len(sg.Players))
		copy(groupPlayers, sg.Players)

		// Determine pairing direction for this scoregroup.
		pairingDown := sg.Score > medianScore || floatEqual(sg.Score, medianScore)

		// Phase 1: Merge incoming floaters into this group (Art. 3.6-3.8).
		// Floaters are sorted per Art. 3.6.3/3.7.3 priority order before
		// merging, so their position in the combined group reflects the
		// correct pairing priority when ExchangeMatch splits into S1/S2.
		if len(pendingFloaters) > 0 {
			isUpperHalf := pairingDown
			sorted := sortFloaters(pendingFloaters, isUpperHalf)
			for _, fe := range sorted {
				groupPlayers = append(groupPlayers, fe.Player)
			}
			pendingFloaters = nil
		}

		// Phase 2: If odd, select a new floater from this group.
		if len(groupPlayers)%2 == 1 {
			var adjacentPlayers []*swisslib.PlayerState
			if idx+1 < len(ordered) {
				adjacentPlayers = ordered[idx+1].Players
			}

			var floater *swisslib.PlayerState
			if pairingDown {
				floater = SelectDownFloater(groupPlayers, adjacentPlayers, forbidden, floatedIDs, isMaxi)
			} else {
				floater = SelectUpFloater(groupPlayers, adjacentPlayers, forbidden, floatedIDs, isMaxi)
			}
			if floater != nil {
				floatedIDs[floater.ID] = true
				dir := FloatDown
				if !pairingDown {
					dir = FloatUp
				}
				pendingFloaters = append(pendingFloaters, FloaterEntry{
					Player:      floater,
					Direction:   dir,
					SourceScore: sg.Score,
				})
				groupPlayers = removePtrs(groupPlayers, floater)
			}
		}

		// Phase 3: Exchange match remaining players within this group (Art. 4).
		pairs, unpaired := ExchangeMatch(groupPlayers, pairingDown, forbidden)

		// Phase 4: Colour exchange pass (Art. 5.2 / 5.7).
		// Swap opponents between pairs to reduce colour conflicts, subject
		// to compatibility and the maxi rating constraint.
		pairs = ColourExchange(pairs, forbidden, isMaxi)

		allPairs = append(allPairs, pairs...)

		// Any unpaired from exchange matching become floaters too.
		for _, u := range unpaired {
			floatedIDs[u.ID] = true
			dir := FloatDown
			if !pairingDown {
				dir = FloatUp
			}
			pendingFloaters = append(pendingFloaters, FloaterEntry{
				Player:      u,
				Direction:   dir,
				SourceScore: sg.Score,
			})
		}
	}

	// Pair remaining floaters across scoregroups.
	if len(pendingFloaters) >= 2 {
		floaterPlayers := make([]*swisslib.PlayerState, len(pendingFloaters))
		for i, fe := range pendingFloaters {
			floaterPlayers[i] = fe.Player
		}
		sort.SliceStable(floaterPlayers, func(i, j int) bool {
			if floaterPlayers[i].Score != floaterPlayers[j].Score {
				return floaterPlayers[i].Score > floaterPlayers[j].Score
			}
			return floaterPlayers[i].TPN < floaterPlayers[j].TPN
		})
		floaterPairs, remaining := greedyPair(floaterPlayers, true, forbidden, nil)
		allPairs = append(allPairs, floaterPairs...)

		// Repair: if greedy pairing left unpaired players (e.g., they already
		// played each other), try to swap them into existing pairs.
		if len(remaining) >= 2 {
			var stillUnpaired []*swisslib.PlayerState
			allPairs, stillUnpaired = repairUnpaired(allPairs, remaining, forbidden)
			remaining = stillUnpaired
		}

		// Safety: any remaining unpaired players receive a PAB.
		for _, u := range remaining {
			result.Byes = append(result.Byes, chesspairing.ByeEntry{
				PlayerID: u.ID,
				Type:     chesspairing.ByePAB,
			})
		}
		pendingFloaters = nil
	}

	// Safety: a single remaining floater also receives a PAB.
	if len(pendingFloaters) == 1 {
		result.Byes = append(result.Byes, chesspairing.ByeEntry{
			PlayerID: pendingFloaters[0].Player.ID,
			Type:     chesspairing.ByePAB,
		})
	}

	// Assign colours and build final result.
	for boardNum, pair := range allPairs {
		isAboveMedian := pair[0].PairingScore > medianScore ||
			pair[0].PairingScore == medianScore
		wID, bID := AllocateColor(pair[0], pair[1], state.CurrentRound, isAboveMedian, topSeedColorPtr(p.opts.TopSeedColor))
		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   boardNum + 1,
			WhiteID: wID,
			BlackID: bID,
		})
	}

	// Re-number boards: sort by max score desc, then min TPN asc.
	sortBoards(result.Pairings, playerPtrs)

	return result, nil
}

// splitByMedian divides score groups into above-median, below-median, and
// the median group itself.
func splitByMedian(groups []swisslib.ScoreGroup, medianScore float64) (above, below []swisslib.ScoreGroup, median *swisslib.ScoreGroup) {
	for i, sg := range groups {
		if floatEqual(sg.Score, medianScore) {
			// This is the median group.
			mg := groups[i]
			median = &mg
		} else if sg.Score > medianScore {
			above = append(above, sg)
		} else {
			below = append(below, sg)
		}
	}
	return
}

// floatEqual compares two float64 values for near-equality with a tolerance
// of 0.001 (sufficient for half-point chess scores).
func floatEqual(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.001
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

// repairUnpaired resolves unpaired players that couldn't be matched among
// themselves (e.g., they already played each other) by swapping them into
// existing pairs. Two strategies are tried:
//
// Strategy 1 (same-pair swap): find a single existing pair {a,b} where
// u1 can pair with one member and u2 with the other. Dissolve {a,b} and
// form {u1,a}+{u2,b} or {u1,b}+{u2,a}.
//
// Strategy 2 (chain swap): dissolve two existing pairs. Pair u1 with a
// member of pair1, and u2 with a member of pair2, then pair the two freed
// members with each other.
func repairUnpaired(pairs [][2]*swisslib.PlayerState, unpaired []*swisslib.PlayerState, forbidden map[[2]string]bool) ([][2]*swisslib.PlayerState, []*swisslib.PlayerState) {
	if len(unpaired) < 2 {
		return pairs, unpaired
	}

	for len(unpaired) >= 2 {
		u1 := unpaired[0]
		u2 := unpaired[1]

		// Strategy 1: same-pair swap.
		if idx, swap := findSamePairSwap(pairs, u1, u2, forbidden); idx >= 0 {
			a, b := pairs[idx][0], pairs[idx][1]
			if swap == 0 {
				pairs[idx] = [2]*swisslib.PlayerState{u1, a}
				pairs = append(pairs, [2]*swisslib.PlayerState{u2, b})
			} else {
				pairs[idx] = [2]*swisslib.PlayerState{u1, b}
				pairs = append(pairs, [2]*swisslib.PlayerState{u2, a})
			}
			unpaired = unpaired[2:]
			continue
		}

		// Strategy 2: chain swap across two different pairs.
		if newPair, ok := findChainSwap(pairs, u1, u2, forbidden); ok {
			pairs = append(pairs, newPair)
			unpaired = unpaired[2:]
			continue
		}

		// Could not repair this pair of unpaired players.
		break
	}

	return pairs, unpaired
}

// findSamePairSwap finds an existing pair where u1 and u2 can each take one
// member. Returns the pair index and swap variant (0 = u1-a/u2-b, 1 = u1-b/u2-a),
// or (-1, 0) if not found.
func findSamePairSwap(pairs [][2]*swisslib.PlayerState, u1, u2 *swisslib.PlayerState, forbidden map[[2]string]bool) (int, int) {
	for pi, ep := range pairs {
		a, b := ep[0], ep[1]
		if IsCompatible(u1, a, forbidden) && IsCompatible(u2, b, forbidden) {
			return pi, 0
		}
		if IsCompatible(u1, b, forbidden) && IsCompatible(u2, a, forbidden) {
			return pi, 1
		}
	}
	return -1, 0
}

// findChainSwap dissolves two existing pairs to accommodate u1 and u2.
// It tries: pair u1 with a member of pair[i], pair u2 with a member of pair[j],
// and pair the two freed members together. If successful, it modifies pairs[i]
// and pairs[j] in-place and returns the new third pair to append.
func findChainSwap(pairs [][2]*swisslib.PlayerState, u1, u2 *swisslib.PlayerState, forbidden map[[2]string]bool) ([2]*swisslib.PlayerState, bool) {
	n := len(pairs)
	for i := range n {
		a, b := pairs[i][0], pairs[i][1]

		// For each member of pair[i] that u1 can pair with:
		type u1Match struct {
			partner *swisslib.PlayerState // member u1 pairs with
			freed   *swisslib.PlayerState // member freed from pair[i]
		}
		var u1Matches []u1Match
		if IsCompatible(u1, a, forbidden) {
			u1Matches = append(u1Matches, u1Match{partner: a, freed: b})
		}
		if IsCompatible(u1, b, forbidden) {
			u1Matches = append(u1Matches, u1Match{partner: b, freed: a})
		}
		if len(u1Matches) == 0 {
			continue
		}

		for j := range n {
			if j == i {
				continue
			}
			c, d := pairs[j][0], pairs[j][1]

			for _, m := range u1Matches {
				// Try: u2-c, freed-d
				if IsCompatible(u2, c, forbidden) && IsCompatible(m.freed, d, forbidden) {
					pairs[i] = [2]*swisslib.PlayerState{u1, m.partner}
					pairs[j] = [2]*swisslib.PlayerState{u2, c}
					return [2]*swisslib.PlayerState{m.freed, d}, true
				}
				// Try: u2-d, freed-c
				if IsCompatible(u2, d, forbidden) && IsCompatible(m.freed, c, forbidden) {
					pairs[i] = [2]*swisslib.PlayerState{u1, m.partner}
					pairs[j] = [2]*swisslib.PlayerState{u2, d}
					return [2]*swisslib.PlayerState{m.freed, c}, true
				}
			}
		}
	}
	return [2]*swisslib.PlayerState{}, false
}

// removePtrs removes a specific player from the pointer slice.
func removePtrs(players []*swisslib.PlayerState, remove *swisslib.PlayerState) []*swisslib.PlayerState {
	result := make([]*swisslib.PlayerState, 0, len(players)-1)
	for _, p := range players {
		if p.ID != remove.ID {
			result = append(result, p)
		}
	}
	return result
}

// derefPlayers converts pointer slice to value slice for BuildScoreGroups.
func derefPlayers(ptrs []*swisslib.PlayerState) []swisslib.PlayerState {
	result := make([]swisslib.PlayerState, len(ptrs))
	for i, p := range ptrs {
		result[i] = *p
	}
	return result
}

// topSeedColorPtr converts the string option to a Color pointer.
func topSeedColorPtr(opt *string) *swisslib.Color {
	if opt == nil {
		return nil
	}
	switch *opt {
	case "white":
		c := swisslib.ColorWhite
		return &c
	case "black":
		c := swisslib.ColorBlack
		return &c
	default:
		return nil
	}
}

// sortBoards sorts pairings for board ordering:
// max score of pair (desc), then min TPN of pair (asc).
func sortBoards(pairings []chesspairing.GamePairing, players []*swisslib.PlayerState) {
	playerMap := make(map[string]*swisslib.PlayerState, len(players))
	for _, p := range players {
		playerMap[p.ID] = p
	}

	sort.SliceStable(pairings, func(i, j int) bool {
		pi1, pi2 := playerMap[pairings[i].WhiteID], playerMap[pairings[i].BlackID]
		pj1, pj2 := playerMap[pairings[j].WhiteID], playerMap[pairings[j].BlackID]

		// Max score of pair.
		maxI := maxScore(pi1, pi2)
		maxJ := maxScore(pj1, pj2)
		if maxI != maxJ {
			return maxI > maxJ
		}

		// Min TPN of pair.
		minI := minTPN(pi1, pi2)
		minJ := minTPN(pj1, pj2)
		return minI < minJ
	})

	// Renumber boards.
	for i := range pairings {
		pairings[i].Board = i + 1
	}
}

// maxScore returns the higher score between two players.
func maxScore(a, b *swisslib.PlayerState) float64 {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return b.Score
	}
	if b == nil {
		return a.Score
	}
	if a.Score > b.Score {
		return a.Score
	}
	return b.Score
}

// minTPN returns the lower TPN between two players.
func minTPN(a, b *swisslib.PlayerState) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return b.TPN
	}
	if b == nil {
		return a.TPN
	}
	if a.TPN < b.TPN {
		return a.TPN
	}
	return b.TPN
}
