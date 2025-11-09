package tre

import (
	"fmt"
	"slices"
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
	ParseOpt
)

type Parsed struct {
	typ   ParseType
	left  *Parsed
	right *Parsed
	class Ranges // ParseClass
	caps  []int  // ParseClass
}

func (p *Parsed) Print(indent int) {
	tab := strings.Repeat("  ", indent)
	switch p.typ {
	case ParseClass:
		fmt.Printf("%s%v class=%v caps=%v\n", tab, p.typ, p.class, p.caps)
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

const debugTrace bool = false

func (p *Lexer) debug(s string) func() {
	if debugTrace {
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

const EOF rune = -1

func (p *Lexer) advance() {
	if p.pos+1 < len(p.inp) {
		p.pos++
		p.cur = p.inp[p.pos]
	} else {
		p.cur = EOF
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
	if ch == EOF {
		return "EOF"
	}
	return fmt.Sprintf("%q", ch)
}

func ParseExpect(p *Lexer, want rune) error {
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
	if ch != EOF && unicode.IsPunct(ch) {
		return ch, nil
	}
	switch ch {
	case 'r':
		return '\r', nil
	case 'n':
		return '\n', nil
	default:
		return 0, fmt.Errorf("%d: unexpected %v after \\", pos-1, showRune(ch))
	}
}

func parseClassChar(p *Lexer, terminal rune) (rune, error) {
	pos := p.pos
	ch := p.next()
	switch ch {
	case '\\':
		var err error
		ch, err = parseEscaped(p)
		if err != nil {
			return 0, err
		}
	default:
		if ch == EOF || ch == terminal || strings.ContainsRune(reservedChars, ch) || !unicode.IsGraphic(ch) {
			return 0, fmt.Errorf("%d: unexpected %v", pos, showRune(ch))
		}
	}
	return ch, nil
}

func parseClassRange(p *Lexer, terminal rune) (rune, rune, error) {
	defer p.debug("parseReClassRange")()
	start, err := parseClassChar(p, terminal)
	if err != nil {
		return 0, 0, err
	}

	if p.peek() == '-' {
		p.advance()
		end, err := parseClassChar(p, terminal)
		if err != nil {
			return 0, 0, err
		}
		if end < start {
			return 0, 0, fmt.Errorf("class range from %v to %v is empty", showRune(start), showRune(end))
		}
		return start, end, nil
	} else {
		return start, start, nil
	}
}

func parseClass(p *Lexer, terminal rune) (Ranges, error) {
	defer p.debug("parseReClass")()
	invert := false
	if p.peek() == '^' {
		p.advance()
		invert = true
	}

	var rs Ranges
	for p.peek() != ']' {
		start, end, err := parseClassRange(p, terminal)
		if err != nil {
			return nil, err
		}
		rs.Add(start, end)
	}

	if err := ParseExpect(p, ']'); err != nil {
		return nil, err
	}

	if invert {
		rs = rs.Invert()
	}
	return rs, nil
}

func parseReChar(p *Lexer, terminal rune) (rune, error) {
	defer p.debug("parseReChar")()
	pos := p.pos
	ch := p.next()
	switch ch {
	case '\\':
		var err error
		ch, err = parseEscaped(p)
		if err != nil {
			return 0, err
		}
	default:
		if ch == EOF || ch == terminal || strings.ContainsRune(reservedChars, ch) || !unicode.IsGraphic(ch) {
			return 0, fmt.Errorf("%d: Unexpected %v", pos, showRune(ch))
		}
	}
	return ch, nil
}

type Parser struct {
	capNum  int
	curCaps []int
}

// parseReAtom parses an re which is not compound or is parenthesized.
// reAtom := "." | char | charclass | ( "?"? re )
func parseReAtom(parser *Parser, lex *Lexer, terminal rune) (*Parsed, error) {
	defer lex.debug("parseReAtom")()
	pos := lex.pos
	peek := lex.peek()
	switch peek {
	case '(':
		lex.advance()
		capNum := 0
		var prevCaps []int
		if lex.peek() == '?' {
			lex.advance()

			parser.capNum++
			capNum = parser.capNum
			prevCaps = parser.curCaps
			parser.curCaps = append(slices.Clone(prevCaps), capNum)
		}

		re1, err := ParseRe(parser, lex, terminal)
		if err != nil {
			return nil, err
		}
		if err := ParseExpect(lex, ')'); err != nil {
			return nil, err
		}

		if capNum != 0 {
			parser.curCaps = prevCaps
			//return &Parsed{typ: ParseCap, left: re1, capNum: capNum}, nil
		}
		return re1, nil

	case '|' | '*' | '+' | '?':
		lex.next()
		return nil, fmt.Errorf("%d: unexpected %v", pos, showRune(peek))

	case '[':
		lex.next()
		rs, err := parseClass(lex, terminal)
		if err != nil {
			return nil, err
		}
		return &Parsed{typ: ParseClass, class: rs, caps: parser.curCaps}, nil

	case '.':
		lex.next()
		return &Parsed{typ: ParseClass, class: FullRanges(), caps: parser.curCaps}, nil

	default:
		ch, err := parseReChar(lex, terminal)
		if err != nil {
			return nil, err
		}
		return &Parsed{typ: ParseClass, class: newRange1(ch), caps: parser.curCaps}, nil
	}
}

// reConcat := reAtom ("*" | "+" | "?") reConcat*
func parseReConcat(parser *Parser, lex *Lexer, terminal rune) (*Parsed, error) {
	defer lex.debug("parseReConcat")()
	re1, err := parseReAtom(parser, lex, terminal)
	if err != nil {
		return nil, err
	}

	for lex.peek() != EOF && lex.peek() != terminal && lex.peek() != ')' && lex.peek() != '|' {
		switch lex.peek() {
		case '*':
			lex.advance()
			re1 = &Parsed{typ: ParseStar, left: re1}
		case '+':
			lex.advance()
			re1 = &Parsed{typ: ParsePlus, left: re1}
		case '?':
			lex.advance()
			re1 = &Parsed{typ: ParseOpt, left: re1}
		default:
			re2, err := parseReConcat(parser, lex, terminal)
			if err != nil {
				return nil, err
			}
			re1 = &Parsed{typ: ParseConcat, left: re1, right: re2}
		}
	}
	return re1, nil
}

// ParseRe parses a regular expression from lex which is terminated by terminal,
// usually EOF. It is the main entry point for parsing regular expressions.
// On successful parse the lexer should be at the terminal rune.
//
// ParseRe parses an re which may be compound.
// re := reConcat ("|" re)*
func ParseRe(parser *Parser, lex *Lexer, terminal rune) (*Parsed, error) {
	defer lex.debug("parseRe")()
	re1, err := parseReConcat(parser, lex, terminal)
	if err != nil {
		return nil, err
	}

	for lex.peek() == '|' {
		lex.advance()
		re2, err := ParseRe(parser, lex, terminal)
		if err != nil {
			return nil, err
		}
		re1 = &Parsed{typ: ParseAlt, left: re1, right: re2}
	}
	return re1, nil
}

func Parse(s string) (*Parsed, error) {
	parser := &Parser{}
	lex := newLexer(s)
	re, err := ParseRe(parser, lex, EOF)
	if err != nil {
		return nil, err
	}

	if err := ParseExpect(lex, EOF); err != nil {
		return nil, err
	}
	return re, nil
}

// ParseBounded parses an RE that is bounded by punctuation, such as /re/.
func ParseBounded(s string) (*Parsed, error) {
	parser := &Parser{}
	lex := newLexer(s)

	pos := lex.pos
	terminal := lex.next()
	if terminal == EOF || !unicode.IsPunct(terminal) {
		return nil, fmt.Errorf("%d: unexpected bounding character %v", pos, showRune(terminal))
	}

	re, err := ParseRe(parser, lex, terminal)
	if err != nil {
		return nil, err
	}

	if err := ParseExpect(lex, terminal); err != nil {
		return nil, err
	}
	if err := ParseExpect(lex, EOF); err != nil {
		return nil, err
	}
	return re, nil
}
