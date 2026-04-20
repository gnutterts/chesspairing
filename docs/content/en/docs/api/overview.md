---
title: "Package Organization"
linkTitle: "Overview"
weight: 1
description: "How the chesspairing module is organized into packages and how they relate."
---

## Root package

The `chesspairing` package defines three interfaces (`Pairer`, `Scorer`, `TieBreaker`) and all shared types (`TournamentState`, `PlayerEntry`, `RoundData`, `GameData`, `PairingResult`, `PlayerScore`, `TieBreakValue`, etc.). It contains no implementation code -- only contracts and data structures.

Config enums (`PairingSystem`, `ScoringSystem`, `GameResult`, `ByeType`) and helper functions (`DefaultTiebreakers()`) also live here. Each enum has a `Parse*` constructor (`ParsePairingSystem`, `ParseScoringSystem`, `ParseGameResult`, `ParseByeType`) for round-tripping config from strings. `PlayedPairs(state, HistoryOptions{})` returns the set of unordered player pairs already played, suitable as a forbidden-pair constraint when building the next round; `HistoryOptions.IncludeForfeits` controls whether single-forfeit games count as played.

## Engine packages

Each pairer and scorer lives in its own package. Every engine package follows the same structure:

- A public `Pairer` or `Scorer` struct.
- An `Options` struct with pointer fields. A nil field means "use the default value."
- `WithDefaults()` method -- fills nil fields with sensible defaults.
- `ParseOptions(map[string]any)` -- parses options from a generic map (typically from JSON config).
- `NewFromMap(map[string]any)` -- constructor for factory instantiation. Returns the engine with options parsed and defaults applied.
- Compile-time interface check:

```go
var _ chesspairing.Pairer = (*Pairer)(nil)
```

### Pairer packages

| Package               | System         | Spec          |
| --------------------- | -------------- | ------------- |
| `pairing/dutch`       | Dutch Swiss    | FIDE C.04.3   |
| `pairing/burstein`    | Burstein Swiss | FIDE C.04.4.2 |
| `pairing/dubov`       | Dubov Swiss    | FIDE C.04.4.1 |
| `pairing/lim`         | Lim Swiss      | FIDE C.04.4.3 |
| `pairing/doubleswiss` | Double-Swiss   | FIDE C.04.5   |
| `pairing/team`        | Team Swiss     | FIDE C.04.6   |
| `pairing/keizer`      | Keizer         | --            |
| `pairing/roundrobin`  | Round-Robin    | FIDE C.05     |

### Scorer packages

| Package            | System                                          |
| ------------------ | ----------------------------------------------- |
| `scoring/standard` | Standard (1-0.5-0, configurable point values)   |
| `scoring/keizer`   | Keizer (iterative convergence, variant support) |
| `scoring/football` | Football (3-1-0, thin wrapper around standard)  |

## Shared libraries

Two internal libraries provide shared logic for Swiss pairers. Callers never import these directly -- they interact only through the public pairer interfaces.

### `pairing/swisslib`

Used by the Dutch and Burstein pairers. Provides:

- `PlayerState` construction from tournament history
- Score groups and brackets
- Bye selection (completability-based)
- Color preference and allocation
- Absolute criteria (C1--C4) and optimization criteria (C8--C21)
- Edge weight computation for Blossom matching (`*big.Int`)
- `PairBracketsGlobal()` -- global Blossom matching with Stage 0.5 completability pre-matching
- Structural validation

### `pairing/lexswiss`

Used by the Double-Swiss and Team Swiss pairers. Provides:

- `ParticipantState` construction
- Score groups
- Bye assignment and up-floater selection
- `PairBracket()` -- lexicographic bracket pairing with pluggable criteria

## Support packages

### `tiebreaker`

Self-registering registry. Each tiebreaker file calls `Register()` in its `init()` function. The registry provides `Get(name)` to retrieve a tiebreaker by ID and `All()` to list available tiebreakers. 25 tiebreakers are registered.

### `trf`

Bidirectional TRF conversion:

- `Read(io.Reader)` -- parses a TRF16 document.
- `Write(io.Writer, *Document)` -- serializes a TRF16 document.
- `ToTournamentState()` -- converts a TRF `Document` to a `chesspairing.TournamentState`.
- `FromTournamentState()` -- converts a `TournamentState` back to a TRF `Document`.
- `Document.Validate()` -- multi-profile validation (General, PairingEngine, FIDE).

### `factory`

Constructs engines by name from a generic config map. Three entry points:

- `NewPairer(name string, opts map[string]any)` -- returns a configured `chesspairing.Pairer`.
- `NewScorer(name string, opts map[string]any)` -- returns a configured `chesspairing.Scorer`.
- `NewTieBreaker(name string)` -- looks up a registered tiebreaker by ID.

Useful when the pairing/scoring system is chosen at runtime from JSON or CLI flags rather than wired in at compile time.

### `standings`

Composes a `Scorer` and a slice of `TieBreaker`s into a presentation-ready table. `Build(ctx, state, scorer, tieBreakers)` runs scoring, runs each tiebreaker, sorts by score then by tiebreaker columns in order, and returns `[]Standing` with shared rank for true ties (standard "1224" ranking). `BuildByID(ctx, state, scorer, tbIDs)` resolves tiebreaker IDs through the registry as a convenience. Wins, draws, and losses are derived from game results, since W/D/L is orthogonal to the scoring rule.

### `algorithm/blossom`

Standalone implementation of Edmonds' maximum weight matching for general graphs. O(n^3). Two variants:

- `MaxWeightMatching(edges, maxCardinality)` -- int64 weights.
- `MaxWeightMatchingBig(edges, maxCardinality)` -- `*big.Int` weights (needed when edge weight bit layouts exceed 64 bits).

Ported from Joris van Rantwijk's Python reference implementation.

### `algorithm/varma`

Standalone lookup tables from FIDE C.05 Annex 2. Used for federation-aware pairing number assignment in round-robin tournaments.

- `Groups()` -- returns Varma groupings for a given player count.
- `Assign()` -- assigns pairing numbers respecting federation constraints.

## Dependency flow

```text
Caller code
  -> chesspairing (interfaces + types)
  -> pairing/* (implement Pairer)
      -> pairing/swisslib or pairing/lexswiss (shared logic)
      -> algorithm/blossom (matching)
  -> scoring/* (implement Scorer)
  -> tiebreaker (implement TieBreaker)
  -> trf (I/O layer)
```

All arrows point inward to the root package. Engine packages depend on the root package and their shared libraries, but never on each other. A pairer package never imports a scorer package, and vice versa. The `trf` package depends on the root package for type definitions but is otherwise independent of the engine packages.
