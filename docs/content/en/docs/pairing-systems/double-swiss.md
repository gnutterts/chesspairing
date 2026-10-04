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
| `topSeedColor`   | `string`     | `"auto"` | Colour of the top seed in round 1. Values: `"auto"`, `"white"`, `"black"`. When `"auto"`, odd boards give the higher-ranked player White and even boards give them Black. |
| `totalRounds`    | `int`        | nil      | Total number of rounds. Used to detect the last round for C8 criteria relaxation.                                                                                         |
| `forbiddenPairs` | `[][]string` | nil      | Pairs of participant IDs that must never be paired together.                                                                                                              |
| `pabPoints`      | `float64`    | `1.5`    | Points for a pairing-allocated bye; by default the value of one win and one draw (C.04.5 Art. 1.4).                                                                       |

## How It Works

The pairing algorithm proceeds in five stages:

### 1. Build Participant States

`lexswiss.BuildParticipantStates` extracts the pairing state for each active participant from `TournamentState`. Pairing scores use standard 1-0.5-0 per game regardless of the tournament's scoring system. Participants are sorted by score descending, then initial rank ascending, and assigned a Tournament Pairing Number (TPN) reflecting their current standing.

Forfeit games are excluded from opponent history, meaning two players who were previously paired in a forfeited game can be paired again.

### 2. Assign the Pairing-Allocated Bye

If the participant count is odd, `lexswiss.AssignPAB` selects the bye recipient per Art. 3.4. The bye goes to the eligible participant (one who has not already received a PAB) with the lowest score, breaking ties by highest TPN. The PAB awards 1.5 points: the points for a match won with one win and one draw.

### 3. Build Score Groups

Participants are grouped by score into score groups, sorted in descending score order. Within each group, participants are ordered by TPN ascending.

### 4. Lexicographic Bracket Pairing

Each score group is paired by Article 3.6 identifier order. The search chooses the sorted set of top members first, then assigns their bottom members in ascending TPN order. An identifier consists of the top-member TPNs followed by the corresponding bottom-member TPNs; the first legal complete identifier is selected. In a six-player first round this is `1-4, 2-5, 3-6`.

**Absolute criteria (always enforced):**

- C1: No two participants play each other more than once.
- Forbidden pairs are never paired.

**Quality criteria (C8 -- colour preferences):**

- C8 checks that two participants with two consecutive same-colour Game 1 assignments who both need the same colour next round are not paired together, since one would necessarily violate the 3-consecutive constraint.
- In the last round (when `TotalRounds` is set and `CurrentRound >= TotalRounds`), C8 is relaxed entirely.

If a score group has an odd number of participants, the upfloater is selected under C.04.5 Art. 3.5.3--3.5.5: candidates are considered by score descending and then TPN ascending; the first is selected only when the rest of its bracket and the target bracket can still be paired. The upfloater must have a compatible opponent in the target group.

### 5. Colour Allocation

`AllocateColor` implements the five-step colour allocation procedure (Art. 4):

1. **Hard constraint**: No participant plays Game 1 as the same colour three times in a row. If a participant has had White in Game 1 for the last two rounds, they must get Black.
2. **Equalise**: The participant with more White Game 1 assignments gets Black.
3. **Alternate**: The participant who had White in Game 1 last round gets Black.
4. **Round 1 board alternation**: Odd-numbered boards give the higher-ranked player White; even-numbered boards give them Black. The `topSeedColor` option can invert this pattern.
5. **Rank tiebreak**: The higher-ranked player (lower TPN) gets White.

After colour allocation, boards are sorted by maximum score in the pair (descending), then minimum TPN in the pair (ascending).

## Comparison

| Aspect                | Double-Swiss       | Dutch/Burstein     | Dubov/Lim                |
| --------------------- | ------------------ | ------------------ | ------------------------ |
| Games per round       | 2 (mini-match)     | 1                  | 1                        |
| Matching algorithm    | Lexicographic DFS  | Global Blossom     | Transposition / Exchange |
| Criteria count        | 2 (C1 + C8)        | 21 (C1-C21)        | 10 (C1-C10)              |
| Colour constraint     | 3-consecutive ban  | 3-consecutive ban  | 3-consecutive ban        |
| PAB value             | 1.5 points         | 1 point            | 1 point                  |
| Shared infrastructure | `pairing/lexswiss` | `pairing/swisslib` | `pairing/swisslib`       |

The lexicographic approach is simpler than Blossom matching: it selects the first valid Article 3.6 identifier rather than optimizing a weighted objective across all brackets. This makes the algorithm deterministic by construction, at the cost of not considering cross-bracket optimization.

## Mathematical Foundations

### Lexicographic Enumeration

For each pair the lower TPN is its top member. The identifier is the ascending top-member TPN sequence followed by the bottom member corresponding to each top member. The search enumerates top-member sets in ascending lexicographic order, then bottom assignments in ascending lexicographic order. The first valid complete identifier is selected; infeasible prefixes are pruned.

### Complexity

For a score group of size k, the worst-case number of candidate pairings is `(k-1)!! = (k-1) * (k-3) * ... * 1`. In practice, the C1 constraint (no repeat opponents) and forbidden pairs prune the search tree substantially. The DFS terminates at the first valid pairing, so average-case performance is much better than the worst case.

### Colour Allocation Priority

The five-step priority chain forms a strict total order over colour assignments. Step 1 (hard constraint) has veto power and can override any preference-based step. Steps 2-5 are applied in sequence as tiebreakers.

## FIDE Reference

- **Regulation**: FIDE C.04.5 (Double-Swiss System)
- **Adopted**: October 2025
- **Effective**: February 2026
- **Key articles**: Art. 3.4 (PAB assignment, 1.5 points), Art. 3.5 (upfloater selection), Art. 3.6 (lexicographic bracket pairing), Art. 4 (colour allocation)
