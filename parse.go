package tre

import (
	"fmt"
	"strings"
	"unicode"
)

const reservedChars = "\\()[]|*+-"

type ParseType int

//go:generate go run golang.org/x/tools/cmd/stringer -type=ParseType
const (
	ParseErr ParseType = iota
	ParseClass
	ParseConcat
	ParseAlt
	ParseStar
	ParsePlus
)

type Parsed struct {
	typ   ParseType
	left  *Parsed
	right *Parsed
	class Ranges
}

func (p *Parsed) Print(indent int) {
	tab := strings.Repeat("  ", indent)
	switch p.typ {
	case ParseClass:
		fmt.Printf("%s%v class=%v\n", tab, p.typ, p.class)
	default:
		fmt.Printf("%s%v\n", tab, p.typ)
	}
	if p.left != nil {
		p.left.Print(indent + 1)
	}
	if p.right != nil {
		p.right.Print(indent + 1)
	}
}

type Lexer struct {
	inp []rune
	cur rune
	pos int

	dbgIndent int
}

func newLexer(s string) *Lexer {
	l := &Lexer{
		inp: []rune(s),
		pos: -1,
	}
	l.advance()
	return l
}

func (p *Lexer) debug(s string) func() {
	if false {
		p.dbgIndent++
		tab := strings.Repeat("--", p.dbgIndent)
		fmt.Printf("%s %s at %d %q\n", tab, s, p.pos, string(p.inp[p.pos:]))
		return func() {
			fmt.Printf("%s end %s\n", tab, s)
			p.dbgIndent--
		}
	}
	return func() {}
}

func (p *Lexer) advance() {
	if p.pos+1 < len(p.inp) {
		p.pos++
		p.cur = p.inp[p.pos]
	} else {
		p.cur = 0 // EOF marker. hackity.
	}
}

func (p *Lexer) peek() rune {
	return p.cur
}

func (p *Lexer) next() rune {
	cur := p.cur
	p.advance()
	return cur
}

func showRune(ch rune) string {
	if ch == 0 {
		return "EOF"
	}
	return fmt.Sprintf("%q", ch)
}

func parseExpect(p *Lexer, want rune) error {
	pos := p.pos
	ch := p.next()
	switch {
	case ch == want:
		return nil
	default:
		return fmt.Errorf("%d: expected %v got %v", pos, showRune(want), showRune(ch))
	}
}

func parseEscaped(p *Lexer) (rune, error) {
	pos := p.pos
	ch := p.next()
	if strings.ContainsRune(reservedChars, ch) {
		return ch, nil
	}
	switch ch {
	case 'r':
		return '\r', nil
	case 'n':
		return '\n', nil
	case 0:
		return 0, fmt.Errorf("%d: unexpected EOF after \\", pos)
	default:
		return 0, fmt.Errorf("%d: unexpected %q", pos-1, "\\"+string(ch))
	}
}

func parseClassChar(p *Lexer) (rune, error) {
	pos := p.pos
	ch := p.next()
	switch ch {
	case '\\':
		var err error
		ch, err = parseEscaped(p)
		if err != nil {
			return 0, err
		}
	case 0:
		return 0, fmt.Errorf("%d: unexpected EOF", pos)
	default:
		if strings.ContainsRune(reservedChars, ch) || !unicode.IsGraphic(ch) {
			return 0, fmt.Errorf("unexpected %q", ch)
		}
	}
	return ch, nil
}

func parseClassRange(p *Lexer) (rune, rune, error) {
	defer p.debug("parseReClassRange")()
	start, err := parseClassChar(p)
	if err != nil {
		return 0, 0, err
	}

	if p.peek() == '-' {
		p.advance()
		end, err := parseClassChar(p)
		if err != nil {
			return 0, 0, err
		}
		return start, end, nil
	} else {
		return start, start, nil
	}
}

func parseClass(p *Lexer) (Ranges, error) {
	defer p.debug("parseReClass")()
	invert := false
	if p.peek() == '^' {
		p.advance()
		invert = true
	}

	var rs Ranges
	for p.peek() != ']' {
		start, end, err := parseClassRange(p)
		if err != nil {
			return nil, err
		}
		rs.Add(start, end)
	}

	if err := parseExpect(p, ']'); err != nil {
		return nil, err
	}

	if invert {
		rs = rs.Invert()
	}
	return rs, nil
}

func parseReChar(p *Lexer) (rune, error) {
	defer p.debug("parseReChar")()
	pos := p.pos
	ch := p.next()
	switch ch {
	case 0:
		return 0, fmt.Errorf("%d: Unexpected EOF", pos)
	case '\\':
		var err error
		ch, err = parseEscaped(p)
		if err != nil {
			return 0, err
		}
	default:
		if strings.ContainsRune(reservedChars, ch) || !unicode.IsGraphic(ch) {
			return 0, fmt.Errorf("%d: Unexpected %q", pos, ch)
		}
	}
	return ch, nil
}

// parseReAtom parses an re which is not compound or is parenthesized.
// reAtom := char | charclass | ( re )
func parseReAtom(lex *Lexer) (*Parsed, error) {
	defer lex.debug("parseReAtom")()
	pos := lex.pos
	peek := lex.peek()
	switch peek {
	case '(':
		lex.advance()
		re1, err := parseRe(lex)
		if err != nil {
			return nil, err
		}
		if err := parseExpect(lex, ')'); err != nil {
			return nil, err
		}
		return re1, nil

	case '|' | '*' | '+':
		lex.next()
		return nil, fmt.Errorf("%d: unexpected %q", pos, peek)

	case '[':
		lex.next()
		rs, err := parseClass(lex)
		if err != nil {
			return nil, err
		}
		return &Parsed{typ: ParseClass, class: rs}, nil

	default:
		ch, err := parseReChar(lex)
		if err != nil {
			return nil, err
		}
		return &Parsed{typ: ParseClass, class: newRange1(ch)}, nil
	}
}

// reConcat := reAtom ("*" | "+") reConcat*
func parseReConcat(lex *Lexer) (*Parsed, error) {
	defer lex.debug("parseReConcat")()
	re1, err := parseReAtom(lex)
	if err != nil {
		return nil, err
	}

	for lex.peek() != 0 && lex.peek() != ')' && lex.peek() != '|' {
		switch lex.peek() {
		case '*':
			lex.advance()
			re1 = &Parsed{typ: ParseStar, left: re1}
		case '+':
			lex.advance()
			re1 = &Parsed{typ: ParsePlus, left: re1}
		default:
			re2, err := parseReConcat(lex)
			if err != nil {
				return nil, err
			}
			re1 = &Parsed{typ: ParseConcat, left: re1, right: re2}
		}
	}
	return re1, nil
}

// parseRe parses an re which may be compound.
// re := reConcat ("|" re)*
func parseRe(lex *Lexer) (*Parsed, error) {
	defer lex.debug("parseRe")()
	re1, err := parseReConcat(lex)
	if err != nil {
		return nil, err
	}

	for lex.peek() == '|' {
		lex.advance()
		re2, err := parseRe(lex)
		if err != nil {
			return nil, err
		}
		re1 = &Parsed{typ: ParseAlt, left: re1, right: re2}
	}
	return re1, nil
}

func Parse(s string) (*Parsed, error) {
	lex := newLexer(s)
	re, err := parseRe(lex)
	if err != nil {
		return nil, err
	}

	if err := parseExpect(lex, 0); err != nil {
		return nil, err
	}
	return re, nil
}
