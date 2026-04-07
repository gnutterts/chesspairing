---
title: "For Arbiters"
linkTitle: "For Arbiters"
weight: 4
description: "A practical guide for chess arbiters who want to understand how chesspairing implements FIDE regulations."
---

Chesspairing is a tournament pairing engine. Given a list of players and their results so far, it produces the pairings for the next round, computes standings, and resolves ties -- all according to the FIDE regulations you already know.

This page maps arbiter-level concerns to the relevant parts of the documentation. You do not need to be a programmer to follow it, though the [CLI Quickstart](../cli-quickstart/) is the fastest way to see the tool in action.

## What it does

Chesspairing handles three jobs:

1. **Pairing** -- deciding who plays whom in the next round, including bye assignment when the player count is odd.
2. **Scoring** -- converting game results into point totals with configurable point values for wins, draws, byes, forfeits, and absences.
3. **Tiebreaking** -- computing tiebreak values to produce final standings when players share the same score.

These three jobs are independent of each other. You can pair a tournament with the Dutch system and score it with Keizer points, or run a round-robin with standard 1-half-0 scoring. The engine does not impose a fixed combination.

## Supported FIDE pairing systems

Chesspairing implements all current FIDE-approved Swiss pairing systems, plus round-robin and Keizer:

| System       | FIDE regulation               | Documentation                                       |
| ------------ | ----------------------------- | --------------------------------------------------- |
| Dutch        | C.04.3                        | [Dutch](/docs/pairing-systems/dutch/)               |
| Burstein     | C.04.4.2                      | [Burstein](/docs/pairing-systems/burstein/)         |
| Dubov        | C.04.4.1                      | [Dubov](/docs/pairing-systems/dubov/)               |
| Lim          | C.04.4.3                      | [Lim](/docs/pairing-systems/lim/)                   |
| Double-Swiss | C.04.5                        | [Double-Swiss](/docs/pairing-systems/double-swiss/) |
| Team Swiss   | C.04.6                        | [Team Swiss](/docs/pairing-systems/team/)           |
| Round-Robin  | C.05 (Berger tables, Annex 1) | [Round-Robin](/docs/pairing-systems/round-robin/)   |
| Keizer       | (not FIDE-regulated)          | [Keizer](/docs/pairing-systems/keizer/)             |

Each pairing system has its own page in the [Pairing Systems](/docs/pairing-systems/) section, explaining how the engine applies the regulation criteria, handles edge cases, and assigns colours.

### Drop-in replacement

The CLI can operate as a drop-in replacement for **bbpPairings** and **JaVaFo**. If you currently use either tool, you can switch to chesspairing without changing your workflow. See [Legacy Mode](/docs/cli/legacy/) for details.

## Practical workflow

A typical round proceeds in three steps:

1. **Prepare a TRF file.** The FIDE Tournament Report File (TRF16) is the standard exchange format for tournament data. Your tournament management software likely exports it. If you need to understand the format, see [TRF16](/docs/formats/trf16/).

2. **Run the pairer.** Feed the TRF file to the CLI:

   ```
   chesspairing pair --system dutch tournament.trf
   ```

   The engine reads the player list and all previous rounds, then outputs the pairings for the next round. Multiple [output formats](/docs/cli/output-formats/) are available (tabular, JSON, XML, board-style).

3. **Get pairings and standings.** The output lists each board with the White and Black player. For standings:
   ```
   chesspairing standings tournament.trf
   ```
   This computes scores and applies tiebreakers in order to produce ranked standings.

For a hands-on walkthrough, see the [CLI Quickstart](../cli-quickstart/).

## Common arbiter questions

The documentation is organized so you can find answers to the questions that come up during a tournament:

| Question                                                      | Where to look                                                                    |
| ------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| How are byes assigned when there is an odd number of players? | [Concepts: Byes](/docs/concepts/byes/)                                           |
| Why did player X get paired against player Y?                 | The page for your pairing system under [Pairing Systems](/docs/pairing-systems/) |
| What happens when a player receives a forfeit?                | [Concepts: Forfeits](/docs/concepts/forfeits/)                                   |
| How are colours allocated?                                    | [Concepts: Colors](/docs/concepts/colors/)                                       |
| What is a floater and why does it matter?                     | [Concepts: Floaters](/docs/concepts/floaters/)                                   |
| How are standings computed?                                   | [Scoring Systems](/docs/scoring/)                                                |
| Which tiebreakers are available?                              | See below, and the [Tiebreakers](/docs/tiebreakers/) section                     |
| How do I verify the engine produced correct pairings?         | [CLI: check](/docs/cli/check/)                                                   |
| How do I validate my TRF file?                                | [CLI: validate](/docs/cli/validate/)                                             |

