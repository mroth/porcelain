package statusv2

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strconv"
)

// Enable debug logging by setting this to a vaild *slog.Logger
var debugLogger = slog.New(slog.DiscardHandler)

// Parse parses the output of `git status --porcelain=v2`.
//
// Additional status headers such as `# branch.*` and `# stash <N>` are parsed if present.
func Parse(r io.Reader) (*Status, error) {
	s := Status{}
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		line := scan.Bytes()
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case '#':
			// parseHeader manages the Branch or Stash field structs of the
			// Status struct directly, so we pass a pointer to the whole struct.
			parseHeader(line, &s)
		case '1':
			entry, err := parseChanged(line)
			if err != nil {
				return nil, err
			}
			s.Entries = append(s.Entries, entry)
		case '2':
			entry, err := parseRenameOrCopy(line)
			if err != nil {
				return nil, err
			}
			s.Entries = append(s.Entries, entry)
		case 'u':
			entry, err := parseUnmerged(line)
			if err != nil {
				return nil, err
			}
			s.Entries = append(s.Entries, entry)
		case '?':
			entry, err := parseUntracked(line)
			if err != nil {
				return nil, err
			}
			s.Entries = append(s.Entries, entry)
		case '!':
			entry, err := parseIgnored(line)
			if err != nil {
				return nil, err
			}
			s.Entries = append(s.Entries, entry)
		}
	}
	return &s, scan.Err()
}

func parseHeader(line []byte, s *Status) {
	fields := bytes.Fields(line)
	if len(fields) < 3 || fields[0][0] != '#' {
		return
	}

	switch string(fields[1]) {
	case "branch.oid":
		ensureBranch(s).OID = string(fields[2])
	case "branch.head":
		ensureBranch(s).Head = string(fields[2])
	case "branch.upstream":
		ensureBranch(s).Upstream = string(fields[2])
	case "branch.ab":
		if len(fields) < 4 {
			debugLogger.Warn("invalid branch.ab header", "line", string(line))
			return
		}
		fmt.Sscanf(string(fields[2]), "+%d", &ensureBranch(s).Ahead)
		fmt.Sscanf(string(fields[3]), "-%d", &ensureBranch(s).Behind)
	case "stash":
		n, err := strconv.ParseInt(string(fields[2]), 10, 0)
		if err != nil {
			// If we can't parse the stash count, just ignore it as invalid
			debugLogger.Warn("invalid stash count", "line", string(line), "error", err)
			return
		}
		s.Stash = &StashInfo{Count: int(n)}
	default:
		// Unknown header, ignore
		// TODO: debug log this in the future? could be interesting to see
	}
}

func ensureBranch(s *Status) *BranchInfo {
	if s.Branch == nil {
		s.Branch = &BranchInfo{}
	}
	return s.Branch
}

// Ordinary changed entries have the following format:
// 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
func parseChanged(line []byte) (ChangedEntry, error) {
	var e ChangedEntry
	fields := bytes.SplitN(line, []byte{' '}, 9)
	if len(fields) < 9 || fields[0][0] != '1' {
		return e, fmt.Errorf("invalid changed entry line: %q", line)
	}

	// Field 1: XY status code
	xy, err := parseXYFlag(fields[1])
	if err != nil {
		return e, err
	}
	e.XY = xy

	// Field 2: Submodule status
	sub, err := parseSubmoduleStatus(fields[2])
	if err != nil {
		return e, err
	}
	e.Sub = sub

	// Fields 3-5: File modes (HEAD, index, worktree)
	e.ModeH, err = parseFileMode(fields[3])
	if err != nil {
		return e, fmt.Errorf("invalid modeH field: %w", err)
	}
	e.ModeI, err = parseFileMode(fields[4])
	if err != nil {
		return e, fmt.Errorf("invalid modeI field: %w", err)
	}
	e.ModeW, err = parseFileMode(fields[5])
	if err != nil {
		return e, fmt.Errorf("invalid modeW field: %w", err)
	}

	// Fields 6-7: Object names (HEAD, index)
	// These are currently usually SHA-1 hashes in hex format, but treat as strings
	// given that they could be other types in the future (e.g. SHA-256 transition)
	e.HashH = string(fields[6])
	e.HashI = string(fields[7])

	// Field 8: Path
	e.Path = string(fields[8])

	return e, nil
}

