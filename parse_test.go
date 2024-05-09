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
		`"a simple string"`,
		`[true,false,true]`,
		`{"hobbies":["football","basketball"],"name":"lihua"}`,
	}
	for _, raw := range tcs {
		res, err := Parse(raw)
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		fmt.Println("result is: ", res)
		resString := toString(res)
		if resString != raw {
			t.Errorf("%s failed, resString is %s", raw, resString)
		}
	}
}
