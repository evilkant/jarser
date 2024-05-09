package main

import (
	"errors"
	"fmt"
	"strings"
)

func Parse(raw string) (any, error) {
	raw = squashWhiteSpace(raw)
	p := newParser(raw)
	return p.parseValue()
}

type parser struct {
	s   string
	i   int // the index of the next character to be processed
	err error
}

func (p *parser) parseValue() (any, error) {
	c := p.nextCh()

	if c == 't' {
		return p.parseTrue()
	}

	if c == 'f' {
		return p.parseFalse()
	}

	if c == 'n' {
		return p.parseNull()
	}

	if c == '"' {
		return p.parseString()
	}

	if c == '{' {
		return p.parseObject()
	}

	if c == '[' {
		return p.parseArray()
	}

	if c == '9' {
		return p.parseNumber()
	}

	return nil, fmt.Errorf("%v not recognized as the start of a json value", c)
}

func (p *parser) matchKeyword(keyword string) error {
	l, r := p.i, p.i+len(keyword)
	if r > len(p.s) {
		r = len(p.s)
	}
	s := p.s[l:r]
	if s == keyword {
		p.advance(len(keyword))
		return nil
	}
	return fmt.Errorf("expect '%s', but got '%s'..", keyword, s)
}

func (p *parser) advance(i int) {
	p.i += i
}

func (p *parser) nextCh() rune {
	return rune(p.s[p.i])
}

func (p *parser) parseTrue() (bool, error) {
	if err := p.matchKeyword("true"); err != nil {
		return false, err
	}
	return true, nil
}

func (p *parser) parseFalse() (bool, error) {
	if err := p.matchKeyword("false"); err != nil {
		return false, err
	}
	return false, nil
}

func (p *parser) parseNull() (any, error) {
	if err := p.matchKeyword("null"); err != nil {
		return nil, err
	}
	return nil, nil
}

func newParser(raw string) *parser {
	return &parser{s: raw, i: 0}
}

func (p *parser) parseObject() (any, error) {
	object := make(map[string]any)

	p.advance(1)
	for p.i < len(p.s) && p.nextCh() != '}' {
		p.skipWhiteSpaces()
		key, err := p.parseString()
		if err != nil {
			return nil, err
		}
		p.skipWhiteSpaces()
		if err := p.matchKeyword(":"); err != nil {
			return nil, err
		}
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		object[key] = value
		p.skipWhiteSpaces()
		if p.nextCh() == '}' {
			break
		}
		if err := p.matchKeyword(","); err != nil {
			return nil, err
		}
	}

	if p.i == len(p.s) {
		return nil, errors.New("object not closed")
	}

	return object, nil
}

func (p *parser) parseString() (string, error) {
	str := ""
	p.advance(1)
	for p.i < len(p.s) && p.nextCh() != '"' {
		c := p.nextCh()
		if c == '\\' {
			if p.i+1 >= len(p.s) {
				return "", errors.New("")
			}
			nc := p.s[p.i+1]
			switch nc {
			case 't':
				str += "\t"
			case 'n':
				str += "\n"
			case '"':
				str += "\""
			default:
				str += string(nc)
			}
		} else {
			str = str + string(c)
		}
		p.advance(1)
	}
	if p.i == len(p.s) {
		return "", errors.New("string not closed")
	}
	p.advance(1)
	return str, nil
}

func (p *parser) parseNumber() (any, error) {
	return "", nil
}

func (p *parser) parseArray() ([]any, error) {
	p.advance(1)

	var res []any
	for p.i < len(p.s) && p.nextCh() != ']' {
		p.skipWhiteSpaces()
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		res = append(res, val)

		p.skipWhiteSpaces()
		if p.nextCh() == ']' {
			break
		}
		if err = p.matchKeyword(","); err != nil {
			return nil, fmt.Errorf("expect comma")
		}
	}
	if p.i == len(p.s) {
		return nil, errors.New("array not closed")
	}
	p.advance(1)
	return res, nil
}

func (p *parser) skipWhiteSpaces() {
	for p.i < len(p.s) && p.nextCh() == ' ' {
		p.advance(1)
	}
}

func squashWhiteSpace(raw string) string {
	return strings.Join(strings.Fields(raw), " ")
}
