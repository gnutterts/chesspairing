package swisslib

import (
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func TestUnrecordedRoundKeepsFloatHistoryAligned(t *testing.T) {
	for _, tc := range []struct {
		name   string
		absent float64
		want   Float
	}{
		{name: "zero points", want: FloatNone},
		{name: "half point", absent: 0.5, want: FloatDown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := &cp.TournamentState{
				Players:       []cp.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}},
				CurrentRound:  3,
				ScoringConfig: cp.ScoringConfig{Options: map[string]any{"pointAbsent": tc.absent}},
				Rounds: []cp.RoundData{
					{Games: []cp.GameData{{WhiteID: "a", BlackID: "c", Result: cp.ResultDraw}}},
					{Games: []cp.GameData{{WhiteID: "a", BlackID: "c", Result: cp.ResultDraw}}},
				},
			}
			players, err := BuildPlayerStates(state)
			if err != nil {
				t.Fatal(err)
			}
			for _, player := range players {
				if player.ID != "b" {
					continue
				}
				if len(player.FloatHistory) != 2 || player.FloatHistory[1] != tc.want {
					t.Fatalf("float history for b = %v, want two rounds ending in %v", player.FloatHistory, tc.want)
				}
				return
			}
			t.Fatal("missing player b")
		})
	}
}
