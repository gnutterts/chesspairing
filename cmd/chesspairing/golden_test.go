// Package main golden tests preserve the CLI's current byte-exact behavior.
// Set GOLDEN_UPDATE=1 and run the Golden tests to review and accept changes.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestGoldenCLI(t *testing.T) {
	for _, input := range cliTRFFixtures(t) {
		input := input
		base := sanitizeGoldenName(input)
		cases := []struct {
			name   string
			args   func(string) []string
			output bool
		}{
			{"pair", func(_ string) []string { return []string{"chesspairing", "pair", "--dutch", input, "--json"} }, false},
			{"pair_text", func(_ string) []string { return []string{"chesspairing", "pair", "--dutch", input} }, false},
			{"pair_wide", func(_ string) []string { return []string{"chesspairing", "pair", "--dutch", input, "--format", "wide"} }, false},
			{"pair_board", func(_ string) []string {
				return []string{"chesspairing", "pair", "--dutch", input, "--format", "board"}
			}, false},
			{"pair_xml", func(_ string) []string { return []string{"chesspairing", "pair", "--dutch", input, "--format", "xml"} }, false},
			{"standings", func(_ string) []string { return []string{"chesspairing", "standings", "--dutch", input, "--json"} }, false},
			{"standings_text", func(_ string) []string { return []string{"chesspairing", "standings", "--dutch", input} }, false},
			{"check", func(_ string) []string { return []string{"chesspairing", "check", "--dutch", input, "--json"} }, false},
			{"validate", func(_ string) []string { return []string{"chesspairing", "validate", input, "--json"} }, false},
			{"validate_text", func(_ string) []string { return []string{"chesspairing", "validate", input} }, false},
			{"convert", func(out string) []string { return []string{"chesspairing", "convert", input, "-o", out} }, true},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name+"_"+base, func(t *testing.T) {
				goldenCLIInvocation(t, tc.name+"_"+base, tc.args, tc.output, "")
			})
		}
	}

	// Exercise every selectable pairing system in both pairing and standings.
	// One stable fixture is sufficient here; every fixture receives Dutch coverage above.
	input := filepath.Join("testdata", "pair-input.trf")
	for _, system := range []string{"burstein", "dubov", "lim", "keizer", "roundrobin"} {
		system := system
		t.Run("system_"+system, func(t *testing.T) {
			goldenCLIInvocation(t, "system_pair_"+system, func(_ string) []string {
				return []string{"chesspairing", "pair", "--" + system, input}
			}, false, "")
			goldenCLIInvocation(t, "system_standings_"+system, func(_ string) []string {
				return []string{"chesspairing", "standings", "--" + system, input}
			}, false, "")
		})
	}

	t.Run("legacy_pair", func(t *testing.T) {
		goldenCLIInvocation(t, "legacy_pair", func(_ string) []string {
			return []string{"chesspairing", "--dutch", input, "-p"}
		}, false, "")
	})
	t.Run("legacy_check", func(t *testing.T) {
		goldenCLIInvocation(t, "legacy_check", func(_ string) []string {
			return []string{"chesspairing", "--dutch", input, "-c"}
		}, false, "")
	})
	t.Run("help", func(t *testing.T) {
		goldenCLIInvocation(t, "help", func(_ string) []string {
			return []string{"chesspairing", "--help"}
		}, false, "")
	})
	for _, tc := range []struct {
		name string
		args func(string) []string
	}{
		{"error_missing_file", func(_ string) []string {
			return []string{"chesspairing", "pair", "--dutch", "testdata/missing.trf"}
		}},
		{"error_unknown_flag", func(_ string) []string {
			return []string{"chesspairing", "pair", "--dutch", "--not-a-flag", input}
		}},
		{"error_convert_trf", func(out string) []string {
			return []string{"chesspairing", "convert", input, "-o", out, "--trf-format", "trf"}
		}},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			goldenCLIInvocation(t, tc.name, tc.args, false, "")
		})
	}

	t.Run("generate", func(t *testing.T) {
		dir := t.TempDir()
		cfg := filepath.Join(dir, "generate.cfg")
		if err := os.WriteFile(cfg, []byte("PlayersNumber=10\nRoundsNumber=3\nForfeitRate=4\nRetiredRate=0\nHalfPointByeRate=0\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		goldenCLIInvocation(t, "generate", func(out string) []string {
			return []string{"chesspairing", "generate", "--dutch", "--config", cfg, "-o", out, "-s", "golden-seed"}
		}, true, dir)
	})
	t.Run("legacy_generate", func(t *testing.T) {
		dir := t.TempDir()
		cfg := filepath.Join(dir, "generate.cfg")
		if err := os.WriteFile(cfg, []byte("PlayersNumber=10\nRoundsNumber=3\nForfeitRate=4\nRetiredRate=0\nHalfPointByeRate=0\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		goldenCLIInvocation(t, "legacy_generate", func(out string) []string {
			return []string{"chesspairing", "--dutch", "-g", cfg, "-o", out, "-s", "golden-seed"}
		}, true, dir)
	})
	t.Run("tiebreakers", func(t *testing.T) {
		goldenCLIInvocation(t, "tiebreakers", func(_ string) []string { return []string{"chesspairing", "tiebreakers", "--json"} }, false, "")
	})
}

