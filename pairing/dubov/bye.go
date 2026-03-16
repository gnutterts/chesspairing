package dubov

import (
	"sort"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// DubovByeSelector selects the bye player per Dubov system rules (Art. 2.3):
// lowest score -> most games played -> highest TPN (lowest-ranked).
type DubovByeSelector struct{}

// SelectBye returns the player to receive the PAB, or nil if all have
// already received one.
func (s DubovByeSelector) SelectBye(players []*swisslib.PlayerState) *swisslib.PlayerState {
	eligible := filterNoByeReceived(players)
	if len(eligible) == 0 {
		return nil
	}

	sort.SliceStable(eligible, func(i, j int) bool {
		// 1. Lowest score first.
		if eligible[i].Score != eligible[j].Score {
			return eligible[i].Score < eligible[j].Score
		}
		// 2. Most games played first.
		gi := swisslib.GamesPlayed(eligible[i])
		gj := swisslib.GamesPlayed(eligible[j])
		if gi != gj {
			return gi > gj
		}
		// 3. Highest TPN (lowest ranking) first.
		return eligible[i].TPN > eligible[j].TPN
	})

	return eligible[0]
}

// filterNoByeReceived returns players who have not yet received a PAB.
func filterNoByeReceived(players []*swisslib.PlayerState) []*swisslib.PlayerState {
	var eligible []*swisslib.PlayerState
	for _, p := range players {
		if !p.ByeReceived {
			eligible = append(eligible, p)
		}
	}
	return eligible
}

// Compile-time check.
var _ swisslib.ByeSelector = DubovByeSelector{}
