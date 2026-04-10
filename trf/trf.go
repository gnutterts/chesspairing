// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package trf implements reading and writing of TRF16 (FIDE Tournament Report
// File) documents. It provides a faithful in-memory representation of TRF data
// and bidirectional conversion to/from chesspairing.TournamentState.
package trf

import "fmt"

// Document represents a complete TRF file (supports both TRF16 and TRF-2026).
type Document struct {
	// Tournament info (header lines 012-132)
	Name           string   `json:"name,omitempty"`           // 012
	City           string   `json:"city,omitempty"`           // 022
	Federation     string   `json:"federation,omitempty"`     // 032
	StartDate      string   `json:"startDate,omitempty"`      // 042
	EndDate        string   `json:"endDate,omitempty"`        // 052
	NumPlayers     int      `json:"numPlayers,omitempty"`     // 062
	NumRated       int      `json:"numRated,omitempty"`       // 072
	NumTeams       int      `json:"numTeams,omitempty"`       // 082
	TournamentType string   `json:"tournamentType,omitempty"` // 092
	ChiefArbiter   string   `json:"chiefArbiter,omitempty"`   // 102
	DeputyArbiter  string   `json:"deputyArbiter,omitempty"`  // 112 (first or only)
	DeputyArbiters []string `json:"deputyArbiters,omitempty"` // 112 (all, for TRF-2026 multiple deputy arbiters)
	TimeControl    string   `json:"timeControl,omitempty"`    // 122
	RoundDates     []string `json:"roundDates,omitempty"`     // 132

	// TRF-2026 header fields (new record types)
	TotalRounds26       int    `json:"totalRounds26,omitempty"`       // 142 (replaces XXR)
	InitialColor26      string `json:"initialColor26,omitempty"`      // 152 (replaces XXC; "B" or "W")
	ScoringSystem       string `json:"scoringSystem,omitempty"`       // 162 (e.g. " W 1.0    D 0.5    L 0.0")
	StartingRankMethod  string `json:"startingRankMethod,omitempty"`  // 172 (e.g. "IND FIDE")
	CodedTournamentType string `json:"codedTournamentType,omitempty"` // 192 (e.g. "FIDE_TEAM_BAKU")
	TieBreakDef         string `json:"tieBreakDef,omitempty"`         // 202 (e.g. "EDET/P,EMGSB/C1/P,BH:MP/C1/P,MPvGP")
	EncodedTimeControl  string `json:"encodedTimeControl,omitempty"`  // 222 (e.g. "40/6000+30:20/3000+30:1500+30")
	TeamInitialColor    string `json:"teamInitialColor,omitempty"`    // 352 (e.g. "WBWB")
	TeamScoringSystem   string `json:"teamScoringSystem,omitempty"`   // 362 (e.g. "TW 2     TD 1     TL 0")

	// TRF-2026 data records
	Absences            []AbsenceRecord       `json:"absences,omitempty"`            // 240 lines
	Accelerations26     []AccelerationRecord  `json:"accelerations26,omitempty"`     // 250 lines (replaces XXS)
	ForbiddenPairs26    []ForbiddenPairRecord `json:"forbiddenPairs26,omitempty"`    // 260 lines (replaces XXP)
	NewTeams            []NewTeamLine         `json:"newTeams,omitempty"`            // 310 lines
	TeamRoundData       []TeamRoundEntry      `json:"teamRoundData,omitempty"`       // 300 lines
	TeamRoundScores     []TeamRoundScoreEntry `json:"teamRoundScores,omitempty"`     // 320 lines
	OldAbsentForfeits   []OldAbsentForfeit    `json:"oldAbsentForfeits,omitempty"`   // 330 lines
	DetailedTeamResults []DetailedTeamResult  `json:"detailedTeamResults,omitempty"` // 801 lines
	SimpleTeamResults   []SimpleTeamResult    `json:"simpleTeamResults,omitempty"`   // 802 lines
	NRSRecords          []NRSRecord           `json:"nrsRecords,omitempty"`          // NRS lines (3-letter federation code)
	Comments            []string              `json:"comments,omitempty"`            // ### lines

	// Extended data lines (TRF16 / legacy)
	TotalRounds    int             `json:"totalRounds,omitempty"`    // XXR
	InitialColor   string          `json:"initialColor,omitempty"`   // XXC (e.g. "white1")
	Acceleration   []string        `json:"acceleration,omitempty"`   // XXS lines (one per line)
	ForbiddenPairs []ForbiddenPair `json:"forbiddenPairs,omitempty"` // XXP lines

	// System-specific extended data lines
	Cycles                  int    `json:"cycles,omitempty"`                  // XXY (Round-Robin: 1=single, 2=double)
	ColorBalance            *bool  `json:"colorBalance,omitempty"`            // XXB (Round-Robin: true/false)
	MaxiTournament          *bool  `json:"maxiTournament,omitempty"`          // XXM (Lim: true/false)
	ColorPreferenceType     string `json:"colorPreferenceType,omitempty"`     // XXT (Team: "A", "B", "none")
	PrimaryScore            string `json:"primaryScore,omitempty"`            // XXG (Team: "match", "game")
	AllowRepeatPairings     *bool  `json:"allowRepeatPairings,omitempty"`     // XXA (Keizer: true/false)
	MinRoundsBetweenRepeats int    `json:"minRoundsBetweenRepeats,omitempty"` // XXK (Keizer: integer)

	// Player data
	Players []PlayerLine `json:"players,omitempty"` // 001 lines, sorted by StartNumber

	// Team data (TRF16 legacy)
	Teams []TeamLine `json:"teams,omitempty"` // 013 lines

	// Unknown/custom lines preserved for round-trip fidelity
	Other []RawLine `json:"other,omitempty"`
}

