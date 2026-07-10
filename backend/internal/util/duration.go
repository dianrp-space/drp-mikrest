package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseRouterOSDuration memparse format duration RouterOS ke time.Duration.
// Format yang didukung: "1h", "30m", "1d", "1d12h30m", "1w", "1w2d", dst.
// RouterOS pakai kombinasi: w (minggu), d (hari), h (jam), m (menit), s (detik).
// Contoh valid: "30m", "1h", "24h", "1d", "7d", "1w", "1d12h", "1h30m".
func ParseRouterOSDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("duration kosong")
	}
	var total time.Duration
	i := 0
	for i < len(s) {
		// baca angka
		start := i
		for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '+') {
			i++
		}
		if start == i {
			return 0, fmt.Errorf("expected digit di posisi %d: %q", i, s)
		}
		numStr := s[start:i]
		n, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid number %q: %w", numStr, err)
		}
		if i >= len(s) {
			return 0, fmt.Errorf("expected unit setelah angka di posisi %d: %q", i, s)
		}
		unit := s[i]
		i++
		switch unit {
		case 'w', 'W':
			total += time.Duration(n) * 7 * 24 * time.Hour
		case 'd', 'D':
			total += time.Duration(n) * 24 * time.Hour
		case 'h', 'H':
			total += time.Duration(n) * time.Hour
		case 'm', 'M':
			total += time.Duration(n) * time.Minute
		case 's', 'S':
			total += time.Duration(n) * time.Second
		default:
			return 0, fmt.Errorf("unit tidak dikenal %q di posisi %d", string(unit), i-1)
		}
	}
	if total <= 0 {
		return 0, fmt.Errorf("duration harus > 0: %q", s)
	}
	return total, nil
}

// ParseRouterOSUptime memparse uptime dari RouterOS yang bisa dalam format:
//   - duration: "1d12h30m", "30m", "1h"
//   - HH:MM:SS: "01:30:00", "00:00:00"
func ParseRouterOSUptime(s string) (time.Duration, error) {
	if s == "" || s == "00:00:00" {
		return 0, nil
	}
	// coba parse sebagai RouterOS duration dulu
	if d, err := ParseRouterOSDuration(s); err == nil {
		return d, nil
	}
	// fallback: HH:MM:SS
	parts := strings.Split(s, ":")
	if len(parts) == 3 {
		h, _ := strconv.Atoi(parts[0])
		m, _ := strconv.Atoi(parts[1])
		sec, _ := strconv.Atoi(parts[2])
		return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second, nil
	}
	return 0, fmt.Errorf("tidak bisa parse uptime: %q", s)
}
