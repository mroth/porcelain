package statusv1

import (
	"encoding"
	"testing"
)

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

func TestXYFlag_MarshalUnmarshalText(t *testing.T) {
	// enforce interface compliance
	var _ encoding.TextMarshaler = (*XYFlag)(nil)
	var _ encoding.TextUnmarshaler = (*XYFlag)(nil)

	tests := []struct {
		xy     XYFlag
		expect string
	}{
		{XYFlag{Unmodified, Modified}, " M"},
		{XYFlag{Added, Unmodified}, "A "},
		{XYFlag{Untracked, Untracked}, "??"},
		{XYFlag{Ignored, Ignored}, "!!"},
	}

	for _, tc := range tests {
		b, err := tc.xy.MarshalText()
		if err != nil {
			t.Errorf("MarshalText() error = %v", err)
		}
		if string(b) != tc.expect {
			t.Errorf("MarshalText() = %q, want %q", b, tc.expect)
		}

		var xy XYFlag
		err = xy.UnmarshalText([]byte(tc.expect))
		if err != nil {
			t.Errorf("UnmarshalText() error = %v", err)
		}
		if xy != tc.xy {
			t.Errorf("UnmarshalText() = %+v, want %+v", xy, tc.xy)
		}
	}

	// Test error case for UnmarshalText
	var xy XYFlag
	err := xy.UnmarshalText([]byte("A"))
	if err == nil {
		t.Errorf("UnmarshalText() should error for input of length != 2")
	}
}
