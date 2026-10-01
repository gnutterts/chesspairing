// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package esgseason generates the deterministic ESG club-season regression.
package esgseason

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/dutch"
	keizerpairing "github.com/gnutterts/chesspairing/pairing/keizer"
	keizerscoring "github.com/gnutterts/chesspairing/scoring/keizer"
)

const rounds = 24

type Season struct {
	Players []Player `json:"players"`
	Rounds  []Round  `json:"rounds"`
}

type Player struct {
	ID                  string `json:"id"`
	Rating              int    `json:"rating"`
	EntryOrder          int    `json:"entryOrder"`
	JoinedRound         int    `json:"joinedRound,omitempty"`
	WithdrawnAfterRound int    `json:"withdrawnAfterRound,omitempty"`
}

type Round struct {
	Number   int       `json:"number"`
	Present  []string  `json:"present"`
	Events   []Event   `json:"events,omitempty"`
	Pairings []Pairing `json:"pairings"`
}

type Event struct {
	Kind    string   `json:"kind"`
	Players []string `json:"players"`
}

type Pairing struct {
	White  string `json:"white"`
	Black  string `json:"black"`
	Result string `json:"result"`
}

type Expected struct {
	Rounds []ExpectedRound `json:"rounds"`
}

type ExpectedRound struct {
	Number    int              `json:"number"`
	Standings []ExpectedPlayer `json:"standings"`
}

type ExpectedPlayer struct {
	ID          string  `json:"id"`
	Points      float64 `json:"points"`
	ValueNumber int     `json:"valueNumber"`
	Position    int     `json:"position"`
}

// Generate constructs the fixed season and its Keizer standings.
func Generate() (Season, Expected, error) {
	season := Season{Players: players()}
	state := chesspairing.TournamentState{Players: entries(season.Players)}
	random := rand.New(rand.NewPCG(20260930, 185902))
	expected := Expected{}
	var doublePair []string

	for number := 1; number <= rounds; number++ {
		state.CurrentRound = number
		events := scriptedEvents(number, doublePair)
		state.PreAssignedByes = preassigned(events)
		pairer := chesspairing.Pairer(keizerpairing.New(keizerpairing.ESGOptions()))
		if number <= 4 {
			pairer = dutch.New(dutch.Options{})
		}
		result, err := pairer.Pair(context.Background(), &state)
		if err != nil {
			return Season{}, Expected{}, fmt.Errorf("round %d pair: %w", number, err)
		}
		games := make([]chesspairing.GameData, len(result.Pairings))
		pairings := make([]Pairing, len(result.Pairings))
		for i, pairing := range result.Pairings {
			gameResult := scriptedResult(random)
			if number == 7 && i == 0 {
				gameResult = chesspairing.ResultDoubleForfeit
				doublePair = []string{pairing.WhiteID, pairing.BlackID}
				events = append(events, Event{Kind: "bothAbsent", Players: doublePair})
			}
			if number == 24 && i == 0 {
				gameResult = chesspairing.ResultBlackWins
				games[i].IsForfeit = true
				events = append(events, Event{Kind: "forfeit", Players: []string{pairing.WhiteID}})
			}
			games[i].WhiteID, games[i].BlackID, games[i].Result = pairing.WhiteID, pairing.BlackID, gameResult
			name, err := resultName(gameResult)
			if err != nil {
				return Season{}, Expected{}, fmt.Errorf("round %d pairing %s-%s: %w", number, pairing.WhiteID, pairing.BlackID, err)
			}
			pairings[i] = Pairing{White: pairing.WhiteID, Black: pairing.BlackID, Result: name}
		}
		for _, bye := range result.Byes {
			if bye.Type == chesspairing.ByePAB {
				events = append(events, Event{Kind: "bye", Players: []string{bye.PlayerID}})
			}
		}
		round := chesspairing.RoundData{Number: number, Games: games, Byes: result.Byes}
		state.Rounds = append(state.Rounds, round)
		// Score against the complete enrolment so the newcomer remains in every
		// historical table and receives its missed-round handicap.
		state.CurrentRound = rounds + 1
		scores, err := keizerscoring.New(keizerscoring.ESGOptions()).Score(context.Background(), &state)
		if err != nil {
			return Season{}, Expected{}, fmt.Errorf("round %d score: %w", number, err)
		}
		expected.Rounds = append(expected.Rounds, expectedRound(number, scores))
		season.Rounds = append(season.Rounds, Round{Number: number, Present: present(state.Players, result.Byes, number), Events: events, Pairings: pairings})
	}
	return season, expected, nil
}

