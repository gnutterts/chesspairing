// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

//go:build ignore

// generate-javafo generates golden test files using JaVaFo as the reference
// Swiss pairing chesspairing. It produces the same scenario.json + round-N.json
// structure as the bbpPairings generator, but stored in testdata/golden-javafo/.
//
// Prerequisites:
//   - Java: /usr/local/opt/openjdk/bin/java (OpenJDK 25+)
//   - JaVaFo: /tmp/javafo/javafo.jar (v2.2)
//
// Usage:
//
//	cd services/api
//	go run ./internal/engine/pairing/dutch/testdata/cmd/generate-javafo
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

const (
	javaPath   = "/usr/local/opt/openjdk/bin/java"
	javafoPath = "/tmp/javafo/javafo.jar"
)

// GoldenScenario describes a test scenario written to scenario.json.
type GoldenScenario struct {
	Description        string                     `json:"description"`
	Players            []chesspairing.PlayerEntry `json:"players"`
	TotalRounds        int                        `json:"totalRounds"`
	ResultStrategy     string                     `json:"resultStrategy"`
	WithdrawAfterRound map[string]int             `json:"withdrawAfterRound,omitempty"`
}

// scenarioDef is the internal definition used to drive generation.
type scenarioDef struct {
	dirName            string
	description        string
	players            []chesspairing.PlayerEntry
	totalRounds        int
	resultStrategy     string
	withdrawAfterRound map[string]int
}

func main() {
	// Verify prerequisites.
	if _, err := os.Stat(javaPath); err != nil {
		log.Fatalf("Java not found at %s: %v", javaPath, err)
	}
	if _, err := os.Stat(javafoPath); err != nil {
		log.Fatalf("JaVaFo JAR not found at %s: %v", javafoPath, err)
	}

	scenarios := buildScenarios()

	// Resolve goldenDir relative to this source file's location.
	_, thisFile, _, _ := runtime.Caller(0)
	goldenDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "golden-javafo")

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
			dirName:        "8-players-5-rounds",
			description:    "8 players, 5 rounds, higher-rated always wins",
			players:        makePlayers([]int{2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700}),
			totalRounds:    5,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "10-players-7-rounds",
			description:    "10 players, 7 rounds, higher-rated always wins",
			players:        makePlayers([]int{2500, 2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700, 1600}),
			totalRounds:    7,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "9-players-5-rounds",
			description:    "9 players (odd count), 5 rounds, bye rotation, higher-rated always wins",
			players:        makePlayers([]int{2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700, 1600}),
			totalRounds:    5,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "12-players-7-rounds",
			description:    "12 players, 7 rounds, higher-rated always wins",
			players:        makePlayers([]int{2600, 2500, 2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700, 1600, 1500}),
			totalRounds:    7,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "20-players-9-rounds",
			description:    "20 players, 9 rounds, higher-rated always wins",
			players:        make20Players(),
			totalRounds:    9,
			resultStrategy: "higher-rated-wins",
		},
		{
			dirName:        "equal-ratings",
			description:    "6 players all rated 2000, 3 rounds, lower player ID wins ties",
			players:        makeEqualPlayers(6, 2000),
			totalRounds:    3,
			resultStrategy: "lower-id-wins",
		},
		{
			dirName:        "withdrawal",
			description:    "8 players, 5 rounds, p8 withdraws after round 2, higher-rated always wins",
			players:        makePlayers([]int{2400, 2300, 2200, 2100, 2000, 1900, 1800, 1700}),
			totalRounds:    5,
			resultStrategy: "higher-rated-wins",
			withdrawAfterRound: map[string]int{
				"p8": 2,
			},
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
			Active:      true,
		}
	}
	return players
}

func make20Players() []chesspairing.PlayerEntry {
	ratings := make([]int, 20)
	for i := 0; i < 20; i++ {
		ratings[i] = 2700 - i*47
	}
	return makePlayers(ratings)
}

func makeEqualPlayers(n, rating int) []chesspairing.PlayerEntry {
	players := make([]chesspairing.PlayerEntry, n)
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("p%d", i+1)
		players[i] = chesspairing.PlayerEntry{
			ID:          id,
			DisplayName: fmt.Sprintf("Player %s", id),
			Rating:      rating,
			Active:      true,
		}
	}
	return players
}

