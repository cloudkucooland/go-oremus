package oremus

import (
	// "bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Reference is the datastructure used for parsing/validating references
type Reference struct {
	prefix    rune
	Book      string
	ChapVerse []chapVerse
}

// chapVerse is the chapter/verse portion of a reference "1:14-2:7a"
type chapVerse struct {
	StartChapter     uint8
	StartVerse       uint8
	StartVerseSuffix string // a, b, ff
	EndChapter       uint8
	EndVerse         uint8
	EndVerseSuffix   string // a, b
}

// some books have a prefix, this is used to normalize those to "1, 2, 3"
var prefixes = map[rune][]string{
	'1': {"1", "1st", "i", "l", "first"},
	'2': {"2", "2nd", "ii", "ll", "second"},
	'3': {"3", "3rd", "iii", "lll", "third"},
}

// a list of known variations of book names (lowercase for ease of matching)
var books = map[string][]string{
	"Genesis":   {"genesis", "gen"},
	"Exodus":    {"exodus", "ex"},
	"Leviticus": {"leviticus", "lev"},
}

// books which have prefixes (John makes this complicated)
var bookswithprefix = []string{
	"kings", "chronicles", "john", "corinthians",
}

// String returns a normalized reference to a scripture passage
func (r *Reference) String() string {
	return "nothing yet"
}

// CleanReference takes a string, returns a normalized string
func CleanReference(in string) (string, error) {
	r, err := ParseReference(in)
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// ParseReference parses a free-form reference to a scripture passage and returns a *Reference
// use the stringer method to get a normalized string format back
func ParseReference(in string) (*Reference, error) {
	newRef := Reference{}
	hasPrefix := false

	chunks := strings.Fields(in)
	for pos, c := range chunks {
		if pos == 0 {
			if b, id := isPrefix(c); b {
				newRef.prefix = id
				hasPrefix = true
				continue
			}
			if b, book := isBook(c); b {
				newRef.Book = book
				continue
			}
			return &newRef, fmt.Errorf("invalid book: %s", c)
		}
		if hasPrefix && pos == 1 {
			if b, book := isBook(c); b {
				if !allowedPrefix(book) {
					return &newRef, fmt.Errorf("invalid book: %s (cannot have prefix)", c)
				}
				newRef.Book = book
				continue
			}
			return &newRef, fmt.Errorf("invalid book: %s (with prefix)", c)
		}
		if (!hasPrefix && pos == 1) || (pos > 1) {
			cv := chapVerse{}
			err := parseChapterVerse(c, &cv)
			if err != nil {
				return &newRef, err
			}
			newRef.ChapVerse = append(newRef.ChapVerse, cv)
		}
	}
	log.Printf("Final: %+v\n", newRef)
	return &newRef, fmt.Errorf("just testing")
}

func isPrefix(in string) (bool, rune) {
	lc := strings.ToLower(in)

	for k, v := range prefixes {
		for _, option := range v {
			if lc == option {
				return true, k
			}
		}
	}
	return false, ' '
}

func isBook(in string) (bool, string) {
	lc := strings.ToLower(in)
	lc = strings.TrimSuffix(lc, ".")

	for k, v := range books {
		for _, option := range v {
			if lc == option {
				return true, k
			}
		}
	}
	return false, ""
}

func allowedPrefix(book string) bool {
	lc := strings.ToLower(book)
	for _, v := range bookswithprefix {
		if lc == v {
			return true
		}
	}
	return false
}

func parseChapterVerse(in string, r *chapVerse) error {
	log.Println(in)

	colons := strings.Split(in, ":")
	log.Printf("%+v\n", colons)
	switch len(colons) {
	case 1:
		// just a chapter
		chapAsInt, err := strconv.Atoi(colons[0])
		if err != nil {
			return err
		}
		r.StartChapter = uint8(chapAsInt)
		return nil
	case 2:
		// chapter and verses
		chapAsInt, err := strconv.Atoi(colons[0])
		if err != nil {
			return err
		}
		r.StartChapter = uint8(chapAsInt)

		// XXX process verse reference
		// split at -
		return nil
	case 3:
		// dafuq?
	}

	return nil
}
