// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
	"github.com/gnutterts/chesspairing/standings"
)

// TestToTournamentState_doubleForfeit verifies that a TRF game in which
// neither player appeared (forfeit loss "-" on both 001 lines) is read as
// chesspairing.ResultDoubleForfeit with IsForfeit set, not as a forfeit win
// for one side.
func TestToTournamentState_doubleForfeit(t *testing.T) {
	input := "001    1      Player One                        2000 NED                         0.0    1  0002 - -\n"
	input += "001    2      Player Two                        1800 NED                         0.0    2  0001 - -\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	if len(state.Rounds) != 1 || len(state.Rounds[0].Games) != 1 {
		t.Fatalf("unexpected round/game count: rounds=%d games=%d", len(state.Rounds), len(state.Rounds[0].Games))
	}

	g := state.Rounds[0].Games[0]
	if g.Result != chesspairing.ResultDoubleForfeit {
		t.Errorf("Result = %q, want %q", g.Result, chesspairing.ResultDoubleForfeit)
	}
	if !g.IsForfeit {
		t.Error("IsForfeit = false, want true")
	}
}

// TestValidate_doubleForfeitConsistent verifies that a double forfeit is not
// reported as a result conflict, both when the two 001 lines carry no color
// ("-") and when they carry opposite colors ("w"/"b").
func TestValidate_doubleForfeitConsistent(t *testing.T) {
	cases := []struct {
		name   string
		color1 Color
		color2 Color
	}{
		{"no colors", ColorNone, ColorNone},
		{"opposite colors", ColorWhite, ColorBlack},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := &Document{
				TournamentType: "Swiss Dutch",
				TotalRounds:    1,
				InitialColor:   "white1",
				Players: []PlayerLine{
					{
						StartNumber: 1,
						Name:        "Alpha",
						Rounds:      []RoundResult{{Opponent: 2, Color: tc.color1, Result: ResultForfeitLoss}},
					},
					{
						StartNumber: 2,
						Name:        "Beta",
						Rounds:      []RoundResult{{Opponent: 1, Color: tc.color2, Result: ResultForfeitLoss}},
					},
				},
			}

			if !areResultsConsistent(ResultForfeitLoss, ResultForfeitLoss) {
				t.Fatal("areResultsConsistent(ForfeitLoss, ForfeitLoss) = false, want true")
			}

			for _, issue := range doc.Validate(ValidatePairingEngine) {
				if strings.Contains(issue.Message, "result conflict") {
					t.Errorf("unexpected result conflict: %s", issue.Message)
				}
			}
		})
	}
}

// TestToTournamentState_inconsistentResults verifies that mutually
// inconsistent results fail loudly instead of silently crediting the first
// side with a win.
func TestToTournamentState_inconsistentResults(t *testing.T) {
	cases := []struct {
		name    string
		result1 string
		result2 string
	}{
		{"both win", "1", "1"},
		{"both forfeit win", "+", "+"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := "001    1      Player One                        2000 NED                         0.0    1  0002 w " + tc.result1 + "\n"
			input += "001    2      Player Two                        1800 NED                         0.0    2  0001 b " + tc.result2 + "\n"

			doc, err := Read(strings.NewReader(input))
			if err != nil {
				t.Fatalf("Read failed: %v", err)
			}

			_, err = doc.ToTournamentState()
			if err == nil {
				t.Fatal("ToTournamentState returned nil error, want inconsistent results error")
			}

			want := "trf: inconsistent results for players 1 and 2 in round 1"
			if err.Error() != want {
				t.Errorf("error = %q, want %q", err.Error(), want)
			}
		})
	}
}

// TestRoundTrip_doubleForfeit verifies the full TRF round trip: "- -" on both
// 001 lines becomes ResultDoubleForfeit and is written back as "- -" on both
// lines (forfeit loss, no color).
func TestRoundTrip_doubleForfeit(t *testing.T) {
	input := "001    1      Player One                        2000 NED                         0.0    1  0002 - -\n"
	input += "001    2      Player Two                        1800 NED                         0.0    2  0001 - -\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}
	if len(state.Rounds) != 1 || len(state.Rounds[0].Games) != 1 {
		t.Fatalf("unexpected round/game count")
	}
	if got := state.Rounds[0].Games[0].Result; got != chesspairing.ResultDoubleForfeit {
		t.Fatalf("Result = %q, want %q", got, chesspairing.ResultDoubleForfeit)
	}

	roundTripped, _ := FromTournamentState(state)
	var buf bytes.Buffer
	if err := Write(&buf, roundTripped); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	doc2, err := Read(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("Read of round-tripped TRF failed: %v\n%s", err, buf.String())
	}
	if len(doc2.Players) != 2 {
		t.Fatalf("round-tripped players = %d, want 2", len(doc2.Players))
	}

	for _, p := range doc2.Players {
		if len(p.Rounds) != 1 {
			t.Fatalf("player %d rounds = %d, want 1", p.StartNumber, len(p.Rounds))
		}
		rr := p.Rounds[0]
		if rr.Color != ColorNone {
			t.Errorf("player %d color = %v, want ColorNone", p.StartNumber, rr.Color)
		}
		if rr.Result != ResultForfeitLoss {
			t.Errorf("player %d result = %v, want ResultForfeitLoss", p.StartNumber, rr.Result)
		}
	}
}

// TestStandings_doubleForfeitScoresZero verifies that both players in a
// double forfeit receive zero points and no win/draw/loss/game count.
func TestStandings_doubleForfeitScoresZero(t *testing.T) {
	doc := &Document{
		TournamentType: "Swiss Dutch",
		TotalRounds:    1,
		InitialColor:   "white1",
		Players: []PlayerLine{
			{
				StartNumber: 1,
				Name:        "Alpha",
				Rating:      2000,
				Rounds:      []RoundResult{{Opponent: 2, Color: ColorWhite, Result: ResultForfeitLoss}},
			},
			{
				StartNumber: 2,
				Name:        "Beta",
				Rating:      1800,
				Rounds:      []RoundResult{{Opponent: 1, Color: ColorBlack, Result: ResultForfeitLoss}},
			},
		},
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	scorer := standard.New(standard.Options{})
	rows, err := standings.Build(context.Background(), state, scorer, nil)
	if err != nil {
		t.Fatalf("standings.Build failed: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("standings rows = %d, want 2", len(rows))
	}

	for _, row := range rows {
		if row.Score != 0 {
			t.Errorf("player %s score = %v, want 0", row.PlayerID, row.Score)
		}
		if row.Wins != 0 || row.Draws != 0 || row.Losses != 0 || row.GamesPlayed != 0 {
			t.Errorf("player %s W/D/L/games = %d/%d/%d/%d, want all 0",
				row.PlayerID, row.Wins, row.Draws, row.Losses, row.GamesPlayed)
		}
	}
}
