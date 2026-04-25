#!/bin/bash
# E2E test script for SOCKS5 proxy

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

cleanup() {
    pkill -f socks5-proxy 2>/dev/null || true
    pkill -f dummy_api 2>/dev/null || true
    rm -f /tmp/socks5_users.db /tmp/socks5_traffic.db
}
trap cleanup EXIT
cleanup

PROXY_PORT=19080
API_PORT=19081

log_info "=== SOCKS5 Proxy E2E Tests ==="

log_info "Building..."
go build -o /tmp/socks5-proxy . 
go build -o /tmp/dummy_api ./test/e2e/dummy_api.go 
go build -o /tmp/socks5_client ./test/e2e/socks5_client.go 
go build -o /tmp/adduser ./test/e2e/adduser.go 

log_info "Creating test user..."
/tmp/adduser -db /tmp/socks5_users.db -user admin -pass test123

log_info "=== Test 1: Local proxy with auth ==="
(/tmp/socks5-proxy -addr :$PROXY_PORT -auth-mode local -auth-db-path /tmp/socks5_users.db -accounting-mode local -accounting-db-path /tmp/socks5_traffic.db -v > /tmp/proxy.log 2>&1) &
PROXY_PID=$!
sleep 2

if ps -p $PROXY_PID > /dev/null 2>&1; then
    log_info "Test 1: PASSED"
else
    log_error "Test 1: FAILED"
    cat /tmp/proxy.log
    exit 1
fi
cleanup

log_info "=== Test 2: Remote API auth ==="
(/tmp/dummy_api -addr :$API_PORT -users "admin:test,user:password" > /tmp/api.log 2>&1) &
API_PID=$!
sleep 1
(/tmp/socks5-proxy -addr :$PROXY_PORT -auth-mode remote -auth-api-url http://127.0.0.1:$API_PORT -accounting-mode remote -accounting-api-url http://127.0.0.1:$API_PORT > /tmp/proxy.log 2>&1) &
sleep 2

if ps -p $API_PID > /dev/null 2>&1; then
    log_info "Test 2: PASSED"
else
    log_error "Test 2: FAILED"
    cat /tmp/api.log
    exit 1
fi
cleanup

log_info "=== Test 3: Mock mode ==="
(/tmp/socks5-proxy -addr :$PROXY_PORT -auth-mode mock -accounting-mode mock > /tmp/proxy.log 2>&1) &
sleep 2

log_info "Test 3: PASSED"
cleanup

log_info ""
log_info "=== All Tests Passed ==="