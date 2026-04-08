---
title: "Legacy Mode"
linkTitle: "Legacy Mode"
weight: 10
description: "Drop-in replacement for bbpPairings and JaVaFo — positional argument interface."
---

Legacy mode provides backward compatibility with bbpPairings and JaVaFo command-line interfaces. It activates automatically when the first argument is not a recognized subcommand (such as `pair`, `check`, `generate`, `validate`, `standings`, `tiebreakers`, `convert`, or `version`), allowing `chesspairing` to serve as a drop-in replacement in existing scripts.

## Usage patterns

```text
chesspairing SYSTEM input-file -p [output-file]    # pair
chesspairing SYSTEM input-file -c                   # check
chesspairing SYSTEM -g [config-file] -o output      # generate
```

The pairing system flag (e.g. `--dutch`) can appear anywhere in the argument list. Flags and positional arguments can be interleaved in any order.

## Flags

| Flag | Description                                                                                                  |
| ---- | ------------------------------------------------------------------------------------------------------------ |
| `-p` | Pair mode. Optional next argument is the output file.                                                        |
| `-c` | Check mode.                                                                                                  |
| `-g` | Generate mode. Optional next argument is a config file.                                                      |
| `-o` | Output file (required for generate mode).                                                                    |
| `-s` | PRNG seed (for generate mode).                                                                               |
| `-r` | Show version. Can be used alone or combined with a mode flag.                                                |
| `-w` | Wide output format (pair mode only).                                                                         |
| `-q` | JaVaFo compatibility flag. Accepted and ignored; an optional numeric argument following it is also consumed. |

Exactly one pairing system flag is required (unless using `-r` alone):

| Flag             | System       |
| ---------------- | ------------ |
| `--dutch`        | Dutch        |
| `--burstein`     | Burstein     |
| `--dubov`        | Dubov        |
| `--lim`          | Lim          |
| `--double-swiss` | Double-Swiss |
| `--team`         | Team Swiss   |
| `--keizer`       | Keizer       |
| `--roundrobin`   | Round-Robin  |

## Mode dispatch

### Pair (`-p`)

Reads the input TRF file, runs the pairing engine, and writes the result.

- Default output is `list` format (bbpPairings-compatible): pairing count on the first line, then `white black` start number pairs.
- With `-w`, output uses `wide` format (tabular with names and ratings).
- Writes to the file specified after `-p`, or to stdout if no file follows.

### Check (`-c`)

Strips the last round from the TRF, re-pairs, and compares against the existing pairings.

- Text output only: `OK: pairings match` or `MISMATCH: generated pairings differ from existing round`.
- Exit code 0 on match, 1 on mismatch or pairing failure.

### Generate (`-g`)

Delegates internally to the `generate` subcommand. The optional config file argument follows `-g`; the output file is specified with `-o`.

### Version (`-r`)

- `-r` alone: prints version information and exits.
- `-r` combined with a mode flag (e.g. `-r -p`): prints version, a blank line, then runs the specified mode.

## Differences from subcommands

| Feature             | Legacy mode              | Subcommand equivalent                  |
| ------------------- | ------------------------ | -------------------------------------- |
| Pair output formats | `list` and `wide` only   | `list`, `wide`, `board`, `xml`, `json` |
| Check JSON output   | Not available            | `--json` flag                          |
| Generate            | Delegates to subcommand  | Full flag set                          |
| Flag parsing        | Manual positional parser | Go `flag` package                      |

## Examples

```bash
# bbpPairings-compatible pairing
chesspairing --dutch tournament.trf -p

# Write pairings to file, wide format
chesspairing --dutch tournament.trf -p pairings.txt -w

# Check pairings
chesspairing --dutch tournament.trf -c

# Generate with seed
chesspairing --dutch -g config.txt -o output.trf -s 42

# Version
chesspairing -r
```

## Exit codes

| Code | Meaning                                   |
| ---- | ----------------------------------------- |
| 0    | Success (or pairings match in check mode) |
| 1    | No valid pairing or pairings mismatch     |
| 3    | Invalid input or missing required flags   |
| 5    | File access error                         |

## See also

- [pair](../pair/) -- full subcommand with all output formats
- [check](../check/) -- full subcommand with JSON output support
- [generate](../generate/) -- full subcommand for TRF generation
