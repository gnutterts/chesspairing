---
title: "Changelog"
linkTitle: "Changelog"
weight: 2
description: "Version history and notable changes."
---

The canonical changelog lives in [`CHANGELOG.md`](https://github.com/gnutterts/chesspairing/blob/main/CHANGELOG.md) at the repository root. It follows a loose [Keep a Changelog](https://keepachangelog.com/) format and Semantic Versioning.

The entries below are rendered from that file, so this page always shows the current history, newest release first. The changelog is written in English.

{{< changelog >}}

## Build-time version string

During development, the version string defaults to `"dev"`. Release versions are set at build time via `-ldflags`:

```bash
go build -ldflags "-X main.version=v1.2.3" ./cmd/chesspairing
```
