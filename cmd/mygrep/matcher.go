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

func matchHere(s string, regex []token) bool {
	currToken := 0

	for i := 0; i < len(s); i++ {
		if i >= len(regex) {
			break
		}

		// TODO: refactor me?
		// if i is last element but regex is not over
		if i == len(s)-1 && i < len(regex)-1 {
			break
		}

		switch t := regex[i].(type) {
		case endOfString:
			return false

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

		currToken++
	}

	if currToken < len(regex)-1 {
		_, ok := regex[currToken+1].(endOfString)
		return ok
	}

	return true
}

func isDigit(char byte) bool {
	return unicode.IsDigit(rune(char))
}

func isLetter(char byte) bool {
	return unicode.IsLetter(rune(char))
}
