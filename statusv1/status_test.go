package statusv1

import "testing"

func TestXYFlag_String(t *testing.T) {
	testcases := []struct {
		name     string
		xy       XYFlag
		expected string
	}{
		{
			name:     "modified in working tree",
			xy:       XYFlag{X: Unmodified, Y: Modified},
			expected: " M",
		},
		{
			name:     "added to index",
			xy:       XYFlag{X: Added, Y: Unmodified},
			expected: "A ",
		},
		{
			name:     "renamed",
			xy:       XYFlag{X: Renamed, Y: Unmodified},
			expected: "R ",
		},
		{
			name:     "untracked",
			xy:       XYFlag{X: Untracked, Y: Untracked},
			expected: "??",
		},
		{
			name:     "ignored",
			xy:       XYFlag{X: Ignored, Y: Ignored},
			expected: "!!",
		},
		{
			name:     "modified both",
			xy:       XYFlag{X: Modified, Y: Modified},
			expected: "MM",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.xy.String()
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}
