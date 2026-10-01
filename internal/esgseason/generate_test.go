// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package esgseason

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/gnutterts/chesspairing"
	keizerscoring "github.com/gnutterts/chesspairing/scoring/keizer"
)

const testdata = "testdata"

func TestSeason(t *testing.T) {
	if os.Getenv("ESGSEASON_UPDATE") == "1" {
		if err := Update(testdata); err != nil {
			t.Fatal(err)
		}
	}
	season := readJSON[Season](t, "season.json")
	expected := readJSON[Expected](t, "expected.json")
	generatedSeason, generatedExpected, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(season, generatedSeason) {
		t.Error("season.json is not the deterministic generated season; run ESGSEASON_UPDATE=1 go test ./internal/esgseason -run TestSeason -count=1")
	}
	if !reflect.DeepEqual(expected, generatedExpected) {
		t.Error("expected.json is not the deterministic generated standings; run ESGSEASON_UPDATE=1 go test ./internal/esgseason -run TestSeason -count=1")
	}
	assertInvariants(t, season, expected)
}

func readJSON[T any](t *testing.T, name string) T {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(testdata, name))
	if err != nil {
		t.Fatal(err)
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func assertInvariants(t *testing.T, season Season, expected Expected) {
	t.Helper()
	if len(season.Players) != 26 || len(season.Rounds) != rounds || len(expected.Rounds) != rounds {
		t.Fatalf("season has %d players, %d rounds and %d standings, want 26, 24 and 24", len(season.Players), len(season.Rounds), len(expected.Rounds))
	}
	players := playersByID(season.Players)
	pairingNumbers := pairingNumbers(t, season.Players)
	met := make(map[string]int)
	byes := make(map[string]int)
	absences := make(map[string]int)
	previous := make(map[string]ExpectedPlayer)
	for index, round := range season.Rounds {
		if round.Number != index+1 {
			t.Errorf("round index %d has number %d", index, round.Number)
		}
		if len(expected.Rounds[index].Standings) != len(season.Players) {
			t.Errorf("round %d has %d standings, want %d", round.Number, len(expected.Rounds[index].Standings), len(season.Players))
		}
		current := standingsByID(t, round.Number, expected.Rounds[index].Standings)
		assertStandings(t, round.Number, current, pairingNumbers)
		assertRound(t, round, players, previous, current, met, byes, absences)
		previous = current
	}
	if absences["P05"] <= 5 {
		t.Errorf("P05 has %d absences, want more than 5 to exercise the limit", absences["P05"])
	}
	assertRoundEightAttendance(t, season, players)
	if !hasDoubleRePair(season) {
		t.Error("the both-absent pair is not paired again in its period")
	}
	assertScores(t, season, expected)
}

func assertRound(t *testing.T, round Round, players map[string]Player, previous, current map[string]ExpectedPlayer, met map[string]int, byes, absences map[string]int) {
	t.Helper()
	seen := make(map[string]bool)
	contributions := make(map[string]float64, len(players))
	for id, player := range players {
		if player.JoinedRound > round.Number {
			contributions[id] = 15
		}
		if player.WithdrawnAfterRound != 0 && round.Number > player.WithdrawnAfterRound {
			absences[id]++
			contributions[id] = absencePoints(absences[id])
		}
	}
	for _, event := range round.Events {
		switch event.Kind {
		case "absent":
			for _, id := range event.Players {
				if seen[id] {
					t.Errorf("round %d absent player %s was paired", round.Number, id)
				}
				seen[id] = true
				absences[id]++
				contributions[id] = absencePoints(absences[id])
			}
		case "external":
			for _, id := range event.Players {
				seen[id] = true
				contributions[id] = 40
			}
		case "bye":
			if len(event.Players) != 1 {
				t.Errorf("round %d bye has %d players", round.Number, len(event.Players))
				continue
			}
			id := event.Players[0]
			seen[id] = true
			contributions[id] = 40
			assertByeFromBottom(t, round, id, players, previous, byes)
			byes[id]++
		case "forfeit":
			if len(event.Players) != 1 {
				t.Errorf("round %d forfeit has %d players", round.Number, len(event.Players))
			}
		case "bothAbsent":
			if len(event.Players) != 2 {
				t.Errorf("round %d both-absent event has %d players", round.Number, len(event.Players))
				continue
			}
			for _, id := range event.Players {
				absences[id]++
				contributions[id] = absencePoints(absences[id])
			}
		case "newcomer", "withdrawn":
			// The player metadata determines these score contributions.
		default:
			t.Errorf("round %d has unknown event %q", round.Number, event.Kind)
		}
	}
	for _, pairing := range round.Pairings {
		if seen[pairing.White] || seen[pairing.Black] {
			t.Errorf("round %d pairs a player more than once", round.Number)
		}
		seen[pairing.White], seen[pairing.Black] = true, true
		result, err := parseResult(pairing.Result)
		if err != nil {
			t.Errorf("round %d pairing %s-%s: %v", round.Number, pairing.White, pairing.Black, err)
			continue
		}
		key := pairKey(pairing.White, pairing.Black)
		if hasEvent(round.Events, "bothAbsent", pairing.White, pairing.Black) {
			if result != chesspairing.ResultDoubleForfeit {
				t.Errorf("round %d both-absent game has result %s", round.Number, pairing.Result)
			}
			if met[key] == round.Number {
				t.Errorf("round %d both-absent pair was recorded as met", round.Number)
			}
			continue
		}
		if previousRound, ok := met[key]; ok {
			if (round.Number-1)/6 == (previousRound-1)/6 {
				t.Errorf("pair %s repeats in period at rounds %d and %d", key, previousRound, round.Number)
			}
			if round.Number-previousRound < 2 {
				t.Errorf("pair %s repeats without a round between at rounds %d and %d", key, previousRound, round.Number)
			}
		}
		met[key] = round.Number
		if hasEvent(round.Events, "forfeit", pairing.White) {
			gain := valueNumber(previous, pairing.White, players)
			contributions[pairing.Black] = gain
			if current[pairing.White].Points-previous[pairing.White].Points != 0 {
				t.Errorf("round %d forfeiting player %s does not score 0", round.Number, pairing.White)
			}
			if current[pairing.Black].Points-previous[pairing.Black].Points != gain {
				t.Errorf("round %d forfeit winner %s gain %g, want forfeiter value number %g", round.Number, pairing.Black, current[pairing.Black].Points-previous[pairing.Black].Points, gain)
			}
			continue
		}
		whiteValue := valueNumber(previous, pairing.White, players)
		blackValue := valueNumber(previous, pairing.Black, players)
		switch result {
		case chesspairing.ResultWhiteWins:
			contributions[pairing.White] = blackValue
		case chesspairing.ResultDraw:
			contributions[pairing.White] = blackValue / 2
			contributions[pairing.Black] = whiteValue / 2
		case chesspairing.ResultBlackWins:
			contributions[pairing.Black] = whiteValue
		default:
			t.Errorf("round %d pairing %s-%s has unexpected result %s", round.Number, pairing.White, pairing.Black, pairing.Result)
		}
	}
	for _, id := range round.Present {
		if !seen[id] {
			t.Errorf("round %d present player %s is neither paired nor assigned a bye", round.Number, id)
		}
	}
	for id := range players {
		if _, ok := current[id]; !ok {
			t.Errorf("round %d withdrawn or inactive player %s is missing from the standings", round.Number, id)
			continue
		}
		got := current[id].Points - previous[id].Points
		if got != contributions[id] {
			t.Errorf("round %d player %s score gain %g, want %g", round.Number, id, got, contributions[id])
		}
	}
}

func assertByeFromBottom(t *testing.T, round Round, bye string, players map[string]Player, previous map[string]ExpectedPlayer, byes map[string]int) {
	t.Helper()
	eligible := append([]string(nil), round.Present...)
	eligible = append(eligible, bye)
	fewest := int(^uint(0) >> 1)
	for _, id := range eligible {
		fewest = min(fewest, byes[id])
	}
	want := ""
	for _, id := range eligible {
		if byes[id] == fewest && (want == "" || rankBeforeRound(id, players, previous) > rankBeforeRound(want, players, previous)) {
			want = id
		}
	}
	if bye != want {
		t.Errorf("round %d bye is %s, want lowest-ranked eligible player %s", round.Number, bye, want)
	}
}

func rankBeforeRound(id string, players map[string]Player, previous map[string]ExpectedPlayer) int {
	if standing, ok := previous[id]; ok {
		return standing.Position
	}
	ranked := make([]Player, 0, len(players))
	for _, player := range players {
		ranked = append(ranked, player)
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Rating != ranked[j].Rating {
			return ranked[i].Rating > ranked[j].Rating
		}
		return ranked[i].EntryOrder < ranked[j].EntryOrder
	})
	for index, player := range ranked {
		if player.ID == id {
			return index + 1
		}
	}
	return 0
}

