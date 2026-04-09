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
type TournamentState struct {
	Players       []PlayerEntry
	Rounds        []RoundData
	CurrentRound  int
	PairingConfig PairingConfig
	ScoringConfig ScoringConfig
	Info          TournamentInfo // Tournament metadata. Zero value if not set.
}

// PlayerEntry represents a player for engine purposes.
type PlayerEntry struct {
	ID          string
	DisplayName string
	Rating      int
	Active      bool
	Federation  string // FIDE federation code (e.g. "NED", "USA", "IND"). Empty if unknown.
	FideID      string // FIDE player ID number. Empty if unknown.
	Title       string // FIDE title code (GM, IM, FM, WGM, WIM, WFM, CM, WCM). Empty if untitled.
	Sex         string // "m" or "w". Empty if unknown.
	BirthDate   string // Birth date as YYYY/MM/DD. Empty if unknown.
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

	return nil
}
