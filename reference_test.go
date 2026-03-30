package oremus

import (
	"testing"
)

var refs = map[string]string{
	"gen":                    "Genesis",                // bare book
	"gen 1":                  "Genesis 1",              // bare chapter
	"   gen 1   ":            "Genesis 1",              // spaces
	"gen 1 - 4":              "Genesis 1-4",            // multiple chapters
	"gen 1,2":                "Genesis 1,2",            // comma separated chapters
	"gen 1:1 - 10":           "Genesis 1:1-10",         // one chapter, spaces
	"gen 1:1 — 10":           "Genesis 1:1-10",         // em dash
	"gen 1:4,6, 7,8 , 9":     "Genesis 1:4,6,7,8,9",    // one chapter, different verses, spaces
	"gen 1:4,6-10,11-14":     "Genesis 1:4,6-10,11-14", // one chapter, different verses, with ranges
	"gen 10:1-7,,9-99":       "Genesis 10:1-7,9-99",    // random comma -- parser fixes
	"gen 10:1-11:3":          "Genesis 10:1-11:3",      // cross chapter boundary
	"gen 1:1-2:4":            "Genesis 1:1-2:4",        // cross chapter boundary
	"gen 1:15ff":             "Genesis 1:15ff",         // ff suffix
	"gen 10:1-7,12:9-99":     "Genesis 10:1-7,12:9-99", // comma starts new chapter
	"gen 1:1a-2:4":           "Genesis 1:1a-2:4",       // suffix on start
	"gen 1:1-2:4a":           "Genesis 1:1-2:4a",       // suffix on end
	"1 john 4:8":             "1 John 4:8",             // prefix
	"1st john 4:8-9":         "1 John 4:8-9",           // prefix with verse range
	"gen 1; ex 2":            "Genesis 1; Exodus 2",
	"gen 1:1-3; john 3:16":   "Genesis 1:1-3; John 3:16",
	"Acts of the apostles 1": "Acts 1",
	"song of songs 1:1":      "Song of Songs 1:1",
	"mark 1:1fff":            "Mark 1:1ff",
	"3 john":                 "3 John",
	"3\tjohn":                "3 John",
	"gen 1-2:3":              "Genesis 1-2:3",
	"gen 1:2-3:4":            "Genesis 1:2-3:4",
	"gen 1;ex 2":             "Genesis 1; Exodus 2",
	"gen 1;;;ex 2":           "Genesis 1; Exodus 2",

	// multi-word books
	"acts of the apostles 1": "Acts 1",
	"song of songs 9:1":      "Song of Songs 9:1",

	// prefixes
	"1 john 1":    "1 John 1",
	"2 peter 1:1": "2 Peter 1:1",
	"1st john 1":  "1 John 1",

	// chapters
	"gen 1-3": "Genesis 1-3",

	// verses
	"gen 1:1-3":   "Genesis 1:1-3",
	"gen 1:1,2,3": "Genesis 1:1,2,3",

	// cross chapter
	"gen 1:1-2:3": "Genesis 1:1-2:3",

	// suffixes
	"gen 1:1a":   "Genesis 1:1a",
	"gen 1:1ff":  "Genesis 1:1ff",
	"gen 1:1fff": "Genesis 1:1ff",

	// commas
	"gen 1,,2":   "Genesis 1,2",
	"gen 1:1,,3": "Genesis 1:1,3",

	// semicolons
	"gen 1;;ex 2": "Genesis 1; Exodus 2",

	// whitespace
	" gen   1:1 ": "Genesis 1:1",

	// dashes
	"gen 1:1—3": "Genesis 1:1-3",
}

// these are things that should not work, just checking the error messages (use `go test -v`)
var badrefs = []string{
	"gen a",
	"gen a::",
	", , ,",
	"gen 40:1-1:1",
	"",
	";_;",
	"*James 1",
	"2nd Luke 9",
	"gen a:b-c:d",
	// "gen 1:0", // Genesis 1
	"gen 1:5-1:1",
	"gen 2-1",
	// "gen 1:", // Genesis 1
	"gen :1",
	"4 john",
	"unknown",
	// "gen 0", // Yeilds "Genesis"
	"gen 1:0",
	"gen 1:",
	"gen :1",
	"gen 1-",
	"gen 1:1-",
	"gen 1:1,",
	"gen 2-1",
	"gen 1:3-1",
	"4 john 1",
	"2 genesis 1",
	"gen 1:1c",
	"gen 1:1$",
}

func TestParseReference(t *testing.T) {
	for k, v := range refs {
		t.Logf("testing %s", k)
		s, err := CleanReference(k)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if s != v {
			t.Errorf("[%s] did not round-trip [%s] [%s]", k, v, s)
		} else {
			t.Logf("[%s] => [%s]", k, s)
		}
	}

	for _, v := range badrefs {
		s, err := CleanReference(v)
		if err == nil {
			t.Errorf("expected error for input %q, got none", v)
		}
		if s != "" {
			t.Logf("[%s] => [%s]", v, s)
		}
	}
}

func TestIdempotent(t *testing.T) {
	input := "gen 1:1 - 10"

	r1, err := ParseReference(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r2, err := ParseReference(r1.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r1.String() != r2.String() {
		t.Errorf("not idempotent: %s != %s", r1.String(), r2.String())
	}
}

func TestInternals(t *testing.T) {
	parsed, _ := ParseReference("gen 1:1-3")

	if parsed.Book != "Genesis" {
		t.Errorf("wrong book")
	}
	if len(parsed.ChapterVerseRange) != 1 {
		t.Errorf("expected 1 range")
	}
}

/*
func TestTables(t *testing.T) {
	for _, tc := range refs {
		t.Run(tc.in, func(t *testing.T) {
			s, err := CleanReference(tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s != tc.out {
				t.Errorf("got %q, want %q", s, tc.out)
			}
		})
	}
} */

func FuzzParseReference(f *testing.F) {
	f.Add("gen 1:1")
	f.Add("john 3:16-18")
	f.Add("gen 1:15ff")
	f.Add("song of songs 1:1")
	f.Add("1 john 4:8")
	f.Add("gen 1-2:3")

	f.Fuzz(func(t *testing.T, input string) {
		r, err := ParseReference(input)
		if err != nil {
			return
		}

		// round-trip must not break
		s := r.String()
		r2, err := ParseReference(s)
		if err != nil {
			t.Fatalf("round-trip parse failed: %v", err)
		}

		if s != r2.String() {
			t.Fatalf("not idempotent: %q != %q", s, r2.String())
		}
	})
}
