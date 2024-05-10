package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func toString(v interface{}) string {
	s, _ := json.Marshal(v)
	return string(s)
}

func TestParseKeyword(t *testing.T) {
	tcs := []string{
		"true",
		"false",
		"null",
		`"a string"`,
		`"a 中文 string"`,
		`-1.23`,
		`[true,false,true]`,
		`{"age":23,"hobbies":["football","basketball"],"name":"lihua"}`,
	}
	for _, raw := range tcs {
		res, err := Parse(raw)
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		fmt.Println("result is: ", res)
		stringRes := toString(res)
		if stringRes != raw {
			t.Errorf("%s failed, res string is %s", raw, stringRes)
		}
	}
}
