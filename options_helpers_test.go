// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package chesspairing_test

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestFloat64Ptr(t *testing.T) {
	v := 3.14
	p := chesspairing.Float64Ptr(v)
	if *p != v {
		t.Errorf("Float64Ptr(%f) = %f, want %f", v, *p, v)
	}
}

func TestIntPtr(t *testing.T) {
	v := 42
	p := chesspairing.IntPtr(v)
	if *p != v {
		t.Errorf("IntPtr(%d) = %d, want %d", v, *p, v)
	}
}

func TestBoolPtr(t *testing.T) {
	v := true
	p := chesspairing.BoolPtr(v)
	if *p != v {
		t.Errorf("BoolPtr(%t) = %t, want %t", v, *p, v)
	}
}

func TestStringPtr(t *testing.T) {
	v := "hello"
	p := chesspairing.StringPtr(v)
	if *p != v {
		t.Errorf("StringPtr(%q) = %q, want %q", v, *p, v)
	}
}

func TestGetFloat64(t *testing.T) {
	m := map[string]any{"a": 1.5, "b": 2, "c": int64(3), "d": "bad"}
	tests := []struct {
		key    string
		want   float64
		wantOK bool
	}{
		{"a", 1.5, true},
		{"b", 2.0, true},
		{"c", 3.0, true},
		{"d", 0, false},
		{"missing", 0, false},
	}
	for _, tt := range tests {
		got, ok := chesspairing.GetFloat64(m, tt.key)
		if ok != tt.wantOK || got != tt.want {
			t.Errorf("GetFloat64(%q) = (%f, %t), want (%f, %t)", tt.key, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestGetInt(t *testing.T) {
	m := map[string]any{"a": 5, "b": int64(6), "c": 7.0, "d": "bad"}
	tests := []struct {
		key    string
		want   int
		wantOK bool
	}{
		{"a", 5, true},
		{"b", 6, true},
		{"c", 7, true},
		{"d", 0, false},
		{"missing", 0, false},
	}
	for _, tt := range tests {
		got, ok := chesspairing.GetInt(m, tt.key)
		if ok != tt.wantOK || got != tt.want {
			t.Errorf("GetInt(%q) = (%d, %t), want (%d, %t)", tt.key, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestGetBool(t *testing.T) {
	m := map[string]any{"a": true, "b": false, "c": "bad"}
	tests := []struct {
		key    string
		want   bool
		wantOK bool
	}{
		{"a", true, true},
		{"b", false, true},
		{"c", false, false},
		{"missing", false, false},
	}
	for _, tt := range tests {
		got, ok := chesspairing.GetBool(m, tt.key)
		if ok != tt.wantOK || got != tt.want {
			t.Errorf("GetBool(%q) = (%t, %t), want (%t, %t)", tt.key, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestGetString(t *testing.T) {
	m := map[string]any{"a": "hello", "b": "", "c": 42}
	tests := []struct {
		key    string
		want   string
		wantOK bool
	}{
		{"a", "hello", true},
		{"b", "", true},
		{"c", "", false},
		{"missing", "", false},
	}
	for _, tt := range tests {
		got, ok := chesspairing.GetString(m, tt.key)
		if ok != tt.wantOK || got != tt.want {
			t.Errorf("GetString(%q) = (%q, %t), want (%q, %t)", tt.key, got, ok, tt.want, tt.wantOK)
		}
	}
}
