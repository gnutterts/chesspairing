---
title: "Color Allocation"
linkTitle: "Color Allocation"
weight: 14
description: "Six color allocation algorithms compared — Dutch, Lim, Double-Swiss, Team, and more."
---

## Overview

After pairing determines _who_ plays _whom_, **color allocation** determines
_who plays White and who plays Black_. Each pairing system implements its own
algorithm with different priority rules, reflecting the different philosophies
of the FIDE regulations.

This page compares the six color allocation algorithms in the codebase.

---

## Common Concepts

All algorithms share these building blocks:

### Color Preference

A player's color preference is derived from their game history:

- **Imbalance**: number of White games minus number of Black games.
- **Consecutive count**: number of games with the same color at the end of
  the history.
- **Preference direction**: the color the player "needs" based on imbalance
  and consecutive history.

### Preference Strength

Most systems classify preferences by strength:

| Strength | Condition                                           | Meaning                                    |
| -------- | --------------------------------------------------- | ------------------------------------------ |
| Absolute | Imbalance $> 1$, or $\geq 2$ consecutive same color | Player _must_ receive the opposite color   |
| Strong   | Imbalance $= 1$, or 1 consecutive same color        | Player _should_ receive the opposite color |
| Mild     | Slight preference from game history                 | Nice to have, but not required             |
| None     | Balanced history                                    | No preference                              |

### Constraint: No 3 Consecutive

The hardest constraint shared by all systems: no player may play 3
consecutive games with the same color. This is checked as a precondition
(in compatibility for Lim, in color allocation for others).

---

## Algorithm 1: Dutch / Burstein (swisslib 6-step)

Used by the Dutch (C.04.3) and Burstein (C.04.4.2) systems. Implementation
in `pairing/swisslib/color.go`.

The algorithm follows bbpPairings' `choosePlayerColor`:

### Step 1: Compatible Preferences

If the two players prefer different colors (or at least one has no
preference), grant both their preferred color. This resolves the majority
of cases.

### Step 2: Absolute Wins

If one player has an **absolute** preference and the other does not, the
absolute preference wins. The other player receives the opposite color.

### Step 3: Strong Beats Non-Strong

If one player has a **strong** preference and the other has only a mild
preference or none, the strong preference wins.

### Step 4: First Color Difference

When both players have the same preference strength, walk **backwards**
through their game histories simultaneously. At the first round where one
player had White and the other had Black (the "first color difference"
round), the player who had the preferred color at that point now receives
the opposite color.

For example, if both prefer White, and in round 3 player A had White while
player B had Black, then player B gets White in the current round (since
player A had it more recently at the divergence point).

### Step 5: Rank Tiebreak

If histories are identical through all rounds, the player with the higher
rank (lower TPN) receives their preferred color.

### Step 6: Board Alternation

For round 1 pairings (no game history), odd-numbered boards give White to
the higher-ranked player and even-numbered boards give White to the lower-
ranked player (or vice versa, depending on the `TopSeedColor` option). This
ensures colors alternate across the board list.

### Top-Scorer Rules

When both players are top scorers (in the highest non-empty score group),
the absolute/strong distinctions from C3 are relaxed. This prevents the
leaders from being unable to play due to color constraints, at the cost of
potentially giving one player a third consecutive same-color game.

---

## Algorithm 2: Dubov

Used by the Dubov system (C.04.4.1). Delegates to the swisslib algorithm
(same 6-step procedure as Dutch/Burstein) after the Dubov-specific pairing
phase completes. The color preferences used by C6 (color preference
violations) during pairing are the same as those used during allocation.

---

## Algorithm 3: Lim (Art. 5)

Used by the Lim system (C.04.4.3). Implementation in `pairing/lim/color.go`.

The Lim algorithm is the most distinctive, featuring **round-parity
awareness** and **median tiebreaking**:

### Round 1

Odd TPN gets the initial color (default White); even TPN gets the opposite.

### Art. 5.3: Must-Alternate

If a player has 2 consecutive games with the same color, they _must_ receive
the opposite color. If both players trigger this rule and need the same
color, the algorithm raises a conflict (this should have been caught by the
compatibility check during pairing).

### Art. 5.2/5.6: Even-Round Equalizing

In even-numbered rounds, the player with more games of one color receives
the opposite color. This actively equalizes the color balance.

### Art. 5.5: Odd-Round Alternating

In odd-numbered rounds, each player receives the opposite of their last
played color. This creates a natural alternation pattern.

### Art. 5.4: History Tiebreak with Median

When the above rules do not resolve the assignment (both players have
identical constraints), walk backwards through game histories:

1. Find the first round where the two players had different colors.
2. The player whose position is **above the median** of the current score
   group gets priority for their preferred color.

"Above the median" means the player's rank is in the upper half of the
score group. This is a deliberate advantage for higher-ranked players in
the Lim philosophy.

---

## Algorithm 4: Double-Swiss (Art. 4)

