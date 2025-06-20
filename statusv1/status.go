// Package statusv1 provides the status codes used in Git porcelain=v1.
//
// Parsing of Git status output in porcelain=v1 format is currently
// unimplemented in favor of the more modern porcelain=v2 format.
package statusv1

// State represents a single character from Git porcelain=v1 status codes.
type State byte

// Git status state codes for index (X) and worktree (Y) changes.
//
// The [Unmodified] state is represented by a space (' ') in porcelain=v1,
// which is different from the '.' used in porcelain=v2.
//
// In porcelain=v1, the XYFlag is used represent untracked ('?') and ignored
// ('!') files, but in porcelain=v2, these become separate entry types.
const (
	Unmodified      State = ' ' // unmodified (no changes)
	Modified        State = 'M' // modified
	TypeChanged     State = 'T' // file type changed (regular file, symbolic link or submodule)
	Added           State = 'A' // added
	Deleted         State = 'D' // deleted
	Renamed         State = 'R' // renamed
	Copied          State = 'C' // copied (if status.renames=copies)
	UpdatedUnmerged State = 'U' // updated but unmerged (merge conflict)

	Untracked State = '?' // untracked files
	Ignored   State = '!' // ignored files
)
