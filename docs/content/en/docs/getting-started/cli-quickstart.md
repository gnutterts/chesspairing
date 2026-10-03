---
title: "CLI Quickstart"
linkTitle: "CLI Quickstart"
weight: 2
description: "Install the chesspairing CLI and pair your first tournament in under five minutes."
---

This page walks you through installing the `chesspairing` command-line tool, generating your first set of pairings from a TRF16 file, and exploring the available output formats.

## Prerequisites

You need [Go](https://go.dev/dl/) 1.26 or later installed. Verify with:

```bash
go version
```

## Install

```bash
go install github.com/gnutterts/chesspairing/cmd/chesspairing@latest
```

Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is on your `PATH`. Confirm the installation:

```bash
chesspairing version
```

## Your first pairing

The `pair` subcommand reads a TRF16 tournament file and generates pairings for the next round. You must specify a pairing system with a flag like `--dutch`, `--burstein`, `--dubov`, `--lim`, `--double-swiss`, `--team`, `--keizer`, or `--roundrobin`.

Suppose you have a file `tournament.trf` with three completed rounds of a Dutch Swiss tournament. Generate the round 4 pairings:

```bash
chesspairing pair --dutch tournament.trf
```

The default output is the **list** format -- a compact, machine-readable format compatible with bbpPairings and JaVaFo. The first line is the number of pairings, followed by one `white black` pair per line (start numbers):

```text
6
5 1
3 8
7 4
2 10
11 6
9 0
```

A `0` on the right indicates a bye.

## Output formats

The `pair` command supports five output formats via the `--format` flag.

### list (default)

```bash
chesspairing pair --dutch tournament.trf --format list
```

Compact pair list as shown above. This is what pairing-management software expects.

### wide

```bash
chesspairing pair --dutch tournament.trf --format wide
```

Shorthand: `-w`

Human-readable table with board numbers, player names, titles, and ratings:

```text
Board  White                Rtg     Black                Rtg
-----  -----                ---     -----                ---
1      5 GM Carlsen, Magnus  2830  -  1 GM Caruana, Fabiano  2786
2      3 IM Doe, Jane        2412  -  8 FM Smith, John       2350
...
```

### board

```bash
chesspairing pair --dutch tournament.trf --format board
```

Numbered board view:

```text
Board 1:  5 -  1
Board 2:  3 -  8
Board 3:  7 -  4
```

### json

```bash
chesspairing pair --dutch tournament.trf --format json
```

Shorthand: `--json`

Structured JSON with board numbers and bye details:

```json
{
  "pairings": [
    { "board": 1, "white": 5, "black": 1 },
    { "board": 2, "white": 3, "black": 8 }
  ],
  "byes": [{ "player": 9, "type": "PAB" }]
}
```

### xml

```bash
chesspairing pair --dutch tournament.trf --format xml
```

XML with player details (names, ratings, titles) on each board element.

## Writing output to a file

Use `-o` to write pairings to a file instead of stdout:

```bash
chesspairing pair --dutch tournament.trf -o round4.txt
```

## Reading from stdin

Pass `-` as the filename to read the TRF from stdin:

```bash
cat tournament.trf | chesspairing pair --dutch -
```

## Other commands

Beyond `pair`, the CLI offers several other subcommands. Run any of them with `--help` for full usage details.

| Command       | What it does                                                                             |
| ------------- | ---------------------------------------------------------------------------------------- |
| `check`       | Re-pairs the last round and compares against the existing pairings to verify correctness |
| `standings`   | Computes and displays tournament standings with configurable scoring and tiebreakers     |
| `validate`    | Validates a TRF16 file against a profile (minimal, standard, or strict)                  |
| `generate`    | Generates a random tournament (bbpPairings RTG-compatible)                               |
| `convert`     | Converts between TRF file formats                                                        |
| `tiebreakers` | Lists all 25 available tiebreaker algorithms                                             |
| `version`     | Shows version and supported pairing systems                                              |

Quick examples:

```bash
# Verify the last round's pairings
chesspairing check --dutch tournament.trf

# Show standings with Buchholz and wins as tiebreakers
chesspairing standings --dutch tournament.trf --tiebreakers buchholz,wins

# Validate a TRF file with strict FIDE checks
chesspairing validate tournament.trf --profile strict
```

## Legacy mode

If you are migrating from bbpPairings or JaVaFo, `chesspairing` supports their positional-argument interface directly:

```bash
# Pair (legacy)
chesspairing --dutch tournament.trf -p

# Check (legacy)
chesspairing --dutch tournament.trf -c
```

Legacy mode activates automatically when the first argument is a system flag rather than a subcommand. See the [Legacy Mode reference](/docs/cli/legacy/) for the full details.

## Exit codes

When scripting with `chesspairing`, check the exit code to determine the result:

| Code | Meaning                                       |
| ---- | --------------------------------------------- |
| 0    | Success                                       |
| 1    | No valid pairing could be produced            |
| 2    | Unexpected runtime error                      |
| 3    | Invalid or malformed input                    |
| 4    | Tournament size exceeds implementation limits |
| 5    | File could not be opened, read, or written    |

## Next steps

- [CLI Reference](/docs/cli/) -- full documentation for every subcommand and flag
- [Output Formats & Exit Codes](/docs/cli/output-formats/) -- detailed format specifications
- [Go Library Quickstart](/docs/getting-started/go-quickstart/) -- use chesspairing as a Go library instead
