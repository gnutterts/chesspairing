// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

//go:build !race

package dubov

// raceEnabled reports whether the tests run under the race detector, which
// slows the search too much for the per-round time limits.
const raceEnabled = false
