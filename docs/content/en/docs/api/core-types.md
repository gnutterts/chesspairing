---
title: "Core Types"
linkTitle: "Core Types"
weight: 2
description: "TournamentState, PlayerEntry, RoundData, GameData, and other foundational types."
---

All core types are defined in the root `chesspairing` package (`result.go`). They form the shared data model that all pairing, scoring, and tiebreaker engines operate on.

## GameResult

`GameResult` is a `string` type representing the outcome of a chess game.

### Constants

| Constant                 | Value       | Meaning               |
| ------------------------ | ----------- | --------------------- |
| `ResultWhiteWins`        | `"1-0"`     | White wins            |
| `ResultBlackWins`        | `"0-1"`     | Black wins            |
| `ResultDraw`             | `"0.5-0.5"` | Draw                  |
| `ResultPending`          | `"*"`       | Not yet played        |
| `ResultForfeitWhiteWins` | `"1-0f"`    | White wins by forfeit |
| `ResultForfeitBlackWins` | `"0-1f"`    | Black wins by forfeit |
| `ResultDoubleForfeit`    | `"0-0f"`    | Both forfeit          |

### Methods

| Method              | Returns | Description                                                                                                                                                                           |
| ------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `IsValid()`         | `bool`  | True if the value is one of the 7 recognized constants.                                                                                                                               |
| `IsRecordable()`    | `bool`  | True if the result can be recorded by a user. All valid results except `ResultPending` are recordable. `ResultPending` is the initial state set by the system when a game is created. |
| `IsForfeit()`       | `bool`  | True if the result is any forfeit (single or double).                                                                                                                                 |
| `IsDoubleForfeit()` | `bool`  | True only for `ResultDoubleForfeit`.                                                                                                                                                  |

### Forfeit semantics

**Single forfeit** (`ResultForfeitWhiteWins`, `ResultForfeitBlackWins`): The winner receives points. The game is excluded from pairing history, meaning the two players can be paired again in a later round as if they had never played.

**Double forfeit** (`ResultDoubleForfeit`): The game is excluded from both scoring and pairing. Neither player receives points, and the game is treated as if it never happened.

## ByeType

`ByeType` is an `int` type (iota-based) classifying how a bye is scored.

### Constants

| Constant            | Value | Description                                        |
| ------------------- | ----- | -------------------------------------------------- |
| `ByePAB`            | `0`   | Pairing-Allocated Bye. Full point. TRF code `"F"`. |
| `ByeHalf`           | `1`   | Half-point bye. TRF code `"H"`.                    |
| `ByeZero`           | `2`   | Zero-point bye. TRF code `"Z"`.                    |
| `ByeAbsent`         | `3`   | Absent/unpaired, unexcused. TRF code `"U"`.        |
| `ByeExcused`        | `4`   | Excused absence (notified in advance).             |
| `ByeClubCommitment` | `5`   | Club commitment (absent for interclub team duty).  |

### Methods

| Method      | Returns  | Description                                                                                                    |
| ----------- | -------- | -------------------------------------------------------------------------------------------------------------- |
| `IsValid()` | `bool`   | True if the value is in the range `ByePAB` through `ByeClubCommitment`.                                        |
| `String()`  | `string` | Human-readable name: `"PAB"`, `"Half"`, `"Zero"`, `"Absent"`, `"Excused"`, `"ClubCommitment"`, or `"Unknown"`. |

## TournamentState

The central data structure. All engines receive a pointer to `TournamentState` and treat it as read-only.

```go
type TournamentState struct {
    Players         []PlayerEntry
    Rounds          []RoundData
    CurrentRound    int
    PreAssignedByes []ByeEntry
    PairingConfig   PairingConfig
    ScoringConfig   ScoringConfig
    Info            TournamentInfo
}
```

