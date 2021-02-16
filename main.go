package main

import (
	"os"

	"github.com/carlmjohnson/exitcode"
	"github.com/spotlightpa/kristy/sitter"
)

func main() {
	exitcode.Exit(sitter.CLI(os.Args[1:]))
}
