package oremus

import (
	"encoding/json"
	"testing"
)

var refs = map[string]string{
	"gen":                "Genesis",                // bare book
	"gen 1":              "Genesis 1",              // bare chapter
	"   gen 1   ":        "Genesis 1",              // spaces
	"gen 1 - 4":          "Genesis 1-4",            // multiple chapters
	"gen 1,2":            "Genesis 1,2",            // comma separated chapters
	"gen 1:1 - 10":       "Genesis 1:1-10",         // one chapter, spaces
	"gen 1:1 — 10":       "Genesis 1:1-10",         // em dash
	"gen 1:4,6, 7,8 , 9": "Genesis 1:4,6,7,8,9",    // one chapter, different verses, spaces
	"gen 1:4,6-10,11-14": "Genesis 1:4,6-10,11-14", // one chapter, different verses, with ranges
	"gen 10:1-7,,9-99":   "Genesis 10:1-7,,9-99",   // random comma -- needs to be fixed
	"gen 10:1-11:3":      "Genesis 10:1-11:3",      // cross chapter boundary
	"gen 1:1-2:4":        "Genesis 1:1-2:4",        // cross chapter boundary
	"gen 1:15ff":         "Genesis 1:15ff",         // ff suffix
	"gen 10:1-7,12:9-99": "Genesis 10:1-7,12:9-99", // comma starts new chapter
	"gen 1:1a-2:4":       "Genesis 1:1a-2:4",       // suffix on start
	"gen 1:1-2:4a":       "Genesis 1:1-2:4a",       // suffix on end
	"1 john 4:8":         "1 John 4:8",             // prefix
	"1st john 4:8-9":     "1 John 4:8-9",           // prefix with verse range
}

// these are things that should not work, just checking the error messages (use `go test -v`)
var badrefs = []string{"gen a", "gen a::", ", , ,", "gen 40:1-1:1", "", ";_;", "*James 1", "2nd Luke 9", "gen a:b-c:d"}

func TestParseReference(t *testing.T) {
	for k, v := range refs {
		t.Logf("testing %s", k)
		parsed, err := ParseReference(k)
		if err != nil {
			t.Fatalf("%v", err)
		}

		s := parsed.String()
		if s != v {
			p, _ := json.Marshal(parsed)
			t.Errorf("[%s] did not round-trip [%s] [%s]\n[%+v]", k, v, s, string(p))
		}
	}

	for _, v := range badrefs {
		t.Logf("testing %s", v)
		parsed, err := ParseReference(v)
		if err != nil {
			t.Logf("%v", err)
			continue
		}

		t.Logf("[%s] => [%s]", v, parsed.String())
	}
}
