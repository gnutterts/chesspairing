---
title: "generate"
linkTitle: "generate"
weight: 4
description: "Generate a complete tournament TRF file with random results for testing."
---

The `generate` subcommand creates a synthetic tournament TRF file by generating random players, pairing each round with the specified engine, and simulating random results. This is primarily a testing and benchmarking tool.

The generated tournament uses realistic Elo-based result simulation: win probability follows the logistic model (`1 / (1 + 10^(-diff/400))`), with configurable draw and forfeit rates.

Unlike `pair` and `check`, this subcommand does not read an existing TRF file. It creates a new tournament from scratch and requires `-o` for the output file.

## Synopsis

```text
chesspairing generate SYSTEM -o output-file [options]
```

The pairing system flag (e.g. `--dutch`) is required and must appear somewhere in the argument list. The `-o` flag specifying the output TRF file is also required.

## Pairing system flags

Exactly one system flag is required:

| Flag             | System                          |
| ---------------- | ------------------------------- |
| `--dutch`        | Dutch (FIDE C.04.3)             |
| `--burstein`     | Burstein (FIDE C.04.4.2)        |
| `--dubov`        | Dubov (FIDE C.04.4.1)           |
| `--lim`          | Lim (FIDE C.04.4.3)             |
| `--double-swiss` | Double-Swiss (FIDE C.04.5)      |
| `--team`         | Team Swiss (FIDE C.04.6)        |
| `--keizer`       | Keizer                          |
| `--roundrobin`   | Round-Robin (FIDE C.05 Annex 1) |

## Options

| Flag       | Type   | Default | Description                     |
| ---------- | ------ | ------- | ------------------------------- |
| `-o`       | string | —       | Output TRF file path (required) |
| `--config` | string | —       | Path to RTG configuration file  |
| `-s`       | string | —       | PRNG seed (integer or string)   |
| `--help`   | —      | —       | Show usage help                 |

## Configuration

When `--config` is provided, the file is parsed as `key=value` pairs (one per line, `#` for comments, blank lines ignored). Unknown keys produce a warning to stderr but do not cause the command to fail.

| Key                    | Type    | Default | Description                                                      |
| ---------------------- | ------- | ------- | ---------------------------------------------------------------- |
| `PlayersNumber`        | int     | 30      | Number of players                                                |
| `RoundsNumber`         | int     | 9       | Number of rounds                                                 |
| `DrawPercentage`       | int     | 30      | Draw probability (%)                                             |
| `ForfeitRate`          | int     | 20      | Inverse forfeit probability (higher = fewer forfeits)            |
| `RetiredRate`          | int     | 100     | Retirement rate (reserved, not yet implemented)                  |
| `HalfPointByeRate`     | int     | 100     | Half-point bye request rate (reserved, not yet implemented)      |
| `HighestRating`        | int     | 2600    | Maximum player rating                                            |
| `LowestRating`         | int     | 1400    | Minimum player rating                                            |
| `PointsForWin`         | float64 | 1.0     | Points for a win (reserved, not yet implemented)                 |
| `PointsForDraw`        | float64 | 0.5     | Points for a draw (reserved, not yet implemented)                |
| `PointsForLoss`        | float64 | 0.0     | Points for a loss (reserved, not yet implemented)                |
| `PointsForZPB`         | float64 | 0.5     | Points for zero-point bye (reserved, not yet implemented)        |
| `PointsForForfeitLoss` | float64 | 0.0     | Points for forfeit loss (reserved, not yet implemented)          |
| `PointsForPAB`         | float64 | 0.0     | Points for pairing-allocated bye (reserved, not yet implemented) |

Example config file:

```text
# Large tournament
PlayersNumber=100
RoundsNumber=11
DrawPercentage=25
HighestRating=2700
LowestRating=1200
```

## Seed behavior

- If `-s` is an integer, it is used directly as the PRNG seed.
- If `-s` is a non-integer string, it is hashed with FNV-1a to produce a 64-bit seed.
- If `-s` is omitted, 8 random bytes from `crypto/rand` are used and the seed is printed to stderr for reproducibility.

The PRNG is `math/rand/v2` with a PCG source.

## Examples

```bash
# Generate a 30-player, 9-round tournament with defaults
chesspairing generate --dutch -o tournament.trf

# Reproducible generation with a numeric seed
chesspairing generate --dutch -o test.trf -s 42

# String seed (hashed via FNV-1a)
chesspairing generate --burstein -o large.trf -s myseed

# With custom configuration
chesspairing generate --burstein -o large.trf --config rtg.conf -s myseed
```

## Player generation

Players are assigned random ratings uniformly distributed in [`LowestRating`, `HighestRating`], then sorted by rating descending. The highest-rated player receives start number 1. Names are generated as "Player N" where N is the start number.

## Result simulation

For each round:

1. Convert the current TRF document to a `TournamentState`.
2. Pair with the specified engine.
3. For each pairing, check for forfeits first -- the per-player forfeit probability is `sqrt(1 - 1/ForfeitRate)`. If both players forfeit, it is a double forfeit; if one forfeits, the other wins by forfeit.
4. For non-forfeit games, determine the result using the logistic Elo model. Draw probability is capped by `DrawPercentage`. The remaining probability mass is split between white and black wins proportional to expected score.
5. Write results back into the TRF document.

## Exit codes

| Code | Meaning                                  |
| ---- | ---------------------------------------- |
| 0    | Tournament generated successfully        |
| 1    | Pairing failed for a round               |
| 2    | Unexpected error during generation       |
| 3    | Invalid input (bad config, missing `-o`) |
| 5    | File access error                        |

## See also

- [pair](../pair/) -- generate pairings for an existing tournament
- [Legacy Mode](../legacy/) -- bbpPairings/JaVaFo drop-in replacement interface
