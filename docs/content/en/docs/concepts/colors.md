---
title: "Colors and Balance"
linkTitle: "Colors & Balance"
weight: 5
description: "Why color allocation matters and how pairing systems balance white and black assignments."
---

## Why color matters

White moves first in chess. Statistically, White wins more often than
Black -- around 55% in master-level games. While this edge is small in
any single game, a player who repeatedly gets the same color is at a
systematic advantage or disadvantage over the tournament. Fair color
distribution is a core responsibility of every pairing system.

## Color preferences

After each round, a player accumulates a color history. From that
history, the pairing system computes a **color preference** -- the
color the player should ideally receive next. Preferences come in
three strengths:

- **Absolute preference.** The player _must_ receive this color.
  Triggered when their color imbalance exceeds 1 (e.g., three Whites
  and one Black) or when they have played the same color in the last
  two consecutive rounds. Violating an absolute preference would create
  three same-color games in a row, which is forbidden in most systems.

- **Strong preference.** The player _should_ receive this color. Their
  color counts are unequal (e.g., two Whites and one Black) but not
  critically so. The pairer will try to satisfy this, but can override
  it if necessary.

- **Mild preference.** The player _would like_ this color, based on
  alternation (the opposite of what they played last round). This is
  the weakest preference and is the first to be sacrificed when
  constraints conflict.

A player with no games yet has no color preference.

## The goal

The color allocation system pursues two goals simultaneously:

1. **Alternate colors** from round to round. If you played White last
   round, you should play Black this round.
2. **Equalise color counts** over the tournament. Your total White
   games and total Black games should stay as close as possible.

These two goals usually align but can conflict. When they do, avoiding
three consecutive same-color games takes priority over equalisation.

## Color allocation happens after pairing

An important architectural point: **color allocation is a separate step
that runs after the pairer has determined who plays whom.** The pairing
algorithms consider color constraints when deciding whether two players
_can_ be paired (an absolute color conflict makes a pairing illegal),
but the actual White/Black assignment for each board happens afterward.

This separation keeps the pairing logic focused on the
constraint-satisfaction problem (who plays whom) while delegating the
color-assignment problem to a dedicated algorithm.

## How each system handles color

### Dutch, Burstein, and Dubov

These three systems share the same color allocation code in the
`swisslib` package. The algorithm follows a 6-step priority:

1. **Compatible preferences.** If one player wants White and the other
   wants Black (or has no preference), both are satisfied.
2. **Absolute preference wins.** If only one player has an absolute
   preference, or one has a stronger imbalance, that player gets their
   preferred color.
3. **Strong preference wins.** If one player has a strong preference
   and the other does not, the strong preference is granted.
4. **Color history tiebreak.** Walk backward through both players'
   color histories and find the most recent round where they had
   different colors. Swap based on that difference.
5. **Rank tiebreak.** If both players want the same color with equal
   strength and identical history, the higher-ranked player gets their
   preference.
6. **Board alternation.** If neither player has any preference (e.g.,
   round 1), alternate by board number: higher-ranked player gets White
   on odd boards, Black on even boards. The TopSeedColor option can
   invert this pattern.

### Lim

The Lim system uses a round-parity approach:

- **Even rounds** aim to _equalise_ color counts (if you have played
  more Whites, you get Black).
- **Odd rounds** aim to _alternate_ (opposite of your last color).
- When both players want the same color, a **median tiebreak** decides:
  in the upper half of the standings, the higher-ranked player gets
  their preference; in the lower half, the lower-ranked player does.
- The mandatory rule (no three consecutive same-color games) always
  takes precedence.

### Double-Swiss

In Double-Swiss, each round is a two-game match where colors
automatically alternate between games. "Color" here refers to who gets
White in Game 1. The system enforces a hard constraint: **no player may
have the same Game-1 color three rounds in a row.** Beyond that, it
follows a 5-step priority: equalise, alternate, rank tiebreak, and
board alternation.

### Team Swiss

Team Swiss uses a 9-step color allocation process, the most complex
of any system. It introduces the **first-team concept**: the team with
the higher score (or higher secondary score, or lower pairing number)
is the "first team" and gets priority in tiebreaks.

The 9 steps handle: initial round assignment, granting single
preferences, satisfying opposite preferences, strong vs. mild (for
Type B preference mode), color difference comparison, alternation from
the most recent divergent round, first-team preference, first-team
alternation, and other-team alternation.

### Keizer

Keizer delegates color allocation to the same `swisslib` code used by
the Dutch, Burstein, and Dubov systems. The full 6-step priority cascade
applies: compatible preferences, absolute wins, strong beats non-strong,
color history tiebreak, rank tiebreak, and board alternation. Forfeit
games are excluded from color history; byes produce a neutral entry.

### Round-Robin

In a round-robin, colors are determined by the Berger table itself.
Board 1 (involving the fixed player) alternates each round. All other
boards assign White to the player in the lower table position. In a
double round-robin with color balancing enabled, the second cycle
reverses all color assignments so that every pair gets one game each
way.

## Further reading

For the mathematical details behind each system's color algorithm, see
[Color Allocation](/docs/algorithms/color-allocation/).
