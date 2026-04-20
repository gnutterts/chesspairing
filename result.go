// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package chesspairing

import "fmt"

// GameResult represents the outcome of a chess game.
type GameResult string

const (
	ResultWhiteWins        GameResult = "1-0"
	ResultBlackWins        GameResult = "0-1"
	ResultDraw             GameResult = "0.5-0.5"
	ResultPending          GameResult = "*"
	ResultForfeitWhiteWins GameResult = "1-0f"
	ResultForfeitBlackWins GameResult = "0-1f"
	ResultDoubleForfeit    GameResult = "0-0f"
)

// IsValid returns true if the game result is a recognized value.
func (gr GameResult) IsValid() bool {
	switch gr {
	case ResultWhiteWins, ResultBlackWins, ResultDraw, ResultPending,
		ResultForfeitWhiteWins, ResultForfeitBlackWins, ResultDoubleForfeit:
		return true
	}
	return false
}

// IsRecordable returns true if the game result is a valid result that can be
// recorded by a user. ResultPending ("*") is valid but not recordable — it is
// the initial state set by the system when a game is created.
func (gr GameResult) IsRecordable() bool {
	switch gr {
	case ResultWhiteWins, ResultBlackWins, ResultDraw,
		ResultForfeitWhiteWins, ResultForfeitBlackWins, ResultDoubleForfeit:
		return true
	}
	return false
}

// IsForfeit returns true if the result is a forfeit (single or double).
// Forfeit games are excluded from pairing history — players can be
// re-paired within the same period.
func (gr GameResult) IsForfeit() bool {
	switch gr {
	case ResultForfeitWhiteWins, ResultForfeitBlackWins, ResultDoubleForfeit:
		return true
	}
	return false
}

// IsDoubleForfeit returns true if both players forfeited.
// Double-forfeit games are excluded from both pairing and scoring —
// the game never happened.
func (gr GameResult) IsDoubleForfeit() bool {
	return gr == ResultDoubleForfeit
}

// ByeType classifies how a bye is scored.
type ByeType int

const (
	ByePAB            ByeType = iota // Pairing-Allocated Bye (full point, TRF "F")
	ByeHalf                          // Half-point bye (TRF "H")
	ByeZero                          // Zero-point bye (TRF "Z")
	ByeAbsent                        // Absent/unpaired, unexcused (TRF "U")
	ByeExcused                       // Excused absence (notified in advance)
	ByeClubCommitment                // Club commitment (absent for interclub team duty)
)

// IsValid returns true if the bye type is a recognized value.
func (bt ByeType) IsValid() bool {
	return bt >= ByePAB && bt <= ByeClubCommitment
}

// String returns the human-readable name of the bye type.
func (bt ByeType) String() string {
	switch bt {
	case ByePAB:
		return "PAB"
	case ByeHalf:
		return "Half"
	case ByeZero:
		return "Zero"
	case ByeAbsent:
		return "Absent"
	case ByeExcused:
		return "Excused"
	case ByeClubCommitment:
		return "ClubCommitment"
	default:
		return "Unknown"
	}
}

// ByeEntry records a bye assignment with its type.
type ByeEntry struct {
	PlayerID string
	Type     ByeType
}

// TournamentInfo holds tournament metadata for display and TRF round-trip fidelity.
// Engines ignore this struct; it is populated from TRF header lines and
// written back when serializing to TRF.
type TournamentInfo struct {
	Name          string
	City          string
	Federation    string // Organizing federation code
	StartDate     string // YYYY/MM/DD
	EndDate       string // YYYY/MM/DD
	ChiefArbiter  string
	DeputyArbiter string
	TimeControl   string   // Allotted time description
	RoundDates    []string // YYYY/MM/DD per round
}

// TournamentState is the read-only snapshot of a tournament passed to engines.
// The caller constructs this from their data source before calling any engine
// method. Engines never perform I/O directly.
//
// Rounds holds completed rounds only (round numbers 1..CurrentRound-1).
// CurrentRound is the 1-based round about to be paired. PreAssignedByes
// declares byes locked in for that upcoming round (e.g. a player notified
// the arbiter in advance that they will skip the round). Pairers exclude
// these players from the matching pool and echo their bye entries back in
// PairingResult.Byes. The roundrobin pairer rejects non-empty
// PreAssignedByes because the Berger schedule is fixed.
type TournamentState struct {
	Players         []PlayerEntry
	Rounds          []RoundData
	CurrentRound    int
	PreAssignedByes []ByeEntry
	PairingConfig   PairingConfig
	ScoringConfig   ScoringConfig
	Info            TournamentInfo // Tournament metadata. Zero value if not set.
}

// PlayerEntry represents a player for engine purposes.
//
// JoinedRound and WithdrawnAfterRound bracket the player's active window.
// JoinedRound = 0 or 1 means the player was present from round 1.
// WithdrawnAfterRound names the last round in which the player participated;
// from *WithdrawnAfterRound + 1 onward they are inactive. nil means the
// player has not withdrawn. Use TournamentState.IsActiveInRound rather
// than reading these fields directly.
type PlayerEntry struct {
	ID                  string
	DisplayName         string
	Rating              int
	Federation          string // FIDE federation code (e.g. "NED", "USA", "IND"). Empty if unknown.
	FideID              string // FIDE player ID number. Empty if unknown.
	Title               string // FIDE title code (GM, IM, FM, WGM, WIM, WFM, CM, WCM). Empty if untitled.
	Sex                 string // "m" or "w". Empty if unknown.
	BirthDate           string // Birth date as YYYY/MM/DD. Empty if unknown.
	JoinedRound         int    // Round number the player joined. 0 or 1 means original player (joined from the start).
	WithdrawnAfterRound *int   // Last round the player participated in; nil means still active.
}

