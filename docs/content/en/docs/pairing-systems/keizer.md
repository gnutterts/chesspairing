---
title: "Keizer System"
linkTitle: "Keizer"
weight: 7
description: "A ranking-based pairing system popular in club play — top-ranked players face each other."
---

The Keizer system pairs players top-down by their current ranking: rank 1 versus rank 2, rank 3 versus rank 4, and so on. In round 1, ranking is determined by rating. From round 2 onward, ranking is determined by the Keizer score computed by the internal Keizer scorer. This tight coupling between pairing and scoring is a defining characteristic of the Keizer system. It is not a FIDE system but is widely used in club chess in Belgium and the Netherlands.

## When to Use

Keizer is appropriate when:

- The tournament is a club competition running over many weeks with irregular attendance (the Keizer scoring system handles absences gracefully).
- You want the strongest active players to face each other every round, producing competitive top-board matchups.
- Repeat pairings are acceptable (or even desirable) in long-running tournaments.
- The tournament uses Keizer scoring, since the pairer depends on the scorer for ranking.

It is not suitable for FIDE-rated events that require an official Swiss system, or for short tournaments where players expect to face different opponents each round.

## Configuration

### CLI

```bash
chesspairing pair --keizer tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/keizer"

// With typed options
p := keizer.New(keizer.Options{
    AllowRepeatPairings:     chesspairing.BoolPtr(true),
    MinRoundsBetweenRepeats: chesspairing.IntPtr(3),
})

// From a generic map (JSON config)
p := keizer.NewFromMap(map[string]any{
    "allowRepeatPairings":     true,
    "minRoundsBetweenRepeats": 5,
})
```

### Options Reference

| Option                    | Type                     | Default | Description                                                                                                                             |
| ------------------------- | ------------------------ | ------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `allowRepeatPairings`     | `bool`                   | `true`  | Whether players can be paired against the same opponent again. When `false`, no repeat pairings are ever allowed.                       |
| `minRoundsBetweenRepeats` | `int`                    | `3`     | Minimum number of rounds that must pass before two players can be paired again. Only applies when `allowRepeatPairings` is `true`.      |
| `scoringOptions`          | `scoring/keizer.Options` | nil     | Configuration for the internal Keizer scorer used for ranking. When nil, the scorer uses its own defaults (24 configurable parameters). |

The `scoringOptions` field accepts the full set of Keizer scoring parameters. Any Keizer scoring option can also be set at the top level of the options map -- `ParseOptions` forwards unrecognized keys to the scoring parser.

## How It Works

### 1. Rank Players

In round 1, players are ranked by rating descending (alphabetical name as tiebreaker). From round 2 onward, the engine instantiates an internal Keizer scorer, runs `Score()` on the current tournament state, and ranks players by their Keizer score descending. Rating is the secondary tiebreaker, and display name is the tertiary tiebreaker.

If scoring fails for any reason, the engine falls back to rating-based ranking.

### 2. Build Pairing History

The engine scans all completed rounds to build a map of which round each pair of players last faced each other. Forfeit games are excluded from this history -- per the project convention, a forfeited game does not count as a real encounter, so the two players can be paired again.

### 3. Pair Top-Down

Players are paired sequentially from the top of the ranking:

- Rank 1 vs Rank 2
- Rank 3 vs Rank 4
- Rank 5 vs Rank 6
- ...and so on

If the player count is odd, the lowest-ranked player receives a pairing-allocated bye.

### 4. Repeat Avoidance

When a proposed pairing would violate the repeat rules, the engine swaps the lower-ranked player in the pair with the nearest available lower-ranked player who is a legal opponent:

- If `allowRepeatPairings` is `false`, any previous encounter is a conflict.
- If `allowRepeatPairings` is `true`, the pairing is only blocked if the last encounter was fewer than `minRoundsBetweenRepeats` rounds ago.

The swap search proceeds downward through the ranking until a compatible partner is found. If no swap is possible, the repeat pairing stands and a note is added to the result.

### 5. Colour Assignment

Color allocation delegates to the same `swisslib.AllocateColor` function used by the Dutch, Burstein, and Dubov systems. The full 6-step priority cascade applies: compatible preferences, absolute wins, strong beats non-strong, first color difference in history, rank tiebreak, and board alternation. See [Color Allocation](/docs/algorithms/color-allocation/) for the detailed algorithm.

Forfeit games do not contribute to colour history. Byes produce a `ColorNone` entry, which is ignored by the preference computation.

## Comparison

| Aspect             | Keizer                          | Dutch Swiss             | Round-Robin                   |
| ------------------ | ------------------------------- | ----------------------- | ----------------------------- |
| Pairing method     | Top-down by score               | Global Blossom matching | Berger table rotation         |
| Repeat pairings    | Allowed (configurable)          | Never                   | Every pair plays exactly once |
| Scoring dependency | Tightly coupled (Keizer scorer) | Independent             | Independent                   |
| Colour allocation  | swisslib 6-step cascade         | 5+ step FIDE rules      | Berger table convention       |
| FIDE regulated     | No                              | Yes (C.04.3)            | Yes (C.05 Annex 1)            |
| Typical use        | Club play, long events          | Open tournaments        | Small closed events           |
| Bye assignment     | Lowest-ranked player            | Completability-based    | Berger table dummy            |

The Keizer system is fundamentally different from Swiss systems. Swiss systems aim to pair players with similar scores while avoiding repeat encounters. Keizer aims to pair the highest-ranked players against each other every round, deliberately creating top-heavy matchups. The ranking evolves each round as Keizer scores change, so a player who loses drops in the ranking and faces weaker opponents next round.

## Mathematical Foundations

### Ranking Function

The ranking for round r is defined as:

```text
rank(p, 1) = sort by rating descending
rank(p, r) = sort by keizerScore(p, r-1) descending    for r > 1
```

where `keizerScore` is computed by the Keizer scorer using iterative convergence (see the [Keizer scoring documentation](/docs/scoring/keizer/) for details on the scoring algorithm).

### Repeat Avoidance Constraint

For two players a and b with last encounter in round L, the pairing is allowed in round R if:

```text
allowRepeatPairings = false:  (a, b) never played
allowRepeatPairings = true:   R - L >= minRoundsBetweenRepeats
```

### Swap Distance

When a conflict occurs at position i (rank i vs rank i+1), the engine searches for the smallest j > i+1 such that `canPair(ranked[i], ranked[j])` is true. The swap replaces ranked[i+1] with ranked[j] in the pairing, maintaining the relative order of all other players. This greedy approach minimizes the disruption to the ranking-based pairing ideal.

## FIDE Reference

The Keizer system is not governed by FIDE regulations. It originated in the Netherlands and is primarily used in club competitions in Belgium and the Netherlands. There is no FIDE handbook article for Keizer pairing or scoring.

The system is sometimes called "Keizer-Sonneborn" in Dutch-language chess literature, though it should not be confused with the Sonneborn-Berger tiebreaker. The scoring method is described in the [Keizer scoring documentation](/docs/scoring/keizer/).
