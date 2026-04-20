---
title: "Installation"
linkTitle: "Installation"
weight: 1
description: "How to install the chesspairing CLI tool."
---

## From source (go install)

With Go 1.24 or later installed:

```bash
go install github.com/gnutterts/chesspairing/cmd/chesspairing@latest
```

This builds the binary and places it in `$GOBIN` (usually `$HOME/go/bin`). Make sure that directory is on your `PATH`.

## Build from source

Clone the repository and build manually:

```bash
git clone https://github.com/gnutterts/chesspairing.git
cd chesspairing
go build -o chesspairing ./cmd/chesspairing
```

To embed a version string at build time:

```bash
go build -ldflags "-X main.version=1.0.0" -o chesspairing ./cmd/chesspairing
```

Without `-ldflags`, the version defaults to `dev`.

## Verify installation

```bash
chesspairing version
```

Expected output:

```text
chesspairing dev

Pairing systems:  dutch, burstein, dubov, lim, doubleswiss, team, keizer, roundrobin
Scoring systems:  standard, keizer, football
Tiebreakers:      25 available
```

## Requirements

- **Go 1.24 or later** for building from source.
- **No external dependencies.** The entire module uses only the Go standard library.
- The binary is statically linked and has no runtime dependencies.

## As a library

If you want to use chesspairing as a Go library rather than a CLI tool, add it to your module:

```bash
go get github.com/gnutterts/chesspairing@latest
```

See the [Go Quickstart](/docs/getting-started/go-quickstart/) for library usage.
