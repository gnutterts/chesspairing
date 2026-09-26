package keizer

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestPointsForResultMatchesScoreForHistoryFreeByeTypes(t *testing.T) {
	cases := []struct {
		typ  chesspairing.ByeType
		want float64
	}{
		{chesspairing.ByePAB, 0.5},
		{chesspairing.ByeHalf, 0.5},
		{chesspairing.ByeZero, 0},
		{chesspairing.ByeClubCommitment, 0.5},
	}
	for _, tc := range cases {
		t.Run(tc.typ.String(), func(t *testing.T) {
			no := false
			opts := Options{SelfVictory: &no}
			s := New(opts)
			state := &chesspairing.TournamentState{Players: []chesspairing.PlayerEntry{{ID: "A"}}, Rounds: []chesspairing.RoundData{{Number: 1, Byes: []chesspairing.ByeEntry{{PlayerID: "A", Type: tc.typ}}}}, CurrentRound: 2}
			scores, err := s.Score(context.Background(), state)
			if err != nil {
				t.Fatal(err)
			}
			p := s.PointsForResult(chesspairing.ResultPending, chesspairing.ResultContext{PlayerRank: 1, PlayerValueNumber: 1, ByeType: &tc.typ})
			if scores[0].Score != tc.want || p != tc.want {
				t.Fatalf("Score() = %v, PointsForResult() = %v, want %v", scores[0].Score, p, tc.want)
			}
		})
	}
}

func TestPointsForResultByeAbsentAndExcusedScoreAsFirstAbsence(t *testing.T) {
	no := false
	s := New(Options{
		SelfVictory:             &no,
		AbsentFixedValue:        chesspairing.IntPtr(7),
		ExcusedAbsentFixedValue: chesspairing.IntPtr(9),
		AbsenceLimit:            chesspairing.IntPtr(1),
		AbsenceDecay:            chesspairing.BoolPtr(true),
	})

	cases := []struct {
		typ  chesspairing.ByeType
		want float64
	}{
		{chesspairing.ByeAbsent, 7},
		{chesspairing.ByeExcused, 9},
	}
	for _, tc := range cases {
		t.Run(tc.typ.String(), func(t *testing.T) {
			got := s.PointsForResult(chesspairing.ResultPending, chesspairing.ResultContext{PlayerRank: 1, PlayerValueNumber: 19, ByeType: &tc.typ})
			if got != tc.want {
				t.Fatalf("PointsForResult(%v) = %v, want %v", tc.typ, got, tc.want)
			}
		})
	}
}
