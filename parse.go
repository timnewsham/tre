package tre

import (
	"bufio"
	"fmt"
	"strings"
	"unicode"
)

type TokType int

//go:generate go run golang.org/x/tools/cmd/stringer -type=TokType
const (
	TokErr TokType = iota
	TokAlt
	TokStar
	TokPlus
	TokOparen
	TokCparen
	TokDot
	TokChar
	TokClass
)

type Token struct {
	typ TokType
	lno int
	lpos int
	lit string
	class Ranges
	err error
}

func (t Token) String() string {
	return fmt.Sprintf("Token %v %d:%d %q %v %v", t.typ, t.lno, t.lpos, t.lit, t.class,  t.err)
}

type Lexer struct {
	inp *bufio.Reader
	err error
	lno int
	lpos int
}

func newLexer(s string) *Lexer {
	return &Lexer{
		inp: bufio.NewReader(strings.NewReader(s)),
		err: nil,
		lno: 1,
		lpos: 1,
	}
}

func (p *Lexer) setErr(err error) {
	if p.err == nil {
		p.err = err
	}
}

// nextRune returns the next rune in the input stream.
// On error, returns NIL with an error set.
func (p *Lexer) nextRune() rune {
	if p.err != nil {
		return 0
	}

	ch, _, err := p.inp.ReadRune()
	if err != nil {
		p.err = err
		return ch
	}

	if ch == 0 {
		p.setErr(fmt.Errorf("unexpected NIL"))
		return 0
	}

	if ch == '\n' {
		p.lno ++
		p.lpos = 1
	} else {
		p.lpos ++
	}

	return ch
}

func (p *Lexer) peek() rune {
	if p.err != nil {
		return 0
	}

	ch, _, err := p.inp.ReadRune()
	if err != nil {
		p.err = err
		return ch
	}
	p.inp.UnreadRune()
	return ch
}

func (p *Lexer) nextEscaped() rune {
	ch := p.nextRune()
	if strings.ContainsRune("\\()|[]*+", ch) {
		return ch
	}
	switch ch {
	case 'r':
		return '\r'
	case 'n':
		return '\n'
	default:
		p.setErr(fmt.Errorf("unexpected %q", "\\" + string(ch)))
		return 0
	}
}

func (p *Lexer) nextClass() Ranges {
	invert := false
	if p.peek() == '^' {
		invert = true
		p.nextRune()
	}

	var rs Ranges
loop:
	for {
		ch := p.nextRune()
		switch(ch) {
		case 0:
			return rs
		case '[':
			p.setErr(fmt.Errorf("unexpected %q", ch))
			return rs
		case ']':
			break loop
		case '-':
			p.setErr(fmt.Errorf("unexpected %q", ch))
			return rs
		case '\\':
			ch = p.nextEscaped()
			if p.err != nil {
				return rs
			}
		default:
			if !unicode.IsGraphic(ch) {
				p.setErr(fmt.Errorf("unexpected %q", ch))
				return rs
			}
		}

		if p.peek() == '-' {
			start := ch
			p.nextRune()

			ch := p.nextRune()
			switch ch {
			case 0:
				return rs
			case '[':
				p.setErr(fmt.Errorf("unexpected %q", ch))
				return rs
			case ']':
				rs.Add1('-')
				break loop
			case '-':
				p.setErr(fmt.Errorf("unexpected %q", ch))
				return rs
			case '\\':
				ch = p.nextEscaped()
				if p.err != nil {
					return rs
				}
			default:
				if !unicode.IsGraphic(ch) {
					p.setErr(fmt.Errorf("unexpected %q", ch))
					return rs
				}
			}
			rs.Add(start, ch)
		} else {
			rs.Add1(ch)
		}
	}

	if invert {
		rs = rs.Invert()
	}
	return rs
}

func (p *Lexer) next() Token {
	lno := p.lno
	lpos := p.lpos

	ch := p.nextRune()
	switch ch {
	case 0:
		return Token{typ: TokErr, lno: lno, lpos: lpos, err: p.err}

	case '\\':
		ch := p.nextEscaped()
		if p.err != nil {
			return Token{typ: TokErr, lno: lno, lpos: lpos, err: p.err}
		}
		return Token{typ: TokChar, lno: lno, lpos: lpos, err: p.err, lit: string(ch)}

	case '|':
		return Token{typ: TokAlt, lno: lno, lpos: lpos, err: p.err, lit: string(ch)}
	case '(':
		return Token{typ: TokOparen, lno: lno, lpos: lpos, err: p.err, lit: string(ch)}
	case ')':
		return Token{typ: TokCparen, lno: lno, lpos: lpos, err: p.err, lit: string(ch)}
	case '*':
		return Token{typ: TokStar, lno: lno, lpos: lpos, err: p.err, lit: string(ch)}
	case '+':
		return Token{typ: TokPlus, lno: lno, lpos: lpos, err: p.err, lit: string(ch)}
	case '.':
		return Token{typ: TokDot, lno: lno, lpos: lpos, err: p.err, lit: string(ch)}
	case '[':
		class := p.nextClass()
		if p.err != nil {
			return Token{typ: TokErr, lno: lno, lpos: lpos, err: p.err}
		}
		return Token{typ: TokClass, lno: lno, lpos: lpos, err: p.err, class: class}

	case ']':
		p.setErr(fmt.Errorf("unexpected %q", ch))
		return Token{typ: TokErr, lno: lno, lpos: lpos, err: p.err}
		
	default:
		if unicode.IsGraphic(ch) {
			return Token{typ: TokChar, lno: lno, lpos: lpos, err: p.err, lit: string(ch)}
		}
		p.setErr(fmt.Errorf("unexpected %q", ch))
		return Token{typ: TokErr, lno: lno, lpos: lpos, err: p.err}
	}
}

type Nfa struct {
}

func Parse(re string) (*Nfa, error) {
	lex := newLexer(re)
	for lex.err == nil {
		tok := lex.next()
		switch tok.typ {
		default:
			fmt.Printf("got token %v\n", tok)
		}
	}
	return nil, fmt.Errorf("todo success")
}
