package statusv1

import (
	"strings"
	"testing"
)

func Test_newZScanner(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "mixed entries with rename",
			input: " M file1.txt\x00R  new.txt\x00old.txt\x00A  file2.txt\x00",
			want: []string{
				" M file1.txt",
				"R  new.txt\x00old.txt",
				"A  file2.txt",
			},
		},
		{
			name:  "multiple renames",
			input: "R  new1.txt\x00old1.txt\x00C  copy.txt\x00orig.txt\x00",
			want: []string{
				"R  new1.txt\x00old1.txt",
				"C  copy.txt\x00orig.txt",
			},
		},
		{
			name:  "regular entries only",
			input: " M file1.txt\x00A  file2.txt\x00?? untracked.txt\x00",
			want: []string{
				" M file1.txt",
				"A  file2.txt",
				"?? untracked.txt",
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  []string{},
		},
		{
			name:  "single rename at EOF without final NUL",
			input: "R  new.txt\x00old.txt",
			want: []string{
				"R  new.txt\x00old.txt",
			},
		},
		{
			name:    "corrupted rename at EOF - missing second path",
			input:   "R  new.txt\x00",
			want:    []string{}, // should error, not return anything
			wantErr: true,
		},
		{
			name:    "corrupted copy at EOF - empty second path",
			input:   "C  new.txt\x00\x00",
			want:    []string{"C  new.txt\x00"},
			wantErr: false, // This should be handled as valid (empty original path)
		},
		{
			name:  "rename in Y position",
			input: " R new.txt\x00old.txt\x00",
			want: []string{
				" R new.txt\x00old.txt",
			},
		},
		{
			name:  "copy in Y position",
			input: " C copy.txt\x00orig.txt\x00",
			want: []string{
				" C copy.txt\x00orig.txt",
			},
		},
		{
			name:  "mixed entries with Y position rename",
			input: " M file1.txt\x00 R new.txt\x00old.txt\x00A  file2.txt\x00",
			want: []string{
				" M file1.txt",
				" R new.txt\x00old.txt",
				"A  file2.txt",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			scanner := newZScanner(strings.NewReader(tc.input))
			var got []string

			for scanner.Scan() {
				token := string(scanner.Bytes())
				if token != "" {
					got = append(got, token)
				}
			}

			if err := scanner.Err(); err != nil {
				if !tc.wantErr {
					t.Fatalf("scanner error: %v", err)
				}
				return
			}

			if tc.wantErr {
				t.Fatalf("expected error, got none")
			}

			if len(got) != len(tc.want) {
				t.Errorf("got %d tokens, want %d: %v vs %v", len(got), len(tc.want), got, tc.want)
				return
			}

			for i, want := range tc.want {
				if got[i] != want {
					t.Errorf("token %d: got %q, want %q", i, got[i], want)
				}
			}
		})
	}
}
