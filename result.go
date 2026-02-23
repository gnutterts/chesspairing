package chesspairing

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

// TournamentState is the read-only snapshot of a tournament passed to engines.
// The service layer constructs this from Firestore data before calling
// any engine method. Engines never write to the database directly.
type TournamentState struct {
	Players       []PlayerEntry
	Rounds        []RoundData
	CurrentRound  int
	PairingConfig PairingConfig
	ScoringConfig ScoringConfig
}

// PlayerEntry represents a player for engine purposes.
type PlayerEntry struct {
	ID          string
	DisplayName string
	Rating      int
	Active      bool
}

// RoundData contains all games for a completed round.
type RoundData struct {
	Number int
	Games  []GameData
	Byes   []string
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
	Byes     []string
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
