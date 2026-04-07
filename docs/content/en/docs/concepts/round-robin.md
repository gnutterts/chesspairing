---
title: "Round-Robin Tournaments"
linkTitle: "Round-Robin"
weight: 2
description: "Every player meets every other player — scheduling with Berger tables and handling odd counts."
---

## Everyone plays everyone

In a round-robin tournament, every player plays every other player
exactly once. There is no matchmaking algorithm, no score groups, and no
floaters -- the entire schedule is determined before the first move. A
tournament with N players needs N-1 rounds (or N rounds if N is odd,
since one player sits out each round).

Round-robin is the gold standard for determining the strongest player
when the field is small enough. It is the format of choice for
Candidates tournaments, national championships, and closed
invitationals where the number of participants is manageable.

## Single and double round-robin

A **single round-robin** (one cycle) gives each pair of players one
game. The Cycles option controls how many times the full schedule is
repeated:

- **Cycles: 1** -- single round-robin. N-1 rounds for N players.
- **Cycles: 2** -- double round-robin. Each pair plays twice with
  reversed colors. 2(N-1) rounds total.

Double round-robin is common when the field is very small (4-8 players)
and the organiser wants more decisive results. In the second cycle,
colors are reversed so that every player gets one game as White and one
as Black against each opponent.

## Berger tables

The schedule is built using FIDE Berger tables (C.05 Annex 1). The
algorithm works as follows:

1. Fix the last player (player N) at a constant position.
2. Rotate the remaining N-1 players through the other positions with a
   fixed stride.
3. Each rotation produces one round of pairings: the player at
   position 0 plays the player at position N-1, position 1 plays
   position N-2, and so on.

This produces a clean schedule where every player meets every other
player exactly once, and color assignments follow directly from the
table positions. Board 1 (the fixed player's board) alternates colors
each round; the other boards assign White to the player in the lower
position index.

For a deeper look at how the tables are constructed, see
[Berger Tables](/docs/algorithms/berger-tables/).

## Odd player counts

When the player count is odd, a dummy "bye" slot is added to make the
count even. In each round, the player paired against the dummy receives
a pairing-allocated bye (PAB) instead of playing. The Berger table
rotation ensures that every player gets exactly one bye over the course
of the cycle.

## Color balancing

Within a single cycle, the Berger table structure naturally distributes
colors fairly. In a double round-robin, the second cycle reverses all
color assignments (when the ColorBalance option is enabled, which it is
by default), so each player gets one White and one Black game against
every opponent.

### Last-two-round swap

In a double round-robin, the boundary between cycle 1 and cycle 2 can
cause a player to have the same color three times in a row (last round
of cycle 1, plus the first two rounds of cycle 2 with reversed colors).
To prevent this, chesspairing swaps the last two rounds of the first
cycle by default (the SwapLastTwoRounds option). This is the standard
FIDE recommendation for double round-robin events.

## Varma tables

When a round-robin includes players from many federations, the initial
player numbering affects who plays whom in which round. FIDE's Varma
tables (C.05 Annex 2) provide a federation-aware method for assigning
pairing numbers so that players from the same federation meet as late in
the tournament as possible, spreading out the "internal" matchups.

chesspairing implements Varma table assignment through the `algorithm/varma`
package. See [Varma Tables](/docs/algorithms/varma-tables/) for details.

## Further reading

- [Round-Robin pairing system reference](/docs/pairing-systems/round-robin/)
- [Berger Tables algorithm](/docs/algorithms/berger-tables/)
- [Varma Tables algorithm](/docs/algorithms/varma-tables/)