// EffectiveTotalRounds returns the total rounds from TRF-2026 (142) if set,
// falling back to TRF16 (XXR). Returns 0 if neither is set.
func (doc *Document) EffectiveTotalRounds() int {
	if doc.TotalRounds26 > 0 {
		return doc.TotalRounds26
	}
	return doc.TotalRounds
}

// EffectiveInitialColor returns the initial color from TRF-2026 (152) if set,
// falling back to TRF16 (XXC). Returns "" if neither is set.
func (doc *Document) EffectiveInitialColor() string {
	if doc.InitialColor26 != "" {
		return doc.InitialColor26
	}
	return doc.InitialColor
}

// PlayerLine represents a single 001 player line.
type PlayerLine struct {
	StartNumber int           `json:"startNumber"`
	Sex         string        `json:"sex,omitempty"`
	Title       string        `json:"title,omitempty"`
	Name        string        `json:"name,omitempty"`
	Rating      int           `json:"rating,omitempty"`
	Federation  string        `json:"federation,omitempty"`
	FideID      string        `json:"fideID,omitempty"`
	BirthDate   string        `json:"birthDate,omitempty"`
	Points      float64       `json:"points"`
	Rank        int           `json:"rank"`
	Rounds      []RoundResult `json:"rounds,omitempty"`
}

// RoundResult is a single round entry from a player's 001 line.
type RoundResult struct {
	Opponent int        `json:"opponent"` // Start number of opponent (0 = no opponent / bye)
	Color    Color      `json:"color"`    // White, Black, or None
	Result   ResultCode `json:"result"`   // Win, Loss, Draw, ForfeitWin, ForfeitLoss, etc.
}

// Color in a TRF round result.
type Color int

const (
	ColorNone  Color = iota // "-" (bye, absent, no game)
	ColorWhite              // "w"
	ColorBlack              // "b"
)

// IsValid returns true if the color is a recognized value.
func (c Color) IsValid() bool {
	return c >= ColorNone && c <= ColorBlack
}

// String returns the human-readable name of the color.
func (c Color) String() string {
	switch c {
	case ColorNone:
		return "None"
	case ColorWhite:
		return "White"
	case ColorBlack:
		return "Black"
	default:
		return "Unknown"
	}
}

// Char returns the TRF character for the color.
func (c Color) Char() byte {
	switch c {
	case ColorWhite:
		return 'w'
	case ColorBlack:
		return 'b'
	default:
		return '-'
	}
}

