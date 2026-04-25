package accounting

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	_ "github.com/glebarez/go-sqlite"

	"vn-socks-proxy/internal/config"
)

var (
	ErrInvalidPayload = errors.New("invalid payload")
)

type TrafficRecorder interface {
	Connect(username, target string) error
	Update(username, target string, bytesSent, bytesRecv uint64, duration time.Duration) error
	Disconnect(username, target string, bytesSent, bytesRecv uint64, duration time.Duration) error
	Close() error
}

type Connection struct {
	ID           int64
	Username    string
	Target      string
	BytesSent  uint64
	BytesRecv  uint64
	StartTime  time.Time
	EndTime    sql.NullTime
}

type LocalRecorder struct {
	db       *sql.DB
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func NewLocalRecorder(dbPath string, interval time.Duration) (*LocalRecorder, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open traffic database: %w", err)
	}
	if err := initTrafficSchema(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}
	r := &LocalRecorder{
		db:       conn,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
	return r, nil
}

func initTrafficSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS connections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		target TEXT NOT NULL,
		bytes_sent INTEGER DEFAULT 0,
		bytes_recv INTEGER DEFAULT 0,
		start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		end_time TIMESTAMP,
		duration_seconds INTEGER
	);
	CREATE INDEX IF NOT EXISTS idx_connections_username ON connections(username);
	CREATE INDEX IF NOT EXISTS idx_connections_start ON connections(start_time);
	`
	_, err := db.Exec(schema)
	return err
}

func (r *LocalRecorder) Connect(username, target string) error {
	_, err := r.db.Exec(
		"INSERT INTO connections (username, target, start_time) VALUES (?, ?, ?)",
		username, target, time.Now(),
	)
	return err
}

func (r *LocalRecorder) Update(username, target string, bytesSent, bytesRecv uint64, duration time.Duration) error {
	_, err := r.db.Exec(
		`UPDATE connections 
		SET bytes_sent = ?, bytes_recv = ?, duration_seconds = ?
		WHERE username = ? AND target = ? AND end_time IS NULL`,
		bytesSent, bytesRecv, int(duration.Seconds()), username, target,
	)
	return err
}

func (r *LocalRecorder) Disconnect(username, target string, bytesSent, bytesRecv uint64, duration time.Duration) error {
	_, err := r.db.Exec(
		`UPDATE connections 
		SET bytes_sent = ?, bytes_recv = ?, end_time = ?, duration_seconds = ?
		WHERE username = ? AND target = ? AND end_time IS NULL`,
		bytesSent, bytesRecv, time.Now(), int(duration.Seconds()), username, target,
	)
	return err
}

func (r *LocalRecorder) Close() error {
	close(r.stopCh)
	r.wg.Wait()
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

type RemoteRecorder struct {
	apiURL   string
	apiKey   string
	client   *http.Client
	mock     bool
}

func NewRemoteRecorder(apiURL, apiKey string, mock bool) *RemoteRecorder {
	return &RemoteRecorder{
		apiURL: apiURL,
		apiKey: apiKey,
		mock:  mock,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type connectPayload struct {
	Username string `json:"username"`
	Target  string `json:"target"`
}

type disconnectPayload struct {
	Username        string `json:"username"`
	Target         string `json:"target"`
	BytesSent      uint64 `json:"bytes_sent"`
	BytesRecv      uint64 `json:"bytes_recv"`
	DurationSeconds int64  `json:"duration_seconds"`
}

type updatePayload struct {
	Username        string `json:"username"`
	Target         string `json:"target"`
	BytesSent      uint64 `json:"bytes_sent"`
	BytesRecv      uint64 `json:"bytes_recv"`
	DurationSeconds int64  `json:"duration_seconds"`
}

func (r *RemoteRecorder) post(endpoint string, payload interface{}) error {
	if r.mock {
		data, _ := json.Marshal(payload)
		log.Printf("[MOCK] POST %s %s", endpoint, string(data))
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, r.apiURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if r.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+r.apiKey)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (r *RemoteRecorder) Connect(username, target string) error {
	return r.post("/api/connect", connectPayload{
		Username: username,
		Target:  target,
	})
}

func (r *RemoteRecorder) Update(username, target string, bytesSent, bytesRecv uint64, duration time.Duration) error {
	return r.post("/api/update", updatePayload{
		Username:         username,
		Target:          target,
		BytesSent:       bytesSent,
		BytesRecv:       bytesRecv,
		DurationSeconds: int64(duration.Seconds()),
	})
}

func (r *RemoteRecorder) Disconnect(username, target string, bytesSent, bytesRecv uint64, duration time.Duration) error {
	return r.post("/api/disconnect", disconnectPayload{
		Username:         username,
		Target:          target,
		BytesSent:       bytesSent,
		BytesRecv:       bytesRecv,
		DurationSeconds: int64(duration.Seconds()),
	})
}

func (r *RemoteRecorder) Close() error {
	return nil
}

func NewRecorder(cfg *config.Config) (TrafficRecorder, error) {
	switch cfg.AccountingMode {
	case config.ModeLocal:
		return NewLocalRecorder(cfg.AccountingDBPath, cfg.AccountingInterval)
	case config.ModeRemote:
		return NewRemoteRecorder(cfg.AccountingAPIURL, cfg.AccountingAPIKey, cfg.MockAPI), nil
	default:
		return nil, fmt.Errorf("unsupported accounting mode: %s", cfg.AccountingMode)
	}
}