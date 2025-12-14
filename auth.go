package roe

import (
	"net/http"
	"strings"
)

const userAgent = "roe-golang/0.1.0"

// Auth handles header generation.
type Auth struct {
	cfg Config
}

func newAuth(cfg Config) Auth {
	return Auth{cfg: cfg}
}

// Headers returns default headers including auth.
func (a Auth) Headers() http.Header {
	h := http.Header{}
	// Strip "Bearer " prefix if user accidentally included it
	key := a.cfg.APIKey
	if strings.HasPrefix(strings.ToLower(key), "bearer ") {
		key = strings.TrimSpace(key[7:])
	}
	h.Set("Authorization", "Bearer "+key)
	h.Set("User-Agent", userAgent)
	return h
}
