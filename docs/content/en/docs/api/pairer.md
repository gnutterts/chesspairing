---
title: "Pairer Interface"
linkTitle: "Pairer"
weight: 3
description: "The Pairer interface and how to use each of the eight pairing implementations."
---

The `Pairer` interface generates pairings for a tournament round. Eight implementations cover all major chess pairing systems.

## Interface definition

```go
type Pairer interface {
    Pair(ctx context.Context, state *TournamentState) (*PairingResult, error)
}
```

`Pair` accepts a `context.Context` and a pointer to `TournamentState`. The state is treated as read-only and is never modified by the engine. The returned `PairingResult` contains board assignments (`Pairings`) and any byes (`Byes`).

All engines accept `context.Context` for forward compatibility. Since all computation is CPU-bound and in-memory, the context is not currently checked for cancellation.

## Implementations

Every pairing engine lives in its own package under `pairing/` and provides two constructors:

- `New(opts Options) *Pairer` -- typed options struct.
- `NewFromMap(m map[string]any) *Pairer` -- generic map (from JSON config or TRF).

Both apply defaults for any unset (nil) fields. See [Options Pattern](../options/) for the nil-means-default convention.

### Creating a pairer

```go
import "github.com/gnutterts/chesspairing/pairing/dutch"

// From a generic options map (e.g. parsed from JSON config):
p := dutch.NewFromMap(nil) // all defaults

// From a typed Options struct (string options require a pointer;
// the package does not provide a StringPtr helper, so use a local variable):
accel := "baku"
color := "white"
p := dutch.New(dutch.Options{
    Acceleration: &accel,
    TopSeedColor: &color,
})
```

All eight engines follow the same pattern. Replace `dutch` with the target package name.

### Compile-time interface check

Every engine package includes a compile-time assertion:

```go
var _ chesspairing.Pairer = (*Pairer)(nil)
```

This ensures the implementation satisfies the `Pairer` interface at compile time.

## Per-engine options

| Engine       | Package               | Key Options                                                                                                                                                                                     |
| ------------ | --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Dutch        | `pairing/dutch`       | `Acceleration` (`*string`: `"none"`, `"baku"`), `TopSeedColor` (`*string`: `"auto"`, `"white"`, `"black"`), `ForbiddenPairs` (`[][]string`)                                                     |
| Burstein     | `pairing/burstein`    | `Acceleration` (`*string`), `TopSeedColor` (`*string`), `ForbiddenPairs` (`[][]string`), `TotalRounds` (`*int`)                                                                                 |
| Dubov        | `pairing/dubov`       | `TopSeedColor` (`*string`), `ForbiddenPairs` (`[][]string`), `TotalRounds` (`*int`)                                                                                                             |
| Lim          | `pairing/lim`         | `TopSeedColor` (`*string`), `ForbiddenPairs` (`[][]string`), `MaxiTournament` (`*bool`)                                                                                                         |
| Double-Swiss | `pairing/doubleswiss` | `TopSeedColor` (`*string`), `ForbiddenPairs` (`[][]string`), `TotalRounds` (`*int`)                                                                                                             |
| Team Swiss   | `pairing/team`        | `TopSeedColor` (`*string`), `ForbiddenPairs` (`[][]string`), `TotalRounds` (`*int`), `ColorPreferenceType` (`*string`: `"A"`, `"B"`, `"none"`), `PrimaryScore` (`*string`: `"match"`, `"game"`) |
| Keizer       | `pairing/keizer`      | `AllowRepeatPairings` (`*bool`, default `true`), `MinRoundsBetweenRepeats` (`*int`, default `3`), `ScoringOptions` (`*keizer.Options`)                                                          |
| Round-Robin  | `pairing/roundrobin`  | `Cycles` (`*int`, default `1`), `ColorBalance` (`*bool`, default `true`), `SwapLastTwoRounds` (`*bool`, default `true`)                                                                         |

### Common options

Most Swiss engines share these options:

- **TopSeedColor** -- Forces the top seed's colour in round 1. Values: `"auto"` (default, engine decides), `"white"`, `"black"`.
- **ForbiddenPairs** -- A list of player ID pairs `[id1, id2]` that must not be paired against each other. The engine treats these as absolute constraints.

