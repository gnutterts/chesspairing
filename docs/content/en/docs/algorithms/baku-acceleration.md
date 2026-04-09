---
title: "Baku Acceleration"
linkTitle: "Acceleration"
weight: 6
description: "Virtual points in early rounds to prevent predictable pairings (FIDE C.04.7)."
---

## Motivation

In a standard Swiss tournament, the first round pairs player 1 against player
$\lceil N/2 \rceil + 1$, player 2 against $\lceil N/2 \rceil + 2$, and so
on. After round 1, the top-half winners all have 1 point and are paired among
themselves in round 2. This creates a predictable pattern where the strongest
players face each other very early, producing decisive results that can make
later rounds less competitive.

**Baku acceleration** (FIDE C.04.7, named after the 2016 Chess Olympiad in
Baku where it was first used at a major event) breaks this pattern by adding
**virtual points** to a subset of players in early rounds. These virtual
points inflate the score groups, mixing players from different rating tiers
into the same bracket. After the acceleration phase ends, the virtual points
are removed and real scores take over.

The implementation lives in `pairing/swisslib/acceleration.go`.

---

## Definitions

Given a tournament with $R$ total rounds and $N$ active players, Baku
acceleration defines four parameters:

### Accelerated Rounds

$$\text{accelerated} = \left\lceil \frac{R}{2} \right\rceil$$

The total number of rounds during which acceleration is active.

### Full Virtual Point Rounds

$$\text{fullVP} = \left\lceil \frac{\text{accelerated}}{2} \right\rceil$$

Rounds $1, 2, \ldots, \text{fullVP}$ award 1.0 virtual points to eligible
players.

### Half Virtual Point Rounds

$$\text{halfVP} = \text{accelerated} - \text{fullVP}$$

Rounds $\text{fullVP} + 1, \ldots, \text{accelerated}$ award 0.5 virtual
points to eligible players.

### Group A Size

$$\text{gaSize} = 2 \cdot \left\lceil \frac{N}{4} \right\rceil$$

The number of players in "Group A" -- the set of players who receive virtual
points. Group A consists of the top-ranked players (those with initial rank
$\leq \text{gaSize}$). The formula ensures Group A is always even-sized.

---

## Virtual Points Function

For player $p$ in round $r$ (1-indexed):

$$\text{VP}(p, r) = \begin{cases} 1.0 & \text{if } \text{rank}(p) \leq \text{gaSize} \text{ and } r \leq \text{fullVP} \\ 0.5 & \text{if } \text{rank}(p) \leq \text{gaSize} \text{ and } \text{fullVP} < r \leq \text{accelerated} \\ 0.0 & \text{otherwise} \end{cases}$$

The virtual points are added to the player's **pairing score** (the score
used for bracket assignment), not their actual tournament score. This means:

- During accelerated rounds, Group A players appear to have higher scores
  than they actually do, placing them in higher brackets.
- Tiebreakers and final standings use the real scores, not the inflated
  pairing scores.
- After round $\text{accelerated}$, all virtual points are zero and pairing
  proceeds normally.

---

## Effect on Brackets

### Without Acceleration

In a 100-player, 9-round tournament after round 1:

- Score group 1.0: ~50 players (all winners)
- Score group 0.5: ~0 players (assuming no draws for simplicity)
- Score group 0.0: ~50 players (all losers)

Round 2 pairs the 50 winners among themselves: player 1 vs ~player 25, player
2 vs ~player 26, etc. The top seeds face strong opposition immediately.

### With Acceleration

The same tournament has $\text{gaSize} = 2 \cdot \lceil 100/4 \rceil = 50$
and $\text{fullVP} = \lceil \lceil 9/2 \rceil / 2 \rceil = 3$. In round 1:

- Group A players (ranks 1--50) have pairing score $0.0 + 1.0 = 1.0$.
- Group B players (ranks 51--100) have pairing score $0.0$.

Round 1 pairs within these inflated brackets. Group A's bracket of 50 players
creates pairings like player 1 vs player 26 (similar to no acceleration).

After round 1, a Group A winner has pairing score $1.0 + 1.0 = 2.0$ for
round 2. A Group B winner has $1.0 + 0.0 = 1.0$. Now the bracket structure
is:

- Pairing score 2.0: ~25 Group A winners
- Pairing score 1.0: ~25 Group A losers + ~25 Group B winners
- Pairing score 0.0: ~25 Group B losers

Round 2 pairs the 25 Group A winners among themselves, but the interesting
bracket is score 1.0, which mixes Group A losers with Group B winners --
players from different rating tiers who would not meet this early without
acceleration.

---

## Worked Example

Tournament: 20 players, 7 rounds.

Parameters:

$$\text{accelerated} = \lceil 7/2 \rceil = 4$$
$$\text{fullVP} = \lceil 4/2 \rceil = 2$$
$$\text{halfVP} = 4 - 2 = 2$$
$$\text{gaSize} = 2 \cdot \lceil 20/4 \rceil = 10$$

Virtual points schedule:

| Round | VP for Group A (ranks 1--10) | VP for Group B (ranks 11--20) |
| ----- | ---------------------------- | ----------------------------- |
| 1     | 1.0                          | 0.0                           |
| 2     | 1.0                          | 0.0                           |
| 3     | 0.5                          | 0.0                           |
| 4     | 0.5                          | 0.0                           |
| 5--7  | 0.0                          | 0.0                           |

The transition from 1.0 to 0.5 virtual points in round 3 provides a gradual
"landing" rather than an abrupt removal. By round 5, all players compete on
real scores alone.

---

## Application in the Pairing Pipeline

Baku acceleration integrates into the Swiss pairing pipeline at the score
group construction step:

1. **Build player states** from tournament history.
2. **Apply acceleration.** For each player, add $\text{VP}(p, r)$ to their
   `PairingScore`. This is done by `ApplyBakuAcceleration` in the swisslib
   package.
3. **Build score groups** using the modified pairing scores.
4. **Proceed with normal pairing** (bracket construction, Blossom matching,
   etc.).

The acceleration is transparent to downstream pairing logic. Score groups and
brackets work with the inflated scores without any special handling.

---

## Properties

**Transitivity.** The virtual point addition preserves the relative ordering
within Group A and within Group B. It only changes cross-group ordering by
elevating Group A above Group B.

**Convergence.** As rounds progress, real score differences dominate the
virtual points. After the acceleration phase, the pairing is entirely
score-driven. The tournament's final standings are unaffected.

**Even Group A.** The $2 \cdot \lceil N/4 \rceil$ formula ensures Group A
always has an even number of players, avoiding the need for a bye within the
accelerated bracket.

---

## Supported Systems

Baku acceleration is supported by the Dutch and Burstein pairing systems
(enabled via the `Acceleration` option). The Dubov, Lim, Double-Swiss, and
Team Swiss systems do not currently implement acceleration.

---

## Related Pages

- [Dutch Pairing](/docs/pairing-systems/dutch/) -- the primary system that
  uses Baku acceleration.
- [Dutch Criteria](../dutch-criteria/) -- the criteria that apply after
  acceleration has modified the score groups.
- [Completability](../completability/) -- Stage 0.5 operates on the
  accelerated score groups.
