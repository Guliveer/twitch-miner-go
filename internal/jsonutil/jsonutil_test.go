package jsonutil

import (
	"encoding/json"
	"testing"
)

func TestIntFromAny(t *testing.T) {
	cases := []struct {
		in   any
		want int
	}{
		{float64(42), 42},
		{float64(-7), -7},
		{int(10), 10},
		{int64(99), 99},
		{json.Number("123"), 123},
		{json.Number("-5"), -5},
		{"not a number", 0},
		{nil, 0},
		{true, 0},
	}
	for _, tc := range cases {
		if got := IntFromAny(tc.in); got != tc.want {
			t.Errorf("IntFromAny(%v) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestFloatFromAny(t *testing.T) {
	cases := []struct {
		in   any
		want float64
	}{
		{float64(3.14), 3.14},
		{int(5), 5.0},
		{int64(7), 7.0},
		{json.Number("2.5"), 2.5},
		{"text", 0},
		{nil, 0},
	}
	for _, tc := range cases {
		if got := FloatFromAny(tc.in); got != tc.want {
			t.Errorf("FloatFromAny(%v) = %f, want %f", tc.in, got, tc.want)
		}
	}
}

func TestStringFromAny(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{"hello", "hello"},
		{"", ""},
		{42, ""},
		{nil, ""},
		{true, ""},
	}
	for _, tc := range cases {
		if got := StringFromAny(tc.in); got != tc.want {
			t.Errorf("StringFromAny(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIntFromMap(t *testing.T) {
	m := map[string]any{"count": float64(7), "name": "foo"}
	if got := IntFromMap(m, "count"); got != 7 {
		t.Errorf("got %d, want 7", got)
	}
	if got := IntFromMap(m, "missing"); got != 0 {
		t.Errorf("missing key: got %d, want 0", got)
	}
	if got := IntFromMap(m, "name"); got != 0 {
		t.Errorf("string key: got %d, want 0", got)
	}
}

func TestStringFromMap(t *testing.T) {
	m := map[string]any{"key": "value", "num": float64(1)}
	if got := StringFromMap(m, "key"); got != "value" {
		t.Errorf("got %q, want %q", got, "value")
	}
	if got := StringFromMap(m, "missing"); got != "" {
		t.Errorf("missing key: got %q, want empty", got)
	}
	if got := StringFromMap(m, "num"); got != "" {
		t.Errorf("non-string key: got %q, want empty", got)
	}
}

func TestBoolFromMap(t *testing.T) {
	m := map[string]any{"on": true, "off": false, "text": "yes"}
	if !BoolFromMap(m, "on") {
		t.Error("expected true")
	}
	if BoolFromMap(m, "off") {
		t.Error("expected false")
	}
	if BoolFromMap(m, "missing") {
		t.Error("missing key: expected false")
	}
	if BoolFromMap(m, "text") {
		t.Error("non-bool value: expected false")
	}
}
