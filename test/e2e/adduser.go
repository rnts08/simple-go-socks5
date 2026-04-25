package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/glebarez/go-sqlite"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbPath := flag.String("db", "./users.db", "Database path")
	user := flag.String("user", "", "Username")
	pass := flag.String("pass", "", "Password")
	flag.Parse()

	if *user == "" || *pass == "" {
		fmt.Fprintf(os.Stderr, "Usage: adduser -db path -user username -pass password\n")
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*pass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", *user, string(hash))
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	log.Printf("Created user: %s", *user)
}