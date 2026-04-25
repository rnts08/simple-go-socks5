package auth

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type DB struct {
	conn *sql.DB
}

func OpenDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := initSchema(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return &DB{conn: conn}, nil
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	`
	_, err := db.Exec(schema)
	return err
}

func (d *DB) GetPasswordHash(username string) (string, error) {
	var hash string
	err := d.conn.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	return hash, nil
}

func (d *DB) GetUser(username string) (*user, error) {
	var u user
	err := d.conn.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (d *DB) CreateUser(username, passwordHash string) (int64, error) {
	result, err := d.conn.Exec(
		"INSERT INTO users (username, password_hash, created_at) VALUES (?, ?, ?)",
		username, passwordHash, time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (d *DB) DeleteUser(username string) error {
	_, err := d.conn.Exec("DELETE FROM users WHERE username = ?", username)
	return err
}

func (d *DB) Close() error {
	return d.conn.Close()
}