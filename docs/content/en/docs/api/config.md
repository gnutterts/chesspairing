---
title: "Configuration and Factory"
linkTitle: "Configuration"
weight: 8
description: "PairingSystem and ScoringSystem enums, config structs, and DefaultTiebreakers."
---

The root package defines enum types and config structs that drive engine selection and configuration. These types are used by the CLI, the `trf` package, and any caller that needs to configure a tournament generically.

## PairingSystem

```go
type PairingSystem string
```

Identifies which pairing algorithm to use. Eight constants are defined:

| Constant             | Value           | FIDE regulation |
| -------------------- | --------------- | --------------- |
| `PairingDutch`       | `"dutch"`       | C.04.3          |
| `PairingBurstein`    | `"burstein"`    | C.04.4.2        |
| `PairingDubov`       | `"dubov"`       | C.04.4.1        |
| `PairingLim`         | `"lim"`         | C.04.4.3        |
| `PairingDoubleSwiss` | `"doubleswiss"` | C.04.5          |
| `PairingTeam`        | `"team"`        | C.04.6          |
| `PairingKeizer`      | `"keizer"`      | --              |
| `PairingRoundRobin`  | `"roundrobin"`  | C.05 Annex 1    |

### IsValid

```go
func (p PairingSystem) IsValid() bool
```

Returns `true` if `p` is one of the eight recognized constants.

## ScoringSystem

```go
type ScoringSystem string
```

Identifies which scoring algorithm to use. Three constants:

| Constant          | Value        |
| ----------------- | ------------ |
| `ScoringStandard` | `"standard"` |
| `ScoringKeizer`   | `"keizer"`   |
| `ScoringFootball` | `"football"` |

### IsValid

```go
func (s ScoringSystem) IsValid() bool
```

Returns `true` if `s` is one of the three recognized constants.

## PairingConfig

```go
type PairingConfig struct {
    System  PairingSystem
    Options map[string]any
}
```

Holds the pairing system selection and its engine-specific options. The `Options` map is passed directly to the engine's `NewFromMap()` constructor. Keys and accepted values depend on the engine -- see the [Options Pattern](../options/) page for details.

Example:

```go
cfg := chesspairing.PairingConfig{
    System: chesspairing.PairingDutch,
    Options: map[string]any{
        "acceleration": "baku",
        "topSeedColor": "white",
    },
}
```

## ScoringConfig

```go
type ScoringConfig struct {
    System      ScoringSystem
    Tiebreakers []string
    Options     map[string]any
}
```

Holds the scoring system selection, the ordered list of tiebreaker IDs, and engine-specific scorer options.

- **System**: which scoring algorithm to use.
- **Tiebreakers**: ordered list of tiebreaker registry IDs (e.g. `"buchholz-cut1"`, `"sonneborn-berger"`). Evaluated in order to break ties in standings.
- **Options**: passed to the scorer's `NewFromMap()`.

Example:

```go
cfg := chesspairing.ScoringConfig{
    System:      chesspairing.ScoringStandard,
    Tiebreakers: []string{"buchholz-cut1", "buchholz", "sonneborn-berger"},
    Options: map[string]any{
        "pointWin":  1.0,
        "pointDraw": 0.5,
    },
}
```

## DefaultTiebreakers

```go
func DefaultTiebreakers(system PairingSystem) []string
```

Returns the FIDE-recommended tiebreaker sequence for the given pairing system. This is used as the default when no tiebreakers are explicitly configured.

| Pairing system                                          | Default tiebreakers                                                 |
| ------------------------------------------------------- | ------------------------------------------------------------------- |
| Swiss (Dutch, Burstein, Dubov, Lim, Double-Swiss, Team) | `buchholz-cut1`, `buchholz`, `sonneborn-berger`, `direct-encounter` |
| Round-Robin                                             | `sonneborn-berger`, `direct-encounter`, `wins`, `koya`              |
| Keizer                                                  | `games-played`, `direct-encounter`, `wins`                          |
| Other/unknown                                           | `direct-encounter`, `wins`                                          |

Usage:

```go
tbs := chesspairing.DefaultTiebreakers(chesspairing.PairingDutch)
// ["buchholz-cut1", "buchholz", "sonneborn-berger", "direct-encounter"]
```

## Putting it together

A `TournamentState` carries both a `PairingConfig` and a `ScoringConfig`. Together they fully describe how a tournament should be paired and scored:

```go
state := chesspairing.TournamentState{
    PairingConfig: chesspairing.PairingConfig{
        System: chesspairing.PairingDutch,
        Options: map[string]any{
            "totalRounds":  9,
            "acceleration": "baku",
        },
    },
    ScoringConfig: chesspairing.ScoringConfig{
        System:      chesspairing.ScoringStandard,
        Tiebreakers: chesspairing.DefaultTiebreakers(chesspairing.PairingDutch),
    },
    // ... players, rounds, etc.
}
```

The CLI factory and the `trf.ToTournamentState()` function both produce `TournamentState` values with these configs populated, so downstream code can instantiate engines generically via `NewFromMap`.
