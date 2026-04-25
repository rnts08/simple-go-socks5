package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"vn-socks-proxy/internal/config"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAuthFailed        = errors.New("authentication failed")
)

type Authenticator interface {
	Validate(username, password string) (bool, error)
	Close() error
}

type user struct {
	ID           int64
	Username    string
	PasswordHash string
	CreatedAt   time.Time
}

type authResponse struct {
	Valid  bool   `json:"valid"`
	Error  string `json:"error,omitempty"`
}

type LocalAuthenticator struct {
	db *DB
}

func NewLocalAuthenticator(dbPath string) (*LocalAuthenticator, error) {
	db, err := OpenDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open auth database: %w", err)
	}
	return &LocalAuthenticator{db: db}, nil
}

func (a *LocalAuthenticator) Validate(username, password string) (bool, error) {
	if username == "" || password == "" {
		return false, ErrInvalidCredentials
	}
	hashedPassword, err := a.db.GetPasswordHash(username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return false, nil
		}
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil, nil
}

func (a *LocalAuthenticator) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

type RemoteAuthenticator struct {
	apiURL string
	apiKey string
	client *http.Client
	mock   bool
}

func NewRemoteAuthenticator(apiURL, apiKey string, mock bool) *RemoteAuthenticator {
	return &RemoteAuthenticator{
		apiURL: apiURL,
		apiKey: apiKey,
		mock:  mock,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (a *RemoteAuthenticator) Validate(username, password string) (bool, error) {
	if username == "" || password == "" {
		return false, ErrInvalidCredentials
	}

	if a.mock {
		log.Printf("[MOCK] Auth validating user=%s (password hidden)", username)
		if username == "admin" && password == "test" {
			log.Printf("[MOCK] Auth success for user=%s", username)
			return true, nil
		}
		log.Printf("[MOCK] Auth failed for user=%s", username)
		return false, nil
	}

	payload := map[string]string{
		"user":     username,
		"password": password,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.apiURL+"/api/login", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if a.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.apiKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusForbidden:
		return false, nil
	default:
		respBody, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("auth request returned unexpected status %d: %s", resp.StatusCode, string(respBody))
	}
}

func (a *RemoteAuthenticator) Close() error {
	return nil
}

type MockAuthenticator struct{}

func NewMockAuthenticator() *MockAuthenticator {
	return &MockAuthenticator{}
}

func (a *MockAuthenticator) Validate(username, password string) (bool, error) {
	log.Printf("[MOCK] Auth validation bypassed for user=%s", username)
	if username == "admin" && password == "test" {
		return true, nil
	}
	return false, nil
}

func (a *MockAuthenticator) Close() error {
	return nil
}

func NewAuthenticator(cfg *config.Config) (Authenticator, error) {
	switch cfg.AuthMode {
	case config.ModeLocal:
		return NewLocalAuthenticator(cfg.AuthDBPath)
	case config.ModeRemote:
		return NewRemoteAuthenticator(cfg.AuthAPIURL, cfg.AuthAPIKey, cfg.MockAPI), nil
	case config.ModeMock:
		return NewMockAuthenticator(), nil
	default:
		return nil, fmt.Errorf("unsupported auth mode: %s", cfg.AuthMode)
	}
}