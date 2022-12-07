package oremus

import (
	"testing"
	// "fmt"
)

var refs = map[string]string{
	"gen":          "Genesis",
	"gen 1":        "Genesis 1",
	"   gen 1   ":  "Genesis 1",
	"gen 1-2":      "Genesis 1-2",
	"gen 1 - 2":    "Genesis 1-2",
	"gen 1:1 - 10": "Genesis 1:1-10",
	"gen 1:1 — 10": "Genesis 1:1-10",
	"gen 1:1a-2:4": "Genesis 1:1a-4:2",
	"gen 1:1-2:4a": "Genesis 1:1-4:2a",
	"1 john 4:8":   "1 John 4:8",
	"1st john 4:8": "1 John 4:8",
}

func TestHelloName(t *testing.T) {
	for k, v := range refs {
		// fmt.Println(k)
		parsed, err := parseReference(k)
		if err != nil {
			t.Fatalf("%v", err)
		}
		s := parsed.String()
		if s != v {
			t.Fatalf("[%s] did not round-trip [%s] [%s]\n%+v", k, v, s, parsed)
		}
	}
}
