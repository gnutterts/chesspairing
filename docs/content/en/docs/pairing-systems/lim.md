---
title: "Lim System"
linkTitle: "Lim"
weight: 4
description: "FIDE C.04.4.3 — median-first processing with four floater types and exchange matching."
---

## Overview

The Lim system is a Swiss variant that processes score groups starting from the extremes and working toward the median, rather than strictly top-down. It classifies floaters into four types (A through D) based on their history and compatibility with adjacent groups, and uses exchange-based matching within each score group instead of Blossom or transposition matching.

As a player, the most noticeable difference from the Dutch system is that the middle score groups are paired last, giving the engine the most information about floaters when handling the most populated part of the field. The system also enforces strict compatibility rules: no three consecutive games with the same colour, and no colour imbalance of three or more.

## When to Use

- Tournaments where you want median-first processing to produce better floater distribution in the middle of the standings.
- Events where explicit floater classification (types A-D) provides more transparent pairing decisions.
- Maxi-format tournaments where a 100-point rating constraint on exchanges and floater selection is required.
- Situations where exchange-based matching (rather than optimization-based) is preferred for auditability.

The Lim system does **not** support Baku acceleration or a configurable `TotalRounds` option. If you need acceleration, use the [Dutch](../dutch/) or [Burstein](../burstein/) system.

## Configuration

### CLI

```bash
chesspairing pair --lim tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/lim"

// With typed options
p := lim.New(lim.Options{
    TopSeedColor:   chesspairing.StringPtr("auto"),
    MaxiTournament: chesspairing.BoolPtr(true),
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})

// From a generic map (e.g., parsed from JSON config)
p := lim.NewFromMap(map[string]any{
    "topSeedColor":   "white",
    "maxiTournament": true,
})
```

### Options

| Option           | Type         | Default  | Description                                                                                         |
| ---------------- | ------------ | -------- | --------------------------------------------------------------------------------------------------- |
| `TopSeedColor`   | `string`     | `"auto"` | Colour for the top seed in round 1. Values: `"auto"`, `"white"`, `"black"`.                         |
| `ForbiddenPairs` | `[][]string` | `nil`    | Pairs of player IDs that must never be paired together.                                             |
| `MaxiTournament` | `bool`       | `false`  | Enables the 100-point rating constraint for exchanges and floater selection (Art. 3.2.3, 3.8, 5.7). |

The Lim system has no `Acceleration` or `TotalRounds` options.

## How It Works

### Step-by-step algorithm

1. **Build player states** from tournament history.

2. **Bye selection** (odd player count): The `LimByeSelector` picks the lowest-ranked player (highest TPN) in the lowest score group who has not already received a PAB. The bye player is removed from the pool.

3. **Build score groups** from the remaining players.

4. **Compute median score**: `roundsPlayed / 2.0`. This divides the field into above-median, below-median, and median groups.

5. **Determine processing order** (Art. 2.2), the defining feature of the Lim system:
   - **Phase 1**: Highest score group down to just above the median.
   - **Phase 2**: Lowest score group up to just below the median (reversed direction).
   - **Phase 3**: Median group last.

   This ensures the median group -- typically the largest and most constrained -- is processed with full knowledge of all floaters from above and below.

6. **For each score group in processing order**:
   - **Merge incoming floaters** into the group. Floaters are sorted per Art. 3.6/3.7 priority: down-floaters before up-floaters in upper-half groups, reversed in lower-half groups.
   - **If odd count, select a floater** to pass to the next group. Selection uses `SelectDownFloater` (above median) or `SelectUpFloater` (below median), considering floater type, colour equalization, and compatibility with the adjacent group.
   - **Exchange match** the remaining even-count group (Art. 4). Players are split into top half (S1) and bottom half (S2) by TPN, with proposed pairings S1[i] vs S2[i]. Incompatible pairs are resolved by exchanging the S2 partner per Art. 4.2 scrutiny order.
   - **Colour exchange pass** (Art. 5.2/5.7): after matching, opponents are swapped between pairs to reduce colour conflicts, subject to compatibility constraints. In maxi-tournaments, exchanged players' ratings must differ by 100 points or less.

7. **Pair remaining floaters** across score group boundaries using greedy matching with repair strategies (same-pair swap and chain swap).

8. **Board ordering**: max score of pair descending, then min TPN ascending.

9. **Colour allocation** via Lim-specific rules (Art. 5) with median-aware tiebreaking: above the median, the higher-ranked player wins colour ties; below the median, the lower-ranked player wins.

