---
title: "Keizer Convergence"
linkTitle: "Keizer Convergence"
weight: 7
description: "Iterative scoring with oscillation detection — how Keizer scores converge in at most 20 iterations."
---

## The Circular Dependency

Keizer scoring has an unusual property: a player's score depends on the
**value numbers** of their opponents, and value numbers are derived from the
**ranking**, which is determined by scores. This creates a circular dependency:

$$\text{scores} \to \text{ranking} \to \text{value numbers} \to \text{scores}$$

The implementation resolves this through **fixed-point iteration**: start with
an initial ranking (by rating), compute scores, re-rank, recompute, and
repeat until the ranking stabilizes. The implementation lives in
`scoring/keizer/keizer.go`.

---

## Value Numbers

Each player's **value number** is derived from their position in the current
ranking. For $N$ active players ranked $1, 2, \ldots, N$:

$$\text{VN}(k) = \text{base} - (k - 1) \times \text{step}$$

where $k$ is the 1-indexed rank, $\text{base}$ defaults to $N$ (the number of
active players), and $\text{step}$ defaults to 1. The top-ranked player has the
highest value number ($N$); the lowest-ranked has the smallest ($1$).

The value numbers are the "currency" of the Keizer system: beating a
highly-ranked opponent is worth more points than beating a lower-ranked one.

---

## Score Computation

For each player $p$, the Keizer score is the sum of contributions from all
their results:

$$S(p) = \text{selfVN}(p) + \sum_{\text{games}} \text{gameValue}(p, g) + \sum_{\text{non-games}} \text{nonGameValue}(p, b)$$

### Self-Victory

Every player receives their own value number once:

$$\text{selfVN}(p) = \text{VN}(\text{rank}(p))$$

This is a standard Keizer convention. It ensures that even a player with no
games has a non-zero score proportional to their rank.

### Game Values

For a game where player $p$ faced opponent $o$ with value number $\text{VN}(o)$:

| Result | Value                                 |
| ------ | ------------------------------------- |
| Win    | $\text{VN}(o) \times \text{winFrac}$  |
| Draw   | $\text{VN}(o) \times \text{drawFrac}$ |
| Loss   | $\text{VN}(o) \times \text{lossFrac}$ |

The fractions default to: win = 1.0, draw = 0.5, loss = 0.0. These are
configurable via options.

### Non-Game Values (Byes and Absences)

Byes and absences do not involve an opponent. Instead, the player's **own**
value number is used as the base:

| Type                  | Value                                 |
| --------------------- | ------------------------------------- |
| Pairing-allocated bye | $\text{VN}(p) \times \text{pabFrac}$  |
| Full-point bye        | $\text{VN}(p) \times \text{fpbFrac}$  |
| Half-point bye        | $\text{VN}(p) \times \text{hpbFrac}$  |
| Zero-point bye        | $0$                                   |
| Club commitment       | $\text{VN}(p) \times \text{clubFrac}$ |
| Absence               | see below                             |

For absences, the first absence uses $\text{VN}(p) \times \text{absentFrac}$.
Successive absences are **decayed** by halving:

$$\text{absenceValue}(p, k) = \frac{\text{VN}(p) \times \text{absentFrac}}{2^{k-1}}$$

where $k$ is the consecutive absence count. There is also a configurable
`AbsenceLimit` (default: 3) -- absences beyond this limit contribute zero.

---

## x2 Integer Arithmetic

To avoid floating-point drift across iterations, the implementation uses
**doubled integer arithmetic**. All value numbers are stored as $2 \times
\text{VN}$, and all fractions are applied via integer multiplication and
division:

$$\text{internal score} = 2 \times S(p)$$

This preserves half-point granularity (draws, half-point byes) without any
floating-point representation. The final exported scores are divided by 2
to recover the original scale.

---

## The Iteration Loop

```text
function KeizerScore(state):
    ranking ← initial ranking (by rating)
    prevRanking ← nil
    prevPrevRanking ← nil

    for iteration = 1 to 20:
        VN ← computeValueNumbers(ranking)
        scores ← computeScores(state, VN)
        newRanking ← sortByScore(scores)

        if newRanking == ranking:
            return scores                  // Converged

        if newRanking == prevPrevRanking:
            // 2-cycle oscillation detected
            scores ← average(scores, prevScores)
            return scores

        prevPrevRanking ← prevRanking
        prevRanking ← ranking
        prevScores ← scores
        ranking ← newRanking

    return scores                          // Max iterations reached
```

### Convergence Check

After each iteration, the new ranking is compared to the previous ranking.
If they are identical (same player ordering), the scores have stabilized and
the loop terminates.

### Oscillation Detection

A **2-cycle oscillation** occurs when the ranking alternates between two
states:

