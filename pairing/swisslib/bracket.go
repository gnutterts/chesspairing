package swisslib

import "sort"

// ScoreGroup holds all players with the same pairing score.
// Players are ordered by TPN ascending within the group.
type ScoreGroup struct {
	Score   float64
	Players []*PlayerState // ordered by TPN ascending
}

// Bracket is the processing unit for the pairing algorithm.
// A homogeneous bracket contains players with the same native score.
// A heterogeneous bracket contains downfloaters merged with a native group.
type Bracket struct {
	Players       []*PlayerState
	Homogeneous   bool
	OriginalScore float64        // native score of this bracket
	Downfloaters  []*PlayerState // players floated down from higher brackets (heterogeneous)
}

// BuildScoreGroups creates score groups from player states.
// Players are grouped by score and each group is sorted by TPN ascending.
// Returns groups in descending score order. Input need not be pre-sorted.
func BuildScoreGroups(players []PlayerState) []ScoreGroup {
	if len(players) == 0 {
		return nil
	}

	// Group by score. Using float64 as map key is safe because pairing scores
	// are always exact multiples of 0.5 (win=1, draw=0.5, loss=0, bye=1),
	// which are exactly representable in IEEE 754.
	groups := make(map[float64][]*PlayerState)
	var scores []float64
	seen := make(map[float64]bool)

	for i := range players {
		p := &players[i]
		if !seen[p.Score] {
			seen[p.Score] = true
			scores = append(scores, p.Score)
		}
		groups[p.Score] = append(groups[p.Score], p)
	}

	// Sort scores descending.
	sort.Float64s(scores)
	// Reverse to descending.
	for i, j := 0, len(scores)-1; i < j; i, j = i+1, j-1 {
		scores[i], scores[j] = scores[j], scores[i]
	}

	// Build result with players sorted by TPN within each group.
	result := make([]ScoreGroup, 0, len(scores))
	for _, score := range scores {
		playerList := groups[score]
		sort.Slice(playerList, func(i, j int) bool {
			return playerList[i].TPN < playerList[j].TPN
		})
		result = append(result, ScoreGroup{
			Score:   score,
			Players: playerList,
		})
	}

	return result
}

// BuildBrackets creates initial homogeneous brackets from score groups.
// Each score group becomes one homogeneous bracket.
// Returns brackets in descending score order.
func BuildBrackets(groups []ScoreGroup) []Bracket {
	// Sort groups by score descending (they should already be, but ensure).
	sorted := make([]ScoreGroup, len(groups))
	copy(sorted, groups)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	brackets := make([]Bracket, 0, len(sorted))
	for _, g := range sorted {
		players := make([]*PlayerState, len(g.Players))
		copy(players, g.Players)
		brackets = append(brackets, Bracket{
			Players:       players,
			Homogeneous:   true,
			OriginalScore: g.Score,
		})
	}
	return brackets
}

// MergeIntoHeterogeneous creates a heterogeneous bracket by merging
// downfloaters into a native bracket.
func MergeIntoHeterogeneous(native Bracket, floaters []*PlayerState) Bracket {
	allPlayers := make([]*PlayerState, 0, len(native.Players)+len(floaters))
	allPlayers = append(allPlayers, floaters...)
	allPlayers = append(allPlayers, native.Players...)

	downfloaters := make([]*PlayerState, len(floaters))
	copy(downfloaters, floaters)

	return Bracket{
		Players:       allPlayers,
		Homogeneous:   false,
		OriginalScore: native.OriginalScore,
		Downfloaters:  downfloaters,
	}
}

// CollapseBrackets merges two consecutive brackets into one.
// Used when a bracket fails to pair and must be combined with the next.
// Players are deduplicated by ID to prevent self-pairings from repeated
// collapse operations.
func CollapseBrackets(upper, lower Bracket) Bracket {
	seen := make(map[string]bool, len(upper.Players)+len(lower.Players))
	allPlayers := make([]*PlayerState, 0, len(upper.Players)+len(lower.Players))
	for _, p := range upper.Players {
		if !seen[p.ID] {
			seen[p.ID] = true
			allPlayers = append(allPlayers, p)
		}
	}
	for _, p := range lower.Players {
		if !seen[p.ID] {
			seen[p.ID] = true
			allPlayers = append(allPlayers, p)
		}
	}

	// Upper bracket players are downfloaters in the collapsed bracket.
	downfloaters := make([]*PlayerState, 0, len(upper.Players))
	downfloaters = append(downfloaters, upper.Players...)

	return Bracket{
		Players:       allPlayers,
		Homogeneous:   false,
		OriginalScore: lower.OriginalScore,
		Downfloaters:  downfloaters,
	}
}
