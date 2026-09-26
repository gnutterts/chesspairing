// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"bytes"
	"strings"
	"testing"
)

// auditPlayer is a literal 001 layout with one round entry at bytes 89-98.
func auditPlayer(result byte) string {
	line := bytes.Repeat([]byte(" "), 99)
	copy(line[0:3], "001")
	copy(line[4:8], "   1")
	copy(line[14:47], "Audit, Player")
	copy(line[80:84], " 0.0")
	copy(line[85:89], "   1")
	copy(line[91:95], "0000")
	line[96] = '-'
	line[98] = result
	return string(line)
}

func TestResultCodes2026(t *testing.T) {
	// TRF-2026, Player Section, result column: the quoted symbols are "1",
	// "0", "=", "+", "-", "W", "D", "L", "H", "F", "U", "Z"; "(blank)
	// equivalent to Z" and "Letter codes are case-insensitive". F/U are
	// excluded here: their meaning changes in a later PR (D22), so these
	// test cases are deliberately withheld for now.
	uppercase := map[byte]ResultCode{
		'1': ResultWin,
		'0': ResultLoss,
		'=': ResultDraw,
		'+': ResultForfeitWin,
		'-': ResultForfeitLoss,
		'W': ResultWinByDefault,
		'D': ResultDrawByDefault,
		'L': ResultLossByDefault,
		'H': ResultHalfBye,
		'Z': ResultZeroBye,
	}
	for code, want := range uppercase {
		t.Run("upper-"+string(code), func(t *testing.T) {
			doc, err := Read(strings.NewReader(auditPlayer(code) + "\n"))
			if err != nil {
				t.Fatalf("Read(%q): %v", code, err)
			}
			if got := doc.Players[0].Rounds[0].Result; got != want {
				t.Fatalf("Read(%q).Result = %v, want %v", code, got, want)
			}

			var out strings.Builder
			if err := Write(&out, doc); err != nil {
				t.Fatalf("Write(%q): %v", code, err)
			}
			doc2, err := Read(strings.NewReader(out.String()))
			if err != nil {
				t.Fatalf("re-Read(%q): %v", code, err)
			}
			if got := doc2.Players[0].Rounds[0].Result; got != want {
				t.Fatalf("re-Read(%q).Result = %v, want %v", code, got, want)
			}
		})
	}

	lowercase := map[byte]ResultCode{
		'w': ResultWinByDefault,
		'd': ResultDrawByDefault,
		'l': ResultLossByDefault,
		'h': ResultHalfBye,
		'z': ResultZeroBye,
	}
	for code, want := range lowercase {
		t.Run("lower-"+string(code), func(t *testing.T) {
			doc, err := Read(strings.NewReader(auditPlayer(code) + "\n"))
			if err != nil {
				t.Fatalf("Read(%q): %v", code, err)
			}
			if got := doc.Players[0].Rounds[0].Result; got != want {
				t.Fatalf("Read(%q).Result = %v, want %v", code, got, want)
			}
		})
	}

	// A blank result with opponent 0000 is equivalent to Z.
	t.Run("blank", func(t *testing.T) {
		doc, err := Read(strings.NewReader(auditPlayer(' ') + "\n"))
		if err != nil {
			t.Fatalf("Read(blank): %v", err)
		}
		r := doc.Players[0].Rounds[0]
		if r.Opponent != 0 || r.Color != ColorNone || r.Result != ResultZeroBye {
			t.Fatalf("Read(blank) = %+v, want {0 None ZeroBye}", r)
		}
	})
}

func TestRecords2026(t *testing.T) {
	// Literal TRF-2026 records: Read+Write must not drop any record.
	input := strings.Join([]string{
		"012 Audit", "022 City", "032 NED", "042 2026/01/01", "052 2026/01/02", "062 1", "072 1", "092 Swiss Dutch", "102 Chief", "112 Deputy", "122 90min", "132 26/01/01",
		"142 7", "152 B", "162 W 1.0    D 0.5    L 0.0", "172 NED FIDE", "182 controller", "192 FIDE_DUTCH_2025", "202 BH", "212 PTS,BH", "222 5400+30", "352 WBWB", "362 TW 2.0   TD 1.0   TL 0.0",
		auditPlayer('Z'),
		"NED    1      Audit, Player                   2000 NED             2000/01/01",
		"013    1 Audit Team                      1",
		"240 H 003 026 047", "250 00.0 02.0 001 003 0001 0090", "260 001 002 125 180 184 216", "299 +    2.0      2.5", "300 008 021 047 0058 0203 0105 0162",
		"310   1 India                            IND     2486   15.0   28.0 11                            1    5   15   28   44",
		"320 01.0 02.0 000 000 050 049 000 046 048 045 000 036 043", "330 +- 004 023 047", "330 -- 004 023 047",
		"801 03 GEO   19 32.5       FFFF       16 w 11=0 1254", "802   3 GEO     19.0   32.5          FPB    4.0         16 w 2.5",
		"XXR 7", "XXC B", "XXS acceleration", "XXP 1 2", "XXY 2", "XXB true", "XXM false", "XXT A", "XXG match", "XXA true", "XXK 1",
	}, "\n") + "\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read literal records: %v", err)
	}
	var out strings.Builder
	if err := Write(&out, doc); err != nil {
		t.Fatalf("Write literal records: %v", err)
	}
	output := out.String()

	codes := []string{
		"001", "013", "012", "022", "032", "042", "052", "062", "072", "092", "102", "112", "122", "132",
		"142", "152", "162", "172", "182", "192", "202", "212", "222", "352", "362",
		"240", "250", "260", "299", "300", "310", "320", "330", "801", "802",
		"XXR", "XXC", "XXS", "XXP", "XXY", "XXB", "XXM", "XXT", "XXG", "XXA", "XXK",
		"NED",
	}
	for _, code := range codes {
		if !strings.Contains(output, code+" ") {
			t.Errorf("record %s was lost in Read+Write", code)
		}
	}
}

func TestOracleRoundTrip(t *testing.T) {
	lf := strings.Join([]string{
		"012 Audit",
		"022 City",
		"092 Swiss Dutch",
		"142 3",
		"152 W",
		"XXR 3",
		"XXC B",
		auditPlayer('1'),
		auditPlayer('0'),
	}, "\n") + "\n"

	tests := []struct {
		name  string
		input string
	}{
		{"lf", lf},
		{"cr", strings.ReplaceAll(lf, "\n", "\r")},
		{"crlf", strings.ReplaceAll(lf, "\n", "\r\n")},
		{"noeol", strings.TrimSuffix(lf, "\n")},
	}

	var want string
	for i, tt := range tests {
		doc, err := Read(strings.NewReader(tt.input))
		if err != nil {
			t.Fatalf("Read(%s): %v", tt.name, err)
		}
		var out strings.Builder
		if err := Write(&out, doc); err != nil {
			t.Fatalf("Write(%s): %v", tt.name, err)
		}
		if i == 0 {
			want = out.String()
			continue
		}
		if out.String() != want {
			t.Errorf("%s round-trip output differs from LF-normalized output", tt.name)
		}
	}
}
