# chesspairing

Pure Go chess tournament pairing, scoring, and tiebreaking engines.

Go 1.26.1 -- zero external dependencies -- 1310 tests

---

## Overview

`chesspairing` implements chess tournament engines in pure Go:

- **Pairing**: FIDE Dutch Swiss (C.04.3), FIDE Burstein Swiss (C.04.4.2), FIDE Dubov Swiss (C.04.4.1), Keizer, Round-robin
- **Scoring**: Standard (1-half-0), Keizer (iterative convergence), Football (3-1-0)
- **Tiebreakers**: 25 algorithms (Buchholz variants, Sonneborn-Berger, Direct Encounter, Performance Rating, and more)

Design principles:

- Pure Go, no CGO, no external dependencies
- All engines operate on in-memory data structures -- no I/O, database, or network dependencies
- Safe for concurrent use when each goroutine supplies its own `TournamentState`
- Core interfaces (`Pairer`, `Scorer`, `TieBreaker`) allow mixing any pairing system with any scoring system

## Installation

```
go get github.com/gnutterts/chesspairing
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/dutch"
)

func main() {
	// Create players.
	players := []cp.PlayerEntry{
		{ID: "p1", DisplayName: "Magnus", Rating: 2850, Active: true},
		{ID: "p2", DisplayName: "Fabiano", Rating: 2790, Active: true},
		{ID: "p3", DisplayName: "Ding", Rating: 2780, Active: true},
		{ID: "p4", DisplayName: "Ian", Rating: 2760, Active: true},
		{ID: "p5", DisplayName: "Alireza", Rating: 2750, Active: true},
		{ID: "p6", DisplayName: "Hikaru", Rating: 2740, Active: true},
	}

	// Build tournament state (round 1, no prior rounds).
	state := &cp.TournamentState{
		Players:      players,
		CurrentRound: 1,
		PairingConfig: cp.PairingConfig{
			System: cp.PairingDutch,
		},
	}

	// Create pairer and generate pairings.
	pairer := dutch.New(dutch.Options{})
	result, err := pairer.Pair(context.Background(), state)
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range result.Pairings {
		fmt.Printf("Board %d: %s vs %s\n", p.Board, p.WhiteID, p.BlackID)
	}
	for _, bye := range result.Byes {
		fmt.Printf("Bye: %s (%s)\n", bye.PlayerID, bye.Type)
	}
}
```

## Packages

| Package | Import Path | Description |
|---------|-------------|-------------|
| `chesspairing` | `github.com/gnutterts/chesspairing` | Core interfaces (`Pairer`, `Scorer`, `TieBreaker`), types (`TournamentState`, `PlayerEntry`, `RoundData`, `GameData`, `GameResult`, `ByeType`, `ByeEntry`, `TournamentInfo`), config |
| `blossom` | `github.com/gnutterts/chesspairing/algorithm/blossom` | Edmonds' maximum weight matching (`int64` and `*big.Int` variants) |
| `varma` | `github.com/gnutterts/chesspairing/algorithm/varma` | Varma Tables (FIDE C.05 Annex 2) — federation-aware number assignment for round-robin |
| `pairing/swisslib` | `github.com/gnutterts/chesspairing/pairing/swisslib` | Shared Swiss pairing library: player state, brackets, criteria (C1-C21), color allocation, bye selection |
| `pairing/dutch` | `github.com/gnutterts/chesspairing/pairing/dutch` | Dutch (FIDE C.04.3) Swiss pairer with global Blossom matching |
| `pairing/burstein` | `github.com/gnutterts/chesspairing/pairing/burstein` | Burstein (FIDE C.04.4.2) Swiss variant with seeding rounds and opposition index |
| `pairing/dubov` | `github.com/gnutterts/chesspairing/pairing/dubov` | Dubov (FIDE C.04.4.1) Swiss variant with ARO-equalization and transposition matching |
| `pairing/keizer` | `github.com/gnutterts/chesspairing/pairing/keizer` | Keizer pairing (top-down by Keizer score, repeat avoidance) |
| `pairing/roundrobin` | `github.com/gnutterts/chesspairing/pairing/roundrobin` | Round-robin pairing (Berger table / circle method) |
| `scoring/standard` | `github.com/gnutterts/chesspairing/scoring/standard` | Standard scoring (1-half-0, configurable point values) |
| `scoring/keizer` | `github.com/gnutterts/chesspairing/scoring/keizer` | Keizer scoring (iterative convergence, value numbers) |
| `scoring/football` | `github.com/gnutterts/chesspairing/scoring/football` | Football scoring (3-1-0, wrapper around standard) |
| `tiebreaker` | `github.com/gnutterts/chesspairing/tiebreaker` | 25 tiebreakers with self-registering registry |
| `trf` | `github.com/gnutterts/chesspairing/trf` | TRF16 (FIDE Tournament Report File) reader, writer, and bidirectional conversion to/from `TournamentState` |

