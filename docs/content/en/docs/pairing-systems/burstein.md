---
title: "Burstein System"
linkTitle: "Burstein"
weight: 2
description: "FIDE C.04.4.2 — a Swiss variant with seeding rounds and opposition-index re-ranking."
---

The Burstein system splits a Swiss tournament into two distinct phases. In the opening rounds -- the seeding rounds -- pairings work exactly like the Dutch system, using initial rankings to create the draw. After that phase ends, every player is re-ranked according to an opposition index that combines Buchholz and Sonneborn-Berger scores. The idea is straightforward: once enough games have been played, your ranking should reflect the strength of the opponents you actually faced and how well you did against them, not just your pre-tournament rating. From that point on, the re-ranked order drives all bracket construction and pairing decisions.

Under the hood, Burstein uses the same global Blossom matching infrastructure as the Dutch system, but it deliberately strips away several layers of optimization. There are no top-scorer rules, no C8 look-ahead into future brackets, and no float optimization criteria. The result is a simpler system that trades some of the Dutch engine's fine-grained control for a fundamentally different approach to ranking fairness.

## When to Use

- **Tournaments where early results should reshape rankings.** If you want the second half of the tournament to pair players based on who they actually played rather than their initial seeding, Burstein is designed for exactly this.
- **Events where float optimization is less important.** The simplified criteria (colour only, no float history) make the system lighter and the pairings easier to explain.
- **As a FIDE-compliant alternative to Dutch.** Burstein is an approved FIDE system (C.04.4.2) and can be used wherever regulations permit a Swiss variant.

For the full 21-criteria treatment, use [Dutch](../dutch/). For ascending-ARO processing, see [Dubov](../dubov/). For median-first processing with floater types, see [Lim](../lim/).

## Configuration

### CLI

```bash
chesspairing pair --burstein tournament.trf
```

The `--burstein` flag selects the Burstein pairing engine. Output format is controlled with `--format` (list, wide, board, xml, json) or the `-w` shorthand for wide output.

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/burstein"

// From typed options
p := burstein.New(burstein.Options{
    Acceleration:   chesspairing.StringPtr("baku"),
    TopSeedColor:   chesspairing.StringPtr("auto"),
    TotalRounds:    chesspairing.IntPtr(9),
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})

// From a generic map (e.g. parsed JSON config)
p := burstein.NewFromMap(map[string]any{
    "acceleration": "none",
    "topSeedColor": "auto",
    "totalRounds":  9.0,
    "forbiddenPairs": []any{[]any{"P1", "P2"}},
})

