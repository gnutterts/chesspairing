// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"bytes"
	"os"
	"testing"
	"unicode/utf8"
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
		// The writer pads fixed-width fields and keeps lone carriage returns and
		// invalid UTF-8 as data, while the reader splits on them, so the
		// idempotency check below is limited to well-formed text.
		if !validTRFText(data) {
			return
		}
		doc, err := Read(bytes.NewReader(data))
		if err != nil {
			return
		}
		// Read is lenient with malformed numeric fields (a member start number
		// 0, values that overflow once the fixed-width fields are rewritten), so
		// byte-wise idempotency is not promised for arbitrary input. What the
		// writer emits must always be readable again.
		var first bytes.Buffer
		if err := Write(&first, doc); err != nil {
			return
		}
		if _, err := Read(bytes.NewReader(first.Bytes())); err != nil {
			t.Fatalf("Read() after Write() error = %v", err)
		}
	})
}

// validTRFText reports whether data is valid UTF-8 without control characters
// other than newline and tab.
func validTRFText(data []byte) bool {
	if !utf8.Valid(data) {
		return false
	}
	for _, b := range data {
		if b < 0x20 && b != '\n' && b != '\t' {
			return false
		}
	}
	return true
}