func absencePoints(count int) float64 {
	if count <= 5 {
		return 20
	}
	return 0
}

func valueNumber(previous map[string]ExpectedPlayer, id string, players map[string]Player) float64 {
	if standing, ok := previous[id]; ok {
		return float64(standing.ValueNumber)
	}
	return float64(61 - rankBeforeRound(id, players, previous))
}

func assertStandings(t *testing.T, round int, standings map[string]ExpectedPlayer, pairingNumbers map[string]int) {
	t.Helper()
	for id, standing := range standings {
		if standing.ValueNumber != 61-standing.Position {
			t.Errorf("round %d player %s value number %d, want %d", round, id, standing.ValueNumber, 61-standing.Position)
		}
	}
	for id, standing := range standings {
		for otherID, other := range standings {
			if id != otherID && standing.Points == other.Points && pairingNumbers[id] < pairingNumbers[otherID] && standing.Position >= other.Position {
				t.Errorf("round %d equal-score players %s and %s are not ordered by pairing number", round, id, otherID)
			}
		}
	}
}

func playersByID(players []Player) map[string]Player {
	result := make(map[string]Player, len(players))
	for _, player := range players {
		result[player.ID] = player
	}
	return result
}

func pairingNumbers(t *testing.T, players []Player) map[string]int {
	t.Helper()
	entries, err := chesspairing.AssignPairingNumbers(entries(players))
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]int, len(entries))
	for _, entry := range entries {
		result[entry.ID] = entry.PairingNumber
	}
	return result
}

