// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// This fixture is an anonymised export of the ESG Keizer competition
// (https://esgemmen.nl). The two Sevilla input errors, "Geen lid" 11.7 → 15
// and bye 24 → 40, remain explicit exceptions below.
//
// The export's field names are preserved so it can be decoded without altering
// the supplied fixture.
type esgFixture struct {
	Players []struct {
		ID     string `json:"id"`
		Rating int    `json:"rating"`
	} `json:"spelers"`
	Rounds []struct {
		Number int `json:"ronde"`
		Games  []struct {
			White  string `json:"wit"`
			Black  string `json:"zwart"`
			Result string `json:"uitslag"`
		} `json:"partijen"`
		Absences []struct {
			Player string  `json:"speler"`
			Reason string  `json:"reden"`
			Via    *string `json:"via"`
		} `json:"afwezig"`
	} `json:"rondes"`
	Standings []struct {
		Round   int `json:"ronde"`
		Players []struct {
			Rank  int     `json:"pos"`
			ID    string  `json:"id"`
			Score float64 `json:"score"`
		} `json:"spelers"`
	} `json:"ranglijst"`
}

func TestESGProfileAgainstSevilla(t *testing.T) {
	// Source: testdata/esg_r2_anoniem.json, ranglijst[ronde 1] and
	// ranglijst[ronde 2]. Scores and Sevilla positions are read directly from
	// those supplied tables; the two documented input errors are corrected to
	// 15 ("Geen lid", round 1) and 40 ("Handmatige paring bye", round 2).
	data, err := os.ReadFile("testdata/esg_r2_anoniem.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture esgFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rounds) != 2 || len(fixture.Standings) != 2 {
		t.Fatalf("fixture has %d rounds and %d standings, want 2 each", len(fixture.Rounds), len(fixture.Standings))
	}

	state := esgState(t, fixture)
	for round := 1; round <= 2; round++ {
		current := *state
		current.Rounds = state.Rounds[:round]
		// Keep late joiner P26 active when reproducing the round-1 table.
		// Its JoinedRound then awards the documented missed-round handicap.
		current.CurrentRound = len(state.Rounds)
		scores, err := New(esgOptions()).Score(context.Background(), &current)
		if err != nil {
			t.Fatalf("round %d Score: %v", round, err)
		}
		actual := make(map[string]chesspairing.PlayerScore, len(scores))
		for _, score := range scores {
			actual[score.PlayerID] = score
		}
		sevilla := fixture.Standings[round-1].Players
		if fixture.Standings[round-1].Round != round || len(sevilla) != len(fixture.Players) {
			t.Fatalf("round %d standings are not a complete supplied table", round)
		}
		for _, expected := range sevilla {
			want := expected.Score
			if expected.ID == "P26" {
				// Its corrected round-1 handicap remains its round-2 total.
				want = 15
			}
			if round == 2 && expected.ID == "P22" {
				want = 40
			}
			got, ok := actual[expected.ID]
			if !ok {
				t.Fatalf("round %d player %s missing from score", round, expected.ID)
			}
			if got.Score != want {
				t.Errorf("round %d player %s: score = %g, want Sevilla/FIDE %g", round, expected.ID, got.Score, want)
			}
		}
	}
}

func esgOptions() Options {
	base, step, absent, club, bye, limit := 60, 1, 20, 40, 40, 5
	selfVictory, frozen, decay := false, true, false
	lateJoin := 15.0
	return Options{
		ValueNumberBase:          &base,
		ValueNumberStep:          &step,
		SelfVictory:              &selfVictory,
		Frozen:                   &frozen,
		AbsentFixedValue:         &absent,
		ClubCommitmentFixedValue: &club,
		ByeFixedValue:            &bye,
		AbsenceLimit:             &limit,
		AbsenceDecay:             &decay,
		LateJoinHandicap:         &lateJoin,
	}
}

func esgState(t *testing.T, fixture esgFixture) *chesspairing.TournamentState {
	t.Helper()
	players := make([]chesspairing.PlayerEntry, len(fixture.Players))
	for i, player := range fixture.Players {
		displayName := player.ID
		// Source: round-1 ranglijst vorigeWaarde values order the equally rated
		// players P23, P24, P22, P25, P26. Display names supply that otherwise
		// unavailable start-order tiebreak to the library.
		switch player.ID {
		case "P23":
			displayName = "P22a"
		case "P24":
			displayName = "P22b"
		case "P22":
			displayName = "P22c"
		}
		players[i] = chesspairing.PlayerEntry{ID: player.ID, DisplayName: displayName, Rating: player.Rating}
		if player.ID == "P26" {
			// Source: round 1 afwezig entry "Geen lid"; JoinedRound makes its
			// missed first round use the ESG late-join handicap of 15.
			players[i].JoinedRound = 2
		}
	}
	rounds := make([]chesspairing.RoundData, len(fixture.Rounds))
	for i, source := range fixture.Rounds {
		round := chesspairing.RoundData{Number: source.Number}
		for _, game := range source.Games {
			result, ok := esgResult(game.Result)
			if !ok {
				t.Fatalf("round %d game %s-%s has unsupported result %q", source.Number, game.White, game.Black, game.Result)
			}
			round.Games = append(round.Games, chesspairing.GameData{WhiteID: game.White, BlackID: game.Black, Result: result})
		}
		for _, absence := range source.Absences {
			switch {
			case absence.Via != nil && *absence.Via == "splExtern":
				round.Byes = append(round.Byes, chesspairing.ByeEntry{PlayerID: absence.Player, Type: chesspairing.ByeClubCommitment})
			case absence.Via != nil && *absence.Via == "splAfwezig":
				round.Byes = append(round.Byes, chesspairing.ByeEntry{PlayerID: absence.Player, Type: chesspairing.ByeAbsent})
			case absence.Reason == "Geen lid":
				// Deliberately no bye: JoinedRound handles this missed round.
			case absence.Reason == "Handmatige paring bye":
				round.Byes = append(round.Byes, chesspairing.ByeEntry{PlayerID: absence.Player, Type: chesspairing.ByePAB})
			default:
				t.Fatalf("round %d player %s has unmapped absence", source.Number, absence.Player)
			}
		}
		rounds[i] = round
	}
	return &chesspairing.TournamentState{Players: players, Rounds: rounds, CurrentRound: len(rounds)}
}

func esgResult(result string) (chesspairing.GameResult, bool) {
	switch result {
	case "1-0":
		return chesspairing.ResultWhiteWins, true
	case "0-1":
		return chesspairing.ResultBlackWins, true
	case "½-½":
		return chesspairing.ResultDraw, true
	default:
		return chesspairing.ResultPending, false
	}
}
