---
title: "Scorer Interface"
linkTitle: "Scorer"
weight: 4
description: "The Scorer interface and the three scoring implementations."
---

The `Scorer` interface calculates standings from game results. Three implementations cover the major scoring systems: Standard (1-0.5-0), Keizer (iterative rank-dependent), and Football (3-1-0).

## Interface definition

```go
type Scorer interface {
    Score(ctx context.Context, state *TournamentState) ([]PlayerScore, error)
    PointsForResult(result GameResult, rctx ResultContext) float64
}
```

Two methods:

- **`Score`** -- Computes scores for all active players in the tournament. Returns one `PlayerScore` per active player, each containing `PlayerID`, `Score`, and `Rank`. Players are ranked by score descending, with rating as the secondary tiebreak.
- **`PointsForResult`** -- Returns the point value for a single game result given its context. Pairers call this internally when they need scoring information (e.g., the Keizer pairer needs Keizer scores for ranking). The `ResultContext` provides opponent/player rank information needed by Keizer scoring; Standard and Football ignore it.

All engines accept `context.Context` for forward compatibility. Since all computation is CPU-bound and in-memory, the context is not currently checked for cancellation.

### ResultContext

```go
type ResultContext struct {
    OpponentRank        int
    OpponentValueNumber int
    PlayerRank          int
    PlayerValueNumber   int
    ByeType             *ByeType
}
```

When `ByeType` is non-nil the entry is a bye of that type and scorers ignore the `Result` field. Otherwise the entry is a played game; forfeits are detected with `Result.IsForfeit()`. The rank and value-number fields are used exclusively by Keizer scoring.

## Implementations

Every scoring engine provides two constructors:

- `New(opts Options) *Scorer` -- typed options struct.
- `NewFromMap(m map[string]any) *Scorer` -- generic map (from JSON config or TRF).

Both apply defaults for any unset (nil) fields. See [Options Pattern](../options/) for the nil-means-default convention.

### Standard

**Package:** `github.com/gnutterts/chesspairing/scoring/standard`

Standard FIDE scoring: fixed points per result, independent of opponent strength. Single pass, deterministic.

**Constructors:**

```go
import "github.com/gnutterts/chesspairing/scoring/standard"

// From a generic options map:
scorer := standard.NewFromMap(nil) // all defaults

// From a typed Options struct:
scorer := standard.New(standard.Options{
    PointWin:  chesspairing.Float64Ptr(1.0),
    PointDraw: chesspairing.Float64Ptr(0.5),
})
```

**Options:**

| Field                  | Type       | Default | Description                                       |
| ---------------------- | ---------- | ------- | ------------------------------------------------- |
| `PointWin`             | `*float64` | `1.0`   | Points for a win                                  |
| `PointDraw`            | `*float64` | `0.5`   | Points for a draw                                 |
| `PointLoss`            | `*float64` | `0.0`   | Points for a loss                                 |
| `PointBye`             | `*float64` | `1.0`   | Points for a pairing-allocated bye                |
| `PointForfeitWin`      | `*float64` | `1.0`   | Points for a forfeit win                          |
| `PointForfeitLoss`     | `*float64` | `0.0`   | Points for a forfeit loss                         |
| `PointAbsent`          | `*float64` | `0.0`   | Points for an unexcused absence (`ByeAbsent`)     |
| `PointExcused`         | `*float64` | `0.0`   | Points for an excused absence (`ByeExcused`)      |
| `PointClubCommitment`  | `*float64` | `0.0`   | Points for a club-commitment absence (`ByeClubCommitment`) |

Each bye type maps to its own option: `ByePAB` to `PointBye`, `ByeHalf` to `PointDraw`, `ByeZero` to `PointLoss`, `ByeAbsent` to `PointAbsent`, `ByeExcused` to `PointExcused`, and `ByeClubCommitment` to `PointClubCommitment`. Double forfeits award zero to both players.

### Keizer

**Package:** `github.com/gnutterts/chesspairing/scoring/keizer`