func standingsByID(t *testing.T, round int, standings []ExpectedPlayer) map[string]ExpectedPlayer {
	t.Helper()
	result := make(map[string]ExpectedPlayer, len(standings))
	for _, standing := range standings {
		if _, ok := result[standing.ID]; ok {
			t.Errorf("round %d lists player %s twice", round, standing.ID)
		}
		result[standing.ID] = standing
	}
	return result
}

func assertScores(t *testing.T, season Season, expected Expected) {
	players := entries(season.Players)
	state := chesspairing.TournamentState{Players: players, CurrentRound: rounds + 1}
	for i, round := range season.Rounds {
		data := chesspairing.RoundData{Number: round.Number}
		for _, pairing := range round.Pairings {
			result, err := parseResult(pairing.Result)
			if err != nil {
				t.Fatalf("round %d pairing %s-%s: %v", round.Number, pairing.White, pairing.Black, err)
			}
			data.Games = append(data.Games, chesspairing.GameData{WhiteID: pairing.White, BlackID: pairing.Black, Result: result, IsForfeit: hasEvent(round.Events, "forfeit", pairing.White)})
		}
		for _, event := range round.Events {
			switch event.Kind {
			case "absent":
				for _, id := range event.Players {
					data.Byes = append(data.Byes, chesspairing.ByeEntry{PlayerID: id, Type: chesspairing.ByeAbsent})
				}
			case "external":
				for _, id := range event.Players {
					data.Byes = append(data.Byes, chesspairing.ByeEntry{PlayerID: id, Type: chesspairing.ByeClubCommitment})
				}
			case "bye":
				data.Byes = append(data.Byes, chesspairing.ByeEntry{PlayerID: event.Players[0], Type: chesspairing.ByePAB})
			}
		}
		state.Rounds = append(state.Rounds, data)
		scores, err := keizerscoring.New(keizerscoring.ESGOptions()).Score(context.Background(), &state)
		if err != nil {
			t.Fatal(err)
		}
		if i == 5 || i == 11 || i == 17 || i == 23 {
			if !reflect.DeepEqual(expectedRound(round.Number, scores), expected.Rounds[i]) {
				t.Errorf("round %d standings differ from expected.json", round.Number)
			}
		}
	}
}

func assertRoundEightAttendance(t *testing.T, season Season, players map[string]Player) {
	t.Helper()
	bothAbsent := eventPlayers(season.Rounds[6].Events, "bothAbsent")
	absent := eventPlayers(season.Rounds[7].Events, "absent")
	if !reflect.DeepEqual(absent, bothAbsent) {
		t.Errorf("round 8 absent players %v, want round-7 both-absent players %v", absent, bothAbsent)
	}
	if len(season.Rounds[7].Present) != len(players)-len(bothAbsent) {
		t.Errorf("round 8 has %d present players, want %d", len(season.Rounds[7].Present), len(players)-len(bothAbsent))
	}
}

func eventPlayers(events []Event, kind string) []string {
	var players []string
	for _, event := range events {
		if event.Kind == kind {
			players = append(players, event.Players...)
		}
	}
	return players
}

func hasDoubleRePair(season Season) bool {
	for _, round := range season.Rounds {
		for _, event := range round.Events {
			if event.Kind != "bothAbsent" {
				continue
			}
			key := pairKey(event.Players[0], event.Players[1])
			for _, later := range season.Rounds[round.Number:] {
				if (later.Number-1)/6 != (round.Number-1)/6 {
					break
				}
				for _, pairing := range later.Pairings {
					if pairKey(pairing.White, pairing.Black) == key {
						return true
					}
				}
			}
		}
	}
	return false
}

func hasEvent(events []Event, kind string, ids ...string) bool {
	for _, event := range events {
		if event.Kind == kind && reflect.DeepEqual(event.Players, ids) {
			return true
		}
	}
	return false
}

func pairKey(first, second string) string {
	if first < second {
		return first + ":" + second
	}
	return second + ":" + first
}

func parseResult(result string) (chesspairing.GameResult, error) {
	switch result {
	case "1-0":
		return chesspairing.ResultWhiteWins, nil
	case "½-½":
		return chesspairing.ResultDraw, nil
	case "0-1":
		return chesspairing.ResultBlackWins, nil
	case "0-0f":
		return chesspairing.ResultDoubleForfeit, nil
	default:
		return chesspairing.GameResult(""), fmt.Errorf("unknown result %q", result)
	}
}
