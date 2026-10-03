---
title: "Testing"
weight: 5
---

Chesspairing uses unit tests, FIDE examples, golden CLI tests, fuzz tests for
TRF processing, and a stored Dutch differential corpus.

## Dutch differential harness

`internal/harness/testdata/corpus/` contains 150 complete tournaments (898
rounds), including forfeits, requested byes, and withdrawals. `oracle.json`
contains the Dutch pairings produced by bbpPairings 6.0.0; `baseline.json`
records the accepted result from chesspairing. The normal test compares pairs
and colours with those stored answers, so it needs no external program and is
reproducible on every platform:

```bash
go test ./internal/harness
```

The oracle is bbpPairings 6.0.0 because it implements the Dutch rules
effective 1 February 2026, which this implementation follows. JaVaFo 2.2
implements the 2017 rules and is therefore not the oracle: among other
changes, the current rules differ on the downfloat after a forfeit loss and on
minimising the score of the player receiving the pairing-allocated bye.

With bbpPairings 6.0.0 on `PATH`, these optional checks verify the stored
oracle or explore freshly generated tournaments:

```bash
go test -tags harness ./internal/harness -run TestHarnessOracle
HARNESS_N=100 HARNESS_SEED=5000 go test -tags harness ./internal/harness -run TestHarnessExplore
```

The corpus is stored rather than regenerated during a test because the random
generator gives different tournaments on platforms with different C++ standard
libraries. See `internal/harness/testdata/README.md` for regeneration steps.
