---
title: "Dubov Criteria"
linkTitle: "Dubov Criteria"
weight: 11
description: "The 10 criteria governing Dubov pairings — MaxT tracking and transposition caps."
---

## Overview

The Dubov system (FIDE C.04.4.1) defines 10 criteria ($C_1$--$C_{10}$) that
govern pairing within each score group. Unlike the Dutch system, which
encodes all optimization criteria into Blossom edge weights for global
matching, the Dubov system processes each score group independently using
**transposition-based matching** with a lexicographic candidate comparison.

The criteria are implemented in `pairing/dubov/criteria.go`.

---

## Absolute Criteria

### C1: No Rematches

Identical to Dutch C1. Two players who have already played each other
(excluding forfeits) shall not be paired again.

Implementation delegates to `swisslib.C1NoRematches`.

### C3: No Absolute Color Conflict

Two players who both have an absolute color preference for the same color
shall not be paired. A player has an absolute color preference when their
color imbalance exceeds 1 or they have 2+ consecutive games with the same
color.

Unlike Dutch C3, there is no top-scorer exception in the Dubov system.

Implementation: `C3NoAbsoluteColorConflict` in `pairing/dubov/criteria.go`.

### Forbidden Pairs

Pairs explicitly forbidden by the tournament organizer are excluded. This is
checked alongside C1 and C3 in `SatisfiesAbsoluteCriteria`.

---

## The MaxT Parameter

A distinctive feature of the Dubov system is the **MaxT** cap on
transpositions. For each score group, the number of transpositions the
algorithm may consider before accepting a candidate pairing is limited by:

$$\text{MaxT} = 2 + \left\lfloor \frac{R}{5} \right\rfloor$$

where $R$ is the number of completed rounds.

| Completed rounds | MaxT |
| ---------------- | ---- |
| 0--4             | 2    |
| 5--9             | 3    |
| 10--14           | 4    |
| 15--19           | 5    |

MaxT controls the trade-off between pairing quality and computational cost.
Early in the tournament, fewer transpositions are needed because most
pairings are straightforward. As the tournament progresses and constraints
accumulate, more flexibility is allowed.

Implementation: `MaxT` function in `pairing/dubov/criteria.go`.

---

## Optimization Criteria (C4--C10)

The optimization criteria are evaluated on candidate pairings and compared
lexicographically. A `DubovCandidateScore` records the violation counts for
C4--C10 plus a transposition index. The `Compare` method performs a strict
lexicographic comparison: C4 violations are compared first; if equal, C5;
and so on.

### C4: Minimize Upfloater Count

Minimize the number of players in the bracket who have floated up from a
lower score group. An upfloater is a player whose float history includes
`FloatUp`.

$$\text{C4 violations} = |\{p \in \text{bracket} : \text{FloatUp} \in \text{history}(p)\}|$$

Implementation: `UpfloatCount` counts `FloatUp` entries in the player's
float history.

### C5: Minimize Upfloater Score Sum

Among the upfloaters counted in C4, minimize the sum of their scores. This
prefers floating up lower-scoring players over higher-scoring ones.

$$\text{C5 violations} = \sum_{\substack{p \in \text{bracket} \\ \text{upfloater}(p)}} \text{score}(p)$$

### C6: Minimize Color Preference Violations

Minimize the number of pairs where both players have a color preference
(strong or absolute) for the same color. Unlike Dutch C10--C13, which
distinguish four levels, Dubov C6 treats all preference conflicts equally.

### C7: Minimize MaxT Upfloater Violations

Count the number of upfloaters whose upfloat count exceeds MaxT. A player
who has floated up too many times (more than MaxT) represents a C7 violation.

$$\text{C7 violations} = |\{p : \text{upfloatCount}(p) > \text{MaxT}\}|$$

### C8: Minimize Consecutive Upfloaters

Count the number of players who floated up in both the current round and
the immediately preceding round. Consecutive upfloats are more disruptive
than isolated ones.

### C9: Minimize MaxT Opponent Violations

Count the number of pairings where one player's opponent has floated up
more than MaxT times. This spreads the burden of facing upfloaters.

### C10: Minimize Consecutive MaxT Violations

