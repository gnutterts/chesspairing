---
title: "standings"
linkTitle: "standings"
weight: 6
description: "Compute and display tournament standings from a TRF file."
---

The `standings` subcommand computes scores, tiebreakers, and final standings from a TRF file. It requires a pairing system flag, which determines the default tiebreaker sequence via `DefaultTiebreakers(system)`. The scoring system, point values, and tiebreaker selection are all configurable.

## Synopsis

```text
chesspairing standings SYSTEM input-file [options]
```

The pairing system flag (e.g. `--dutch`) is required and must appear somewhere in the argument list -- it can come before or after the input file. The input file can be a filesystem path or `-` for stdin.

Flags and positional arguments can be interleaved in any order.

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

The system flag is consumed before other flags are parsed, so it can appear anywhere in the argument list. It does not select a pairing engine -- it is only used to determine the default tiebreaker sequence.

## Flags

| Flag             | Type    | Default    | Description                                                |
| ---------------- | ------- | ---------- | ---------------------------------------------------------- |
| `--scoring`      | string  | `standard` | Scoring system: `standard`, `keizer`, `football`           |
| `--tiebreakers`  | string  | --         | Comma-separated tiebreaker IDs (overrides system defaults) |
| `--win`          | float64 | --         | Override points for a win                                  |
| `--draw`         | float64 | --         | Override points for a draw                                 |
| `--loss`         | float64 | --         | Override points for a loss                                 |
| `--forfeit-win`  | float64 | --         | Override points for a forfeit win                          |
| `--bye`          | float64 | --         | Override points for a bye                                  |
| `--forfeit-loss` | float64 | --         | Override points for a forfeit loss                         |
| `--json`         | bool    | `false`    | Output as JSON                                             |
| `--help`         | --      | --         | Show usage help                                            |

Point override flags only take effect when set to a value >= 0. They map to scoring option keys: `pointWin`, `pointDraw`, `pointLoss`, `pointForfeitWin`, `pointBye`, `pointForfeitLoss`. These overrides are applied on top of any scoring options already present in the TRF file.

## Examples

```bash
# Default standings (standard scoring, system-default tiebreakers)
chesspairing standings --dutch tournament.trf

# Football scoring with custom tiebreakers
chesspairing standings --dutch tournament.trf --scoring football --tiebreakers buchholz,wins,progressive

# Custom point values
chesspairing standings --dutch tournament.trf --win 3 --draw 1 --loss 0

# Keizer scoring
chesspairing standings --keizer tournament.trf --scoring keizer

# JSON output
chesspairing standings --dutch tournament.trf --json

# Read from stdin
cat tournament.trf | chesspairing standings --dutch -
```

## Output

### Text format (default)

Tab-aligned table with rank, player ID, name, score, and one column per tiebreaker. The tiebreaker column headers use each tiebreaker's display name.

```text
Rank  ID  Name              Score  Buchholz Cut 1  Buchholz  Wins
----  --  ----              -----  --------------  --------  ----
1     5   GM Smith          7.0    28.5            35.0      6
2     1   IM Jones          7.0    27.0            34.5      5
3     12  FM Brown          6.5    26.0            33.0      5
```

Tied players share the same rank. Scores and tiebreaker values are formatted as integers when they are whole numbers, with one decimal place otherwise.

When there are no standings to display, the output is `(no standings)`.

### JSON format

```json
{
  "standings": [
    {
      "rank": 1,
      "playerId": "5",
      "displayName": "GM Smith",
      "score": 7,
      "tieBreakers": [
        { "id": "buchholz-cut1", "name": "Buchholz Cut 1", "value": 28.5 },
        { "id": "buchholz", "name": "Buchholz", "value": 35 },
        { "id": "wins", "name": "Wins", "value": 6 }
      ],
      "gamesPlayed": 9,
      "wins": 6,
      "draws": 2,
      "losses": 1
    }
  ],
  "scoring": "standard",
  "tiebreakers": ["buchholz-cut1", "buchholz", "wins"]
}
```

Each standing entry includes game statistics (`gamesPlayed`, `wins`, `draws`, `losses`) computed from the round data. The top-level `tiebreakers` array lists the tiebreaker IDs that were requested (including any that failed and were skipped).

## Tiebreaker selection

If `--tiebreakers` is not specified, the default tiebreaker sequence for the given pairing system is used. The defaults from `DefaultTiebreakers()` are:

| System                                          | Default tiebreakers                                                 |
| ----------------------------------------------- | ------------------------------------------------------------------- |
| Dutch, Burstein, Dubov, Lim, Double-Swiss, Team | `buchholz-cut1`, `buchholz`, `sonneborn-berger`, `direct-encounter` |
| Round-Robin                                     | `sonneborn-berger`, `direct-encounter`, `wins`, `koya`              |
| Keizer                                          | `games-played`, `direct-encounter`, `wins`                          |

Unknown tiebreaker IDs print a warning to stderr and are skipped. Failed tiebreaker computations also print a warning and are skipped. Neither case causes the command to fail.

Available tiebreaker IDs can be listed with the [tiebreakers](../tiebreakers-cmd/) subcommand.

## Standings computation

1. Parse the TRF file and convert to `TournamentState`
2. Run the selected scorer (`standard`, `keizer`, or `football`) to produce `PlayerScore` values
3. Compute each tiebreaker in order, skipping any that fail
4. Sort by score descending, then by each tiebreaker value in order (descending)
5. Assign shared ranks -- players with identical scores and identical tiebreaker values receive the same rank; the next rank is the position in the sorted list (e.g. two players tied at rank 1 are followed by rank 3)

## Exit codes

| Code | Meaning                                                            |
| ---- | ------------------------------------------------------------------ |
| 0    | Standings computed successfully                                    |
| 2    | Unexpected error (scoring failure, JSON encoding error)            |
| 3    | Invalid input (missing system flag, malformed TRF, unknown scorer) |
| 5    | File access error                                                  |

## See also

- [tiebreakers](../tiebreakers-cmd/) -- list available tiebreaker IDs
- [pair](../pair/) -- generate pairings for the next round
- [Output Formats and Exit Codes](../output-formats/) -- detailed format specifications and all exit codes
