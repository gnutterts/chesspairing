---
title: ChessPairing
description: "Go library and CLI for chess tournament pairing, scoring, and tiebreaking according to FIDE regulations."
---

{{< blocks/cover title="ChessPairing" image_anchor="top" height="med" color="dark" >}}

<div class="mx-auto">
  <a class="btn btn-lg btn-primary me-3 mb-4" href="/docs/">
    Documentation
  </a>
  <a class="btn btn-lg btn-secondary me-3 mb-4" href="https://github.com/gnutterts/chesspairing">
    GitHub
  </a>
  <p class="lead mt-4">A Go module for chess tournament pairing, scoring, and tiebreaking.</p>
</div>
{{< /blocks/cover >}}

{{% blocks/lead color="primary" %}}

Go library and CLI tool for pairing chess tournaments according to FIDE regulations. Supports Swiss (Dutch, Burstein, Dubov, Lim, Double-Swiss, Team), Round-Robin, and Keizer. No external dependencies.

{{% /blocks/lead %}}

{{< blocks/section color="white" type="row" >}}

{{% blocks/feature icon="fa-chess" title="Players and Arbiters" url="/docs/getting-started/for-arbiters/" %}}
How pairings work, what the tiebreakers measure, and how to read the output.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-terminal" title="CLI" url="/docs/getting-started/cli-quickstart/" %}}
Install the tool, feed it a TRF file, get pairings back. Works as a drop-in for bbpPairings and JaVaFo.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-code" title="Go API" url="/docs/getting-started/go-quickstart/" %}}
`go get github.com/gnutterts/chesspairing` and pair programmatically. No dependencies, safe for concurrent use.
{{% /blocks/feature %}}

{{< /blocks/section >}}
