package main

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidPattern = errors.New("invalid pattern")
	ErrUnexpectedEnd  = fmt.Errorf("%w: unexpected end of pattern", ErrInvalidPattern)
)

type token interface {
	fmt.Stringer
	isToken()
}

type char byte

type anyDigit struct{}

type anyLetter struct{}

type charGroup struct {
	chars    []byte // enumeration of all possible chars
	negative bool
}

type startOfString struct{}

type endOfString struct{}

type oneOrMore struct{}

type zeroOrMore struct{}

type wildcard struct{}

type alteration struct {
	words []string
}

func parsePattern(s string) ([]token, error) {
	tokens := []token{}

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			if i == len(s)-1 {
				return nil, fmt.Errorf("%w: unexpected end of pattern string", ErrInvalidPattern)
			}

			switch next := s[i+1]; next {
			case 'd':
				tokens = append(tokens, anyDigit{})
			case 'w':
				tokens = append(tokens, anyLetter{})
			case '\\':
				tokens = append(tokens, char(s[i]))
				i++
			default:
				return nil, fmt.Errorf(`%w: expected '\d' or '\w', got '\%s'`, next)
			}

			i++

		case '[':
			i++
			if i == len(s) {
				return nil, ErrUnexpectedEnd
			}

			cg := charGroup{
				negative: s[i] == '^',
			}
			if cg.negative {
				i++ // go forward to the actual character group
			}

			for ; i < len(s) && s[i] != ']'; i++ {
				cg.chars = append(cg.chars, s[i])
			}

			if i == len(s) { // make sure we're here not because the string ended but because of group ended
				return nil, ErrUnexpectedEnd
			}

			i++ // skip closing bracket

			tokens = append(tokens, cg)

		case '^':
			tokens = append(tokens, startOfString{})

		case '$':
			tokens = append(tokens, endOfString{})

		case '+':
			if i == 0 {
				return nil, fmt.Errorf("%w: expected '+' to have preceding token", ErrInvalidPattern)
			}

			tokens = append(tokens, oneOrMore{})

		case '?':
			if i == 0 {
				return nil, fmt.Errorf("%w: expected '?' to have preceding token", ErrInvalidPattern)
			}

			tokens = append(tokens, zeroOrMore{})

		case '.':
			tokens = append(tokens, wildcard{})

		case '(':
			i++
			if i == len(s) {
				return nil, ErrUnexpectedEnd
			}

			closing := strings.Index(s[i:], ")")
			if closing == -1 {
				return nil, fmt.Errorf("%w: alteration: expected ')'", ErrInvalidPattern)
			}

			a := alteration{
				words: strings.Split(s[i:i+closing], "|"),
			}
			tokens = append(tokens, a)

			i += closing

		default:
			tokens = append(tokens, char(s[i]))
		}
	}

	return tokens, nil
}

func (char) isToken()         {}
func (c char) String() string { return fmt.Sprintf("%c", byte(c)) }

func (anyDigit) isToken()         {}
func (a anyDigit) String() string { return "\\d" }

func (anyLetter) isToken()         {}
func (a anyLetter) String() string { return "\\w" }

func (charGroup) isToken() {}
func (cg charGroup) String() string {
	sb := strings.Builder{}
	sb.WriteByte('[')

	if cg.negative {
		sb.WriteByte('^')
	}

	for _, c := range cg.chars {
		sb.WriteByte(c)
	}

	sb.WriteByte(']')

	return sb.String()
}

func (startOfString) isToken() {}
func (startOfString) String() string {
	return "^"
}

func (endOfString) isToken() {}
func (endOfString) String() string {
	return "$"
}

func (oneOrMore) isToken() {}
func (oneOrMore) String() string {
	return "+"
}

func (zeroOrMore) isToken() {}
func (zeroOrMore) String() string {
	return "?"
}

func (wildcard) isToken() {}
func (wildcard) String() string {
	return "."
}

func (alteration) isToken() {}
func (a alteration) String() string {
	return "(" + strings.Join(a.words, "|") + ")"
}
