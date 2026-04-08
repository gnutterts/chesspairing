---
title: "Dutch System"
linkTitle: "Dutch"
weight: 1
description: "The standard FIDE Swiss pairing system (C.04.3) — global Blossom matching with 21 optimization criteria."
---

If you have ever played in a rated Swiss tournament, the Dutch system almost certainly decided who you sat across from. It is the default FIDE Swiss algorithm and the one most arbiters and players take for granted. The engine groups everyone by score, splits each group into an upper and lower half, and tries to pair across the halves while respecting colour history, avoiding rematches, and distributing floaters as fairly as the maths allow. Behind that simple description sits a global Blossom matching graph, 21 layered optimization criteria, and edge weights that require arbitrary-precision integers to encode.

## When to Use

- **Standard rated tournaments.** Any FIDE-rated or nationally rated Swiss event that does not specify an alternative system.
- **Open and closed Swiss events** of any size, from a weekend rapid with 20 players to an open with 500+.
- **When you need maximum pairing quality.** The 21 criteria and global matching produce pairings that satisfy FIDE rules as closely as mathematically possible.
- **When Baku acceleration is wanted.** The Dutch engine natively supports FIDE C.04.7 acceleration for early-round diversity.

For events that need opposition-index re-ranking after an initial phase, consider [Burstein](../burstein/). For ARO-ordered processing, see [Dubov](../dubov/). For median-first floater logic, see [Lim](../lim/).

## Configuration

### CLI

```bash
chesspairing pair --dutch tournament.trf
```

The `--dutch` flag selects the Dutch pairing engine. Output format is controlled with `--format` (list, wide, board, xml, json) or the `-w` shorthand for wide output.

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/dutch"

// From typed options
p := dutch.New(dutch.Options{
    Acceleration:   chesspairing.StringPtr("baku"),
    TopSeedColor:   chesspairing.StringPtr("white"),
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})

// From a generic map (e.g. parsed JSON config)
p := dutch.NewFromMap(map[string]any{
    "acceleration": "baku",
    "topSeedColor": "white",
    "forbiddenPairs": []any{[]any{"P1", "P2"}},
})

