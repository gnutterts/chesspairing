---
title: "TRF16 I/O"
linkTitle: "TRF I/O"
weight: 7
description: "Reading, writing, and converting TRF16 files with the trf package."
---

The `trf` package (`github.com/gnutterts/chesspairing/trf`) handles reading, writing, validating, and converting FIDE Tournament Report Files. It supports both TRF16 (legacy) and TRF-2026 formats.

## Core types

### Document

The main type representing a complete TRF file. Key fields (abbreviated):

```go
type Document struct {
    // Tournament info (header lines 012-132)
    Name           string   `json:"name,omitempty"`
    City           string   `json:"city,omitempty"`
    Federation     string   `json:"federation,omitempty"`
    StartDate      string   `json:"startDate,omitempty"`
    EndDate        string   `json:"endDate,omitempty"`
    NumPlayers     int      `json:"numPlayers,omitempty"`
    NumRated       int      `json:"numRated,omitempty"`
    TournamentType string   `json:"tournamentType,omitempty"`
    ChiefArbiter   string   `json:"chiefArbiter,omitempty"`
    TimeControl    string   `json:"timeControl,omitempty"`
    RoundDates     []string `json:"roundDates,omitempty"`

    // TRF-2026 header fields
    TotalRounds26  int    `json:"totalRounds26,omitempty"`  // 142
    InitialColor26 string `json:"initialColor26,omitempty"` // 152

    // TRF16 extended data
    TotalRounds    int             `json:"totalRounds,omitempty"`    // XXR
    InitialColor   string          `json:"initialColor,omitempty"`   // XXC
    Acceleration   []string        `json:"acceleration,omitempty"`   // XXS
    ForbiddenPairs []ForbiddenPair `json:"forbiddenPairs,omitempty"` // XXP

    // System-specific fields (see below)
    Cycles                  int    `json:"cycles,omitempty"`                  // XXY
    ColorBalance            *bool  `json:"colorBalance,omitempty"`            // XXB
    MaxiTournament          *bool  `json:"maxiTournament,omitempty"`          // XXM
    ColorPreferenceType     string `json:"colorPreferenceType,omitempty"`     // XXT
    PrimaryScore            string `json:"primaryScore,omitempty"`            // XXG
    AllowRepeatPairings     *bool  `json:"allowRepeatPairings,omitempty"`     // XXA
    MinRoundsBetweenRepeats int    `json:"minRoundsBetweenRepeats,omitempty"` // XXK

    // Player and team data
    Players []PlayerLine `json:"players,omitempty"` // 001 lines
    Teams   []TeamLine   `json:"teams,omitempty"`   // 013 lines

    // Comment region
    Comments               []string    `json:"comments,omitempty"`               // free-form ### lines
    ChesspairingDirectives []Directive `json:"chesspairingDirectives,omitempty"` // ### chesspairing:<verb>

    // Unknown lines preserved for round-trip fidelity
    Other []RawLine `json:"other,omitempty"`
}
```

All struct fields have JSON tags for serialization.

`ChesspairingDirectives` carries typed `### chesspairing:<verb> k=v ...`
comment lines that express data outside the FIDE TRF vocabulary, such
as `ByeExcused`/`ByeClubCommitment` pre-assigned byes and per-player
withdrawals. Each entry is a `Directive{Verb string, Params
map[string]string}`. See [TRF-2026
Extensions](/docs/formats/trf-extensions/) for the directive grammar
and bridging behaviour.

#### EffectiveTotalRounds

```go
func (doc *Document) EffectiveTotalRounds() int
```

Returns the total rounds from TRF-2026 field 142 if set, falling back to TRF16 field XXR. Returns 0 if neither is set.

#### EffectiveInitialColor

```go
func (doc *Document) EffectiveInitialColor() string
```

Returns the initial color from TRF-2026 field 152 if set, falling back to TRF16 field XXC. Returns `""` if neither is set.

### PlayerLine

```go
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
```

Represents a single 001 player line. The `Rounds` slice contains one `RoundResult` per round the player has data for.

### RoundResult

```go
type RoundResult struct {
    Opponent int        `json:"opponent"` // Start number (0 = bye/absent)
    Color    Color      `json:"color"`
    Result   ResultCode `json:"result"`
}
```

A single round entry from a player's 001 line.

### Color

```go
type Color int

const (
    ColorNone  Color = 0 // "-" (bye, absent, no game)
    ColorWhite Color = 1 // "w"
    ColorBlack Color = 2 // "b"
)
```

Methods:

