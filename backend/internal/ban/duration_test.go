package ban

import (
	"testing"
	"time"
)

func TestParseDuration_Presets(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"1h", time.Hour},
		{"24h", 24 * time.Hour},
		{"1d", 24 * time.Hour},
		{"168h", 168 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
		{"720h", 720 * time.Hour},
		{"30d", 30 * 24 * time.Hour},
		{"3600", time.Hour},
		{"permanent", 0},
		{"PERMANENT", 0},
		{" 1d ", 24 * time.Hour},
	}
	for _, c := range cases {
		got, err := ParseDuration(c.in)
		if err != nil {
			t.Fatalf("ParseDuration(%q) unexpected err: %v", c.in, err)
		}
		if got != c.want {
			t.Fatalf("ParseDuration(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseDuration_Invalid(t *testing.T) {
	t.Parallel()
	for _, in := range []string{"", "abc", "-1h", "0h", "-5"} {
		if _, err := ParseDuration(in); err == nil {
			t.Fatalf("ParseDuration(%q) expected error, got nil", in)
		}
	}
}
