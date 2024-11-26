package main

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrSyntax        = errors.New("syntax error")
	ErrUnexpectedEnd = fmt.Errorf("%w: unexpected end of pattern", ErrSyntax)
)

type token interface {
	fmt.Stringer
	isToken()
}

type (
	char          byte
	anyDigit      struct{} // \d
	anyLetter     struct{} // \w
	startOfString struct{} // ^
	endOfString   struct{} // $
	oneOrMore     struct{} // +
	zeroOrOne     struct{} // ?
	zeroOrMore    struct{} // *
	wildcard      struct{} // .
	charGroup     struct {
		chars    []byte // enumeration of all possible chars
		negative bool
	} // [abc] or [^abc]
	captureGroup  struct{ patterns [][]token } // ([abc]+)
	backReference int                          // \1
)

func parsePattern(s string) ([]token, error) {
	tokens := []token{}

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			if i == len(s)-1 {
				return nil, syntaxErrorf("unexpected end of pattern string")
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
				if '0' < next && next <= '9' {
					ref := int(next - 48)

					tokens = append(tokens, backReference(ref-1)) // make it start from zero instead of one
				} else {
					return nil, syntaxErrorf(`expected '\d' or '\w', got '\%s'`, string(next))
				}
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

			tokens = append(tokens, cg)

		case '^':
			tokens = append(tokens, startOfString{})

		case '$':
			tokens = append(tokens, endOfString{})

		case '+':
			if i == 0 {
				return nil, syntaxErrorf("expected '+' to have preceding token")
			}

			tokens = append(tokens, oneOrMore{})

		case '?':
			if i == 0 {
				return nil, syntaxErrorf("expected '?' to have preceding token")
			}

			tokens = append(tokens, zeroOrOne{})

		case '*':
			if i == 0 {
				return nil, syntaxErrorf("expected '*' to have preceding token")
			}

			tokens = append(tokens, zeroOrMore{})

		case '.':
			tokens = append(tokens, wildcard{})

		case '(':
			closing := closingParentesis(s[i:])
			if closing == -1 {
				return nil, syntaxErrorf("alteration: expected ')'")
			}

			content := s[i+1 : i+closing]

			var patterns [][]token

			if !hasNestedGroup(content) {
				words := strings.Split(content, "|")
				patterns = make([][]token, 0, len(words))

				for _, w := range words {
					p, err := parsePattern(w)
					if err != nil {
						return nil, err
					}

					patterns = append(patterns, p)
				}
			} else {
				internal, err := parsePattern(content)
				if err != nil {
					return nil, err
				}

				patterns = [][]token{internal}
			}

			tokens = append(tokens, captureGroup{patterns})

			i += closing

		default:
			tokens = append(tokens, char(s[i]))
		}
	}

	return tokens, nil
}

// todo: improve me
func hasNestedGroup(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '(' {
			return true
		}
	}
	return false
}

func closingParentesis(s string) int {
	opened := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			opened++
		case ')':
			opened--
			if opened == 0 {
				return i
			}
		}
	}

	return -1
}

func syntaxErrorf(format string, a ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrSyntax, fmt.Sprintf(format, a...))
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

func (zeroOrOne) isToken() {}
func (zeroOrOne) String() string {
	return "?"
}

func (zeroOrMore) isToken() {}
func (zeroOrMore) String() string {
	return "*"
}

func (wildcard) isToken() {}
func (wildcard) String() string {
	return "."
}

func (captureGroup) isToken() {}
func (cg captureGroup) String() string {
	sb := &strings.Builder{}

	sb.WriteString("(")
	for i, p := range cg.patterns {
		for _, t := range p {
			sb.WriteString(t.String())
		}

		if i != len(cg.patterns)-1 {
			sb.WriteString("|")
		}
	}
	sb.WriteString(")")

	return sb.String()
}

func (backReference) isToken() {}
func (b backReference) String() string {
	return fmt.Sprintf(`\%d`, b+1)
}
