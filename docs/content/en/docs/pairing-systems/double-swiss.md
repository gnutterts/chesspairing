---
title: "Double-Swiss System"
linkTitle: "Double-Swiss"
weight: 5
description: "FIDE C.04.5 — lexicographic pairing for large tournaments where players are paired twice."
---

The Double-Swiss system treats each round as a two-game match. Players are paired using lexicographic bracket pairing, and colours alternate within each match so both players get one game as White and one as Black. Approved by FIDE in October 2025 and effective from February 2026, it is designed for large open tournaments where a Swiss format benefits from the increased statistical reliability of mini-matches.

## When to Use

Double-Swiss is appropriate when:

- The tournament has enough rounds relative to the number of participants for meaningful Swiss pairing, but you want to reduce the variance that comes from single-game results.
- You want each round to produce a decisive match result more often, since a two-game match can end in a win even if one game is drawn.
- The event can accommodate the longer playing time required by two-game matches per round.

It is less suitable when round time is constrained to a single game, or when participants expect the traditional one-game-per-round Swiss format.

## Configuration

### CLI

```bash
chesspairing pair --double-swiss tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/doubleswiss"

// With typed options
p := doubleswiss.New(doubleswiss.Options{
    TopSeedColor: chesspairing.StringPtr("auto"),
    TotalRounds:  chesspairing.IntPtr(9),
    ForbiddenPairs: [][]string{{"player-1", "player-2"}},
})

// From a generic map (JSON config)
p := doubleswiss.NewFromMap(map[string]any{
    "topSeedColor": "auto",
    "totalRounds":  9,
})
```

### Options Reference

| Option           | Type         | Default  | Description                                                                                                                                                               |
| ---------------- | ------------ | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `topSeedColor`   | `string`     | `"auto"` | Initial colour drawn before round 1. Values: `"auto"`, `"white"`, `"black"`; it is given to a new higher-ranked player with an odd TPN. |
| `totalRounds`    | `int`        | nil      | Total number of rounds. Used to detect the last round for C8 criteria relaxation.                                                                                         |
| `forbiddenPairs` | `[][]string` | nil      | Pairs of participant IDs that must never be paired together.                                                                                                              |
| `pabPoints`      | `float64`    | `1.5`    | Points for a pairing-allocated bye; by default the value of one win and one draw (C.04.5 Art. 1.4).                                                                       |

## How It Works

The pairing algorithm proceeds in five stages:

### 1. Build Participant States

`lexswiss.BuildParticipantStates` extracts the pairing state for each active participant from `TournamentState`. Pairing scores use standard 1-0.5-0 per game regardless of the tournament's scoring system. The Tournament Pairing Number (TPN) is fixed by the initial order (Art. 1.1); for pairing purposes participants are sorted by score descending and then TPN ascending (Art. 1.2). It is never renumbered for a round.

Callers must record a match as two `GameData` entries in the same round: game 1 in the pairing colours and game 2 with colours reversed. A single entry remains supported as a one-game match. Colour history has one entry per match: game 1's scheduled colour, only when at least one game was played. A match is forfeited only when one player forfeits every recorded game; a single forfeited game is otherwise a regularly played match for C1, C2, and matches played. A game is treated as a forfeit when either its result is a forfeit or its `IsForfeit` flag is set.

### 2. Assign the Pairing-Allocated Bye

If the participant count is odd, `lexswiss.AssignPAB` selects the bye recipient per Art. 3.4. Among C2-eligible participants (no earlier PAB, forfeit win, or deprecated full-point bye), it first retains only those whose removal leaves a complete legal C1 pairing. It then selects the lowest score, the highest number of matches actually played, and finally the largest fixed TPN, in that order. A forfeited match is not played; a Double-Swiss match is otherwise played even if one game was forfeited. The PAB awards 1.5 points: the points for a match won with one win and one draw.

### 3. Build Score Groups

Participants are grouped by score into score groups, sorted in descending score order. Within each group, participants are ordered by TPN ascending.

### 4. Lexicographic Bracket Pairing

Each score group is paired by Article 3.6 identifier order. The search chooses the sorted set of top members first, then assigns their bottom members in ascending TPN order. An identifier consists of the top-member TPNs followed by the corresponding bottom-member TPNs; the first legal complete identifier is selected. In a six-player first round this is `1-4, 2-5, 3-6`.

