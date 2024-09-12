package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode"
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

	ok := matchLine(line, pattern)
	if !ok {
		os.Exit(1)
	}
}

func matchLine(line []byte, pattern string) (ok bool) {
	switch pattern {
	case `\d`:
		ok = bytes.ContainsAny(line, "012345678")
	case `\w`:
		for _, char := range line {
			if unicode.IsLetter(rune(char)) || unicode.IsDigit(rune(char)) {
				return true
			}
		}
	default:
		ok = bytes.ContainsAny(line, pattern)
	}

	return ok
}
