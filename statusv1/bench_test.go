package statusv1

import (
	"bytes"
	"io"
	"testing"
)

func BenchmarkParse_Sample(b *testing.B) {
	r := bytes.NewReader(samplePorcelainV1Output)

	b.ReportAllocs()
	for b.Loop() {
		Parse(r)
		r.Seek(0, io.SeekStart) // reset reader for next iteration
	}
}

func BenchmarkParseZ_Sample(b *testing.B) {
	r := bytes.NewReader(samplePorcelainV1ZOutput)

	b.ReportAllocs()
	for b.Loop() {
		ParseZ(r)
		r.Seek(0, io.SeekStart) // reset reader for next iteration
	}
}

func Benchmark_parseEntry_Modified(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseEntry(sampleEntryModified)
	}
}

func Benchmark_parseEntry_Renamed(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseEntry(sampleEntryRenamed)
	}
}

func Benchmark_parseEntryZ_Modified(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseEntryZ(sampleEntryModified)
	}
}

func Benchmark_parseEntryZ_Renamed(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		parseEntryZ(sampleEntryRenamedZ)
	}
}