### Engine-specific notes

**Dutch (FIDE C.04.3):** The most widely used Swiss system. Uses global Blossom matching with 21 quality criteria. `Acceleration` enables Baku acceleration (FIDE C.04.7), which assigns virtual points in early rounds to create more varied pairings.

**Burstein (FIDE C.04.4.2):** Uses seeding rounds (delegated to Dutch matching) followed by opposition-index-based matching. The number of seeding rounds is `min(floor(totalRounds/2), 4)`. Set `TotalRounds` to control this; if nil, it is derived from the state.

**Dubov (FIDE C.04.4.1):** An ARO-equalization Swiss variant. Splits score groups by colour preference, sorts by ascending ARO, and uses transposition-based matching with 10 criteria.

**Lim (FIDE C.04.4.3):** Processes score groups in median-first order and uses exchange-based matching. Has four floater types (A-D). `MaxiTournament` enables the 100-point rating constraint for exchanges and floater selection.

**Double-Swiss (FIDE C.04.5):** Each round is a 2-game match. Uses lexicographic bracket pairing. `TotalRounds` is used to determine the last round for criteria relaxation.

**Team Swiss (FIDE C.04.6):** Pairs teams (each `PlayerEntry` represents a team). `ColorPreferenceType` selects Type A (simple) or Type B (strong + mild) colour preferences. `PrimaryScore` chooses between match points and game points for pairing ranking.

**Keizer:** Top-down pairing by Keizer score. `AllowRepeatPairings` controls re-pairing, with `MinRoundsBetweenRepeats` setting the gap. `ScoringOptions` configures the internal Keizer scorer used for ranking; when nil, scoring defaults apply. Color allocation uses the swisslib 6-step cascade (same as Dutch/Burstein).

**Round-Robin (FIDE C.05 Annex 1):** Uses FIDE Berger tables. `Cycles` sets the number of complete round-robins (2 = double round-robin with reversed colours). `SwapLastTwoRounds` follows the FIDE recommendation to swap the last two rounds of cycle 1 in a double round-robin to avoid three consecutive same-colour games at the cycle boundary.

## Usage example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/pairing/dutch"
)

func main() {
    state := &chesspairing.TournamentState{
        Players: []chesspairing.PlayerEntry{
            {ID: "1", DisplayName: "Alice",   Rating: 2400},
            {ID: "2", DisplayName: "Bob",     Rating: 2350},
            {ID: "3", DisplayName: "Charlie", Rating: 2300},
            {ID: "4", DisplayName: "Diana",   Rating: 2250},
            {ID: "5", DisplayName: "Eve",     Rating: 2200},
        },
        CurrentRound: 1,
        PairingConfig: chesspairing.PairingConfig{
            System: chesspairing.PairingDutch,
        },
    }

    pairer := dutch.NewFromMap(nil)

    result, err := pairer.Pair(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }

    for _, g := range result.Pairings {
        fmt.Printf("Board %d: %s (W) vs %s (B)\n", g.Board, g.WhiteID, g.BlackID)
    }
    for _, b := range result.Byes {
        fmt.Printf("Bye: %s (%s)\n", b.PlayerID, b.Type)
    }
}
```

With an odd number of players, exactly one player receives a pairing-allocated bye (PAB). The bye assignee is determined by the engine's bye selection algorithm.

## Error handling

Engines return an error for invalid input states. Common error conditions:

- No active players in the tournament.
- `TournamentState.Validate()` fails (empty IDs, duplicates, round count mismatch).
- Insufficient players to form any pairing (e.g. one active player with no bye possible).

Pairing engines never panic. All exceptional conditions are reported through the returned `error` value.

## Data flow

```text
Caller builds TournamentState
  -> Pairer.Pair(ctx, state) returns *PairingResult
     -> PairingResult.Pairings: board assignments for the round
     -> PairingResult.Byes: bye assignments (typically 0 or 1)
     -> PairingResult.Notes: engine diagnostic messages
```

The caller is responsible for recording the pairing result into `RoundData` before calling `Pair` again for the next round.
