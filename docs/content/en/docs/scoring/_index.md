---
title: "Scoring Systems"
linkTitle: "Scoring"
weight: 40
description: "Three scoring engines — Standard (1-½-0), Keizer (iterative convergence), and Football (3-1-0)."
---

Chesspairing includes three scoring engines. Each one implements the `Scorer` interface and converts game results into player scores. Scoring is independent of pairing -- any scoring engine can be combined with any pairing system.

## At a glance

| System                | Point scheme                         | Iteration                   | Primary use case                          |
| --------------------- | ------------------------------------ | --------------------------- | ----------------------------------------- |
| [Standard](standard/) | 1 - 0.5 - 0 (configurable)           | None (single pass)          | FIDE-rated events, Swiss, Round-Robin     |
| [Keizer](keizer/)     | Dynamic (opponent-strength-weighted) | Iterative (up to 20 rounds) | Club tournaments, competitive league play |
| [Football](football/) | 3 - 1 - 0 (configurable)             | None (single pass)          | Events wanting stronger win incentives    |

## How they differ

**Standard scoring** assigns fixed point values: 1 for a win, 0.5 for a draw, 0 for a loss. Every point value is configurable, including separate values for byes, forfeit wins, forfeit losses, and absences. Because the value of a result never depends on the opponent, a single pass through the results produces final standings. This is the scoring system used in virtually all rated chess events.

**Keizer scoring** makes a result's value depend on the opponent's current ranking. Beating the top-ranked player is worth more than beating the bottom-ranked player. Since rankings depend on scores and scores depend on rankings, Keizer uses iterative convergence: compute scores, re-rank, recompute, and repeat until the ranking stabilises (or for a maximum of 20 iterations). The engine uses x2 integer arithmetic internally and includes 2-cycle oscillation detection with averaging to guarantee termination. Keizer scoring is designed for club events where rewarding play against strong opposition creates competitive incentives.

**Football scoring** is a thin wrapper around Standard scoring with different defaults: 3 for a win, 1 for a draw, 0 for a loss. The higher win/draw ratio discourages short draws and rewards decisive games. All point values remain configurable. Under the hood, Football delegates entirely to the Standard engine with adjusted default parameters.

## Scoring and pairing interaction

Pairing and scoring are intentionally decoupled. A tournament can use Swiss pairing with Keizer scoring, Round-Robin with Football scoring, or any other combination. The Swiss pairers use standard 1-0.5-0 scoring internally for score group formation regardless of the tournament's public scoring system -- this is by design, since FIDE Swiss regulations define score groups in terms of standard points.

The one exception is the Keizer pairer, which uses Keizer scores to determine pairing order. Using the Keizer pairer with a non-Keizer scoring system would produce arbitrary pairings, so this combination is not meaningful.

## Forfeit and bye handling

All three scoring engines handle the same set of special results:

| Result type    | Standard (default) | Keizer                         | Football (default) |
| -------------- | ------------------ | ------------------------------ | ------------------ |
| OTB Win        | 1.0                | Dynamic (opponent-rank-based)  | 3.0                |
| OTB Draw       | 0.5                | Dynamic                        | 1.0                |
| OTB Loss       | 0.0                | Dynamic                        | 0.0                |
| Forfeit Win    | 1.0                | Fixed fraction of self-victory | 3.0                |
| Forfeit Loss   | 0.0                | 0.0                            | 0.0                |
| PAB (Bye)      | 1.0                | Fixed fraction of self-victory | 3.0                |
| Absent         | 0.0                | 0.0                            | 0.0                |
| Double Forfeit | 0.0 / 0.0          | 0.0 / 0.0                      | 0.0 / 0.0          |

In Keizer scoring, forfeit wins and byes are computed as a configurable fraction of the "self-victory" value -- the points a player would receive for beating themselves at their current ranking. This keeps non-game results proportional to a player's strength.

## Interface

All three engines implement the same interface:

```go
type Scorer interface {
    Score(ctx context.Context, state TournamentState) ([]PlayerScore, error)
}
```

The returned `PlayerScore` slice contains one entry per player with their total score. The scorer never modifies the input state. Each engine also provides `NewFromMap(map[string]any)` for generic instantiation from configuration maps.
