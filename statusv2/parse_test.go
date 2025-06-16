package statusv2

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestParse tests the Parse function with various test cases.
// Each test case specifies a file containing the output of `git status --porcelain=v2`.
// Files should be placed in the "testdata" directory.
func TestParse(t *testing.T) {
	testcases := []struct {
		file    string
		desc    string // optional human readable description
		want    *GitStatusV2
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
