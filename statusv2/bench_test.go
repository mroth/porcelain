package statusv2

import (
	"io"
	"strings"
	"testing"
)

func BenchmarkParse_Sample(b *testing.B) {
	r := strings.NewReader(samplePorcelainV2Output)

	b.ReportAllocs()
	for b.Loop() {
		Parse(r)
		r.Seek(0, io.SeekStart) // reset reader for next iteration
	}
}

func Benchmark_parseHeaders(b *testing.B) {
	var s Status
	var (
		branchOID      = []byte("# branch.oid 34064be349d4a03ed158aba170d8d2db6ff9e3e0")
		branchHead     = []byte("# branch.head main")
		branchUpstream = []byte("# branch.upstream origin/main")
		branchAB       = []byte("# branch.ab +6 -3")
		stash          = []byte("# stash 3")
	)

	b.ReportAllocs()
	for b.Loop() {
		parseHeader(branchOID, &s)
		parseHeader(branchHead, &s)
		parseHeader(branchUpstream, &s)
		parseHeader(branchAB, &s)
		parseHeader(stash, &s)
	}
}

func Benchmark_parseChange(b *testing.B) {
	var simple = []byte("1 M. N... 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 1234567890abcdef1234567890abcdef12345678 file_changed.txt")

	b.ReportAllocs()
	for b.Loop() {
		parseChanged(simple)
	}
}

func Benchmark_parseRenameOrCopy(b *testing.B) {
	var simple = []byte("2 R. N... 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 1234567890abcdef1234567890abcdef12345678 R100 file_renamed.txt\tfile_original.txt")

	b.ReportAllocs()
	for b.Loop() {
		parseRenameOrCopy(simple, tabSeparator)
	}
}

func Benchmark_parseUnmerged(b *testing.B) {
	var simple = []byte("1 U. N... 100644 100644 100644 1234567890abcdef1234567890abcdef12345678 1234567890abcdef1234567890abcdef12345678 file_unmerged.txt")

	b.ReportAllocs()
	for b.Loop() {
		parseUnmerged(simple)
	}
}

func Benchmark_parseUntracked(b *testing.B) {
	var simple = []byte("? file_untracked.txt")

	b.ReportAllocs()
	for b.Loop() {
		parseUntracked(simple)
	}
}

func Benchmark_parseIgnored(b *testing.B) {
	var simple = []byte("! file_ignored.txt")

	b.ReportAllocs()
	for b.Loop() {
		parseIgnored(simple)
	}
}
