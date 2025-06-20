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
// Additional status headers such as `--branch` and `--show-status` are parsed if present.
func Parse(r io.Reader) (*Status, error) {
	return parseWithScanner(bufio.NewScanner(r), tabSeparator)
}

// ParseZ parses the output of `git status --porcelain=v2 -z`.
//
// The -z flag changes line termination from LF to NUL and path separation in rename/copy
// entries from tab to NUL.
//
// For additional details, see the documentation of [Parse].
func ParseZ(r io.Reader) (*Status, error) {
	return parseWithScanner(newZScanner(r), nulSeparator)
}

// renamePathSep represents the byte used to separate paths in rename/copy entries
type renamePathSep byte

const (
	tabSeparator renamePathSep = '\t'   // Normal mode: paths separated by tab
	nulSeparator renamePathSep = '\x00' // -z mode: paths separated by NUL
)

func parseWithScanner(scanner *bufio.Scanner, pathSep renamePathSep) (*Status, error) {
	s := Status{}
	for scanner.Scan() {
		line := scanner.Bytes()
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
			entry, err := parseRenameOrCopy(line, pathSep)
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
	return &s, scanner.Err()
}

// Headers take the form of `# <key> <values...>` where <key> is a string like
// "branch.oid" or "stash". As per the specification, parsers should ignore
// unknown headers, so we don't return an error if the header is not recognized.
func parseHeader(line []byte, s *Status) {
	line, ok := bytes.CutPrefix(line, []byte("# "))
	if !ok {
		return
	}

	headerKey, value, found := bytes.Cut(line, []byte{' '})
	if !found {
		return
	}

	switch string(headerKey) {
	case "branch.oid":
		ensureBranch(s).OID = string(value)
	case "branch.head":
		ensureBranch(s).Head = string(value)
	case "branch.upstream":
		ensureBranch(s).Upstream = string(value)
	case "branch.ab":
		fmt.Sscanf(string(value), "+%d -%d", &ensureBranch(s).Ahead, &ensureBranch(s).Behind)
	case "stash":
		n, err := strconv.ParseInt(string(value), 10, 0)
		if err != nil {
			// If we can't parse the stash count, just ignore it as invalid
			debugLogger.Warn("invalid stash count", "line", string(line), "error", err)
			return
		}
		s.Stash = &StashInfo{Count: int(n)}
	default:
		debugLogger.Debug("unrecognized status header", "line", string(line))
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
	var zero ChangedEntry
	fields := bytes.SplitN(line, []byte{' '}, 9)
	if len(fields) < 9 || !bytes.HasPrefix(fields[0], []byte{'1'}) {
		return zero, fmt.Errorf("invalid changed entry line: %q", line)
	}

	// Field 1: XY status code
	xy, err := parseXYFlag(fields[1])
	if err != nil {
		return zero, err
	}

	// Field 2: Submodule status
	sub, err := parseSubmoduleStatus(fields[2])
	if err != nil {
		return zero, err
	}

	// Fields 3-5: File modes (HEAD, index, worktree)
	modeH, err := parseFileMode(fields[3])
	if err != nil {
		return zero, fmt.Errorf("invalid modeH field: %w", err)
	}
	modeI, err := parseFileMode(fields[4])
	if err != nil {
		return zero, fmt.Errorf("invalid modeI field: %w", err)
	}
	modeW, err := parseFileMode(fields[5])
	if err != nil {
		return zero, fmt.Errorf("invalid modeW field: %w", err)
	}

	// Fields 6-7: Object names (HEAD, index)
	// These are currently usually SHA-1 hashes in hex format, but treat as strings
	// given that they could be other types in the future (e.g. SHA-256 transition)
	hashH := string(fields[6])
	hashI := string(fields[7])

	// Field 8: Path
	path := string(fields[8])

	return ChangedEntry{
		XY:    xy,
		Sub:   sub,
		ModeH: modeH,
		ModeI: modeI,
		ModeW: modeW,
		HashH: hashH,
		HashI: hashI,
		Path:  path,
	}, nil
}

// Renamed or copied entries have the following format:
// 2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path><sep><origPath>
func parseRenameOrCopy(line []byte, pathSep renamePathSep) (RenameOrCopyEntry, error) {
	var zero RenameOrCopyEntry
	fields := bytes.SplitN(line, []byte{' '}, 10)
	if len(fields) < 10 || !bytes.HasPrefix(fields[0], []byte{'2'}) {
		return zero, fmt.Errorf("invalid rename or copy entry line: %q", line)
	}

	// Field 1: XY status code
	xy, err := parseXYFlag(fields[1])
	if err != nil {
		return zero, err
	}

	// Field 2: Submodule status
	sub, err := parseSubmoduleStatus(fields[2])
	if err != nil {
		return zero, err
	}

	// Fields 3-5: File modes (HEAD, index, worktree)
	modeH, err := parseFileMode(fields[3])
	if err != nil {
		return zero, fmt.Errorf("invalid modeH field: %w", err)
	}
	modeI, err := parseFileMode(fields[4])
	if err != nil {
		return zero, fmt.Errorf("invalid modeI field: %w", err)
	}
	modeW, err := parseFileMode(fields[5])
	if err != nil {
		return zero, fmt.Errorf("invalid modeW field: %w", err)
	}

	// Fields 6-7: Object names (HEAD, index)
	// These are currently usually SHA-1 hashes in hex format, but treat as strings
	// given that they could be other types in the future (e.g. SHA-256 transition)
	hashH := string(fields[6])
	hashI := string(fields[7])

	// Field 8: Rename or copy score
	// The rename or copy score (denoting the percentage of similarity between
	// the source and target of the move or copy). For example "R100" or "C75".
	score := string(fields[8])

	// Field 9: <path><sep><origPath>
	// The target path (new path) and the origin path (old path) are separated
	// by tab (ASCII 0x09), except in -z mode, where they are separated by NUL
	// (ASCII 0x00).
	sep := []byte{byte(pathSep)}
	pathBytes, origBytes, found := bytes.Cut(fields[9], sep)
	if !found {
		return zero, fmt.Errorf("invalid rename/copy path entry format: %q", fields[9])
	}
	path := string(pathBytes)
	orig := string(origBytes)

	return RenameOrCopyEntry{
		XY:    xy,
		Sub:   sub,
		ModeH: modeH,
		ModeI: modeI,
		ModeW: modeW,
		HashH: hashH,
		HashI: hashI,
		Score: score,
		Path:  path,
		Orig:  orig,
	}, nil
}

// Unmerged entries have the following format:
// u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
func parseUnmerged(line []byte) (UnmergedEntry, error) {
	var zero UnmergedEntry
	fields := bytes.SplitN(line, []byte{' '}, 11)
	if len(fields) < 11 || !bytes.HasPrefix(fields[0], []byte{'u'}) {
		return zero, fmt.Errorf("invalid unmerged entry line: %q", line)
	}

	// Field 1: XY status code
	xy, err := parseXYFlag(fields[1])
	if err != nil {
		return zero, err
	}

	// Field 2: Submodule status
	sub, err := parseSubmoduleStatus(fields[2])
	if err != nil {
		return zero, err
	}

	// Fields 3-6: File modes (stage 1, stage 2, stage 3, worktree)
	mode1, err := parseFileMode(fields[3])
	if err != nil {
		return zero, fmt.Errorf("invalid mode1 field: %w", err)
	}
	mode2, err := parseFileMode(fields[4])
	if err != nil {
		return zero, fmt.Errorf("invalid mode2 field: %w", err)
	}
	mode3, err := parseFileMode(fields[5])
	if err != nil {
		return zero, fmt.Errorf("invalid mode3 field: %w", err)
	}
	modeW, err := parseFileMode(fields[6])
	if err != nil {
		return zero, fmt.Errorf("invalid modeW field: %w", err)
	}

	// Fields 7-9: Object names (stage 1, stage 2, stage 3)
	hash1 := string(fields[7])
	hash2 := string(fields[8])
	hash3 := string(fields[9])

	// Field 10: Path
	path := string(fields[10])

	return UnmergedEntry{
		XY:    xy,
		Sub:   sub,
		Mode1: mode1,
		Mode2: mode2,
		Mode3: mode3,
		ModeW: modeW,
		Hash1: hash1,
		Hash2: hash2,
		Hash3: hash3,
		Path:  path,
	}, nil
}

// Untracked items have the following format:
// ? <path>
func parseUntracked(line []byte) (UntrackedEntry, error) {
	pathBytes, ok := bytes.CutPrefix(line, []byte{'?', ' '})
	if !ok {
		return UntrackedEntry{}, fmt.Errorf("invalid untracked entry line: %q", line)
	}

	return UntrackedEntry{Path: string(pathBytes)}, nil
}

// Ignored items have the following format:
// ! <path>
func parseIgnored(line []byte) (IgnoredEntry, error) {
	pathBytes, ok := bytes.CutPrefix(line, []byte{'!', ' '})
	if !ok {
		return IgnoredEntry{}, fmt.Errorf("invalid ignored entry line: %q", line)
	}

	return IgnoredEntry{Path: string(pathBytes)}, nil
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
