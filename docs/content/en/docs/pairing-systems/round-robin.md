---
title: "Round-Robin"
linkTitle: "Round-Robin"
weight: 8
description: "Every player meets every other player — FIDE Berger table scheduling with optional double round-robin."
---

The Round-Robin system schedules every player to face every other player exactly once (single round-robin) or twice with reversed colours (double round-robin). Pairings are generated deterministically from the FIDE Berger tables using a rotation algorithm. There is no matching optimization and no criteria evaluation -- the schedule is fully determined before the tournament begins by the player count and cycle configuration.

## When to Use

Round-Robin is appropriate when:

- The tournament is small enough that every player can face every other player (typically 6-16 players).
- Fairness requires that all players meet, eliminating the possibility that two strong players avoid each other through Swiss pairing.
- The event is a closed championship, qualification tournament, or league match.
- Double round-robin is used for events where each pair should play with both colours.

It is not suitable for large open tournaments where the number of rounds would be impractical (n-1 rounds for n players in a single cycle).

## Configuration

### CLI

```bash
chesspairing pair --roundrobin tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/roundrobin"

// With typed options
p := roundrobin.New(roundrobin.Options{
    Cycles:            chesspairing.IntPtr(2),
    ColorBalance:      chesspairing.BoolPtr(true),
    SwapLastTwoRounds: chesspairing.BoolPtr(true),
})

// From a generic map (JSON config)
p := roundrobin.NewFromMap(map[string]any{
    "cycles":            2,
    "colorBalance":      true,
    "swapLastTwoRounds": true,
})
```

### Options Reference

| Option              | Type   | Default | Description                                                                                                                                          |
| ------------------- | ------ | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cycles`            | `int`  | `1`     | Number of complete round-robins. `1` = single (each pair plays once). `2` = double (each pair plays twice with reversed colours).                    |
| `colorBalance`      | `bool` | `true`  | Whether to reverse all colours in even cycles (cycle 2, 4, ...) of a multi-cycle round-robin.                                                        |
| `swapLastTwoRounds` | `bool` | `true`  | Whether to swap the last two rounds of cycle 1 in a double round-robin. Only applies when `cycles` is `2` and there are at least 2 rounds per cycle. |

## How It Works

### 1. Table Setup

The engine determines the table size n from the number of active players. If the player count is odd, a dummy "BYE" player is added to make n even. The player paired against the dummy in any round receives a pairing-allocated bye.

Key values:

- **Rounds per cycle** = n - 1
- **Total rounds** = rounds per cycle multiplied by the number of cycles

### 2. Berger Table Rotation

The FIDE Berger table is generated using a fixed-point rotation algorithm:

1. Fix the last player (index n-1) at position n-1 for all rounds. For odd player counts, this fixed position is the bye dummy.
2. Rotate the remaining n-1 players through positions 0 to n-2. The rotation uses a stride of n/2 - 1 per round.
3. For round r within a cycle, the position of player j is: `positions[j] = ((j - r * stride) mod m + m) mod m`, where m = n - 1.

### 3. Pairing from Positions

Pairings are formed by matching positions symmetrically: position 0 with position n-1, position 1 with position n-2, and so on through position n/2 - 1 with position n/2.

If either position in a pair corresponds to the bye dummy (odd player count), the real player receives a bye instead of a game.

### 4. Colour Assignment

Colours follow the FIDE Berger table conventions:

- **Board 1** (the fixed player against the rotating player at position 0): In even rounds (0-based), the rotating player gets White. In odd rounds, the fixed player gets White.
- **Other boards**: The player with the lower position index (the "top-row" player) gets White.

### 5. Cycle Colour Reversal

In multi-cycle tournaments with `colorBalance` enabled, all colours are reversed in even cycles (0-based cycle index 1, 3, ...). This ensures that in a double round-robin, every pair plays once with each colour assignment.

### 6. Last-Two-Round Swap

In a double round-robin (`cycles` = 2) with `swapLastTwoRounds` enabled and at least 2 rounds per cycle, the engine swaps the second-to-last and last rounds of cycle 1. This prevents three or more consecutive games with the same colour at the boundary between cycle 1 and cycle 2.

Without this swap, a player who has White in the last round of cycle 1 would also have White in the first round of cycle 2 (since cycle 2 reverses colours from cycle 1, but the first round of cycle 2 maps to round 1 of the Berger table, not the last). The swap breaks this pattern.

## Comparison

| Aspect            | Round-Robin                              | Dutch Swiss                     | Keizer              |
| ----------------- | ---------------------------------------- | ------------------------------- | ------------------- |
| Schedule          | Fully predetermined                      | Dynamic per round               | Dynamic per round   |
| Repeat encounters | Every pair plays exactly once (or twice) | Never                           | Configurable        |
| Rounds required   | n - 1 per cycle                          | Typically log2(n) to 2\*log2(n) | Unlimited           |
| Player capacity   | Small (6-16 typical)                     | Large (any size)                | Medium (club-sized) |
| Colour allocation | Berger table convention                  | FIDE multi-step                 | Simple alternation  |
| Bye handling      | Dummy player rotation                    | Completability-based            | Lowest-ranked       |
| FIDE regulated    | Yes (C.05 Annex 1)                       | Yes (C.04.3)                    | No                  |

## Mathematical Foundations

### Berger Table Formula

For n players (including dummy if odd), m = n - 1 rotating players, stride s = n/2 - 1:

```text
position(j, r) = ((j - r * s) mod m + m) mod m    for j in [0, m-1]
position(m, r) = m                                  (fixed player)
```

Pairings for round r:

```text
pair(i, r) = (player[position(i, r)], player[position(n-1-i, r)])    for i in [0, n/2 - 1]
```

### Round Count

| Configuration     | Rounds   |
| ----------------- | -------- |
| Single RR, n even | n - 1    |
| Single RR, n odd  | n        |
| Double RR, n even | 2(n - 1) |
| Double RR, n odd  | 2n       |

The odd-player case adds 1 to the effective player count (the dummy), so rounds per cycle = n.

### Colour Balance Property

In a single round-robin with n players:

- When n is even: each player gets (n-2)/2 games as White and (n-2)/2 as Black, plus one game whose colour depends on their position (not perfectly balanced for all players).
- The fixed player alternates colours each round.

In a double round-robin with `colorBalance` enabled: each pair plays exactly once with each colour assignment, giving perfect pairwise colour balance.

### Last-Two-Round Swap Correctness

The swap is applied to cycle 1 only. Let R = rounds per cycle. In cycle 1, the logical round mapping becomes:

```text
actual round R-2 -> Berger round R-1
actual round R-1 -> Berger round R-2
```

This changes the colour of the last two rounds in cycle 1, breaking the consecutive-same-colour pattern that would otherwise occur at the cycle 1/cycle 2 boundary. The swap is only meaningful for double round-robin and is disabled for single round-robin or when rounds per cycle is less than 2.

## FIDE Reference

- **Regulation**: FIDE C.05 Annex 1 (Berger Tables)
- **Key rules**: Berger table rotation for scheduling, colour assignment by board position, colour reversal in double round-robin cycles
- **Pre-processing**: For tournaments using FIDE pairing number assignment, the [Varma tables](/docs/algorithms/varma-tables/) (FIDE C.05 Annex 2) can be used for federation-aware number assignment before round-robin pairing begins
