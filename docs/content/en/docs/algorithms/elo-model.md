---
title: "Elo Probability Model"
linkTitle: "Elo Model"
weight: 8
description: "The expected score function and its role in performance rating calculations."
---

## The Logistic Model

The Elo rating system models the probability of a game outcome as a function
of the rating difference between two players. Given player $A$ with rating
$R_A$ and player $B$ with rating $R_B$, the **expected score** for player $A$
is:

$$E_A = \frac{1}{1 + 10^{-(R_A - R_B) / 400}}$$

This is a logistic function with base 10 and a scaling factor of 400. The
key properties:

- $E_A = 0.5$ when $R_A = R_B$ (equal ratings imply equal chances).
- $E_A \to 1$ as $(R_A - R_B) \to +\infty$.
- $E_A \to 0$ as $(R_A - R_B) \to -\infty$.
- $E_A + E_B = 1$ (the scores are complementary).

---

## Rating Difference and Winning Probability

Setting $d = R_A - R_B$, the expected score simplifies to:

$$E(d) = \frac{1}{1 + 10^{-d/400}}$$

Some reference values:

| $d$ (rating difference) | $E(d)$ (expected score) | Interpretation  |
| ----------------------- | ----------------------- | --------------- |
| 0                       | 0.50                    | Equal chances   |
| 100                     | 0.64                    | Slight favorite |
| 200                     | 0.76                    | Clear favorite  |
| 400                     | 0.91                    | Strong favorite |
| 800                     | 0.99                    | Near certain    |
| $-100$                  | 0.36                    | Slight underdog |
| $-200$                  | 0.24                    | Clear underdog  |

The function is monotonically increasing: a larger rating advantage always
means a higher expected score. It is also symmetric:

$$E(-d) = 1 - E(d)$$

---

## Inverse Function: Rating Difference from Score

The inverse of the expected score function gives the rating difference
corresponding to an observed winning percentage $p$:

$$d(p) = -400 \cdot \log_{10}\!\left(\frac{1}{p} - 1\right)$$

This is used by the [FIDE B.02 table](../fide-b02/) to convert tournament
performance (fractional score) into a rating difference. For example, a
player who scores 7/10 = 0.70 against their opponents has:

$$d(0.70) = -400 \cdot \log_{10}\!\left(\frac{1}{0.70} - 1\right) = -400 \cdot \log_{10}(0.4286) \approx 147$$

This means the player performed approximately 147 rating points above the
average rating of their opponents.

### Boundary Behavior

The inverse function has singularities at $p = 0$ and $p = 1$:

- $d(0) = -\infty$: a player who loses every game performs infinitely below
  their opponents.
- $d(1) = +\infty$: a player who wins every game performs infinitely above
  their opponents.

FIDE handles this by clamping: $d$ is bounded to $[-800, +800]$ in the B.02
table. A player with a perfect score receives $d = +800$; a player with zero
score receives $d = -800$.

---

## Use in Performance Rating (TPR)

The **Tournament Performance Rating** (FIDE Art. 10.2) estimates the rating
a player would need to achieve their observed results against their specific
opponents:

$$\text{TPR} = \text{ARO} + d(p)$$

where:

- $\text{ARO} = \frac{1}{n} \sum_{i=1}^{n} R_i$ is the Average Rating of
  Opponents (computed over actual games, excluding forfeits and byes).
- $p = \frac{\text{score}}{n}$ is the fractional score (wins count 1, draws
  count 0.5).
- $d(p)$ is looked up from the [FIDE B.02 table](../fide-b02/) (not computed
  from the formula directly, for consistency with FIDE's official tabulated
  values).

The result is rounded to the nearest integer.

---

## Use in Performance Points (PTP)

The **Performance Points** tiebreaker (FIDE Art. 10.3) works in reverse.
Instead of computing a rating from a score, it finds the minimum rating $R^*$
such that the expected total score against the actual opponents equals or
exceeds the actual score:

$$R^* = \min \left\{ R \;\middle|\; \sum_{i=1}^{n} E(R - R_i) \geq S \right\}$$

where $R_i$ are the opponent ratings and $S$ is the actual score.

This is solved by **binary search** over the rating range
$[\min_i R_i - 800, \max_i R_i + 800]$ with 0.5 rating-point precision.
The expected score function $E(d)$ is evaluated using the
[FIDE B.02 table](../fide-b02/) at each step.

Special cases:

| Condition               | Result                   |
| ----------------------- | ------------------------ |
| $S = 0$ (zero score)    | $R^* = \min_i R_i - 800$ |
| $S = n$ (perfect score) | $R^* = \max_i R_i + 800$ |
| No games                | $R^* = 0$                |

---

## Why Base 10 and Factor 400?

The original Elo system (1960s) used a normal distribution approximation.
FIDE later adopted the logistic model with base 10 and a scaling factor of
400 because:

1. **400-point difference** corresponds to approximately a 10:1 odds ratio
   ($E(400) \approx 0.91$), which provides an intuitive scale for chess
   players.
2. **Base 10** keeps the arithmetic simple: a difference of 400 points means
   a factor of $10^1 = 10$ in the odds.
3. The logistic function is simpler to compute and has heavier tails than the
   normal distribution, better matching observed chess results where upsets
   occur more frequently than a normal model predicts.

---

## Relationship to Other Models

| Model               | Formula                      | Used by                                 |
| ------------------- | ---------------------------- | --------------------------------------- |
| FIDE Elo (logistic) | $E = (1 + 10^{-d/400})^{-1}$ | FIDE ratings, this implementation       |
| USCF Elo (logistic) | $E = (1 + 10^{-d/400})^{-1}$ | Same formula, different K-factors       |
| Glicko              | Logistic with uncertainty    | Glicko-2 rating system                  |
| Bradley-Terry       | $E = e^{d} / (1 + e^{d})$    | General paired comparison (natural log) |

The FIDE logistic model is a special case of the Bradley-Terry model with
the substitution $d_{\text{BT}} = \frac{d \cdot \ln(10)}{400}$.

---

## Related Pages

- [FIDE B.02 Conversion Table](../fide-b02/) -- the discretized lookup table
  used instead of the continuous formula.
- [Performance Rating Tiebreakers](/docs/tiebreakers/performance/) -- TPR and
  PTP tiebreakers that use the Elo model.
