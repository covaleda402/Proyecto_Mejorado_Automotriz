package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the runtime configuration of the backend. Every value comes from
// the environment; no secret has a default, so a missing one stops the process
// instead of silently starting with a known key.
type Config struct {
	DatabaseDSN     string
	HTTPPort        string
	AllowedOrigin   string
	TokenSecret     string
	TokenTTL        time.Duration
	RequestTimeout  time.Duration
	DatabaseTimeout time.Duration
}

// Load reads the configuration and fails when a required variable is absent.
func Load() (Config, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("MYSQL_URL")
	}
	if dsn != "" {
		dsn = normalizeDSN(dsn)
	} else {
		host, err := required("MYSQL_HOST")
		if err != nil {
			return Config{}, err
		}
		port, err := required("MYSQL_PORT")
		if err != nil {
			return Config{}, err
		}
		database, err := required("MYSQL_DATABASE")
		if err != nil {
			return Config{}, err
		}
		user, err := required("MYSQL_USER")
		if err != nil {
			return Config{}, err
		}
		password, err := required("MYSQL_PASSWORD")
		if err != nil {
			return Config{}, err
		}
		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4&loc=UTC",
			user, password, net.JoinHostPort(host, port), database,
		)
	}

	secret, err := required("TOKEN_SECRET")
	if err != nil {
		return Config{}, err
	}
	origin := optional("ALLOWED_ORIGIN", "*")

	return Config{
		DatabaseDSN:     dsn,
		HTTPPort:        optional("PORT", optional("HTTP_PORT", "8080")),
		AllowedOrigin:   origin,
		TokenSecret:     secret,
		TokenTTL:        minuteDuration("TOKEN_TTL_MINUTE", 480),
		RequestTimeout:  secondDuration("REQUEST_TIMEOUT_SECOND", 15),
		DatabaseTimeout: secondDuration("DATABASE_TIMEOUT_SECOND", 5),
	}, nil
}

func required(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func optional(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func secondDuration(name string, fallback int) time.Duration {
	return time.Duration(positiveNumber(name, fallback)) * time.Second
}

func minuteDuration(name string, fallback int) time.Duration {
	return time.Duration(positiveNumber(name, fallback)) * time.Minute
}

func positiveNumber(name string, fallback int) int {
	parsed, err := strconv.Atoi(os.Getenv(name))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

// normalizeDSN converts standard cloud connection strings (mysql://user:pass@host:port/db)
// into the format expected by go-sql-driver/mysql (user:pass@tcp(host:port)/db?params).
func normalizeDSN(raw string) string {
	if !strings.HasPrefix(raw, "mysql://") {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	user := u.User.Username()
	password, _ := u.User.Password()
	host := u.Host
	dbName := strings.TrimPrefix(u.Path, "/")
	query := u.Query()
	if query.Get("parseTime") == "" {
		query.Set("parseTime", "true")
	}
	if query.Get("charset") == "" {
		query.Set("charset", "utf8mb4")
	}
	if query.Get("loc") == "" {
		query.Set("loc", "UTC")
	}
	if query.Get("tls") == "" {
		query.Set("tls", "true")
	}
	// TiDB Cloud uses ?ssl-mode=REQUIRED, but Go driver uses tls=true
	query.Del("ssl-mode")
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", user, password, host, dbName, query.Encode())
}
