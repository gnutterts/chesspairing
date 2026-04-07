---
title: "For Researchers"
linkTitle: "For Researchers"
weight: 5
description: "Entry point for mathematicians and computer scientists interested in the algorithms behind tournament pairing."
---

Chess tournament pairing is a constrained combinatorial optimization problem. Given a set of players with game histories, ratings, colour histories, and various eligibility constraints, the task is to produce a set of pairings that satisfies hard constraints (no repeat opponents, no third consecutive same-colour game) while optimizing a lexicographic objective over a dozen or more soft criteria (score-group homogeneity, colour equalization, minimizing float distance, rating-order preservation).

Chesspairing solves this problem for all current FIDE-regulated pairing systems, three scoring systems, and 25 tiebreakers. Everything is implemented in pure Go with zero external dependencies -- the source code is the single source of truth for every algorithm described below.

This page surveys the key algorithmic components and points you to the detailed write-ups in the [Algorithms](/docs/algorithms/) section.

## Maximum weight matching via Edmonds' Blossom algorithm

The core of the Dutch, Burstein, and Dubov Swiss pairers is a reduction to maximum weight matching in a general (non-bipartite) graph. Each eligible player pair becomes an edge, and the pairing criteria are encoded into the edge weight such that the maximum weight matching corresponds to the optimal pairing.

Chesspairing includes a full implementation of Edmonds' Blossom algorithm (O(n^3)), ported from Joris van Rantwijk's Python reference. Two variants are provided:

- **`MaxWeightMatching`** -- operates on `int64` edge weights.
- **`MaxWeightMatchingBig`** -- operates on `*big.Int` edge weights, which is necessary in practice because the edge weight encoding exceeds 64 bits for realistic tournament sizes.

See [Blossom Algorithm](/docs/algorithms/blossom/).

## Edge weight encoding

The FIDE Dutch system defines over 16 pairing criteria arranged in strict priority order (C1 through C21, with some criteria being absolute constraints and others optimization targets). Rather than running the Blossom algorithm once per criterion level with iterative fixing, chesspairing packs all criteria into a single `*big.Int` per edge. Each criterion occupies a fixed-width bit field, and the fields are ordered from most-significant (highest priority) to least-significant (lowest priority).

This reduces the multi-objective lexicographic optimization to a single maximum weight matching call. The bit layout is designed so that satisfying a higher-priority criterion always outweighs any combination of lower-priority criteria.

See [Edge Weight Encoding](/docs/algorithms/edge-weights/).

## Completability pre-matching (Stage 0.5)

When a tournament round has an odd number of active players, exactly one player must receive a pairing-allocated bye. The choice of bye recipient affects the feasibility and quality of the remaining pairings. Selecting the wrong player can make it impossible to pair the rest of the field without violating absolute constraints.

Chesspairing uses a completability pre-matching phase (Stage 0.5, mirroring the approach from bbpPairings) that runs a simplified Blossom matching for each bye candidate. A candidate is viable only if the remaining players can be fully paired. Among viable candidates, the one from the lowest score group with the highest pairing number is selected -- matching FIDE regulations for bye assignment.

See [Completability](/docs/algorithms/completability/).

## Lexicographic bracket pairing

The Double-Swiss (C.04.5) and Team Swiss (C.04.6) systems use a different approach from Blossom matching. Within each score group, players are split into a top half (S1) and bottom half (S2). The algorithm attempts to pair S1[1] with S2[1], S1[2] with S2[2], and so on. When a pairing is infeasible, it backtracks and produces the lexicographically smallest valid pairing by trying S2 permutations in order.

Quality criteria (colour equalization, float minimization) are evaluated for each candidate pairing and used to select among feasible alternatives. The backtracking is bounded by the score group size, keeping it practical for tournament-scale inputs.

See [Lexicographic Pairing](/docs/algorithms/lexicographic/).

## Lim exchange matching

The Lim system (C.04.4.3) takes yet another approach. It classifies players as four floater types (A through D) based on their float history, then processes score groups from the median outward. Within each score group, an exchange-based matching procedure systematically tries transpositions of the lower-ranked subgroup, accepting the first pairing that satisfies compatibility constraints.

The floater selection and exchange order are deterministic, ensuring reproducible pairings. This is a fundamentally different algorithmic structure from both the Blossom-based and lexicographic approaches.

See [Lim Exchange Matching](/docs/algorithms/lim-exchange/).

## Berger table rotation for round-robin

Round-robin pairing follows the FIDE Berger tables (C.05, Annex 1). For n players, round k is generated by a rotation of pairing numbers with player n held fixed. When n is odd, the fixed position becomes the bye slot. The implementation supports multiple cycles (double round-robin, etc.) and an optional swap of the last two rounds to improve colour balance.

See [Berger Tables](/docs/algorithms/berger-tables/).

