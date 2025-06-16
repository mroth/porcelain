package statusv2

import (
	"errors"
	"io"
)

// Parse parses the output of `git status --porcelain=v2`.
//
// Additional status headers such as `# branch.*` and `# stash <N>` are parsed if present.
func Parse(input io.Reader) (*GitStatusV2, error) {
	return nil, errors.New("not implemented")
}

// ParseZ parses the output of `git status --porcelain=v2 -z`.
//
// -z output is a special mode that uses a NUL (ASCII 0x00) byte as line terminator instead of newline,
// as well as disabling any quoting of special characters in pathnames.
// func ParseZ(input io.Reader) (*GitStatusV2, error) {
// 	panic("not implemented")
// }
