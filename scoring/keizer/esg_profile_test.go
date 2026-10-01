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
// (https://esgemmen.nl). The export's field names are preserved so it can be
// decoded without altering the supplied fixture.
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
			Result int    `json:"uitslag"`
		} `json:"partijen"`
		Absences []struct {
			Player string  `json:"speler"`
			Reason int     `json:"reden"`
			Via    *string `json:"via"`
		} `json:"afwezig"`
	} `json:"rondes"`
	Standings []struct {
		Round   int `json:"ronde"`
		Players []struct {
			Rank  int     `json:"pos"`
			ID    string  `json:"id"`
			Score float64 `json:"score"`
			Value int     `json:"waarde"`
		} `json:"spelers"`
	} `json:"ranglijst"`
}

func TestESGProfileAgainstSevilla(t *testing.T) {
	// Source: testdata/esg_r2_anoniem.json, ranglijst[ronde 1] and
	// ranglijst[ronde 2]. Scores, positions and value numbers are read directly
	// from those supplied tables.
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
		// Keep late joiner P25 active when reproducing the round-1 table.
		// Its JoinedRound then awards the documented missed-round handicap.
		current.CurrentRound = len(state.Rounds)
		scores, err := New(ESGOptions()).Score(context.Background(), &current)
		if err != nil {
			t.Fatalf("round %d Score: %v", round, err)
		}
		actual := make(map[string]chesspairing.PlayerScore, len(scores))
		for _, score := range scores {
			actual[score.PlayerID] = score
		}
		sevilla := fixture.Standings[round-1].Players
		if fixture.Standings[round-1].Round != round-1 || len(sevilla) != len(fixture.Players) {
			t.Fatalf("round %d standings are not a complete supplied table", round)
		}
		for _, expected := range sevilla {
			got, ok := actual[expected.ID]
			if !ok {
				t.Fatalf("round %d player %s missing from score", round, expected.ID)
			}
			if got.Score != expected.Score {
				t.Errorf("round %d player %s: score = %g, want Sevilla %g", round, expected.ID, got.Score, expected.Score)
			}
			if got.Rank-1 != expected.Rank {
				t.Errorf("round %d player %s: position = %d, want Sevilla %d", round, expected.ID, got.Rank-1, expected.Rank)
			}
			if value := 61 - got.Rank; value != expected.Value {
				t.Errorf("round %d player %s: value number = %d, want Sevilla %d", round, expected.ID, value, expected.Value)
			}
		}
	}
}

func esgState(t *testing.T, fixture esgFixture) *chesspairing.TournamentState {
	t.Helper()
	players := make([]chesspairing.PlayerEntry, len(fixture.Players))
	for i, player := range fixture.Players {
		displayName := player.ID
		// Source: round-1 ranglijst vorigeWaarde values order the equally rated
		// players P22, P23, P21, P24, P25. Display names supply that otherwise
		// unavailable start-order tiebreak to the library.
		switch player.ID {
		case "P22":
			displayName = "P21a"
		case "P23":
			displayName = "P21b"
		case "P21":
			displayName = "P21c"
		case "P24":
			displayName = "P21d"
		}
		players[i] = chesspairing.PlayerEntry{ID: player.ID, DisplayName: displayName, Rating: player.Rating}
		if player.ID == "P25" {
			// Source: round 1 afwezig entry with via "splNieuw"; JoinedRound
			// makes its missed first round use the ESG late-join handicap of 15.
			players[i].JoinedRound = 2
		}
	}
	rounds := make([]chesspairing.RoundData, len(fixture.Rounds))
	for i, source := range fixture.Rounds {
		round := chesspairing.RoundData{Number: source.Number + 1}
		for _, game := range source.Games {
			result, ok := esgResult(game.Result)
			if !ok {
				t.Fatalf("round %d game %s-%s has unsupported result %d", source.Number, game.White, game.Black, game.Result)
			}
			round.Games = append(round.Games, chesspairing.GameData{WhiteID: game.White, BlackID: game.Black, Result: result})
		}
		for _, absence := range source.Absences {
			switch {
			case absence.Via != nil && *absence.Via == "splExtern":
				round.Byes = append(round.Byes, chesspairing.ByeEntry{PlayerID: absence.Player, Type: chesspairing.ByeClubCommitment})
			case absence.Via != nil && *absence.Via == "splAfwezig":
				round.Byes = append(round.Byes, chesspairing.ByeEntry{PlayerID: absence.Player, Type: chesspairing.ByeAbsent})
			case absence.Via != nil && *absence.Via == "splNieuw":
				// Deliberately no bye: JoinedRound handles this missed round.
			case absence.Via != nil && *absence.Via == "splBye":
				round.Byes = append(round.Byes, chesspairing.ByeEntry{PlayerID: absence.Player, Type: chesspairing.ByePAB})
			default:
				t.Fatalf("round %d player %s has unmapped absence", source.Number, absence.Player)
			}
		}
		rounds[i] = round
	}
	return &chesspairing.TournamentState{Players: players, Rounds: rounds, CurrentRound: len(rounds)}
}

func esgResult(result int) (chesspairing.GameResult, bool) {
	switch result {
	case 1:
		return chesspairing.ResultWhiteWins, true
	case 3:
		return chesspairing.ResultBlackWins, true
	default:
		return chesspairing.ResultPending, false
	}
}
