package statusv2

import (
	"strings"
	"testing"
)

func TestZScanner(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{
			name: "basic entries",
			input: "1 M. N... 100644 100644 100644 hash1 hash2 file.txt\x00" +
				"? untracked.txt\x00" +
				"! ignored.txt\x00",
			expected: []string{
				"1 M. N... 100644 100644 100644 hash1 hash2 file.txt",
				"? untracked.txt",
				"! ignored.txt",
			},
		},
		{
			name: "mixed entries",
			input: "1 M. N... 100644 100644 100644 hash1 hash2 modified.txt\x00" +
				"2 R. N... 100644 100644 100644 hash1 hash2 R100 renamed.txt\x00original.txt\x00" +
				"? untracked.txt\x00",
			expected: []string{
				"1 M. N... 100644 100644 100644 hash1 hash2 modified.txt",
				"2 R. N... 100644 100644 100644 hash1 hash2 R100 renamed.txt\x00original.txt",
				"? untracked.txt",
			},
		},
		{
			name:     "rename entry with double NUL",
			input:    "2 R. N... 100644 100644 100644 hash1 hash2 R100 newpath.txt\x00oldpath.txt\x00",
			expected: []string{"2 R. N... 100644 100644 100644 hash1 hash2 R100 newpath.txt\x00oldpath.txt"},
		},
		{
			name:     "rename entry at EOF",
			input:    "2 R. N... 100644 100644 100644 hash1 hash2 R100 newpath.txt\x00oldpath.txt",
			expected: []string{"2 R. N... 100644 100644 100644 hash1 hash2 R100 newpath.txt\x00oldpath.txt"},
		},
		{
			name:     "entry at EOF without NUL terminator",
			input:    "? untracked_no_nul.txt",
			expected: []string{"? untracked_no_nul.txt"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:    "malformed rename entry at end (missing second path)",
			input:   "2 R. N... 100644 100644 100644 hash1 hash2 R100 newpath.txt\x00",
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			scanner := newZScanner(strings.NewReader(tc.input))

			var results []string
			for scanner.Scan() {
				results = append(results, string(scanner.Bytes()))
			}

			err := scanner.Err()
			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Scanner error: %v", err)
			}

			if len(results) != len(tc.expected) {
				t.Fatalf("Expected %d entries, got %d", len(tc.expected), len(results))
			}

			for i, result := range results {
				if result != tc.expected[i] {
					t.Errorf("Entry %d: expected %q, got %q", i, tc.expected[i], result)
				}
			}
		})
	}
}
