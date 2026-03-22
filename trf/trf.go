// Package trf implements reading and writing of TRF16 (FIDE Tournament Report
// File) documents. It provides a faithful in-memory representation of TRF data
// and bidirectional conversion to/from chesspairing.TournamentState.
package trf

import "fmt"

// Document represents a complete TRF16 file.
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
	DeputyArbiter  string   `json:"deputyArbiter,omitempty"`  // 112
	TimeControl    string   `json:"timeControl,omitempty"`    // 122
	RoundDates     []string `json:"roundDates,omitempty"`     // 132

	// Extended data lines
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

	// Team data
	Teams []TeamLine `json:"teams,omitempty"` // 013 lines

	// Unknown/custom lines preserved for round-trip fidelity
	Other []RawLine `json:"other,omitempty"`
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
