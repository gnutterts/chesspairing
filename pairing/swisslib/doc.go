// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package swisslib provides shared data structures and algorithms for the
// Swiss pairing engines. Dutch (C.04.3), Burstein (C.04.4.2), Dubov
// (C.04.4.1), and Lim (C.04.4.3) all build on this foundation, which
// mirrors the FIDE regulation structure where C.04.1 and C.04.2 define
// rules common to every Swiss system.
//
// # Testing strategy
//
// This package's own test suite reports approximately 31% statement
// coverage. That figure is misleading on its own. swisslib is shared
// infrastructure exercised primarily through its callers. When coverage
// is measured against the union of the Dutch, Burstein, Dubov, and Lim
// test suites, swisslib reaches roughly 94%:
//
//	go test -coverpkg=./pairing/swisslib -coverprofile=cover.out \
//	    ./pairing/dutch ./pairing/burstein ./pairing/dubov \
//	    ./pairing/lim ./pairing/swisslib
//	go tool cover -func=cover.out
//
// The remaining uncovered lines are debug helpers (Player.String) and
// rarely triggered fallbacks. New unit tests should target genuinely
// untested edge cases rather than chasing the per-package number.
package swisslib
