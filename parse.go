package main

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

func Parse(raw string) (*value, error) {
	p := newParser(raw)
	return p.parseValue()
}

type parser struct {
	s string
	i int // the index of the next rune to be processed
}

func (p *parser) advance() {
	p.advanceN(1)
}

func (p *parser) advanceN(n int) {
	for ; n > 0; n-- {
		_, w := utf8.DecodeRuneInString(p.s[p.i:])
		p.i += w
	}
}

const emptyRune = rune(0)

func (p *parser) getRune() rune {
	return p.peekRune(0)
}

func (p *parser) peekRune(i int) rune {
	var r rune
	var w int
	offset := p.i
	for ; i >= 0; i-- {
		if offset >= len(p.s) {
			return emptyRune
		}
		r, w = utf8.DecodeRuneInString(p.s[offset:])
		offset += w
	}
	return r
}

func (p *parser) advanceOneThenGetRune() rune {
	p.advance()
	return p.getRune()
}

func (p *parser) parseValue() (*value, error) {
	r := p.getRune()
	if r == 't' {
		return p.parseTrue()
	}

	if r == 'f' {
		return p.parseFalse()
	}

	if r == 'n' {
		return p.parseNull()
	}

	if r == '"' {
		return p.parseString()
	}

	if unicode.IsDigit(r) || r == '-' {
		return p.parseNumber()
	}

	if r == '{' {
		return p.parseObject()
	}

	if r == '[' {
		return p.parseArray()
	}

	return nil, fmt.Errorf("%v not recognized as the start of a json value", string(r))
}

func (p *parser) matchLiteral(literal string) error {
	l, r := p.i, p.i+len(literal)
	if r > len(p.s) {
		r = len(p.s)
	}
	s := p.s[l:r]
	if s == literal {
		// advance in bytes, not in runes
		p.i += len(literal)
		return nil
	}
	return fmt.Errorf("expect '%s', but got '%s'..", literal, s)
}

func (p *parser) parseTrue() (*value, error) {
	if err := p.matchLiteral("true"); err != nil {
		return nil, err
	}
	return &value{t: _true}, nil
}

func (p *parser) parseFalse() (*value, error) {
	if err := p.matchLiteral("false"); err != nil {
		return nil, err
	}
	return &value{t: _false}, nil
}

func (p *parser) parseNull() (*value, error) {
	if err := p.matchLiteral("null"); err != nil {
		return nil, err
	}
	return &value{t: _null}, nil
}

func (p *parser) parseString() (*value, error) {
	if p.getRune() != '"' {
		return nil, errors.New("expect '\"'")
	}
	p.advance()

	str := ""
	for r := p.getRune(); r != emptyRune && r != '"'; r = p.advanceOneThenGetRune() {
		if r != '\\' {
			str = str + string(r)
			continue
		}
		nr := p.peekRune(1)
		if nr == emptyRune {
			return nil, errors.New("expect rune after '\\'")
		}
		switch nr {
		case 't':
			str += "\t"
		case 'n':
			str += "\n"
		case '"':
			str += "\""
		default:
			str += string(nr)
		}
		p.advance()
	}
	if p.getRune() == emptyRune {
		return nil, errors.New("string not closed")
	}
	p.advance()
	return &value{t: _string, str: str}, nil
}

func (p *parser) parseNumber() (*value, error) {
	if r := p.getRune(); r != '-' && !unicode.IsDigit(r) {
		return nil, errors.New("expect minus sign or digit")
	}

	numStr := ""

	var r rune
	if r = p.getRune(); r == '-' {
		numStr += "-"
		p.advance()
	}
	for r = p.getRune(); r != emptyRune && unicode.IsDigit(r); r = p.advanceOneThenGetRune() {
		numStr += string(r)
	}
	if r == '.' {
		numStr += "."
		p.advance()
		if !unicode.IsDigit(p.getRune()) {
			return nil, errors.New("expect digit after '.'")
		}
	}
	for r = p.getRune(); r != emptyRune && unicode.IsDigit(r); r = p.advanceOneThenGetRune() {
		numStr += string(r)
	}

	f, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s as float", numStr)
	}

	return &value{t: _number, num: f}, nil
}

func newParser(raw string) *parser {
	return &parser{s: raw, i: 0}
}

func (p *parser) parseArray() (*value, error) {
	if p.getRune() != '[' {
		return nil, errors.New("expect '['")
	}
	p.advance()

	var res []*value
	for r := p.getRune(); r != emptyRune && r != ']'; r = p.getRune() {
		p.skipWhiteSpaces()
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		res = append(res, val)

		p.skipWhiteSpaces()
		if p.getRune() == ']' {
			break
		}
		if err = p.matchLiteral(","); err != nil {
			return nil, err
		}

	}
	if p.getRune() == emptyRune {
		return nil, errors.New("array not closed")
	}
	p.advance()

	return &value{t: _array, arr: res}, nil
}

func (p *parser) parseObject() (*value, error) {
	if p.getRune() != '{' {
		return nil, errors.New("expect '{'")
	}
	p.advance()

	object := make(map[string]*value)

	for r := p.getRune(); r != emptyRune && r != '}'; r = p.getRune() {
		p.skipWhiteSpaces()
		name, err := p.parseString()
		if err != nil {
			return nil, err
		}
		p.skipWhiteSpaces()
		if err := p.matchLiteral(":"); err != nil {
			return nil, err
		}
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		object[name.str] = value
		p.skipWhiteSpaces()
		if p.getRune() == '}' {
			break
		}
		if err := p.matchLiteral(","); err != nil {
			return nil, err
		}
	}

	if p.getRune() == emptyRune {
		return nil, errors.New("object not closed")
	}
	p.advance()
	return &value{t: _object, obj: object}, nil
}

func (p *parser) skipWhiteSpaces() {
	for unicode.IsSpace(p.getRune()) {
		p.advance()
	}
}
