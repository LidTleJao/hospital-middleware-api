// Package config loads and validates the settings the service needs at boot.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds every setting the hospital middleware API needs to run.
type Config struct {
	Port        string
	GinMode     string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
	HISBaseURL  string
	HISTimeout  time.Duration
}

// Load reads configuration from the environment. It reports every missing
// required variable at once rather than failing on the first one, so a
// misconfigured deployment can be fixed in a single pass.
func Load() (Config, error) {
	var missing []string

	databaseURL := required("DATABASE_URL", &missing)
	JwtSecret := required("JWT_SECRET", &missing)

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	jwtTTL, err := duration("JWT_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	hisTimeout, err := duration("HIS_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:        optional("APP_PORT", "8080"),
		GinMode:     optional("GIN_MODE", "release"),
		HISBaseURL:  optional("HIS_BASE_URL", "https://hospital-a.api.co.th"),
		DatabaseURL: databaseURL,
		JWTSecret:   JwtSecret,
		JWTTTL:      jwtTTL,
		HISTimeout:  hisTimeout,
	}, nil
}

// required returns the value of key, appending key to missing when it is unset.
func required(key string, missing *[]string) string {
	value := os.Getenv(key)
	if value == "" {
		*missing = append(*missing, key)
	}
	return value
}

// optional returns the value of key, or fallback when it is unset.
func optional(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// duration parses key as a Go duration such as "30s" or "24h", falling back
// when it is unset.
func duration(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", key, raw, err)
	}
	return parsed, nil
}
