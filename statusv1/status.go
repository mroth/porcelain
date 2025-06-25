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

// XYFlag represents the two-character status code in porcelain=v1 format.
// The X position shows the status of the index, and the Y position shows
// the status of the working tree.
type XYFlag struct {
	X State // index status
	Y State // working tree status
}

// String returns the XY status as a two-character string.
func (xy XYFlag) String() string { return string(xy.X) + string(xy.Y) }

// Entry represents a single file entry in git status --porcelain=v1 output.
type Entry struct {
	XY       XYFlag // two-character status code
	Path     string // current path of the file
	OrigPath string // original path for renamed/copied files (empty if not renamed/copied)
}

// Status represents the parsed output of git status --porcelain=v1.
//
// The Header field contains any header lines from the output, which may be present
// when using flags such as --branch.  These lines are always prefixed with `##`.
type Status struct {
	Headers []string // header lines (prefixed with `##`), if present
	Entries []Entry  // file entries
}