## Pairing Systems

All pairers implement `chesspairing.Pairer`:

```go
type Pairer interface {
    Pair(ctx context.Context, state *TournamentState) (*PairingResult, error)
}
```

### Dutch Swiss (FIDE C.04.3)

The most widely used Swiss pairing system. Players are grouped by score, then paired within brackets using all 21 FIDE quality criteria (C1-C21) with global Blossom maximum weight matching.

**Options:**

| Option | Values | Default | Description |
|--------|--------|---------|-------------|
| `Acceleration` | `"none"`, `"baku"` | `"none"` | Baku acceleration mode |
| `TopSeedColor` | `"auto"`, `"white"`, `"black"` | `"auto"` | Top seed color in round 1 |
| `ForbiddenPairs` | `[][]string` | `nil` | Player ID pairs that must not be paired |

```go
import "github.com/gnutterts/chesspairing/pairing/dutch"

pairer := dutch.New(dutch.Options{})

// Or with options:
accel := "baku"
pairer = dutch.New(dutch.Options{Acceleration: &accel})

result, err := pairer.Pair(ctx, state)
```

### Burstein Swiss (FIDE C.04.4.2)

Uses seeding rounds (delegating to Dutch matching) followed by opposition-index-based re-ranking. Seeding rounds = min(floor(totalRounds/2), 4). Uses criteria C1-C8 only (no float tracking).

**Options:**

| Option | Values | Default | Description |
|--------|--------|---------|-------------|
| `TopSeedColor` | `"auto"`, `"white"`, `"black"` | `"auto"` | Top seed color in round 1 |
| `ForbiddenPairs` | `[][]string` | `nil` | Player ID pairs that must not be paired |
| `TotalRounds` | `*int` | derived from state | Planned number of rounds (for seeding round calculation) |

```go
import "github.com/gnutterts/chesspairing/pairing/burstein"

pairer := burstein.New(burstein.Options{})
result, err := pairer.Pair(ctx, state)
```

### Dubov Swiss (FIDE C.04.4.1)

ARO-equalization Swiss variant. Score groups are split by colour preference (G1/G2), sorted by ascending ARO, and matched using transposition-based search with 10 criteria (C1-C10).

**Options:**

| Option | Values | Default | Description |
|--------|--------|---------|-------------|
| `TopSeedColor` | `"auto"`, `"white"`, `"black"` | `"auto"` | Top seed color in round 1 |
| `ForbiddenPairs` | `[][]string` | `nil` | Player ID pairs that must not be paired |
| `TotalRounds` | `*int` | derived from state | Planned number of rounds |

```go
import "github.com/gnutterts/chesspairing/pairing/dubov"

pairer := dubov.New(dubov.Options{})
result, err := pairer.Pair(ctx, state)
```

### Keizer

Outside-in pairing: rank 1 vs rank N, rank 2 vs rank N-1, and so on. The middle player gets a bye if there is an odd count. Includes repeat avoidance with a configurable gap and color balancing based on most recent game.

**Options:**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `AllowRepeatPairings` | `*bool` | `true` | Allow rematches within the tournament |
| `MinRoundsBetweenRepeats` | `*int` | `3` | Minimum rounds before a rematch is allowed |

```go
import "github.com/gnutterts/chesspairing/pairing/keizer"

pairer := keizer.New(keizer.Options{})
result, err := pairer.Pair(ctx, state)
```

### Round-Robin

Berger table / circle method. Supports single and double round-robin. For double round-robin, colors are reversed in the second cycle when `ColorBalance` is true.

**Options:**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Cycles` | `*int` | `1` | Number of complete round-robins (2 = double RR) |
| `ColorBalance` | `*bool` | `true` | Reverse colors in even cycles |

```go
import "github.com/gnutterts/chesspairing/pairing/roundrobin"

cycles := 2
pairer := roundrobin.New(roundrobin.Options{Cycles: &cycles})
result, err := pairer.Pair(ctx, state)
```

## Scoring Systems

All scorers implement `chesspairing.Scorer`:

```go
type Scorer interface {
    Score(ctx context.Context, state *TournamentState) ([]PlayerScore, error)
    PointsForResult(result GameResult, rctx ResultContext) float64
}
```

### Standard (1-half-0)

FIDE default. Each result awards a fixed number of points. All point values are configurable.

| Result | Default Points |
|--------|---------------|
| Win | 1.0 |
| Draw | 0.5 |
| Loss | 0.0 |
| Bye | 1.0 |
| Forfeit win | 1.0 |
| Forfeit loss | 0.0 |
| Absent | 0.0 |

```go
import "github.com/gnutterts/chesspairing/scoring/standard"

