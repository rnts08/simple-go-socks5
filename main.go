package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"

	"vn-socks-proxy/internal/accounting"
	"vn-socks-proxy/internal/auth"
	"vn-socks-proxy/internal/config"
)

var (
	Commands      = []string{"CONNECT", "BIND", "UDP ASSOCIATE"}
	AddrType     = []string{"", "IPv4", "", "Domain", "IPv6"}
	Verbose      = false
	errAddrType   = fmt.Errorf("socks addr type not supported")
	errVer       = fmt.Errorf("socks version not supported")
	errAuthExtraData = fmt.Errorf("socks authentication received extra data")
	errReqExtraData = fmt.Errorf("socks request received extra data")
	errCmd        = fmt.Errorf("socks only supports CONNECT command")
	connStats    = make(map[net.Conn]*ClientStats)
	statsMutex   sync.RWMutex
)

const (
	socksVer5       = 0x05
	socksCmdConnect = 0x01
)

type ClientStats struct {
	StartTime time.Time
	BytesSent uint64
	BytesRecv uint64
	Username string
	Target   string
}

func handShake(conn net.Conn) error {
	buf := make([]byte, 258)

	n, err := io.ReadAtLeast(conn, buf, 2)
	if err != nil {
		return fmt.Errorf("failed to read handshake: %w", err)
	}

	if buf[0] != socksVer5 {
		return errVer
	}

	nmethods := int(buf[1])
	msgLen := nmethods + 2

	if n < msgLen {
		if _, err = io.ReadFull(conn, buf[n:msgLen]); err != nil {
			return fmt.Errorf("failed to read methods: %w", err)
		}
	} else if n > msgLen {
		return errAuthExtraData
	}

	_, err = conn.Write([]byte{socksVer5, 0})
	return err
}

func parseTarget(conn net.Conn) (string, error) {
	const (
		idVer   = 0
		idCmd   = 1
		idType = 3
		idIP0  = 4
		idDmLen = 4
		idDm0  = 5

		typeIPv4 = 1
		typeDm  = 3
		typeIPv6 = 4

		lenIPv4 = 10
		lenIPv6 = 22
		lenDmBase = 7
	)

	buf := make([]byte, 263)
	n, err := io.ReadAtLeast(conn, buf, idDmLen+1)
	if err != nil {
		return "", fmt.Errorf("failed to read target: %w", err)
	}

	if buf[idVer] != socksVer5 {
		return "", errVer
	}

	if buf[idCmd] != socksCmdConnect {
		return "", errCmd
	}

	var reqLen int
	switch buf[idType] {
	case typeIPv4:
		reqLen = lenIPv4
	case typeIPv6:
		reqLen = lenIPv6
	case typeDm:
		reqLen = int(buf[idDmLen]) + lenDmBase
	default:
		return "", errAddrType
	}

	if n < reqLen {
		if _, err := io.ReadFull(conn, buf[n:reqLen]); err != nil {
			return "", fmt.Errorf("failed to read full request: %w", err)
		}
	} else if n > reqLen {
		return "", errReqExtraData
	}

	var host string
	switch buf[idType] {
	case typeIPv4:
		host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
	case typeIPv6:
		host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	}

	port := binary.BigEndian.Uint16(buf[reqLen-2:])
	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}

type readWriter struct {
	net.Conn
	stats *ClientStats
	rec   accounting.TrafficRecorder
}

func (rw *readWriter) Read(b []byte) (int, error) {
	n, err := rw.Conn.Read(b)
	if n > 0 {
		statsMutex.Lock()
		rw.stats.BytesRecv += uint64(n)
		statsMutex.Unlock()
	}
	return n, err
}

func (rw *readWriter) Write(b []byte) (int, error) {
	n, err := rw.Conn.Write(b)
	if n > 0 {
		statsMutex.Lock()
		rw.stats.BytesSent += uint64(n)
		statsMutex.Unlock()
	}
	return n, err
}

