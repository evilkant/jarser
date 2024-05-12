package main

import (
	"fmt"
	"sort"
	"strings"
)

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
