---
title: "Go Library Quickstart"
linkTitle: "Go Quickstart"
weight: 3
description: "Add chesspairing to your Go project and generate your first pairing programmatically."
---

This guide walks you through adding chesspairing to a Go project, building a
tournament state, generating pairings, scoring rounds, and computing
tiebreakers. By the end you will have a working program that pairs a
four-player Swiss tournament.

## Prerequisites

- Go 1.24 or later

chesspairing is a pure Go module with zero external dependencies.

## Install

Add the module to your project:

```bash
go get github.com/gnutterts/chesspairing
```

Then import the root package and the engine packages you need:

```go
import (
    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/pairing/dutch"
    "github.com/gnutterts/chesspairing/scoring/standard"
    "github.com/gnutterts/chesspairing/tiebreaker"
)
```

## Core interfaces

All engines implement one of three interfaces defined in the root package:

```go
// Pairer generates pairings for a round given tournament state.
type Pairer interface {
    Pair(ctx context.Context, state *TournamentState) (*PairingResult, error)
}

// Scorer calculates standings from game results.
type Scorer interface {
    Score(ctx context.Context, state *TournamentState) ([]PlayerScore, error)
    PointsForResult(result GameResult, rctx ResultContext) float64
}

// TieBreaker computes a single tiebreak value for each player.
type TieBreaker interface {
    ID() string
    Name() string
    Compute(ctx context.Context, state *TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
```

Pairing and scoring are independent -- a tournament can use any combination
(for example, Swiss pairing with Keizer scoring).

## Build a TournamentState

Every engine method takes a `*TournamentState`. This is a read-only snapshot
that you construct from your own data source:

```go
state := &chesspairing.TournamentState{
    Players: []chesspairing.PlayerEntry{
        {ID: "1", DisplayName: "Alice",   Rating: 2100},
        {ID: "2", DisplayName: "Bob",     Rating: 1950},
        {ID: "3", DisplayName: "Charlie", Rating: 1800},
        {ID: "4", DisplayName: "Diana",   Rating: 1750},
    },
    Rounds:       nil, // no rounds played yet
    CurrentRound: 0,
    PairingConfig: chesspairing.PairingConfig{
        System: chesspairing.PairingDutch,
    },
    ScoringConfig: chesspairing.ScoringConfig{
        System:      chesspairing.ScoringStandard,
        Tiebreakers: []string{"buchholz-cut1", "buchholz", "sonneborn-berger"},
    },
}
```

`PlayerEntry` supports additional optional fields such as `Federation`,
`FideID`, `Title`, `Sex`, and `BirthDate`. Set only what you have. To
mark a player as permanently withdrawn after some round `N`, set
`WithdrawnAfterRound = &N`; the player is then excluded from pairing
in every round greater than `N`. Use `state.IsActiveInRound(playerID,
round)` and `state.ActivePlayerIDs(round)` to query the active set.

`Rounds` contains a `[]RoundData` with completed game results. For a brand-new
tournament this is nil or empty.

## Generate pairings

Create a pairer and call `Pair`:

```go
pairer := dutch.New(dutch.Options{})
result, err := pairer.Pair(context.Background(), state)
if err != nil {
    log.Fatal(err)
}

for _, p := range result.Pairings {
    fmt.Printf("Board %d: %s (White) vs %s (Black)\n", p.Board, p.WhiteID, p.BlackID)
}
for _, bye := range result.Byes {
    fmt.Printf("Bye: %s (%s)\n", bye.PlayerID, bye.Type)
}
```

`PairingResult` contains `Pairings` (a slice of `GamePairing` with `Board`,
`WhiteID`, `BlackID`), `Byes` (a slice of `ByeEntry`), and optional `Notes`.

### Available pairers

| System       | Package               | FIDE regulation |
| ------------ | --------------------- | --------------- |
| Dutch        | `pairing/dutch`       | C.04.3          |
| Burstein     | `pairing/burstein`    | C.04.4.2        |
| Dubov        | `pairing/dubov`       | C.04.4.1        |
| Lim          | `pairing/lim`         | C.04.4.3        |
| Double-Swiss | `pairing/doubleswiss` | C.04.5          |
| Team Swiss   | `pairing/team`        | C.04.6          |
| Keizer       | `pairing/keizer`      | --              |
| Round-robin  | `pairing/roundrobin`  | C.05 Annex 1    |

All pairers follow the same pattern: `New(Options{})` or `NewFromMap(map[string]any)`.

## Score the tournament

After recording game results, score the round:

```go
scorer := standard.New(standard.Options{})

// Add round 1 results to the state.
state.Rounds = []chesspairing.RoundData{
    {
        Number: 1,
        Games: []chesspairing.GameData{
            {WhiteID: "1", BlackID: "4", Result: chesspairing.ResultWhiteWins},
            {WhiteID: "2", BlackID: "3", Result: chesspairing.ResultDraw},
        },
    },
}
state.CurrentRound = 1

scores, err := scorer.Score(context.Background(), state)
if err != nil {
    log.Fatal(err)
}
for _, s := range scores {
    fmt.Printf("Player %s: %.1f pts (rank %d)\n", s.PlayerID, s.Score, s.Rank)
}
```