// Renamed or copied entries have the following format:
// 2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path><sep><origPath>
func parseRenameOrCopy(line []byte) (RenameOrCopyEntry, error) {
	var e RenameOrCopyEntry
	fields := bytes.SplitN(line, []byte{' '}, 10)
	if len(fields) < 10 || fields[0][0] != '2' {
		return e, fmt.Errorf("invalid rename or copy entry line: %q", line)
	}

	// Field 1: XY status code
	xy, err := parseXYFlag(fields[1])
	if err != nil {
		return e, err
	}
	e.XY = xy

	// Field 2: Submodule status
	sub, err := parseSubmoduleStatus(fields[2])
	if err != nil {
		return e, err
	}
	e.Sub = sub

	// Fields 3-5: File modes (HEAD, index, worktree)
	e.ModeH, err = parseFileMode(fields[3])
	if err != nil {
		return e, fmt.Errorf("invalid modeH field: %w", err)
	}
	e.ModeI, err = parseFileMode(fields[4])
	if err != nil {
		return e, fmt.Errorf("invalid modeI field: %w", err)
	}
	e.ModeW, err = parseFileMode(fields[5])
	if err != nil {
		return e, fmt.Errorf("invalid modeW field: %w", err)
	}

	// Fields 6-7: Object names (HEAD, index)
	// These are currently usually SHA-1 hashes in hex format, but treat as strings
	// given that they could be other types in the future (e.g. SHA-256 transition)
	e.HashH = string(fields[6])
	e.HashI = string(fields[7])

	// Field 8: Rename or copy score
	// The rename or copy score (denoting the percentage of similarity between
	// the source and target of the move or copy). For example "R100" or "C75".
	e.Score = string(fields[8])

	// Field 9: <path><sep><origPath>
	// The target path (new path) and the origin path (old path) are separated
	// by tab (ASCII 0x09), except in -z mode, where they are separated by NUL
	// (ASCII 0x00).
	var sep = []byte{'\t'}
	path, orig, found := bytes.Cut(fields[9], sep)
	if !found {
		return e, fmt.Errorf("invalid rename/copy path entry format: %q", fields[9])
	}
	e.Path = string(path)
	e.Orig = string(orig)

	return e, nil
}

// Unmerged entries have the following format:
// u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
func parseUnmerged(line []byte) (UnmergedEntry, error) {
	var e UnmergedEntry
	fields := bytes.SplitN(line, []byte{' '}, 11)
	if len(fields) < 11 || fields[0][0] != 'u' {
		return e, fmt.Errorf("invalid unmerged entry line: %q", line)
	}

	// Field 1: XY status code
	xy, err := parseXYFlag(fields[1])
	if err != nil {
		return e, err
	}
	e.XY = xy

	// Field 2: Submodule status
	sub, err := parseSubmoduleStatus(fields[2])
	if err != nil {
		return e, err
	}
	e.Sub = sub

	// Fields 3-6: File modes (stage 1, stage 2, stage 3, worktree)
	e.Mode1, err = parseFileMode(fields[3])
	if err != nil {
		return e, fmt.Errorf("invalid mode1 field: %w", err)
	}
	e.Mode2, err = parseFileMode(fields[4])
	if err != nil {
		return e, fmt.Errorf("invalid mode2 field: %w", err)
	}
	e.Mode3, err = parseFileMode(fields[5])
	if err != nil {
		return e, fmt.Errorf("invalid mode3 field: %w", err)
	}
	e.ModeW, err = parseFileMode(fields[6])
	if err != nil {
		return e, fmt.Errorf("invalid modeW field: %w", err)
	}

	// Fields 7-9: Object names (stage 1, stage 2, stage 3)
	e.Hash1 = string(fields[7])
	e.Hash2 = string(fields[8])
	e.Hash3 = string(fields[9])

	// Field 10: Path
	e.Path = string(fields[10])

	return e, nil
}

// Untracked items have the following format:
// ? <path>
func parseUntracked(line []byte) (UntrackedEntry, error) {
	if !bytes.HasPrefix(line, []byte{'?', ' '}) {
		return UntrackedEntry{}, fmt.Errorf("invalid untracked entry line: %q", line)
	}

	path := string(line[2:]) // Skip the "? " prefix
	return UntrackedEntry{Path: path}, nil
}

// Ignored items have the following format:
// ! <path>
func parseIgnored(line []byte) (IgnoredEntry, error) {
	if !bytes.HasPrefix(line, []byte{'!', ' '}) {
		return IgnoredEntry{}, fmt.Errorf("invalid ignored entry line: %q", line)
	}

	path := string(line[2:]) // Skip the "! " prefix
	return IgnoredEntry{Path: path}, nil
}

func parseSubmoduleStatus(field []byte) (SubmoduleStatus, error) {
	var s SubmoduleStatus
	if len(field) != 4 {
		return s, fmt.Errorf("invalid submodule status field: %q", field)
	}
	return SubmoduleStatus{
		IsSubmodule:      field[0] == 'S',
		CommitChanged:    field[1] == 'C',
		HasModifications: field[2] == 'M',
		HasUntracked:     field[3] == 'U',
	}, nil
}

func parseFileMode(field []byte) (FileMode, error) {
	mode, err := strconv.ParseUint(string(field), 8, 32)
	if err != nil {
		return 0, err
	}
	return FileMode(mode), nil
}

func parseXYFlag(field []byte) (XYFlag, error) {
	if len(field) != 2 {
		return XYFlag{}, fmt.Errorf("invalid XY field: expected 2 characters, got %d", len(field))
	}
	return XYFlag{State(field[0]), State(field[1])}, nil
}

// ParseZ parses the output of `git status --porcelain=v2 -z`.
//
// -z output is a special mode that uses a NUL (ASCII 0x00) byte as line terminator instead of newline,
// as well as disabling any quoting of special characters in pathnames.
// func ParseZ(input io.Reader) (*GitStatusV2, error) {
// 	panic("not implemented")
// }
