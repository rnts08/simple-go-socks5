# SOCKS5 Proxy

A SOCKS5 proxy server written in Go with authentication and traffic accounting.

## Features

- **SOCKS5 Protocol**: RFC 1928 compliant CONNECT command support
- **Authentication**: Local SQLite database or remote API validation
- **Traffic Accounting**: Local SQLite logging or remote API reporting
- **Flexible Modes**: Run in local-only, remote-only, or both modes
- **Mock Mode**: Test without real backends by enabling mock API

## Usage

```bash
./socks5 [options]
```

### Options

| Flag | Default | Description |
|------|---------|------------|
| `-addr` | `:8080` | Proxy listen address |
| `-v` | `false` | Log all proxy requests |
| `-auth-mode` | `local` | Auth mode: `local`, `remote`, or `mock` |
| `-auth-db-path` | `./users.db` | Path to auth SQLite database |
| `-auth-api-url` | | Remote auth API URL |
| `-auth-api-key` | | API key for remote auth |
| `-accounting-mode` | `local` | Accounting mode: `local`, `remote`, `both`, or `mock` |
| `-accounting-db-path` | `./traffic.db` | Path to traffic SQLite database |
| `-accounting-api-url` | | Remote accounting API URL |
| `-accounting-api-key` | | API key for remote accounting |
| `-accounting-interval` | `60s` | Interval for periodic accounting updates |
| `-mock-api` | `false` | Mock API calls (log instead of real requests) |

### Examples

Run with local SQLite databases:
```bash
./socks5 -v
```

Run with remote API:
```bash
./socks5 \
  -auth-mode remote \
  -auth-api-url https://api.example.com \
  -auth-api-key your-key \
  -accounting-mode remote \
  -accounting-api-url https://api.example.com \
  -accounting-api-key your-key
```

Run in mock mode for testing:
```bash
./socks5 -mock-api -v
```

## API Contracts

### Authentication

**POST** `/api/login`
```json
{"user": "email@address.com", "password": "password"}
```
- `200 OK` = Valid credentials
- `403 Forbidden` = Invalid credentials

### Traffic Accounting

**POST** `/api/connect`
```json
{"username": "user", "target": "1.2.3.4:443"}
```

**POST** `/api/update`
```json
{"username": "user", "target": "1.2.3.4:443", "bytes_sent": 1024, "bytes_recv": 2048, "duration_seconds": 60}
```

**POST** `/api/disconnect`
```json
{"username": "user", "target": "1.2.3.4:443", "bytes_sent": 2048, "bytes_recv": 4096, "duration_seconds": 120}
```

## Database Schemas

### Users (local auth)
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Connections (traffic accounting)
```sql
CREATE TABLE connections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    target TEXT NOT NULL,
    bytes_sent INTEGER DEFAULT 0,
    bytes_recv INTEGER DEFAULT 0,
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP,
    duration_seconds INTEGER
);
```

## Building

```bash
go build -o socks5 .
```

## Testing

```bash
go test ./...
```

## Credits

Originally based on [ring04h/s5.go](http://github.com/ring04h/s5.go)