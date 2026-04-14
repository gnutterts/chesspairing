# Security policy

## Supported versions

This project is in early development and has no users beyond the author yet.
Only the `main` branch is supported. There are no security backports.

## Reporting a vulnerability

For anything you would not want to disclose in public — credential leaks,
remote code execution paths, parser exploits with security implications —
please use GitHub's private vulnerability reporting:

<https://github.com/gnutterts/chesspairing/security/advisories/new>

For everything else (including correctness bugs in pairings, scoring, or
tiebreakers), open a public issue at
<https://github.com/gnutterts/chesspairing/issues>.

## Scope

The library has no I/O, no network calls, and no external dependencies, so
the realistic attack surface is small:

- The TRF parser (`trf/`) processes untrusted text input and is the most
  likely place for parsing-related bugs. It is fuzz-tested but not
  exhaustively audited.
- The CLI reads files and writes to stdout/stderr; it does not execute
  shell commands or fetch network resources.

## Response

This is a solo, hobby-time project. Best-effort response within a few days
is the realistic expectation, not a guarantee.
