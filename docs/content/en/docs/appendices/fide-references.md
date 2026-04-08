---
title: "FIDE Handbook References"
linkTitle: "FIDE References"
weight: 1
description: "Links to relevant FIDE Handbook sections for each pairing and scoring system."
---

This page lists the FIDE Handbook sections relevant to each pairing system, scoring rule, and tiebreaker implemented in the chesspairing module. Links use the format `https://handbook.fide.com/chapter/CXXXX` where applicable, but exact URLs may change over time. The section numbers are provided for manual lookup.

## Pairing Systems

| System            | FIDE Section | Handbook Chapter                                      |
| ----------------- | ------------ | ----------------------------------------------------- |
| Dutch Swiss       | C.04.3       | General Rules, Dutch System                           |
| Burstein Swiss    | C.04.4.2     | Systems based on the Dutch System, Burstein System    |
| Dubov Swiss       | C.04.4.1     | Systems based on the Dutch System, Dubov System       |
| Lim Swiss         | C.04.4.3     | Systems based on the Dutch System, Lim System         |
| Double-Swiss      | C.04.5       | Double-Swiss System                                   |
| Team Swiss        | C.04.6       | Team Swiss System                                     |
| Baku Acceleration | C.04.7       | Accelerated pairings for the Swiss System             |
| Round-Robin       | C.05         | Round-Robin System                                    |
| Berger Tables     | C.05 Annex 1 | Pairing tables for Round-Robin tournaments            |
| Varma Tables      | C.05 Annex 2 | Initial number assignment for Round-Robin tournaments |

### Dutch Swiss (C.04.3)

The Dutch system is the most widely used Swiss pairing method. It defines absolute criteria (C1--C4) that must be satisfied, and optimization criteria (C5--C21) that the algorithm maximizes. The chesspairing implementation uses a global Blossom matching architecture with a 7-phase bracket loop.

Link: [https://handbook.fide.com/chapter/C0403](https://handbook.fide.com/chapter/C0403)

### Burstein Swiss (C.04.4.2)

The Burstein system is a Dutch variant that distinguishes between seeding rounds and post-seeding rounds. In post-seeding rounds, players are re-ranked using an opposition index derived from Buchholz and Sonneborn-Berger values.

Link: [https://handbook.fide.com/chapter/C04042](https://handbook.fide.com/chapter/C04042)

### Dubov Swiss (C.04.4.1)

The Dubov system uses Average Rating of Opponents (ARO) for sorting within score groups and defines its own set of criteria (C1--C10). It uses ascending-ARO sorting and a transposition-based matching approach.

Link: [https://handbook.fide.com/chapter/C04041](https://handbook.fide.com/chapter/C04041)

### Lim Swiss (C.04.4.3)

The Lim system processes score groups in median-first order and uses exchange-based matching. It classifies floaters into four types (A--D) and defines specific compatibility constraints.

Link: [https://handbook.fide.com/chapter/C04043](https://handbook.fide.com/chapter/C04043)

### Double-Swiss (C.04.5)

The Double-Swiss system uses lexicographic bracket pairing and a 5-step colour allocation priority. It is designed for tournaments where players play two games per round against different opponents.

Link: [https://handbook.fide.com/chapter/C0405](https://handbook.fide.com/chapter/C0405)

### Team Swiss (C.04.6)

The Team Swiss system adapts Swiss pairing for team competitions. It supports configurable colour preference types (A, B, or None) and uses a 9-step colour allocation process.

Link: [https://handbook.fide.com/chapter/C0406](https://handbook.fide.com/chapter/C0406)

### Baku Acceleration (C.04.7)

Baku acceleration assigns virtual points in early rounds to separate top-seeded players, reducing the number of decisive games between top players in the opening rounds. The chesspairing module implements this as an option for the Dutch and Burstein pairers.

Link: [https://handbook.fide.com/chapter/C0407](https://handbook.fide.com/chapter/C0407)

### Round-Robin (C.05)

The Round-Robin system pairs every player against every other player. The chesspairing module implements FIDE Berger tables for scheduling and supports single and double round-robin formats with configurable colour balancing.

Link: [https://handbook.fide.com/chapter/C05](https://handbook.fide.com/chapter/C05)

### Berger Tables (C.05 Annex 1)

Berger tables define the standard pairing schedule for Round-Robin tournaments. The tables specify which players meet in each round and which player has White.

### Varma Tables (C.05 Annex 2)

Varma tables provide a federation-aware initial number assignment for Round-Robin tournaments. This ensures players from the same federation are distributed across rounds as evenly as possible.

## Scoring and Tiebreakers

| Topic                   | FIDE Section | Description                                                   |
| ----------------------- | ------------ | ------------------------------------------------------------- |
| General scoring rules   | C.02         | Defines standard point allocation for wins, draws, and losses |
| Tie-breaking procedures | B.02         | Defines approved tiebreaker methods and their computation     |

### Standard Scoring (C.02)

FIDE C.02 covers the general rules for scoring in chess competitions. The chesspairing module implements standard scoring (1--0.5--0) with configurable point values for wins, draws, losses, byes, forfeits, and absences.

### Tiebreakers (B.02)

FIDE B.02 defines approved tie-breaking procedures. The chesspairing module implements 25 tiebreakers, including all commonly used FIDE methods. See the [Tiebreakers](/docs/tiebreakers/) documentation for the full list.

Link: [https://handbook.fide.com/chapter/B02](https://handbook.fide.com/chapter/B02)

## Related Regulations

- **C.01 -- General Rules for Competitions**: Covers the overarching rules that apply to all FIDE-rated competitions, including definitions and general procedures.
- **C.02 -- Scoring**: Defines how game results translate to points.
- **FIDE Rating Regulations**: Separate from the pairing and scoring rules, these govern how individual player ratings are calculated and updated. The chesspairing module does not compute ratings but uses player ratings as input for pairing and tiebreaking.

## Finding FIDE Regulations

The FIDE Handbook is available at [https://handbook.fide.com](https://handbook.fide.com). Regulations are organized by chapter number. If a direct link above does not resolve, navigate to the handbook and search by section number (e.g., C.04.3).
