package statusv1

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Sample porcelain=v1 lines
var (
	sampleHeaderBranch      = []byte("## main...origin/main [ahead 1]")
	sampleHeaderUnknown     = []byte("## unknown header line")
	sampleEntryModified     = []byte(" M file_modified.txt")
	sampleEntryAdded        = []byte("A  file_added.txt")
	sampleEntryDeleted      = []byte("D  file_deleted.txt")
	sampleEntryRenamed      = []byte("R  file_original.txt -> file_renamed.txt")
	sampleEntryCopied       = []byte("C  file_original.txt -> file_copied.txt")
	sampleEntryUntracked    = []byte("?? file_untracked.txt")
	sampleEntryIgnored      = []byte("!! file_ignored.txt")
	sampleEntryBothModified = []byte("MM file_both_modified.txt")
)

// Some entries that differ in -z format for benchmarking ParseZ
var (
	sampleEntryRenamedZ = []byte("R  file_renamed.txt\x00file_original.txt")
	sampleEntryCopiedZ  = []byte("C  file_copied.txt\x00file_original.txt")
)

// samplePorcelainV1Output is a contrived sample output of:
// `git status --porcelain=v1 --branch`.
// It contains one entry for each common status type, a branch header, and a junk header to test parsing.
var samplePorcelainV1Output = bytes.Join([][]byte{
	sampleHeaderBranch,
	sampleHeaderUnknown,
	sampleEntryModified,
	sampleEntryAdded,
	sampleEntryDeleted,
	sampleEntryRenamed,
	sampleEntryCopied,
	sampleEntryUntracked,
	sampleEntryIgnored,
	sampleEntryBothModified,
}, []byte("\n"))

// samplePorcelainV1ZOutput is a contrived sample output of:
// `git status --porcelain=v1 --branch -z`.
// It contains one entry for each common status type, a branch header, and a junk header to test parsing.
// It should parse to the same result as samplePorcelainV1Output, but uses NUL terminators and the -z format for renames/copies.
var samplePorcelainV1ZOutput = bytes.Join([][]byte{
	sampleHeaderBranch,
	sampleHeaderUnknown,
	sampleEntryModified,
	sampleEntryAdded,
	sampleEntryDeleted,
	sampleEntryRenamedZ,
	sampleEntryCopiedZ,
	sampleEntryUntracked,
	sampleEntryIgnored,
	sampleEntryBothModified,
}, []byte("\x00"))

var sampleParsedStatus = Status{
	Headers: []string{
		"## main...origin/main [ahead 1]",
		"## unknown header line",
	},
	Entries: []Entry{
		{XY: XYFlag{Unmodified, Modified}, Path: "file_modified.txt"},
		{XY: XYFlag{Added, Unmodified}, Path: "file_added.txt"},
		{XY: XYFlag{Deleted, Unmodified}, Path: "file_deleted.txt"},
		{XY: XYFlag{Renamed, Unmodified}, Path: "file_renamed.txt", OrigPath: "file_original.txt"},
		{XY: XYFlag{Copied, Unmodified}, Path: "file_copied.txt", OrigPath: "file_original.txt"},
		{XY: XYFlag{Untracked, Untracked}, Path: "file_untracked.txt"},
		{XY: XYFlag{Ignored, Ignored}, Path: "file_ignored.txt"},
		{XY: XYFlag{Modified, Modified}, Path: "file_both_modified.txt"},
	},
}

func TestParse(t *testing.T) {
	r := bytes.NewReader(samplePorcelainV1Output)
	want := &sampleParsedStatus
	got, err := Parse(r)
	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
	}

}

func TestParseZ(t *testing.T) {
	r := bytes.NewReader(samplePorcelainV1ZOutput)
	want := &sampleParsedStatus
	got, err := ParseZ(r)
	if err != nil {
		t.Errorf("ParseZ() error = %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParseZ() mismatch (-want +got):\n%s", diff)
	}
}

