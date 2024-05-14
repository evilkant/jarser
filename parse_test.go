package main

import (
	"encoding/json"
	"testing"
)

func toString(v interface{}) string {
	s, _ := json.Marshal(v)
	return string(s)
}

func TestParse(t *testing.T) {
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
		stringRes := res.Generate()
		if stringRes != raw {
			t.Errorf("%s failed, res string is %s", raw, stringRes)
		}
	}
}

func TestGet(t *testing.T) {
	raw := `{"info":{"age":23,"hobbies":["football","basketball"],"name":"lihua"}}`
	path := "info.hobbies.#1"
	val, err := Get(raw, path)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	hobby, ok := val.(string)
	if !ok {
		t.Errorf("%v not a string", val)
	}
	if hobby != "basketball" {
		t.Errorf("%s not basketball", hobby)
	}
	t.Logf("hobby is %s", hobby)
}