In Keizer scoring, each player gets a value number based on their current rank. Beating a strong opponent (high value number) earns more points than beating a weaker one. Absences receive a fraction of the player's own value number.

The algorithm is iterative: scores determine rankings, which determine value numbers, which change scores. It converges within 20 iterations using x2 integer arithmetic (doubled scores) to eliminate floating-point drift. A 2-cycle oscillation detector averages alternating rankings when convergence stalls.

**Constructors:**

```go
import "github.com/gnutterts/chesspairing/scoring/keizer"

// From a generic options map:
scorer := keizer.NewFromMap(nil) // KeizerForClubs defaults

// From a typed Options struct:
scorer := keizer.New(keizer.Options{
    WinFraction:  chesspairing.Float64Ptr(1.0),
    DrawFraction: chesspairing.Float64Ptr(0.5),
    SelfVictory:  chesspairing.BoolPtr(true),
})
```

**Key options (25 total):**

| Field                    | Type       | Default      | Description                                                                                                      |
| ------------------------ | ---------- | ------------ | ---------------------------------------------------------------------------------------------------------------- |
| `ValueNumberBase`        | `*int`     | player count | Top-ranked player's value number                                                                                 |
| `ValueNumberStep`        | `*int`     | `1`          | Decrement per rank position                                                                                      |
| `WinFraction`            | `*float64` | `1.0`        | Fraction of opponent's value number for a win                                                                    |
| `DrawFraction`           | `*float64` | `0.5`        | Fraction of opponent's value number for a draw                                                                   |
| `LossFraction`           | `*float64` | `0.0`        | Fraction of opponent's value number for a loss                                                                   |
| `ForfeitWinFraction`     | `*float64` | `1.0`        | Fraction of opponent's value number for a forfeit win                                                            |
| `ForfeitLossFraction`    | `*float64` | `0.0`        | Fraction of opponent's value number for a forfeit loss                                                           |
| `DoubleForfeitFraction`  | `*float64` | `0.0`        | Fraction of opponent's value number for a double forfeit                                                         |
| `ByeValueFraction`       | `*float64` | `0.50`       | Fraction of own value number for a PAB                                                                           |
| `HalfByeFraction`        | `*float64` | `0.50`       | Fraction of own value number for a half-point bye                                                                |
| `ZeroByeFraction`        | `*float64` | `0.0`        | Fraction of own value number for a zero-point bye                                                                |
| `AbsentPenaltyFraction`  | `*float64` | `0.35`       | Fraction of own value number for an unexcused absence                                                            |
| `ExcusedAbsentFraction`  | `*float64` | `0.35`       | Fraction of own value number for an excused absence                                                              |
| `ClubCommitmentFraction` | `*float64` | `0.70`       | Fraction of own value number for interclub duty absence                                                          |
| `SelfVictory`            | `*bool`    | `true`       | Add player's own value number to their total (once, not per round)                                               |
| `AbsenceLimit`           | `*int`     | `5`          | Max absences that score points (0 = unlimited). Club commitments exempt                                          |
| `AbsenceDecay`           | `*bool`    | `false`      | Halve absence bonus for each successive absence                                                                  |
| `Frozen`                 | `*bool`    | `false`      | Disable iterative convergence; score each round with the ranking at the time                                     |
| `LateJoinHandicap`       | `*float64` | `0`          | Fixed score per round missed before joining. Requires `PlayerEntry.JoinedRound`. Exempt from absence limit/decay |

Six fixed-value override fields (`ByeFixedValue`, `HalfByeFixedValue`, `ZeroByeFixedValue`, `AbsentFixedValue`, `ExcusedAbsentFixedValue`, `ClubCommitmentFixedValue`) replace the corresponding fraction calculation with a fixed integer score when non-nil.

**Variant presets:**

