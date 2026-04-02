package version

import (
	"fmt"
	"strconv"
	"strings"
)

// Number and GitCommit are set at build time via -ldflags.
var (
	Number    = "dev"
	GitCommit = "unknown"
)

// Version represents a 3-segment semantic version.
type Version struct {
	Major, Minor, Patch int
}

// String returns a human-readable version string.
// Returns "dev" if no version was injected at build time.
func String() string {
	if Number == "" || Number == "dev" {
		return "dev"
	}
	return fmt.Sprintf("%s (%s)", Number, GitCommit)
}

// Parse parses a version string like "1.2.3" or "v1.2.3".
func Parse(s string) (Version, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version %q: expected 3 segments", s)
	}
	var v Version
	var err error
	if v.Major, err = strconv.Atoi(parts[0]); err != nil {
		return Version{}, fmt.Errorf("invalid major version: %w", err)
	}
	if v.Minor, err = strconv.Atoi(parts[1]); err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %w", err)
	}
	if v.Patch, err = strconv.Atoi(parts[2]); err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %w", err)
	}
	return v, nil
}

// Compare returns -1 if a < b, 0 if a == b, 1 if a > b.
func Compare(a, b Version) int {
	if a.Major != b.Major {
		if a.Major < b.Major {
			return -1
		}
		return 1
	}
	if a.Minor != b.Minor {
		if a.Minor < b.Minor {
			return -1
		}
		return 1
	}
	if a.Patch != b.Patch {
		if a.Patch < b.Patch {
			return -1
		}
		return 1
	}
	return 0
}
