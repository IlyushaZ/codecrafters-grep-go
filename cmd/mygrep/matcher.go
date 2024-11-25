package main

import (
	"fmt"
	"strings"
	"unicode"
)

func MatchString(pattern, s string) (bool, error) {
	ts, err := parsePattern(pattern)
	if err != nil {
		return false, err
	}

	if len(ts) == 0 {
		return true, nil
	}

	if _, ok := ts[0].(startOfString); ok {
		return matchHere(ts[1:], s).match, nil
	}

	orig := s
	match := false

	for i := 0; i < len(orig) && !match; i++ {
		s = orig[i:]
		match = matchHere(ts, s).match
	}

	return match, nil
}

type matchResult struct {
	match bool
	end   int // non-zero if matched
}

func matchHere(pattern []token, s string) matchResult {
	pos := 0 // position in s

	captured := []string{}

	for i, tkn := range pattern {
		switch t := tkn.(type) {
		case endOfString:
			return matchResult{pos == len(s), pos} // TODO: pos?

		case anyDigit:
			if pos == len(s) || !isDigit(s[pos]) {
				return matchResult{false, 0}
			}

		case anyLetter:
			if pos == len(s) || !isLetter(s[pos]) {
				return matchResult{false, 0}
			}

		case wildcard:
			// do nothing, because any value matches it

		case char:
			if pos == len(s) || s[pos] != byte(t) {
				// check so we can safely check the next token
				if last := i == len(pattern)-1; last {
					return matchResult{false, 0}
				}

				if _, zeroOrOne := pattern[i+1].(zeroOrOne); !zeroOrOne {
					return matchResult{false, 0}
				}

				// continue without incrementing the pos in order to cover the case with zero occurrences
				continue
			}

		case charGroup:
			if pos == len(s) {
				return matchResult{false, 0}
			}

			contains := false
			for _, c := range t.chars {
				if s[pos] == c {
					contains = true
				}
			}

			if t.negative == contains {
				return matchResult{false, 0}
			}

		case captureGroup:
			match := false

			for _, p := range t.patterns {
				if m := matchHere(p, s[pos:]); m.match {
					match = true
					captured = append(captured, s[pos:pos+m.end])
					pos += m.end
				}
			}

			if !match {
				return matchResult{false, 0}
			}

			continue

		case backReference:
			if len(captured) < int(t) { // normally should not happen, as it's checked by the parser
				panic(fmt.Sprintf("invalid back reference %d. len(captured)=%d", t, len(captured)))
			}

			if len(s[pos:]) < len(captured[t]) {
				return matchResult{false, 0}
			}

			if !strings.HasPrefix(s[pos:], captured[t]) {
				return matchResult{false, 0}
			}

			pos += len(captured[t])

			continue // continue without incrementing the pos

		case zeroOrOne:
			prev := pattern[i-1]
			if !matchHere([]token{prev}, s[pos:pos+1]).match {
				continue // it's zero, no need to go forward
			}

		case oneOrMore:
			prev := pattern[i-1]
			var next token
			if i != len(pattern)-1 {
				next = pattern[i+1]
			}

			for {
				if pos >= len(s) {
					break
				}

				// cover corner-case when + is preceded by .
				// . will match every char until the end of string
				// and there will be no chance that next matchHere call will result in false
				if next != nil && matchHere([]token{next}, s[pos:pos+1]).match {
					break
				}

				if !matchHere([]token{prev}, s[pos:pos+1]).match {
					break
				}

				pos++
			}

			continue // avoid incrementing pos one more time
		}

		pos++
	}

	return matchResult{true, pos}
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
