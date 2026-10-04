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
| `wins`                  | Games Won               | Games won; forfeit wins count in predetermined pairings            |
| `win`                   | Rounds Won              | OTB wins + forfeit wins + PAB                                   |
| `black-games`           | Games with Black        | Number of games played as Black; forfeit wins count in predetermined pairings |
| `black-wins`            | Black Wins              | Wins with Black pieces; forfeit wins count in predetermined pairings |
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

## Unplayed rounds

Tie-breakers build one per-round record for every player. In Swiss tournaments, FIDE C.07:2026 Article 16 classifies unplayed rounds and uses capped dummy opponents; forfeit wins and losses, requested byes, and absences are unplayed, while full-point byes count as played. In round robins and other predetermined pairings, Article 15.2 treats forfeits as regular encounters, except for ratings-based tie-breaks and Type-B forfeit losses. Pending games are not completed encounters.

This means:

- A forfeit win contributes to Buchholz, Sonneborn-Berger, and other opponent-score-based calculations through the Article 16 dummy (Swiss) or the scheduled opponent (predetermined pairings).
- Pending games are not completed encounters.
- Ratings-based tie-breaks only use opponents played over the board, even in predetermined pairings.

## DefaultTiebreakers

The root package provides library default tiebreaker sequences per pairing system. The Chief Organiser chooses the actual sequence under FIDE C.07:2026 Article 4.1; the library defaults are used when no list is configured. The team defaults assume match points are the primary score (Article 13):

```go
import "github.com/gnutterts/chesspairing"

tbs := chesspairing.DefaultTiebreakers(chesspairing.PairingDutch)
// Returns: ["buchholz-cut1", "buchholz", "sonneborn-berger", "direct-encounter"]
```

| Pairing system                                  | Default tiebreakers                                                 |
| ----------------------------------------------- | ------------------------------------------------------------------- |
| Dutch, Burstein, Dubov, Lim, Double-Swiss       | `buchholz-cut1`, `buchholz`, `sonneborn-berger`, `direct-encounter` |
| Team Swiss                                      | `buchholz-mp-cut1`, `buchholz-mp`, `emmsb`, `mpvgp`                 |
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

    // Step 2: Compute tiebreakers in the configured library order.
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
