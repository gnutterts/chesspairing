---
title: chesspairing
---

{{< blocks/cover title="chesspairing" image_anchor="top" height="med" color="dark" >}}
<div class="mx-auto">
  <a class="btn btn-lg btn-primary me-3 mb-4" href="/docs/">
    Documentation
  </a>
  <a class="btn btn-lg btn-secondary me-3 mb-4" href="https://github.com/gnutterts/chesspairing">
    GitHub
  </a>
  <p class="lead mt-4">Chess tournament pairing, scoring, and tiebreaking algorithms in pure Go.</p>
</div>
{{< /blocks/cover >}}

{{% blocks/lead color="primary" %}}

**Eight pairing systems. Three scoring engines. Twenty-five tiebreakers.**

A zero-dependency Go module implementing FIDE regulations for Swiss (Dutch, Burstein, Dubov, Lim, Double-Swiss, Team), Round-Robin, and Keizer tournaments. Ships with a CLI tool that serves as a drop-in replacement for bbpPairings and JaVaFo.

{{% /blocks/lead %}}

{{< blocks/section color="white" type="row" >}}

{{% blocks/feature icon="fa-chess" title="For Chess Players and Arbiters" url="/docs/getting-started/for-arbiters/" %}}
Understand how pairings are generated, why you got that opponent, and what each tiebreaker actually measures.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-terminal" title="For CLI Users" url="/docs/getting-started/cli-quickstart/" %}}
Install the tool and pair your first tournament in under five minutes. Supports TRF16 input with five output formats.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-code" title="For Go Developers" url="/docs/getting-started/go-quickstart/" %}}
Add `chesspairing` to your project with `go get`. Clean interfaces, no dependencies, safe for concurrent use.
{{% /blocks/feature %}}

{{< /blocks/section >}}

{{< blocks/section color="light" type="row" >}}

{{% blocks/feature icon="fa-chess-board" title="FIDE Compliant" %}}
Implements FIDE C.04.3 (Dutch), C.04.4.1 (Dubov), C.04.4.2 (Burstein), C.04.4.3 (Lim), C.04.5 (Double-Swiss), C.04.6 (Team Swiss), and C.05 (Round-Robin) with Berger tables and Varma number assignment.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-square-root-variable" title="Mathematically Rigorous" url="/docs/algorithms/" %}}
Built on Edmonds' Blossom matching algorithm with big.Int edge weights encoding 16+ criteria fields. Every optimization decision is traceable to a FIDE regulation article.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-vial" title="Thoroughly Tested" %}}
Over 1300 tests including golden file comparisons against bbpPairings and JaVaFo reference implementations, plus fuzz testing for the TRF parser.
{{% /blocks/feature %}}

{{< /blocks/section >}}