// ResultCode is a TRF result character.
type ResultCode int

const (
	ResultWin           ResultCode = iota // "1" - win (played)
	ResultLoss                            // "0" - loss (played)
	ResultDraw                            // "=" - draw
	ResultForfeitWin                      // "+" - win by forfeit
	ResultForfeitLoss                     // "-" - loss by forfeit
	ResultHalfBye                         // "H" - half-point bye
	ResultFullBye                         // "F" - full-point bye (PAB)
	ResultUnpaired                        // "U" - unpaired (absent, 0 pts)
	ResultZeroBye                         // "Z" - zero-point bye
	ResultNotPlayed                       // "*" - not yet played
	ResultWinByDefault                    // "W" - win, opponent absent
	ResultDrawByDefault                   // "D" - draw by default
	ResultLossByDefault                   // "L" - loss by default
)

// IsValid returns true if the result code is a recognized value.
func (rc ResultCode) IsValid() bool {
	return rc >= ResultWin && rc <= ResultLossByDefault
}

// String returns the human-readable name of the result code.
func (rc ResultCode) String() string {
	switch rc {
	case ResultWin:
		return "Win"
	case ResultLoss:
		return "Loss"
	case ResultDraw:
		return "Draw"
	case ResultForfeitWin:
		return "ForfeitWin"
	case ResultForfeitLoss:
		return "ForfeitLoss"
	case ResultHalfBye:
		return "HalfBye"
	case ResultFullBye:
		return "FullBye"
	case ResultUnpaired:
		return "Unpaired"
	case ResultZeroBye:
		return "ZeroBye"
	case ResultNotPlayed:
		return "NotPlayed"
	case ResultWinByDefault:
		return "WinByDefault"
	case ResultDrawByDefault:
		return "DrawByDefault"
	case ResultLossByDefault:
		return "LossByDefault"
	default:
		return "Unknown"
	}
}

// Char returns the TRF character for the result code.
func (rc ResultCode) Char() byte {
	switch rc {
	case ResultWin:
		return '1'
	case ResultLoss:
		return '0'
	case ResultDraw:
		return '='
	case ResultForfeitWin:
		return '+'
	case ResultForfeitLoss:
		return '-'
	case ResultHalfBye:
		return 'H'
	case ResultFullBye:
		return 'F'
	case ResultUnpaired:
		return 'U'
	case ResultZeroBye:
		return 'Z'
	case ResultNotPlayed:
		return '*'
	case ResultWinByDefault:
		return 'W'
	case ResultDrawByDefault:
		return 'D'
	case ResultLossByDefault:
		return 'L'
	default:
		return '?'
	}
}

// parseResultChar converts a TRF result character to a ResultCode.
func parseResultChar(ch byte) (ResultCode, bool) {
	switch ch {
	case '1':
		return ResultWin, true
	case '0':
		return ResultLoss, true
	case '=':
		return ResultDraw, true
	case '+':
		return ResultForfeitWin, true
	case '-':
		return ResultForfeitLoss, true
	case 'H':
		return ResultHalfBye, true
	case 'F':
		return ResultFullBye, true
	case 'U':
		return ResultUnpaired, true
	case 'Z':
		return ResultZeroBye, true
	case '*':
		return ResultNotPlayed, true
	case 'W':
		return ResultWinByDefault, true
	case 'D':
		return ResultDrawByDefault, true
	case 'L':
		return ResultLossByDefault, true
	default:
		return 0, false
	}
}

// parseColorChar converts a TRF color character to a Color.
func parseColorChar(ch byte) (Color, bool) {
	switch ch {
	case 'w':
		return ColorWhite, true
	case 'b':
		return ColorBlack, true
	case '-':
		return ColorNone, true
	default:
		return 0, false
	}
}

// TeamLine represents a 013 team line.
type TeamLine struct {
	TeamNumber int    `json:"teamNumber"`
	TeamName   string `json:"teamName"`
	Members    []int  `json:"members"` // Start numbers of team members
}

// ForbiddenPair represents an XXP forbidden pair entry.
type ForbiddenPair struct {
	Player1 int `json:"player1"` // Start number
	Player2 int `json:"player2"` // Start number
}

