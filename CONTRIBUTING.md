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
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run ./...

# Vet
go vet ./...
```

All three checks must pass before submitting changes.

## Guidelines

- No external dependencies. The module is pure stdlib Go.
- Return errors rather than panicking.
- Use natural, descriptive commit messages (no conventional commit prefixes).
- New features should include tests.

For detailed coding conventions, see the
[contributing guide](https://chesspairing.nl/docs/appendices/contributing/)
on the documentation site.

## Reporting issues

File issues at <https://github.com/gnutterts/chesspairing/issues>. Include
the input data (TRF file or TournamentState construction) and the expected
vs. actual output when reporting bugs.
