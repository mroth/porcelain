package statusv2

import (
	"testing"
)

// Fuzz test for parseChanged function
func FuzzParseChanged(f *testing.F) {
	// Add some seed inputs
	f.Add([]byte("1 M. N... 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 1234567890abcdef1234567890abcdef12345678 file_changed.txt"))
	f.Add([]byte("1 A. N... 000000 100644 100644 0000000000000000000000000000000000000000 fa49b077972391ad58037050f2a75f74e3671e92 file_added.txt"))
	f.Add([]byte("1 D. N... 100644 000000 000000 1234567890abcdef1234567890abcdef12345678 0000000000000000000000000000000000000000 file_deleted.txt"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parser should never panic, only return an error for invalid input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseChanged panicked with input %q: %v", data, r)
			}
		}()
		parseChangedEntry(data)
	})
}

// Fuzz test for parseRenameOrCopy function
func FuzzParseRenameOrCopy(f *testing.F) {
	// Add some seed inputs with both tab and NUL separators (including mismatches)
	f.Add([]byte("2 C. N... 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 1234567890abcdef1234567890abcdef12345678 C75 file_copied.txt\tfile_source.txt"), byte('\t'))
	f.Add([]byte("2 C. N... 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 1234567890abcdef1234567890abcdef12345678 C75 file_copied.txt\tfile_source.txt"), byte('\x00'))
	f.Add([]byte("2 R. N... 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 1234567890abcdef1234567890abcdef12345678 R100 file_renamed.txt\x00file_original.txt"), byte('\x00'))
	f.Add([]byte("2 R. N... 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 1234567890abcdef1234567890abcdef12345678 R100 file_renamed.txt\x00file_original.txt"), byte('\t'))

	f.Fuzz(func(t *testing.T, data []byte, sep byte) {
		// Parser should never panic, only return an error for invalid input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseRenameOrCopy panicked with input %q, sep %q: %v", data, sep, r)
			}
		}()
		parseRenameOrCopyEntry(data, renamePathSep(sep))
	})
}

// Fuzz test for parseUnmerged function
func FuzzParseUnmerged(f *testing.F) {
	// Add some seed inputs
	f.Add([]byte("u UU N... 100644 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 abcdef1234567890abcdef1234567890abcdef12 fedcba0987654321fedcba0987654321fedcba09 merge_conflict.txt"))
	f.Add([]byte("u DD N... 100644 000000 000000 000000 1234567890abcdef1234567890abcdef12345678 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 deleted_by_both.txt"))
	f.Add([]byte("u AU N... 000000 100644 000000 100644 0000000000000000000000000000000000000000 1234567890abcdef1234567890abcdef12345678 0000000000000000000000000000000000000000 added_by_us.txt"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parser should never panic, only return an error for invalid input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseUnmerged panicked with input %q: %v", data, r)
			}
		}()
		parseUnmergedEntry(data)
	})
}

// Fuzz test for parseUntracked function
func FuzzParseUntracked(f *testing.F) {
	// Add some seed inputs
	f.Add([]byte("? file_untracked.txt"))
	f.Add([]byte("? path/to/untracked.txt"))
	f.Add([]byte("? untracked with spaces.txt"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parser should never panic, only return an error for invalid input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseUntracked panicked with input %q: %v", data, r)
			}
		}()
		parseUntrackedEntry(data)
	})
}

// Fuzz test for parseIgnored function
func FuzzParseIgnored(f *testing.F) {
	// Add some seed inputs
	f.Add([]byte("! file_ignored.txt"))
	f.Add([]byte("! build/ignored.o"))
	f.Add([]byte("! .DS_Store"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parser should never panic, only return an error for invalid input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseIgnored panicked with input %q: %v", data, r)
			}
		}()
		parseIgnoredEntry(data)
	})
}

// Fuzz test for parseHeader function
func FuzzParseHeader(f *testing.F) {
	// Add some seed inputs
	f.Add(sampleHeaderBranchOID)
	f.Add(sampleHeaderBranchHead)
	f.Add(sampleHeaderBranchUpstream)
	f.Add(sampleHeaderBranchAB)
	f.Add(sampleHeaderStash)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parser should never panic, only return without action on invalid input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseHeader panicked with input %q: %v", data, r)
			}
		}()
		var s Status
		parseHeaderEntry(data, &s)
	})
}
