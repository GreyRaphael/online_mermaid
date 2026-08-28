package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	Version   = "v1.0.2"
	GitCommit = "dev"
	BuildDate = "unknown"
)

func init() {
	if GitCommit == "dev" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			for _, s := range bi.Settings {
				if s.Key == "vcs.revision" && len(s.Value) >= 7 {
					GitCommit = s.Value[:7]
				}
				if s.Key == "vcs.time" {
					BuildDate = s.Value
				}
			}
		}
	}
}

func FormattedVersion() string {
	return fmt.Sprintf("online-mermaid version %s (commit: %s, built at: %s)", Version, GitCommit, BuildDate)
}

type Config struct {
	Addr         string
	DBPath       string
	Username     string
	PasswordHash []byte
	SessionTTL   time.Duration
	SecureCookie bool
}

var ErrVersionRequested = errors.New("version requested")

func Parse(args []string) (Config, error) {
	fs := flag.NewFlagSet("online-mermaid", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var cfg Config
	var sessionTTL string
	var showVersion bool
	fs.BoolVar(&showVersion, "v", false, "print version and exit")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")
	fs.StringVar(&cfg.Addr, "addr", envOr("ONLINE_MERMAID_ADDR", "0.0.0.0:8850"), "HTTP listen address")
	fs.StringVar(&cfg.DBPath, "db-path", envOr("ONLINE_MERMAID_DB", ""), "SQLite database file path")
	fs.StringVar(&cfg.Username, "admin-user", envOr("ONLINE_MERMAID_ADMIN_USERNAME", "admin"), "administrator username")
	passwordHash := os.Getenv("ONLINE_MERMAID_ADMIN_PASSWORD_HASH")
	fs.StringVar(&passwordHash, "password-hash", passwordHash, "administrator bcrypt password hash")
	fs.StringVar(&sessionTTL, "session-ttl", envOr("ONLINE_MERMAID_SESSION_TTL", "72h"), "session lifetime")
	fs.BoolVar(&cfg.SecureCookie, "secure-cookie", envBool("ONLINE_MERMAID_SECURE_COOKIE", false), "mark session cookie Secure")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if showVersion {
		return Config{}, ErrVersionRequested
	}

	cfg.Username = strings.TrimSpace(cfg.Username)
	if cfg.Username == "" {
		return Config{}, errors.New("admin username cannot be empty")
	}

	// If no password hash is provided, use default password "admin123" for quick start
	if passwordHash == "" {
		defaultHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return Config{}, fmt.Errorf("generate default password hash: %w", err)
		}
		passwordHash = string(defaultHash)
	}

	if _, err := bcrypt.Cost([]byte(passwordHash)); err != nil {
		return Config{}, fmt.Errorf("invalid bcrypt password hash: %w", err)
	}
	cfg.PasswordHash = []byte(passwordHash)

	var err error
	cfg.SessionTTL, err = time.ParseDuration(sessionTTL)
	if err != nil || cfg.SessionTTL <= 0 {
		return Config{}, fmt.Errorf("invalid session TTL %q", sessionTTL)
	}

	if cfg.DBPath == "" {
		cfg.DBPath = defaultDBPath()
	}
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
		return Config{}, fmt.Errorf("create database directory: %w", err)
	}

	return cfg, nil
}

func defaultDBPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, hErr := os.UserHomeDir()
		if hErr != nil {
			return filepath.Join(".", "data", "mermaid.db")
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "online-mermaid", "mermaid.db")
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
