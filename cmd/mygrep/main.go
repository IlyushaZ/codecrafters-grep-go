package main

import (
	"fmt"
	"io"
	"os"
)

// Usage: echo <input_text> | your_grep.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	match, err := MatchString(pattern, string(line))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: can't match string: %v", err)
		os.Exit(2)
	}

	if !match {
		os.Exit(1)
	}
}
