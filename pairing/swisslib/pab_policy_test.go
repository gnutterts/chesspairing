package swisslib

import (
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func TestForfeitC2UsesConfiguredWinThreshold(t *testing.T) {
	state := &cp.TournamentState{
		CurrentRound: 2,
		Players:      []cp.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}},
		Rounds: []cp.RoundData{{Games: []cp.GameData{
			{WhiteID: "a", BlackID: "b", Result: cp.ResultForfeitWhiteWins},
			{WhiteID: "c", BlackID: "d", Result: cp.ResultDoubleForfeit},
		}}},
	}
	for _, tc := range []struct {
		name     string
		options  map[string]any
		want     bool
		wantLoss bool
	}{
		{name: "default", want: true},
		{name: "half-point forfeit win", options: map[string]any{"pointForfeitWin": 0.5}, want: false},
		{name: "full-point forfeit loss", options: map[string]any{"pointForfeitWin": 0.5, "pointForfeitLoss": 1}, wantLoss: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state.ScoringConfig.Options = tc.options
			computed := ComputePABIneligibility(state)
			players, err := BuildPlayerStates(state)
			if err != nil {
				t.Fatal(err)
			}
			for _, player := range players {
				want := player.ID == "a" && tc.want || player.ID != "a" && tc.wantLoss
				if player.PABIneligible.FullPointUnplayed != want || computed[player.ID].FullPointUnplayed != want {
					t.Fatalf("C2 for %s: build=%v compute=%v, want %v", player.ID, player.PABIneligible, computed[player.ID], want)
				}
			}
		})
	}
}
