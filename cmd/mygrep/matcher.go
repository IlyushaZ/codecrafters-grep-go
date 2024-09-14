package main

import (
	"unicode"
)

func matchString(s string, pattern string) (bool, error) {
	tokens, err := parseString(pattern)
	if err != nil {
		return false, err
	}

	if len(tokens) == 0 {
		return true, nil
	}

	if _, ok := tokens[0].(startOfString); ok {
		return matchHere(s, tokens[1:]), nil
	}

	origLine := s
	match := false

	for i := 0; i < len(origLine) && !match; i++ {
		s = origLine[i:]
		match = matchHere(s, tokens)
	}

	return match, nil
}

func matchHere(s string, pattern []token) bool {
	pos := 0 // position in s

	for i, tkn := range pattern {
		switch t := tkn.(type) {
		case endOfString:
			return pos == len(s)

		case char:
			if pos == len(s) || s[pos] != byte(t) {
				return false
			}

		case anyDigit:
			if pos == len(s) || !isDigit(s[pos]) {
				return false
			}

		case anyChar:
			if pos == len(s) || !isLetter(s[pos]) {
				return false
			}

		case charGroup:
			if pos == len(s) {
				return false
			}

			contains := false
			for _, c := range t.chars {
				if s[pos] == c {
					contains = true
				}
			}

			if t.negative == contains {
				return false
			}

		case oneOrMore:
			prev := pattern[i-1]

			for {
				if pos >= len(s) {
					break
				}

				match := matchHere(s[pos:pos+1], []token{prev})
				if !match {
					break
				}

				pos++
			}

			continue // avoid incrementing pos one more time
		}

		pos++
	}

	return true
}

func isDigit(char byte) bool {
	return unicode.IsDigit(rune(char))
}

func isLetter(char byte) bool {
	return unicode.IsLetter(rune(char))
}