| Field             | Type             | Description                                                                                            |
| ----------------- | ---------------- | ------------------------------------------------------------------------------------------------------ |
| `Players`         | `[]PlayerEntry`  | All players registered in the tournament.                                                              |
| `Rounds`          | `[]RoundData`    | Completed rounds with game results and byes.                                                           |
| `CurrentRound`    | `int`            | The next round to be paired (1-based).                                                                 |
| `PreAssignedByes` | `[]ByeEntry`     | Byes locked in for the upcoming round. Pairers exclude these players from matching and pass the entries through to `PairingResult.Byes` unchanged. The PAB-uniqueness rule applies only to algorithmically allocated byes. The roundrobin pairer rejects non-empty values because the Berger schedule is fixed. |
| `PairingConfig`   | `PairingConfig`  | Pairing system selection and engine-specific options.                                                  |
| `ScoringConfig`   | `ScoringConfig`  | Scoring system selection, tiebreaker list, and scoring options.                                        |
| `Info`            | `TournamentInfo` | Tournament metadata. Zero value if not set. Engines ignore this field.                                 |

### Validate()

```go
func (s *TournamentState) Validate() error
```

Checks structural invariants and returns an error describing the first problem found, or `nil` if valid. The checks are:

- At least one player exists.
- No player has an empty `ID`.
- No duplicate player IDs.
- `CurrentRound` does not exceed `len(Rounds)`.
- Every entry in `PreAssignedByes` references a known player, has a unique player ID, and carries a valid `ByeType`.
- Every `WithdrawnAfterRound` value is positive and does not exceed `CurrentRound`.

The state also exposes two helpers that respect joins and withdrawals:

- `IsActiveInRound(playerID, round)` reports whether a player is participating in a given 1-indexed round. `round <= 0` skips the round filter and just rejects withdrawn players.
- `ActivePlayerIDs(round)` returns the IDs of players active in the given round, in the order they appear in `Players`.

## PlayerEntry

Represents a single player for engine purposes.

```go
type PlayerEntry struct {
    ID                  string
    DisplayName         string
    Rating              int
    Federation          string
    FideID              string
    Title               string
    Sex                 string
    BirthDate           string
    JoinedRound         int
    WithdrawnAfterRound *int
}
```

