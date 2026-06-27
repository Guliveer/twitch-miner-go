package utils

import (
	"math"
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Hello World", "hello-world"},
		{"Tom Clancy's Rainbow Six Siege", "tom-clancys-rainbow-six-siege"},
		{"PUBG: BATTLEGROUNDS", "pubg-battlegrounds"},
		{"Dota 2", "dota-2"},
		{"---leading-trailing---", "leading-trailing"},
		{"multiple   spaces", "multiple-spaces"},
		{"special!@#$%chars", "special-chars"},
		{"", ""},
		{"already-slugified", "already-slugified"},
	}
	for _, tc := range cases {
		if got := Slugify(tc.in); got != tc.want {
			t.Errorf("Slugify(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestMillify(t *testing.T) {
	cases := []struct {
		n         int
		precision int
		want      string
	}{
		{0, 2, "0"},
		{999, 2, "999"},
		{1000, 2, "1K"},
		{1500, 2, "1.5K"},
		{1234, 0, "1K"},
		{1_000_000, 2, "1M"},
		{1_500_000, 1, "1.5M"},
		{1_000_000_000, 2, "1B"},
		{-5000, 2, "-5K"},
		{50_000, 2, "50K"},
		{100, -1, "100"}, // negative precision falls back to 2, still < 1K
	}
	for _, tc := range cases {
		if got := Millify(tc.n, tc.precision); got != tc.want {
			t.Errorf("Millify(%d, %d) = %q, want %q", tc.n, tc.precision, got, tc.want)
		}
	}
}

func TestPercentage(t *testing.T) {
	cases := []struct {
		a, b int
		want int
	}{
		{50, 100, 50},
		{1, 4, 25},
		{0, 100, 0},
		{100, 0, 0},
		{0, 0, 0},
		{100, 100, 100},
		{3, 10, 30},
	}
	for _, tc := range cases {
		if got := Percentage(tc.a, tc.b); got != tc.want {
			t.Errorf("Percentage(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestFloatRound(t *testing.T) {
	cases := []struct {
		n      float64
		digits int
		want   float64
	}{
		{3.14159, 2, 3.14},
		{3.145, 2, 3.15},
		{1.0, 0, 1.0},
		{1.5, 0, 2.0},
		{-2.555, 2, -2.56},
		{0, 4, 0},
		{100.0, 2, 100.0},
	}
	for _, tc := range cases {
		got := FloatRound(tc.n, tc.digits)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("FloatRound(%v, %d) = %v, want %v", tc.n, tc.digits, got, tc.want)
		}
	}
}

func TestSafeGoRecoversPanic(t *testing.T) {
	done := make(chan struct{})
	SafeGo(func() {
		defer close(done)
		panic("test panic — should be recovered")
	})
	<-done // will hang forever if SafeGo doesn't recover
}
