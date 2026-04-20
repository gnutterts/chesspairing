// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

//go:build ignore

// generate-bbppairings generates golden test files using bbpPairings as the
// reference Burstein Swiss pairing engine.
//
// Prerequisites:
//   - bbpPairings binary on PATH
//
// Usage:
//
//	go run ./pairing/burstein/testdata/cmd/generate-bbppairings/main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/gnutterts/chesspairing"
)

const bbpPairingsPath = "bbpPairings"

// GoldenScenario describes a test scenario written to scenario.json.
type GoldenScenario struct {
	Description        string                     `json:"description"`
	Players            []chesspairing.PlayerEntry `json:"players"`
	TotalRounds        int                        `json:"totalRounds"`
	ResultStrategy     string                     `json:"resultStrategy"`
	WithdrawAfterRound map[string]int             `json:"withdrawAfterRound,omitempty"`
}

type scenarioDef struct {
	dirName            string
	description        string
	players            []chesspairing.PlayerEntry
	totalRounds        int
	resultStrategy     string
	withdrawAfterRound map[string]int
}

func main() {
	if _, err := exec.LookPath(bbpPairingsPath); err != nil {
		log.Fatalf("bbpPairings not found on PATH: %v", err)
	}

	scenarios := buildScenarios()

	_, thisFile, _, _ := runtime.Caller(0)
	goldenDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "golden-bbppairings")

	for _, sc := range scenarios {
		fmt.Printf("=== Generating: %s ===\n", sc.dirName)
		if err := generateScenario(goldenDir, sc); err != nil {
			log.Fatalf("scenario %s failed: %v", sc.dirName, err)
		}
		fmt.Printf("    Done (%d rounds)\n", sc.totalRounds)
	}

	fmt.Println("\nAll scenarios generated successfully.")
}

func buildScenarios() []scenarioDef {
	return []scenarioDef{
		{
			dirName:        "6-players-3-rounds",
			description:    "6 players, 3 rounds, seeding rounds only, higher-rated wins",
			players:        makePlayers([]int{2200, 2100, 2000, 1900, 1800, 1700}),
			totalRounds:    3,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "8-players-5-rounds",
			description:    "8 players, 5 rounds, higher-rated wins",
			players:        makePlayers([]int{2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700}),
			totalRounds:    5,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "10-players-5-rounds",
			description:    "10 players, 5 rounds, 2 seeding + 3 post-seeding, higher-rated wins",
			players:        makePlayers([]int{2500, 2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700, 1600}),
			totalRounds:    5,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "12-players-7-rounds",
			description:    "12 players, 7 rounds, 3 seeding + 4 post-seeding, higher-rated wins",
			players:        makePlayers([]int{2600, 2500, 2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700, 1600, 1500}),
			totalRounds:    7,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "20-players-9-rounds",
			description:    "20 players, 9 rounds, 4 seeding + 5 post-seeding, higher-rated wins",
			players:        make20Players(),
			totalRounds:    9,
			resultStrategy: "higher-rated-wins",
		},
	}
}

func makePlayers(ratings []int) []chesspairing.PlayerEntry {
	players := make([]chesspairing.PlayerEntry, len(ratings))
	for i, r := range ratings {
		id := fmt.Sprintf("p%d", i+1)
		players[i] = chesspairing.PlayerEntry{
			ID:          id,
			DisplayName: fmt.Sprintf("Player %d", r),
			Rating:      r,
		}
	}
	return players
}

func make20Players() []chesspairing.PlayerEntry {
	ratings := make([]int, 20)
	for i := range 20 {
		ratings[i] = 2700 - i*47
	}
	return makePlayers(ratings)
}

