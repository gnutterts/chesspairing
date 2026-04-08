---
title: "Football Scoring"
linkTitle: "Football"
weight: 3
description: "A 3-1-0 scoring system that rewards wins more heavily than draws."
---

Football scoring borrows from association football: 3 points for a win, 1 for a draw, 0 for a loss. This shifts the incentive structure compared to standard 1-half-0 scoring -- a win is now worth three draws instead of two, which discourages passive play and rewards decisive results.

The implementation is a thin wrapper around the [standard scorer](/docs/scoring/standard/). It applies football-specific defaults and then delegates all scoring logic to the standard engine. This means the algorithm, ranking rules, and result handling are identical to standard scoring; only the point values differ.

## When to use

- **Club events that want to discourage draws.** In standard scoring, two draws equal one win. In football scoring, three draws equal one win, making draws significantly less attractive.
- **Informal or rapid tournaments.** Football scoring is popular in casual and rapid-play events where decisive games create a more dynamic tournament atmosphere.
- **Any format.** Football scoring works with Swiss, round-robin, or any other pairing system -- it is purely a scoring change.

Football scoring is not used in FIDE-rated events.

## Configuration

### CLI

Pass options through the TRF `XXY` field or via `--config`:

```bash
chesspairing pair --config '{"scoring": {"pointWin": 3.0, "pointDraw": 1.0}}' tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/scoring/football"

// With all defaults (3-1-0).
scorer := football.New(standard.Options{})

// Override specific values.
scorer := football.New(standard.Options{
    PointDraw: chesspairing.Float64Ptr(0.5),
})

// From a generic map.
scorer := football.NewFromMap(map[string]any{
    "pointDraw": 0.5,
})

// Use the Scorer interface.
scores, err := scorer.Score(ctx, &state)
points := scorer.PointsForResult(result, rctx)
```

The `Scorer` type satisfies the `chesspairing.Scorer` interface at compile time:

```go
var _ chesspairing.Scorer = (*football.Scorer)(nil)
```

Note that `football.New()` and `football.NewFromMap()` both accept `standard.Options` -- the same options struct used by the standard scorer. Any nil fields receive football defaults instead of standard defaults.

### Options reference

Football scoring uses the same `standard.Options` struct with different defaults:

| Field              | JSON key           | Football default | Standard default | Description                               |
| ------------------ | ------------------ | ---------------- | ---------------- | ----------------------------------------- |
| `PointWin`         | `pointWin`         | 3.0              | 1.0              | Points for an over-the-board win.         |
| `PointDraw`        | `pointDraw`        | 1.0              | 0.5              | Points for a draw.                        |
| `PointLoss`        | `pointLoss`        | 0.0              | 0.0              | Points for a loss.                        |
| `PointBye`         | `pointBye`         | 3.0              | 1.0              | Points for a pairing-allocated bye (PAB). |
| `PointForfeitWin`  | `pointForfeitWin`  | 3.0              | 1.0              | Points for winning by forfeit.            |
| `PointForfeitLoss` | `pointForfeitLoss` | 0.0              | 0.0              | Points for losing by forfeit.             |
| `PointAbsent`      | `pointAbsent`      | 0.0              | 0.0              | Points when absent.                       |

All fields are `*float64`. When you explicitly set a value, it takes precedence over the football default. For instance, setting `PointDraw` to 0.5 gives you a 3-0.5-0 system.

## How it works

Football scoring delegates entirely to the standard scoring engine. The `Score()` and `PointsForResult()` methods call through to the underlying `standard.Scorer` instance.

The only difference is in default initialization: where `standard.Options.WithDefaults()` fills nil fields with 1-half-0 values, the football scorer fills nil fields with 3-1-0 values before passing the options to the standard engine. Once defaults are applied, the standard engine handles everything -- game processing, bye handling, absence detection, and ranking.

The algorithm is the same single-pass approach described in [standard scoring](/docs/scoring/standard/):

1. Initialize zero scores for all active players.
2. Process each round: score games, score byes, detect absences.
3. Rank by score descending, rating descending, display name ascending.

## Examples

### Default football scoring

```go
scorer := football.New(standard.Options{})
scores, _ := scorer.Score(ctx, &state)
// Player who won 3, drew 1, lost 1: 3×3.0 + 1×1.0 + 1×0.0 = 10.0
```

### Custom draw value

Reduce the draw value to make wins even more dominant:

```go
scorer := football.New(standard.Options{
    PointDraw: chesspairing.Float64Ptr(0.5),
})
// A win (3.0) is now worth six draws (6×0.5 = 3.0).
```

### Comparing with standard scoring

Consider a player with 5 wins, 3 draws, and 2 losses:

| System   | Calculation              | Total |
| -------- | ------------------------ | ----- |
| Standard | 5(1.0) + 3(0.5) + 2(0.0) | 6.5   |
| Football | 5(3.0) + 3(1.0) + 2(0.0) | 18.0  |

The relative standings between players stay the same when all players have identical numbers of games. Where football scoring changes outcomes is when players have different win/draw ratios. A player with 4 wins and 4 draws (4 wins and 2 losses in standard: 5.0 vs 6.0) gets 16.0 in football, while a player with 6 draws and 2 wins gets 8.0 -- a much larger gap than in standard scoring.

## Related

- [Standard scoring](/docs/scoring/standard/) -- the underlying engine and the full options reference
- [Scoring concepts](/docs/concepts/scoring/) -- overview of all three scoring systems
- [Keizer scoring](/docs/scoring/keizer/) -- the ranking-based alternative
- [Scorer interface](/docs/api/scorer/) -- API reference for the `Scorer` interface
