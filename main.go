package main

import "fmt"

func main() {
	raw := `{"info":{"age":23,"hobbies":["football","basketball"],"name":"lihua"}}`
	value, _ := Parse(raw)
	fmt.Printf("%s\n", value.Generate())
	secondHobby, _ := Get(raw, "info.hobbies.#1")
	fmt.Printf("%s\n", secondHobby)
}