scorer := standard.New(standard.Options{})
scores, err := scorer.Score(ctx, state)
```

### Keizer

Iterative convergence scoring. Each player is assigned a value number based on their current rank (top player = N, decreasing by step). Points for a win equal the opponent's value number. Rankings and value numbers update iteratively until convergence.

```go
import "github.com/gnutterts/chesspairing/scoring/keizer"

scorer := keizer.New(keizer.Options{})
scores, err := scorer.Score(ctx, state)
```

### Football (3-1-0)

Wrapper around standard scoring with football-style defaults: 3 for a win, 1 for a draw, 0 for a loss. All point values remain configurable via `standard.Options`.

```go
import "github.com/gnutterts/chesspairing/scoring/football"

scorer := football.New(standard.Options{})
scores, err := scorer.Score(ctx, state)
```

## Tiebreakers

All tiebreakers implement `chesspairing.TieBreaker` and self-register via `init()`:

```go
type TieBreaker interface {
    ID() string
    Name() string
    Compute(ctx context.Context, state *TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
```

### Available Tiebreakers

| ID | Name | Description |
|----|------|-------------|
| `buchholz` | Buchholz | Sum of all opponents' scores |
| `buchholz-cut1` | Buchholz Cut 1 | Buchholz minus lowest opponent score |
| `buchholz-cut2` | Buchholz Cut 2 | Buchholz minus two lowest opponent scores |
| `buchholz-median` | Buchholz Median | Buchholz minus highest and lowest |
| `buchholz-median2` | Buchholz Median-2 | Buchholz minus two highest and two lowest |
| `sonneborn-berger` | Sonneborn-Berger | Sum of beaten opponents' scores + half of drawn opponents' scores |
| `direct-encounter` | Direct Encounter | Head-to-head score between tied players |
| `wins` | Games Won (OTB) | OTB wins only (excludes forfeits and byes) |
| `win` | Rounds Won | Rounds with a win result (OTB + forfeit wins + PAB) |
| `black-games` | Games with Black | Games played as black (excludes forfeits) |
| `black-wins` | Black Wins | OTB wins with the black pieces |
| `rounds-played` | Rounds Played | Number of rounds actually played |
| `standard-points` | Standard Points | Score under 1-half-0 regardless of scoring system |
| `pairing-number` | Pairing Number | Tournament pairing number (initial seed) |
| `koya` | Koya System | Score against opponents with 50%+ |
| `progressive` | Progressive Score | Cumulative score after each round |
| `aro` | Avg Rating of Opponents | Mean rating of all opponents faced |
| `fore-buchholz` | Fore Buchholz | Buchholz with pending games treated as draws |
| `avg-opponent-buchholz` | Avg Opponent Buchholz | Average of opponents' Buchholz scores |
| `performance-rating` | Performance Rating | TPR = ARO + dp(p) per FIDE B.02 |
| `performance-points` | Performance Points | Lowest rating with expected score >= actual |
| `avg-opponent-tpr` | Avg Opponent TPR | Average of opponents' TPR values |
| `avg-opponent-ptp` | Avg Opponent PTP | Average of opponents' PTP values |
| `player-rating` | Player Rating | Player's own rating |
| `games-played` | Games Played | Total games played (rewards participation) |

### Using the Registry

```go
import "github.com/gnutterts/chesspairing/tiebreaker"

// Look up by ID.
tb, err := tiebreaker.Get("buchholz-cut1")
if err != nil {
    log.Fatal(err)
}

values, err := tb.Compute(ctx, state, scores)

// List all registered IDs.
ids := tiebreaker.All()

// FIDE-recommended defaults per pairing system.
defaults := chesspairing.DefaultTiebreakers(chesspairing.PairingDutch)
// Returns: ["buchholz-cut1", "buchholz", "sonneborn-berger", "direct-encounter"]
```

## Game Results

| Constant | String Value | Description |
|----------|-------------|-------------|
| `ResultWhiteWins` | `"1-0"` | White wins (played) |
| `ResultBlackWins` | `"0-1"` | Black wins (played) |
| `ResultDraw` | `"0.5-0.5"` | Draw (played) |
| `ResultPending` | `"*"` | Game not yet played (initial state) |
| `ResultForfeitWhiteWins` | `"1-0f"` | White wins by forfeit (excluded from pairing history) |
| `ResultForfeitBlackWins` | `"0-1f"` | Black wins by forfeit (excluded from pairing history) |
| `ResultDoubleForfeit` | `"0-0f"` | Both players forfeit (excluded from pairing and scoring) |

Helper methods on `GameResult`: `IsValid()`, `IsRecordable()`, `IsForfeit()`, `IsDoubleForfeit()`.

## Testing

```
go test -race -count=1 ./...   # 583 tests across 14 packages
```

## License

TBD
