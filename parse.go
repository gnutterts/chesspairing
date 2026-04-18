// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package chesspairing

import (
	"fmt"
	"strings"
)

// ParseScoringSystem parses a string into a ScoringSystem. Matching is
// case-insensitive and surrounding whitespace is trimmed. Accepted values
// are the canonical names "standard", "keizer", "football".
//
// Returns an error wrapping the input on unknown values. Empty input is
// rejected so the empty string and typos surface at the parse boundary
// instead of propagating as a default.
func ParseScoringSystem(s string) (ScoringSystem, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "standard":
		return ScoringStandard, nil
	case "keizer":
		return ScoringKeizer, nil
	case "football":
		return ScoringFootball, nil
	case "":
		return "", fmt.Errorf("empty scoring system")
	default:
		return "", fmt.Errorf("unknown scoring system: %q", s)
	}
}

// ParsePairingSystem parses a string into a PairingSystem. Matching is
// case-insensitive and surrounding whitespace is trimmed. Accepted values
// are the canonical names ("dutch", "burstein", "dubov", "lim",
// "doubleswiss", "team", "keizer", "roundrobin") plus the bbpPairings /
// JaVaFo aliases ("fide-dutch", "fide-burstein", "fide-dubov", "fide-lim",
// "double-swiss", "round-robin", "rr").
//
// Returns an error wrapping the input on unknown values. Empty input is
// rejected.
func ParsePairingSystem(s string) (PairingSystem, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "dutch", "fide-dutch":
		return PairingDutch, nil
	case "burstein", "fide-burstein":
		return PairingBurstein, nil
	case "dubov", "fide-dubov":
		return PairingDubov, nil
	case "lim", "fide-lim":
		return PairingLim, nil
	case "doubleswiss", "double-swiss":
		return PairingDoubleSwiss, nil
	case "team":
		return PairingTeam, nil
	case "keizer":
		return PairingKeizer, nil
	case "roundrobin", "round-robin", "rr":
		return PairingRoundRobin, nil
	case "":
		return "", fmt.Errorf("empty pairing system")
	default:
		return "", fmt.Errorf("unknown pairing system: %q", s)
	}
}

// ParseGameResult parses a string into a GameResult. Matching is
// case-insensitive and surrounding whitespace is trimmed. Internal spaces
// around the dash are tolerated ("1 - 0" parses as ResultWhiteWins).
//
// Accepted spellings:
//
//	"1-0", "1 - 0"           -> ResultWhiteWins
//	"0-1", "0 - 1"           -> ResultBlackWins
//	"0.5-0.5", "1/2-1/2",    -> ResultDraw
//	"0.5 - 0.5", "1/2 - 1/2"
//	"*"                      -> ResultPending
//	"1-0f", "1 - 0 f"        -> ResultForfeitWhiteWins
//	"0-1f", "0 - 1 f"        -> ResultForfeitBlackWins
//	"0-0f", "0 - 0 f"        -> ResultDoubleForfeit
//
// Returns an error wrapping the input on unknown values. Empty input is
// rejected.
func ParseGameResult(s string) (GameResult, error) {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" {
		return "", fmt.Errorf("empty game result")
	}
	// Normalize internal whitespace.
	t = strings.Join(strings.Fields(t), "")
	switch t {
	case "1-0":
		return ResultWhiteWins, nil
	case "0-1":
		return ResultBlackWins, nil
	case "0.5-0.5", "1/2-1/2", "½-½":
		return ResultDraw, nil
	case "*":
		return ResultPending, nil
	case "1-0f":
		return ResultForfeitWhiteWins, nil
	case "0-1f":
		return ResultForfeitBlackWins, nil
	case "0-0f":
		return ResultDoubleForfeit, nil
	default:
		return "", fmt.Errorf("unknown game result: %q", s)
	}
}

// ParseByeType parses a string into a ByeType. Matching is case-insensitive
// and surrounding whitespace is trimmed. Accepted spellings include both the
// String() forms and the TRF letter codes:
//
//	"PAB", "F"            -> ByePAB
//	"Half", "H"           -> ByeHalf
//	"Zero", "Z"           -> ByeZero
//	"Absent", "U"         -> ByeAbsent
//	"Excused"             -> ByeExcused
//	"ClubCommitment"      -> ByeClubCommitment
//
// The TRF letters are accepted because they are the canonical TRF-side
// spelling and downstream consumers reading TRF data may have them in hand.
// String() output is unchanged.
//
// Returns an error wrapping the input on unknown values. Empty input is
// rejected.
func ParseByeType(s string) (ByeType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "pab", "f":
		return ByePAB, nil
	case "half", "h":
		return ByeHalf, nil
	case "zero", "z":
		return ByeZero, nil
	case "absent", "u":
		return ByeAbsent, nil
	case "excused":
		return ByeExcused, nil
	case "clubcommitment":
		return ByeClubCommitment, nil
	case "":
		return 0, fmt.Errorf("empty bye type")
	default:
		return 0, fmt.Errorf("unknown bye type: %q", s)
	}
}
