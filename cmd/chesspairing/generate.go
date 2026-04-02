// cmd/chesspairing/generate.go
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	mrand "math/rand/v2"
	"os"
	"sort"
	"strconv"
	"strings"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/trf"
)

// rtgConfig holds RTG configuration (bbpPairings-compatible keys).
type rtgConfig struct {
	PlayersNumber        int
	RoundsNumber         int
	DrawPercentage       int
	ForfeitRate          int
	RetiredRate          int
	HalfPointByeRate     int
	HighestRating        int
	LowestRating         int
	PointsForWin         float64
	PointsForDraw        float64
	PointsForLoss        float64
	PointsForZPB         float64
	PointsForForfeitLoss float64
	PointsForPAB         float64
}

func defaultRTGConfig() rtgConfig {
	return rtgConfig{
		PlayersNumber:        30,
		RoundsNumber:         9,
		DrawPercentage:       30,
		ForfeitRate:          20,
		RetiredRate:          100,
		HalfPointByeRate:     100,
		HighestRating:        2600,
		LowestRating:         1400,
		PointsForWin:         1.0,
		PointsForDraw:        0.5,
		PointsForLoss:        0.0,
		PointsForZPB:         0.5,
		PointsForForfeitLoss: 0.0,
		PointsForPAB:         0.0,
	}
}

func runGenerate(args []string, stdout, stderr io.Writer) int {
	// Parse args manually (legacy-style positional args)
	parsed, err := parseLegacyArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitInvalidInput
	}

	if parsed.system == "" {
		fmt.Fprintln(stderr, "error: system flag required for generation (e.g. --dutch)")
		return ExitInvalidInput
	}

	if parsed.outputFile == "" {
		fmt.Fprintln(stderr, "error: -o output file required for generation")
		return ExitInvalidInput
	}

	// Load config
	cfg := defaultRTGConfig()
	if parsed.configFile != "" {
		if err := loadRTGConfig(parsed.configFile, &cfg); err != nil {
			fmt.Fprintf(stderr, "error: loading config: %v\n", err)
			return ExitFileAccess
		}
	}

	// Set up PRNG
	var rng *mrand.Rand
	if parsed.seed != "" {
		seed := parseSeed(parsed.seed)
		rng = mrand.New(mrand.NewPCG(uint64(seed), 0))
	} else {
		// Random seed from crypto/rand
		var seedBytes [8]byte
		if _, err := rand.Read(seedBytes[:]); err != nil {
			fmt.Fprintf(stderr, "error: generating random seed: %v\n", err)
			return ExitUnexpected
		}
		seed := int64(binary.LittleEndian.Uint64(seedBytes[:]))
		fmt.Fprintf(stderr, "seed: %d\n", seed)
		rng = mrand.New(mrand.NewPCG(uint64(seed), 0))
	}

	// Generate players
	doc := generatePlayers(rng, cfg)

	// Set tournament metadata
	doc.Name = "Generated Tournament"
	doc.TotalRounds = cfg.RoundsNumber

	// Generate rounds
	ctx := context.Background()
	for round := 1; round <= cfg.RoundsNumber; round++ {
		state, err := doc.ToTournamentState()
		if err != nil {
			fmt.Fprintf(stderr, "error: round %d state: %v\n", round, err)
			return ExitUnexpected
		}

		state.PairingConfig.System = parsed.system
		pairer, err := newPairer(parsed.system, state.PairingConfig.Options)
		if err != nil {
			fmt.Fprintf(stderr, "error: round %d pairer: %v\n", round, err)
			return ExitUnexpected
		}

		result, err := pairer.Pair(ctx, state)
		if err != nil {
			fmt.Fprintf(stderr, "error: round %d pairing failed: %v\n", round, err)
			return ExitNoPairing
		}

		// Generate results for each pairing
		applyRandomResults(rng, cfg, result, state)

		// Append round to document
		appendRoundToDoc(doc, result, round)
	}

	// Write output
	out, err := os.Create(parsed.outputFile)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot create %s: %v\n", parsed.outputFile, err)
		return ExitFileAccess
	}
	defer out.Close()

	if err := trf.Write(out, doc); err != nil {
		fmt.Fprintf(stderr, "error: writing TRF: %v\n", err)
		return ExitUnexpected
	}

	return ExitSuccess
}

