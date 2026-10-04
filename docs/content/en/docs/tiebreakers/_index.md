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

The Chief Organiser chooses the tie-break order in the tournament regulations (FIDE C.07:2026 Article 4.1). `DefaultTiebreakers()` supplies library defaults, which can be replaced in the configuration:

- **Swiss tournaments**: Buchholz Cut-1, Buchholz, Sonneborn-Berger, Direct Encounter
- **Team Swiss**: Buchholz MP Cut-1, Buchholz MP, Extended Sonneborn-Berger (MP/MP), Match Points or Game Points
- **Round-Robin**: Sonneborn-Berger, Direct Encounter, Games Won, Koya
- **Keizer**: Games Played, Direct Encounter, Games Won

For events with many tied players, Buchholz variants are the most discriminating because they incorporate the entire tournament's result network. Performance-based tiebreakers (TPR, PTP) are useful in large opens where rating-based strength measurement is meaningful. Head-to-head tiebreakers like Direct Encounter are decisive when a small group of players is tied.

## Unplayed rounds

Opponent-based tie-breakers build one record for every player and round. In Swiss tournaments, FIDE C.07:2026 Article 16 classifies unplayed rounds, adjusts opponents' scores where required, and uses capped dummy opponents for the participant's own calculation. In round robins and other predetermined pairings, Article 15.2 instead treats forfeits as regular encounters except for ratings-based tie-breaks and Type-B forfeit losses. Pending games do not count as games played over the board.

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

The `tiebreaker.All()` function returns the registered individual names,
`tiebreaker.TeamAll()` returns the team-only names, and the CLI's
`tiebreakers` subcommand lists both with descriptions.

## Team competitions

For team scoring, `mpvgp` supplies the secondary MP or GP score. The extended
Sonneborn-Berger variants are `emmsb`, `emgsb`, `egmsb`, and `eggsb`; use
`emmsb-cut1` for its Cut-1 variant. `buchholz-mp` and `buchholz-mp-cut1`
reference opponents' MP. Board-level matches additionally support
`board-count`, `top-board-results`, and `bottom-board-elimination`. See FIDE
C.07:2026 Articles 12 and 13.


## Interface

Every tiebreaker implements:

```go
type TieBreaker interface {
    Compute(ctx context.Context, state TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
```

The `scores` parameter provides the current standings (from any scoring engine). The returned `TieBreakValue` slice contains one entry per player with the computed tiebreak value. Tiebreakers never modify the input state or scores.

## Registered IDs

`aro`, `aro-cut1`, `avg-opponent-buchholz`, `avg-opponent-fore-buchholz`, `avg-opponent-ptp`, `avg-opponent-tpr`, `black-games`, `black-wins`, `board-count`, `bottom-board-elimination`, `buchholz`, `buchholz-cut1`, `buchholz-cut2`, `buchholz-median`, `buchholz-median2`, `buchholz-mp`, `buchholz-mp-cut1`, `direct-encounter`, `eggsb`, `egmsb`, `emgsb`, `emmsb`, `emmsb-cut1`, `fore-buchholz`, `games-played`, `ge`, `koya`, `mpvgp`, `pairing-number`, `performance-points`, `performance-rating`, `player-rating`, `progressive`, `progressive-cut1`, `rounds-played`, `sonneborn-berger`, `sonneborn-berger-cut1`, `standard-points`, `top-board-results`, `win`, `wins`
