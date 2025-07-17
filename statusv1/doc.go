/*
Package statusv1 parses the output of `git status --porcelain=v1`.

This package provides parsing for Git's machine-readable status format,
supporting both regular line-terminated output and NUL-terminated output (with
-z flag). This package implements parsing of the original porcelain=v1 format,
which is simpler than the more modern porcelain=v2 format.

# Basic Usage

[Parse] takes an [io.Reader] containing `git status --porcelain=v1` output.

	r := bytes.NewReader(gitStatusOutput)
	status, err := statusv1.Parse(r)
	if err != nil {
	    log.Fatal(err)
	}

[ParseZ] provides a variant that will work with NUL-terminated git status output
(from -z flag).

# Working with Results

The [Status] struct contains parsed information, notably the list of file
entries, which can be accessed via the [Status.Entries] field. Each entry is
represented by an [Entry] struct, which contains the XY status flags and file
paths.

# Git Status Format

This package parses Git's porcelain=v1 format, which provides machine-readable
output with basic information about file status and branch state. The format is
stable across Git versions and designed for programmatic consumption.

The porcelain=v1 format outputs one line per file:

	XY PATH
	XY ORIG_PATH -> PATH  (for renames/copies)

Where XY is a two-character status code indicating the state of the file in the
index (X) and working tree (Y).

There is also an alternate -z format recommended for machine parsing. In that
format, the status field is the same, but some other things change. First, the
-> is omitted from rename entries and the field order is reversed (e.g from ->
to becomes to from). Second, a NUL (ASCII 0) follows each filename, replacing
space as a field separator and the terminating newline (but a space still
separates the status field from the first filename). Third, filenames containing
special characters are not specially formatted; no quoting or backslash-escaping
is performed.

In most cases, users should prefer the more modern porcelain=v2 format, which
provides more detailed information and additional features. See the [statusv2]
package for parsing porcelain=v2 output.

For more information about the porcelain=v1 format, see the Git documentation
for [git status].

[git status]: https://git-scm.com/docs/git-status#_porcelain_format_version_1
*/
package statusv1