// parseSeed converts a seed string to int64.
// Integer strings are parsed directly; non-integers are hashed via FNV-1a.
func parseSeed(s string) int64 {
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

func loadRTGConfig(filename string, cfg *rtgConfig) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	lineNum := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "PlayersNumber":
			v, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.PlayersNumber = v
		case "RoundsNumber":
			v, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.RoundsNumber = v
		case "DrawPercentage":
			v, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.DrawPercentage = v
		case "ForfeitRate":
			v, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.ForfeitRate = v
		case "RetiredRate":
			v, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.RetiredRate = v
		case "HalfPointByeRate":
			v, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.HalfPointByeRate = v
		case "HighestRating":
			v, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.HighestRating = v
		case "LowestRating":
			v, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.LowestRating = v
		case "PointsForWin":
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.PointsForWin = v
		case "PointsForDraw":
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.PointsForDraw = v
		case "PointsForLoss":
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.PointsForLoss = v
		case "PointsForZPB":
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.PointsForZPB = v
		case "PointsForForfeitLoss":
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.PointsForForfeitLoss = v
		case "PointsForPAB":
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("line %d: %s: %w", lineNum, key, err)
			}
			cfg.PointsForPAB = v
		}
	}
	return scanner.Err()
}

func generatePlayers(rng *mrand.Rand, cfg rtgConfig) *trf.Document {
	doc := &trf.Document{}
	ratingRange := cfg.HighestRating - cfg.LowestRating

	// Generate random ratings
	ratings := make([]int, cfg.PlayersNumber)
	for i := range ratings {
		ratings[i] = cfg.LowestRating + rng.IntN(ratingRange+1)
	}
	// Sort descending (highest rated = start number 1)
	sort.Sort(sort.Reverse(sort.IntSlice(ratings)))

	for i, rating := range ratings {
		doc.Players = append(doc.Players, trf.PlayerLine{
			StartNumber: i + 1,
			Name:        fmt.Sprintf("Player %d", i+1),
			Rating:      rating,
			Rank:        i + 1,
		})
	}
	doc.NumPlayers = cfg.PlayersNumber

	return doc
}

func applyRandomResults(rng *mrand.Rand, cfg rtgConfig, result *cp.PairingResult, state *cp.TournamentState) {
	// Build rating lookup
	ratingMap := make(map[string]int, len(state.Players))
	for _, p := range state.Players {
		ratingMap[p.ID] = p.Rating
	}

	drawProb := float64(cfg.DrawPercentage) / 100.0

	for i := range result.Pairings {
		whiteRating := float64(ratingMap[result.Pairings[i].WhiteID])
		blackRating := float64(ratingMap[result.Pairings[i].BlackID])

		// Forfeit check
		if cfg.ForfeitRate > 0 {
			forfeitProb := math.Sqrt(1.0 - 1.0/float64(cfg.ForfeitRate))
			whiteForfeit := rng.Float64() > forfeitProb
			blackForfeit := rng.Float64() > forfeitProb
			if whiteForfeit && blackForfeit {
				result.Notes = append(result.Notes, fmt.Sprintf("forfeit:%d:double", i))
				continue
			}
			if whiteForfeit {
				result.Notes = append(result.Notes, fmt.Sprintf("forfeit:%d:black-wins", i))
				continue
			}
			if blackForfeit {
				result.Notes = append(result.Notes, fmt.Sprintf("forfeit:%d:white-wins", i))
				continue
			}
		}

		// Expected score for white using logistic model
		ratingDiff := whiteRating - blackRating
		expectedWhite := 1.0 / (1.0 + math.Pow(10, -ratingDiff/400.0))

		// Draw probability capped
		actualDrawProb := math.Min(drawProb, 2.0-expectedWhite*2.0)
		if actualDrawProb < 0 {
			actualDrawProb = 0
		}

		roll := rng.Float64()
		if roll < actualDrawProb {
			result.Notes = append(result.Notes, fmt.Sprintf("result:%d:draw", i))
		} else if roll < actualDrawProb+(1.0-actualDrawProb)*expectedWhite {
			result.Notes = append(result.Notes, fmt.Sprintf("result:%d:white-wins", i))
		} else {
			result.Notes = append(result.Notes, fmt.Sprintf("result:%d:black-wins", i))
		}
	}
}

