---
title: "Dubov System"
linkTitle: "Dubov"
weight: 3
description: "FIDE C.04.4.1 — ascending ARO processing with transposition-based matching."
---

## Overview

The Dubov system is a Swiss variant that aims to equalize the Average Rating of Opponents (ARO) across players in the same score group. Where the [Dutch system](../dutch/) splits brackets by pairing number and uses global Blossom matching, Dubov splits by colour preference (G1/G2) and sorts the white-seeking half by ascending ARO. Matching is done through transpositions of the opponent group rather than weighted graph optimization.

If you are coming from the Dutch system, the most visible difference as a player is that your opponent selection depends on who you have already faced (through ARO) rather than solely on your tournament pairing number.

## When to Use

- Tournaments where you want to spread strong opponents more evenly across the field, rather than concentrating them at the top boards.
- Events where ARO equalization is a priority, such as round-robin replacement formats or norm-seeking tournaments.
- Situations where the simpler 10-criterion model (vs. 21 in Dutch) is preferred for transparency.

The Dubov system does **not** support Baku acceleration. If you need accelerated pairings, use the [Dutch](../dutch/) or [Burstein](../burstein/) system instead.

## Configuration

### CLI

```bash
chesspairing pair --dubov tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/dubov"

// With typed options
p := dubov.New(dubov.Options{
    TopSeedColor: chesspairing.StringPtr("auto"),
    TotalRounds:  chesspairing.IntPtr(9),
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})

// From a generic map (e.g., parsed from JSON config)
p := dubov.NewFromMap(map[string]any{
    "topSeedColor": "black",
    "totalRounds":  9.0,
})
```

### Options

| Option           | Type         | Default  | Description                                                                          |
| ---------------- | ------------ | -------- | ------------------------------------------------------------------------------------ |
| `TopSeedColor`   | `string`     | `"auto"` | Colour for the top seed in round 1. Values: `"auto"`, `"white"`, `"black"`.          |
| `ForbiddenPairs` | `[][]string` | `nil`    | Pairs of player IDs that must never be paired together.                              |
| `TotalRounds`    | `int`        | derived  | Planned total rounds. Used for internal calculations; derived from state if not set. |

The Dubov system has no `Acceleration` option. Baku acceleration is not part of the C.04.4.1 specification.

## How It Works

### Step-by-step algorithm

1. **Build player states** from tournament history (game results, colour history, float history, opponents).

2. **Bye selection** (odd player count): The `DubovByeSelector` picks the bye player before matching begins. Selection criteria, in order: lowest score, most games played, highest TPN (lowest-ranked). The bye player is removed from the pool before brackets are formed.

3. **Build score groups and brackets** from the remaining players.

4. **Build rating map** for ARO computation.

