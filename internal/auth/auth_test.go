package auth

import (
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestLocalAuthenticator_Validate(t *testing.T) {
	tmpFile := t.TempDir() + "/test_users.db"
	defer os.Remove(tmpFile)

	db, err := OpenDB(tmpFile)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	_, err = db.CreateUser("testuser", string(hashedPassword))
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	auth := &LocalAuthenticator{db: db}

	tests := []struct {
		name       string
		username   string
		password   string
		wantValid  bool
	}{
		{"valid credentials", "testuser", "testpassword", true},
		{"invalid password", "testuser", "wrongpassword", false},
		{"invalid username", "nonexistent", "testpassword", false},
		{"empty credentials rejected", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := auth.Validate(tt.username, tt.password)
			if valid != tt.wantValid {
				t.Errorf("expected valid=%v, got %v", tt.wantValid, valid)
			}
		})
	}
}

func TestLocalAuthenticator_UserNotFound(t *testing.T) {
	tmpFile := t.TempDir() + "/test_users_empty.db"
	defer os.Remove(tmpFile)

	db, err := OpenDB(tmpFile)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	auth := &LocalAuthenticator{db: db}

	valid, _ := auth.Validate("nonexistent", "anypassword")
	if valid != false {
		t.Errorf("expected valid=false for nonexistent user")
	}
}

func TestDB_CreateUser(t *testing.T) {
	tmpFile := t.TempDir() + "/test_create.db"
	defer os.Remove(tmpFile)

	db, err := OpenDB(tmpFile)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	id, err := db.CreateUser("newuser", string(hashedPassword))
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}

	user, err := db.GetUser("newuser")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if user.Username != "newuser" {
		t.Errorf("expected username newuser, got %s", user.Username)
	}
}

func TestDB_DeleteUser(t *testing.T) {
	tmpFile := t.TempDir() + "/test_delete.db"
	defer os.Remove(tmpFile)

	db, err := OpenDB(tmpFile)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	_, _ = db.CreateUser("todelete", string(hashedPassword))

	err = db.DeleteUser("todelete")
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	_, err = db.GetUser("todelete")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}