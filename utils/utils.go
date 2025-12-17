package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// IsBotUA checks if the user agent indicates a bot
func IsBotUA(ua string) bool {
	return strings.Contains(strings.ToLower(ua), "bot")
}

// GetClientIP extracts the client IP from the request
func GetClientIP(r *http.Request) string {
	xfw := r.Header.Get("X-Forwarded-For")
	if xfw != "" {
		parts := strings.Split(xfw, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}
	ra := r.RemoteAddr
	if i := strings.LastIndex(ra, ":"); i > 0 {
		return ra[:i]
	}
	return ra
}

// SafeTargetURL validates and normalizes a target URL
func SafeTargetURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("invalid url")
	}
	if !u.IsAbs() || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("unsupported scheme")
	}
	return u.String(), nil
}

// ParseSize parses a size string like "300x250" into width and height
func ParseSize(tsize string) (int, int) {
	w, h := 300, 250
	s := strings.ToLower(strings.TrimSpace(tsize))
	if s == "" {
		return w, h
	}
	parts := strings.Split(s, "x")
	if len(parts) == 2 {
		if ww, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil && ww > 0 {
			w = ww
		}
		if hh, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil && hh > 0 {
			h = hh
		}
	}
	return w, h
}

// AtoiOrZero converts a string to int, returning 0 on error
func AtoiOrZero(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

// GetScheme determines the request scheme (http/https)
func GetScheme(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if xfp := r.Header.Get("X-Forwarded-Proto"); xfp != "" {
		scheme = strings.TrimSpace(strings.Split(xfp, ",")[0])
	}
	return scheme
}