5. **Process brackets top-down.** For each bracket:
   - **Split into G1 and G2.** In round 1, G1 is the first half by TPN and G2 is the rest. In later rounds, G1 contains players preferring White, G2 contains players preferring Black or having no preference. Groups are balanced so `|G1| = floor(n/2)`.
   - **Sort G1 by ascending ARO** (ties broken by ascending TPN). This is the defining characteristic of the Dubov system: the player with the lowest ARO in G1 is paired first.
   - **Generate G2 transpositions** (up to 120 permutations using Narayana Pandita's next-permutation algorithm). Each transposition proposes sequential pairings: G1[0] vs G2[0], G1[1] vs G2[1], etc.
   - **Evaluate each transposition** against absolute criteria (C1, C3, forbidden pairs). If any pair violates an absolute criterion, the entire transposition is rejected. Valid transpositions are scored using criteria C4-C10.
   - **Select the best transposition** by comparing candidate scores lexicographically.

6. **Upward collapse on failure.** If a bracket produces only floaters and no pairs, it collapses into the adjacent bracket and matching is retried.

7. **MaxT upfloater limit**: `2 + floor(CompletedRounds / 5)`. This limits how many times a player can be floated upward before the engine penalizes further floats.

8. **Board ordering**: max player score descending, bracket score descending, min TPN ascending.

9. **Colour allocation** via Dubov's 5-rule algorithm (Art. 5). This delegates to the shared `swisslib.AllocateColor` without top-scorer-specific rules.

### Bye selection details

The Dubov bye rule differs from the Dutch system's completability-based approach. Instead of analyzing which player removal leads to the best overall matching, Dubov selects deterministically:

1. Lowest score
2. Most games played (among tied scores)
3. Highest TPN / lowest rank (final tiebreak)

Players who have already received a PAB are excluded.

### Criteria

The Dubov system uses 10 criteria, compared to 21 in the Dutch system:

| Criterion | Type     | Description                                                                    |
| --------- | -------- | ------------------------------------------------------------------------------ |
| C1        | Absolute | No rematches (players must not have already played each other)                 |
| C3        | Absolute | No absolute colour conflicts (both players needing the same colour absolutely) |
| C4        | Quality  | Minimize upfloater count                                                       |
| C5        | Quality  | Maximize upfloater score sum (prefer floating higher-scored players)           |
| C6        | Quality  | Minimize colour preference violations                                          |
| C7        | Quality  | Minimize upfloaters at or above MaxT                                           |
| C8        | Quality  | Minimize consecutive-round upfloaters                                          |
| C9        | Quality  | Minimize upfloater opponents at or above MaxT                                  |
| C10       | Quality  | Minimize consecutive-round MaxT violations                                     |

Candidate comparison order: C4, then C5 (reversed -- higher is better), then C6-C10 lexicographically, then transposition index as final tiebreak.

### Errors

| Error                  | Condition                                         |
| ---------------------- | ------------------------------------------------- |
| `ErrTooFewPlayers`     | Fewer than 1 active player                        |
| `ErrNoPairingPossible` | No valid pairing exists for the remaining players |

## Comparison with Dutch

| Aspect          | Dubov                                      | Dutch                                    |
| --------------- | ------------------------------------------ | ---------------------------------------- |
| Group split     | G1/G2 by colour preference                 | S1/S2 by TPN (top half / bottom half)    |
| G1 sort order   | Ascending ARO, then TPN                    | Descending score, then ascending TPN     |
| Matching        | Transposition-only (cap 120)               | Global Blossom matching                  |
| Exchanges       | None                                       | S1/S2 exchanges allowed                  |
| Criteria        | 10 (C1-C10)                                | 21 (C1-C4, C8-C21)                       |
| Bye selection   | Deterministic (lowest ARO in lowest group) | Completability pre-matching              |
| Acceleration    | Not supported                              | Baku acceleration (C.04.7)               |
| Upfloater limit | MaxT = 2 + floor(Rnds/5)                   | Blossom edge weights handle floater cost |

## Mathematical Foundations

The Dubov system's core mechanism is simpler than the Dutch system's Blossom-based approach. Instead of computing optimal weighted matchings, it enumerates G2 permutations and evaluates each against a fixed criteria hierarchy.

- **ARO computation**: arithmetic mean of opponents' ratings, excluding forfeits. See `pairing/dubov/aro.go`.
- **Transposition generation**: Narayana Pandita's algorithm for lexicographic next-permutation, capped at 120 transpositions per bracket.
- **Criteria scoring**: the `DubovCandidateScore` type implements a total ordering over C4-C10 violations with lexicographic comparison.
- **Dubov criteria details**: [Dubov Criteria](../../algorithms/dubov-criteria/) covers the full C1-C10 specification.
- **Colour allocation**: [Colour Allocation](../../algorithms/color-allocation/) describes the shared allocation rules used across Swiss variants.

## FIDE Reference

The Dubov system is defined in FIDE Handbook C.04.4.1. It specifies the ARO-equalization approach, G1/G2 colour-preference splitting, transposition-based matching, the 10-criterion evaluation model, and the MaxT upfloater limit formula.
