# Versioning policy

Chesspairing follows [Semantic Versioning](https://semver.org/). Until 1.0.0,
a public API break is made only in a minor release and is listed in
[CHANGELOG.md](CHANGELOG.md). Patch releases do not intentionally break the
public API.

The supported Go versions are the two latest Go releases. The module declares
its minimum version in [go.mod](go.mod); CI tests `oldstable` and `stable` on
Linux, macOS, and Windows.

A deprecated public API remains available for at least one minor release
before removal. Its replacement and intended removal release are documented in
the changelog.
