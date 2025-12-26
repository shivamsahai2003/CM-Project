package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func IsBotUA(ua string) bool {
	return strings.Contains(strings.ToLower(ua), "bot")
}

func GetClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	if i := strings.LastIndex(r.RemoteAddr, ":"); i > 0 {
		return r.RemoteAddr[:i]
	}
	return r.RemoteAddr
}

func SafeTargetURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !u.IsAbs() || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("invalid url")
	}
	return u.String(), nil
}

func ParseSize(tsize string) (int, int) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(tsize)), "x")
	w, h := 300, 250
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

func AtoiOrZero(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func GetScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if xfp := r.Header.Get("X-Forwarded-Proto"); xfp != "" {
		return strings.TrimSpace(strings.Split(xfp, ",")[0])
	}
	return "http"
}

func CountKeywordSlots(templatePath string) int {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return 0
	}
	return len(regexp.MustCompile(`class="keyword-item"`).FindAllString(string(content), -1))
}

func CountAdSlots(templatePath string) int {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return 0
	}
	matches := regexp.MustCompile(`\{\{\.AdHref\d+\}\}`).FindAllString(string(content), -1)
	unique := make(map[string]struct{})
	for _, m := range matches {
		unique[m] = struct{}{}
	}
	return len(unique)
}
