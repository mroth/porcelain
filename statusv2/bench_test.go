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
		parseHeader(sampleHeaderBranchOID, &s)
		parseHeader(sampleHeaderBranchHead, &s)
		parseHeader(sampleHeaderBranchUpstream, &s)
		parseHeader(sampleHeaderBranchAB, &s)
		parseHeader(sampleHeaderStash, &s)
	}
}

func Benchmark_parseChange(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseChanged(sampleEntryChanged)
	}
}

func Benchmark_parseRenameOrCopy(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseRenameOrCopy(sampleEntryRenamed, tabSeparator)
	}
}

func Benchmark_parseUnmerged(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseUnmerged(sampleEntryUnmerged)
	}
}

func Benchmark_parseUntracked(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseUntracked(sampleEntryUntracked)
	}
}

func Benchmark_parseIgnored(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseIgnored(sampleEntryIgnored)
	}
}
