# Dutch harness data

`corpus/` holds 150 complete tournaments (seeds 1001-1150, four configurations, 9 to 48
players, 3 to 9 rounds, with forfeits, requested byes and withdrawals). They were
generated once with the random tournament generator of bbpPairings v6.0.0 on Linux
(GCC, libstdc++, the platform bbpPairings is developed on).

`oracle.json` holds bbpPairings' own pairing (`--dutch -p`) for every round of every
tournament: 898 answers. They were computed on macOS and reproduced byte for byte by a
build on Linux (g++ 15.2), so the pairing is platform independent. The generator is not:
the same seed gives different tournaments with libc++ and libstdc++, which is why the
tournaments are stored instead of generated on every run.

`baseline.json` is the accepted state of our Dutch pairing against that oracle: all 898
rounds equal, pairs and colours. Any difference fails the test.

`regress/` holds 15 rounds, found by the nightly exploration, where an earlier version of
our engine differed from bbpPairings, with bbpPairings' exact answer; they run in the normal
test run as well.

## Tests

- `go test ./internal/harness` runs `TestHarnessCorpus`: our pairing against the stored
  answers. No external program needed, same result everywhere.
- `go test -tags harness ./internal/harness -run TestHarnessOracle` needs
  `bbpPairings` 6.0.0 on `PATH` and checks that it still gives the stored answers.
- `HARNESS_N=100 HARNESS_SEED=5000 go test -tags harness ./internal/harness -run TestHarnessExplore`
  pairs freshly generated tournaments with both engines to look for new differences. The
  nightly workflow runs it with a seed that changes every day.

## Regenerating

On Linux with bbpPairings v6.0.0 (tag commit `16a000f9811de322b0e835d5643198226165b5a9`,
built with `make`) on `PATH`:

    HARNESS_WRITE_CORPUS=1 go test -tags harness ./internal/harness -run TestHarnessWriteCorpus
    HARNESS_WRITE_ORACLE=1 go test -tags harness ./internal/harness -run TestHarnessWriteOracle

Then update `baseline.json` from the output of `TestHarnessCorpus`.

## Why not JaVaFo

JaVaFo 2.2 (2018) implements the 2017 Dutch rules. bbpPairings 6.0.0 implements the rules
effective 1 February 2026, which our engine follows. They differ where the rules changed
(for example the downfloat given after a forfeit loss and the minimisation of the score of
the player who receives the pairing-allocated bye), so JaVaFo is not used as an oracle.
