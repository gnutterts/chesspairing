---
title: "FIDE B.02 Conversion Table"
linkTitle: "B.02 Table"
weight: 9
description: "The lookup table converting winning percentages to rating differences and vice versa."
---

## Purpose

The [Elo probability model](../elo-model/) defines a continuous function
mapping rating differences to expected scores. In practice, FIDE does not
use the continuous formula directly. Instead, **FIDE Regulation B.02 Table
8.1b** provides a discrete lookup table with 101 entries, and all official
calculations use this table with interpolation.

The implementation in `tiebreaker/ratingtable.go` stores this table and
provides two lookup functions: `dpFromP` (score to rating difference) and
`expectedScore` (rating difference to score).

---

## The Table

The table maps fractional scores $p \in [0.00, 1.00]$ to rating differences
$d_p \in [-800, +800]$. Selected entries:

| $p$  | $d_p$  | $p$  | $d_p$ |
| ---- | ------ | ---- | ----- |
| 0.00 | $-800$ | 0.50 | $0$   |
| 0.01 | $-677$ | 0.55 | $36$  |
| 0.05 | $-470$ | 0.60 | $72$  |
| 0.10 | $-366$ | 0.65 | $110$ |
| 0.15 | $-296$ | 0.70 | $149$ |
| 0.20 | $-240$ | 0.75 | $193$ |
| 0.25 | $-193$ | 0.80 | $240$ |
| 0.30 | $-149$ | 0.85 | $296$ |
| 0.35 | $-110$ | 0.90 | $366$ |
| 0.40 | $-72$  | 0.95 | $470$ |
| 0.45 | $-36$  | 0.99 | $677$ |
|      |        | 1.00 | $800$ |

The table is symmetric around $p = 0.50$: $d_p = -d_{1-p}$.

The full 101-entry table is stored as a constant array indexed by
$\lfloor 100 \cdot p \rfloor$.

---

## Forward Lookup: dpFromP

Given a fractional score $p$ (score divided by number of games), find the
corresponding rating difference $d_p$.

### Algorithm

1. **Clamp** $p$ to $[0, 1]$.
2. **Scale** to the table index: $i = p \times 100$.
3. **Integer part**: $\lfloor i \rfloor$ gives the lower table index.
4. **Fractional part**: $f = i - \lfloor i \rfloor$.
5. **Interpolate**:

$$d_p = d[\lfloor i \rfloor] + f \cdot \left(d[\lceil i \rceil] - d[\lfloor i \rfloor]\right)$$

If $\lfloor i \rfloor = \lceil i \rceil$ (exact table entry), no
interpolation is needed.

### Boundary Handling

- $p = 0.00$: returns $-800$ (minimum rating difference).
- $p = 1.00$: returns $+800$ (maximum rating difference).

These are FIDE's conventional clamps. A player who wins every game is treated
as performing 800 points above the average opponent; a player who loses every
game is treated as 800 points below.

---

## Inverse Lookup: expectedScore

Given a rating difference $d$, find the expected fractional score $E(d)$.
This is the inverse of `dpFromP`.

### Algorithm

1. **Clamp** $d$ to $[-800, +800]$.
2. **Binary search** through the table (which is sorted by $d_p$) to find
   the two entries bracketing $d$:

   $$d[j] \leq d < d[j+1]$$

3. **Interpolate**:

$$E = \frac{j}{100} + \frac{d - d[j]}{d[j+1] - d[j]} \cdot \frac{1}{100}$$

This converts the table index back to a fractional score, interpolating
within the bracketing interval.

### Boundary Handling

- $d \leq -800$: returns $0.00$.
- $d \geq +800$: returns $1.00$.

---

## Interpolation Accuracy

The table entries are spaced at 1% intervals in $p$, so the maximum
interpolation error depends on the curvature of the Elo function within
each interval. Near $p = 0.50$ (where the function is nearly linear), the
error is negligible. Near the extremes ($p$ close to 0 or 1), the function
curves sharply and interpolation is less precise.

For the values used in FIDE tiebreaking calculations, the interpolation
error is well within the rounding tolerance of 1 rating point. The final
TPR and PTP values are rounded to integers, absorbing any sub-unit
interpolation artifacts.

---

## Use in Tiebreakers

### Tournament Performance Rating (TPR)

The `performancerating` tiebreaker computes:

$$\text{TPR} = \text{ARO} + d_p\!\left(\frac{S}{n}\right)$$

where $S$ is the player's score and $n$ is the number of rated games
(excluding forfeits and byes). The $d_p$ lookup uses this table.

### Performance Points (PTP)

The `performancepoints` tiebreaker uses `expectedScore` iteratively during
binary search:

$$\sum_{i=1}^{n} E(R - R_i) \stackrel{?}{\geq} S$$

Each $E(R - R_i)$ call goes through the inverse lookup in this table.

### Average Opponent TPR (APRO)

Computes TPR for each opponent, then averages. Each opponent's TPR uses the
$d_p$ lookup.

### Average Opponent PTP (APPO)

Computes PTP for each opponent, then averages. Each opponent's PTP uses the
$E(d)$ lookup.

---

## Table vs Formula

The FIDE B.02 table values are _close to_ but _not identical to_ the
continuous logistic formula:

$$d_{\text{formula}}(p) = -400 \cdot \log_{10}\!\left(\frac{1}{p} - 1\right)$$

For example:

| $p$  | Table $d_p$ | Formula $d_{\text{formula}}$ | Difference |
| ---- | ----------- | ---------------------------- | ---------- |
| 0.50 | 0           | 0.0                          | 0          |
| 0.60 | 72          | 72.2                         | 0.2        |
| 0.70 | 149         | 146.8                        | 2.2        |
| 0.80 | 240         | 240.8                        | 0.8        |
| 0.90 | 366         | 366.0                        | 0.0        |
| 0.95 | 470         | 476.2                        | 6.2        |

The differences are small (within 7 rating points) but exist because the
FIDE table was originally derived from normal distribution approximations
and has been retained for historical consistency. The implementation uses
the table, not the formula, to match FIDE's official calculations.

---

## Related Pages

- [Elo Probability Model](../elo-model/) -- the continuous model that the
  table discretizes.
- [Performance Rating Tiebreakers](/docs/tiebreakers/performance/) -- the
  tiebreakers that consume these lookup functions.
