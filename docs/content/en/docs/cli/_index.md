---
title: "CLI Reference"
linkTitle: "CLI Reference"
weight: 60
description: "Complete reference for the chesspairing command-line tool — all subcommands, flags, and output formats."
---

## Overview

The `chesspairing` binary processes [TRF16](https://www.fide.com) tournament files and provides pairing, validation, scoring, and utility functions. All subcommands follow a consistent pattern: they accept a TRF file as input, perform an operation, and produce output to stdout (or an output file via `-o`).

The general invocation pattern is:

```text
chesspairing COMMAND [SYSTEM] input.trf [options]
```

Where `COMMAND` is one of the subcommands listed below, `SYSTEM` is a pairing system flag (required by some commands), and `input.trf` is the path to a TRF16 tournament file.

## Subcommands

| Command                         | Description                                  | Requires System Flag |
| ------------------------------- | -------------------------------------------- | -------------------- |
| [pair](pair/)                   | Generate pairings for the next round         | Yes                  |
| [check](check/)                 | Verify existing pairings match engine output | Yes                  |
| [generate](generate/)           | Generate a complete TRF with random results  | Yes                  |
| [validate](validate/)           | Validate TRF file structure                  | No                   |
| [standings](standings/)         | Compute and display standings                | Yes                  |
| [tiebreakers](tiebreakers-cmd/) | List available tiebreakers                   | No                   |
| [convert](convert/)             | Re-serialize a TRF file                      | No                   |
| [version](version/)             | Display version info                         | No                   |

Each subcommand has its own page with full usage examples and flag descriptions. Use `chesspairing <command> --help` for inline help.

## System flags

Commands that require a pairing system accept one of these flags before the input file:

| Flag             | Pairing System                  |
| ---------------- | ------------------------------- |
| `--dutch`        | Dutch (FIDE C.04.3)             |
| `--burstein`     | Burstein (FIDE C.04.4.2)        |
| `--dubov`        | Dubov (FIDE C.04.4.1)           |
| `--lim`          | Lim (FIDE C.04.4.3)             |
| `--double-swiss` | Double-Swiss (FIDE C.04.5)      |
| `--team`         | Team Swiss (FIDE C.04.6)        |
| `--keizer`       | Keizer                          |
| `--roundrobin`   | Round-robin (FIDE C.05 Annex 1) |

The system flag determines which pairing engine (and its associated default scorer) is used for the operation. When a TRF file contains system-specific `XX` fields, those options are passed through to the engine automatically.

## Input handling

Input files can be specified as:

- A file path: `chesspairing pair --dutch tournament.trf`
- A dash (`-`) for stdin: `cat tournament.trf | chesspairing pair --dutch -`

If no input file is specified, the tool reports an error and exits with code 3 (`ExitInvalidInput`).

## Output

Most commands write to stdout by default. Where supported, use `-o` to redirect output to a file. The `--json` flag is available on most commands for machine-readable output. The `pair` subcommand additionally supports `--format` with values `list`, `wide`, `board`, `xml`, and `json`. See [Output Formats and Exit Codes](output-formats/) for details.

## Legacy mode

When invoked without a recognized subcommand, the tool falls back to legacy mode -- a bbpPairings/JaVaFo-compatible positional argument interface. This allows `chesspairing` to serve as a drop-in replacement in existing toolchains:

```bash
chesspairing --dutch input.trf -p
chesspairing --dutch input.trf -c
```

See [Legacy Mode](legacy/) for the full positional argument interface.

## Exit codes

| Code | Constant           | Meaning                                    |
| ---- | ------------------ | ------------------------------------------ |
| 0    | `ExitSuccess`      | Operation completed successfully           |
| 1    | `ExitNoPairing`    | No valid pairing could be produced         |
| 2    | `ExitUnexpected`   | Unexpected runtime error                   |
| 3    | `ExitInvalidInput` | Invalid or malformed input                 |
| 4    | `ExitSizeOverflow` | Tournament size exceeds limits             |
| 5    | `ExitFileAccess`   | File could not be opened, read, or written |

See [Output Formats and Exit Codes](output-formats/) for detailed descriptions.
