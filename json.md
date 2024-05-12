Reference: https://www.json.org/json-en.html

JSON (JavaScript Object Notation) is a data-exchange text format.

Each JSON text represents a value. A value can be a string, or a number, or true or false or null, or an object or an array.
It is easy to know which type of value you are parsing ---- (skipping spaces) just look at the first symbol.
```
n --> null
t --> true
f --> false
" --> string
0-9/- --> number
[ --> array
{ --> object
```

By the way, parsing, is just structuring a linear representation (symbols come one after another in a line) in accordance to some rule (sometimes called grammar).

The goal of this repo is just to be able to (1) parse a string of JSON format into a tree-like data structure, and (2) convert the constructed data structure back to the original JSON-formatted string.