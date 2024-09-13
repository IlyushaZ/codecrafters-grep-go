package main

import (
	"unicode"
)

func matchString(s string, pattern string) (bool, error) {
	tokens, err := parseString(pattern)
	if err != nil {
		return false, err
	}

	origLine := s
	match := false

	for i := 0; i < len(origLine) && !match; i++ {
		s = origLine[i:]
		match = matchFromStart(s, tokens)
	}

	return match, nil
}

func matchFromStart(s string, regex []token) bool {
	for i := 0; i < len(s); i++ {
		if i >= len(regex) {
			break
		}

		if i == len(s)-1 && i < len(regex)-1 { // TODO: maybe this can be improved?
			return false
		}

		switch t := regex[i].(type) {
		case char:
			if s[i] != byte(t) {
				return false
			}

		case anyDigit:
			if !isDigit(s[i]) {
				return false
			}

		case anyChar:
			if !isLetter(s[i]) {
				return false
			}

		case charGroup:
			contains := false
			for _, c := range t.chars {
				if s[i] == c {
					contains = true
				}
			}

			if t.negative == contains {
				return false
			}
		}
	}

	return true
}

func isDigit(char byte) bool {
	return unicode.IsDigit(rune(char))
}

func isLetter(char byte) bool {
	return unicode.IsLetter(rune(char))
}