func goldenCLIInvocation(t *testing.T, name string, args func(string) []string, hasOutput bool, extraTemp string) {
	t.Helper()
	dir := t.TempDir()
	outPath := filepath.Join(dir, "output.trf")
	var stdout, stderr bytes.Buffer
	argv := args(outPath)
	code := run(argv, &stdout, &stderr)
	normalize := func(s string) string {
		s = strings.ReplaceAll(s, dir, "<TMP>")
		if extraTemp != "" {
			s = strings.ReplaceAll(s, extraTemp, "<TMP>")
		}
		// Keep recorded paths and missing-file errors byte-identical on Unix and
		// Windows. argv itself remains native when passed to run above.
		s = filepath.ToSlash(s)
		return strings.NewReplacer(
			"no such file or directory", "file not found",
			"The system cannot find the file specified.", "file not found",
			"The system cannot find the path specified.", "file not found",
		).Replace(s)
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "ARGS: %s\nEXIT: %d\nSTDOUT:\n%sSTDERR:\n%s", normalize(strings.Join(argv[1:], " ")), code, normalize(stdout.String()), normalize(stderr.String()))
	if hasOutput {
		data, err := os.ReadFile(outPath)
		fmt.Fprintf(&b, "OUTPUT ERROR: %s\nOUTPUT BYTES:\n", normalize(fmt.Sprint(err)))
		if err == nil {
			b.Write(data)
		}
	}
	goldenCLI(t, filepath.Join("testdata", "golden", name+".txt"), b.Bytes())
}

func cliTRFFixtures(t *testing.T) []string {
	t.Helper()
	roots := []string{"../../trf/testdata", "../../pairing/dutch/testdata", "testdata"}
	var paths []string
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if path != root && filepath.Base(path) == "golden" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".trf" {
				// A .trf fixture is always significant, including malformed files:
				// its parser error is observable golden output.
				paths = append(paths, path)
				return nil
			}
			if ext != ".input" {
				// .input is the only non-TRF fixture format used for valid TRF input.
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if len(data) >= 3 && (string(data[:3]) == "001" || string(data[:3]) == "012") {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Strings(paths)
	return paths
}

func sanitizeGoldenName(path string) string {
	path = filepath.ToSlash(path)
	for strings.HasPrefix(path, "../") {
		path = strings.TrimPrefix(path, "../")
	}
	return strings.NewReplacer("/", "__", ".", "_").Replace(path)
}

func goldenCLI(t *testing.T, path string, got []byte) {
	t.Helper()
	if os.Getenv("GOLDEN_UPDATE") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("GOLDEN_UPDATE=1: overwrote %s; output was not compared", path)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden %s: %v (run with GOLDEN_UPDATE=1)", path, err)
	}
	if bytes.Equal(want, got) {
		return
	}
	t.Fatalf("golden mismatch: %s\n%s", path, compactCLIDiff(want, got))
}

func compactCLIDiff(want, got []byte) string {
	w, g := strings.Split(string(want), "\n"), strings.Split(string(got), "\n")
	n := min(len(w), len(g))
	line := n
	for i := 0; i < n; i++ {
		if w[i] != g[i] {
			line = i
			break
		}
	}
	var b strings.Builder
	for i := max(0, line-1); i < min(max(len(w), len(g)), line+3); i++ {
		if i < len(w) && i < len(g) && w[i] == g[i] {
			fmt.Fprintf(&b, " %d %s\n", i+1, w[i])
			continue
		}
		if i < len(w) {
			fmt.Fprintf(&b, "-%d %s\n", i+1, w[i])
		}
		if i < len(g) {
			fmt.Fprintf(&b, "+%d %s\n", i+1, g[i])
		}
	}
	return b.String()
}
