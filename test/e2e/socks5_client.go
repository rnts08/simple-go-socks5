package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

var (
	proxyHost = flag.String("proxy", "127.0.0.1:8080", "SOCKS5 proxy address")
	target  = flag.String("target", "example.com:80", "Target to connect to")
	user    = flag.String("user", "", "Username for auth")
	pass    = flag.String("pass", "", "Password for auth")
	verbose = flag.Bool("v", false, "Verbose output")
)

	const socksVer5 = 0x05

func main() {
	flag.Parse()

	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	log.Printf("Connecting to proxy at %s", *proxyHost)
	log.Printf("Target: %s", *target)

	conn, err := net.Dial("tcp", *proxyHost)
	if err != nil {
		log.Fatalf("Failed to connect to proxy: %v", err)
	}
	defer conn.Close()

	// SOCKS5 handshake
	if err := handshake(conn); err != nil {
		log.Fatalf("Handshake failed: %v", err)
	}

	// Authenticate if credentials provided
	if *user != "" {
		if err := authenticate(conn, *user, *pass); err != nil {
			log.Fatalf("Auth failed: %v", err)
		}
		log.Printf("Authenticated as %s", *user)
	}

	// Connect to target
	if err := connectTarget(conn, *target); err != nil {
		log.Fatalf("Connect failed: %v", err)
	}

	log.Printf("Connected to %s", *target)

	// Send test data and verify
	testData := []byte("GET / HTTP/1.0\r\nHost: example.com\r\n\r\n")
	n, err := conn.Write(testData)
	if err != nil {
		log.Printf("Write error: %v", err)
	} else {
		log.Printf("Wrote %d bytes", n)
	}

	// Read response
	buf := make([]byte, 4096)
	n, err = conn.Read(buf)
	if err != nil && err != io.EOF {
		log.Printf("Read error: %v", err)
	} else {
		log.Printf("Read %d bytes", n)
		if *verbose {
			log.Printf("Response: %s", string(buf[:n]))
		}
	}

	log.Println("Test completed successfully")
}

func handshake(conn net.Conn) error {
	// Send greeting: version 5, 1 auth method (none)
	_, err := conn.Write([]byte{socksVer5, 0x01, 0x00})
	if err != nil {
		return err
	}

	// Read server auth selection
	buf := make([]byte, 2)
	if _, err := io.ReadAtLeast(conn, buf, 2); err != nil {
		return err
	}

	if buf[0] != socksVer5 {
		return fmt.Errorf("unexpected version: %d", buf[0])
	}

	// 0x00 = no auth, 0x02 = username/password
	if buf[1] != 0x00 && buf[1] != 0x02 {
		return fmt.Errorf("unsupported auth method: %d", buf[1])
	}

	log.Printf("Handshake complete, auth method: %d", buf[1])
	return nil
}

func authenticate(conn net.Conn, user, pass string) error {
	// Send username/password auth request
	// Format: version(1) + userlen(1) + user + passlen(1) + pass
	data := []byte{0x01}
	data = append(data, byte(len(user)))
	data = append(data, user...)
	data = append(data, byte(len(pass)))
	data = append(data, pass...)

	_, err := conn.Write(data)
	if err != nil {
		return err
	}

	// Read response
	buf := make([]byte, 2)
	if _, err := io.ReadAtLeast(conn, buf, 2); err != nil {
		return err
	}

	if buf[1] != 0x00 {
		return fmt.Errorf("auth failed with code: %d", buf[1])
	}

	return nil
}

func connectTarget(conn net.Conn, target string) error {
	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	// Build connect request
	req := []byte{
		socksVer5,        // Version
		0x01,           // CONNECT command
		0x00,           // Reserved
		0x03,           // DOMAIN address type
		byte(len(host)),  // Domain length
	}
	req = append(req, host...)
	req = append(req, byte(port>>8), byte(port&0xff))

	_, err = conn.Write(req)
	if err != nil {
		return err
	}

	// Read reply
	buf := make([]byte, 10)
	if _, err := io.ReadAtLeast(conn, buf, 10); err != nil {
		return err
	}

	if buf[0] != socksVer5 {
		return fmt.Errorf("invalid version in reply")
	}

	if buf[1] != 0x00 {
		return fmt.Errorf("connect failed with code: %d", buf[1])
	}

	log.Printf("Connected to %s", target)
	return nil
}