## Varma tables for pairing number assignment

In round-robin tournaments, the initial assignment of pairing numbers to players affects colour distribution across the tournament. The Varma tables (FIDE C.05, Annex 2) provide a federation-aware assignment that avoids players from the same federation meeting in early rounds. The implementation includes the full lookup tables and a federation-aware assignment algorithm.

See [Varma Tables](/docs/algorithms/varma-tables/).

## Baku acceleration

In the opening rounds of a large Swiss tournament, the top score groups contain many players with identical scores, making the pairings within those groups somewhat arbitrary. Baku acceleration (FIDE C.04.7) assigns virtual points to top-seeded players in early rounds, creating more differentiated score groups and producing more meaningful pairings from the start. The virtual points are removed after the acceleration phase ends.

See [Baku Acceleration](/docs/algorithms/baku-acceleration/).

## Keizer scoring convergence

Keizer scoring is an iterative algorithm. Each player's score depends on the scores of their opponents (stronger opponents are worth more points), which in turn depend on the scores of their opponents, and so on. The implementation resolves this circularity by iterating: compute scores, re-rank, recompute, and repeat until the ranking stabilizes.

All arithmetic uses doubled integers to avoid floating-point issues. Convergence is guaranteed in practice but the implementation includes oscillation detection and caps the iteration count at 20. In most tournaments, the ranking stabilizes within 3-5 iterations.

See [Keizer Convergence](/docs/algorithms/keizer-convergence/).

## Colour allocation

After pairings are determined, colours must be assigned to each board. Each FIDE pairing system specifies its own colour allocation procedure with different priority rules. The general approach balances colour history, respects colour preferences and due-colour, and avoids giving the same colour three times in a row. The Dutch system uses a multi-step priority with alternation as the tiebreaker; the Team Swiss system has a 9-step procedure; and the Double-Swiss system uses a 5-step priority.

See [Colour Allocation](/docs/algorithms/color-allocation/).

## Dutch and Dubov optimization criteria

The Dutch system defines criteria C1 through C21 (absolute constraints C1-C6, optimization criteria C8-C21). The Dubov system defines its own set of ten criteria (C1-C10) with different priorities. Both sets are implemented as functions that evaluate a proposed pairing against the tournament state and contribute to edge weights.

See [Dutch Criteria](/docs/algorithms/dutch-criteria/) and [Dubov Criteria](/docs/algorithms/dubov-criteria/).

## FIDE B.02 conversion table

Several performance-based tiebreakers (TPR, PTP, APRO, APPO) require converting between expected scores and rating differences. The FIDE B.02 conversion table provides this mapping. Chesspairing includes the full table and interpolation logic.

See [FIDE B.02](/docs/algorithms/fide-b02/) and [Elo Model](/docs/algorithms/elo-model/).

## Reading the code

The entire codebase is structured for readability. Key entry points:

| Algorithm                     | Package               | Entry function                              |
| ----------------------------- | --------------------- | ------------------------------------------- |
| Blossom matching              | `algorithm/blossom`   | `MaxWeightMatching`, `MaxWeightMatchingBig` |
| Edge weight computation       | `pairing/swisslib`    | `ComputeBaseEdgeWeight`                     |
| Completability                | `pairing/swisslib`    | `PairBracketsGlobal` (Stage 0.5)            |
| Dutch pairing                 | `pairing/dutch`       | `Pair`                                      |
| Burstein pairing              | `pairing/burstein`    | `Pair`                                      |
| Dubov pairing                 | `pairing/dubov`       | `Pair`                                      |
| Lim pairing                   | `pairing/lim`         | `Pair`                                      |
| Lexicographic bracket pairing | `pairing/lexswiss`    | `PairBracket`                               |
| Double-Swiss pairing          | `pairing/doubleswiss` | `Pair`                                      |
| Team Swiss pairing            | `pairing/team`        | `Pair`                                      |
| Berger table rotation         | `pairing/roundrobin`  | `Pair`                                      |
| Varma assignment              | `algorithm/varma`     | `Groups`, `Assign`                          |
| Baku acceleration             | `pairing/swisslib`    | `AdjustScoreGroups`                         |
| Keizer scoring                | `scoring/keizer`      | `Score`                                     |
| Tiebreaker registry           | `tiebreaker`          | `Get`, `All`                                |

For a guided tour of the type system and interfaces, see the [API Overview](/docs/api/overview/) and [Core Types](/docs/api/core-types/).

## Next steps

- [Algorithms](/docs/algorithms/) -- mathematical deep-dives with formulas, pseudocode, and proof sketches
- [API Reference](/docs/api/) -- the Go type system, interfaces, and package organization
- [Go Quickstart](../go-quickstart/) -- use the library directly in Go code
- [Pairing Systems](/docs/pairing-systems/) -- regulation-level documentation for each pairing engine
