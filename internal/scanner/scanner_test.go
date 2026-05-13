package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "scanner-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some dummy files
	os.MkdirAll(filepath.Join(tmpDir, "sales"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "sales", "orders.yaml"), []byte("dummy"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "sales", "ignore_me.txt"), []byte("dummy"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "root.yaml"), []byte("dummy"), 0644)

	// Run the scanner
	files, err := Scan(tmpDir, ".yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	// Validate the found files
	var rootFile, salesFile *File
	for i, f := range files {
		if f.BaseName == "root" {
			rootFile = &files[i]
		} else if f.BaseName == "orders" {
			salesFile = &files[i]
		}
	}

	if rootFile == nil || salesFile == nil {
		t.Fatalf("failed to find expected files in results: %v", files)
	}

	if rootFile.RelativeDir != "" {
		t.Errorf("expected rootFile RelativeDir to be empty, got %q", rootFile.RelativeDir)
	}

	if salesFile.RelativeDir != "sales" {
		t.Errorf("expected salesFile RelativeDir to be 'sales', got %q", salesFile.RelativeDir)
	}
}

func TestScan_NonExistentDir(t *testing.T) {
	_, err := Scan("/path/that/does/not/exist/hopefully", ".yaml")
	if err == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}
}