### Four floater types

The Lim system classifies each floater candidate based on two factors: whether the player has already floated into the current group, and whether the player has a compatible opponent in the adjacent group.

| Type | Already floated? | Compatible opponent in adjacent? | Priority                     |
| ---- | ---------------- | -------------------------------- | ---------------------------- |
| A    | Yes              | No                               | Worst (highest disadvantage) |
| B    | Yes              | Yes                              |                              |
| C    | No               | No                               |                              |
| D    | No               | Yes                              | Best (least disadvantage)    |

When selecting a floater, the engine prefers type D (least disadvantaged) first. This minimizes the damage to the player who must float, since they have the best chance of being paired in the next group.

### Compatibility rules

Two players are compatible (Art. 2.1) if all of the following hold:

1. They have not already played each other.
2. They are not a forbidden pair.
3. At least one legal colour assignment exists where neither player would have:
   - The same colour in three consecutive rounds.
   - A colour imbalance of three or more (e.g., 5 Whites and 2 Blacks).

### Exchange matching details

The exchange algorithm (Art. 4) works within a single score group:

1. Split players into S1 (lower TPNs) and S2 (higher TPNs).
2. Propose initial pairings: S1[0] vs S2[0], S1[1] vs S2[1], etc.
3. Scrutinize each pair. When pairing downward, scrutiny starts from the highest-numbered player in S1. When pairing upward, scrutiny starts from the lowest-numbered.
4. For each incompatible pairing, try S2 exchanges first (proposed partner, then remaining S2 players in exchange order), then S1 cross-half partners.
5. If complete pairing fails, fall back to greedy matching.

### Maxi-tournament mode

When `MaxiTournament` is enabled:

- **Floater selection** (Art. 3.2.3): if the selected floater's rating differs from the reference player by more than 100 points, the reference player (lowest TPN when floating down, highest TPN when floating up) is chosen instead, overriding floater type priority.
- **Floater opponent selection** (Art. 3.8): candidates whose rating differs from the floater by more than 100 points are excluded.
- **Colour exchange** (Art. 5.7): opponent swaps between pairs are only allowed if the exchanged players' ratings differ by 100 points or less.

### Error handling

The Lim pairer does not return sentinel errors. When pairing is partially impossible, it returns the best partial result with additional byes assigned to players who could not be paired.

## Comparison with Dutch

| Aspect                 | Lim                                                       | Dutch                                         |
| ---------------------- | --------------------------------------------------------- | --------------------------------------------- |
| Processing order       | Median-first (high-to-median, low-to-median, median last) | Strict top-down                               |
| Matching               | Exchange-based (S1/S2 with Art. 4 scrutiny)               | Global Blossom matching                       |
| Floater classification | 4 types (A-D) with priority ordering                      | No explicit types; Blossom handles float cost |
| Compatibility          | Explicit 3-consecutive / 3-imbalance rules                | Absolute colour criteria (C3)                 |
| Colour allocation      | Median-aware tiebreaking (Art. 5.4)                       | Standard swisslib allocation                  |
| Maxi-tournament        | 100-point rating constraint on exchanges                  | Not supported                                 |
| Acceleration           | Not supported                                             | Baku acceleration (C.04.7)                    |
| Bye selection          | Lowest rank in lowest group                               | Completability pre-matching                   |
| Criteria count         | Compatibility-based (no numbered quality criteria)        | 21 criteria (C1-C4, C8-C21)                   |

## Mathematical Foundations

The Lim system uses a deterministic exchange algorithm rather than optimization-based matching. Its complexity is lower per bracket, but the three-phase processing order and floater type classification add structural complexity.

- **Exchange matching**: [Lim Exchange Algorithm](../../algorithms/lim-exchange/) covers the Art. 4 exchange procedure, scrutiny order, and cross-half matching.
- **Colour allocation**: [Colour Allocation](../../algorithms/color-allocation/) describes the shared allocation rules. The Lim system adds median-aware tiebreaking on top.
- **Floater classification**: the four types (A-D) are defined in `pairing/lim/floater.go`. Classification depends on float history and compatibility with adjacent group members.
- **Compatibility checking**: `pairing/lim/compatibility.go` implements the three-consecutive and three-imbalance constraints using colour history analysis.

## FIDE Reference

The Lim system is defined in FIDE Handbook C.04.4.3. It specifies the median-first processing order, four floater types with priority selection, exchange-based matching within score groups, compatibility constraints on colour sequences and imbalance, and the optional maxi-tournament rating restriction.
