// Command porcelain2go converts porcelain output of `git status` into a Go struct.
// It reads from stdin and writes to stdout, so it can be used in a pipeline.
// It is primarily intended for use in testing and debugging on the CLI.
//
// Usage example:
//
//	git status --porcelain=v1 | porcelain2go -format v1
//	git status --porcelain=v2 | porcelain2go -format v2
//	git status --porcelain=v1 -z | porcelain2go -format v1z
//	git status --porcelain=v2 -z | porcelain2go -format v2z
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mroth/porcelain/statusv1"
	"github.com/mroth/porcelain/statusv2"
)

var (
	porcelainVersion = flag.String("format", "v2", "porcelain version to parse [v1, v1z, v2, v2z]")
)

type StatusParser func(io.Reader) (any, error)

func getStatusParser(format string) (StatusParser, error) {
	switch format {
	case "v1":
		return func(r io.Reader) (any, error) { return statusv1.Parse(r) }, nil
	case "v1z":
		return func(r io.Reader) (any, error) { return statusv1.ParseZ(r) }, nil
	case "v2":
		return func(r io.Reader) (any, error) { return statusv2.Parse(r) }, nil
	case "v2z":
		return func(r io.Reader) (any, error) { return statusv2.ParseZ(r) }, nil
	default:
		return nil, fmt.Errorf("unsupported -format flag value: %s", format)
	}
}

func main() {
	flag.Parse()
	parser, err := getStatusParser(*porcelainVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		flag.Usage()
		os.Exit(2)
	}

	in := bufio.NewReader(os.Stdin)
	results, err := parser(in)
	if err != nil {
		log.Fatalf("fatal: error parsing porcelain output: %v", err)
	}

	out, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("fatal: error marshaling results to JSON: %v", err)
	}
	fmt.Println(string(out))
}
