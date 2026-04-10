---
title: "Pairing Systems"
linkTitle: "Pairing Systems"
weight: 30
description: "Eight pairing algorithms — from FIDE Swiss variants to Round-Robin and Keizer."
---

Chesspairing implements eight pairing engines. Each one satisfies the `Pairer` interface, which means any engine can be combined with any scoring system and any set of tiebreakers. The choice of pairing system depends on the tournament format, the regulations in force, and the organiser's goals.

## Swiss systems

Five engines implement FIDE Swiss pairing regulations. They all share the same high-level idea -- group players by score and pair across groups -- but differ in matching strategy, floater handling, colour allocation, and optimization criteria.

| System                        | FIDE regulation | Matching strategy                            | Best suited for                                             |
| ----------------------------- | --------------- | -------------------------------------------- | ----------------------------------------------------------- |
| [Dutch](dutch/)               | C.04.3          | Global Blossom (21 criteria)                 | Standard rated tournaments, any size                        |
| [Burstein](burstein/)         | C.04.4.2        | Global Blossom + opposition index re-ranking | Events with a seeding phase followed by competitive pairing |
| [Dubov](dubov/)               | C.04.4.1        | Transposition-based, ARO-ordered             | Events prioritising opponent strength balance               |
| [Lim](lim/)                   | C.04.4.3        | Exchange-based, median-first                 | Events wanting explicit floater control                     |
| [Double-Swiss](double-swiss/) | C.04.5          | Lexicographic bracket pairing                | Large events needing faster pairing computation             |

All five Swiss engines handle bye assignment, colour balancing, rematch avoidance, and forfeit exclusion. They differ primarily in how they resolve conflicts when a perfect pairing is not possible.

### Choosing between Swiss variants

**Dutch** is the default. It encodes 21 criteria into a single Blossom matching problem, guaranteeing a globally optimal solution within FIDE constraints. Unless regulations or tournament characteristics call for something else, Dutch is the right choice.

**Burstein** extends Dutch with an opposition-index mechanism. During early "seeding" rounds, pairings follow standard Dutch rules. After the seeding phase, players are re-ranked by Buchholz and Sonneborn-Berger indices, creating more balanced opposition in later rounds. This suits events where early rounds sort the field and later rounds should match similarly-performing players.

**Dubov** processes score groups in ascending ARO (Average Rating of Opponents) order rather than descending pairing number. This spreads strong opposition more evenly across the draw. It uses transposition-based matching within score groups, which is simpler than Blossom but handles most practical cases well.

**Lim** uses a median-first processing order (middle score groups first, then outward) and explicit floater classification (types A through D). The exchange-based matching within score groups gives the arbiter a more predictable pairing process at the cost of some optimality compared to global Blossom matching.

**Double-Swiss** uses lexicographic bracket pairing from the shared `lexswiss` library. It pairs faster than Blossom-based systems and includes an explicit ban on three consecutive games with the same colour. It targets large open events where computational speed matters.

## Team Swiss

[Team Swiss](team/) (FIDE C.04.6) pairs teams rather than individuals. It shares the lexicographic bracket pairing infrastructure with Double-Swiss but adds team-specific colour preference (types A, B, or None based on board-1 history) and a 9-step colour allocation procedure. The primary score can be match points or game points, configurable via options.

## Non-Swiss systems

| System                      | Matching strategy                 | Best suited for                            |
| --------------------------- | --------------------------------- | ------------------------------------------ |
| [Round-Robin](round-robin/) | FIDE Berger tables (C.05 Annex 1) | Fixed-field events, league play            |
| [Keizer](keizer/)           | Top-down by Keizer score          | Club events wanting competitive incentives |

**Round-Robin** generates pairings from FIDE Berger rotation tables. Every player meets every other player exactly once per cycle, with configurable multi-cycle support and optional last-two-round swap for double round-robin events. There is no score-based matching -- the schedule is fully determined before the tournament begins.

**Keizer** ranks players by Keizer score (computed by the Keizer scoring engine) and pairs top-down: first plays second, third plays fourth, and so on. Repeat avoidance pushes opponents apart when they have already met. Keizer pairing only makes sense with Keizer scoring, since the rankings that drive the pairing depend on the iterative Keizer score computation.

## Interface

All eight engines implement the same interface:

```go
type Pairer interface {
    Pair(ctx context.Context, state TournamentState) (PairingResult, error)
}
```

The `TournamentState` contains all player data, round history, game results, and configuration. The returned `PairingResult` contains the list of game pairings (board assignments with white/black) and any bye entries. The pairing engine never modifies the input state.

Every engine also provides a `NewFromMap(map[string]any)` constructor for instantiation from generic configuration (JSON, TRF options, CLI flags). Engine-specific options are documented on each engine's page.
