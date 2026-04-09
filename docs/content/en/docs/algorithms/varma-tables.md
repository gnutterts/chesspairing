---
title: "Varma Table Assignment"
linkTitle: "Varma Tables"
weight: 5
description: "Federation-aware pairing number assignment for round-robin tournaments (FIDE C.05 Annex 2)."
---

## The Problem

In a [round-robin tournament](../berger-tables/), the Berger rotation
determines which pairing numbers meet in each round. If two players from the
same federation happen to have adjacent pairing numbers, they will be
scheduled in an early round. Tournament organizers generally prefer to
**spread** same-federation encounters across the schedule so that federation
teammates do not cluster in early or late rounds.

FIDE C.05 Annex 2 defines the Varma table system: a set of lookup tables
that partition pairing number slots into groups, combined with an assignment
algorithm that distributes federations across those groups.

The implementation lives in `algorithm/varma/`.

---

## Varma Groups

For $N$ players (even), the pairing numbers $1, 2, \ldots, N$ are partitioned
into four groups labelled **A**, **B**, **C**, and **D**. The assignment is
defined by lookup tables for each even player count from 10 to 24.

The key property: players assigned to the same Varma group will meet each
other in rounds that are maximally spread across the schedule. Players in
different groups meet in intermediate rounds. By placing all players from a
given federation into the same group (when possible), their mutual games are
spread optimally.

### Table Structure

Each table entry maps a pairing number to its group. For example, with 10
players:

| Pairing Number | 1   | 2   | 3   | 4   | 5   | 6   | 7   | 8   | 9   | 10  |
| -------------- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Group          | A   | B   | C   | D   | D   | C   | B   | A   | A   | B   |

The tables are stored as constant slices in `algorithm/varma/varma.go`.

### Small Tournaments

For $N \leq 8$ players, the Varma tables are trivial: all players are placed
in group A. The federation separation benefit only manifests with 9 or more
players, where the groups have enough members to provide meaningful
separation.

### Odd Player Counts

For odd $N$, the implementation rounds up to $N + 1$ (adding a dummy) and
uses the lookup table for the even count. The dummy player's pairing number
is then filtered out, leaving $N$ assignments. The dummy's slot effectively
becomes the bye position.

---

## The Assignment Algorithm

Given the Varma group table and a list of players with federation labels,
the `Assign` function distributes players to pairing numbers:

### Step 1: Filter Active Players

Remove withdrawn or absent players. Only active players receive pairing
numbers.

### Step 2: Get Group Table

Look up (or compute) the Varma group table for the player count. For counts
above 24, the implementation falls back to a direct assignment without
federation optimization.

### Step 3: Group Players by Federation

Partition players by federation, sorting federations by size (largest first).
This greedy ordering ensures the largest federations get first pick of
groups, maximizing the separation benefit.

### Step 4: Best-Fit Assignment

For each federation (largest first):

1. **Find the best-fit group.** The best-fit group is the one with the most
   remaining slots that is still large enough to hold all players from this
   federation. If no single group fits, the federation is split across
   multiple groups (spilling into the next-best group).

2. **Assign players to slots.** Within each federation, players are ordered
   alphabetically by display name and assigned to the available slots in
   their designated group(s).

The best-fit strategy is a bin-packing heuristic. It does not guarantee a
globally optimal federation separation, but it works well in practice because:

- The largest federations are placed first, getting the most favorable groups.
- Smaller federations fill the remaining gaps.
- The Varma table structure ensures that even suboptimal group choices
  provide reasonable round separation.

### Step 5: Return Ordered Players

The output is the player list ordered by assigned pairing number. This
ordering is then used by the [Berger rotation](../berger-tables/) to
construct the round schedule.

---

## Example

Consider a 12-player tournament with three federations:

- Federation X: 5 players
- Federation Y: 4 players
- Federation Z: 3 players

The Varma table for 12 players has groups of size 3 each (A: 3 slots, B: 3,
C: 3, D: 3).

1. **Federation X** (5 players, largest): no single group of size 3 fits all 5. Assign 3 to group A, spill 2 to group B.
2. **Federation Y** (4 players): group B has 1 remaining slot, too small.
   Assign 3 to group C (perfect fit), spill 1 to group B.
3. **Federation Z** (3 players): group D has 3 slots. Perfect fit.

Result: Group A gets 3 from X; Group B gets 2 from X + 1 from Y; Group C
gets 3 from Y; Group D gets 3 from Z. Same-federation games for Z are
maximally spread (all in group D). Federation X's 5 players span two groups
(A + B), giving moderate but not perfect separation.

---

## Mathematical Justification

The Varma group structure exploits a property of the Berger rotation. In a
schedule with stride $s = n/2 - 1$, two players at positions $p$ and $q$
meet in round:

$$r = \frac{(q - p) \cdot s^{-1} \bmod (n - 1)}{1}$$

where $s^{-1}$ is the modular inverse of $s$ modulo $n - 1$ (which exists
because $\gcd(s, n - 1) = 1$ for the Berger stride). Players whose position
difference $|p - q|$ maps to a round near $\lfloor (n-1)/2 \rfloor$ meet in
the middle of the schedule.

The Varma tables are constructed so that players within the same group have
position differences that produce rounds near the midpoint -- maximizing the
"distance" from round 1 and round $n - 1$.

---

## Limitations

- **Player count range.** Explicit lookup tables exist only for 9--24
  players. For tournaments larger than 24, the Varma assignment falls back to
  a direct ordering without federation optimization.
- **Imperfect separation.** When a federation is larger than any single
  group, its players must span multiple groups, reducing the separation
  benefit.
- **Alphabetical tiebreak.** Within a federation, players are ordered by
  display name. This is a conventional choice with no mathematical
  significance.

---

## Complexity

The assignment algorithm is $O(F \cdot G + N)$ where $F$ is the number of
federations, $G = 4$ is the number of groups, and $N$ is the player count.
The federation sorting is $O(F \log F)$. For realistic tournament sizes, the
entire procedure is effectively $O(N)$.

---

## Related Pages

- [Berger Table Rotation](../berger-tables/) -- the schedule that Varma
  assignment feeds into.
- [Round-Robin Pairing](/docs/pairing-systems/round-robin/) -- the pairing
  system that uses both Varma tables and Berger rotation.
