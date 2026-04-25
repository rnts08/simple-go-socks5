package accounting

import (
	"os"
	"testing"
	"time"
)

func TestLocalRecorder_Connect(t *testing.T) {
	tmpFile := t.TempDir() + "/test_traffic.db"
	defer os.Remove(tmpFile)

	rec, err := NewLocalRecorder(tmpFile, 60*time.Second)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Close()

	err = rec.Connect("testuser", "target.example.com:443")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLocalRecorder_Update(t *testing.T) {
	tmpFile := t.TempDir() + "/test_traffic_update.db"
	defer os.Remove(tmpFile)

	rec, err := NewLocalRecorder(tmpFile, 60*time.Second)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Close()

	err = rec.Connect("testuser", "target.example.com:443")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	err = rec.Update("testuser", "target.example.com:443", 1024, 2048, 30*time.Second)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLocalRecorder_Disconnect(t *testing.T) {
	tmpFile := t.TempDir() + "/test_traffic_disconnect.db"
	defer os.Remove(tmpFile)

	rec, err := NewLocalRecorder(tmpFile, 60*time.Second)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Close()

	err = rec.Connect("testuser", "target.example.com:443")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	err = rec.Disconnect("testuser", "target.example.com:443", 1024, 2048, 60*time.Second)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLocalRecorder_FullFlow(t *testing.T) {
	tmpFile := t.TempDir() + "/test_traffic_flow.db"
	defer os.Remove(tmpFile)

	rec, err := NewLocalRecorder(tmpFile, 60*time.Second)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer rec.Close()

	username := "testuser"
	target := "target.example.com:443"
	bytesSent := uint64(1024)
	bytesRecv := uint64(2048)
	duration := 30 * time.Second

	err = rec.Connect(username, target)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	err = rec.Update(username, target, bytesSent, bytesRecv, duration)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	err = rec.Disconnect(username, target, bytesSent*2, bytesRecv*2, duration*2)
	if err != nil {
		t.Fatalf("disconnect failed: %v", err)
	}
}