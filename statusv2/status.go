// Package statusv2 implements parsing of `git status --porcelain=v2` output.
package statusv2

import "strconv"

// GitStatusV2 represents a full --porcelain=v2 snapshot.
// If you invoked `--branch`, Branch will be non-nil. If `--show-stash` was passed and N>0, Stash.Count == N.
type GitStatusV2 struct {
	Branch  *BranchInfo // nil if `--branch` not passed
	Stash   *StashInfo  // nil if `--show-stash` not passed or count == 0
	Entries []Entry     // in the order lines appeared; can be ChangedEntry, RenameOrCopyEntry, UnmergedEntry, UntrackedEntry, or IgnoredEntry
}

// BranchInfo holds data from all “# branch.*” headers (requires --branch).
type BranchInfo struct {
	OID      string // commit hash or "(initial)"
	Head     string // branch name or "(detached)"
	Upstream string // upstream branch name (empty if unset)
	Ahead    int    // how many commits ahead of upstream
	Behind   int    // how many commits behind upstream
}

// StashInfo holds the stash header “# stash <N>” (requires --show-stash).
type StashInfo struct {
	Count int // number of stash entries
}

// EntryType enumerates the kinds of per-file lines in porcelain v2.
type EntryType int

const (
	EntryTypeChanged      EntryType = iota // “1” ordinary changed entry
	EntryTypeRenameOrCopy                  // “2” rename or copy
	EntryTypeUnmerged                      // “u” merge-conflict entry
	EntryTypeUntracked                     // “?” untracked file
	EntryTypeIgnored                       // “!” ignored file
)

// Entry is a union interface; each implementation corresponds to one line-type.
type Entry interface {
	Type() EntryType
}

// State is one of the valid XY status bytes (as in “<XY>” field).
type State byte

const (
	Unmodified      State = '.' // unmodified
	Modified        State = 'M' // modified
	TypeChanged     State = 'T' // file type changed (e.g. regular→symlink)
	Added           State = 'A' // added
	Deleted         State = 'D' // deleted
	Renamed         State = 'R' // renamed
	Copied          State = 'C' // copied (if status.renames=copies)
	UpdatedUnmerged State = 'U' // updated but unmerged (merge conflict)
)

// XYFlag holds the two-character XY status (staged + unstaged).
// “<XY>” in “1”, “2” and “u” lines, where unchanged is “.” not space.
type XYFlag [2]State

func (xy XYFlag) X() State       { return xy[0] }
func (xy XYFlag) Y() State       { return xy[1] }
func (xy XYFlag) String() string { return string(xy[0]) + string(xy[1]) }

// A FileMode represents the kind of tree entries used by git. It resembles
// regular file systems modes, although FileModes are considerably simpler.
type FileMode uint32

// For more information on possible FileMode values, see:
// https://pkg.go.dev/github.com/go-git/go-git/v5/plumbing/filemode#FileMode
const (
	FileModeEmpty      FileMode = 0
	FileModeDir        FileMode = 0040000
	FileModeRegular    FileMode = 0100644
	FileModeExecutable FileMode = 0100755
	FileModeSymlink    FileMode = 0120000
	FileModeSubmodule  FileMode = 0160000
)

func (m FileMode) String() string {
	return strconv.FormatUint(uint64(m), 8)
}

// SubmoduleStatus represents the 4-character “<sub>” field (for lines “1”, “2” and “u”).
// “N…” means not a submodule; “S<c><m><u>” for a submodule: <c>='C' if commit changed, <m>='M' if tracked-content changed, <u>='U' if untracked files present.
type SubmoduleStatus struct {
	IsSubmodule      bool // true if first char == 'S'
	CommitChanged    bool // true if second char == 'C'
	HasModifications bool // true if third char == 'M'
	HasUntracked     bool // true if fourth char == 'U'
}

// ChangedEntry models a “1” line: an ordinary changed file (not rename/copy).
//
//	1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
type ChangedEntry struct {
	XY    XYFlag          // two-character XY status
	Sub   SubmoduleStatus // 4-character submodule field
	ModeH FileMode        // <mH> filemode in HEAD (octal string)
	ModeI FileMode        // <mI> filemode in index (octal string)
	ModeW FileMode        // <mW> filemode in worktree (octal string)
	HashH string          // <hH> object name (SHA-1) in HEAD
	HashI string          // <hI> object name (SHA-1) in index
	Path  string          // file path relative to repo root
}

func (ChangedEntry) Type() EntryType { return EntryTypeChanged }

// RenameOrCopyEntry models a “2” line: renamed or copied file.
//
//	2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path><sep><origPath>
type RenameOrCopyEntry struct {
	XY    XYFlag          // two-character XY status
	Sub   SubmoduleStatus // 4-character submodule field
	ModeH FileMode        // <mH> filemode in HEAD (octal string)
	ModeI FileMode        // <mI> filemode in index (octal string)
	ModeW FileMode        // <mW> filemode in worktree (octal string)
	HashH string          // <hH> object name (SHA-1) in HEAD
	HashI string          // <hI> object name (SHA-1) in index
	Score string          // "<X><score>" (e.g. "R100" or "C75")
	Path  string          // target path (new path)
	Orig  string          // origin path (old path)
}

func (RenameOrCopyEntry) Type() EntryType { return EntryTypeRenameOrCopy }

// UnmergedEntry models a “u” line: an unmerged (conflicted) file.
//
//	u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
type UnmergedEntry struct {
	XY    XYFlag          // two-character XY status (e.g. “DD”, “AU”, “UU”)
	Sub   SubmoduleStatus // 4-character submodule field
	Mode1 FileMode        // <m1> mode of stage 1 (common base)
	Mode2 FileMode        // <m2> mode of stage 2 (ours)
	Mode3 FileMode        // <m3> mode of stage 3 (theirs)
	ModeW FileMode        // <mW> mode in worktree
	Hash1 string          // <h1> SHA of stage-1 blob
	Hash2 string          // <h2> SHA of stage-2 blob
	Hash3 string          // <h3> SHA of stage-3 blob
	Path  string          // file path relative to repo root
}

func (UnmergedEntry) Type() EntryType { return EntryTypeUnmerged }

// UntrackedEntry models a “?” line: an untracked file.
//
//	? <path>
type UntrackedEntry struct {
	Path string // file path
}

func (UntrackedEntry) Type() EntryType { return EntryTypeUntracked }

// IgnoredEntry models a “!” line: an ignored file.
//
//	! <path>
type IgnoredEntry struct {
	Path string // file path
}

func (IgnoredEntry) Type() EntryType { return EntryTypeIgnored }
