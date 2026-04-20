---
title: "TieBreaker Interface"
linkTitle: "TieBreaker"
weight: 5
description: "The TieBreaker interface, the self-registering registry, and how to compute tiebreak values."
---

The `TieBreaker` interface computes a single numeric tiebreak value for each player. The `tiebreaker` package provides 25 implementations and a self-registering registry for lookup by ID.

## Interface definition

```go
type TieBreaker interface {
    ID() string
    Name() string
    Compute(ctx context.Context, state *TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
```

Three methods:

- **`ID`** -- Short machine identifier (e.g., `"buchholz-cut1"`). Used in configuration and the registry.
- **`Name`** -- Human-readable display name (e.g., `"Buchholz Cut-1"`).
- **`Compute`** -- Takes the tournament state and current scores (from a `Scorer`), returns one `TieBreakValue` per player. The scores slice is needed because many tiebreakers depend on opponents' scores.

All engines accept `context.Context` for forward compatibility. Since all computation is CPU-bound and in-memory, the context is not currently checked for cancellation.

### TieBreakValue

```go
type TieBreakValue struct {
    PlayerID string
    Value    float64
}
```

The returned slice contains one entry per player in `scores`, in the same order.

## Registry

Tiebreakers self-register via `init()` functions. The `tiebreaker` package exposes three registry functions:

```go
import "github.com/gnutterts/chesspairing/tiebreaker"

// Get a tiebreaker by ID.
tb, err := tiebreaker.Get("buchholz-cut1")

// List all registered IDs (unsorted).
ids := tiebreaker.All() // returns []string

// Register a custom tiebreaker (call in init() only).
tiebreaker.Register("my-tb", func() chesspairing.TieBreaker {
    return &myTieBreaker{}
})
```

All writes to the registry happen during `init()`. After initialization completes, the registry is read-only and safe for concurrent access without synchronization.

## All 25 registered tiebreakers

| ID                      | Name                    | Description                                                     |
| ----------------------- | ----------------------- | --------------------------------------------------------------- |
| `buchholz`              | Buchholz                | Sum of all opponents' scores                                    |
| `buchholz-cut1`         | Buchholz Cut-1          | Drop lowest opponent score                                      |
| `buchholz-cut2`         | Buchholz Cut-2          | Drop two lowest opponent scores                                 |
| `buchholz-median`       | Buchholz Median         | Drop highest and lowest opponent scores                         |
| `buchholz-median2`      | Buchholz Median-2       | Drop two highest and two lowest opponent scores                 |
| `sonneborn-berger`      | Sonneborn-Berger        | Sum of opponents' scores weighted by result against each        |
| `direct-encounter`      | Direct Encounter        | Head-to-head score among tied players                           |
| `wins`                  | Games Won (OTB)         | OTB wins only, excludes forfeit wins                            |
| `win`                   | Rounds Won              | OTB wins + forfeit wins + PAB                                   |
| `black-games`           | Games with Black        | Number of games played as Black, excludes forfeits              |
| `black-wins`            | Black Wins              | OTB wins with Black pieces                                      |
| `rounds-played`         | Rounds Played           | Total rounds where the player participated                      |
| `standard-points`       | Standard Points         | Score using 1-0.5-0 regardless of the tournament scoring system |
| `pairing-number`        | Pairing Number          | Tournament pairing number (TPN, lower is better)                |
| `koya`                  | Koya System             | Score against opponents with >= 50% score                       |
| `progressive`           | Progressive Score       | Cumulative round-by-round score                                 |
| `aro`                   | Avg Rating of Opponents | Average rating of opponents                                     |
| `fore-buchholz`         | Fore Buchholz           | Buchholz treating pending games as draws                        |
| `avg-opponent-buchholz` | Avg Opponent Buchholz   | Average of opponents' Buchholz scores                           |
| `performance-rating`    | Performance Rating      | Tournament Performance Rating (TPR)                             |
| `performance-points`    | Performance Points      | Tournament Performance Points (PTP)                             |
| `avg-opponent-tpr`      | Avg Opponent TPR        | Average of opponents' TPR (APRO)                                |
| `avg-opponent-ptp`      | Avg Opponent PTP        | Average of opponents' PTP (APPO)                                |
| `player-rating`         | Player Rating           | Player's own rating (RTNG)                                      |
| `games-played`          | Games Played            | Total games played (excluding forfeits)                         |

## Forfeit exclusion

All opponent-based tiebreakers use the shared `buildOpponentData` function, which excludes all forfeited games (single and double forfeits) from the opponent list. This means:

- Forfeit wins/losses do not contribute to Buchholz, Sonneborn-Berger, or any opponent-score-based calculation.
- Pending games are also excluded.
- Only OTB results (`ResultWhiteWins`, `ResultBlackWins`, `ResultDraw`) are counted.

This ensures forfeits do not distort tiebreak calculations.

## DefaultTiebreakers

The root package provides FIDE-recommended tiebreaker sequences per pairing system:

```go
import "github.com/gnutterts/chesspairing"

tbs := chesspairing.DefaultTiebreakers(chesspairing.PairingDutch)
// Returns: ["buchholz-cut1", "buchholz", "sonneborn-berger", "direct-encounter"]
```

| Pairing system                                  | Default tiebreakers                                                 |
| ----------------------------------------------- | ------------------------------------------------------------------- |
| Dutch, Burstein, Dubov, Lim, Double-Swiss, Team | `buchholz-cut1`, `buchholz`, `sonneborn-berger`, `direct-encounter` |
| Round-Robin                                     | `sonneborn-berger`, `direct-encounter`, `wins`, `koya`              |
| Keizer                                          | `games-played`, `direct-encounter`, `wins`                          |

See [Tiebreakers](/docs/tiebreakers/) for detailed explanations of each algorithm.

## Usage example

Score the tournament, then compute tiebreakers in sequence to build final standings:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/scoring/standard"
    "github.com/gnutterts/chesspairing/tiebreaker"
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
        CurrentRound:  2,
        PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDutch},
    }

    // Step 1: Compute scores.
    scorer := standard.NewFromMap(nil)
    scores, err := scorer.Score(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }

    // Step 2: Compute tiebreakers in FIDE-recommended order.
    tbIDs := chesspairing.DefaultTiebreakers(state.PairingConfig.System)
    tbResults := make(map[string][]chesspairing.TieBreakValue, len(tbIDs))

    for _, id := range tbIDs {
        tb, err := tiebreaker.Get(id)
        if err != nil {
            log.Fatal(err)
        }
        values, err := tb.Compute(context.Background(), state, scores)
        if err != nil {
            log.Fatal(err)
        }
        tbResults[id] = values
    }

    // Step 3: Display standings with tiebreak values.
    for _, ps := range scores {
        fmt.Printf("Rank %d: %s (%.1f pts)", ps.Rank, ps.PlayerID, ps.Score)
        for _, id := range tbIDs {
            for _, tv := range tbResults[id] {
                if tv.PlayerID == ps.PlayerID {
                    fmt.Printf("  %s=%.2f", id, tv.Value)
                    break
                }
            }
        }
        fmt.Println()
    }
}
```

## Data flow

```text
Scorer.Score(ctx, state) returns []PlayerScore
  -> pass scores to each TieBreaker.Compute(ctx, state, scores)
     -> returns []TieBreakValue (one per player)
  -> combine into []Standing for final ranked output
```

The `Standing` type combines scores and tiebreakers into a single ranked output:

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
