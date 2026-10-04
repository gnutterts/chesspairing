// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"errors"
	"math"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// BakuAccelerationRounds returns the number of accelerated rounds, full virtual
// point rounds, and half virtual point rounds for the Baku acceleration system
// (FIDE C.04.7).
//
//   - accelerated = ceil(totalRounds / 2)
//   - fullVP = ceil(accelerated / 2)
//   - halfVP = accelerated - fullVP
func BakuAccelerationRounds(totalRounds int) (accelerated, fullVP, halfVP int) {
	accelerated = int(math.Ceil(float64(totalRounds) / 2.0))
	fullVP = int(math.Ceil(float64(accelerated) / 2.0))
	halfVP = accelerated - fullVP
	return
}

// BakuGASize returns the size of Group A for Baku acceleration: 2 *
// ceil(participantsBeforeFirstRound / 4).
func BakuGASize(participantsBeforeFirstRound int) int {
	return 2 * int(math.Ceil(float64(participantsBeforeFirstRound)/4.0))
}

// BakuGroupA returns the participants in group A of the Baku acceleration
// (C.04.7 1.2 and 1.3).
func BakuGroupA(players []chesspairing.PlayerEntry) (map[string]bool, error) {
	numbered, err := chesspairing.AssignPairingNumbers(players)
	if err != nil {
		return nil, err
	}
	initial := make([]chesspairing.PlayerEntry, 0, len(numbered))
	for _, player := range numbered {
		if player.JoinedRound <= 1 {
			initial = append(initial, player)
		}
	}
	if len(initial) == 0 {
		return map[string]bool{}, nil
	}
	sort.Slice(initial, func(i, j int) bool {
		return initial[i].PairingNumber < initial[j].PairingNumber
	})
	gaSize := min(BakuGASize(len(initial)), len(initial))
	last := initial[gaSize-1].PairingNumber
	groupA := make(map[string]bool, gaSize)
	for _, player := range numbered {
		if player.PairingNumber <= last {
			groupA[player.ID] = true
		}
	}
	return groupA, nil
}

// BakuPremise reports whether the scoring configuration satisfies C.04.7 1.1.
func BakuPremise(cfg chesspairing.ScoringConfig) error {
	if cfg.System == "" || cfg.System == chesspairing.ScoringStandard {
		opts := standard.ParseOptions(cfg.Options).WithDefaults()
		if math.Abs(*opts.PointWin-2**opts.PointDraw) < 1e-9 && math.Abs(*opts.PointLoss) < 1e-9 {
			return nil
		}
	}
	return errors.New("scoring does not meet C.04.7 1.1 for Baku acceleration: a win must score two draws and a loss zero")
}

// BakuVirtualPoints returns the virtual points for a player in a given round
// under Baku acceleration.
//
// winPoints is the number of points awarded for a win by the tournament's
// scoring configuration.
//
//   - GA player in a full VP round: winPoints
//   - GA player in a half VP round: winPoints / 2
//   - All other cases: 0.0
func BakuVirtualPoints(winPoints float64, totalRounds, currentRound int, isGA bool) float64 {
	if !isGA {
		return 0.0
	}

	accelerated, fullVP, _ := BakuAccelerationRounds(totalRounds)

	if currentRound <= fullVP {
		return winPoints
	}

	if currentRound <= accelerated {
		return winPoints / 2
	}

	return 0.0
}

// ApplyBakuAcceleration modifies PairingScore for each player by adding virtual
// points based on the Baku acceleration system.
//
// winPoints is the number of points awarded for a win by the tournament's
// scoring configuration.
//
// Players in groupA receive virtual points. Players outside Group A are not
// modified.
func ApplyBakuAcceleration(winPoints float64, players []PlayerState, currentRound, totalRounds int, groupA map[string]bool) {
	for i := range players {
		vp := BakuVirtualPoints(winPoints, totalRounds, currentRound, groupA[players[i].ID])
		players[i].PairingScore = players[i].Score + vp
	}
}

// AddVirtualPoints gives every player the virtual points of the round being
// paired and orders the players by the resulting pairing score (then by
// pairing number), renumbering their live rank TPN. From here on Score and
// PairingScore hold the pairing score of C.04.7 1.5, which is what the
// scoregroups, the criteria and the bye choice are based on; the callers keep
// the real scores if they need them.
func AddVirtualPoints(players []PlayerState, virtualPoints func(playerID string) float64) {
	for i := range players {
		vp := virtualPoints(players[i].ID)
		players[i].Score += vp
		players[i].PairingScore = players[i].Score
	}
	sort.SliceStable(players, func(i, j int) bool {
		if players[i].Score != players[j].Score {
			return players[i].Score > players[j].Score
		}
		return players[i].InitialRank < players[j].InitialRank
	})
	for i := range players {
		players[i].TPN = i + 1
	}
}
