package statusv1

import (
	"bytes"
	"testing"
)

// FuzzParse tests the Parse function with arbitrary input
func FuzzParse(f *testing.F) {
	// Add seed inputs covering various entry types
	f.Add([]byte(" M file.txt\nA  added.txt\nD  deleted.txt"))
	f.Add([]byte("R  new.txt -> old.txt\n?? untracked.txt"))
	f.Add([]byte("!! ignored.txt\nMM conflict.txt"))
	f.Add([]byte("C  copy.txt -> orig.txt"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parser should never panic, only return an error for invalid input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parse panicked with input %q: %v", data, r)
			}
		}()
		Parse(bytes.NewReader(data))
	})
}

// FuzzParseZ tests the ParseZ function with arbitrary input
func FuzzParseZ(f *testing.F) {
	// Add seed inputs in -z format
	f.Add([]byte(" M file.txt\x00A  added.txt\x00"))
	f.Add([]byte("R  new.txt\x00old.txt\x00?? untracked.txt\x00"))
	f.Add([]byte("C  copy.txt\x00orig.txt\x00"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parser should never panic, only return an error for invalid input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ParseZ panicked with input %q: %v", data, r)
			}
		}()
		ParseZ(bytes.NewReader(data))
	})
}
