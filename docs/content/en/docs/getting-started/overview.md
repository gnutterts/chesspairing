---
title: "Overview"
linkTitle: "Overview"
weight: 1
description: "What chesspairing is and how its components fit together."
---

## What is chesspairing?

Chesspairing is a pure Go module (`github.com/gnutterts/chesspairing`) that handles three core tournament operations: **pairing** (deciding who plays whom), **scoring** (turning game results into standings), and **tiebreaking** (separating players with equal scores).

It has zero external dependencies -- the entire module runs on the Go standard library alone. There is no I/O, no database access, and no network communication. Every engine operates on in-memory data structures, which makes it straightforward to embed in a server, desktop application, or automated pipeline. It is safe for concurrent use when each goroutine supplies its own tournament state.

The module implements all six FIDE-approved Swiss pairing systems (as defined in the FIDE Handbook sections C.04.3 through C.04.6), plus Keizer pairing and FIDE round-robin pairing. You can combine any pairing system with any scoring system -- for example, Swiss pairing with Keizer scoring, or round-robin with football scoring. Nothing forces a particular combination.

## Architecture

Three interfaces in the root package define what the engines do:

```go
type Pairer interface {
    Pair(ctx context.Context, state *TournamentState) (*PairingResult, error)
}

type Scorer interface {
    Score(ctx context.Context, state *TournamentState) ([]PlayerScore, error)
    PointsForResult(result GameResult, rctx ResultContext) float64
}

type TieBreaker interface {
    ID() string
    Name() string
    Compute(ctx context.Context, state *TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
```

A `TournamentState` is the single input to all three. It is a read-only snapshot containing the player list, all completed rounds (with game results and byes), the current round number, and the pairing/scoring configuration. The calling code constructs this snapshot from whatever storage it uses, and the engines return results without side effects.

### Data flow

```text
Caller builds TournamentState
  |
  +---> Pairer.Pair()       --> PairingResult (board pairings + byes)
  |
  +---> Scorer.Score()      --> []PlayerScore (score per player, ranked)
  |
  +---> TieBreaker.Compute()--> []TieBreakValue (one numeric value per player)
```

The three steps are independent. You can call `Pair` without scoring, compute scores without tiebreakers, or run all three in sequence. A typical tournament round looks like:

1. Build a `TournamentState` from your data source.
2. Call `Scorer.Score()` to produce the current standings (the pairer uses these internally too).
3. Call `Pairer.Pair()` to generate pairings for the next round.
4. After the round is played and results are recorded, call `Scorer.Score()` again and then one or more `TieBreaker.Compute()` calls to produce final rankings.

## What's included

### Pairing systems

| System       | FIDE Ref | Package               | Notes                                                 |
| ------------ | -------- | --------------------- | ----------------------------------------------------- |
| Dutch        | C.04.3   | `pairing/dutch`       | Global Blossom matching, Baku acceleration            |
| Burstein     | C.04.4.2 | `pairing/burstein`    | Seeding/post-seeding rounds with opposition index     |
| Dubov        | C.04.4.1 | `pairing/dubov`       | ARO-based ranking, 10 dedicated criteria              |
| Lim          | C.04.4.3 | `pairing/lim`         | Median-first processing, exchange-based matching      |
| Double-Swiss | C.04.5   | `pairing/doubleswiss` | Lexicographic bracket pairing                         |
| Team Swiss   | C.04.6   | `pairing/team`        | Team-level pairing with configurable color preference |
| Keizer       | --       | `pairing/keizer`      | Top-down by Keizer score, repeat avoidance            |
| Round-Robin  | C.05     | `pairing/roundrobin`  | FIDE Berger tables, multi-cycle support               |

Each pairing engine implements the `Pairer` interface and has an `Options` struct for system-specific configuration. See the [Pairing Systems](/docs/pairing-systems/) section for detailed documentation of each system.

### Scoring systems

| System   | Default Points        | Package            | Notes                                                     |
| -------- | --------------------- | ------------------ | --------------------------------------------------------- |
| Standard | 1 -- 0.5 -- 0         | `scoring/standard` | Configurable point values for wins, draws, byes, forfeits |
| Keizer   | Iterative convergence | `scoring/keizer`   | Rank-dependent scoring with 24 configurable parameters    |
| Football | 3 -- 1 -- 0           | `scoring/football` | Thin wrapper around Standard with different defaults      |