| Field                 | Type     | Description                                                                                                                                                                                                                |
| --------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ID`                  | `string` | Unique player identifier. Must not be empty.                                                                                                                                                                               |
| `DisplayName`         | `string` | Player name for display purposes.                                                                                                                                                                                          |
| `Rating`              | `int`    | Player rating (e.g. FIDE Elo). Used for seeding and tiebreakers.                                                                                                                                                           |
| `Federation`          | `string` | FIDE federation code (e.g. `"NED"`, `"USA"`, `"IND"`). Empty if unknown. Used by Varma assignment for round-robin.                                                                                                         |
| `FideID`              | `string` | FIDE player ID number. Empty if unknown.                                                                                                                                                                                   |
| `Title`               | `string` | FIDE title code (`"GM"`, `"IM"`, `"FM"`, `"WGM"`, `"WIM"`, `"WFM"`, `"CM"`, `"WCM"`). Empty if untitled.                                                                                                                   |
| `Sex`                 | `string` | `"m"` or `"w"`. Empty if unknown.                                                                                                                                                                                          |
| `BirthDate`           | `string` | Birth date as `YYYY/MM/DD`. Empty if unknown.                                                                                                                                                                              |
| `JoinedRound`         | `int`    | Round number when the player joined. 0 or 1 means original player (joined from the start). Used by Keizer scoring for late-joiner handicap calculation.                                                                    |
| `WithdrawnAfterRound` | `*int`   | Last round in which the player participated. `nil` means the player is still active. From `*WithdrawnAfterRound + 1` onward the player is excluded from pairing and from scoring. Use `IsActiveInRound` rather than reading this field directly. Per-round skips that do not amount to a withdrawal are expressed as a pre-assigned `ByeAbsent` or `ByeExcused` instead. |

## RoundData

Contains all games and byes for a completed round.

```go
type RoundData struct {
    Number int
    Games  []GameData
    Byes   []ByeEntry
}
```

| Field    | Type         | Description                      |
| -------- | ------------ | -------------------------------- |
| `Number` | `int`        | Round number, 1-based.           |
| `Games`  | `[]GameData` | All games played in this round.  |
| `Byes`   | `[]ByeEntry` | All byes assigned in this round. |

## GameData

A single game result for engine consumption.

```go
type GameData struct {
    WhiteID   string
    BlackID   string
    Result    GameResult
    IsForfeit bool
}
```

| Field       | Type         | Description                                                                                                         |
| ----------- | ------------ | ------------------------------------------------------------------------------------------------------------------- |
| `WhiteID`   | `string`     | Player ID of the white player.                                                                                      |
| `BlackID`   | `string`     | Player ID of the black player.                                                                                      |
| `Result`    | `GameResult` | The game outcome.                                                                                                   |
| `IsForfeit` | `bool`       | Redundant with `Result.IsForfeit()` but provided for convenience so callers do not need to check the result string. |

## ByeEntry

Records a bye assignment with its type.

```go
type ByeEntry struct {
    PlayerID string
    Type     ByeType
}
```

| Field      | Type      | Description                      |
| ---------- | --------- | -------------------------------- |
| `PlayerID` | `string`  | The player who received the bye. |
| `Type`     | `ByeType` | How the bye is scored.           |

## ResultContext

Provides additional context to `Scorer.PointsForResult()` when calculating points for a specific game result.

```go
type ResultContext struct {
    OpponentRank        int
    OpponentValueNumber int
    PlayerRank          int
    PlayerValueNumber   int
    ByeType             *ByeType
}
```

| Field                 | Type       | Description                                                                                                                |
| --------------------- | ---------- | -------------------------------------------------------------------------------------------------------------------------- |
| `OpponentRank`        | `int`      | Opponent's current rank (1-based).                                                                                         |
| `OpponentValueNumber` | `int`      | Opponent's Keizer value number (rank-derived).                                                                             |
| `PlayerRank`          | `int`      | Current player's rank.                                                                                                     |
| `PlayerValueNumber`   | `int`      | Current player's Keizer value number.                                                                                      |
| `ByeType`             | `*ByeType` | When non-nil, indicates the entry is a bye of the given type rather than a played game. Scorers ignore the `Result` field. |

Forfeit detection is no longer a flag on `ResultContext`. Callers derive it from `Result.IsForfeit()` directly. This struct is primarily used by the Keizer scoring system, where point values depend on the opponent's rank and value number. Standard and football scoring ignore the rank/value fields and dispatch on `ByeType` and `Result`.

## PairingResult

Output of `Pairer.Pair()`. Contains the board assignments and any byes for the round.

```go
type PairingResult struct {
    Pairings []GamePairing
    Byes     []ByeEntry
    Notes    []string
}
```

| Field      | Type            | Description                                                                |
| ---------- | --------------- | -------------------------------------------------------------------------- |
| `Pairings` | `[]GamePairing` | Board assignments for the round.                                           |
| `Byes`     | `[]ByeEntry`    | Byes assigned by the engine (typically at most one PAB for Swiss systems). |
| `Notes`    | `[]string`      | Engine diagnostic messages (e.g. criteria relaxation warnings).            |

## GamePairing

A single board assignment within a `PairingResult`.

```go
type GamePairing struct {
    Board   int
    WhiteID string
    BlackID string
}
```

| Field     | Type     | Description                                        |
| --------- | -------- | -------------------------------------------------- |
| `Board`   | `int`    | Board number, 1-indexed. Board 1 is the top board. |
| `WhiteID` | `string` | Player ID assigned the white pieces.               |
| `BlackID` | `string` | Player ID assigned the black pieces.               |

## PlayerScore

Output of `Scorer.Score()`. One entry per player.

```go
type PlayerScore struct {
    PlayerID string
    Score    float64
    Rank     int
}
```

| Field      | Type      | Description                                               |
| ---------- | --------- | --------------------------------------------------------- |
| `PlayerID` | `string`  | The player's unique identifier.                           |
| `Score`    | `float64` | The player's total score under the active scoring system. |
| `Rank`     | `int`     | The player's rank by score (1-based, ties possible).      |

## TieBreakValue

Output of `TieBreaker.Compute()`. One entry per player.

```go
type TieBreakValue struct {
    PlayerID string
    Value    float64
}
```

| Field      | Type      | Description                                  |
| ---------- | --------- | -------------------------------------------- |
| `PlayerID` | `string`  | The player's unique identifier.              |
| `Value`    | `float64` | The computed tiebreak value for this player. |

## Standing

Final ranked output combining score and tiebreakers. All fields have JSON struct tags for serialization.

```go
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
```

| Field         | JSON Key      | Type           | Description                         |
| ------------- | ------------- | -------------- | ----------------------------------- |
| `Rank`        | `rank`        | `int`          | Final rank (1-based).               |
| `PlayerID`    | `playerId`    | `string`       | Player identifier.                  |
| `DisplayName` | `displayName` | `string`       | Player name.                        |
| `Score`       | `score`       | `float64`      | Total score.                        |
| `TieBreakers` | `tieBreakers` | `[]NamedValue` | Ordered tiebreak values.            |
| `GamesPlayed` | `gamesPlayed` | `int`          | Total games played (excludes byes). |
| `Wins`        | `wins`        | `int`          | Total wins.                         |
| `Draws`       | `draws`       | `int`          | Total draws.                        |
| `Losses`      | `losses`      | `int`          | Total losses.                       |

## NamedValue

Pairs a tiebreaker identifier with its computed value. Used within `Standing.TieBreakers`.

```go
type NamedValue struct {
    ID    string  `json:"id"`
    Name  string  `json:"name"`
    Value float64 `json:"value"`
}
```

| Field   | JSON Key | Type      | Description                                               |
| ------- | -------- | --------- | --------------------------------------------------------- |
| `ID`    | `id`     | `string`  | Tiebreaker registry identifier (e.g. `"buchholz-cut1"`).  |
| `Name`  | `name`   | `string`  | Human-readable tiebreaker name (e.g. `"Buchholz Cut 1"`). |
| `Value` | `value`  | `float64` | Computed tiebreak value.                                  |

## TournamentInfo

Metadata struct for display and TRF round-trip fidelity. Engines ignore this struct entirely. It is populated from TRF header lines during parsing and written back when serializing to TRF.

```go
type TournamentInfo struct {
    Name          string
    City          string
    Federation    string
    StartDate     string
    EndDate       string
    ChiefArbiter  string
    DeputyArbiter string
    TimeControl   string
    RoundDates    []string
}
```

| Field           | Type       | Description                        |
| --------------- | ---------- | ---------------------------------- |
| `Name`          | `string`   | Tournament name.                   |
| `City`          | `string`   | City where the tournament is held. |
| `Federation`    | `string`   | Organizing federation code.        |
| `StartDate`     | `string`   | Start date as `YYYY/MM/DD`.        |
| `EndDate`       | `string`   | End date as `YYYY/MM/DD`.          |
| `ChiefArbiter`  | `string`   | Chief arbiter name.                |
| `DeputyArbiter` | `string`   | Deputy arbiter name.               |
| `TimeControl`   | `string`   | Allotted time description.         |
| `RoundDates`    | `[]string` | Per-round dates as `YYYY/MM/DD`.   |

## Config Types

### PairingConfig

```go
type PairingConfig struct {
    System  PairingSystem
    Options map[string]any
}
```

Selects the pairing algorithm and passes engine-specific options. The `Options` map is parsed by each engine's `ParseOptions()` function. See [Options Pattern](../options/) for details.

### ScoringConfig

```go
type ScoringConfig struct {
    System      ScoringSystem
    Tiebreakers []string
    Options     map[string]any
}
```

Selects the scoring algorithm, specifies the ordered tiebreaker list, and passes scoring options. Tiebreaker IDs correspond to the [tiebreaker registry](../tiebreaker/).