// RawLine preserves an unrecognized line for round-trip fidelity.
type RawLine struct {
	Code string `json:"code"` // The 3-character line code
	Data string `json:"data"` // Everything after the code and space
}

// ParseError describes a TRF parsing error with line context.
type ParseError struct {
	Line    int    `json:"line"`    // 1-based line number in the input
	Code    string `json:"code"`    // Line code (e.g., "001", "012", "XXR")
	Message string `json:"message"` // Human-readable description
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("trf: line %d (%s): %s", e.Line, e.Code, e.Message)
}

// isByeResult returns true if the result code represents a bye (no opponent).
func (rc ResultCode) isByeResult() bool {
	switch rc {
	case ResultHalfBye, ResultFullBye, ResultUnpaired, ResultZeroBye:
		return true
	default:
		return false
	}
}

// --- TRF-2026 new record types ---

// AbsenceRecord represents a 240 line (absent participants).
// Format: 240 T RRR TOI1 TOI2 ...
// T = "F" (full forfeit) or "H" (half-point bye)
// RRR = round number
// TOI = team-or-individual start numbers
type AbsenceRecord struct {
	Type    string `json:"type"`    // "F" or "H"
	Round   int    `json:"round"`   // Round number
	Players []int  `json:"players"` // Start numbers of absent players/teams
}

// AccelerationRecord represents a 250 line (Baku acceleration data).
// Format: 250 MMMM GGGG RRF RRL PPPF PPPL
// MMMM = match points to add, GGGG = game points to add
// RRF/RRL = first/last round, PPPF/PPPL = first/last player
type AccelerationRecord struct {
	MatchPoints float64 `json:"matchPoints"` // Match points to add (for team)
	GamePoints  float64 `json:"gamePoints"`  // Game points to add
	FirstRound  int     `json:"firstRound"`  // First round of acceleration
	LastRound   int     `json:"lastRound"`   // Last round of acceleration
	FirstPlayer int     `json:"firstPlayer"` // First player/team number
	LastPlayer  int     `json:"lastPlayer"`  // Last player/team number
	Raw         string  `json:"raw"`         // Raw line data for round-trip
}

// ForbiddenPairRecord represents a 260 line (forbidden/prohibited pairings).
// Format: 260 RR1 RRL TOI1 TOI2 ...
// RR1/RRL = first/last round of restriction
// TOI = player/team numbers that cannot be paired against each other
type ForbiddenPairRecord struct {
	FirstRound int    `json:"firstRound"` // First round of restriction
	LastRound  int    `json:"lastRound"`  // Last round of restriction
	Players    []int  `json:"players"`    // Start numbers that are mutually forbidden
	Raw        string `json:"raw"`        // Raw line data for round-trip
}

// NewTeamLine represents a 310 line (TRF-2026 team record).
// Format: 310 SSS NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN FFFFF EEEEEE MMMMMM GGGGGG RRR  PPP1 PPP2...
type NewTeamLine struct {
	TeamNumber  int     `json:"teamNumber"`            // 3-char team number
	TeamName    string  `json:"teamName"`              // 32-char team name
	Federation  string  `json:"federation,omitempty"`  // 5-char federation
	AvgRating   float64 `json:"avgRating,omitempty"`   // 6-char average rating
	MatchPoints float64 `json:"matchPoints,omitempty"` // 6-char match points
	GamePoints  float64 `json:"gamePoints,omitempty"`  // 6-char game points
	Rank        int     `json:"rank,omitempty"`        // 3-char rank
	Members     []int   `json:"members"`               // Member start numbers (4 chars each)
}

// TeamRoundEntry represents a 300 line (team round data / board assignments).
// Format: 300 RRR TT1 TT2 PPP1 PPP2 PPP3 PPP4
// RRR = round, TT1/TT2 = team numbers, PPP = player start numbers on boards
type TeamRoundEntry struct {
	Round  int   `json:"round"`  // Round number
	Team1  int   `json:"team1"`  // First team number
	Team2  int   `json:"team2"`  // Second team number
	Boards []int `json:"boards"` // Player start numbers for each board (0 = empty)
}