- **KeizerForClubs (default):** All nil -- uses the defaults listed above.
- **Classic KNSB sixths:** `WinFraction=1`, `DrawFraction=0.5`, `LossFraction=0`, `ByeValueFraction=4/6`, `AbsentPenaltyFraction=2/6`, `ClubCommitmentFraction=2/3`, `ExcusedAbsentFraction=2/6`, `AbsenceLimit=5`.
- **FreeKeizer:** `LossFraction=1/6`, `ByeValueFraction=4/6`, `AbsentPenaltyFraction=2/6`, `AbsenceLimit=5`.

### Football

**Package:** `github.com/gnutterts/chesspairing/scoring/football`

Thin wrapper around Standard scoring with football defaults: 3 for a win, 1 for a draw, 0 for a loss. Rewards decisive results more heavily than standard scoring.

Football uses `standard.Options` directly -- there is no separate `football.Options` type. All Standard options are available since Football delegates entirely to the Standard scorer internally.

**Constructors:**

```go
import "github.com/gnutterts/chesspairing/scoring/football"

// From a generic options map:
scorer := football.NewFromMap(nil) // football defaults (3-1-0)

// From a typed Options struct (uses standard.Options):
scorer := football.New(standard.Options{
    PointWin: chesspairing.Float64Ptr(3.0),
})
```

**Football defaults (vs. Standard):**

| Field                 | Football | Standard |
| --------------------- | -------- | -------- |
| `PointWin`            | `3.0`    | `1.0`    |
| `PointDraw`           | `1.0`    | `0.5`    |
| `PointLoss`           | `0.0`    | `0.0`    |
| `PointBye`            | `3.0`    | `1.0`    |
| `PointForfeitWin`     | `3.0`    | `1.0`    |
| `PointForfeitLoss`    | `0.0`    | `0.0`    |
| `PointAbsent`         | `0.0`    | `0.0`    |
| `PointExcused`        | `0.0`    | `0.0`    |
| `PointClubCommitment` | `0.0`    | `0.0`    |

## Compile-time interface check

Every scoring package includes a compile-time assertion:

```go
var _ chesspairing.Scorer = (*Scorer)(nil)
```

This ensures the implementation satisfies the `Scorer` interface at compile time.

## Usage example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/scoring/standard"
)

func main() {
    state := &chesspairing.TournamentState{
        Players: []chesspairing.PlayerEntry{
            {ID: "1", DisplayName: "Alice",   Rating: 2400},
            {ID: "2", DisplayName: "Bob",     Rating: 2350},
            {ID: "3", DisplayName: "Charlie", Rating: 2300},
            {ID: "4", DisplayName: "Diana",   Rating: 2250},
        },
        Rounds: []chesspairing.RoundData{
            {
                Number: 1,
                Games: []chesspairing.GameData{
                    {WhiteID: "1", BlackID: "4", Result: chesspairing.ResultWhiteWins},
                    {WhiteID: "2", BlackID: "3", Result: chesspairing.ResultDraw},
                },
            },
        },
        CurrentRound: 2,
    }

    scorer := standard.NewFromMap(nil)

    scores, err := scorer.Score(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }

    for _, ps := range scores {
        fmt.Printf("Rank %d: %s (%.1f pts)\n", ps.Rank, ps.PlayerID, ps.Score)
    }
    // Output:
    // Rank 1: 1 (1.0 pts)
    // Rank 2: 2 (0.5 pts)
    // Rank 3: 3 (0.5 pts)
    // Rank 4: 4 (0.0 pts)
}
```

## Error handling

Scorers return an error for invalid input states. If the tournament has no players, `Score` returns `nil, nil` (not an error).

Scorers never panic. All exceptional conditions are reported through the returned `error` value.

## Data flow

```text
Caller builds TournamentState
  -> Scorer.Score(ctx, state) returns []PlayerScore
     -> PlayerScore.PlayerID: player identifier
     -> PlayerScore.Score: total points
     -> PlayerScore.Rank: 1-based rank position

  -> Scorer.PointsForResult(result, rctx) returns float64
     -> Used internally by pairers for scoregroup formation
```

The `[]PlayerScore` slice is ordered by rank (index 0 = rank 1). Pass this slice to `TieBreaker.Compute()` to compute tiebreak values.