func generateScenario(goldenDir string, sc scenarioDef) error {
	dir := filepath.Join(goldenDir, sc.dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	scenarioFile := GoldenScenario{
		Description:        sc.description,
		Players:            sc.players,
		TotalRounds:        sc.totalRounds,
		ResultStrategy:     sc.resultStrategy,
		WithdrawAfterRound: sc.withdrawAfterRound,
	}
	if err := writeJSON(filepath.Join(dir, "scenario.json"), scenarioFile); err != nil {
		return err
	}

	ratingByID := make(map[string]int, len(sc.players))
	for _, p := range sc.players {
		ratingByID[p.ID] = p.Rating
	}

	players := make([]chesspairing.PlayerEntry, len(sc.players))
	copy(players, sc.players)
	sort.Slice(players, func(i, j int) bool {
		return players[i].Rating > players[j].Rating
	})

	playerIndex := make(map[string]int, len(players))
	for i, p := range players {
		playerIndex[p.ID] = i + 1
	}

	var rounds []chesspairing.RoundData
	withdrawn := make(map[string]bool)

	for round := 1; round <= sc.totalRounds; round++ {
		if sc.withdrawAfterRound != nil && round > 1 {
			for pid, afterRound := range sc.withdrawAfterRound {
				if afterRound == round-1 {
					withdrawn[pid] = true
					fmt.Printf("    Round %d: player %s withdrawn\n", round, pid)
				}
			}
		}

		trfContent := buildTRF(players, rounds, playerIndex, sc.totalRounds, withdrawn)

		tmpDir, err := os.MkdirTemp("", "bbp-burstein-*")
		if err != nil {
			return fmt.Errorf("create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		trfFile := filepath.Join(tmpDir, "input.trf")
		if err := os.WriteFile(trfFile, []byte(trfContent), 0o644); err != nil {
			return fmt.Errorf("write TRF round %d: %w", round, err)
		}

		outputFile := filepath.Join(tmpDir, "output.txt")
		cmd := exec.Command(bbpPairingsPath, "--burstein", trfFile, "-p", outputFile)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("bbpPairings --burstein round %d failed: %v\nOutput: %s\nTRF:\n%s", round, err, string(out), trfContent)
		}

		outputData, err := os.ReadFile(outputFile) //nolint:gosec // generator tool
		if err != nil {
			return fmt.Errorf("read bbpPairings output round %d: %w", round, err)
		}

		result, err := parseBBPOutput(string(outputData), playerIndex)
		if err != nil {
			return fmt.Errorf("parse bbpPairings output round %d: %w", round, err)
		}

		result.Notes = []string{"Pairings generated by bbpPairings --burstein (Burstein Swiss system)"}

		fmt.Printf("    Round %d: %d pairings, %d byes\n", round, len(result.Pairings), len(result.Byes))

		roundFile := filepath.Join(dir, fmt.Sprintf("round-%d.json", round))
		if err := writeJSON(roundFile, result); err != nil {
			return err
		}

		rd := chesspairing.RoundData{
			Number: round,
			Games:  make([]chesspairing.GameData, len(result.Pairings)),
			Byes:   result.Byes,
		}
		for i, p := range result.Pairings {
			res := determineResult(p.WhiteID, p.BlackID, ratingByID, sc.resultStrategy)
			rd.Games[i] = chesspairing.GameData{
				WhiteID:   p.WhiteID,
				BlackID:   p.BlackID,
				Result:    res,
				IsForfeit: false,
			}
		}
		rounds = append(rounds, rd)
	}

	return nil
}

func buildTRF(players []chesspairing.PlayerEntry, rounds []chesspairing.RoundData,
	playerIndex map[string]int, totalRounds int, withdrawn map[string]bool) string {

	var b strings.Builder
	b.WriteString("012 Tournament\n")
	fmt.Fprintf(&b, "062 %d\n", len(players))
	b.WriteString("092 Swiss Burstein\n")
	fmt.Fprintf(&b, "XXR %d\n", totalRounds)
	b.WriteString("XXC white1\n")

	for _, p := range players {
		sn := playerIndex[p.ID]
		isWithdrawn := withdrawn[p.ID]
		writeTRFPlayerLine(&b, sn, p, rounds, playerIndex, isWithdrawn)
	}

	return b.String()
}

func writeTRFPlayerLine(b *strings.Builder, startNum int, player chesspairing.PlayerEntry,
	rounds []chesspairing.RoundData, playerIndex map[string]int, isWithdrawn bool) {

	name := player.DisplayName
	if len(name) > 33 {
		name = name[:33]
	}

	totalPoints := 0.0
	for _, round := range rounds {
		pts, _ := playerRoundResult(player.ID, round)
		totalPoints += pts
	}

	header := make([]byte, 89)
	for i := range header {
		header[i] = ' '
	}

	copy(header[0:3], "001")
	copy(header[4:8], fmt.Sprintf("%4d", startNum))
	copy(header[14:14+len(name)], name)
	copy(header[48:52], fmt.Sprintf("%4d", player.Rating))
	copy(header[53:56], "NED")
	copy(header[80:84], formatPoints(totalPoints))

	b.Write(header)

	for _, round := range rounds {
		writeRoundResult(b, player.ID, round, playerIndex, isWithdrawn)
	}

	if isWithdrawn {
		b.WriteString("  0000 - Z")
	}

	b.WriteString("\n")
}

func playerRoundResult(playerID string, round chesspairing.RoundData) (float64, int) {
	for _, bye := range round.Byes {
		if bye.PlayerID == playerID {
			return 1.0, 0
		}
	}

	for _, game := range round.Games {
		if game.WhiteID == playerID {
			switch game.Result {
			case chesspairing.ResultWhiteWins:
				return 1.0, 0
			case chesspairing.ResultBlackWins:
				return 0.0, 0
			case chesspairing.ResultDraw:
				return 0.5, 0
			default:
				return 0.0, 0
			}
		}
		if game.BlackID == playerID {
			switch game.Result {
			case chesspairing.ResultWhiteWins:
				return 0.0, 0
			case chesspairing.ResultBlackWins:
				return 1.0, 0
			case chesspairing.ResultDraw:
				return 0.5, 0
			default:
				return 0.0, 0
			}
		}
	}

	return 0.0, 0
}

func writeRoundResult(b *strings.Builder, playerID string, round chesspairing.RoundData,
	playerIndex map[string]int, isWithdrawn bool) {

	for _, bye := range round.Byes {
		if bye.PlayerID == playerID {
			b.WriteString("  0000 - F")
			return
		}
	}

	for _, game := range round.Games {
		if game.WhiteID == playerID {
			oppSN := playerIndex[game.BlackID]
			result := gameResultChar(game.Result, true)
			fmt.Fprintf(b, "  %4d w %c", oppSN, result)
			return
		}
		if game.BlackID == playerID {
			oppSN := playerIndex[game.WhiteID]
			result := gameResultChar(game.Result, false)
			fmt.Fprintf(b, "  %4d b %c", oppSN, result)
			return
		}
	}

	if isWithdrawn {
		b.WriteString("  0000 - Z")
	} else {
		b.WriteString("  0000 - U")
	}
}

func gameResultChar(result chesspairing.GameResult, isWhite bool) byte {
	switch result {
	case chesspairing.ResultWhiteWins:
		if isWhite {
			return '1'
		}
		return '0'
	case chesspairing.ResultBlackWins:
		if isWhite {
			return '0'
		}
		return '1'
	case chesspairing.ResultDraw:
		return '='
	default:
		return '*'
	}
}

func formatPoints(pts float64) string {
	s := fmt.Sprintf("%.1f", pts)
	for len(s) < 4 {
		s = " " + s
	}
	return s
}

func parseBBPOutput(output string, playerIndex map[string]int) (*chesspairing.PairingResult, error) {
	reverseIndex := make(map[int]string, len(playerIndex))
	for id, sn := range playerIndex {
		reverseIndex[sn] = id
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty bbpPairings output")
	}

	numPairs, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return nil, fmt.Errorf("parse pair count: %w", err)
	}

	if len(lines) < numPairs+1 {
		return nil, fmt.Errorf("expected %d pair lines, got %d", numPairs, len(lines)-1)
	}

	result := &chesspairing.PairingResult{}
	board := 0

	for i := 1; i <= numPairs; i++ {
		fields := strings.Fields(strings.TrimSpace(lines[i]))
		if len(fields) != 2 {
			return nil, fmt.Errorf("pair line %d: expected 2 fields, got %d", i, len(fields))
		}

		whiteSN, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, fmt.Errorf("pair line %d: parse white SN: %w", i, err)
		}
		blackSN, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("pair line %d: parse black SN: %w", i, err)
		}

		if blackSN == 0 {
			whiteID, ok := reverseIndex[whiteSN]
			if !ok {
				return nil, fmt.Errorf("pair line %d: unknown white SN %d", i, whiteSN)
			}
			result.Byes = append(result.Byes, chesspairing.ByeEntry{PlayerID: whiteID, Type: chesspairing.ByePAB})
			continue
		}
		if whiteSN == 0 {
			blackID, ok := reverseIndex[blackSN]
			if !ok {
				return nil, fmt.Errorf("pair line %d: unknown black SN %d", i, blackSN)
			}
			result.Byes = append(result.Byes, chesspairing.ByeEntry{PlayerID: blackID, Type: chesspairing.ByePAB})
			continue
		}

		whiteID, ok := reverseIndex[whiteSN]
		if !ok {
			return nil, fmt.Errorf("pair line %d: unknown white SN %d", i, whiteSN)
		}
		blackID, ok := reverseIndex[blackSN]
		if !ok {
			return nil, fmt.Errorf("pair line %d: unknown black SN %d", i, blackSN)
		}

		board++
		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   board,
			WhiteID: whiteID,
			BlackID: blackID,
		})
	}

	return result, nil
}

func determineResult(whiteID, blackID string, ratings map[string]int, strategy string) chesspairing.GameResult {
	switch strategy {
	case "higher-rated-wins":
		wr := ratings[whiteID]
		br := ratings[blackID]
		if wr > br {
			return chesspairing.ResultWhiteWins
		}
		if br > wr {
			return chesspairing.ResultBlackWins
		}
		return chesspairing.ResultDraw
	default:
		log.Fatalf("unknown result strategy: %s", strategy)
		return ""
	}
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