// TeamRoundScoreEntry represents a 320 line (team round-by-round scores).
// Format: 320 TTT GGGG RRR1 RRR2 ...
type TeamRoundScoreEntry struct {
	TeamNumber int      `json:"teamNumber"`
	GamePoints float64  `json:"gamePoints"`
	Scores     []string `json:"scores"` // Per-round scores as strings (may be empty)
	Raw        string   `json:"raw"`    // Raw line data for round-trip
}

// OldAbsentForfeit represents a 330 line (legacy absent/forfeit records).
// Format: 330 TT RRR WWW BBB
// TT = result code ("+-", "-+", "--")
// RRR = round, WWW = white team, BBB = black team
type OldAbsentForfeit struct {
	ResultType string `json:"resultType"` // "+-", "-+", "--"
	Round      int    `json:"round"`      // Round number
	WhiteTeam  int    `json:"whiteTeam"`  // White team number
	BlackTeam  int    `json:"blackTeam"`  // Black team number
}

// DetailedTeamResult represents an 801 line (detailed team match results per round).
// Format: 801 TT NNNNN MMMM GGGG T01 C RRRR BBBB ...
type DetailedTeamResult struct {
	TeamNumber  int                 `json:"teamNumber"`
	TeamName    string              `json:"teamName"`
	MatchPoints float64             `json:"matchPoints"`
	GamePoints  float64             `json:"gamePoints"`
	Rounds      []DetailedTeamRound `json:"rounds"`
	Raw         string              `json:"raw"` // Raw line data for round-trip
}

// DetailedTeamRound is a single round entry in an 801 line.
type DetailedTeamRound struct {
	Opponent   int    `json:"opponent"`          // Opponent team number (0 = no match)
	Color      string `json:"color"`             // "w" or "b"
	Results    string `json:"results"`           // Individual board results (e.g. "=0=1")
	BoardOrder string `json:"boardOrder"`        // Board order (e.g. "1234")
	ByeType    string `json:"byeType,omitempty"` // "FFFF", "HHHH", "ZZZZ", "UUUU" for byes, empty for normal
}

// SimpleTeamResult represents an 802 line (simpler team round results).
// Format: 802 TTT NNNNN MMMMMM GGGGGG T01 C GGGGf ...
type SimpleTeamResult struct {
	TeamNumber  int               `json:"teamNumber"`
	TeamName    string            `json:"teamName"`
	MatchPoints float64           `json:"matchPoints"`
	GamePoints  float64           `json:"gamePoints"`
	Rounds      []SimpleTeamRound `json:"rounds"`
	Raw         string            `json:"raw"` // Raw line data for round-trip
}

// SimpleTeamRound is a single round entry in an 802 line.
type SimpleTeamRound struct {
	Opponent   int     `json:"opponent"`          // Opponent team number (0 = no match)
	Color      string  `json:"color"`             // "w" or "b"
	GamePoints float64 `json:"gamePoints"`        // Game points scored
	Forfeit    bool    `json:"forfeit,omitempty"` // Whether result includes forfeits
	ByeType    string  `json:"byeType,omitempty"` // "FPB", "HPB", "ZPB", "PAB" for byes, empty for normal
}

// NRSRecord represents a national rating system record line.
// Lines start with a 3-letter federation code instead of a numeric record code.
// Format is similar to 001 but may include national rating and sub-federation data.
type NRSRecord struct {
	Federation     string `json:"federation"`               // 3-letter federation code (e.g. "IND")
	StartNumber    int    `json:"startNumber"`              // Player start number
	Title          string `json:"title,omitempty"`          // Title
	Name           string `json:"name,omitempty"`           // Player name
	NationalRating int    `json:"nationalRating,omitempty"` // National rating
	SubFederation  string `json:"subFederation,omitempty"`  // Sub-federation code
	NationalID     string `json:"nationalID,omitempty"`     // National ID
	BirthDate      string `json:"birthDate,omitempty"`      // Birth date
	Raw            string `json:"raw"`                      // Raw line data for round-trip
}
