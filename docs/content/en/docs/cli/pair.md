---
title: "pair"
linkTitle: "pair"
weight: 2
description: "Read a TRF file and generate pairings for the next round."
---

The `pair` subcommand reads a TRF16 file, runs the specified pairing engine, and outputs the next round's pairings. This is the primary subcommand for most workflows.

## Synopsis

```text
chesspairing pair SYSTEM input-file [options]
```

The pairing system flag (e.g. `--dutch`) is required and must appear somewhere in the argument list -- it can come before or after the input file. The input file can be a filesystem path or `-` for stdin.

Flags and positional arguments can be interleaved in any order. A bare `--` terminates flag processing; everything after it is treated as positional.

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

The system flag is consumed before other flags are parsed, so it can appear anywhere in the argument list.

## Options

| Flag       | Type   | Default | Description                                           |
| ---------- | ------ | ------- | ----------------------------------------------------- |
| `--format` | string | `list`  | Output format: `list`, `wide`, `board`, `xml`, `json` |
| `-w`       | bool   | `false` | Shorthand for `--format wide`                         |
| `--json`   | bool   | `false` | Shorthand for `--format json` (backward compatible)   |
| `-o`       | string | stdout  | Write output to a file instead of stdout              |
| `--help`   | —      | —       | Show usage help                                       |

**Format resolution priority:** `--format` > `-w` > `--json` > default `list`.

If `--format` is set explicitly, `-w` and `--json` are ignored. If multiple shorthands are given without `--format`, `-w` takes precedence over `--json`.

## Examples

```bash
# Default list format (bbpPairings-compatible)
chesspairing pair --dutch tournament.trf

# Wide tabular format with names and ratings
chesspairing pair --dutch tournament.trf -w

# JSON output written to file
chesspairing pair --dutch tournament.trf --format json -o pairings.json

# Board format with the Lim pairing system
chesspairing pair --lim tournament.trf --format board

# XML format
chesspairing pair --dutch tournament.trf --format xml

# Read from stdin
cat tournament.trf | chesspairing pair --dutch -

# Flags and input file can be interleaved
chesspairing pair tournament.trf --burstein -o result.txt --format wide
```

## Output formats

### list (default)

Compact, machine-readable format compatible with bbpPairings and JaVaFo. The first line is the number of board pairings. Each subsequent line is a `white black` pair of start numbers. Byes are listed after the pairings as `player 0`.

```text
5
5 1
3 12
8 2
6 9
4 11
7 0
```

### wide

Human-readable table with board numbers, player start numbers, titles, names, and ratings. Byes are listed after the last board.

```text
Board  White              Rtg   -  Black              Rtg
-----  -----              ---      -----              ---
1      5 GM Smith          2600  -  1 IM Jones         2500
2      3 WGM Lee           2400  -  12 FM Petrov       2350
       7 Brown             1800     Bye (PAB)
```

### board

Compact numbered board list with start numbers only. Byes follow the last board.

```text
Board  1:  5 -  1
Board  2:  3 - 12
Board  3:  8 -  2
Bye:  7
```

### json

Structured JSON with a `pairings` array and an optional `byes` array. Board numbers are 1-indexed. Bye types use their string representation (`PAB`, `Half`, `Zero`, `Absent`, `Excused`, `ClubCommitment`).

```json
{
  "pairings": [
    { "board": 1, "white": 5, "black": 1 },
    { "board": 2, "white": 3, "black": 12 },
    { "board": 3, "white": 8, "black": 2 }
  ],
  "byes": [{ "player": 7, "type": "PAB" }]
}
```

The `byes` key is omitted when there are no byes.

### xml

XML document with player metadata (name, rating, title) on each board element. Includes a root `<pairings>` element with `round`, `boards`, and `byes` attributes.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<pairings round="4" boards="3" byes="1">
  <board number="1">
    <white number="5" name="Smith" rating="2600" title="GM"></white>
    <black number="1" name="Jones" rating="2500" title="IM"></black>
  </board>
  <bye number="7" name="Brown" type="PAB"></bye>
</pairings>
```

See [Output Formats and Exit Codes](../output-formats/) for full format specifications.

## Exit codes

| Code | Meaning                                                            |
| ---- | ------------------------------------------------------------------ |
| 0    | Pairings generated successfully                                    |
| 1    | No valid pairing could be produced                                 |
| 3    | Invalid input (malformed TRF, unknown format, missing system flag) |
| 5    | File could not be opened or written                                |

## Pairing engine options

Each pairing system accepts engine-specific options through the TRF file's `XXY` extension field. These options control behavior such as Baku acceleration, top-seed color, forbidden pairs, and total rounds. Refer to the documentation for each [pairing system](../../pairing-systems/) for available options.

## See also

- [check](../check/) -- verify existing pairings against engine output
- [generate](../generate/) -- generate pairings and output an updated TRF
- [Legacy Mode](../legacy/) -- bbpPairings/JaVaFo drop-in replacement interface
- [Output Formats and Exit Codes](../output-formats/) -- detailed format specifications and all exit codes