result, err := p.Pair(ctx, &state)
```

### Options

| Option           | Type         | Default  | Description                                                                                                                 |
| ---------------- | ------------ | -------- | --------------------------------------------------------------------------------------------------------------------------- |
| `Acceleration`   | `*string`    | `"none"` | `"none"` or `"baku"`. Baku acceleration (FIDE C.04.7) adds virtual points in early rounds.                                  |
| `TopSeedColor`   | `*string`    | `"auto"` | `"auto"`, `"white"`, or `"black"`. Controls which colour the top seed receives in round 1.                                  |
| `ForbiddenPairs` | `[][]string` | `nil`    | Pairs of player IDs that must never be paired together, enforced as an absolute criterion.                                  |
| `TotalRounds`    | `*int`       | derived  | Planned total rounds in the tournament. Used to compute the seeding round count. If nil, derived from the tournament state. |

All fields follow the pointer-nil pattern: a nil value means "use the default". Call `WithDefaults()` explicitly or let `New()` do it for you.

### Errors

| Error                  | Condition                                              |
| ---------------------- | ------------------------------------------------------ |
| `ErrTooFewPlayers`     | Fewer than 2 active players in the tournament state.   |
| `ErrNoPairingPossible` | No valid pairing exists given the current constraints. |

## How It Works

### 1. Build player states

Every active player's tournament history is compiled into a `PlayerState`: score, colour history, opponent list, float history, bye status, and tournament pairing number (TPN).

### 2. Determine the current phase

The tournament is divided into seeding rounds and post-seeding rounds:

```text
SeedingRounds = min(floor(TotalRounds / 2), 4)
```

For a 9-round tournament, the first 4 rounds are seeding rounds. For a 7-round tournament, the first 3 are. The formula caps at 4 regardless of tournament length.

| Total rounds | Seeding rounds |
| ------------ | -------------- |
| 3            | 1              |
| 5            | 2              |
| 7            | 3              |
| 9+           | 4              |

### 3. Seeding rounds: TPN-based ranking

During seeding rounds, players keep their original TPN order. Bracket construction and matching work the same as in the Dutch system. This phase produces a baseline of results that the opposition index can later work with.

### 4. Post-seeding rounds: opposition-index re-ranking

After the seeding phase, players are re-ranked using `RankByOppositionIndex()`. The re-ranking uses three tiebreakers applied in order:

1. **Score** (descending) -- same primary sort as standard ranking.
2. **Buchholz** (descending) -- sum of all opponents' scores. Higher Buchholz means you faced stronger opposition overall.
3. **Sonneborn-Berger** (descending) -- sum of (your result against each opponent multiplied by that opponent's score). Rewards winning against strong opponents more than winning against weak ones.
4. **Original TPN** (ascending) -- breaks any remaining ties.

After sorting, new TPN values are assigned sequentially (1, 2, 3, ...). These re-assigned TPNs drive all subsequent bracket construction and S1/S2 splits.

Forfeit games are excluded from the Sonneborn-Berger computation. Buchholz includes all opponents (including inactive players) to avoid penalizing a player whose opponent withdrew.

### 5. Apply Baku acceleration (optional)

Same as the Dutch system: when `Acceleration` is `"baku"`, virtual points are added to Group A players' pairing scores. See [Baku Acceleration](/docs/algorithms/baku-acceleration/) for the calculation details.

### 6. Build score groups

Players are partitioned into score groups ordered from highest to lowest. Within each group, players are sorted by TPN ascending -- which in post-seeding rounds means sorted by opposition index.

### 7. Global Blossom matching

The matching uses the same `PairBracketsGlobal` infrastructure as the Dutch system (completability pre-matching for odd counts, incremental bracket processing), but with key differences in the criteria context:

- **No top-scorer rules.** The `TopScorers` map is empty, so final-round colour protection for leading players does not apply.
- **No C8 look-ahead.** The `LookAhead` function is not set, so there is no check that floaters allow the next bracket to be pairable. This simplifies matching at the cost of occasionally producing suboptimal float distributions.
- **Colour criteria only (C10-C13).** The edge weights encode absolute colour violations, strong colour preference satisfaction, mild colour preference satisfaction, and colour imbalance minimization. Float criteria C14-C21 are not used.

The Blossom matching still produces globally optimal pairings within the reduced criteria set.

### 8. Colour allocation

Colours are allocated using the same six-priority algorithm as the Dutch system, but without top-scorer rules. The `topScorerRules` parameter is set to `false`, so the final-round special handling for leading players is skipped.

### 9. Bye assignment

If the player count is odd, the single unmatched player from the Blossom matching receives a pairing-allocated bye (PAB).

## Comparison with Other Systems

| Aspect                    | Burstein                               | [Dutch](../dutch/) | [Dubov](../dubov/)  | [Lim](../lim/)                |
| ------------------------- | -------------------------------------- | ------------------ | ------------------- | ----------------------------- |
| **Ranking method**        | TPN in seeding, opposition index after | TPN throughout     | ARO-based           | TPN throughout                |
| **Seeding rounds**        | Yes (up to 4)                          | No                 | No                  | No                            |
| **Matching algorithm**    | Global Blossom                         | Global Blossom     | Transposition-based | Exchange-based                |
| **Optimization criteria** | C10-C13 (colour only)                  | C1-C21 (full)      | C1-C10              | Compatibility + floater types |
| **C8 look-ahead**         | No                                     | Yes                | No                  | No                            |
| **Float criteria**        | None                                   | C14-C21            | None                | Floater types A-D             |
| **Top-scorer rules**      | No                                     | Yes (final round)  | No                  | No                            |
| **Extra options**         | `TotalRounds`                          | --                 | `TotalRounds`       | `MaxiTournament`              |

The Burstein system occupies a middle ground: it uses the same powerful Blossom matching engine as Dutch but applies a smaller, more focused set of criteria. The opposition-index re-ranking is its distinguishing feature -- no other Swiss variant in this library re-orders players based on opponent strength after the opening phase.

## Mathematical Foundations

The Burstein pairer shares most of its algorithmic infrastructure with the Dutch system:

- **[Blossom Matching](/docs/algorithms/blossom/)** -- Edmonds' O(n^3) maximum weight matching, using `*big.Int` edge weights.
- **[Edge Weight Encoding](/docs/algorithms/edge-weights/)** -- Same bit-packed encoding as Dutch, but only the colour-related fields (C10-C13) carry meaningful weight; float fields (C14-C21) are zeroed.
- **[Completability Pre-matching](/docs/algorithms/completability/)** -- Stage 0.5 bye determination, identical to Dutch.
- **[Baku Acceleration](/docs/algorithms/baku-acceleration/)** -- Same virtual point system when enabled.
- **[Colour Allocation](/docs/algorithms/color-allocation/)** -- Same six-priority procedure, minus top-scorer rules.

### Opposition Index

The opposition index is computed per player as a tuple (Buchholz, Sonneborn-Berger, TPN). Players are re-sorted by score first, then by this tuple.

**Buchholz** is the sum of all opponents' standard pairing scores (1-0.5-0):

```text
Buchholz(i) = sum over all opponents j of Score(j)
```

**Sonneborn-Berger** weights each opponent's score by the result achieved against them:

```text
SB(i) = sum over all games g of Result(i, g) * Score(opponent(i, g))
```

where `Result(i, g)` is 1 for a win, 0.5 for a draw, and 0 for a loss. Forfeit games are excluded.

The combination ensures that among players with equal scores, those who faced and beat stronger opposition are ranked higher. This produces more meaningful bracket compositions in the post-seeding phase than raw TPN order would.

## FIDE Reference

The Burstein system is defined in FIDE regulation C.04.4.2. The implementation covers:

- **C.04.4.2 Article 1** -- Definition of seeding rounds and the seeding round formula.
- **C.04.4.2 Article 2** -- Opposition index computation (Buchholz + Sonneborn-Berger) and re-ranking procedure for post-seeding rounds.
- **C.04.4.2 Article 3** -- Pairing procedure using score brackets with S1/S2 splits, sharing the absolute criteria (C1-C4) with the Dutch system.
- **C.04.4.2 Article 4** -- Optimization limited to colour criteria (C10-C13); float optimization criteria (C14-C21) are not applied.
- **C.04.7** -- Baku acceleration (shared with Dutch, optional).

The system delegates to the same Blossom matching infrastructure as the Dutch engine, with the criteria context configured to reflect the Burstein-specific rules: empty top-scorer map, no look-ahead function, and colour-only optimization weights.
