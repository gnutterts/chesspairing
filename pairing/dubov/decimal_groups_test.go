package dubov

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestDecimalScoreGroup(t *testing.T) {
	players := []*swisslib.PlayerState{{ID: "a", Score: 0.1 + 0.2}, {ID: "b", Score: 0.3}}
	groups := groupPlayers(players)
	if len(groups.scores) != 1 || len(groups.byScore[groups.scores[0]]) != 2 {
		t.Errorf("groups = %+v", groups)
	}
}