// RoundData contains all games for a completed round.
type RoundData struct {
	Number int
	Games  []GameData
	Byes   []ByeEntry
}

// GameData is a single game result for engine consumption.
type GameData struct {
	WhiteID   string
	BlackID   string
	Result    GameResult
	IsForfeit bool
}

// ResultContext provides additional information needed by scoring systems
// when calculating points for a specific game result.
type ResultContext struct {
	OpponentRank        int
	OpponentValueNumber int
	PlayerRank          int
	PlayerValueNumber   int
	IsBye               bool
	IsAbsent            bool
	IsForfeit           bool
}

// PairingResult is returned by a Pairer.
type PairingResult struct {
	Pairings []GamePairing
	Byes     []ByeEntry
	Notes    []string
}

// GamePairing is a single pairing assignment for a round.
type GamePairing struct {
	Board   int
	WhiteID string
	BlackID string
}

// PlayerScore holds a player's calculated score from the scoring engine.
type PlayerScore struct {
	PlayerID string
	Score    float64
	Rank     int
}

// TieBreakValue is a single tiebreak computation for one player.
type TieBreakValue struct {
	PlayerID string
	Value    float64
}

// Standing is the final ranked output combining score and tiebreakers.
type Standing struct {
	Rank        int          `json:"rank"`
	PlayerID    string       `json:"playerId"`
	DisplayName string       `json:"displayName"`
	Score       float64      `json:"score"`
	TieBreakers []NamedValue `json:"tieBreakers"`
	GamesPlayed int          `json:"gamesPlayed"`
	Wins        int          `json:"wins"`
	Draws       int          `json:"draws"`
	Losses      int          `json:"losses"`
}

// NamedValue pairs a tiebreaker identifier with its computed value.
type NamedValue struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// Validate checks structural invariants of the tournament state.
// Returns an error describing the first problem found, or nil if valid.
func (s *TournamentState) Validate() error {
	if len(s.Players) == 0 {
		return fmt.Errorf("no players in tournament state")
	}

	seen := make(map[string]bool, len(s.Players))
	for i, p := range s.Players {
		if p.ID == "" {
			return fmt.Errorf("empty player ID at index %d", i)
		}
		if seen[p.ID] {
			return fmt.Errorf("duplicate player ID %q", p.ID)
		}
		seen[p.ID] = true
	}

	if s.CurrentRound > len(s.Rounds) {
		return fmt.Errorf("CurrentRound (%d) exceeds number of rounds (%d)", s.CurrentRound, len(s.Rounds))
	}

	if len(s.PreAssignedByes) > 0 {
		seenBye := make(map[string]bool, len(s.PreAssignedByes))
		for i, b := range s.PreAssignedByes {
			if !seen[b.PlayerID] {
				return fmt.Errorf("PreAssignedByes[%d]: unknown player ID %q", i, b.PlayerID)
			}
			if seenBye[b.PlayerID] {
				return fmt.Errorf("PreAssignedByes[%d]: duplicate player ID %q", i, b.PlayerID)
			}
			if !b.Type.IsValid() {
				return fmt.Errorf("PreAssignedByes[%d]: invalid bye type %d for player %q", i, b.Type, b.PlayerID)
			}
			seenBye[b.PlayerID] = true
		}
	}

	for i, p := range s.Players {
		if p.WithdrawnAfterRound == nil {
			continue
		}
		w := *p.WithdrawnAfterRound
		if w <= 0 {
			return fmt.Errorf("player %q (Players[%d]): WithdrawnAfterRound %d must be positive", p.ID, i, w)
		}
		if w > s.CurrentRound {
			return fmt.Errorf("player %q (Players[%d]): WithdrawnAfterRound %d exceeds CurrentRound %d", p.ID, i, w, s.CurrentRound)
		}
	}

	return nil
}

// IsActiveInRound reports whether the player with the given ID exists in the
// tournament and is participating in the given 1-indexed round. A player is
// active in round r when their JoinedRound is at most r (treating 0 as 1)
// and either WithdrawnAfterRound is nil or *WithdrawnAfterRound >= r.
// Unknown player IDs return false.
//
// As a convenience for scoring callers that operate over the entire played
// history without a specific round anchor, round <= 0 means "no round filter":
// any enrolled player who has not been withdrawn (WithdrawnAfterRound == nil)
// is considered active. A withdrawn player is excluded regardless of when.
func (s *TournamentState) IsActiveInRound(playerID string, round int) bool {
	for i := range s.Players {
		p := &s.Players[i]
		if p.ID != playerID {
			continue
		}
		if round <= 0 {
			return p.WithdrawnAfterRound == nil
		}
		joined := p.JoinedRound
		if joined < 1 {
			joined = 1
		}
		if joined > round {
			return false
		}
		if p.WithdrawnAfterRound != nil && *p.WithdrawnAfterRound < round {
			return false
		}
		return true
	}
	return false
}

// ActivePlayerIDs returns the IDs of players active in the given round, in
// the order they appear in s.Players. See IsActiveInRound for the predicate,
// including the round <= 0 convenience.
func (s *TournamentState) ActivePlayerIDs(round int) []string {
	out := make([]string, 0, len(s.Players))
	for i := range s.Players {
		p := &s.Players[i]
		if round <= 0 {
			if p.WithdrawnAfterRound == nil {
				out = append(out, p.ID)
			}
			continue
		}
		joined := p.JoinedRound
		if joined < 1 {
			joined = 1
		}
		if joined > round {
			continue
		}
		if p.WithdrawnAfterRound != nil && *p.WithdrawnAfterRound < round {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}