result, err := p.Pair(ctx, &state)
```

### Options

| Option           | Type         | Default  | Description                                                                                                             |
| ---------------- | ------------ | -------- | ----------------------------------------------------------------------------------------------------------------------- |
| `Acceleration`   | `*string`    | `"none"` | `"none"` or `"baku"`. Baku acceleration (FIDE C.04.7) adds virtual points in early rounds to mix rating tiers sooner.   |
| `TopSeedColor`   | `*string`    | `"auto"` | `"auto"`, `"white"`, or `"black"`. Controls which colour the top seed receives in round 1. Subsequent boards alternate. |
| `ForbiddenPairs` | `[][]string` | `nil`    | Pairs of player IDs that must never be paired together, enforced as an absolute criterion alongside C1 and C3.          |

All fields follow the pointer-nil pattern: a nil value means "use the default". Call `WithDefaults()` explicitly or let `New()` do it for you.

### Errors

| Error                  | Condition                                              |
| ---------------------- | ------------------------------------------------------ |
| `ErrTooFewPlayers`     | Fewer than 2 active players in the tournament state.   |
| `ErrNoPairingPossible` | No valid pairing exists given the current constraints. |

## How It Works

The Dutch pairer follows seven steps, matching the architecture used by bbpPairings:

### 1. Build player states

Every active player's tournament history is compiled into a `PlayerState`: score, colour history, opponent list, float history, bye status, and tournament pairing number (TPN).

### 2. Apply Baku acceleration (optional)

When `Acceleration` is set to `"baku"`, virtual points are added to each player's pairing score according to FIDE C.04.7:

- **Group A size** = 2 \* ceil(N / 4), where N is the total number of players.
- **Accelerated rounds** = ceil(totalRounds / 2). The first half of these use 1.0 virtual point; the second half use 0.5.
- Only Group A players (those with initial rank within GA size) receive virtual points.

This pushes top-rated players into different score brackets in early rounds, preventing them from all clustering at the top immediately.

### 3. Build score groups

Players are partitioned into score groups ordered from highest score to lowest. Within each group, players are sorted by TPN ascending (strongest first).

### 4. Global Blossom matching

This is the core of the algorithm. Rather than pairing each bracket in isolation (which can produce suboptimal results), the Dutch engine builds a single global matching graph containing all players and runs Edmonds' maximum weight Blossom algorithm to find the optimal pairing.

The process has two stages:

**Stage 0.5 -- Completability pre-matching** (odd player counts only). A simplified Blossom matching determines which player will receive the pairing-allocated bye (PAB). The simplified edge weights encode bye eligibility, score maximization, and top-scorer protection. The unmatched player's score feeds into the real edge weights.

**Main matching -- 7-phase bracket loop.** Score groups are processed top-down. For each bracket, edges are inserted into the global graph with weights encoding all 21 criteria. The Blossom algorithm runs incrementally, committing pairs from the current bracket before moving to the next. This incremental approach mirrors bbpPairings' `computeMatching` procedure.

### 5. Board ordering

Committed pairs are sorted for board assignment:

1. **Max score descending** -- the pair containing the higher-scoring player comes first.
2. **Bracket score descending** -- among pairs with the same max player score, homogeneous pairs (both players native to the bracket) come before heterogeneous pairs (one player floated in).
3. **Min TPN ascending** -- ties are broken by the stronger player's TPN (lower number = higher board).

### 6. Colour allocation

Colours are assigned using a six-priority algorithm that mirrors bbpPairings' `choosePlayerNeutralColor` and `choosePlayerColor`:

1. Compatible preferences -- both players' preferences can be satisfied simultaneously.
2. Absolute preference wins -- a player with colour imbalance > 1 or 2+ consecutive same colour gets priority.
3. Strong preference beats non-strong -- imbalance > 0 (but not absolute) outranks a mild preference.
4. First colour difference -- walk backwards through both players' colour histories and swap from the most recent round where they differed.
5. Same-colour conflict -- when both want the same colour at equal strength, the higher-ranked player gets their preference.
6. No preference -- alternate by board number (higher-ranked gets White on odd boards by default, controlled by `TopSeedColor`).

In the final round, top-scorer rules apply: players with more than 50% of the maximum possible score receive special consideration to avoid colour-based competitive disadvantage.

### 7. Bye assignment

If the player count is odd, the single unmatched player from the Blossom matching receives a pairing-allocated bye (PAB).

## Comparison with Other Systems

| Aspect                 | Dutch                      | [Burstein](../burstein/)               | [Dubov](../dubov/)  | [Lim](../lim/)                |
| ---------------------- | -------------------------- | -------------------------------------- | ------------------- | ----------------------------- |
| **Matching algorithm** | Global Blossom             | Global Blossom                         | Transposition-based | Exchange-based                |
| **Criteria count**     | 21 (C1-C21)                | C1-C4, C10-C13 only                    | 10 (C1-C10)         | Compatibility + floater types |
| **S1/S2 splits**       | Yes (upper/lower half)     | Yes                                    | G1/G2 split         | No (exchange within group)    |
| **C8 look-ahead**      | Yes (MatchBracketFeasible) | No                                     | No                  | No                            |
| **Float criteria**     | C14-C21 (full)             | None                                   | No float criteria   | Floater types A-D             |
| **Top-scorer rules**   | Yes (final round)          | No                                     | No                  | No                            |
| **Player ranking**     | TPN throughout             | TPN in seeding, opposition index after | ARO-based           | TPN throughout                |
| **Processing order**   | Top-down by score          | Top-down by score                      | Ascending ARO       | Median-first                  |

The Dutch system is the most comprehensive in terms of optimization criteria. Burstein intentionally strips away float criteria and top-scorer rules in favour of opposition-index re-ranking. Dubov and Lim use fundamentally different matching strategies (transposition and exchange, respectively) instead of Blossom matching.

## Mathematical Foundations

The Dutch pairer relies on several algorithms documented in the [Algorithms](/docs/algorithms/) section:

- **[Blossom Matching](/docs/algorithms/blossom/)** -- Edmonds' O(n^3) maximum weight matching for general graphs. The `algorithm/blossom/` package provides both `int64` and `*big.Int` variants.
- **[Edge Weight Encoding](/docs/algorithms/edge-weights/)** -- The 16+ criteria fields are packed into a single `*big.Int` edge weight using positional bit encoding. Higher-priority criteria occupy more significant bits, so the Blossom algorithm naturally prefers pairings that satisfy the most important criteria.
- **[Completability Pre-matching](/docs/algorithms/completability/)** -- Stage 0.5 uses a simplified Blossom run with reduced edge weights to determine the bye recipient before the main matching.
- **[Dutch Criteria](/docs/algorithms/dutch-criteria/)** -- Detailed breakdown of all 21 criteria: C1-C4 (absolute), C5-C7 (quality), C8 (look-ahead), C9 (bye assignee), C10-C13 (colour optimization), C14-C21 (float optimization).
- **[Baku Acceleration](/docs/algorithms/baku-acceleration/)** -- Virtual point calculation, Group A sizing, and round classification.
- **[Colour Allocation](/docs/algorithms/color-allocation/)** -- The six-priority colour assignment procedure.

### Why big.Int?

Each edge weight encodes 16+ fields across score-group-sized bit ranges. For a tournament with many score groups, the total bit width easily exceeds 64 bits. The `algorithm/blossom/` package provides `MaxWeightMatchingBig` specifically for this case. The int64 variant is used only in completability pre-matching where the simplified weights fit in 64 bits.

## FIDE Reference

The Dutch system is defined in FIDE regulation C.04.3. The implementation covers:

- **C.04.3 Article 1** -- Definitions (score bracket, score group, pairing bracket, S1/S2 halves, heterogeneous brackets, floaters).
- **C.04.3 Article 2** -- Absolute criteria C1-C4 (no rematches, no second bye, colour limits, forbidden pairs).
- **C.04.3 Article 3** -- Quality criteria C5-C7 (maximize pairs per bracket, maximize paired scores, minimize score differences).
- **C.04.3 Article 4** -- C8 look-ahead (floaters must allow the next bracket to be pairable).
- **C.04.3 Article 5** -- Optimization criteria C9-C21 (bye placement, colour preferences, float history).
- **C.04.3 Annex A** -- Board ordering and initial colour allocation rules.
- **C.04.7** -- Baku acceleration (virtual points, Group A, accelerated round count).

The S1/S2 half-split, Narayana Pandita transposition order, and combination-based exchange enumeration follow the procedures described in the FIDE handbook for deterministic traversal of candidate pairings within each bracket.
