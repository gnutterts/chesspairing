# Contributing

Thanks for your interest. This project is not actively seeking contributions,
but bug reports, test cases, and well-considered patches are welcome.

## Before you contribute

By submitting a contribution (pull request, patch, or any other form), you
agree that all intellectual property rights to that contribution are
transferred to the project author (Gert Nutterts). This keeps the license
clean and simple. You may be asked to sign a short contribution agreement
to formalize this.

Every accepted contributor is credited by name in CONTRIBUTORS.md.

See the [LICENSE.md](LICENSE.md) §Contributions for the full terms.

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
[contributing guide](https://gnutterts.github.io/chesspairing/docs/appendices/contributing/)
on the documentation site.

## Reporting issues

File issues at <https://github.com/gnutterts/chesspairing/issues>. Include
the input data (TRF file or TournamentState construction) and the expected
vs. actual output when reporting bugs.
