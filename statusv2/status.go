package statusv2

import "strconv"

// Status represents parsed git status --porcelain=v2 output.
//
// Branch contains branch information if --branch was used.
// Stash contains stash count if --show-stash was used and stashes exist.
// Entries contains all file status entries in the order they appeared.
type Status struct {
	Branch  *BranchInfo // nil if `--branch` not passed
	Stash   *StashInfo  // nil if `--show-stash` not passed or count == 0
	Entries []Entry     // in the order lines appeared; can be ChangedEntry, RenameOrCopyEntry, UnmergedEntry, UntrackedEntry, or IgnoredEntry
}

// BranchInfo contains branch information from git status --branch output.
//
// Available when --branch flag is used. Contains current branch state,
// upstream tracking information, and ahead/behind commit counts.
type BranchInfo struct {
	OID      string // current commit hash, or "(initial)" for new repos
	Head     string // current branch name, or "(detached)" for detached HEAD
	Upstream string // upstream branch name (empty if no upstream set)
	Ahead    int    // commits ahead of upstream
	Behind   int    // commits behind upstream
}

// StashInfo contains stash information from git status --show-stash output.
//
// Available when --show-stash flag is used and stashes exist.
type StashInfo struct {
	Count int // number of stash entries
}

// EntryType identifies the kind of file status entry.
type EntryType int

// Entry type constants corresponding to git status line prefixes.
const (
	EntryTypeChanged      EntryType = iota // "1" - modified files
	EntryTypeRenameOrCopy                  // "2" - renamed or copied files
	EntryTypeUnmerged                      // "u" - merge conflict files
	EntryTypeUntracked                     // "?" - untracked files
	EntryTypeIgnored                       // "!" - ignored files
)

// Entry represents a file status entry. Use type switching to access specific fields:
//
//	switch e := entry.(type) {
//	case ChangedEntry:
//		// Access e.Path, etc.
//	case RenameOrCopyEntry:
//		// Access e.Path, e.Orig, etc.
//	}
type Entry interface {
	Type() EntryType
}

// State represents a single character from Git porcelain=v2 XY status codes.
type State byte

// Git status state codes for index (X) and worktree (Y) changes.
//
// The [Unmodified] state is represented by a dot ('.') in porcelain=v2,
// which is different from the space (' ') used in porcelain=v1.
//
// Untracked and ignored files are no longer represented in XY flag states in
// porcelain=v2, but rather as separate entry types ([UntrackedEntry] and
// [IgnoredEntry]).
const (
	Unmodified      State = '.' // unmodified (no changes)
	Modified        State = 'M' // modified
	TypeChanged     State = 'T' // file type changed (regular file, symbolic link or submodule)
	Added           State = 'A' // added
	Deleted         State = 'D' // deleted
	Renamed         State = 'R' // renamed
	Copied          State = 'C' // copied (if status.renames=copies)
	UpdatedUnmerged State = 'U' // updated but unmerged (merge conflict)
)

// XYFlag holds the two-character XY status codes (index + worktree).
// X represents staged changes, Y represents unstaged changes.
// Unchanged files use "." in porcelain=v2, not space.
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

// String returns the octal string representation of the FileMode, e.g. "100644".
// Note that this is different from Go octal formatting, which uses a leading "0".
func (m FileMode) String() string {
	return strconv.FormatUint(uint64(m), 8)
}

// SubmoduleStatus represents submodule state information.
//
// For regular files, IsSubmodule is false and other fields are ignored.
// For submodules, the flags indicate different types of changes:
//   - CommitChanged: submodule commit differs from what's recorded
//   - HasModifications: tracked files within submodule have changes
//   - HasUntracked: untracked changes exist within submodule
type SubmoduleStatus struct {
	IsSubmodule      bool // true if this entry represents a submodule
	CommitChanged    bool // true if submodule commit has changed
	HasModifications bool // true if submodule has tracked changes
	HasUntracked     bool // true if submodule has untracked changes
}

// TODO: add String() method to SubmoduleStatus?
// <sub>       A 4 character field describing the submodule state.
// 	    "N..." when the entry is not a submodule
// 	    "S<c><m><u>" when the entry is a submodule
// 	    <c> is "C" if the commit changed; otherwise "."
// 	    <m> is "M" if it has tracked changes; otherwise "."
// 	    <u> is "U" if there are untracked changes; otherwise "."

// ChangedEntry represents a modified file (added, modified, deleted, etc).
//
// Corresponds to porcelain=v2 status lines starting with "1". Does not include
// renamed or copied files (see [RenameOrCopyEntry]).
type ChangedEntry struct {
	XY    XYFlag          // staged and unstaged XY values
	Sub   SubmoduleStatus // submodule state information
	ModeH FileMode        // file mode in HEAD commit
	ModeI FileMode        // file mode in index (staged)
	ModeW FileMode        // file mode in worktree (unstaged)
	HashH string          // object hash in HEAD commit
	HashI string          // object hash in index (staged)
	Path  string          // file path relative to repository root
}

func (ChangedEntry) Type() EntryType { return EntryTypeChanged }

// RenameOrCopyEntry represents a renamed or copied file.
//
// Corresponds to porcelain=v2 status lines starting with "2". Includes both the
// original and new file paths, plus a similarity score.
type RenameOrCopyEntry struct {
	XY    XYFlag          // staged and unstaged XY values
	Sub   SubmoduleStatus // submodule state information
	ModeH FileMode        // file mode in HEAD commit
	ModeI FileMode        // file mode in index (staged)
	ModeW FileMode        // file mode in worktree (unstaged)
	HashH string          // object hash in HEAD commit
	HashI string          // object hash in index (staged)
	Score string          // similarity score (e.g. "R100", "C75")
	Path  string          // new file path
	Orig  string          // original file path
}

func (RenameOrCopyEntry) Type() EntryType { return EntryTypeRenameOrCopy }

// UnmergedEntry represents a file with merge conflicts.
//
// Corresponds to porcelain=v2 status lines starting with "u". Contains
// information about all three merge stages: base (1), ours (2), and theirs (3).
type UnmergedEntry struct {
	XY    XYFlag          // conflict type XY values
	Sub   SubmoduleStatus // submodule state information
	Mode1 FileMode        // file mode in stage 1 (common base)
	Mode2 FileMode        // file mode in stage 2 (ours)
	Mode3 FileMode        // file mode in stage 3 (theirs)
	ModeW FileMode        // file mode in worktree
	Hash1 string          // object hash in stage 1 (common base)
	Hash2 string          // object hash in stage 2 (ours)
	Hash3 string          // object hash in stage 3 (theirs)
	Path  string          // file path relative to repository root
}

func (UnmergedEntry) Type() EntryType { return EntryTypeUnmerged }

// UntrackedEntry represents an untracked file.
//
// Corresponds to git status lines starting with "?".
type UntrackedEntry struct {
	Path string // file path relative to repository root
}

func (UntrackedEntry) Type() EntryType { return EntryTypeUntracked }

// IgnoredEntry represents an ignored file.
//
// Corresponds to git status lines starting with "!" (when --ignored is used).
type IgnoredEntry struct {
	Path string // file path relative to repository root
}

func (IgnoredEntry) Type() EntryType { return EntryTypeIgnored }
