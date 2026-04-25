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