---
title: "Team Swiss System"
linkTitle: "Team Swiss"
weight: 6
description: "FIDE C.04.6 — Swiss pairing for team competitions with board-level color allocation."
---

The Team Swiss system pairs teams rather than individual players, using the same lexicographic bracket pairing infrastructure as the Double-Swiss system. Each `PlayerEntry` in `TournamentState` represents a team. The system features a nine-step colour allocation procedure and two configurable colour preference types that control how aggressively the engine tries to satisfy colour balance at the team level. Approved by FIDE in October 2025 and effective from February 2026.

## When to Use

Team Swiss is appropriate when:

- The competition is between teams (e.g. club championships, Olympiad-style events).
- You need a Swiss format that handles team-level pairing while respecting board-level colour assignments.
- The event uses match points (team wins a match) or game points (sum of individual board results) as the primary ranking criterion.

It is not suitable for individual tournaments. For individual two-game-match Swiss tournaments, use the [Double-Swiss](../double-swiss/) system.

## Configuration

### CLI

```bash
chesspairing pair --team tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/team"

// With typed options
p := team.New(team.Options{
    TopSeedColor:        chesspairing.StringPtr("white"),
    TotalRounds:         chesspairing.IntPtr(11),
    ColorPreferenceType: chesspairing.StringPtr("A"),
    PrimaryScore:        chesspairing.StringPtr("match"),
    ForbiddenPairs:      [][]string{{"team-1", "team-2"}},
})

// From a generic map (JSON config)
p := team.NewFromMap(map[string]any{
    "topSeedColor":        "white",
    "totalRounds":         11,
    "colorPreferenceType": "B",
    "primaryScore":        "game",
})
```

### Options Reference

| Option                | Type         | Default   | Description                                                                                                                                                                  |
| --------------------- | ------------ | --------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `topSeedColor`        | `string`     | `"white"` | Initial colour from drawing of lots before round 1 (Art. 4.1). Values: `"white"`, `"black"`. Note: unlike other systems, this defaults to `"white"`, not `"auto"`.           |
| `totalRounds`         | `int`        | nil       | Total number of rounds. Used to determine the last two rounds (C7/C10 relaxation) and last round (Type B mild preference calculation).                                       |
| `colorPreferenceType` | `string`     | `"A"`     | Colour preference rules (Art. 1.7). Values: `"A"` (simple), `"B"` (strong + mild), `"none"` (disabled).                                                                      |
| `primaryScore`        | `string`     | `"match"` | Score used for pairing (Art. 1.2). Values: `"match"` (match points), `"game"` (game points). The other score becomes the secondary score for colour allocation (Art. 4.2.2). |
| `forbiddenPairs`      | `[][]string` | nil       | Pairs of team IDs that must never be paired together.                                                                                                                        |

## How It Works

The pairing algorithm proceeds in five stages, sharing most of its infrastructure with the Double-Swiss system through the `pairing/lexswiss` package.

### 1. Build Participant States

`lexswiss.BuildParticipantStates` constructs the pairing state for each active team. Each team is treated as a single participant. Pairing scores use standard 1-0.5-0 regardless of the tournament's scoring system. Teams are sorted by score descending, then initial rank ascending, and assigned a TPN.

### 2. Assign the Pairing-Allocated Bye

If the team count is odd, `assignTeamPAB` selects the bye recipient per Art. 3.4. Team Swiss adds an extra tiebreaker compared to the base `lexswiss.AssignPAB`: among teams with the same lowest score, the team with the most matches played (longest colour history) receives the bye first, then largest TPN.

### 3. Build Score Groups and Criteria

Teams are grouped by score into descending-order score groups. The criteria function is built based on the colour preference type and round context:

**C8 (colour preferences):** Two teams with the same colour preference direction (both wanting White, or both wanting Black) cannot both be satisfied in a pairing. Such pairings are rejected during lexicographic enumeration.

**C9 (Type B only):** Two teams with the same strong colour preference violate C9. Since strong preferences imply the same direction, this is already captured by C8 for same-colour cases.

**C7 and C10** are relaxed in the last two rounds (when `TotalRounds` is set and `CurrentRound >= TotalRounds - 1`).

### 4. Lexicographic Bracket Pairing

Identical to Double-Swiss: depth-first search enumerating pairings in lexicographic TPN order, with backtracking. Odd-sized score groups float the lowest-ranked team up to the bracket above.

