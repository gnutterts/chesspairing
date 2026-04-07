---
title: "Go API Reference"
linkTitle: "Go API"
weight: 70
description: "Hand-written API documentation for the chesspairing Go module — types, interfaces, and usage patterns."
---

This is hand-written API documentation for the `github.com/gnutterts/chesspairing` Go module. The code is the single source of truth; these pages describe what the code does, not the other way around.

## Module info

- **Module path**: `github.com/gnutterts/chesspairing`
- **Go version**: 1.26.1
- **External dependencies**: none (stdlib only)

## Core interfaces

The root package defines three interfaces:

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

## Data flow

```text
*TournamentState
  -> Pairer.Pair()       -> *PairingResult
  -> Scorer.Score()      -> []PlayerScore
  -> TieBreaker.Compute() -> []TieBreakValue
```

Pairing and scoring are independent. Any pairer can combine with any scorer. A tournament can use Swiss pairing with Keizer scoring, or Round-robin pairing with Football scoring.

## Context parameter

All methods accept `context.Context` as their first parameter for API compatibility. Currently, no engine checks for cancellation -- all computation is CPU-bound and in-memory.

## Concurrency

All engines are safe for concurrent use when each goroutine supplies its own `TournamentState`. There is no shared mutable state.

## Packages

| Package                          | Purpose                                  |
| -------------------------------- | ---------------------------------------- |
| [`chesspairing`](overview/)      | Interfaces, shared types, config enums   |
| [`pairing/dutch`](pairer/)       | Dutch Swiss pairer (C.04.3)              |
| [`pairing/burstein`](pairer/)    | Burstein Swiss pairer (C.04.4.2)         |
| [`pairing/dubov`](pairer/)       | Dubov Swiss pairer (C.04.4.1)            |
| [`pairing/lim`](pairer/)         | Lim Swiss pairer (C.04.4.3)              |
| [`pairing/doubleswiss`](pairer/) | Double-Swiss pairer (C.04.5)             |
| [`pairing/team`](pairer/)        | Team Swiss pairer (C.04.6)               |
| [`pairing/keizer`](pairer/)      | Keizer pairer                            |
| [`pairing/roundrobin`](pairer/)  | Round-Robin pairer (C.05)                |
| [`scoring/standard`](scorer/)    | Standard scoring (1-0.5-0)               |
| [`scoring/keizer`](scorer/)      | Keizer scoring (iterative)               |
| [`scoring/football`](scorer/)    | Football scoring (3-1-0)                 |
| [`tiebreaker`](tiebreaker/)      | 25 tiebreaker implementations + registry |
| [`trf`](trf/)                    | TRF16/TRF-2026 I/O and validation        |
| [`algorithm/blossom`](overview/) | Edmonds' maximum weight matching         |
| [`algorithm/varma`](overview/)   | Varma lookup tables (C.05 Annex 2)       |

## Sub-pages

- [Package Organization](overview/) -- how packages relate and the dependency flow
- [Core Types](core-types/) -- `TournamentState`, `PlayerEntry`, `GameData`, result types
- [Pairer Interface](pairer/) -- `Pairer` contract, `PairingResult`, all pairer implementations
- [Scorer Interface](scorer/) -- `Scorer` contract, `PlayerScore`, scoring engines
- [TieBreaker Interface](tiebreaker/) -- `TieBreaker` contract, registry, all 25 tiebreakers
- [Options Pattern](options/) -- how engine options work (pointer fields, `WithDefaults`, `ParseOptions`)
- [Config & Enums](config/) -- `PairingSystem`, `ScoringSystem`, `DefaultTiebreakers()`
- [TRF Package](trf/) -- TRF16 reading, writing, validation, and bidirectional conversion
