---
title: "Standard Scoring"
linkTitle: "Standard"
weight: 1
description: "The classic 1-½-0 scoring system with configurable point values for wins, draws, byes, forfeits, and absences."
---

Standard scoring assigns a fixed number of points for each game outcome: 1 for a win, 0.5 for a draw, 0 for a loss. Every point value is configurable, but the default values follow FIDE conventions used in virtually all rated Swiss and round-robin events.

Because points depend only on the result -- not on the opponent -- standard scoring requires no iteration and produces deterministic standings after a single pass through the results.

## When to use

Standard scoring is the right choice for most tournaments:

- **FIDE-rated events.** Required by FIDE regulations for all officially rated competitions.
- **Swiss and round-robin tournaments.** The expected scoring system for these formats.
- **Any event where simplicity matters.** Points are easy to explain and verify: a win is always worth the same, regardless of who you beat.

Even when a tournament uses a different public scoring system (like Keizer), the Swiss pairers use standard scoring internally to form score groups.

## Configuration

### CLI

Pass scoring options through the TRF `XXY` field or via `--config`:

```bash
chesspairing pair --config '{"scoring": {"pointWin": 1.0, "pointDraw": 0.5, "pointBye": 0.5}}' tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/scoring/standard"

// With explicit options (nil fields use defaults).
scorer := standard.New(standard.Options{
    PointBye: chesspairing.Float64Ptr(0.5),
})

// From a generic map (e.g. parsed from JSON config).
scorer := standard.NewFromMap(map[string]any{
    "pointBye": 0.5,
})

// Use the Scorer interface.
scores, err := scorer.Score(ctx, &state)
points := scorer.PointsForResult(result, rctx)
```

The `Scorer` type satisfies the `chesspairing.Scorer` interface at compile time:

```go
var _ chesspairing.Scorer = (*standard.Scorer)(nil)
```

### Options reference

All fields are `*float64`. A `nil` value means "use the default." This distinguishes "not configured" from "explicitly set to zero."

| Field              | JSON key           | Default | Description                               |
| ------------------ | ------------------ | ------- | ----------------------------------------- |
| `PointWin`         | `pointWin`         | 1.0     | Points for an over-the-board win.         |
| `PointDraw`        | `pointDraw`        | 0.5     | Points for a draw.                        |
| `PointLoss`        | `pointLoss`        | 0.0     | Points for a loss.                        |
| `PointBye`         | `pointBye`         | 1.0     | Points for a pairing-allocated bye (PAB). |
| `PointForfeitWin`  | `pointForfeitWin`  | 1.0     | Points for winning by forfeit.            |
| `PointForfeitLoss` | `pointForfeitLoss` | 0.0     | Points for losing by forfeit.             |
| `PointAbsent`      | `pointAbsent`      | 0.0     | Points when absent (no game and no bye).  |

## How it works

### Score()

The `Score()` method makes a single pass through all rounds to accumulate points:

1. **Initialize.** Create a zero score for every active player.

2. **Process each round.** For every round in the tournament:
   - **Games.** Each game is scored according to its result:
     - _Double forfeit_ -- both players receive 0 (the game is treated as if it never happened).
     - _Single forfeit_ -- the winner receives `PointForfeitWin`, the loser receives `PointForfeitLoss`.
     - _Regular result_ -- `PointWin`/`PointDraw`/`PointLoss` as appropriate. Pending games (`*`) contribute nothing.

   - **Byes.** Each bye is scored by type:
     - PAB -- `PointBye`
     - Half-point bye -- `PointDraw`
     - Zero-point bye -- `PointLoss`
     - Absent bye -- `PointAbsent`

   - **Absent detection.** Any active player who neither played a game nor received a bye in the round is treated as absent and receives `PointAbsent`.

3. **Rank.** Players are sorted by score descending, then by rating descending, then by display name ascending.

### PointsForResult()

Returns the point value for a single result. The method checks conditions in this order:

1. If `IsAbsent` -- return `PointAbsent`.
2. If `IsBye` -- return `PointBye`.
3. If `IsForfeit` -- return `PointForfeitWin` or `PointForfeitLoss` based on the result direction.
4. Otherwise -- return `PointWin`, `PointDraw`, or 0 for a regular game.

This priority order means that a forfeit win in a bye context is scored as a bye, not as a forfeit.

## Examples

### Default FIDE scoring

```go
scorer := standard.New(standard.Options{})
scores, _ := scorer.Score(ctx, &state)
// Player who won 3, drew 1, lost 1: 3×1.0 + 1×0.5 + 1×0.0 = 3.5
```

### Half-point PAB

Some organisers prefer a half-point bye instead of a full point:

```go
scorer := standard.New(standard.Options{
    PointBye: chesspairing.Float64Ptr(0.5),
})
```

### Penalizing absences

Award a negative score for unexcused absences:

```go
scorer := standard.New(standard.Options{
    PointAbsent: chesspairing.Float64Ptr(-1.0),
})
```

## Related

- [Scoring concepts](/docs/concepts/scoring/) -- overview of all three scoring systems and how they interact with pairing
- [Football scoring](/docs/scoring/football/) -- the 3-1-0 variant built on top of standard scoring
- [Keizer scoring](/docs/scoring/keizer/) -- iterative ranking-based alternative
- [Byes](/docs/concepts/byes/) -- bye types and how they are scored
- [Forfeits and absences](/docs/concepts/forfeits/) -- how forfeit results affect scoring and pairing history
- [Scorer interface](/docs/api/scorer/) -- API reference for the `Scorer` interface
