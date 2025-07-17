package statusv2

import "testing"

func TestXYFlag_Accessors(t *testing.T) {
	flags := XYFlag{Modified, Unmodified}
	wantX, wantY := Modified, Unmodified
	x, y := flags.X, flags.Y
	if x != wantX {
		t.Errorf("Expected X to be %v, got %v", wantX, x)
	}
	if y != wantY {
		t.Errorf("Expected Y to be %v, got %v", wantY, y)
	}
}

func TestXYFlag_String(t *testing.T) {
	testcases := []struct {
		name     string
		xy       XYFlag
		expected string
	}{
		{
			xy:       XYFlag{Modified, Unmodified},
			expected: "M.",
		},
		{
			xy:       XYFlag{Added, Deleted},
			expected: "AD",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.xy.String()
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestFileMode_String(t *testing.T) {
	testcases := []struct {
		name     string
		mode     FileMode
		expected string
	}{
		{
			name:     "regular file",
			mode:     FileModeRegular,
			expected: "100644",
		},
		{
			name:     "directory",
			mode:     FileModeDir,
			expected: "40000",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.mode.String()
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestSubmoduleStatus_String(t *testing.T) {
	testcases := []struct {
		name     string
		status   SubmoduleStatus
		expected string
	}{
		{
			name:     "not a submodule",
			status:   SubmoduleStatus{},
			expected: "N...",
		},
		{
			name: "commit changed",
			status: SubmoduleStatus{
				IsSubmodule:      true,
				CommitChanged:    true,
				HasModifications: false,
				HasUntracked:     false,
			},
			expected: "SC..",
		},
		{
			name: "has modifications",
			status: SubmoduleStatus{
				IsSubmodule:      true,
				CommitChanged:    false,
				HasModifications: true,
				HasUntracked:     false,
			},
			expected: "S.M.",
		},
		{
			name: "has modifications",
			status: SubmoduleStatus{
				IsSubmodule:      true,
				CommitChanged:    false,
				HasModifications: true,
				HasUntracked:     false,
			},
			expected: "S.M.",
		},
		{
			name: "all fields",
			status: SubmoduleStatus{
				IsSubmodule:      true,
				CommitChanged:    true,
				HasModifications: true,
				HasUntracked:     true,
			},
			expected: "SCMU",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.status.String()
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestEntry_Type(t *testing.T) {
	testcases := []struct {
		entry     Entry
		entryType EntryType
	}{
		{ChangedEntry{}, EntryTypeChanged},
		{RenameOrCopyEntry{}, EntryTypeRenameOrCopy},
		{UnmergedEntry{}, EntryTypeUnmerged},
		{UntrackedEntry{}, EntryTypeUntracked},
		{IgnoredEntry{}, EntryTypeIgnored},
	}

	for _, tc := range testcases {
		result := tc.entry.Type()
		if result != tc.entryType {
			t.Errorf("Expected %v, got %v", tc.entryType, result)
		}
	}
}