## Available tiebreakers

Chesspairing provides 25 tiebreakers. The table below groups them by category, following the structure used in FIDE handbook section C.07. Each tiebreaker is identified by a short ID that you use in configuration.

### Buchholz family

Sum of opponents' scores, with variants that exclude extreme values.

| Tiebreaker                | ID                      |
| ------------------------- | ----------------------- |
| Buchholz (full)           | `buchholz`              |
| Buchholz Cut-1            | `buchholz-cut1`         |
| Buchholz Cut-2            | `buchholz-cut2`         |
| Buchholz Median           | `buchholz-median`       |
| Buchholz Median-2         | `buchholz-median2`      |
| Fore Buchholz             | `fore-buchholz`         |
| Average Opponent Buchholz | `avg-opponent-buchholz` |

For details, see [Buchholz](/docs/tiebreakers/buchholz/) and [Opponent Buchholz](/docs/tiebreakers/opponent-buchholz/).

### Head-to-head

Results between specific opponents, or results weighted by opponent score.

| Tiebreaker       | ID                 |
| ---------------- | ------------------ |
| Direct Encounter | `direct-encounter` |
| Sonneborn-Berger | `sonneborn-berger` |

See [Head-to-Head](/docs/tiebreakers/head-to-head/).

### Result-based

Derived directly from game outcomes.

| Tiebreaker                                              | ID                |
| ------------------------------------------------------- | ----------------- |
| Games Won (OTB wins only)                               | `wins`            |
| Rounds Won (OTB wins + forfeit wins + PAB)              | `win`             |
| Standard Points (1-half-0 regardless of scoring system) | `standard-points` |
| Progressive (cumulative) Score                          | `progressive`     |
| Koya System                                             | `koya`            |

See [Results](/docs/tiebreakers/results/).

### Performance-based

Derived from player ratings and the FIDE B.02 conversion table.

| Tiebreaker                    | ID                   |
| ----------------------------- | -------------------- |
| Average Rating of Opponents   | `aro`                |
| Tournament Performance Rating | `performance-rating` |
| Performance Points            | `performance-points` |
| Average Opponent TPR          | `avg-opponent-tpr`   |
| Average Opponent PTP          | `avg-opponent-ptp`   |

See [Performance](/docs/tiebreakers/performance/).

### Color and activity

Participation and colour distribution measures.

| Tiebreaker       | ID              |
| ---------------- | --------------- |
| Games with Black | `black-games`   |
| Black Wins       | `black-wins`    |
| Rounds Played    | `rounds-played` |
| Games Played     | `games-played`  |

See [Color & Activity](/docs/tiebreakers/color-activity/).

### Ordering

Deterministic final tiebreakers when all else is equal.

| Tiebreaker           | ID               |
| -------------------- | ---------------- |
| Pairing Number (TPN) | `pairing-number` |
| Player Rating        | `player-rating`  |

See [Ordering](/docs/tiebreakers/ordering/).

### FIDE defaults

When you do not specify tiebreakers explicitly, chesspairing applies FIDE-recommended defaults for each pairing system. For Swiss systems, these are Buchholz Cut-1, Buchholz, Sonneborn-Berger, and Direct Encounter. Round-robin defaults to Sonneborn-Berger, Direct Encounter, Wins, and Koya. You can override these in the [configuration](/docs/formats/configuration/).

## Scoring systems

Three scoring engines are available, each configurable with custom point values:

| System   | Default points          | Documentation                       |
| -------- | ----------------------- | ----------------------------------- |
| Standard | Win 1, Draw 0.5, Loss 0 | [Standard](/docs/scoring/standard/) |
| Football | Win 3, Draw 1, Loss 0   | [Football](/docs/scoring/football/) |
| Keizer   | Iterative convergence   | [Keizer](/docs/scoring/keizer/)     |

All three handle byes, forfeits, and absences with configurable point values. See the [Scoring Systems](/docs/scoring/) section for full details.

## Next steps

- [CLI Quickstart](../cli-quickstart/) -- pair a tournament from the command line in five minutes
- [Pairing Systems](/docs/pairing-systems/) -- detailed documentation for each pairing algorithm
- [Concepts](/docs/concepts/) -- Swiss system fundamentals, byes, colours, floaters, forfeits
- [CLI Reference](/docs/cli/) -- all available commands and output formats
