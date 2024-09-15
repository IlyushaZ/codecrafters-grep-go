package main

import (
	"unicode"
)

func MatchString(pattern, s string) (bool, error) {
	tokens, err := parsePattern(pattern)
	if err != nil {
		return false, err
	}

	if len(tokens) == 0 {
		return true, nil
	}

	if _, ok := tokens[0].(startOfString); ok {
		return matchHere(tokens[1:], s), nil
	}

	orig := s
	match := false

	for i := 0; i < len(orig) && !match; i++ {
		s = orig[i:]
		match = matchHere(tokens, s)
	}

	return match, nil
}

func matchHere(pattern []token, s string) bool {
	pos := 0 // position in s

	for i, tkn := range pattern {
		switch t := tkn.(type) {
		case endOfString:
			return pos == len(s)

		case anyDigit:
			if pos == len(s) || !isDigit(s[pos]) {
				return false
			}

		case anyLetter:
			if pos == len(s) || !isLetter(s[pos]) {
				return false
			}

		case wildcard:
			// do nothing, because any value matches it

		case char:
			if pos == len(s) || s[pos] != byte(t) {
				// check so we can safely check the next token
				if last := i == len(pattern)-1; last {
					return false
				}

				if _, optional := pattern[i+1].(zeroOrMore); !optional {
					return false
				}

				// continue without incrementing the pos in order to cover the case with zero occurrences
				continue
			}

		case alteration:
			match := false

			for _, w := range t.words {
				ts := tokenizeString(w)

				if matchHere(ts, s[pos:pos+len(ts)]) {
					match = true
					pos += len(ts)
					break
				}
			}

			if !match {
				return false
			}

			// continue without incrementing the pos because we've already moved it after matching
			continue

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

		case oneOrMore, zeroOrMore:
			prev := pattern[i-1]

			for {
				if pos >= len(s) {
					break
				}

				match := matchHere([]token{prev}, s[pos:pos+1])
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

func tokenizeString(s string) []token {
	ts := make([]token, 0, len(s))
	for i := 0; i < len(s); i++ {
		ts = append(ts, char(s[i]))
	}
	return ts
}