func Test_parseEntry(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		want    Entry
		wantErr bool
	}{
		{
			name:  "modified in worktree",
			input: " M file.txt",
			want:  Entry{XY: XYFlag{Unmodified, Modified}, Path: "file.txt"},
		},
		{
			name:  "added to index",
			input: "A  file.txt",
			want:  Entry{XY: XYFlag{Added, Unmodified}, Path: "file.txt"},
		},
		{
			name:  "renamed",
			input: "R  old.txt -> new.txt",
			want:  Entry{XY: XYFlag{Renamed, Unmodified}, Path: "new.txt", OrigPath: "old.txt"},
		},
		{
			name:  "copied",
			input: "C  orig.txt -> copy.txt",
			want:  Entry{XY: XYFlag{Copied, Unmodified}, Path: "copy.txt", OrigPath: "orig.txt"},
		},
		{
			name:  "untracked",
			input: "?? untracked.txt",
			want:  Entry{XY: XYFlag{Untracked, Untracked}, Path: "untracked.txt"},
		},
		{
			name:  "ignored",
			input: "!! ignored.txt",
			want:  Entry{XY: XYFlag{Ignored, Ignored}, Path: "ignored.txt"},
		},
		{
			name:  "quoted path",
			input: "A  \"path with spaces.txt\"",
			want:  Entry{XY: XYFlag{Added, Unmodified}, Path: "\"path with spaces.txt\""},
		},
		{
			name:  "quoted path with escape",
			input: "A  \"path\\nwith\\nnewline.txt\"",
			want:  Entry{XY: XYFlag{Added, Unmodified}, Path: "\"path\\nwith\\nnewline.txt\""},
		},
		{
			name:  "renamed with quoted paths",
			input: "R  \"old path.txt\" -> \"new path.txt\"",
			want:  Entry{XY: XYFlag{Renamed, Unmodified}, Path: "\"new path.txt\"", OrigPath: "\"old path.txt\""},
		},
		// Edge cases
		{
			name:    "line too short",
			input:   "M",
			wantErr: true,
		},
		{
			name:    "missing space after XY",
			input:   "M file.txt",
			wantErr: true,
		},
		{
			name:    "invalid rename format - empty target",
			input:   "R  old.txt -> ",
			wantErr: true,
		},
		{
			name:    "invalid rename format - empty source",
			input:   "R   -> new.txt",
			wantErr: true,
		},
		{
			name:  "malformed rename - missing arrow (treated as path)",
			input: "R  old.txt new.txt",
			want:  Entry{XY: XYFlag{Renamed, Unmodified}, Path: "old.txt new.txt"},
			// this is likely an error in reality, but as per porcelain=v1 spec,
			// it is parsed as a valid path with no original path.
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseEntry([]byte(tc.input))
			if (err != nil) != tc.wantErr {
				t.Errorf("parseEntry() error = %v, wantErr %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("parseEntry() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_parseEntryZ(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		want    Entry
		wantErr bool
	}{
		{
			name:  "modified in worktree",
			input: " M file.txt",
			want:  Entry{XY: XYFlag{Unmodified, Modified}, Path: "file.txt"},
		},
		{
			name:  "added to index",
			input: "A  file.txt",
			want:  Entry{XY: XYFlag{Added, Unmodified}, Path: "file.txt"},
		},
		{
			name:  "renamed in X position (-z format)",
			input: "R  new.txt\x00old.txt",
			want:  Entry{XY: XYFlag{Renamed, Unmodified}, Path: "new.txt", OrigPath: "old.txt"},
		},
		{
			name:  "copied in X position (-z format)",
			input: "C  copy.txt\x00orig.txt",
			want:  Entry{XY: XYFlag{Copied, Unmodified}, Path: "copy.txt", OrigPath: "orig.txt"},
		},
		{
			name:  "renamed in Y position (-z format)",
			input: " R new.txt\x00old.txt",
			want:  Entry{XY: XYFlag{Unmodified, Renamed}, Path: "new.txt", OrigPath: "old.txt"},
		},
		{
			name:  "copied in Y position (-z format)",
			input: " C copy.txt\x00orig.txt",
			want:  Entry{XY: XYFlag{Unmodified, Copied}, Path: "copy.txt", OrigPath: "orig.txt"},
		},
		{
			name:  "untracked",
			input: "?? untracked.txt",
			want:  Entry{XY: XYFlag{Untracked, Untracked}, Path: "untracked.txt"},
		},
		{
			name:  "path with spaces (no quoting in -z)",
			input: "A  path with spaces.txt",
			want:  Entry{XY: XYFlag{Added, Unmodified}, Path: "path with spaces.txt"},
		},
		// Edge cases
		{
			name:    "entry too short",
			input:   "M",
			wantErr: true,
		},
		{
			name:    "missing space after XY",
			input:   "M file.txt",
			wantErr: true,
		},
		{
			name:  "rename entry with missing second path (treated as path)",
			input: "R  new.txt\x00",
			want:  Entry{XY: XYFlag{Renamed, Unmodified}, Path: "new.txt"},
		},
		{
			name:  "empty original path with NUL terminator",
			input: "R  new.txt\x00\x00",
			want:  Entry{XY: XYFlag{Renamed, Unmodified}, Path: "new.txt", OrigPath: "\x00"},
		},
		{
			name:    "missing space after XY with tab",
			input:   "M\tfile.txt",
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseEntryZ([]byte(tc.input))
			if (err != nil) != tc.wantErr {
				t.Errorf("parseEntryZ() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("parseEntryZ() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_parseXYFlag(t *testing.T) {
	testcases := []struct {
		name     string
		input    []byte
		expected XYFlag
		wantErr  bool
	}{
		{
			name:     "valid modified",
			input:    []byte(" M"),
			expected: XYFlag{X: Unmodified, Y: Modified},
		},
		{
			name:     "valid added",
			input:    []byte("A "),
			expected: XYFlag{X: Added, Y: Unmodified},
		},
		{
			name:     "valid untracked",
			input:    []byte("??"),
			expected: XYFlag{X: Untracked, Y: Untracked},
		},
		{
			name:    "too short",
			input:   []byte("M"),
			wantErr: true,
		},
		{
			name:    "too long",
			input:   []byte("ABC"),
			wantErr: true,
		},
		{
			name:    "empty",
			input:   []byte(""),
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseXYFlag(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseXYFlag() error = %v, wantErr %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("parseXYFlag() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