| Method          | Signature                                          | Description                                |
| --------------- | -------------------------------------------------- | ------------------------------------------ |
| `IsValid`       | `func (c Color) IsValid() bool`                    | True for ColorNone, ColorWhite, ColorBlack |
| `String`        | `func (c Color) String() string`                   | Returns `"None"`, `"White"`, or `"Black"`  |
| `Char`          | `func (c Color) Char() byte`                       | Returns `'-'`, `'w'`, or `'b'`             |
| `MarshalJSON`   | `func (c Color) MarshalJSON() ([]byte, error)`     | JSON serialization                         |
| `UnmarshalJSON` | `func (c *Color) UnmarshalJSON(data []byte) error` | JSON deserialization                       |

### ResultCode

```go
type ResultCode int
```

13 constants representing all possible TRF result characters:

| Constant              | Value | Char | Meaning                  |
| --------------------- | ----- | ---- | ------------------------ |
| `ResultWin`           | 0     | `1`  | Win (played)             |
| `ResultLoss`          | 1     | `0`  | Loss (played)            |
| `ResultDraw`          | 2     | `=`  | Draw                     |
| `ResultForfeitWin`    | 3     | `+`  | Win by forfeit           |
| `ResultForfeitLoss`   | 4     | `-`  | Loss by forfeit          |
| `ResultHalfBye`       | 5     | `H`  | Half-point bye           |
| `ResultFullBye`       | 6     | `F`  | Full-point bye (PAB)     |
| `ResultUnpaired`      | 7     | `U`  | Unpaired (absent, 0 pts) |
| `ResultZeroBye`       | 8     | `Z`  | Zero-point bye           |
| `ResultNotPlayed`     | 9     | `*`  | Not yet played           |
| `ResultWinByDefault`  | 10    | `W`  | Win, opponent absent     |
| `ResultDrawByDefault` | 11    | `D`  | Draw by default          |
| `ResultLossByDefault` | 12    | `L`  | Loss by default          |

Methods: `IsValid() bool`, `String() string`, `Char() byte`, `MarshalJSON`, `UnmarshalJSON`.

## Reading

```go
func Read(r io.Reader) (*Document, error)
```

Parses a TRF file from any `io.Reader`. Supports both TRF16 and TRF-2026 formats (detected automatically from the line codes present).

Returns a `*ParseError` for parsing failures. ParseError includes the 1-based line number and line code for diagnostics:

```go
type ParseError struct {
    Line    int    `json:"line"`
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

Example:

```go
f, err := os.Open("tournament.trf")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

doc, err := trf.Read(f)
if err != nil {
    var pe *trf.ParseError
    if errors.As(err, &pe) {
        log.Printf("parse error at line %d (%s): %s", pe.Line, pe.Code, pe.Message)
    }
    log.Fatal(err)
}
```

## Writing

```go
func Write(w io.Writer, doc *Document) error
```

Serializes a Document to TRF format, writing to any `io.Writer`. Handles both TRF16 and TRF-2026 fields -- it writes whichever fields are populated in the Document.

```go
var buf bytes.Buffer
if err := trf.Write(&buf, doc); err != nil {
    log.Fatal(err)
}
// buf.String() contains the TRF file content
```

## Converting

Bidirectional conversion between `Document` and `chesspairing.TournamentState`.

### ToTournamentState

```go
func (doc *Document) ToTournamentState() (*chesspairing.TournamentState, error)
```

Converts a TRF Document to a `TournamentState` suitable for engine use.

Key behaviors:

- **Player IDs** are string start numbers (`"1"`, `"2"`, ...).
- **Rounds** are reconstructed by cross-referencing per-player results. Games are deduplicated (each game appears once, not twice).
- **Pairing system** is inferred from the TRF tournament type field (e.g. `"Swiss Dutch"` maps to `PairingDutch`).
- **System-specific options** from XX fields (XXR, XXC, XXS, XXP, XXY, XXB, XXM, XXT, XXG, XXA, XXK) and TRF-2026 equivalents are placed into `PairingConfig.Options` for round-tripping.
- **Scoring config** defaults to `ScoringStandard` with `DefaultTiebreakers()` for the inferred pairing system.

```go
doc, err := trf.Read(reader)
if err != nil {
    log.Fatal(err)
}

