package statusv1

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// newZScanner creates a scanner that tokenizes git status --porcelain=v1 -z
// output, returning each entry as a token, omitting the NUL byte that serves as
// the line terminator.
//
// It handles the complex case for rename/copy entries (XY status codes where
// X is 'R' or 'C') which contain two NUL bytes: one as the path separator and
// another as the line terminator. Regular entries only have the line terminator
// NUL byte.
func newZScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(porcelainv1ZSplitFunc)
	return scanner
}

// porcelainv1ZSplitFunc is a custom [bufio.SplitFunc] that handles the dual NUL byte issue
// in porcelain v1 -z output. For rename/copy entries (where the first character of the XY
// status is 'R' or 'C'), it looks for the second NUL byte as the true line terminator,
// while for all other entries it uses the first NUL byte as the terminator.
func porcelainv1ZSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Look for first NUL byte. For rename/copy entries, this will be the path
	// separator, and for all other entries, this is the entry terminator.
	firstNUL := bytes.IndexByte(data, '\x00')
	if firstNUL == -1 {
		if atEOF && len(data) > 0 {
			// No NUL found but we're at EOF, return remaining data
			return len(data), data, nil
		}
		// Need more data
		return 0, nil, nil
	}

	// Check if this is a rename/copy entry (XY status where X or Y is 'R' or 'C')
	// Format: "XY path\x00origpath\x00"
	if len(data) >= 2 && (data[0] == 'R' || data[0] == 'C' || data[1] == 'R' || data[1] == 'C') {
		// Look for the second NUL byte (the line terminator)
		secondNUL := bytes.IndexByte(data[firstNUL+1:], '\x00')
		if secondNUL == -1 {
			if atEOF {
				// At EOF with only one NUL - check if we have both paths
				if firstNUL+1 < len(data) {
					// We have data after the first NUL, treat as second path
					return len(data), data, nil
				}
				// Only one path, this is invalid for this entry type
				return 0, nil, fmt.Errorf("malformed rename/copy entry: missing original path")
			}
			// Need more data to find the second NUL
			return 0, nil, nil
		}

		// Return the entire rename/copy entry including the internal NUL path
		// separator, advancing past the second NUL byte entry terminator.
		totalLength := firstNUL + 1 + secondNUL
		return totalLength + 1, data[:totalLength], nil
	}

	// Normal case: return up to first NUL as the token,
	// advancing the scanner past the entry terminator.
	return firstNUL + 1, data[:firstNUL], nil
}
