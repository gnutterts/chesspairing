# Contributing

Thanks for your interest. This project is not actively seeking contributions,
but bug reports, test cases, and well-considered patches are welcome.

## Before you contribute

By submitting a pull request or patch, you agree that your contribution is
licensed under the [Apache License 2.0](LICENSE), the same license that
covers the rest of this project. See Section 5 of the license for details.

## Development workflow

```bash
# Run tests
go test -race -count=1 ./...

# Lint
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2 run ./...

# Vet
go vet ./...
```

All three checks must pass before submitting changes. The `main` branch is
protected: every change goes through a pull request, and CI must pass before it
can be merged. The golangci-lint version above matches the one pinned in
`.github/workflows/ci.yml`; Renovate keeps both up to date.

## Guidelines

- No external dependencies. The module is pure stdlib Go.
- Return errors rather than panicking.
- Use natural, descriptive commit messages (no conventional commit prefixes).
- New features should include tests.

For detailed coding conventions, see the
[contributing guide](https://chesspairing.nl/docs/appendices/contributing/)
on the documentation site.

## Use of AI tools

This project is developed with the help of AI tools for analysis, code, tests
and reviews. A human maintainer reviews and merges every change; the test suite,
the FIDE reference cases and the cross-checks against independent pairing
engines are the final arbiter, not the tool that produced a change. Commits made
with AI assistance carry the trailer `Assisted-by: AI tools (see CONTRIBUTING.md)`.
Configuration files for AI assistants (such as `CLAUDE.md` or `AGENTS.md`) are
not part of the repository; CI rejects them.

## Reporting issues

File issues at <https://github.com/gnutterts/chesspairing/issues>. Include
the input data (TRF file or TournamentState construction) and the expected
vs. actual output when reporting bugs.
