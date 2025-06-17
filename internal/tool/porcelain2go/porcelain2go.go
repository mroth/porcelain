// Command porcelain2go converts porcelain output of `git status` into a Go struct.
// It reads from stdin and writes to stdout, so it can be used in a pipeline.
// It is primarily intended for use in testing and debugging on the CLI.
package main

import (
	"bufio"
	"flag"
	"log"
	"os"

	"github.com/mroth/porcelain/statusv2"
	"github.com/yassinebenaid/godump"
)

var (
	porcelainVersion = flag.String("porcelain", "v2", "porcelain version to parse (e.g. v2)")
)

func main() {
	flag.Parse()
	if *porcelainVersion != "v2" {
		log.Fatal("only -porcelain=v2 is currently supported")
	}

	in := bufio.NewReader(os.Stdin)
	results, err := statusv2.Parse(in)
	if err != nil {
		log.Fatalf("error parsing porcelain output: %v", err)
	}
	godump.Dump(*results)
}
