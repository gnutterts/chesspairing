package lim

import (
	"context"
	"sort"

	chesspairing "github.com/gnutterts/chesspairing"
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
	var allFloaters []*swisslib.PlayerState

	for idx, sg := range ordered {
		groupPlayers := make([]*swisslib.PlayerState, len(sg.Players))
		copy(groupPlayers, sg.Players)

		// Merge any existing floaters into this group for exchange matching.
		groupPlayers = append(groupPlayers, allFloaters...)
		allFloaters = nil

		// If odd, select a floater before exchange matching.
		if len(groupPlayers)%2 == 1 {
			// Determine adjacent group's players for floater selection.
			var adjacentPlayers []*swisslib.PlayerState
			if idx+1 < len(ordered) {
				adjacentPlayers = ordered[idx+1].Players
			}

			// Determine pairing direction.
			pairingDown := sg.Score > medianScore || (sg.Score-medianScore > -0.001 && sg.Score-medianScore < 0.001)

			var floater *swisslib.PlayerState
			if pairingDown {
				floater = SelectDownFloater(groupPlayers, adjacentPlayers, forbidden)
			} else {
				floater = SelectUpFloater(groupPlayers, adjacentPlayers, forbidden)
			}
			if floater != nil {
				allFloaters = append(allFloaters, floater)
				groupPlayers = removePtrs(groupPlayers, floater)
			}
		}

		pairingDown := sg.Score > medianScore || (sg.Score-medianScore > -0.001 && sg.Score-medianScore < 0.001)
		pairs, unpaired := ExchangeMatch(groupPlayers, pairingDown, forbidden)
		allPairs = append(allPairs, pairs...)
		allFloaters = append(allFloaters, unpaired...)
	}

	// Pair remaining floaters across scoregroups.
	if len(allFloaters) >= 2 {
		sort.SliceStable(allFloaters, func(i, j int) bool {
			if allFloaters[i].Score != allFloaters[j].Score {
				return allFloaters[i].Score > allFloaters[j].Score
			}
			return allFloaters[i].TPN < allFloaters[j].TPN
		})
		floaterPairs, _ := greedyPair(allFloaters, true, forbidden, nil)
		allPairs = append(allPairs, floaterPairs...)
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
		diff := sg.Score - medianScore
		if diff < 0 {
			diff = -diff
		}
		if diff < 0.001 {
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
