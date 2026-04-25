package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type Mode string

const (
	ModeLocal  Mode = "local"
	ModeRemote Mode = "remote"
	ModeBoth  Mode = "both"
	ModeMock  Mode = "mock"
)

type Config struct {
	ListenAddr      string
	Verbose         bool
	AuthMode        Mode
	AuthDBPath      string
	AuthAPIURL      string
	AuthAPIKey      string
	AccountingMode Mode
	AccountingDBPath string
	AccountingAPIURL string
	AccountingAPIKey string
	AccountingInterval time.Duration
	MockAPI         bool
}

var cfg *Config

func Parse() *Config {
	if cfg != nil {
		return cfg
	}

	listenAddr := flag.String("addr", ":8080", "proxy listen address")
	verbose := flag.Bool("v", false, "log all proxy requests")
	authMode := flag.String("auth-mode", "local", "auth mode: local, remote, or mock")
	authDBPath := flag.String("auth-db-path", "./users.db", "path to auth SQLite database")
	authAPIURL := flag.String("auth-api-url", "", "remote auth API URL")
	authAPIKey := flag.String("auth-api-key", "", "API key for remote auth")
	accountingMode := flag.String("accounting-mode", "local", "accounting mode: local, remote, both, or mock")
	accountingDBPath := flag.String("accounting-db-path", "./traffic.db", "path to traffic SQLite database")
	accountingAPIURL := flag.String("accounting-api-url", "", "remote accounting API URL")
	accountingAPIKey := flag.String("accounting-api-key", "", "API key for remote accounting")
	accountingInterval := flag.Duration("accounting-interval", 60*time.Second, "interval for periodic accounting updates")
	mockAPI := flag.Bool("mock-api", false, "mock API calls (log instead of making real requests)")

	flag.Parse()

	cfg = &Config{
		ListenAddr:          *listenAddr,
		Verbose:             *verbose,
		AuthMode:            Mode(*authMode),
		AuthDBPath:          *authDBPath,
		AuthAPIURL:          *authAPIURL,
		AuthAPIKey:          *authAPIKey,
		AccountingMode:     Mode(*accountingMode),
		AccountingDBPath:    *accountingDBPath,
		AccountingAPIURL:   *accountingAPIURL,
		AccountingAPIKey:   *accountingAPIKey,
		AccountingInterval:  *accountingInterval,
		MockAPI:             *mockAPI,
	}

	if err := cfg.validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return cfg
}

func (c *Config) validate() error {
	if c.AuthMode == ModeRemote && c.AuthAPIURL == "" {
		return fmt.Errorf("auth-mode=remote requires auth-api-url")
	}
	if c.AuthMode == ModeLocal {
		if c.AuthDBPath == "" {
			return fmt.Errorf("auth-mode=local requires auth-db-path")
		}
	}
	if c.AccountingMode == ModeRemote && c.AccountingAPIURL == "" {
		return fmt.Errorf("accounting-mode=remote requires accounting-api-url")
	}
	if c.AccountingMode == ModeLocal || c.AccountingMode == ModeBoth {
		if c.AccountingDBPath == "" {
			return fmt.Errorf("accounting-mode=local or both requires accounting-db-path")
		}
	}
	return nil
}