Used by the Double-Swiss system (C.04.5). Implementation in
`pairing/doubleswiss/color.go`.

Double-Swiss allocates colors for **Game 1** of each 2-game match (Game 2
automatically reverses colors). The algorithm has 5 steps:

### Step 1: Hard Constraint (No 3 Consecutive)

If giving a player a specific color would create 3 consecutive games with
that color, assign the opposite. This is the only absolute constraint.

### Step 2: Equalize

The player with more games of one color receives the opposite color. This
balances the overall color distribution.

### Step 3: Alternate

Each player receives the opposite of their last played color.

### Step 4: Round-1 Board Alternation

In round 1, odd-numbered boards give White to the higher-ranked player;
even boards reverse. The `TopSeedColor` option controls the starting color.

### Step 5: Rank Tiebreak

The player with the higher rank (lower TPN) receives White.

The 3-consecutive ban (Step 1) is unique to Double-Swiss: it is checked as
a hard constraint during color allocation rather than during compatibility
(as in Lim).

---

## Algorithm 5: Team Swiss (Art. 4, 9-step)

Used by the Team Swiss system (C.04.6). Implementation in
`pairing/team/color.go` and `pairing/team/color_pref.go`.

This is the most complex color algorithm, with 9 steps and the concept of
a **first team**:

### Color Preference Types

Team Swiss supports two preference computation modes:

| Type   | Description                                                                                                                               |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| Type A | Simple: color difference $< -1$ or last 2 games Black implies White preference. Symmetric for Black.                                      |
| Type B | Strong + Mild: same conditions give "strong" preference; additional conditions (color difference $= \pm 1$, etc.) give "mild" preference. |

### First Team

The **first team** in a pairing is determined by:

1. Higher primary score, or
2. If tied, higher secondary score, or
3. If still tied, lower TPN (higher seed).

The first-team concept gives one team slight priority in ambiguous cases.

### The 9 Steps

1. **No history**: if neither team has game history, assign by TPN parity
   and initial color setting.
2. **One preference**: if only one team has a color preference, grant it.
3. **Opposite preferences**: if preferences differ, grant both.
4. **Strong beats non-strong** (Type B only): if one has a strong preference
   and the other has only mild, the strong preference wins.
5. **Lower color difference**: the team with the lower color difference
   (fewer Whites relative to Blacks) receives White.
6. **Alternation from history**: walk backwards through game histories to
   find the most recent divergence. The team due for a change gets it.
7. **First team preference**: grant the first team's preference.
8. **First team alternation**: give the first team the opposite of their
   last color.
9. **Other team alternation**: give the non-first team the opposite of
   their last color.

Steps 7--9 are progressive fallbacks for when all prior rules are
indeterminate.

---

## Comparison Table

| Feature              | Dutch/Burstein               | Lim                                    | Double-Swiss      | Team Swiss                              |
| -------------------- | ---------------------------- | -------------------------------------- | ----------------- | --------------------------------------- |
| Steps                | 6                            | 5 + median                             | 5                 | 9                                       |
| Preference levels    | Absolute, Strong, Mild, None | Binary + must-alternate                | Binary            | Type A (simple) or Type B (strong/mild) |
| History walk         | Backward to first difference | Backward + median                      | N/A               | Backward to recent divergence           |
| Round parity         | No                           | Yes (even = equalize, odd = alternate) | No                | No                                      |
| Median tiebreak      | No                           | Yes                                    | No                | No                                      |
| First-entity concept | No                           | No                                     | No                | Yes (first team)                        |
| 3-consecutive check  | During compatibility         | During compatibility                   | During allocation | N/A (inherent in preferences)           |
| Board alternation    | Round 1                      | Round 1 by TPN parity                  | Round 1           | N/A                                     |
| Top-scorer exception | Yes                          | No                                     | No                | No                                      |

---

## Design Rationale

The different algorithms reflect different philosophies:

- **Dutch/Burstein**: maximizes color satisfaction across the tournament via
  history-based tiebreaking. The backward walk ensures that long-term color
  patterns are considered, not just recent games.

- **Lim**: emphasizes round-level fairness. Even rounds actively equalize;
  odd rounds alternate. The median tiebreak adds a subtle ranking advantage
  that rewards better tournament performance.

- **Double-Swiss**: prioritizes the match-level experience. Since each match
  is 2 games with reversed colors, only Game 1's color matters for the
  sequence. The algorithm is simpler because the match structure inherently
  provides color balance.

- **Team Swiss**: adds organizational complexity for team events. The
  first-team concept ensures that the team with better tournament performance
  gets slight priority in ambiguous color decisions, reflecting the
  competitive hierarchy.

---

## Related Pages

- [Dutch Criteria](../dutch-criteria/) -- C10--C13 color optimization in
  the edge weights.
- [Lim Exchange Matching](../lim-exchange/) -- compatibility checks that
  include color feasibility.
- [Concepts: Colors & Balance](/docs/concepts/colors/) -- general
  introduction to color allocation for chess players.