**Absolute criteria (always enforced):**

- C1: No two participants play each other more than once.
- Forbidden pairs are never paired.

**Quality criterion (C8):**

- Except in the last round, C8 minimises the number of upfloaters whose opponent was a floater in the previous round.
- Pairings are considered in identifier order; the first pairing with the minimum C8 count is selected. C8 does not apply in the last round (when `TotalRounds` is set and `CurrentRound >= TotalRounds`).
- Colour is not a pairing criterion. It is allocated only after the complete round-pairing has been made.

Brackets are built from the highest unpaired scoregroup downwards under C.04.5 Art. 3.3.2. All residents stay in their bracket. Upfloater sets are tried with the smallest working number first (C4), then the best score profile (C5), then C6, C7, and finally Article 3.5.4 TPN order. A set must produce a legal Article 3.6 bracket and leave a C1/C2 completion for the remaining players. Failed brackets are not pooled; an impossible round is reported under Art. 3.3.3.

### 5. Colour Allocation

`AllocateColor` applies Article 4.3 to each pair, after pairing:

1. When both players have not yet played a match, the higher-ranked player receives the initial colour for an odd TPN and the opposite colour for an even TPN.
2. Otherwise, White goes to the player with fewer previous Whites.
3. If still tied, colours alternate from the most recent match in which the players had opposite colours; unplayed matches are removed before comparison.
4. If still tied, the higher-ranked player's colour alternates from that player's last played round.
5. Finally, the opponent's colour alternates from that opponent's last played round.

After colour allocation, boards are sorted by maximum score in the pair (descending), then minimum TPN in the pair (ascending).

## Comparison

| Aspect                | Double-Swiss       | Dutch/Burstein     | Dubov/Lim                |
| --------------------- | ------------------ | ------------------ | ------------------------ |
| Games per round       | 2 (mini-match)     | 1                  | 1                        |
| Matching algorithm    | Identifier order + weighted matching for C8 | Global Blossom     | Transposition / Exchange |
| Criteria count        | C1, C2 and C8      | 21 (C1-C21)        | 10 (C1-C10)              |
| Colour allocation     | After pairing, Art. 4 | 3-consecutive ban | 3-consecutive ban        |
| PAB value             | 1.5 points         | 1 point            | 1 point                  |
| Shared infrastructure | `pairing/lexswiss` | `pairing/swisslib` | `pairing/swisslib`       |

The lexicographic approach is simpler than Blossom matching: it selects the first Article 3.6 identifier among the pairings with the minimum C8 count, rather than optimising a weighted objective across all brackets. This makes the algorithm deterministic by construction, at the cost of not considering cross-bracket optimisation.

## Mathematical Foundations

### Lexicographic Enumeration

For each pair the lower TPN is its top member. The identifier is the ascending top-member TPN sequence followed by the bottom member corresponding to each top member. The search enumerates top-member sets in ascending lexicographic order, then bottom assignments in ascending lexicographic order. Among legal complete identifiers, the first with the minimum C8 count is selected; infeasible prefixes are pruned.

### Complexity

A bracket of size k has up to `(k-1)!! = (k-1) * (k-3) * ... * 1` pairings, so they are never enumerated. Without C8 cost the search walks the identifiers and prunes every branch that cannot be completed under C1 and the forbidden pairs. With C8, one weighted matching first gives the minimal C8 count; the pairing is then built in identifier order, and each step is checked with another weighted matching on the remaining players, so the work is polynomial. Every check is charged to a search limit; when it runs out the round fails with the partial pairing.

### Colour Allocation Priority

Article 4.3 is a descending-priority sequence: initial colour for two new players, fewer Whites, the most recent opposite colours, the higher-ranked player's last colour, then the opponent's last colour. It allocates colours after the pairing and does not veto a pairing.

## FIDE Reference

- **Regulation**: FIDE C.04.5 (Double-Swiss System)
- **Adopted**: October 2025
- **Effective**: February 2026
- **Key articles**: Art. 3.4.1-3.4.4 (PAB assignment, 1.5 points), Art. 3.5 (upfloater selection), Art. 3.6 (lexicographic bracket pairing), Art. 4 (colour allocation)