$$R_k \to R_{k+1} \to R_k \to R_{k+1} \to \cdots$$

This happens when two players with very close scores keep swapping rank
positions, which changes their value numbers just enough to swap them back.
The implementation detects this by comparing the current ranking to the
ranking from _two_ iterations ago (`prevPrevRanking`). If they match, the
system is oscillating between two fixed points.

The resolution is **averaging**: the scores from the last two iterations are
averaged element-wise. This produces a stable intermediate score that places
the oscillating players at the midpoint of their two alternating positions.

---

## Convergence Analysis

### Why Does It Converge?

The Keizer scoring function $f : \text{rankings} \to \text{rankings}$ maps
a ranking to a new ranking via score computation and re-sorting. The domain
is finite (there are $N!$ possible rankings of $N$ players). Therefore:

1. The sequence $R_0, R_1, R_2, \ldots$ must eventually enter a cycle (by
   the pigeonhole principle).
2. If the cycle length is 1, the ranking has converged to a fixed point.
3. If the cycle length is 2, the oscillation detector fires and resolves it
   by averaging.
4. Longer cycles are theoretically possible but not observed in practice. The
   20-iteration cap guarantees termination regardless.

### Practical Convergence Speed

In practice, Keizer scoring converges within 2--5 iterations for typical
club tournaments (20--60 players, 7--11 rounds). The initial ranking (by
rating) is usually close to the final ranking (by Keizer score), so only a
few players swap positions before stabilizing.

Tournaments where many players have similar ratings and scores may require
more iterations, as small value number changes can cascade through the
ranking. The worst observed case in testing is approximately 12 iterations.

### The 20-Iteration Cap

The implementation enforces a hard limit of 20 iterations. If convergence
has not been reached by then, the last computed scores are returned as-is.
This is a safety measure; in practice it is never triggered for realistic
tournament data.

---

## Worked Example

A 5-player tournament after 3 rounds. Initial ranking by rating:

| Rank | Player | Rating |
| ---- | ------ | ------ |
| 1    | Alice  | 2100   |
| 2    | Bob    | 2050   |
| 3    | Carol  | 1950   |
| 4    | Dave   | 1900   |
| 5    | Eve    | 1800   |

With defaults (base = 5, step = 1), initial value numbers: Alice = 5,
Bob = 4, Carol = 3, Dave = 2, Eve = 1.

**Iteration 1.** Compute scores using these value numbers. Suppose:

- Alice beat Bob (VN 4) and Carol (VN 3), lost to Dave (VN 2): game score = $4 + 3 + 0 = 7$, self = 5, total = 12.
- Bob beat Dave (VN 2) and Eve (VN 1), lost to Alice (VN 5): game score = $2 + 1 + 0 = 3$, self = 4, total = 7.
- Carol beat Eve (VN 1), drew Dave (VN 2), lost to Alice (VN 5): game score = $1 + 1 + 0 = 2$, self = 3, total = 5.
- Dave beat Alice (VN 5), drew Carol (VN 3), lost to Bob (VN 4): game score = $5 + 1.5 + 0 = 6.5$, self = 2, total = 8.5.
- Eve lost to Bob (VN 4) and Carol (VN 3): game score = 0, self = 1, total = 1.

New ranking: Alice (12), Dave (8.5), Bob (7), Carol (5), Eve (1).

**Iteration 2.** Dave and Bob swapped ranks (2 and 3). Recompute value
numbers: Alice = 5, Dave = 4, Bob = 3, Carol = 2, Eve = 1.

Recalculate scores with the new value numbers. If the new ranking matches --
Alice, Dave, Bob, Carol, Eve -- the iteration converges.

---

## Fixed-Value Overrides

The options allow overriding the formula-derived value numbers with fixed
values for specific result types. For example, `FixedWinValue` bypasses the
opponent's value number entirely and awards a constant for every win. When
set, the game value becomes:

$$\text{gameValue} = \text{fixedValue}$$

regardless of the opponent's ranking. This turns Keizer into a simpler
point-based system while retaining the iterative framework for non-overridden
result types.

---

## Complexity

Each iteration is $O(N \cdot G)$ where $N$ is the number of players and $G$
is the maximum number of games per player. Re-sorting is $O(N \log N)$.
With at most 20 iterations:

$$O(20 \cdot (N \cdot G + N \log N)) = O(N \cdot G)$$

since $G$ is bounded by the number of rounds (at most $N - 1$ in round-robin,
typically 7--11 in Swiss).

---

## Related Pages

- [Keizer Scoring](/docs/scoring/keizer/) -- configuration and usage of the
  Keizer scoring system.
- [Keizer Pairing](/docs/pairing-systems/keizer/) -- the pairing system that
  uses Keizer scores for ranking-based pairing.
