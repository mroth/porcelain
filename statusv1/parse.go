package statusv1

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// Parse parses git status --porcelain=v1 output from an io.Reader.
//
// Headers: When using --branch in conjunction with git status --porcelain=v1,
// the output may contain header lines, for example, `## main...origin/main
// [ahead 1]`. These lines are preserved with ordering intact in the Headers
// field of the returned Status struct, but are not parsed as they are not
// documented as part of the --porcelain=v1 format.
//
// Path Handling: Paths containing special characters may be quoted by Git
// according to core.quotePath configuration. This function preserves paths
// exactly as provided by Git without unquoting. If your application needs
// unquoted paths, consider using [ParseZ] with the -z flag instead, as Git
// does not quote paths in -z format.
func Parse(r io.Reader) (*Status, error) {
	scanner := bufio.NewScanner(r)
	status := &Status{}

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue // skip empty lines
		}

		if bytes.HasPrefix(line, []byte("##")) {
			status.Headers = append(status.Headers, string(line))
			continue
		}

		entry, err := parseEntry(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line %q: %w", line, err)
		}

		status.Entries = append(status.Entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return status, nil
}

// ParseZ parses git status --porcelain=v1 -z output from an io.Reader.
//
// In the -z format, entries are terminated by NUL bytes instead of newlines,
// and rename entries have the format "XY to\x00from\x00" instead of "XY from -> to".
//
// Headers: When using --branch in conjunction with git status --porcelain=v1,
// the output may contain header lines, for example, `## main...origin/main
// [ahead 1]`. These lines are preserved with ordering intact in the Headers
// field of the returned Status struct, but are not parsed as they are not
// documented as part of the --porcelain=v1 format.
//
// Path Handling: In -z format, Git does not quote paths containing special
// characters, so all paths are provided as-is. This function preserves paths
// exactly as provided by Git.
func ParseZ(r io.Reader) (*Status, error) {
	scanner := newZScanner(r)
	status := &Status{}

	for scanner.Scan() {
		entry := scanner.Bytes()
		if len(entry) == 0 {
			continue // skip empty entries
		}

		if bytes.HasPrefix(entry, []byte("##")) {
			status.Headers = append(status.Headers, string(entry))
			continue
		}

		parsedEntry, err := parseEntryZ(entry)
		if err != nil {
			return nil, fmt.Errorf("failed to parse entry %q: %w", entry, err)
		}

		status.Entries = append(status.Entries, parsedEntry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return status, nil
}

// parseEntry parses a single line from git status --porcelain=v1 output.
// Format: "XY PATH" or "XY ORIG_PATH -> PATH"
func parseEntry(line []byte) (Entry, error) {
	if len(line) < 3 {
		return Entry{}, fmt.Errorf("line too short: %q", line)
	}

	// Parse XY status
	xy, err := parseXYFlag(line[:2])
	if err != nil {
		return Entry{}, err
	}

	// Skip the space after XY
	if line[2] != ' ' {
		return Entry{}, fmt.Errorf("expected space after XY status, got %q", line[2])
	}

	pathPart := line[3:]

	// Check for rename/copy format: "ORIG_PATH -> PATH"
	separator := []byte(" -> ")
	if origPath, newPath, found := bytes.Cut(pathPart, separator); found {
		// Check for empty parts
		if len(origPath) == 0 || len(newPath) == 0 {
			return Entry{}, fmt.Errorf("invalid rename format: %q", pathPart)
		}

		return Entry{
			XY:       xy,
			Path:     string(newPath),
			OrigPath: string(origPath),
		}, nil
	}

	// Regular format: just "PATH"
	return Entry{
		XY:   xy,
		Path: string(pathPart),
	}, nil
}

// parseEntryZ parses a single entry from git status --porcelain=v1 -z output.
// In -z format, rename entries contain both paths: "XY to\x00from".
func parseEntryZ(entry []byte) (Entry, error) {
	if len(entry) < 3 {
		return Entry{}, fmt.Errorf("entry too short: %q", entry)
	}

	// Parse XY status
	xy, err := parseXYFlag(entry[:2])
	if err != nil {
		return Entry{}, err
	}

	// Skip the space after XY
	if entry[2] != ' ' {
		return Entry{}, fmt.Errorf("expected space after XY status, got %q", entry[2])
	}

	pathPart := entry[3:]

	// For renames/copies in -z format, we have "to\x00from"
	// R or C can appear in either X or Y position
	if xy.X == Renamed || xy.X == Copied || xy.Y == Renamed || xy.Y == Copied {
		if newPath, origPath, found := bytes.Cut(pathPart, []byte{'\x00'}); found {
			// This is a rename: "to\x00from"
			return Entry{
				XY:       xy,
				Path:     string(newPath),
				OrigPath: string(origPath),
			}, nil
		}
		// If we don't have the NUL separator, it might be a malformed entry
		// but we'll treat it as just the new path
	}

	// Regular format: just the path (or malformed rename with only new path)
	return Entry{
		XY:   xy,
		Path: string(pathPart),
	}, nil
}

func parseXYFlag(field []byte) (XYFlag, error) {
	if len(field) != 2 {
		return XYFlag{}, fmt.Errorf("invalid XY field: expected 2 characters, got %d", len(field))
	}
	return XYFlag{X: State(field[0]), Y: State(field[1])}, nil
}