func appendRoundToDoc(doc *trf.Document, result *cp.PairingResult, roundNum int) {
	// Parse notes to determine results
	resultMap := make(map[int]string) // pairing index → result type
	for _, note := range result.Notes {
		parts := strings.SplitN(note, ":", 3)
		if len(parts) != 3 {
			continue
		}
		idx, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		resultMap[idx] = parts[0] + ":" + parts[2]
	}

	// Build start number lookup from player IDs
	playerIdx := make(map[string]int) // player ID → index in doc.Players
	for i, pl := range doc.Players {
		playerIdx[fmt.Sprintf("%d", pl.StartNumber)] = i
	}

	// Add round results to each player
	for i, pairing := range result.Pairings {
		whiteIdx, wOK := playerIdx[pairing.WhiteID]
		blackIdx, bOK := playerIdx[pairing.BlackID]
		if !wOK || !bOK {
			continue
		}

		res := resultMap[i]
		var whiteResult, blackResult trf.ResultCode
		switch res {
		case "result:white-wins":
			whiteResult = trf.ResultWin
			blackResult = trf.ResultLoss
		case "result:black-wins":
			whiteResult = trf.ResultLoss
			blackResult = trf.ResultWin
		case "result:draw":
			whiteResult = trf.ResultDraw
			blackResult = trf.ResultDraw
		case "forfeit:white-wins":
			whiteResult = trf.ResultForfeitWin
			blackResult = trf.ResultForfeitLoss
		case "forfeit:black-wins":
			whiteResult = trf.ResultForfeitLoss
			blackResult = trf.ResultForfeitWin
		case "forfeit:double":
			whiteResult = trf.ResultForfeitLoss
			blackResult = trf.ResultForfeitLoss
		default:
			whiteResult = trf.ResultDraw
			blackResult = trf.ResultDraw
		}

		doc.Players[whiteIdx].Rounds = append(doc.Players[whiteIdx].Rounds, trf.RoundResult{
			Opponent: doc.Players[blackIdx].StartNumber,
			Color:    trf.ColorWhite,
			Result:   whiteResult,
		})
		doc.Players[blackIdx].Rounds = append(doc.Players[blackIdx].Rounds, trf.RoundResult{
			Opponent: doc.Players[whiteIdx].StartNumber,
			Color:    trf.ColorBlack,
			Result:   blackResult,
		})
	}

	// Handle byes
	for _, bye := range result.Byes {
		idx, ok := playerIdx[bye.PlayerID]
		if !ok {
			continue
		}
		var byeResult trf.ResultCode
		switch bye.Type {
		case cp.ByePAB:
			byeResult = trf.ResultFullBye
		case cp.ByeHalf:
			byeResult = trf.ResultHalfBye
		case cp.ByeZero:
			byeResult = trf.ResultZeroBye
		default:
			byeResult = trf.ResultUnpaired
		}
		doc.Players[idx].Rounds = append(doc.Players[idx].Rounds, trf.RoundResult{
			Opponent: 0,
			Color:    trf.ColorNone,
			Result:   byeResult,
		})
	}

	// Players who neither played nor got a bye are absent
	playedThisRound := make(map[string]bool)
	for _, p := range result.Pairings {
		playedThisRound[p.WhiteID] = true
		playedThisRound[p.BlackID] = true
	}
	for _, b := range result.Byes {
		playedThisRound[b.PlayerID] = true
	}
	for _, pl := range doc.Players {
		pid := fmt.Sprintf("%d", pl.StartNumber)
		if !playedThisRound[pid] {
			idx := playerIdx[pid]
			doc.Players[idx].Rounds = append(doc.Players[idx].Rounds, trf.RoundResult{
				Opponent: 0,
				Color:    trf.ColorNone,
				Result:   trf.ResultUnpaired,
			})
		}
	}

	// Clear notes after consuming them
	result.Notes = nil
}
