package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
)

var (
	addr        = flag.String("addr", ":8081", "API server listen address")
	validUsers = flag.String("users", "admin:test,user:password", "Valid users (user:pass,user:pass)")
	printReq   = flag.Bool("print", false, "Print all requests")
)

	type LoginRequest struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}

	type TrafficRequest struct {
		Username        string `json:"username"`
		Target         string `json:"target"`
		BytesSent      uint64 `json:"bytes_sent"`
		BytesRecv      uint64 `json:"bytes_recv"`
		DurationSeconds int64  `json:"duration_seconds"`
	}

	var (
		mu     sync.RWMutex
		users   map[string]string
		records = make(map[string]int)
	)

func main() {
	flag.Parse()
	users = parseUsers(*validUsers)

	log.Printf("Starting test API server on %s", *addr)
	log.Printf("Valid users: %v", users)

	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/connect", handleConnect)
	http.HandleFunc("/api/update", handleUpdate)
	http.HandleFunc("/api/disconnect", handleDisconnect)
	http.HandleFunc("/stats", handleStats)

	go func() {
		if err := http.ListenAndServe(*addr, nil); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	log.Println("Shutting down...")
}

func parseUsers(s string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, ":")
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("login: decode error: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if *printReq {
		log.Printf("LOGIN: user=%s", req.User)
	}

	mu.RLock()
	validPass, ok := users[req.User]
	mu.RUnlock()

	if !ok || req.Password != validPass {
		log.Printf("LOGIN: failed for %s", req.User)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	log.Printf("LOGIN: success for %s", req.User)
	w.WriteHeader(http.StatusOK)
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TrafficRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if *printReq {
		log.Printf("CONNECT: %s -> %s", req.Username, req.Target)
	}

	key := req.Username + "->" + req.Target
	mu.Lock()
	records[key]++
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TrafficRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if *printReq {
		log.Printf("UPDATE: %s -> %s sent=%d recv=%d dur=%ds",
			req.Username, req.Target, req.BytesSent, req.BytesRecv, req.DurationSeconds)
	}

	key := req.Username + "->" + req.Target
	mu.Lock()
	records[key]++
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TrafficRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if *printReq {
		log.Printf("DISCONNECT: %s -> %s sent=%d recv=%d dur=%ds",
			req.Username, req.Target, req.BytesSent, req.BytesRecv, req.DurationSeconds)
	}

	w.WriteHeader(http.StatusOK)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid_users": len(users),
		"records":    records,
	})
}