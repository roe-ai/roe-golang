package roe

import (
	"net/url"
	"os"
	"regexp"
	"strings"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{12}$`)

func isUUIDString(v any) bool {
	str, ok := v.(string)
	if !ok {
		return false
	}
	return uuidPattern.MatchString(str)
}

func isFilePath(val string) bool {
	if val == "" || isUUIDString(val) {
		return false
	}
	info, err := os.Stat(val)
	return err == nil && !info.IsDir()
}

func looksLikePath(val string) bool {
	return strings.Contains(val, "/") || strings.Contains(val, "\\") || strings.HasPrefix(val, ".")
}

func isHTTPURL(val string) bool {
	parsed, err := url.Parse(val)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return parsed.Host != ""
}
