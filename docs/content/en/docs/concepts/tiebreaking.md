---
title: "Tiebreaking"
linkTitle: "Tiebreaking"
weight: 4
description: "When players share the same score, tiebreakers determine who ranks higher."
---

## Why tiebreakers exist

In a Swiss tournament with 40 players and 7 rounds, it is common for
several players to finish on the same score. Two players on 5.5/7 need
to be ranked -- who gets the trophy? Tiebreakers answer that question
by computing a secondary value for each player that distinguishes
otherwise equal scores.

## How tiebreakers are applied

Tiebreakers are configured as an ordered list. The scoring engine
computes every tiebreaker for every player, then applies them in
sequence:

1. Players are first ranked by score.
2. Among players with equal scores, the first tiebreaker is compared.
3. If the first tiebreaker is also equal, the second tiebreaker is
   compared.
4. This continues until a difference is found or all tiebreakers are
   exhausted.

The order matters. Placing Buchholz first means "strength of opponents"
is valued above all else; placing Direct Encounter first means
head-to-head results take priority.

## Categories of tiebreakers

chesspairing provides 25 tiebreakers, which fall into six categories:

### Opponent-based

These measure how strong your opponents were. The logic: beating
opponents who themselves scored well is more impressive than beating
opponents who scored poorly.

- **Buchholz** -- sum of all opponents' final scores. The most widely
  used Swiss tiebreaker.
- **Buchholz Cut-1** -- Buchholz minus the lowest opponent score.
  Reduces the penalty for one weak opponent.
- **Buchholz Cut-2** -- Buchholz minus the two lowest opponent scores.
- **Buchholz Median** -- Buchholz minus the highest and lowest opponent
  scores. Removes outliers in both directions.
- **Buchholz Median-2** -- Buchholz minus the two highest and two
  lowest.
- **Fore Buchholz** -- Buchholz where pending (unplayed) games are
  treated as draws. Useful for mid-tournament standings.
- **Average Opponent Buchholz** -- Buchholz divided by games played.

### Performance-based

These estimate how well you played relative to your rating.

- **Performance Rating (TPR)** -- the rating at which your result would
  be expected, computed from the FIDE B.02 conversion table.
- **Performance Points (PTP)** -- expected score based on rating
  differences, using FIDE expected-score tables.
- **Average Rating of Opponents (ARO)** -- the mean rating of your
  opponents.
- **Average Opponent TPR (APRO)** -- the mean performance rating of
  your opponents.
- **Average Opponent PTP (APPO)** -- the mean performance points of
  your opponents.

### Result-based

These focus on the quality of your individual results.

- **Games Won** -- number of over-the-board wins (excludes forfeits).
- **Rounds Won** -- number of wins including forfeit wins and PAB byes.
- **Progressive Score** -- cumulative (running) score after each round.
  Rewards early wins more than late wins.

### Head-to-head

These look at the results between the specific players who are tied.

- **Direct Encounter** -- the score between the tied players in their
  games against each other. Only meaningful when tied players actually
  played.
- **Sonneborn-Berger** -- for each opponent, multiply your result
  against them by their final score, then sum. Rewards wins against
  high-scoring opponents and penalises draws against low-scoring ones.
- **Koya System** -- your score against opponents who finished in the
  top half of the standings. Common in round-robin events.

### Activity-based

These reflect participation level.

- **Games Played** -- total number of games played (excludes forfeits).
- **Games with Black** -- number of games played as Black (excludes
  forfeits). Tiebreaks in favour of players who had a harder color
  schedule.

### Ordering

These provide deterministic final tiebreaking when all other tiebreakers
are equal.

- **Pairing Number** -- the player's tournament pairing number (TPN).
  Lower number = higher rank.
- **Player Rating** -- the player's rating. Higher = better.

## Default tiebreakers by system

Each pairing system has a recommended default tiebreaker sequence. These
defaults are returned by `DefaultTiebreakers()` and used when no
explicit tiebreaker list is configured:

**Swiss systems** (Dutch, Burstein, Dubov, Lim, Double-Swiss, Team):

1. Buchholz Cut-1
2. Buchholz
3. Sonneborn-Berger
4. Direct Encounter

**Round-Robin:**

1. Sonneborn-Berger
2. Direct Encounter
3. Games Won
4. Koya

**Keizer:**

1. Games Played
2. Direct Encounter
3. Games Won

These defaults follow FIDE recommendations. You can override them with
any combination of the 25 available tiebreakers.

## The tiebreaker registry

All tiebreakers self-register at startup via Go's `init()` mechanism.
You look them up by string ID:

```go
tb, err := tiebreaker.Get("buchholz-cut1")
values, err := tb.Compute(ctx, state, scores)
```

The 25 registered IDs are: `buchholz`, `buchholz-cut1`, `buchholz-cut2`,
`buchholz-median`, `buchholz-median2`, `sonneborn-berger`,
`direct-encounter`, `wins`, `win`, `black-games`, `black-wins`,
`rounds-played`, `standard-points`, `pairing-number`, `koya`,
`progressive`, `aro`, `fore-buchholz`, `avg-opponent-buchholz`,
`performance-rating`, `performance-points`, `avg-opponent-tpr`,
`avg-opponent-ptp`, `player-rating`, `games-played`.

## Further reading

- [Buchholz tiebreakers](/docs/tiebreakers/buchholz/)
- [Performance-based tiebreakers](/docs/tiebreakers/performance/)
- [Result-based tiebreakers](/docs/tiebreakers/results/)
- [Head-to-head tiebreakers](/docs/tiebreakers/head-to-head/)
- [Color, activity & ordering tiebreakers](/docs/tiebreakers/color-activity/)
