/*
Package statusv2 parses the output of `git status --porcelain=v2`.

This package provides parsing for Git's machine-readable status format, supporting
both regular line-terminated output and NUL-terminated output (with -z flag).

# Basic Usage

[Parse] takes an [io.Reader] containing `git status --porcelain=v2` output. Branch
and stash information are also parsed if the `--branch` and/or `--show-stash` flags were
used with the command.

	r := bytes.NewReader(gitStatusOutput)
	status, err := statusv2.Parse(r)
	if err != nil {
	    log.Fatal(err)
	}

[ParseZ] provides a variant that will work with NUL-terminated git status output (from -z flag):

# Working with Results

The [Status] struct contains parsed information:

	// Access branch and stash information
	fmt.Printf("Branch: %s\n", status.Branch.Head)
	fmt.Printf("Ahead: %d, Behind: %d\n", status.Branch.Ahead, status.Branch.Behind)

	// Iterate through file entries
	for _, entry := range status.Entries {
	    switch e := entry.(type) {
	    case ChangedEntry:
	        fmt.Printf("Changed: %s (flags: %s)\n", e.Path, e.XY)
	    case RenameOrCopyEntry:
	        fmt.Printf("Renamed: %s -> %s\n", e.OrigPath, e.Path)
	    case UnmergedEntry:
	        fmt.Printf("Conflict: %s [%o/%o/%o]\n", e.Path, e.Mode1, e.Mode2, e.Mode3)
	    case UntrackedEntry:
	        fmt.Printf("Untracked: %s\n", e.Path)
	    case IgnoredEntry:
	        fmt.Printf("Ignored: %s\n", e.Path)
	    }
	}

# Entry Types

The package defines several entry types that implement the [Entry] interface:

  - [ChangedEntry] - Files with changes in index or worktree
  - [RenameOrCopyEntry] - Files that have been renamed or copied
  - [UnmergedEntry] - Files with merge conflicts
  - [UntrackedEntry] - Files not tracked by Git
  - [IgnoredEntry] - Files ignored by Git

Each entry type has specific fields relevant to its status. Use type switching
to access the specific fields for each entry type.

# Git Status Format

This package parses Git's porcelain=v2 format, which provides machine-readable
output with detailed information about file status, branch state, and stash
information. The format is stable across Git versions and designed for
programmatic consumption.

For more information about the porcelain=v2 format, see the Git documentation
for [git status].

[git status]: https://git-scm.com/docs/git-status#_porcelain_format_version_2
*/
package statusv2