### 5. Colour Allocation (9-Step)

`AllocateColor` implements the nine-step colour allocation procedure of Art. 4:

1. **Determine the first-team** (Art. 4.2): Higher primary score, then higher secondary score (if available), then smaller TPN.
2. **No history** (Art. 4.3.1): If neither team has played a match, assign by TPN parity. Odd TPN gets the initial colour; even TPN gets the opposite.
3. **One preference** (Art. 4.3.2): If only one team has a colour preference, grant it.
4. **Opposite preferences** (Art. 4.3.3): If both teams have opposite preferences, grant both.
5. **Strong vs non-strong** (Art. 4.3.4, Type B only): If only one team has a strong preference, grant it.
6. **Colour difference** (Art. 4.3.5): The team with the lower colour difference (fewer whites minus blacks) gets White.
7. **Alternation** (Art. 4.3.6): Find the most recent round where one team had White and the other had Black, then alternate.
8. **First-team preference** (Art. 4.3.7): Grant the first-team's colour preference.
9. **Last colour alternation** (Art. 4.3.8-9): Alternate from the first-team's last played colour; if still tied, alternate from the other team's last colour.

Colour is determined by the first board assignment (Art. 1.6.1): whichever team gets White on board 1 is considered "White" for the match.

## Colour Preference Types

### Type A (Simple)

A team has a preference for White if its colour difference is below -1, or if CD is 0 or -1 and the last two played matches were both Black. The symmetric rule applies for Black. Otherwise, no preference. This is a binary system: the team either has a preference or it does not.

### Type B (Strong + Mild)

Type B adds a second tier. Strong preferences use the same conditions as Type A. Mild preferences apply when:

- CD is -1: mild preference for White.
- CD is 0 and it is not the last round and the last played match was Black: mild preference for White.
- Symmetric rules apply for Black.

In the last round, mild preferences disappear (CD = 0 produces no preference). The strong/mild distinction affects step 5 of colour allocation: a strong preference takes priority over a non-strong one.

## Comparison

| Aspect                  | Team Swiss             | Double-Swiss       | Dutch                 |
| ----------------------- | ---------------------- | ------------------ | --------------------- |
| Participants            | Teams                  | Individual players | Individual players    |
| Matching algorithm      | Lexicographic DFS      | Lexicographic DFS  | Global Blossom        |
| Colour allocation steps | 9                      | 5                  | FIDE C.04.3 rules     |
| Colour preference types | 3 (A, B, None)         | N/A                | N/A                   |
| Primary score options   | Match or game points   | Game points        | Standard points       |
| Criteria relaxation     | Last 2 rounds (C7/C10) | Last round (C8)    | No special relaxation |
| Default `topSeedColor`  | `"white"`              | `"auto"`           | `"auto"`              |
| Shared infrastructure   | `pairing/lexswiss`     | `pairing/lexswiss` | `pairing/swisslib`    |

## Mathematical Foundations

### Colour Difference

The colour difference (CD) for a team is defined as:

```text
CD = (number of White assignments) - (number of Black assignments)
```

Rounds where the team had no game (bye, absence) are excluded. CD drives the preference computation: negative CD pushes toward White, positive CD pushes toward Black.

### First-Team Ordering

The first-team determination in Art. 4.2 creates a strict total order for each pair:

```text
first(a, b) = a  if  score(a) > score(b)
            = a  if  score(a) == score(b) and secondary(a) > secondary(b)
            = a  if  scores equal and TPN(a) < TPN(b)
```

This ordering ensures deterministic colour allocation when all other tiebreakers are exhausted.

### Lexicographic Pairing

The lexicographic enumeration is identical to Double-Swiss. See the [Double-Swiss](../double-swiss/#lexicographic-enumeration) mathematical foundations for the formal description.

## FIDE Reference

- **Regulation**: FIDE C.04.6 (Team Swiss System)
- **Adopted**: October 2025
- **Effective**: February 2026
- **Key articles**: Art. 1.2 (primary/secondary score), Art. 1.6.1 (colour by first board), Art. 1.7 (colour preference types), Art. 3.4 (PAB assignment), Art. 3.5 (upfloater selection), Art. 3.6 (lexicographic bracket pairing), Art. 4 (9-step colour allocation)
