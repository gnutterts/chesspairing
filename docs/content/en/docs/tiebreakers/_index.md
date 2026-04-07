---
title: "Tiebreakers"
linkTitle: "Tiebreakers"
weight: 50
description: "25 tiebreaker implementations grouped by category — from Buchholz variants to performance ratings."
---

Chesspairing provides 25 tiebreakers grouped into seven categories. Each tiebreaker implements the `TieBreaker` interface and computes a single numeric value per player from the tournament state and standings. Tiebreakers self-register via a central registry, so they can be selected by name at runtime.

## Categories

| Category                                | Tiebreakers                                                                | What they measure                                   |
| --------------------------------------- | -------------------------------------------------------------------------- | --------------------------------------------------- |
| [Buchholz](buchholz/)                   | buchholz, buchholz-cut1, buchholz-cut2, buchholz-median, buchholz-median2  | Sum of opponents' scores (with cut/median variants) |
| [Performance](performance/)             | performance-rating, performance-points, avg-opponent-tpr, avg-opponent-ptp | Strength of play relative to opposition             |
| [Results](results/)                     | wins, win, standard-points, progressive, rounds-played, games-played       | Direct measures of game outcomes                    |
| [Head-to-Head](head-to-head/)           | direct-encounter, sonneborn-berger, koya                                   | Results among tied or top-half opponents            |
| [Color & Activity](color-activity/)     | black-games, black-wins                                                    | Colour distribution metrics                         |
| [Ordering](ordering/)                   | pairing-number, player-rating                                              | Static player attributes for final tiebreaking      |
| [Opponent Buchholz](opponent-buchholz/) | fore-buchholz, avg-opponent-buchholz                                       | Buchholz-derived metrics of opponent quality        |

## Choosing tiebreakers

FIDE regulations recommend specific tiebreaker sequences depending on the tournament format. The `DefaultTiebreakers()` function in the root package returns the FIDE-recommended sequence for each pairing system. Typical choices:

- **Swiss tournaments**: Buchholz Cut 1, Buchholz, Sonneborn-Berger, Progressive
- **Round-Robin**: Sonneborn-Berger, Direct Encounter, Wins, Games with Black
- **Keizer**: The Keizer score itself is the primary ranking; additional tiebreakers are rarely needed

For events with many tied players, Buchholz variants are the most discriminating because they incorporate the entire tournament's result network. Performance-based tiebreakers (TPR, PTP) are useful in large opens where rating-based strength measurement is meaningful. Head-to-head tiebreakers like Direct Encounter are decisive when a small group of players is tied.

## Forfeit exclusion

All tiebreakers that examine opponent data use the shared `buildOpponentData` function, which excludes forfeited games from the opponent list. This means:

- A forfeit win does not add the absent opponent to your Buchholz calculation.
- A double forfeit is excluded from both players' tiebreak computations entirely.
- Only games actually played over the board (including draws) contribute to opponent-based tiebreakers.

This matches FIDE tiebreaker regulations, which treat forfeits as non-games for tiebreaking purposes.

## Registry

Tiebreakers self-register via `init()` functions:

```go
func init() {
    Register("buchholz", func() chesspairing.TieBreaker {
        return &Buchholz{variant: buchholzFull}
    })
}
```

At runtime, retrieve a tiebreaker by its registered name:

```go
tb, err := tiebreaker.Get("buchholz-cut1")
if err != nil {
    // unknown tiebreaker name
}
values, err := tb.Compute(ctx, state, scores)
```

The `tiebreaker.All()` function returns all 25 registered names, and the CLI's `tiebreakers` subcommand lists them with descriptions.

## Interface

Every tiebreaker implements:

```go
type TieBreaker interface {
    Compute(ctx context.Context, state TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
```

The `scores` parameter provides the current standings (from any scoring engine). The returned `TieBreakValue` slice contains one entry per player with the computed tiebreak value. Tiebreakers never modify the input state or scores.

## All 25 registered IDs

`buchholz`, `buchholz-cut1`, `buchholz-cut2`, `buchholz-median`, `buchholz-median2`, `sonneborn-berger`, `direct-encounter`, `wins`, `win`, `black-games`, `black-wins`, `rounds-played`, `standard-points`, `pairing-number`, `koya`, `progressive`, `aro`, `fore-buchholz`, `avg-opponent-buchholz`, `performance-rating`, `performance-points`, `avg-opponent-tpr`, `avg-opponent-ptp`, `player-rating`, `games-played`
