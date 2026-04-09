---
title: "Berger Table Rotation"
linkTitle: "Berger Tables"
weight: 4
description: "FIDE Berger tables for round-robin scheduling — the rotation algorithm and last-two-round swap."
---

## The Scheduling Problem

A **round-robin** tournament requires every player to play every other player
exactly once (single round-robin) or exactly twice (double round-robin). For
$N$ players, a single round-robin has $\binom{N}{2} = \frac{N(N-1)}{2}$
games distributed across $N - 1$ rounds (or $N$ rounds if $N$ is odd, with
one bye per round).

Johann Berger published a systematic schedule construction in 1895 that FIDE
adopted as the standard (C.05 Annex 1). The method fixes one player in place
and rotates all others, producing a balanced schedule with good color
alternation properties.

The implementation lives in `pairing/roundrobin/roundrobin.go`.

---

## Setup

Let $N$ be the number of players. If $N$ is odd, add a **dummy player**
(numbered $N$) to make the count even; any player paired against the dummy
receives a bye. Set $n = N$ if even, $n = N + 1$ if odd.

Number the positions $0, 1, 2, \ldots, n - 1$. Player at position $n - 1$ is
**fixed** (the "pivot"). The remaining $n - 1$ players rotate.

---

## Rotation Formula

For round $r$ (0-indexed), the player at position $j$ in round $r$ is the
player who was originally at position:

$$\text{positions}[j] = (j - r \cdot s) \bmod (n - 1) \quad \text{for } j < n - 1$$

where the **stride** is:

$$s = \frac{n}{2} - 1$$

The player at position $n - 1$ does not move. After $n - 1$ rounds, every
non-fixed position has been visited by every player exactly once, and every
pair of players has been scheduled exactly once.

### Why this stride?

The stride $s = n/2 - 1$ is chosen so that each rotation moves players
approximately half the table forward. This maximizes color alternation: a
player who had White in one round is likely to have Black in the next, because
they move to the opposite side of the pairing table.

The choice of stride is not unique -- any value coprime to $n - 1$ produces a
valid schedule. The Berger stride $n/2 - 1$ is the standard because it
optimizes color balance properties.

---

## Pairing Construction

In each round $r$, pair the players at positions as follows:

1. **Board 1**: position $0$ versus position $n - 1$ (the fixed player).
2. **Board $k$** (for $k = 2, 3, \ldots, n/2$): position $k - 1$ versus
   position $n - 1 - (k - 1) = n - k$.

This gives $n/2$ boards per round. If $N$ was odd, the player paired against
the dummy receives a bye instead of a game.

---

## Color Assignment

Colors are assigned per FIDE convention:

- **Board 1**: the rotating player (position $0$) alternates color each
  round. In round 0 they play White; in round 1, Black; and so on.
  Equivalently, the fixed player (position $n - 1$) plays Black in even
  rounds and White in odd rounds.
- **Other boards**: the player at the lower position index plays White.

Formally, for board $k > 1$ with players at positions $a < b$:

$$\text{White} = \text{player at position } a, \quad \text{Black} = \text{player at position } b$$

For board 1 in round $r$:

$$\text{White} = \begin{cases} \text{rotating player} & \text{if } r \text{ is even} \\ \text{fixed player} & \text{if } r \text{ is odd} \end{cases}$$

---

## Double Round-Robin

A double round-robin consists of two **cycles**. In cycle 2, every pairing
from cycle 1 is repeated with reversed colors:

- If player $A$ had White against player $B$ in cycle 1, then $B$ has White
  against $A$ in cycle 2.

The round numbering continues: cycle 1 uses rounds $0, 1, \ldots, n - 2$ and
cycle 2 uses rounds $n - 1, n, \ldots, 2(n - 1) - 1$.

### The Last-Two-Round Swap

A problem arises at the cycle boundary. In the last round of cycle 1 and the
first round of cycle 2, the same player pairs occur with reversed colors. A
player who had White in round $n - 2$ would have Black in round $n - 1$ with
the same opponent, then immediately face a different opponent in round $n$
with the color that continues from round $n - 2$. This can create sequences
of three consecutive games with the same color.

The FIDE-recommended fix is to **swap the last two rounds of cycle 1** (not
cycle 2). When the `SwapLastTwoRounds` option is enabled:

- Round $n - 3$ gets the pairings originally scheduled for round
  $n - 2$.
- Round $n - 2$ gets the pairings originally scheduled for round
  $n - 3$.

This disrupts the three-consecutive-color pattern at the cost of a minor
schedule irregularity in the final two rounds of the first cycle.

---

## Worked Example: 6 Players

With $N = 6$ (even), $n = 6$, stride $s = 6/2 - 1 = 2$.

Positions in round 0: players 0, 1, 2, 3, 4 rotate; player 5 is fixed.

| Round | Positions after rotation | Board 1 | Board 2 | Board 3 |
| ----- | ------------------------ | ------- | ------- | ------- |
| 0     | 0 1 2 3 4 **5**          | 0 vs 5  | 1 vs 4  | 2 vs 3  |
| 1     | 3 4 0 1 2 **5**          | 3 vs 5  | 4 vs 2  | 0 vs 1  |
| 2     | 1 2 3 4 0 **5**          | 1 vs 5  | 2 vs 0  | 3 vs 4  |
| 3     | 4 0 1 2 3 **5**          | 4 vs 5  | 0 vs 3  | 1 vs 2  |
| 4     | 2 3 4 0 1 **5**          | 2 vs 5  | 3 vs 1  | 4 vs 0  |

Every pair appears exactly once across 5 rounds. Board 1 alternates the
color of the rotating player: White in rounds 0, 2, 4; Black in rounds 1, 3.

---

## Odd Player Count

For $N = 5$, add dummy player 5 to get $n = 6$. The schedule is identical to
the example above, but any game involving player 5 becomes a bye for the
opponent:

- Round 0: player 0 has a bye (was paired against dummy 5).
- Round 1: player 2 has a bye.
- And so on.

Each player receives exactly one bye across the tournament.

---

## Properties

**Completeness.** Every pair of real players is scheduled exactly once per
cycle. This follows from the rotation being a cyclic permutation of order
$n - 1$ on the non-fixed positions.

**Color balance.** Each player gets at most $\lceil (n-1)/2 \rceil$ games
with one color and at least $\lfloor (n-1)/2 \rfloor$ with the other. The
fixed player alternates perfectly. Rotating players have near-perfect
alternation due to the stride choice.

**Round count.** Single round-robin: $n - 1$ rounds. Double round-robin:
$2(n - 1)$ rounds.

---

## Complexity

The schedule construction is $O(n^2)$ -- each of the $n - 1$ rounds produces
$n/2$ pairings, giving $O(n^2)$ total operations. The rotation itself is
$O(1)$ per player per round (a modular arithmetic operation).

---

## Related Pages

- [Round-Robin Pairing](/docs/pairing-systems/round-robin/) -- the pairing
  system that uses Berger tables.
- [Varma Tables](../varma-tables/) -- federation-aware pairing number
  assignment that determines _which_ player sits at _which_ position before
  the Berger rotation begins.
