// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package standings_test

import (
	"context"
	"errors"
	"testing"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
	"github.com/gnutterts/chesspairing/standings"
	"github.com/gnutterts/chesspairing/tiebreaker"
)

func players(ids ...string) []cp.PlayerEntry {
	out := make([]cp.PlayerEntry, len(ids))
	for i, id := range ids {
		out[i] = cp.PlayerEntry{ID: id, DisplayName: id}
	}
	return out
}

func game(white, black string, r cp.GameResult) cp.GameData {
	return cp.GameData{
		WhiteID:   white,
		BlackID:   black,
		Result:    r,
		IsForfeit: r == cp.ResultForfeitWhiteWins || r == cp.ResultForfeitBlackWins || r == cp.ResultDoubleForfeit,
	}
}

func findStanding(rows []cp.Standing, id string) *cp.Standing {
	for i := range rows {
		if rows[i].PlayerID == id {
			return &rows[i]
		}
	}
	return nil
}

func TestBuildSingleRoundDecisive(t *testing.T) {
	state := &cp.TournamentState{
		Players: players("p1", "p2"),
		Rounds: []cp.RoundData{{
			Number: 1,
			Games:  []cp.GameData{game("p1", "p2", cp.ResultWhiteWins)},
		}},
		CurrentRound: 1,
	}
	rows, err := standings.Build(context.Background(), state, standard.New(standard.Options{}.WithDefaults()), nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if rows[0].PlayerID != "p1" || rows[0].Rank != 1 || rows[0].Wins != 1 || rows[0].Losses != 0 {
		t.Errorf("row 0: %+v", rows[0])
	}
	if rows[1].PlayerID != "p2" || rows[1].Rank != 2 || rows[1].Wins != 0 || rows[1].Losses != 1 {
		t.Errorf("row 1: %+v", rows[1])
	}
}

func TestBuildDrawSharedRank(t *testing.T) {
	state := &cp.TournamentState{
		Players: players("p1", "p2"),
		Rounds: []cp.RoundData{{
			Number: 1,
			Games:  []cp.GameData{game("p1", "p2", cp.ResultDraw)},
		}},
		CurrentRound: 1,
	}
	rows, err := standings.Build(context.Background(), state, standard.New(standard.Options{}.WithDefaults()), nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if rows[0].Rank != 1 || rows[1].Rank != 1 {
		t.Errorf("want both rank 1, got %d and %d", rows[0].Rank, rows[1].Rank)
	}
	for _, r := range rows {
		if r.Draws != 1 || r.Wins != 0 || r.Losses != 0 {
			t.Errorf("%s: want 1 draw, got %+v", r.PlayerID, r)
		}
	}
}

func TestBuildSingleForfeitCountsAsWinLoss(t *testing.T) {
	state := &cp.TournamentState{
		Players: players("p1", "p2"),
		Rounds: []cp.RoundData{{
			Number: 1,
			Games:  []cp.GameData{game("p1", "p2", cp.ResultForfeitWhiteWins)},
		}},
		CurrentRound: 1,
	}
	rows, err := standings.Build(context.Background(), state, standard.New(standard.Options{}.WithDefaults()), nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	w := findStanding(rows, "p1")
	l := findStanding(rows, "p2")
	if w == nil || l == nil {
		t.Fatal("missing player rows")
	}
	if w.Wins != 1 || w.GamesPlayed != 1 {
		t.Errorf("forfeit-winner: %+v", *w)
	}
	if l.Losses != 1 || l.GamesPlayed != 1 {
		t.Errorf("forfeit-loser: %+v", *l)
	}
}

// TestBuildDoubleForfeitZeroAcrossBoard is the visible behavior change versus
// the CLI's previous buildStandings: a double forfeit no longer increments
// GamesPlayed, and contributes 0 wins / 0 draws / 0 losses to both players.
func TestBuildDoubleForfeitZeroAcrossBoard(t *testing.T) {
	state := &cp.TournamentState{
		Players: players("p1", "p2"),
		Rounds: []cp.RoundData{{
			Number: 1,
			Games:  []cp.GameData{game("p1", "p2", cp.ResultDoubleForfeit)},
		}},
		CurrentRound: 1,
	}
	rows, err := standings.Build(context.Background(), state, standard.New(standard.Options{}.WithDefaults()), nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, r := range rows {
		if r.GamesPlayed != 0 || r.Wins != 0 || r.Draws != 0 || r.Losses != 0 {
			t.Errorf("%s: want all zero, got %+v", r.PlayerID, r)
		}
		if r.Score != 0 {
			t.Errorf("%s: want score 0, got %v", r.PlayerID, r.Score)
		}
	}
	// Both rank 1 (everyone tied at zero).
	if rows[0].Rank != 1 || rows[1].Rank != 1 {
		t.Errorf("ranks: %d, %d", rows[0].Rank, rows[1].Rank)
	}
}

func TestBuildPendingGameSkipped(t *testing.T) {
	state := &cp.TournamentState{
		Players: players("p1", "p2"),
		Rounds: []cp.RoundData{{
			Number: 1,
			Games:  []cp.GameData{game("p1", "p2", cp.ResultPending)},
		}},
		CurrentRound: 1,
	}
	rows, err := standings.Build(context.Background(), state, standard.New(standard.Options{}.WithDefaults()), nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, r := range rows {
		if r.GamesPlayed != 0 {
			t.Errorf("%s: pending should not count, got played=%d", r.PlayerID, r.GamesPlayed)
		}
	}
}

func TestBuildScoreTieBrokenByTiebreaker(t *testing.T) {
	// p1 beats p3, p2 beats p3, p1 draws p2.
	// p1 and p2 each: 1.5 points.
	// Buchholz: p1 opps {p2, p3} = 1.5 + 0; p2 opps {p1, p3} = 1.5 + 0.
	// Tied on Buchholz too. Use Sonneborn-Berger to differentiate would
	// also tie here, so instead construct a clearer asymmetry.
	//
	// Setup: p1 beats p2, p3 beats p4, p1 draws p3, p2 beats p4.
	// Round 1: p1-p2 1-0, p3-p4 1-0
	// Round 2: p1-p3 0.5-0.5, p2-p4 1-0
	// Scores: p1=1.5, p2=1.0, p3=1.5, p4=0
	// Buchholz: p1 opps={p2,p3}=1+1.5=2.5; p3 opps={p4,p1}=0+1.5=1.5
	// So p1 ranks above p3 by Buchholz.
	state := &cp.TournamentState{
		Players: players("p1", "p2", "p3", "p4"),
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{
				game("p1", "p2", cp.ResultWhiteWins),
				game("p3", "p4", cp.ResultWhiteWins),
			}},
			{Number: 2, Games: []cp.GameData{
				game("p1", "p3", cp.ResultDraw),
				game("p2", "p4", cp.ResultWhiteWins),
			}},
		},
		CurrentRound: 2,
	}
	bh, err := tiebreaker.Get("buchholz")
	if err != nil {
		t.Fatalf("get tiebreaker: %v", err)
	}
	rows, err := standings.Build(context.Background(), state,
		standard.New(standard.Options{}.WithDefaults()),
		[]cp.TieBreaker{bh})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if rows[0].PlayerID != "p1" || rows[0].Rank != 1 {
		t.Errorf("expected p1 rank 1, got %+v", rows[0])
	}
	if rows[1].PlayerID != "p3" || rows[1].Rank != 2 {
		t.Errorf("expected p3 rank 2, got %+v", rows[1])
	}
}

func TestBuildByIDResolvesRegistry(t *testing.T) {
	state := &cp.TournamentState{
		Players: players("p1", "p2"),
		Rounds: []cp.RoundData{{
			Number: 1,
			Games:  []cp.GameData{game("p1", "p2", cp.ResultWhiteWins)},
		}},
		CurrentRound: 1,
	}
	rows, err := standings.BuildByID(context.Background(), state,
		standard.New(standard.Options{}.WithDefaults()),
		[]string{"buchholz"})
	if err != nil {
		t.Fatalf("BuildByID: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if len(rows[0].TieBreakers) != 1 || rows[0].TieBreakers[0].ID != "buchholz" {
		t.Errorf("want one buchholz tiebreaker, got %+v", rows[0].TieBreakers)
	}
}

func TestBuildByIDUnknownIDError(t *testing.T) {
	state := &cp.TournamentState{Players: players("p1"), CurrentRound: 0}
	_, err := standings.BuildByID(context.Background(), state,
		standard.New(standard.Options{}.WithDefaults()),
		[]string{"definitely-not-a-real-tiebreaker"})
	if err == nil {
		t.Fatal("want error for unknown tiebreaker id")
	}
}

func TestBuildEmptyTieBreakerList(t *testing.T) {
	state := &cp.TournamentState{
		Players: players("p1", "p2"),
		Rounds: []cp.RoundData{{
			Number: 1,
			Games:  []cp.GameData{game("p1", "p2", cp.ResultWhiteWins)},
		}},
		CurrentRound: 1,
	}
	rows, err := standings.Build(context.Background(), state,
		standard.New(standard.Options{}.WithDefaults()), nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, r := range rows {
		if len(r.TieBreakers) != 0 {
			t.Errorf("%s: want no tiebreakers, got %d", r.PlayerID, len(r.TieBreakers))
		}
	}
}

func TestBuildNilStateError(t *testing.T) {
	_, err := standings.Build(context.Background(), nil,
		standard.New(standard.Options{}.WithDefaults()), nil)
	if err == nil {
		t.Fatal("want error for nil state")
	}
}

func TestBuildNilScorerError(t *testing.T) {
	state := &cp.TournamentState{Players: players("p1")}
	_, err := standings.Build(context.Background(), state, nil, nil)
	if err == nil {
		t.Fatal("want error for nil scorer")
	}
}

// failingScorer returns an error; used to verify error propagation.
type failingScorer struct{}

func (failingScorer) Score(_ context.Context, _ *cp.TournamentState) ([]cp.PlayerScore, error) {
	return nil, errors.New("scorer boom")
}

func (failingScorer) PointsForResult(_ cp.GameResult, _ cp.ResultContext) float64 { return 0 }

func TestBuildScorerErrorPropagated(t *testing.T) {
	state := &cp.TournamentState{Players: players("p1")}
	_, err := standings.Build(context.Background(), state, failingScorer{}, nil)
	if err == nil {
		t.Fatal("want propagated scorer error")
	}
}

// failingTB returns an error from Compute.
type failingTB struct{}

func (failingTB) ID() string   { return "boom" }
func (failingTB) Name() string { return "Boom" }
func (failingTB) Compute(_ context.Context, _ *cp.TournamentState, _ []cp.PlayerScore) ([]cp.TieBreakValue, error) {
	return nil, errors.New("tb boom")
}

func TestBuildTieBreakerErrorPropagated(t *testing.T) {
	state := &cp.TournamentState{
		Players: players("p1", "p2"),
		Rounds: []cp.RoundData{{
			Number: 1,
			Games:  []cp.GameData{game("p1", "p2", cp.ResultWhiteWins)},
		}},
		CurrentRound: 1,
	}
	_, err := standings.Build(context.Background(), state,
		standard.New(standard.Options{}.WithDefaults()),
		[]cp.TieBreaker{failingTB{}})
	if err == nil {
		t.Fatal("want propagated tiebreaker error")
	}
}

// TestBuildExcludesWithdrawnPlayer confirms a player marked
// WithdrawnAfterRound does not appear in the standings table for any
// round after the withdrawal — this is the user-visible side of the
// PlayerEntry.Active redesign.
func TestBuildExcludesWithdrawnPlayer(t *testing.T) {
	withdrawnAfter := 1
	state := &cp.TournamentState{
		Players: []cp.PlayerEntry{
			{ID: "p1", DisplayName: "p1"},
			{ID: "p2", DisplayName: "p2"},
			{ID: "p3", DisplayName: "p3", WithdrawnAfterRound: &withdrawnAfter},
		},
		Rounds: []cp.RoundData{
			{
				Number: 1,
				Games: []cp.GameData{
					game("p1", "p2", cp.ResultWhiteWins),
				},
				Byes: []cp.ByeEntry{{PlayerID: "p3", Type: cp.ByePAB}},
			},
			{
				Number: 2,
				Games:  []cp.GameData{game("p1", "p2", cp.ResultDraw)},
			},
		},
		CurrentRound: 2,
	}
	rows, err := standings.Build(context.Background(), state,
		standard.New(standard.Options{}.WithDefaults()), nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if findStanding(rows, "p3") != nil {
		t.Errorf("withdrawn player p3 must not appear in standings, got rows: %+v", rows)
	}
	if len(rows) != 2 {
		t.Errorf("want 2 rows, got %d: %+v", len(rows), rows)
	}
}
