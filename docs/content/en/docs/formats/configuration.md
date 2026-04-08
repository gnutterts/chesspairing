---
title: "Configuration Format"
linkTitle: "Configuration"
weight: 4
description: "The configuration map format used by engine factories and the generate subcommand."
---

## Engine options map

All pairing and scoring engines accept a `map[string]any` for configuration. This map can come from:

- **TRF system-specific fields** -- parsed by `ToTournamentState()` from XX codes and TRF-2026 headers
- **JSON configuration** -- deserialized from a JSON object
- **CLI flags** -- mapped to option keys by the `standings` subcommand

Keys are strings matching the Go struct field names (camelCase). Values are parsed with type coercion via the helper functions `GetFloat64()`, `GetInt()`, `GetBool()`, and `GetString()` from `options_helpers.go`. `GetFloat64` accepts `float64`, `int`, and `int64` values; `GetInt` accepts `int`, `int64`, and `float64`; `GetBool` and `GetString` accept their native types only.

### Pairing option keys

Options common to multiple Swiss pairers:

| Key              | Type       | Description                                          | Used by         |
| ---------------- | ---------- | ---------------------------------------------------- | --------------- |
| `totalRounds`    | int        | Total planned rounds                                 | All Swiss       |
| `topSeedColor`   | string     | Initial color for top seed                           | All Swiss       |
| `forbiddenPairs` | `[][2]int` | Pairs of start numbers that must not play each other | All Swiss       |
| `acceleration`   | string     | Acceleration type (`"baku"`)                         | Dutch, Burstein |

Round-Robin options:

| Key                 | Type | Description                                   |
| ------------------- | ---- | --------------------------------------------- |
| `cycles`            | int  | Number of cycles (`1` = single, `2` = double) |
| `colorBalance`      | bool | Enable color balancing                        |
| `swapLastTwoRounds` | bool | Swap last two rounds for double round-robin   |

Lim options:

| Key              | Type | Description                 |
| ---------------- | ---- | --------------------------- |
| `maxiTournament` | bool | Enable maxi-tournament mode |

Team Swiss options:

| Key                   | Type   | Description                                           |
| --------------------- | ------ | ----------------------------------------------------- |
| `colorPreferenceType` | string | Color preference algorithm: `"A"`, `"B"`, or `"none"` |
| `primaryScore`        | string | Primary scoring metric: `"match"` or `"game"`         |

Keizer options:

| Key                       | Type | Description                      |
| ------------------------- | ---- | -------------------------------- |
| `allowRepeatPairings`     | bool | Allow repeat pairings            |
| `minRoundsBetweenRepeats` | int  | Minimum rounds between rematches |

### Scoring option keys

Standard and Football scoring options:

| Key                | Type    | Default (Standard) | Default (Football) | Description               |
| ------------------ | ------- | ------------------ | ------------------ | ------------------------- |
| `pointWin`         | float64 | `1.0`              | `3.0`              | Points for a win          |
| `pointDraw`        | float64 | `0.5`              | `1.0`              | Points for a draw         |
| `pointLoss`        | float64 | `0.0`              | `0.0`              | Points for a loss         |
| `pointBye`         | float64 | `1.0`              | `3.0`              | Points for PAB            |
| `pointForfeitWin`  | float64 | `1.0`              | `3.0`              | Points for a forfeit win  |
| `pointForfeitLoss` | float64 | `0.0`              | `0.0`              | Points for a forfeit loss |
| `pointAbsent`      | float64 | `0.0`              | `0.0`              | Points for absence        |

### CLI flag to option key mapping

The `standings` subcommand maps its flags to scoring option keys:

| CLI flag         | Option key         |
| ---------------- | ------------------ |
| `--win`          | `pointWin`         |
| `--draw`         | `pointDraw`        |
| `--loss`         | `pointLoss`        |
| `--forfeit-win`  | `pointForfeitWin`  |
| `--bye`          | `pointBye`         |
| `--forfeit-loss` | `pointForfeitLoss` |

Flag values of `-1` (the default) are treated as unset and do not override engine defaults.

## RTG configuration file

The `generate` subcommand accepts a configuration file via `--config` for the Random Tournament Generator. The file uses a simple `key=value` format with `#` comments:

```text
# RTG configuration
PlayersNumber=100
RoundsNumber=11
DrawPercentage=25
HighestRating=2800
LowestRating=1200
```

### RTG configuration keys

| Key                    | Type    | Default | Description                                                                   |
| ---------------------- | ------- | ------- | ----------------------------------------------------------------------------- |
| `PlayersNumber`        | int     | `30`    | Number of players to generate                                                 |
| `RoundsNumber`         | int     | `9`     | Number of rounds to simulate                                                  |
| `DrawPercentage`       | int     | `30`    | Draw probability as a percentage (0-100)                                      |
| `ForfeitRate`          | int     | `20`    | Forfeit rarity factor (higher = rarer; probability is `1 - sqrt(1 - 1/rate)`) |
| `RetiredRate`          | int     | `100`   | Retirement rarity factor                                                      |
| `HalfPointByeRate`     | int     | `100`   | Half-point bye rarity factor                                                  |
| `HighestRating`        | int     | `2600`  | Upper bound for random ratings                                                |
| `LowestRating`         | int     | `1400`  | Lower bound for random ratings                                                |
| `PointsForWin`         | float64 | `1.0`   | Points awarded for a win                                                      |
| `PointsForDraw`        | float64 | `0.5`   | Points awarded for a draw                                                     |
| `PointsForLoss`        | float64 | `0.0`   | Points awarded for a loss                                                     |
| `PointsForZPB`         | float64 | `0.5`   | Points for zero-point bye                                                     |
| `PointsForForfeitLoss` | float64 | `0.0`   | Points for a forfeit loss                                                     |
| `PointsForPAB`         | float64 | `0.0`   | Points for pairing-allocated bye                                              |

Unknown keys are silently ignored. Blank lines and lines starting with `#` are skipped.

### Result simulation

The RTG uses a logistic Elo model for result simulation. Given a rating difference `d = whiteRating - blackRating`, the expected score for White is:

```text
E(white) = 1 / (1 + 10^(-d/400))
```

The draw probability is scaled based on the configured `DrawPercentage` and capped to remain consistent with the expected score. Forfeits are determined independently per player using the `ForfeitRate` parameter before result simulation.

### Seeding

The RTG seeds the PRNG using the `-s` flag. Integer strings are parsed directly as `int64` seeds. Non-integer strings are hashed via FNV-1a to produce a deterministic seed. If no seed is provided, a cryptographically random seed is generated and printed to stderr for reproducibility.
