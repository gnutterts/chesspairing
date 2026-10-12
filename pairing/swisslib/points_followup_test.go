package swisslib

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestExcusedWinValueBlocksPAB(t *testing.T) {
	state := &chesspairing.TournamentState{ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringStandard, Options: map[string]any{"pointExcused": 1.0}}, Players: []chesspairing.PlayerEntry{{ID: "p"}}, Rounds: []chesspairing.RoundData{{Byes: []chesspairing.ByeEntry{{PlayerID: "p", Type: chesspairing.ByeExcused}}}}, CurrentRound: 2}
	if !ComputePABIneligibility(state)["p"].Any() {
		t.Fatal("excused win must bar PAB")
	}
	players := mustBuildPlayerStates(t, state)
	if !players[0].PABIneligible.Any() {
		t.Fatal("pairer must see PAB ineligibility")
	}
	state.PreAssignedByes = []chesspairing.ByeEntry{{PlayerID: "p", Type: chesspairing.ByePAB}}
	if err := ValidatePreAssignedPAB(state); err == nil {
		t.Fatal("pre-assigned PAB accepted")
	}
}

func TestUnplayedDrawKeepsDrawPoints(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  chesspairing.ScoringConfig
		want float64
	}{
		{"default", chesspairing.ScoringConfig{}, 0.5},
		{"football", chesspairing.ScoringConfig{System: chesspairing.ScoringFootball}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := &chesspairing.TournamentState{ScoringConfig: tc.cfg, Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}}, Rounds: []chesspairing.RoundData{{Games: []chesspairing.GameData{{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultDraw, IsForfeit: true}}}}, CurrentRound: 2}
			for _, p := range mustBuildPlayerStates(t, state) {
				if p.Score != tc.want || len(p.Opponents) != 0 {
					t.Errorf("%s: score %g, opponents %v", p.ID, p.Score, p.Opponents)
				}
			}
		})
	}
}

func TestScoreGroupKeyLargeAndDecimalShift(t *testing.T) {
	players := []PlayerState{{ID: "a", Score: 0.1 + 0.2, PairingScore: 0.1 + 0.2}, {ID: "b", Score: 0.3, PairingScore: 0.3}, {ID: "huge", Score: 1e300, PairingScore: 1e300}, {ID: "larger", Score: 2e300, PairingScore: 2e300}, {ID: "lower", Score: 0, PairingScore: 0}}
	groups := BuildScoreGroups(players)
	if len(groups) != 4 || len(groups[2].Players) != 2 {
		t.Fatalf("groups: %+v", groups)
	}
	params := ComputeEdgeWeightParams(groups, 2)
	if got, want := params.scoreGroupShift(players[0].Score), params.ScoreGroupShifts[groups[2].Score]; got != want || got == 0 {
		t.Errorf("shift = %d, want %d", got, want)
	}
}

func TestEmptyScoringSystemUsesOptions(t *testing.T) {
	cfg := chesspairing.ScoringConfig{Options: map[string]any{"pointWin": 2.0, "pointDraw": 1.0}}
	if got := PairingWinPoints(cfg); got != 2 {
		t.Fatalf("PairingWinPoints = %v, want 2", got)
	}
	if err := BakuPremise(chesspairing.ScoringConfig{Options: map[string]any{"pointWin": 3.0}}); err == nil {
		t.Fatal("Baku accepted non-proportional scoring with an empty system")
	}
}

func TestDecimalScoresAreNormalised(t *testing.T) {
	// Three draws of 0.1 and one win of 0.3 are the same score.
	cfg := chesspairing.ScoringConfig{System: chesspairing.ScoringStandard, Options: map[string]any{"pointWin": 0.3, "pointDraw": 0.1}}
	state := &chesspairing.TournamentState{
		ScoringConfig: cfg,
		Players:       []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultDraw}, {WhiteID: "c", BlackID: "d", Result: chesspairing.ResultWhiteWins}}},
			{Number: 2, Games: []chesspairing.GameData{{WhiteID: "a", BlackID: "d", Result: chesspairing.ResultDraw}, {WhiteID: "b", BlackID: "c", Result: chesspairing.ResultDraw}}},
			{Number: 3, Games: []chesspairing.GameData{{WhiteID: "a", BlackID: "c", Result: chesspairing.ResultDraw}, {WhiteID: "b", BlackID: "d", Result: chesspairing.ResultDraw}}},
		},
		CurrentRound: 4,
	}
	players := mustBuildPlayerStates(t, state)
	score := map[string]float64{}
	for _, p := range players {
		score[p.ID] = p.Score
	}
	// a: three draws (0.1+0.1+0.1); c: win 0.3 then two draws = 0.5; compare a with 0.3.
	if score["a"] != 0.3 {
		t.Fatalf("a score = %v, want exactly 0.3", score["a"])
	}
}