state, err := doc.ToTournamentState()
if err != nil {
    log.Fatal(err)
}
// state.PairingConfig.System == chesspairing.PairingDutch
// state.Players[0].ID == "1"
```

### FromTournamentState

```go
func FromTournamentState(state *chesspairing.TournamentState) (*Document, map[string]int)
```

Creates a Document from a TournamentState. Returns the Document and a mapping from player ID to assigned start number.

Key behaviors:

- **Start numbers** are assigned by sorting players by rating descending, then by name ascending, then by ID ascending (for determinism).
- **Player map** lets callers look up which start number was assigned to each player ID.
- **Round results** are reconstructed from `state.Rounds` game data and byes.
- **Tournament type** and **XX fields** are set from `PairingConfig.System` and `PairingConfig.Options`.

```go
doc, playerMap := trf.FromTournamentState(&state)
// playerMap["player-uuid-1"] == 1  (highest rated)
// playerMap["player-uuid-2"] == 2
err := trf.Write(os.Stdout, doc)
```

## Validation

```go
func (doc *Document) Validate(profile ValidationProfile) []ValidationIssue
```

Checks a Document for completeness according to a validation profile. Each profile is a superset of the previous one.

### Profiles

| Profile        | Constant                    | Checks                                                                                                                                     |
| -------------- | --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| General        | `ValidateGeneral` (0)       | Structural integrity: players exist, unique start numbers, NumPlayers consistency                                                          |
| Pairing Engine | `ValidatePairingEngine` (1) | General + XXR/142 (total rounds), XXC/152 (initial color), 092 (tournament type), opponent symmetry, color consistency, result consistency |
| FIDE           | `ValidateFIDE` (2)          | Pairing Engine + tournament name, dates, time control, per-player name/federation/FIDE ID                                                  |

### ValidationIssue

```go
type ValidationIssue struct {
    Field    string   // e.g. "012", "XXR", "player.3.fideID"
    Severity Severity // SeverityError or SeverityWarning
    Message  string   // Human-readable description
}
```

- **SeverityError**: must fix for the profile to be satisfied.
- **SeverityWarning**: advisory, not blocking.

Example:

```go
issues := doc.Validate(trf.ValidatePairingEngine)
for _, issue := range issues {
    fmt.Printf("[%s] %s: %s\n", issue.Severity, issue.Field, issue.Message)
}
```

## System-specific extended fields

TRF16 uses XX-prefixed line codes for engine-specific configuration. These are stored on the Document and round-trip through `ToTournamentState()` / `FromTournamentState()`.

### Common fields

| Field            | Line code | Type              | Description                                             |
| ---------------- | --------- | ----------------- | ------------------------------------------------------- |
| `TotalRounds`    | XXR       | `int`             | Total number of rounds in the tournament                |
| `InitialColor`   | XXC       | `string`          | Initial color assignment for top seed (e.g. `"white1"`) |
| `Acceleration`   | XXS       | `[]string`        | Baku acceleration data lines                            |
| `ForbiddenPairs` | XXP       | `[]ForbiddenPair` | Pairs of start numbers that must not be paired          |

### Round-Robin fields

| Field          | Line code | Type    | Description                                           |
| -------------- | --------- | ------- | ----------------------------------------------------- |
| `Cycles`       | XXY       | `int`   | Number of cycles (1 = single, 2 = double round-robin) |
| `ColorBalance` | XXB       | `*bool` | Enable FIDE color balancing                           |

### Lim field

| Field            | Line code | Type    | Description              |
| ---------------- | --------- | ------- | ------------------------ |
| `MaxiTournament` | XXM       | `*bool` | Lim maxi-tournament mode |

### Team Swiss fields

| Field                 | Line code | Type     | Description                                      |
| --------------------- | --------- | -------- | ------------------------------------------------ |
| `ColorPreferenceType` | XXT       | `string` | Color preference type: `"A"`, `"B"`, or `"none"` |
| `PrimaryScore`        | XXG       | `string` | Primary score type: `"match"` or `"game"`        |

### Keizer fields

| Field                     | Line code | Type    | Description                              |
| ------------------------- | --------- | ------- | ---------------------------------------- |
| `AllowRepeatPairings`     | XXA       | `*bool` | Whether repeat pairings are permitted    |
| `MinRoundsBetweenRepeats` | XXK       | `int`   | Minimum rounds between repeated pairings |

### TRF-2026 equivalents

TRF-2026 introduces new line codes that replace some XX fields. The Document stores both, and the `Effective*` methods provide unified access:

- **142** (`TotalRounds26`) replaces XXR -- use `EffectiveTotalRounds()`.
- **152** (`InitialColor26`) replaces XXC -- use `EffectiveInitialColor()`.
- **250** (`Accelerations26`) replaces XXS.
- **260** (`ForbiddenPairs26`) replaces XXP.