func pipeWhenClose(conn net.Conn, target string, stats *ClientStats, rec accounting.TrafficRecorder) {
	if Verbose {
		log.Println("Connecting to remote:", target)
	}

	remoteConn, err := net.DialTimeout("tcp", target, 15*time.Second)
	if err != nil {
		log.Println("Failed to connect to remote:", err)
		return
	}
	defer remoteConn.Close()

	tcpAddr := remoteConn.LocalAddr().(*net.TCPAddr)
	reply := make([]byte, 10)
	reply[0], reply[1], reply[2] = 0x05, 0x00, 0x00

	ip := tcpAddr.IP.To4()
	if ip == nil {
		ip = tcpAddr.IP.To16()
		reply[3] = 0x04
	} else {
		reply[3] = 0x01
	}
	copy(reply[4:], ip)
	reply[8] = byte(tcpAddr.Port >> 8)
	reply[9] = byte(tcpAddr.Port & 0xff)

	if _, err := conn.Write(reply); err != nil {
		log.Println("Failed to send response to client:", err)
		return
	}

	startTime := time.Now()
	if rec != nil {
		_ = rec.Connect(stats.Username, target)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(remoteConn, conn)
		conn.Close()
	}()

	go func() {
		defer wg.Done()
		io.Copy(conn, remoteConn)
		remoteConn.Close()
	}()

	wg.Wait()

	duration := time.Since(startTime)
	statsMutex.RLock()
	bytesSent := stats.BytesSent
	bytesRecv := stats.BytesRecv
	statsMutex.RUnlock()

	if rec != nil {
		_ = rec.Disconnect(stats.Username, target, bytesSent, bytesRecv, duration)
	}

	if Verbose {
		log.Printf("Connection closed: %s -> %s, duration: %v, sent: %d, recv: %d",
			stats.Username, target, duration, bytesSent, bytesRecv)
	}
}

func handleConnection(conn net.Conn, authenticator auth.Authenticator, recorder accounting.TrafficRecorder) {
	stats := &ClientStats{
		StartTime: time.Now(),
		Username:  "anonymous",
	}

	statsMutex.Lock()
	connStats[conn] = stats
	statsMutex.Unlock()

	defer func() {
		statsMutex.Lock()
		delete(connStats, conn)
		statsMutex.Unlock()
		conn.Close()
	}()

	if err := handShake(conn); err != nil {
		log.Println("Handshake failed:", err)
		return
	}

	if authenticator != nil {
		username, err := handleAuth(conn, authenticator)
		if err != nil {
			log.Println("Authentication failed:", err)
			sendAuthFailure(conn)
			return
		}
		stats.Username = username
		_, _ = conn.Write([]byte{socksVer5, 0x00})
	}

	target, err := parseTarget(conn)
	if err != nil {
		log.Println("Failed to parse target:", err)
		return
	}
	stats.Target = target

	pipeWhenClose(conn, target, stats, recorder)
}

func handleAuth(conn net.Conn, authenticator auth.Authenticator) (string, error) {
	buf := make([]byte, 258)

	n, err := io.ReadAtLeast(conn, buf, 4)
	if err != nil {
		return "", fmt.Errorf("failed to read auth: %w", err)
	}

	if buf[0] != 0x01 {
		return "", fmt.Errorf("unsupported auth version: %d", buf[0])
	}

	offset := 2
	if n < offset+2 {
		return "", fmt.Errorf("incomplete auth data")
	}

	userLen := int(buf[offset])
	offset++
	if n < offset+userLen {
		return "", fmt.Errorf("incomplete username")
	}
	username := string(buf[offset : offset+userLen])
	offset += userLen

	passLen := int(buf[offset])
	offset++
	if n < offset+passLen {
		return "", fmt.Errorf("incomplete password")
	}
	password := string(buf[offset : offset+passLen])

	valid, err := authenticator.Validate(username, password)
	if err != nil {
		return "", fmt.Errorf("auth validation error: %w", err)
	}
	if !valid {
		return "", fmt.Errorf("invalid credentials")
	}

	return username, nil
}

func sendAuthFailure(conn net.Conn) {
	_, _ = conn.Write([]byte{socksVer5, 0x01})
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cfg := config.Parse()

	Verbose = cfg.Verbose

	var authenticator auth.Authenticator
	var recorder accounting.TrafficRecorder
	var authErr, recErr error

	if cfg.AuthMode != config.ModeMock {
		authenticator, authErr = auth.NewAuthenticator(cfg)
		if authErr != nil {
			log.Printf("Warning: failed to initialize authenticator: %v", authErr)
		}
		defer func() {
			if authenticator != nil {
				authenticator.Close()
			}
		}()
	}

	if cfg.AccountingMode != config.ModeMock {
		recorder, recErr = accounting.NewRecorder(cfg)
		if recErr != nil {
			log.Printf("Warning: failed to initialize recorder: %v", recErr)
		}
		defer func() {
			if recorder != nil {
				recorder.Close()
			}
		}()
	}

	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	log.Printf("SOCKS5 proxy listening on %s", cfg.ListenAddr)
	if cfg.Verbose {
		log.Printf("Auth mode: %s, Accounting mode: %s", cfg.AuthMode, cfg.AccountingMode)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			return
		}
		if Verbose {
			log.Println("New connection from:", conn.RemoteAddr())
		}
		go handleConnection(conn, authenticator, recorder)
	}
}