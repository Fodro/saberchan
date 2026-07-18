package ban

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// presetAliases maps human-friendly day presets to their Go duration string.
var presetAliases = map[string]string{
	"1d":  "24h",
	"7d":  "168h",
	"30d": "720h",
}

// ParseDuration parses a ban duration preset accepted from admin requests:
// a Go duration string (e.g. "1h", "24h", "168h"), a day-shorthand alias
// ("1d", "7d", "30d"), a plain number of seconds ("3600"), or "permanent".
//
// It returns 0 for "permanent", which Service.Ban/BanPost treat as a
// permanent ban.
func ParseDuration(s string) (time.Duration, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return 0, fmt.Errorf("duration is required")
	}
	if s == "permanent" {
		return 0, nil
	}
	if alias, ok := presetAliases[s]; ok {
		s = alias
	}

	if d, err := time.ParseDuration(s); err == nil {
		if d <= 0 {
			return 0, fmt.Errorf("duration must be positive")
		}
		return d, nil
	}

	if secs, err := strconv.ParseInt(s, 10, 64); err == nil {
		if secs <= 0 {
			return 0, fmt.Errorf("duration must be positive")
		}
		return time.Duration(secs) * time.Second, nil
	}

	return 0, fmt.Errorf("invalid duration %q", s)
}
