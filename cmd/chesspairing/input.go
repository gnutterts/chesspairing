// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"os"
)

// openInput opens the named file for reading. If name is "-", it returns
// os.Stdin wrapped in a no-op closer. Returns an error if name is empty.
func openInput(name string) (io.ReadCloser, error) {
	if name == "" {
		return nil, fmt.Errorf("no input file specified")
	}
	if name == "-" {
		return io.NopCloser(os.Stdin), nil
	}
	return os.Open(name)
}
