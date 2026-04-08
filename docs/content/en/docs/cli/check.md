---
title: "check"
linkTitle: "check"
weight: 3
description: "Verify that existing pairings match the engine's output."
---

The `check` subcommand verifies that the pairings in the last round of a TRF file match what the specified pairing engine would produce. It strips the last round from the tournament state, re-pairs using the given system, and compares the result against the existing pairings.

This is useful for validating that a tournament was paired correctly, or for testing engine implementations against reference files.

## Synopsis

```text
chesspairing check SYSTEM input-file [options]
```

The pairing system flag (e.g. `--dutch`) is required and must appear somewhere in the argument list. The input file can be a filesystem path or `-` for stdin.

## Comparison logic

The comparison is order-independent (set-based). It checks that:

- The same player pairs appear (regardless of board assignment)
- The same bye assignments are present
- The pairing count matches

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

| Flag     | Type | Default | Description           |
| -------- | ---- | ------- | --------------------- |
| `--json` | bool | `false` | Output result as JSON |
| `--help` | --   | --      | Show usage help       |

## Examples

```bash
# Text output
chesspairing check --dutch tournament.trf

# Read from stdin
chesspairing check --dutch - < tournament.trf

# JSON output
chesspairing check --dutch tournament.trf --json
```

## Output

**Text format** (default):

- `OK: pairings match` when the re-paired result matches.
- `MISMATCH: generated pairings differ from existing round` when they differ.

**JSON format:**

```json
{
  "match": true,
  "system": "dutch",
  "round": 5
}
```

The `round` field indicates which round was checked (the last round of the input file).

## Exit codes

| Code | Meaning                                                       |
| ---- | ------------------------------------------------------------- |
| 0    | Pairings match                                                |
| 1    | Pairings mismatch or pairing failed                           |
| 3    | Invalid input (no rounds, malformed TRF, missing system flag) |
| 5    | File could not be opened                                      |

## How it works

1. Parse the TRF file and convert to a `TournamentState`.
2. Save the last round's pairings and byes.
3. Remove the last round from the tournament state and decrement the current round.
4. Re-pair with the specified engine using the tournament's pairing options.
5. Compare the re-generated pairings against the saved last round (set-based comparison of white/black pairs and bye assignments).

## See also

- [pair](../pair/) -- generate pairings for the next round
- [generate](../generate/) -- generate pairings and output an updated TRF
- [Legacy Mode](../legacy/) -- bbpPairings/JaVaFo drop-in replacement interface
