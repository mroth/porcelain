// Command porcelain2go converts porcelain output of `git status` into a Go struct.
// It reads from stdin and writes to stdout, so it can be used in a pipeline.
// It is primarily intended for use in testing and debugging on the CLI.
//
// Usage example:
//
//	git status --porcelain=v2 | porcelain2go -format v2
//	git status --porcelain=v2 -z | porcelain2go -format v2z
package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"

	"github.com/mroth/porcelain/statusv2"
	"github.com/yassinebenaid/godump"
)

var (
	porcelainVersion = flag.String("format", "v2", "porcelain version to parse [v2, v2z]")
)

func main() {
	flag.Parse()
	var parserMap = map[string]func(io.Reader) (*statusv2.Status, error){
		"v2":  statusv2.Parse,
		"v2z": statusv2.ParseZ,
	}
	parser, found := parserMap[*porcelainVersion]
	if !found {
		log.Fatalf("unsupported -porcelain version: %s", *porcelainVersion)
	}

	in := bufio.NewReader(os.Stdin)
	results, err := parser(in)
	if err != nil {
		log.Fatalf("error parsing porcelain output: %v", err)
	}
	godump.Dump(*results)
}
