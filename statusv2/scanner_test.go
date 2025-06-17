package statusv2

import (
	"strings"
	"testing"
)

func TestZScanner_BasicEntries(t *testing.T) {
	// Test basic entries that should split on first NUL
	input := "1 M. N... 100644 100644 100644 hash1 hash2 file.txt\x00" +
		"? untracked.txt\x00" +
		"! ignored.txt\x00"

	scanner := newZScanner(strings.NewReader(input))

	expected := []string{
		"1 M. N... 100644 100644 100644 hash1 hash2 file.txt",
		"? untracked.txt",
		"! ignored.txt",
	}

	var results []string
	for scanner.Scan() {
		results = append(results, string(scanner.Bytes()))
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Scanner error: %v", err)
	}

	if len(results) != len(expected) {
		t.Fatalf("Expected %d entries, got %d", len(expected), len(results))
	}

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Entry %d: expected %q, got %q", i, expected[i], result)
		}
	}
}

func TestZScanner_RenameEntry(t *testing.T) {
	// Test rename/copy entry that should include internal NUL separator
	input := "2 R. N... 100644 100644 100644 hash1 hash2 R100 newpath.txt\x00oldpath.txt\x00"

	scanner := newZScanner(strings.NewReader(input))

	expected := "2 R. N... 100644 100644 100644 hash1 hash2 R100 newpath.txt\x00oldpath.txt"

	if !scanner.Scan() {
		t.Fatal("Expected to scan one entry")
	}

	result := string(scanner.Bytes())
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Should not have more entries
	if scanner.Scan() {
		t.Error("Expected only one entry")
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Scanner error: %v", err)
	}
}

func TestZScanner_RenameAtEOF(t *testing.T) {
	// Test rename entry at EOF with both paths (should work fine)
	input := "2 R. N... 100644 100644 100644 hash1 hash2 R100 newpath.txt\x00oldpath.txt"

	scanner := newZScanner(strings.NewReader(input))

	expected := "2 R. N... 100644 100644 100644 hash1 hash2 R100 newpath.txt\x00oldpath.txt"

	if !scanner.Scan() {
		t.Fatal("Expected to scan one entry")
	}

	result := string(scanner.Bytes())
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Should not have more entries
	if scanner.Scan() {
		t.Error("Expected only one entry")
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Scanner error: %v", err)
	}
}

func TestZScanner_MixedEntries(t *testing.T) {
	// Test mix of normal and rename entries
	input := "1 M. N... 100644 100644 100644 hash1 hash2 modified.txt\x00" +
		"2 R. N... 100644 100644 100644 hash1 hash2 R100 renamed.txt\x00original.txt\x00" +
		"? untracked.txt\x00"

	scanner := newZScanner(strings.NewReader(input))

	expected := []string{
		"1 M. N... 100644 100644 100644 hash1 hash2 modified.txt",
		"2 R. N... 100644 100644 100644 hash1 hash2 R100 renamed.txt\x00original.txt",
		"? untracked.txt",
	}

	var results []string
	for scanner.Scan() {
		results = append(results, string(scanner.Bytes()))
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Scanner error: %v", err)
	}

	if len(results) != len(expected) {
		t.Fatalf("Expected %d entries, got %d", len(expected), len(results))
	}

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Entry %d: expected %q, got %q", i, expected[i], result)
		}
	}
}

func TestZScanner_EmptyInput(t *testing.T) {
	scanner := newZScanner(strings.NewReader(""))

	if scanner.Scan() {
		t.Error("Expected no entries for empty input")
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Scanner error: %v", err)
	}
}

func TestZScanner_CorruptedRename(t *testing.T) {
	// Test rename entry with only one NUL (corrupted data)
	input := "2 R. N... 100644 100644 100644 hash1 hash2 R100 newpath.txt\x00"

	scanner := newZScanner(strings.NewReader(input))

	// Should not scan successfully
	if scanner.Scan() {
		t.Error("Expected scan to fail for corrupted rename entry")
	}

	// Should have an error
	if err := scanner.Err(); err == nil {
		t.Error("Expected error for corrupted rename entry")
	}
}
