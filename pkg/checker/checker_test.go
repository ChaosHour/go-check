package checker

import (
	"os"
	"strings"
	"testing"
)

func TestReadMyCnf(t *testing.T) {
	// Create a temporary directory and .my.cnf file
	content := `[client]
user = testuser
password = testpass
host = localhost
`
	tmpDir, err := os.MkdirTemp("", "test_home")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cnfFile := tmpDir + "/.my.cnf"
	err = os.WriteFile(cnfFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Temporarily set HOME to the temp dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	user, password, host, err := ReadMyCnf()
	if err != nil {
		t.Fatalf("ReadMyCnf failed: %v", err)
	}

	if user != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", user)
	}
	if password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", password)
	}
	if host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", host)
	}
}

func TestReadMyCnfMissingFile(t *testing.T) {
	// Set HOME to a non-existent directory
	os.Setenv("HOME", "/nonexistent")
	defer os.Setenv("HOME", "")

	_, _, _, err := ReadMyCnf()
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestReadMyCnfIncomplete(t *testing.T) {
	content := `[client]
user = testuser
`
	tmpFile, err := os.CreateTemp("", ".my.cnf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", strings.TrimSuffix(tmpFile.Name(), "/.my.cnf"))
	defer os.Setenv("HOME", oldHome)

	_, _, _, err = ReadMyCnf()
	if err == nil {
		t.Error("Expected error for incomplete credentials, got nil")
	}
}
