package statusv2

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// newZScanner creates a scanner that properly tokenizes git status --porcelain=v2 -z output.
// It handles the complex case where rename/copy entries (type "2") contain two NUL bytes:
// one as the path separator and another as the line terminator. Regular entries only
// have the line terminator NUL byte.
func newZScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(porcelainv2ZSplitFunc)
	return scanner
}

// porcelainv2ZSplitFunc is a custom split function that handles the dual NUL byte issue
// in porcelain v2 -z output. For rename/copy entries (starting with "2 "), it looks for
// the second NUL byte as the true line terminator, while for all other entries it uses
// the first NUL byte as the terminator.
func porcelainv2ZSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Look for first NUL byte
	firstNUL := bytes.IndexByte(data, '\x00')
	if firstNUL == -1 {
		if atEOF && len(data) > 0 {
			// No NUL found but we're at EOF, return remaining data
			return len(data), data, nil
		}
		// Need more data
		return 0, nil, nil
	}

	// Check if this is a rename/copy entry (starts with "2 ")
	if len(data) >= 2 && data[0] == '2' && data[1] == ' ' {
		// Look for the second NUL byte (the line terminator)
		secondNUL := bytes.IndexByte(data[firstNUL+1:], '\x00')
		if secondNUL == -1 {
			if atEOF {
				// At EOF with only one NUL - check if we have both paths
				if firstNUL+1 < len(data) {
					// We have data after the first NUL, treat as second path
					return len(data), data, nil
				} else {
					// Only one path, this is corruption
					return 0, nil, fmt.Errorf("malformed rename/copy entry: missing second path")
				}
			}
			// Need more data to find the second NUL
			return 0, nil, nil
		}

		// Return the entire rename/copy entry including the internal NUL separator
		totalLen := firstNUL + 1 + secondNUL + 1
		return totalLen, data[:firstNUL+1+secondNUL], nil
	}

	// Normal case: return up to first NUL
	return firstNUL + 1, data[:firstNUL], nil
}