func generateScenario(goldenDir string, sc scenarioDef) error {
	dir := filepath.Join(goldenDir, sc.dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	// Write scenario.json.
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

	// Build a rating lookup from the original player list.
	ratingByID := make(map[string]int, len(sc.players))
	for _, p := range sc.players {
		ratingByID[p.ID] = p.Rating
	}

	// Sort players by rating descending.
	players := make([]chesspairing.PlayerEntry, len(sc.players))
	copy(players, sc.players)
	sort.Slice(players, func(i, j int) bool {
		return players[i].Rating > players[j].Rating
	})

	// Player index: ID -> 1-based start number.
	playerIndex := make(map[string]int, len(players))
	for i, p := range players {
		playerIndex[p.ID] = i + 1
	}

	// Track tournament state.
	var rounds []chesspairing.RoundData
	withdrawn := make(map[string]bool)

	for round := 1; round <= sc.totalRounds; round++ {
		// Apply withdrawals.
		if sc.withdrawAfterRound != nil && round > 1 {
			for pid, afterRound := range sc.withdrawAfterRound {
				if afterRound == round-1 {
					withdrawn[pid] = true
					fmt.Printf("    Round %d: player %s withdrawn\n", round, pid)
				}
			}
		}

		// Generate TRF for this round.
		trf := buildTRF(players, rounds, playerIndex, sc.totalRounds, withdrawn)

		// Write TRF to temp file.
		trfFile := filepath.Join(dir, fmt.Sprintf("round-%d.trf", round))
		if err := os.WriteFile(trfFile, []byte(trf), 0o644); err != nil {
			return fmt.Errorf("write TRF round %d: %w", round, err)
		}

		// Run JaVaFo.
		outputFile := filepath.Join(dir, fmt.Sprintf("round-%d-javafo.txt", round))
		cmd := exec.Command(javaPath, "-ea", "-jar", javafoPath, trfFile, "-p", outputFile)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("JaVaFo round %d failed: %v\nOutput: %s", round, err, string(out))
		}

		// Parse JaVaFo output.
		outputData, err := os.ReadFile(outputFile) //nolint:gosec // generator tool
		if err != nil {
			return fmt.Errorf("read JaVaFo output round %d: %w", round, err)
		}

		result, err := parseJaVaFoOutput(string(outputData), playerIndex)
		if err != nil {
			return fmt.Errorf("parse JaVaFo output round %d: %w", round, err)
		}

		result.Notes = []string{"Pairings generated by JaVaFo 2.2 (FIDE Dutch Swiss system)"}

		fmt.Printf("    Round %d: %d pairings, %d byes\n", round, len(result.Pairings), len(result.Byes))

		// Write round JSON.
		roundFile := filepath.Join(dir, fmt.Sprintf("round-%d.json", round))
		if err := writeJSON(roundFile, result); err != nil {
			return err
		}

		// Build round data with deterministic results for next round's TRF.
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

		// Clean up TRF and JaVaFo output files (keep only JSON).
		os.Remove(trfFile)
		os.Remove(outputFile)
	}

	return nil
}

// buildTRF generates a TRF string for the current tournament state.
func buildTRF(players []chesspairing.PlayerEntry, rounds []chesspairing.RoundData,
	playerIndex map[string]int, totalRounds int, withdrawn map[string]bool) string {

	var b strings.Builder

	b.WriteString("012 Tournament\n")
	fmt.Fprintf(&b, "062 %d\n", len(players))
	b.WriteString("092 Swiss Dutch\n")
	fmt.Fprintf(&b, "XXR %d\n", totalRounds)
	b.WriteString("XXC white1\n")

	for _, p := range players {
		sn := playerIndex[p.ID]
		isWithdrawn := withdrawn[p.ID]
		writeTRFPlayerLine(&b, sn, p, rounds, playerIndex, isWithdrawn)
	}

	return b.String()
}

// writeTRFPlayerLine writes a single 001 line for a player.
func writeTRFPlayerLine(b *strings.Builder, startNum int, player chesspairing.PlayerEntry,
	rounds []chesspairing.RoundData, playerIndex map[string]int, isWithdrawn bool) {

	name := player.DisplayName
	if len(name) > 33 {
		name = name[:33]
	}

	// Calculate total points.
	totalPoints := 0.0
	for _, round := range rounds {
		pts, _ := playerRoundResult(player.ID, round, playerIndex)
		totalPoints += pts
	}

	// Build fixed-width 89-char header.
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

	// Round results.
	for _, round := range rounds {
		writeRoundResult(b, player.ID, round, playerIndex, isWithdrawn)
	}

	// Withdrawn marker for current round.
	if isWithdrawn {
		b.WriteString("  0000 - Z")
	}

	b.WriteString("\n")
}

func playerRoundResult(playerID string, round chesspairing.RoundData,
	playerIndex map[string]int) (float64, int) {

	for _, bye := range round.Byes {
		if bye.PlayerID == playerID {
			return 1.0, 0
		}
	}

	for _, game := range round.Games {
		if game.WhiteID == playerID {
			oppSN := playerIndex[game.BlackID]
			switch game.Result {
			case chesspairing.ResultWhiteWins:
				return 1.0, oppSN
			case chesspairing.ResultBlackWins:
				return 0.0, oppSN
			case chesspairing.ResultDraw:
				return 0.5, oppSN
			default:
				return 0.0, oppSN
			}
		}
		if game.BlackID == playerID {
			oppSN := playerIndex[game.WhiteID]
			switch game.Result {
			case chesspairing.ResultWhiteWins:
				return 0.0, oppSN
			case chesspairing.ResultBlackWins:
				return 1.0, oppSN
			case chesspairing.ResultDraw:
				return 0.5, oppSN
			default:
				return 0.0, oppSN
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

// parseJaVaFoOutput parses JaVaFo's compact pairing output.
// Format:
//
//	N            (number of pairs)
//	white black  (start numbers, 0 = bye)
//	...
func parseJaVaFoOutput(output string, playerIndex map[string]int) (*chesspairing.PairingResult, error) {
	// Build reverse index.
	reverseIndex := make(map[int]string, len(playerIndex))
	for id, sn := range playerIndex {
		reverseIndex[sn] = id
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty JaVaFo output")
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

		// Bye: opponent = 0.
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
	case "lower-id-wins":
		if whiteID < blackID {
			return chesspairing.ResultWhiteWins
		}
		if blackID < whiteID {
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