`Score` returns `[]PlayerScore` sorted by rank. Each entry has `PlayerID`,
`Score`, and `Rank`.

### Available scorers

| System   | Package            | Default points          |
| -------- | ------------------ | ----------------------- |
| Standard | `scoring/standard` | 1 - 0.5 - 0             |
| Football | `scoring/football` | 3 - 1 - 0               |
| Keizer   | `scoring/keizer`   | Iterative ranking-based |

## Compute tiebreakers

Tiebreakers are looked up from a global registry by ID:

```go
tb, err := tiebreaker.Get("buchholz-cut1")
if err != nil {
    log.Fatal(err)
}

values, err := tb.Compute(context.Background(), state, scores)
if err != nil {
    log.Fatal(err)
}
for _, v := range values {
    fmt.Printf("Player %s: Buchholz Cut 1 = %.1f\n", v.PlayerID, v.Value)
}
```

There are 25 registered tiebreakers. Common IDs include `buchholz`,
`buchholz-cut1`, `sonneborn-berger`, `direct-encounter`, `wins`, `aro`,
`performance-rating`, and `koya`. Use `tiebreaker.All()` to list them all.

## Engine options

Each engine has an `Options` struct with pointer fields. A nil field means
"use the default." Pass an empty struct for default behavior:

```go
// All defaults:
pairer := dutch.New(dutch.Options{})

// Override one setting:
accel := "baku"
pairer := dutch.New(dutch.Options{
    Acceleration: &accel,
})
```

For dynamic configuration (JSON, config files), use `NewFromMap`:

```go
pairer := dutch.NewFromMap(map[string]any{
    "acceleration": "baku",
    "topSeedColor": "white",
})
```

## Complete example

The following program creates a four-player tournament, generates round 1
pairings, records results, scores the round, and computes a tiebreaker:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/pairing/dutch"
    "github.com/gnutterts/chesspairing/scoring/standard"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

func main() {
    ctx := context.Background()

    // 1. Define players and build tournament state.
    state := &chesspairing.TournamentState{
        Players: []chesspairing.PlayerEntry{
            {ID: "1", DisplayName: "Alice",   Rating: 2100},
            {ID: "2", DisplayName: "Bob",     Rating: 1950},
            {ID: "3", DisplayName: "Charlie", Rating: 1800},
            {ID: "4", DisplayName: "Diana",   Rating: 1750},
        },
        CurrentRound: 0,
        PairingConfig: chesspairing.PairingConfig{
            System: chesspairing.PairingDutch,
        },
        ScoringConfig: chesspairing.ScoringConfig{
            System: chesspairing.ScoringStandard,
        },
    }

    // 2. Generate round 1 pairings.
    pairer := dutch.New(dutch.Options{})
    result, err := pairer.Pair(ctx, state)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Round 1 Pairings:")
    for _, p := range result.Pairings {
        fmt.Printf("  Board %d: %s (W) vs %s (B)\n", p.Board, p.WhiteID, p.BlackID)
    }

    // 3. Record results and add round to state.
    //    (In a real application, results come from user input.)
    state.Rounds = []chesspairing.RoundData{
        {
            Number: 1,
            Games:  toGameData(result.Pairings),
            Byes:   result.Byes,
        },
    }
    state.CurrentRound = 1

    // 4. Score the round.
    scorer := standard.New(standard.Options{})
    scores, err := scorer.Score(ctx, state)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("\nStandings after Round 1:")
    for _, s := range scores {
        fmt.Printf("  %d. Player %s — %.1f pts\n", s.Rank, s.PlayerID, s.Score)
    }

    // 5. Compute a tiebreaker.
    tb, err := tiebreaker.Get("buchholz-cut1")
    if err != nil {
        log.Fatal(err)
    }
    values, err := tb.Compute(ctx, state, scores)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("\nBuchholz Cut 1:")
    for _, v := range values {
        fmt.Printf("  Player %s: %.1f\n", v.PlayerID, v.Value)
    }
}

// toGameData converts pairings into game results for demonstration purposes.
// The first board is a white win, the rest are draws.
func toGameData(pairings []chesspairing.GamePairing) []chesspairing.GameData {
    games := make([]chesspairing.GameData, len(pairings))
    for i, p := range pairings {
        result := chesspairing.ResultDraw
        if i == 0 {
            result = chesspairing.ResultWhiteWins
        }
        games[i] = chesspairing.GameData{
            WhiteID: p.WhiteID,
            BlackID: p.BlackID,
            Result:  result,
        }
    }
    return games
}
```

## Next steps

- [API Reference](/docs/api/) -- full documentation of all types, interfaces,
  and engine options
- [CLI Quickstart](../cli-quickstart/) -- use chesspairing from the command
  line with TRF16 files
- Browse the available [pairing systems](/docs/pairing-systems/) and
  [scoring systems](/docs/scoring/) to find the right configuration
  for your tournament