Each scoring engine implements the `Scorer` interface. See the [Scoring Systems](/docs/scoring/) section for details on configuration and behavior.

### Tiebreakers

Chesspairing includes 25 tiebreaker implementations that self-register through a central registry. They cover the full range of FIDE-recognized methods:

| Category           | Tiebreakers                                                                                                            |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------- |
| Buchholz family    | Buchholz, Buchholz Cut-1, Buchholz Cut-2, Buchholz Median, Buchholz Median-2, Fore Buchholz, Average Opponent Buchholz |
| Head-to-head       | Direct Encounter, Koya System                                                                                          |
| Results-based      | Games Won, Rounds Won, Progressive Score, Standard Points                                                              |
| Performance        | Performance Rating (TPR), Performance Points (PTP), Average Opponent TPR, Average Opponent PTP, Player Rating          |
| Color and activity | Games with Black, Black Wins, Rounds Played, Games Played                                                              |
| Rating-based       | Average Rating of Opponents (ARO)                                                                                      |
| Administrative     | Pairing Number (TPN)                                                                                                   |

Every tiebreaker implements the `TieBreaker` interface and can be referenced by its string ID (e.g. `"buchholz-cut1"`, `"sonneborn-berger"`, `"direct-encounter"`). See the [Tiebreakers](/docs/tiebreakers/) section for the full registry and computation details.

## Two ways to use it

### Command-line tool

The `chesspairing` CLI reads [TRF16 files](/docs/formats/trf16/) (the standard FIDE tournament exchange format) and produces pairings, standings, and validation reports. It offers eight subcommands:

| Command       | Purpose                                           |
| ------------- | ------------------------------------------------- |
| `pair`        | Generate pairings for the next round              |
| `check`       | Verify existing pairings match engine output      |
| `generate`    | Output an updated TRF file with pairings appended |
| `validate`    | Validate a TRF file against multiple profiles     |
| `standings`   | Compute and display current standings             |
| `tiebreakers` | List available tiebreakers                        |
| `convert`     | Re-serialize a TRF file                           |
| `version`     | Print version information                         |

Output is available in five formats: plain list (bbpPairings-compatible), wide tabular, board view, XML, and JSON. A legacy mode provides drop-in compatibility with bbpPairings and JaVaFo command-line conventions.

See the [CLI Quickstart](/docs/getting-started/cli-quickstart/) to get started, or the [CLI Reference](/docs/cli/) for complete documentation.

### Go library

Add the module to your project:

```bash
go get github.com/gnutterts/chesspairing
```

Then construct a `TournamentState`, pick your engines, and call them:

```go
import (
    "context"
    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/pairing/dutch"
    "github.com/gnutterts/chesspairing/scoring/standard"
)

state := &chesspairing.TournamentState{
    Players:      players,
    Rounds:       rounds,
    CurrentRound: 3,
    PairingConfig: chesspairing.PairingConfig{
        System: chesspairing.PairingDutch,
    },
    ScoringConfig: chesspairing.ScoringConfig{
        System: chesspairing.ScoringStandard,
    },
}

pairer := dutch.NewFromMap(nil) // default options
result, err := pairer.Pair(context.Background(), state)
```

Every engine also provides a `NewFromMap(map[string]any)` constructor for instantiation from configuration maps, making it straightforward to wire up engines from JSON config or database records.

See the [Go Library Quickstart](/docs/getting-started/go-quickstart/) for a complete walkthrough, or the [API Reference](/docs/api/) for type and package documentation.

## Next steps

- [CLI Quickstart](/docs/getting-started/cli-quickstart/) -- install the tool and pair a tournament in under five minutes
- [Go Library Quickstart](/docs/getting-started/go-quickstart/) -- add the module to a Go project and generate pairings programmatically
- [For Arbiters](/docs/getting-started/for-arbiters/) -- understand how chesspairing maps to FIDE regulations
- [For Researchers](/docs/getting-started/for-researchers/) -- explore the algorithm implementations and matching strategies
- [Concepts](/docs/concepts/) -- background on Swiss systems, scoring, tiebreaking, colors, byes, and floaters
