package statusv2

import (
	"bytes"
	"io"
	"testing"
)

func BenchmarkParse_Sample(b *testing.B) {
	r := bytes.NewReader(samplePorcelainV2Output)

	b.ReportAllocs()
	for b.Loop() {
		Parse(r)
		r.Seek(0, io.SeekStart) // reset reader for next iteration
	}
}

func Benchmark_parseHeaders(b *testing.B) {
	var s Status

	b.ReportAllocs()
	for b.Loop() {
		parseHeaderEntry(sampleHeaderBranchOID, &s)
		parseHeaderEntry(sampleHeaderBranchHead, &s)
		parseHeaderEntry(sampleHeaderBranchUpstream, &s)
		parseHeaderEntry(sampleHeaderBranchAB, &s)
		parseHeaderEntry(sampleHeaderStash, &s)
	}
}

func Benchmark_parseChange(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseChangedEntry(sampleEntryChanged)
	}
}

func Benchmark_parseRenameOrCopy(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseRenameOrCopyEntry(sampleEntryRenamed, tabSeparator)
	}
}

func Benchmark_parseUnmerged(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseUnmergedEntry(sampleEntryUnmerged)
	}
}

func Benchmark_parseUntracked(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseUntrackedEntry(sampleEntryUntracked)
	}
}

func Benchmark_parseIgnored(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseIgnoredEntry(sampleEntryIgnored)
	}
}