Count the number of players who exceeded the MaxT upfloat limit in both
the current round and the preceding round.

---

## Candidate Scoring

Each candidate pairing for a score group receives a `DubovCandidateScore`
containing:

```text
DubovCandidateScore {
    C4Violations    int  // upfloater count
    C5Violations    int  // upfloater score sum
    C6Violations    int  // color preference violations
    C7Violations    int  // MaxT upfloater violations
    C8Violations    int  // consecutive upfloaters
    C9Violations    int  // MaxT opponent violations
    C10Violations   int  // consecutive MaxT violations
    Transposition   int  // transposition index (lower = better)
}
```

The `Compare` method returns $-1$, $0$, or $+1$ by comparing fields in
order from C4 to C10. If all violation counts are equal, the transposition
index breaks the tie (lower is better, corresponding to the more "natural"
pairing order).

A score is **perfect** when all violation counts are zero and the
transposition index is zero. A perfect score means the natural pairing
order satisfies all optimization criteria.

---

## Processing Order: Ascending ARO

Unlike the Dutch system, which processes score groups from highest to lowest,
the Dubov system processes players within each score group in **ascending
ARO** (Average Rating of Opponents) order.

The ARO is computed from the player's actual game history:

$$\text{ARO}(p) = \frac{1}{|G(p)|} \sum_{g \in G(p)} \text{rating}(\text{opponent}(g))$$

where $G(p)$ is the set of games played by $p$ (excluding forfeits).

Ascending ARO processing means players who have faced weaker opponents so
far are paired first. This tends to produce pairings that equalize average
opponent strength across the tournament.

Implementation: `pairing/dubov/aro.go`.

---

## Matching Algorithm

The Dubov system uses **transposition-based matching** within each score
group:

1. **Sort** players by ascending ARO.
2. **Split** into two halves: G1 (upper, higher ARO) and G2 (lower, lower
   ARO).
3. **Generate transpositions** of G2 (permutations that rearrange the
   pairing partners).
4. For each transposition (up to MaxT):
   - Pair G1[i] with G2[i] for each position $i$.
   - Check absolute criteria (C1, C3, forbidden).
   - Score the candidate pairing (C4--C10).
   - If perfect, accept immediately.
   - Otherwise, record as the best candidate if it improves on the
     previous best (by lexicographic comparison).
5. Accept the best candidate found within MaxT transpositions.

This is fundamentally different from the Dutch approach (global Blossom
matching) and the Lim approach (exchange matching). The transposition cap
MaxT limits the search space, trading optimality for predictability.

Implementation: `pairing/dubov/matching.go`.

---

## Comparison with Dutch Criteria

| Aspect              | Dutch ($C_1$--$C_{21}$)             | Dubov ($C_1$--$C_{10}$)              |
| ------------------- | ----------------------------------- | ------------------------------------ |
| Criteria count      | 21                                  | 10                                   |
| Absolute criteria   | C1, C3 (with top-scorer exception)  | C1, C3 (no exception)                |
| Color criteria      | 4 levels (C10--C13)                 | 1 level (C6)                         |
| Float criteria      | 8 (C14--C21, round $R-1$ and $R-2$) | 4 (C7--C10, MaxT-based)              |
| Matching method     | Global Blossom                      | Transposition with MaxT cap          |
| Processing order    | Descending score                    | Ascending ARO within bracket         |
| Score groups        | Global across all groups            | One group at a time                  |
| Computational model | Polynomial (Blossom $O(n^3)$)       | Bounded search (MaxT transpositions) |

The Dubov system is simpler and more predictable. The MaxT cap ensures
that the algorithm's behavior can be understood by arbiters: at most
$\text{MaxT}$ alternative pairings are considered before accepting the best
one found. The Dutch system's Blossom matching considers all possible
pairings simultaneously, which is more powerful but harder to trace by hand.

---

## Related Pages

- [Dubov Pairing](/docs/pairing-systems/dubov/) -- the pairing system these
  criteria govern.
- [Dutch Criteria](../dutch-criteria/) -- the Dutch alternative with 21
  criteria.
- [Color Allocation](../color-allocation/) -- how Dubov resolves color
  preferences after pairing.
