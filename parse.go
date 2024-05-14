package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Parse(raw string) (*value, error) {
	p := newParser(raw)
	return p.parseValue()
}

func Get(raw string, path string) (any, error) {
	p := newParser(raw)
	val, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	keys := strings.Split(path, ".")
	for _, k := range keys {
		val, err = val.Access(k)
		if err != nil {
			return nil, err
		}
	}
	return val.Value(), nil
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

type value struct {
	t   _type
	str string
	num float64 // todo: a json nunmber type would be better
	arr []*value
	obj map[string]*value
}

type _type int32

const (
	_ _type = iota
	_true
	_false
	_null
	_string
	_number
	_array
	_object
)

func (v *value) Generate() string {
	if v.t == _true {
		return "true"
	}
	if v.t == _false {
		return "false"
	}
	if v.t == _null {
		return "null"
	}
	if v.t == _string {
		return fmt.Sprintf("\"%s\"", v.str)
	}
	if v.t == _number {
		if float64(int64(v.num)) == v.num {
			return fmt.Sprintf("%d", int64(v.num))
		}
		return fmt.Sprintf("%.2f", v.num)
	}
	if v.t == _array {
		strs := make([]string, 0, len(v.arr))
		for _, elem := range v.arr {
			str := elem.Generate()
			strs = append(strs, str)
		}
		return fmt.Sprintf("[%s]", strings.Join(strs, ","))
	}
	if v.t == _object {
		strs := make([]string, 0, len(v.obj))
		for name, val := range v.obj {
			str := fmt.Sprintf("\"%s\":%s", name, val.Generate())
			strs = append(strs, str)
		}
		sort.Strings(strs)
		return fmt.Sprintf("{%s}", strings.Join(strs, ","))
	}
	return "the answer is 42"
}

func (v *value) Value() any {
	if v.t == _true {
		return true
	}
	if v.t == _false {
		return false
	}
	if v.t == _null {
		return nil
	}
	if v.t == _string {
		return v.str
	}
	if v.t == _number {
		return v.num
	}
	if v.t == _array {
		ret := make([]any, 0, len(v.arr))
		for _, e := range v.arr {
			ret = append(ret, e.Value())
		}
		return ret
	}
	if v.t == _object {
		ret := make(map[string]any)
		for k, v := range v.obj {
			ret[k] = v.Value()
		}
		return ret
	}
	return nil
}

func (v *value) Access(key string) (*value, error) {
	if strings.HasPrefix(key, "#") {
		idx, err := strconv.Atoi(key[1:])
		if err != nil {
			return nil, err
		}
		if v.t != _array {
			return nil, errors.New("not an array")
		}
		if idx >= len(v.arr) {
			return nil, errors.New("index out of bound")
		}
		return v.arr[idx], nil
	}

	if v.t != _object {
		return nil, errors.New("not an object")
	}

	val, ok := v.obj[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}
	return val, nil
}
