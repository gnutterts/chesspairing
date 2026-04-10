// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"bytes"
	"os"
	"testing"
)

func FuzzRead(f *testing.F) {
	// Seed corpus with valid TRF fragments.
	f.Add([]byte("012 Test Tournament\n"))
	f.Add([]byte("001    1  GM Kasparov, Garry                   2850 RUS  4100018    1963/04/13  1.0    1  0002 w 1\n"))
	f.Add([]byte("XXR 5\nXXC white1\nXXP 1 2\n"))
	f.Add([]byte("013    1Chess Club Amsterdam                    1   2   3   4\n"))
	f.Add([]byte("XXB true\nXXM false\nXXA true\n"))
	f.Add([]byte("XYZ Unknown line content\n"))

	// Seed with testdata/basic.trf if it exists.
	if data, err := os.ReadFile("testdata/basic.trf"); err == nil {
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		doc, _ := Read(bytes.NewReader(data))
		if doc != nil {
			var buf bytes.Buffer
			_ = Write(&buf, doc)
		}
	})
}
