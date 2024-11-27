package main

import (
	"fmt"
	"strings"
	"unicode"
)

func MatchString(pattern, s string) (bool, error) {
	p, err := parsePattern(pattern)
	if err != nil {
		return false, err
	}

	if len(p) == 0 {
		return true, nil
	}

	m := newMatcher()

	if _, ok := p[0].(startOfString); ok {
		return m.matchHere(p[1:], s).match, nil
	}

	orig := s
	match := false

	for i := 0; i < len(orig) && !match; i++ {
		s = orig[i:]
		match = m.matchHere(p, s).match
	}

	return match, nil
}

type matcher struct {
	groupID  int
	captured map[int]string
	next     token // for corner-cases
}

func newMatcher() *matcher {
	return &matcher{groupID: 0, captured: make(map[int]string)}
}

type matchResult struct {
	match bool
	end   int // non-zero if matched
}

func (m *matcher) matchHere(pattern []token, s string) matchResult {
	pos := 0 // position in s

	for i, tkn := range pattern {
		switch t := tkn.(type) {
		case endOfString:
			return matchResult{pos == len(s), pos}

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

			m.groupID++
			currentGroup := m.groupID

			// we need to save the next token for cases
			// when capture group contains a pattern that will match the whole s until the end
			if i != len(pattern)-1 {
				m.next = pattern[i+1]
			}

			for _, p := range t.patterns {
				if mr := m.matchHere(p, s[pos:]); mr.match {
					match = true
					m.captured[currentGroup] = s[pos : pos+mr.end]
					pos += mr.end
					m.next = nil
					break
				}
			}

			if !match {
				return matchResult{false, 0}
			}

			continue

		case backReference:
			val := m.captured[int(t)+1]

			if val == "" {
				panic(fmt.Sprintf("invalid back reference %d, %v", t, m.captured))
			}

			if len(s[pos:]) < len(val) {
				return matchResult{false, 0}
			}

			if !strings.HasPrefix(s[pos:], val) {
				return matchResult{false, 0}
			}

			pos += len(val)

			continue // continue without incrementing the pos

		case zeroOrOne:
			prev := pattern[i-1]
			if !m.matchHere([]token{prev}, s[pos:pos+1]).match {
				continue // it's zero, no need to go forward
			}

		case oneOrMore:
			prev := pattern[i-1]
			if i != len(pattern)-1 {
				m.next = pattern[i+1]
			}

			for {
				if pos >= len(s) {
					break
				}

				// cover corner-case when + is preceded by .
				// . will match every char until the end of string
				// and there will be no chance that next matchHere call will result in false
				if m.next != nil && m.matchHere([]token{m.next}, s[pos:pos+1]).match {
					break
				}

				if !m.matchHere([]token{prev}, s[pos:pos+1]).match {
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