func players() []Player {
	players := make([]Player, 26)
	for i := range players {
		players[i] = Player{ID: fmt.Sprintf("P%02d", i+1), Rating: 2100 - i*31, EntryOrder: i + 1}
	}
	players[24].Rating = 0
	players[24].JoinedRound = 3
	players[25].WithdrawnAfterRound = 15
	return players
}

func entries(players []Player) []chesspairing.PlayerEntry {
	entries := make([]chesspairing.PlayerEntry, len(players))
	for i, player := range players {
		entries[i] = chesspairing.PlayerEntry{ID: player.ID, DisplayName: player.ID, Rating: player.Rating, JoinedRound: player.JoinedRound}
		if player.WithdrawnAfterRound != 0 {
			withdrawn := player.WithdrawnAfterRound
			entries[i].WithdrawnAfterRound = &withdrawn
		}
	}
	return entries
}

func scriptedEvents(round int, doublePair []string) []Event {
	var events []Event
	switch round {
	case 2:
		events = append(events, Event{Kind: "absent", Players: []string{"P05"}})
	case 4, 6, 10, 14:
		events = append(events, Event{Kind: "absent", Players: []string{"P05", fmt.Sprintf("P%02d", round/2+5)}})
	case 18, 22:
		events = append(events, Event{Kind: "absent", Players: []string{"P05"}})
	case 20:
		events = append(events, Event{Kind: "absent", Players: []string{"P09"}})
	case 23, 24:
		events = append(events, Event{Kind: "absent", Players: []string{"P05"}})
	case 3, 11:
		events = append(events, Event{Kind: "external", Players: []string{fmt.Sprintf("P%02d", round/2+3)}})
		events = append(events, Event{Kind: "absent", Players: []string{fmt.Sprintf("P%02d", round/2+4)}})
	case 19:
		events = append(events, Event{Kind: "external", Players: []string{"P12"}})
	case 8:
		// The round-7 double-forfeit players are both absent; everyone else
		// attends normally. They can therefore be paired again later this period.
		for _, id := range doublePair {
			events = append(events, Event{Kind: "absent", Players: []string{id}})
		}
	}
	if round == 3 {
		events = append(events, Event{Kind: "newcomer", Players: []string{"P25"}})
	}
	if round == 16 {
		events = append(events, Event{Kind: "withdrawn", Players: []string{"P26"}})
	}
	return events
}

func preassigned(events []Event) []chesspairing.ByeEntry {
	var byes []chesspairing.ByeEntry
	for _, event := range events {
		switch event.Kind {
		case "absent":
			for _, id := range event.Players {
				byes = append(byes, chesspairing.ByeEntry{PlayerID: id, Type: chesspairing.ByeAbsent})
			}
		case "external":
			for _, id := range event.Players {
				byes = append(byes, chesspairing.ByeEntry{PlayerID: id, Type: chesspairing.ByeClubCommitment})
			}
		}
	}
	return byes
}

func present(players []chesspairing.PlayerEntry, byes []chesspairing.ByeEntry, round int) []string {
	absent := make(map[string]bool, len(byes))
	for _, bye := range byes {
		absent[bye.PlayerID] = true
	}
	var ids []string
	for _, player := range players {
		if (player.JoinedRound == 0 || player.JoinedRound <= round) && (player.WithdrawnAfterRound == nil || round <= *player.WithdrawnAfterRound) && !absent[player.ID] {
			ids = append(ids, player.ID)
		}
	}
	return ids
}

func scriptedResult(random *rand.Rand) chesspairing.GameResult {
	switch random.IntN(3) {
	case 0:
		return chesspairing.ResultWhiteWins
	case 1:
		return chesspairing.ResultDraw
	default:
		return chesspairing.ResultBlackWins
	}
}

func resultName(result chesspairing.GameResult) (string, error) {
	switch result {
	case chesspairing.ResultWhiteWins:
		return "1-0", nil
	case chesspairing.ResultDraw:
		return "½-½", nil
	case chesspairing.ResultBlackWins:
		return "0-1", nil
	case chesspairing.ResultDoubleForfeit:
		return "0-0f", nil
	default:
		return "", fmt.Errorf("unknown game result %q", result)
	}
}

func expectedRound(number int, scores []chesspairing.PlayerScore) ExpectedRound {
	standing := make([]ExpectedPlayer, len(scores))
	for i, score := range scores {
		standing[i] = ExpectedPlayer{ID: score.PlayerID, Points: score.Score, Position: score.Rank, ValueNumber: 61 - score.Rank}
	}
	return ExpectedRound{Number: number, Standings: standing}
}

// Update writes the generated fixtures. It is intentionally called only by a
// test guarded by ESGSEASON_UPDATE=1.
func Update(directory string) error {
	season, expected, err := Generate()
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(directory, "season.json"), season); err != nil {
		return err
	}
	return writeJSON(filepath.Join(directory, "expected.json"), expected)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
