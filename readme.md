# SOCKS5 Proxy

A production-ready SOCKS5 proxy server written in Go with authentication and traffic accounting capabilities.

## Features

- **SOCKS5 Protocol**: RFC 1928 compliant CONNECT command support
- **Authentication**: 
  - Local SQLite database with bcrypt password hashing
  - Remote API validation endpoint
- **Traffic Accounting**:
  - Local SQLite logging
  - Remote API reporting with connect/update/disconnect events
- **Flexible Modes**: Run in local-only, remote-only, or both modes simultaneously
- **Mock Mode**: Test without real backends - logs requests instead of making API calls
- **System Integration**: systemd service and SysVinit script for Debian/Ubuntu

## Requirements

### Runtime

- **Go 1.25+**: If building from source
- **Linux**: For systemd/SysVinit integration
- **Network access**: For remote API modes

### For Local Authentication

- Write access to SQLite database path
- `users` table (see schema below)

### For Local Traffic Accounting

- Write access to SQLite database path
- `connections` table (see schema below)

### For Remote API

The server makes HTTP POST requests to your API endpoints:

| Endpoint | Purpose |
|----------|---------|
| `POST {auth-api-url}/api/login` | Validate user credentials |
| `POST {accounting-api-url}/api/connect` | Log new connection |
| `POST {accounting-api-url}/api/update` | Periodic traffic update |
| `POST {accounting-api-url}/api/disconnect` | Log connection close |

## Download & Build

### Option 1: Clone and Build

```bash
# Clone the repository
git clone https://github.com/yourrepo/socks5.git
cd socks5

# Build
go build -o socks5 .
# or
make build
```

### Option 2: Download Release

Download the latest .deb package from the releases page for your architecture:

```bash
# Install (Debian/Ubuntu)
sudo dpkg -i go-socks5_*.deb

# Or extract manually
dpkg-deb -x go-socks5_*.deb /tmp/socks5
```

## Usage

### Quick Start (Local Mode)

```bash
# Run with defaults (local SQLite)
./socks5 -v
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8080` | Proxy listen address |
| `-v` | `false` | Enable verbose logging |
| `-auth-mode` | `local` | Auth mode: `local`, `remote`, or `mock` |
| `-auth-db-path` | `./users.db` | Path to auth SQLite database |
| `-auth-api-url` | | Remote auth API base URL |
| `-auth-api-key` | | API key for remote auth |
| `-accounting-mode` | `local` | Accounting mode: `local`, `remote`, `both`, or `mock` |
| `-accounting-db-path` | `./traffic.db` | Path to traffic SQLite database |
| `-accounting-api-url` | | Remote accounting API base URL |
| `-accounting-api-key` | | API key for remote accounting |
| `-accounting-interval` | `60s` | Interval for periodic updates |
| `-mock-api` | `false` | Mock mode (log only, no real API calls) |

### Examples

#### Local SQLite (Default)

```bash
./socks5 -v
```

#### Remote API Only

```bash
./socks5 \
  -auth-mode remote \
  -auth-api-url https://api.example.com \
  -auth-api-key your-secret-key \
  -accounting-mode remote \
  -accounting-api-url https://api.example.com \
  -accounting-api-key your-secret-key
```

#### Both Local and Remote

```bash
./socks5 \
  -auth-mode local \
  -auth-db-path /var/lib/socks5/users.db \
  -accounting-mode both \
  -accounting-db-path /var/lib/socks5/traffic.db \
  -accounting-api-url https://api.example.com \
  -accounting-api-key your-key
```

#### Mock Mode (Testing)

```bash
./socks5 -mock-api -v
```

## Remote API Specification

### Authentication API

The server POSTs to `{auth-api-url}/api/login`

**Request:**
```json
{
  "user": "username@example.com",
  "password": "userpassword"
}
```

Headers:
- `Content-Type: application/json`
- `Authorization: Bearer {auth-api-key}` (if configured)

**Response:**
- `200 OK` - Valid credentials
- `403 Forbidden` - Invalid credentials
- Other - Error (connection closed)

### Traffic Accounting API

The server POSTs to `{accounting-api-url}/api/...`

#### Connect Event
**POST** `/api/connect`

```json
{
  "username": "user", 
  "target": "1.2.3.4:443"
}
```

#### Update Event  
**POST** `/api/update`

```json
{
  "username": "user",
  "target": "1.2.3.4:443",
  "bytes_sent": 1024,
  "bytes_recv": 2048,
  "duration_seconds": 60
}
```

#### Disconnect Event
**POST** `/api/disconnect`

```json
{
  "username": "user",
  "target": "1.2.3.4:443",
  "bytes_sent": 2048,
  "bytes_recv": 4096,
  "duration_seconds": 120
}
```

Headers for all:
- `Content-Type: application/json`
- `Authorization: Bearer {accounting-api-key}` (if configured)

## Database Schemas

### Users Table (Local Auth)

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_users_username ON users(username);
```

### Connections Table (Traffic Accounting)

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
CREATE INDEX idx_connections_username ON connections(username);
CREATE INDEX idx_connections_start ON connections(start_time);
```

## Installation (Linux)

### Using Makefile (Recommended)

```bash
make install
```

This installs:
- Binary: `/usr/local/bin/socks5`
- Config dir: `/etc/socks5/`
- Default config: `/etc/default/socks5`
- systemd service: `/etc/systemd/system/socks5.service`
- SysV init: `/etc/init.d/socks5`

### Start the Service

```bash
# systemd
systemctl start socks5
systemctl enable socks5  # enable on boot

# SysVinit
/etc/init.d/socks5 start
update-rc.d socks5 defaults
```

### Uninstall

```bash
make uninstall
```

## Testing

```bash
go test ./...
# or
make test
```

## Credits

Originally based on [ring04h/s5.go](http://github.com/ring04h/s5.go)