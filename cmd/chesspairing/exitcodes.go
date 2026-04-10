// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/exitcodes.go
package main

const (
	ExitSuccess      = 0 // Operation completed successfully
	ExitNoPairing    = 1 // No valid pairing could be produced
	ExitUnexpected   = 2 // Unexpected runtime error
	ExitInvalidInput = 3 // Invalid or malformed input file
	ExitSizeOverflow = 4 // Tournament size exceeds implementation limits
	ExitFileAccess   = 5 // File could not be opened, read, or written
)
