package statusv2

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// samplePorcelainV2Output is a contrived sample output of `git status --porcelain=v2 --branch --show-status`.
// It contains branch information, a stash entry, and one changed file entry for each
// of the EntryType variants: Changed, RenameOrCopy, Unmerged, Untracked, and Ignored.
const samplePorcelainV2Output = `# comment skipped non-standard header
# branch.oid 34064be349d4a03ed158aba170d8d2db6ff9e3e0
# branch.head main
# branch.upstream origin/main
# branch.ab +6 -3
# stash 3
# branch.upstream origin/main
1 M. N... 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 1234567890abcdef1234567890abcdef12345678 file_changed.txt
2 R. N... 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 1234567890abcdef1234567890abcdef12345678 R100 file_renamed.txt	file_original.txt
1 U. N... 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 1234567890abcdef1234567890abcdef12345678 file_unmerged.txt
? file_untracked.txt
! file_ignored.txt
`

func TestParse(t *testing.T) {
	r := strings.NewReader(samplePorcelainV2Output)
	got, err := Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	want := &Status{
		Branch: &BranchInfo{
			OID:      "34064be349d4a03ed158aba170d8d2db6ff9e3e0",
			Head:     "main",
			Upstream: "origin/main",
			Ahead:    6,
			Behind:   3,
		},
		Stash: &StashInfo{Count: 3},
		Entries: []Entry{
			ChangedEntry{
				XY:    XYFlag{Modified, Unmodified},
				Sub:   SubmoduleStatus{IsSubmodule: false},
				ModeH: 0100644,
				ModeI: 0100644,
				ModeW: 0100644,
				HashH: "1234567890abcdef1234567890abcdef12345678",
				HashI: "1234567890abcdef1234567890abcdef12345678",
				Path:  "file_changed.txt",
			},
			RenameOrCopyEntry{
				XY:    XYFlag{Renamed, Unmodified},
				Sub:   SubmoduleStatus{IsSubmodule: false},
				ModeH: 0100644,
				ModeI: 0100644,
				ModeW: 0100644,
				HashH: "1234567890abcdef1234567890abcdef12345678",
				HashI: "1234567890abcdef1234567890abcdef12345678",
				Score: "R100",
				Path:  "file_renamed.txt",
				Orig:  "file_original.txt",
			},
			ChangedEntry{
				XY:    XYFlag{UpdatedUnmerged, Unmodified},
				Sub:   SubmoduleStatus{IsSubmodule: false},
				ModeH: 0100644,
				ModeI: 0100644,
				ModeW: 0100644,
				HashH: "1234567890abcdef1234567890abcdef12345678",
				HashI: "1234567890abcdef1234567890abcdef12345678",
				Path:  "file_unmerged.txt",
			},
			UntrackedEntry{
				Path: "file_untracked.txt",
			},
			IgnoredEntry{
				Path: "file_ignored.txt",
			},
		},
	}
	if cmp.Diff(want, got) != "" {
		t.Errorf("Parse() mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestParseZ(t *testing.T) {
	// Test ParseZ with basic functionality
	input := "1 M. N... 100644 100644 100644 hash1 hash2 modified.txt\x00" +
		"2 R. N... 100644 100644 100644 hash1 hash2 R100 newpath.txt\x00oldpath.txt\x00" +
		"? untracked.txt\x00"

	got, err := ParseZ(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseZ() error = %v", err)
	}
	want := &Status{
		Entries: []Entry{
			ChangedEntry{
				XY:    XYFlag{Modified, Unmodified},
				Sub:   SubmoduleStatus{IsSubmodule: false},
				ModeH: 0100644,
				ModeI: 0100644,
				ModeW: 0100644,
				HashH: "hash1",
				HashI: "hash2",
				Path:  "modified.txt",
			},
			RenameOrCopyEntry{
				XY:    XYFlag{Renamed, Unmodified},
				Sub:   SubmoduleStatus{IsSubmodule: false},
				ModeH: 0100644,
				ModeI: 0100644,
				ModeW: 0100644,
				HashH: "hash1",
				HashI: "hash2",
				Score: "R100",
				Path:  "newpath.txt",
				Orig:  "oldpath.txt",
			},
			UntrackedEntry{
				Path: "untracked.txt",
			},
		},
	}
	if cmp.Diff(want, got) != "" {
		t.Errorf("ParseZ() mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

// TestParseGolden tests the Parse function with various test cases.
// Each test case specifies a file containing the output of `git status --porcelain=v2`.
// Files should be placed in the "testdata" directory.
func TestParseGolden(t *testing.T) {
	testcases := []struct {
		file    string
		desc    string // optional human readable description
		want    *Status
		wantErr bool
	}{}

	for _, tc := range testcases {
		t.Run(tc.file, func(t *testing.T) {
			f, err := os.Open("testdata/" + tc.file)
			if err != nil {
				t.Fatalf("fatal: failed to open test file %q: %v", tc.file, err)
			}
			defer f.Close()

			got, err := Parse(f)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ParseFile() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestParseChanged tests the parseChanged function with various valid and invalid inputs.
func TestParseChanged(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		want    ChangedEntry
		wantErr bool
	}{
		{
			name:  "simple added file",
			input: "1 A. N... 000000 100644 100644 0000000000000000000000000000000000000000 fa49b077972391ad58037050f2a75f74e3671e92 file_add_clean.txt",
			want: ChangedEntry{
				XY:    XYFlag{Added, Unmodified},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				ModeH: FileMode(0),
				ModeI: FileMode(0100644),
				ModeW: FileMode(0100644),
				HashH: "0000000000000000000000000000000000000000",
				HashI: "fa49b077972391ad58037050f2a75f74e3671e92",
				Path:  "file_add_clean.txt",
			},
			wantErr: false,
		},
		{
			name:  "modified file in both index and worktree",
			input: "1 MM N... 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 543f44d8a781da3a5623de35c3e20b21df7c4557 file_add_modify.txt",
			want: ChangedEntry{
				XY:    XYFlag{Modified, Modified},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				ModeH: FileMode(0100644),
				ModeI: FileMode(0100644),
				ModeW: FileMode(0100644),
				HashH: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				HashI: "543f44d8a781da3a5623de35c3e20b21df7c4557",
				Path:  "file_add_modify.txt",
			},
			wantErr: false,
		},
		{
			name:  "type changed file",
			input: "1 MT N... 100644 100644 120000 f2376e2bab6c5194410bd8a55630f83f933d2f34 fe1ffd21b578c50773a92520eccf43ddd258f530 file_add_type.txt",
			want: ChangedEntry{
				XY:    XYFlag{Modified, TypeChanged},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				ModeH: FileMode(0100644),
				ModeI: FileMode(0100644),
				ModeW: FileMode(0120000),
				HashH: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				HashI: "fe1ffd21b578c50773a92520eccf43ddd258f530",
				Path:  "file_add_type.txt",
			},
			wantErr: false,
		},
		{
			name:  "deleted file",
			input: "1 D. N... 100644 000000 000000 3706ca2490889bfa11c40ab8e9f8852a27182ffb 0000000000000000000000000000000000000000 file_delete_recreate_type.txt",
			want: ChangedEntry{
				XY:    XYFlag{Deleted, Unmodified},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				ModeH: FileMode(0100644),
				ModeI: FileMode(0),
				ModeW: FileMode(0),
				HashH: "3706ca2490889bfa11c40ab8e9f8852a27182ffb",
				HashI: "0000000000000000000000000000000000000000",
				Path:  "file_delete_recreate_type.txt",
			},
			wantErr: false,
		},
		{
			name:  "worktree deletion",
			input: "1 .D N... 100644 100644 000000 98c28f9a9834de8aa406c64935e72f5851fddcc3 98c28f9a9834de8aa406c64935e72f5851fddcc3 file_delete_worktree.txt",
			want: ChangedEntry{
				XY:    XYFlag{Unmodified, Deleted},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				ModeH: FileMode(0100644),
				ModeI: FileMode(0100644),
				ModeW: FileMode(0),
				HashH: "98c28f9a9834de8aa406c64935e72f5851fddcc3",
				HashI: "98c28f9a9834de8aa406c64935e72f5851fddcc3",
				Path:  "file_delete_worktree.txt",
			},
			wantErr: false,
		},
		{
			name:  "submodule with changes",
			input: "1 MM SCM. 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 543f44d8a781da3a5623de35c3e20b21df7c4557 submodule_path",
			want: ChangedEntry{
				XY:    XYFlag{Modified, Modified},
				Sub:   SubmoduleStatus{IsSubmodule: true, CommitChanged: true, HasModifications: true, HasUntracked: false},
				ModeH: FileMode(0100644),
				ModeI: FileMode(0100644),
				ModeW: FileMode(0100644),
				HashH: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				HashI: "543f44d8a781da3a5623de35c3e20b21df7c4557",
				Path:  "submodule_path",
			},
			wantErr: false,
		},
		{
			name:  "submodule with untracked files",
			input: "1 .M S..U 160000 160000 160000 abcdef1234567890abcdef1234567890abcdef12 1234567890abcdef1234567890abcdef12345678 submodule_untracked",
			want: ChangedEntry{
				XY:    XYFlag{Unmodified, Modified},
				Sub:   SubmoduleStatus{IsSubmodule: true, CommitChanged: false, HasModifications: false, HasUntracked: true},
				ModeH: FileMode(0160000),
				ModeI: FileMode(0160000),
				ModeW: FileMode(0160000),
				HashH: "abcdef1234567890abcdef1234567890abcdef12",
				HashI: "1234567890abcdef1234567890abcdef12345678",
				Path:  "submodule_untracked",
			},
			wantErr: false,
		},
		{
			name:    "invalid line - wrong prefix",
			input:   "2 A. N... 000000 100644 100644 0000000000000000000000000000000000000000 fa49b077972391ad58037050f2a75f74e3671e92 file.txt",
			wantErr: true,
		},
		{
			name:    "invalid line - too few fields",
			input:   "1 A. N... 000000 100644 100644 0000000000000000000000000000000000000000",
			wantErr: true,
		},
		{
			name:    "empty line",
			input:   "",
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseChanged([]byte(tc.input))
			if (err != nil) != tc.wantErr {
				t.Errorf("parseChanged() error = %v, wantErr %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("parseChanged() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestParseRenameOrCopied tests the parseRenameOrCopy function with various valid and invalid inputs.
func TestParseRenameOrCopied(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		want    RenameOrCopyEntry
		wantErr bool
	}{
		{
			name:  "simple rename",
			input: "2 R. N... 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 f2376e2bab6c5194410bd8a55630f83f933d2f34 R100 file_renamed_clean.txt\tfile_delete_index.txt",
			want: RenameOrCopyEntry{
				XY:    XYFlag{Renamed, Unmodified},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				ModeH: FileMode(0100644),
				ModeI: FileMode(0100644),
				ModeW: FileMode(0100644),
				HashH: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				HashI: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Score: "R100",
				Path:  "file_renamed_clean.txt",
				Orig:  "file_delete_index.txt",
			},
			wantErr: false,
		},
		{
			name:  "rename with deletion",
			input: "2 RD N... 100644 100644 000000 f2376e2bab6c5194410bd8a55630f83f933d2f34 f2376e2bab6c5194410bd8a55630f83f933d2f34 R100 file_renamed_delete.txt\tfile_rename_delete.txt",
			want: RenameOrCopyEntry{
				XY:    XYFlag{Renamed, Deleted},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				ModeH: FileMode(0100644),
				ModeI: FileMode(0100644),
				ModeW: FileMode(0),
				HashH: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				HashI: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Score: "R100",
				Path:  "file_renamed_delete.txt",
				Orig:  "file_rename_delete.txt",
			},
			wantErr: false,
		},
		{
			name:  "rename with modification",
			input: "2 RM N... 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 f2376e2bab6c5194410bd8a55630f83f933d2f34 R100 file_renamed_modify.txt\tfile_rename_modify.txt",
			want: RenameOrCopyEntry{
				XY:    XYFlag{Renamed, Modified},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				ModeH: FileMode(0100644),
				ModeI: FileMode(0100644),
				ModeW: FileMode(0100644),
				HashH: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				HashI: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Score: "R100",
				Path:  "file_renamed_modify.txt",
				Orig:  "file_rename_modify.txt",
			},
			wantErr: false,
		},
		{
			name:  "rename with type change",
			input: "2 RT N... 100644 100644 120000 f2376e2bab6c5194410bd8a55630f83f933d2f34 f2376e2bab6c5194410bd8a55630f83f933d2f34 R100 file_renamed_type.txt\tfile_rename_source.txt",
			want: RenameOrCopyEntry{
				XY:    XYFlag{Renamed, TypeChanged},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				ModeH: FileMode(0100644),
				ModeI: FileMode(0100644),
				ModeW: FileMode(0120000),
				HashH: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				HashI: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Score: "R100",
				Path:  "file_renamed_type.txt",
				Orig:  "file_rename_source.txt",
			},
			wantErr: false,
		},
		{
			name:  "copy with lower similarity score",
			input: "2 C. N... 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 a1b2c3d4e5f6789012345678901234567890abcd C75 copied_file.txt\toriginal_file.txt",
			want: RenameOrCopyEntry{
				XY:    XYFlag{Copied, Unmodified},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				ModeH: FileMode(0100644),
				ModeI: FileMode(0100644),
				ModeW: FileMode(0100644),
				HashH: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				HashI: "a1b2c3d4e5f6789012345678901234567890abcd",
				Score: "C75",
				Path:  "copied_file.txt",
				Orig:  "original_file.txt",
			},
			wantErr: false,
		},
		{
			name:  "submodule rename",
			input: "2 R. SCM. 160000 160000 160000 abcdef1234567890abcdef1234567890abcdef12 1234567890abcdef1234567890abcdef12345678 R100 submodule_new\tsubmodule_old",
			want: RenameOrCopyEntry{
				XY:    XYFlag{Renamed, Unmodified},
				Sub:   SubmoduleStatus{IsSubmodule: true, CommitChanged: true, HasModifications: true, HasUntracked: false},
				ModeH: FileMode(0160000),
				ModeI: FileMode(0160000),
				ModeW: FileMode(0160000),
				HashH: "abcdef1234567890abcdef1234567890abcdef12",
				HashI: "1234567890abcdef1234567890abcdef12345678",
				Score: "R100",
				Path:  "submodule_new",
				Orig:  "submodule_old",
			},
			wantErr: false,
		},
		{
			name:  "paths with special characters",
			input: "2 R. N... 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 f2376e2bab6c5194410bd8a55630f83f933d2f34 R100 path/with spaces/new.txt\tpath/with spaces/old.txt",
			want: RenameOrCopyEntry{
				XY:    XYFlag{Renamed, Unmodified},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				ModeH: FileMode(0100644),
				ModeI: FileMode(0100644),
				ModeW: FileMode(0100644),
				HashH: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				HashI: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Score: "R100",
				Path:  "path/with spaces/new.txt",
				Orig:  "path/with spaces/old.txt",
			},
			wantErr: false,
		},
		{
			name:    "invalid line - wrong prefix",
			input:   "1 R. N... 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 f2376e2bab6c5194410bd8a55630f83f933d2f34 R100 new.txt\told.txt",
			wantErr: true,
		},
		{
			name:    "invalid line - too few fields",
			input:   "2 R. N... 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 f2376e2bab6c5194410bd8a55630f83f933d2f34",
			wantErr: true,
		},
		{
			name:    "missing tab separator in paths",
			input:   "2 R. N... 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 f2376e2bab6c5194410bd8a55630f83f933d2f34 R100 new.txt old.txt",
			wantErr: true,
		},
		{
			name:    "empty line",
			input:   "",
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseRenameOrCopy([]byte(tc.input), tabSeparator)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseRenameOrCopy() error = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr {
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("parseRenameOrCopy() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// TestParseUnmerged tests the parseUnmerged function with various valid and invalid inputs.
func TestParseUnmerged(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		want    UnmergedEntry
		wantErr bool
	}{
		{
			name:  "both modified conflict",
			input: "u UU N... 100644 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 b3266c11446a04580631ad3edf7e20789dc477d0 0942ce73bfaae4c3356c727901d1b4b933cf7b88 merge_both_modified.txt",
			want: UnmergedEntry{
				XY:    XYFlag{UpdatedUnmerged, UpdatedUnmerged},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				Mode1: FileModeRegular,
				Mode2: FileModeRegular,
				Mode3: FileModeRegular,
				ModeW: FileModeRegular,
				Hash1: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Hash2: "b3266c11446a04580631ad3edf7e20789dc477d0",
				Hash3: "0942ce73bfaae4c3356c727901d1b4b933cf7b88",
				Path:  "merge_both_modified.txt",
			},
			wantErr: false,
		},
		{
			name:  "both deleted conflict",
			input: "u DD N... 100644 000000 000000 000000 f2376e2bab6c5194410bd8a55630f83f933d2f34 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 merge_both_deleted.txt",
			want: UnmergedEntry{
				XY:    XYFlag{Deleted, Deleted},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				Mode1: FileModeRegular,
				Mode2: FileModeEmpty,
				Mode3: FileModeEmpty,
				ModeW: FileModeEmpty,
				Hash1: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Hash2: "0000000000000000000000000000000000000000",
				Hash3: "0000000000000000000000000000000000000000",
				Path:  "merge_both_deleted.txt",
			},
			wantErr: false,
		},
		{
			name:  "added by us conflict",
			input: "u AU N... 100644 000000 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 0000000000000000000000000000000000000000 0942ce73bfaae4c3356c727901d1b4b933cf7b88 merge_added_by_us.txt",
			want: UnmergedEntry{
				XY:    XYFlag{Added, UpdatedUnmerged},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				Mode1: FileModeRegular,
				Mode2: FileModeEmpty,
				Mode3: FileModeRegular,
				ModeW: FileModeRegular,
				Hash1: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Hash2: "0000000000000000000000000000000000000000",
				Hash3: "0942ce73bfaae4c3356c727901d1b4b933cf7b88",
				Path:  "merge_added_by_us.txt",
			},
			wantErr: false,
		},
		{
			name:  "deleted by them conflict",
			input: "u UD N... 100644 100644 000000 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 f0e618c170ab07669cef49d4a84ced48a603ce0c 0000000000000000000000000000000000000000 merge_deleted_by_them.txt",
			want: UnmergedEntry{
				XY:    XYFlag{UpdatedUnmerged, Deleted},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				Mode1: FileModeRegular,
				Mode2: FileModeRegular,
				Mode3: FileModeEmpty,
				ModeW: FileModeRegular,
				Hash1: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Hash2: "f0e618c170ab07669cef49d4a84ced48a603ce0c",
				Hash3: "0000000000000000000000000000000000000000",
				Path:  "merge_deleted_by_them.txt",
			},
			wantErr: false,
		},
		{
			name:  "deleted by us conflict",
			input: "u DU N... 100644 000000 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 0000000000000000000000000000000000000000 7d987256966029afce5bde6f7eeca94d17f267d0 merge_deleted_by_us.txt",
			want: UnmergedEntry{
				XY:    XYFlag{Deleted, UpdatedUnmerged},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				Mode1: FileModeRegular,
				Mode2: FileModeEmpty,
				Mode3: FileModeRegular,
				ModeW: FileModeRegular,
				Hash1: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Hash2: "0000000000000000000000000000000000000000",
				Hash3: "7d987256966029afce5bde6f7eeca94d17f267d0",
				Path:  "merge_deleted_by_us.txt",
			},
			wantErr: false,
		},
		{
			name:  "both added conflict",
			input: "u AA N... 100644 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 b3266c11446a04580631ad3edf7e20789dc477d0 7d987256966029afce5bde6f7eeca94d17f267d0 merge_both_added.txt",
			want: UnmergedEntry{
				XY:    XYFlag{Added, Added},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				Mode1: FileModeRegular,
				Mode2: FileModeRegular,
				Mode3: FileModeRegular,
				ModeW: FileModeRegular,
				Hash1: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Hash2: "b3266c11446a04580631ad3edf7e20789dc477d0",
				Hash3: "7d987256966029afce5bde6f7eeca94d17f267d0",
				Path:  "merge_both_added.txt",
			},
			wantErr: false,
		},
		{
			name:  "submodule conflict",
			input: "u UU SCM. 160000 160000 160000 160000 abcdef1234567890abcdef1234567890abcdef12 1234567890abcdef1234567890abcdef12345678 fedcba0987654321fedcba0987654321fedcba09 submodule_conflict",
			want: UnmergedEntry{
				XY:    XYFlag{UpdatedUnmerged, UpdatedUnmerged},
				Sub:   SubmoduleStatus{IsSubmodule: true, CommitChanged: true, HasModifications: true, HasUntracked: false},
				Mode1: FileModeSubmodule,
				Mode2: FileModeSubmodule,
				Mode3: FileModeSubmodule,
				ModeW: FileModeSubmodule,
				Hash1: "abcdef1234567890abcdef1234567890abcdef12",
				Hash2: "1234567890abcdef1234567890abcdef12345678",
				Hash3: "fedcba0987654321fedcba0987654321fedcba09",
				Path:  "submodule_conflict",
			},
			wantErr: false,
		},
		{
			name:  "path with special characters",
			input: "u UU N... 100644 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 b3266c11446a04580631ad3edf7e20789dc477d0 0942ce73bfaae4c3356c727901d1b4b933cf7b88 path/with spaces/conflict.txt",
			want: UnmergedEntry{
				XY:    XYFlag{UpdatedUnmerged, UpdatedUnmerged},
				Sub:   SubmoduleStatus{IsSubmodule: false, CommitChanged: false, HasModifications: false, HasUntracked: false},
				Mode1: FileModeRegular,
				Mode2: FileModeRegular,
				Mode3: FileModeRegular,
				ModeW: FileModeRegular,
				Hash1: "f2376e2bab6c5194410bd8a55630f83f933d2f34",
				Hash2: "b3266c11446a04580631ad3edf7e20789dc477d0",
				Hash3: "0942ce73bfaae4c3356c727901d1b4b933cf7b88",
				Path:  "path/with spaces/conflict.txt",
			},
			wantErr: false,
		},
		{
			name:    "invalid line - wrong prefix",
			input:   "1 UU N... 100644 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 b3266c11446a04580631ad3edf7e20789dc477d0 0942ce73bfaae4c3356c727901d1b4b933cf7b88 file.txt",
			wantErr: true,
		},
		{
			name:    "invalid line - too few fields",
			input:   "u UU N... 100644 100644 100644 100644 f2376e2bab6c5194410bd8a55630f83f933d2f34 b3266c11446a04580631ad3edf7e20789dc477d0",
			wantErr: true,
		},
		{
			name:    "empty line",
			input:   "",
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseUnmerged([]byte(tc.input))
			if (err != nil) != tc.wantErr {
				t.Errorf("parseUnmerged() error = %v, wantErr %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("parseUnmerged() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseXYFlag(t *testing.T) {
	testcases := []struct {
		name    string
		input   []byte
		want    XYFlag
		wantErr bool
	}{
		{
			name:    "valid XY flag - both modified",
			input:   []byte("MM"),
			want:    XYFlag{Modified, Modified},
			wantErr: false,
		},
		{
			name:    "valid XY flag - added clean",
			input:   []byte("A."),
			want:    XYFlag{Added, Unmodified},
			wantErr: false,
		},
		{
			name:    "valid XY flag - deleted both",
			input:   []byte("DD"),
			want:    XYFlag{Deleted, Deleted},
			wantErr: false,
		},
		{
			name:    "valid XY flag - unmerged",
			input:   []byte("UU"),
			want:    XYFlag{UpdatedUnmerged, UpdatedUnmerged},
			wantErr: false,
		},
		{
			name:    "valid XY flag - type changed",
			input:   []byte("TT"),
			want:    XYFlag{TypeChanged, TypeChanged},
			wantErr: false,
		},
		{
			name:    "valid XY flag - renamed",
			input:   []byte("R."),
			want:    XYFlag{Renamed, Unmodified},
			wantErr: false,
		},
		{
			name:    "valid XY flag - copied",
			input:   []byte("C."),
			want:    XYFlag{Copied, Unmodified},
			wantErr: false,
		},
		{
			name:    "invalid XY flag - too short",
			input:   []byte("M"),
			wantErr: true,
		},
		{
			name:    "invalid XY flag - too long",
			input:   []byte("MMM"),
			wantErr: true,
		},
		{
			name:    "invalid XY flag - empty",
			input:   []byte(""),
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseXYFlag(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseXYFlag() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("parseXYFlag() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseSubmoduleStatus(t *testing.T) {
	testcases := []struct {
		name    string
		input   []byte
		want    SubmoduleStatus
		wantErr bool
	}{
		{
			name:  "not a submodule",
			input: []byte("N..."),
			want: SubmoduleStatus{
				IsSubmodule:      false,
				CommitChanged:    false,
				HasModifications: false,
				HasUntracked:     false,
			},
			wantErr: false,
		},
		{
			name:  "submodule commit changed",
			input: []byte("SC.."),
			want: SubmoduleStatus{
				IsSubmodule:      true,
				CommitChanged:    true,
				HasModifications: false,
				HasUntracked:     false,
			},
			wantErr: false,
		},
		{
			name:  "submodule with modifications",
			input: []byte("S.M."),
			want: SubmoduleStatus{
				IsSubmodule:      true,
				CommitChanged:    false,
				HasModifications: true,
				HasUntracked:     false,
			},
			wantErr: false,
		},
		{
			name:  "submodule with untracked files",
			input: []byte("S..U"),
			want: SubmoduleStatus{
				IsSubmodule:      true,
				CommitChanged:    false,
				HasModifications: false,
				HasUntracked:     true,
			},
			wantErr: false,
		},
		{
			name:  "submodule with all changes",
			input: []byte("SCMU"),
			want: SubmoduleStatus{
				IsSubmodule:      true,
				CommitChanged:    true,
				HasModifications: true,
				HasUntracked:     true,
			},
			wantErr: false,
		},
		{
			name:    "invalid length - too short",
			input:   []byte("N.."),
			wantErr: true,
		},
		{
			name:    "invalid length - too long",
			input:   []byte("N...."),
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   []byte(""),
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSubmoduleStatus(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseSubmoduleStatus() error = %v, wantErr %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("parseSubmoduleStatus() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseFileMode(t *testing.T) {
	testcases := []struct {
		name    string
		input   []byte
		want    FileMode
		wantErr bool
	}{
		{
			name:    "regular file",
			input:   []byte("100644"),
			want:    FileModeRegular,
			wantErr: false,
		},
		{
			name:    "executable file",
			input:   []byte("100755"),
			want:    FileModeExecutable,
			wantErr: false,
		},
		{
			name:    "symbolic link",
			input:   []byte("120000"),
			want:    FileModeSymlink,
			wantErr: false,
		},
		{
			name:    "submodule",
			input:   []byte("160000"),
			want:    FileModeSubmodule,
			wantErr: false,
		},
		{
			name:    "directory",
			input:   []byte("040000"),
			want:    FileModeDir,
			wantErr: false,
		},
		{
			name:    "empty/zero",
			input:   []byte("000000"),
			want:    FileModeEmpty,
			wantErr: false,
		},
		{
			name:    "invalid non-octal characters",
			input:   []byte("10064g"),
			wantErr: true,
		},
		{
			name:    "invalid decimal number",
			input:   []byte("100648"),
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   []byte(""),
			wantErr: true,
		},
		{
			name:    "non-numeric input",
			input:   []byte("abcdef"),
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseFileMode(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseFileMode() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("parseFileMode() = %o, want %o", got, tc.want)
			}
		})
	}
}
