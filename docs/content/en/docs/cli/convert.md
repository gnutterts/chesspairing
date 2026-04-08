---
title: "convert"
linkTitle: "convert"
weight: 8
description: "Re-serialize a TRF file, normalizing formatting."
---

The `convert` subcommand reads a TRF16 file and writes it back out, normalizing field ordering and formatting. This is useful for cleaning up hand-edited TRF files or standardizing output from other tools.

## Synopsis

```text
chesspairing convert input-file -o output-file [options]
```

No pairing system flag required. The input file can be a filesystem path or `-` for stdin.

## Options

| Flag           | Type   | Default    | Description                              |
| -------------- | ------ | ---------- | ---------------------------------------- |
| `-o`           | string | (required) | Output file path                         |
| `--trf-format` | string | `trf2026`  | Output format: `trf`, `trfbx`, `trf2026` |
| `--help`       | --     | --         | Show usage help                          |

Both `-o` and the input file are required. Omitting either produces exit code 3.

## TRF format flag

The `--trf-format` flag accepts three values: `trf`, `trfbx`, and `trf2026`. Unknown values are rejected with exit code 3.

**Important:** Only `trf2026` is fully supported. When `trf` or `trfbx` is specified, the CLI writes to stderr:

```text
warning: --trf-format FORMAT not yet supported by library, writing default format
```

The actual output always uses the default TRF serializer regardless of the format value. This flag exists for forward compatibility; alternate serializers will be added in future releases.

## Examples

```bash
# Normalize a TRF file
chesspairing convert tournament.trf -o normalized.trf

# Read from stdin
chesspairing convert - -o output.trf < tournament.trf

# Explicitly set output format (default)
chesspairing convert tournament.trf -o output.trf --trf-format trf2026
```

## Exit codes

| Code | Meaning                                                |
| ---- | ------------------------------------------------------ |
| 0    | Success                                                |
| 2    | Write error (TRF serialization failed)                 |
| 3    | Invalid input (missing args, bad TRF, unknown format)  |
| 5    | File access error (cannot open input or create output) |

## See also

- [validate](../validate/) -- validate a TRF file against a profile
- [pair](../pair/) -- generate pairings from a TRF file
