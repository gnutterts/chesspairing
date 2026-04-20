---
title: "Contributing"
linkTitle: "Contributing"
weight: 3
description: "How to contribute to the chesspairing project."
---

## Prerequisites

- Go 1.24 or later
- No external dependencies are used -- the module is pure stdlib Go, and this is intentional

## Development Workflow

```bash
# Clone
git clone https://github.com/gnutterts/chesspairing.git
cd chesspairing

# Run tests
go test -race -count=1 ./...

# Lint
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run ./...

# Vet
go vet ./...
```

All three checks must pass before submitting changes.

## Code Style

- Standard Go formatting (`gofmt`). Do not use alternative formatters.
- Return errors rather than panicking.
- All engine methods accept `context.Context` as the first parameter.
- Engine configuration uses the pointer-field Options pattern: nil fields mean "use default". Each Options struct provides `WithDefaults()` and `ParseOptions(map[string]any)`.
- Compile-time interface checks in each engine package (e.g., `var _ chesspairing.Pairer = (*Pairer)(nil)`).
- No external dependencies. If you need functionality not in the standard library, implement it within the module.

## Commit Messages

Use natural, descriptive commit messages. No conventional commit prefixes (no `feat:`, `fix:`, etc.). Examples:

- "Add Dutch pairer with global Blossom matching"
- "Wire up Keizer scoring with iterative convergence"
- "Fix color allocation edge case in odd player counts"

## Testing

- All changes must pass the existing test suite (~1325 tests across 19 packages).
- New features should include tests.
- White-box tests (same package) are the norm. The root package uses black-box tests (`chesspairing_test`).
- The Dutch pairer uses golden file tests with self-generated, JaVaFo 2.2, and bbpPairings reference pairings.
- Fuzz testing is available for the TRF parser (`trf/fuzz_test.go`).

Run the full test suite with the race detector enabled:

```bash
go test -race -count=1 ./...
```

## License for contributions

By submitting a pull request or patch, you agree that your contribution is
licensed under the [Apache License 2.0](https://github.com/gnutterts/chesspairing/blob/main/LICENSE),
the same license that covers the rest of this project. See Section 5 of the
license for details.

## Areas of Interest

The following areas could benefit from contributions:

- **Additional pairing systems** -- New FIDE or non-FIDE pairing system implementations.
- **Performance optimization** -- Profiling and improving the Blossom matching or iterative scoring algorithms.
- **Documentation** -- Corrections, clarifications, and additional examples.
- **Bug reports** -- Reports from real tournament usage are particularly valuable for validating edge cases.

## Reporting Issues

File issues at [https://github.com/gnutterts/chesspairing/issues](https://github.com/gnutterts/chesspairing/issues). Include the input data (TRF file or `TournamentState` construction) and the expected vs. actual output when reporting bugs.
