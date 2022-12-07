package oremus

import (
	"bytes"
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
	StartVerseSuffix rune // a, b, f
	EndChapter       uint8
	EndVerse         uint8
	EndVerseSuffix   rune // a, b, f
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
	var unset rune
	first := true
	var cacheChapter uint8

	buf := bytes.Buffer{}
	if r.prefix != unset {
		buf.WriteRune(r.prefix)
		buf.WriteRune(' ')
	}
	buf.WriteString(r.Book)
	buf.WriteRune(' ')

	for _, v := range r.ChapVerse {
		if !first {
			buf.WriteString(",")
		} else {
			first = false
		}

		if v.StartChapter != cacheChapter {
			chap := fmt.Sprintf("%d", v.StartChapter)
			buf.WriteString(chap)
		}
		if v.StartVerse != 0 {
			if v.StartChapter != cacheChapter {
				buf.WriteRune(':')
			}
			verse := fmt.Sprintf("%d", v.StartVerse)
			buf.WriteString(verse)
			if v.StartVerseSuffix != unset {
				if v.StartVerseSuffix != 'f' {
					buf.WriteRune(v.StartVerseSuffix)
				} else {
					buf.WriteString("ff")
				}
			}
		}
		if v.EndChapter != v.StartChapter || v.EndVerse != v.StartVerse {
			buf.WriteRune('-')
		}
		if v.EndChapter != v.StartChapter {
			chap := fmt.Sprintf("%d", v.EndChapter)
			buf.WriteString(chap)
			if v.EndVerse != 0 {
				buf.WriteRune(':')
			}
		}
		if v.EndChapter != v.StartChapter || v.EndVerse != v.StartVerse {
			verse := fmt.Sprintf("%d", v.EndVerse)
			buf.WriteString(verse)
			if v.EndVerseSuffix != unset {
				if v.EndVerseSuffix != 'f' {
					buf.WriteRune(v.EndVerseSuffix)
				} else {
					buf.WriteString("ff")
				}
			}
		}
		cacheChapter = v.EndChapter
	}

	return buf.String()
}

// CleanReference takes a string, returns a normalized string for a semi-colon separated list of scripture references
func CleanReference(in string) (string, error) {
	r, err := parseReferences(in)
	if err != nil {
		return "", err
	}

	buf := bytes.Buffer{}
	first := true
	for _, v := range r {
		if !first {
			buf.WriteString("; ")
		} else {
			first = false
		}
		buf.WriteString(v.String())
	}
	log.Printf("\nin: %s\nout: %s\n", in, buf.String())
	return buf.String(), fmt.Errorf("still testing")
}

// ParseReference parses a free-form reference to a scripture passage and returns a []*Reference
// use the stringer method to get a normalized string format back
func parseReferences(in string) ([]*Reference, error) {
	var out []*Reference

	refs := strings.Split(in, ";")
	for _, r := range refs {
		parsed, err := parseReference(r)
		if err != nil {
			return out, err
		}
		out = append(out, parsed)
	}
	return out, nil
}

// parse reference parses a single free-form reference to a scripture passage and returns a *Reference
// use the stringer method to get a normalized string format back
func parseReference(in string) (*Reference, error) {
	newRef := Reference{}
	hasPrefix := false
	var err error
	buf := bytes.Buffer{}

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

		// concat the rest and process letter-by-letter
		if (!hasPrefix && pos == 1) || (pos > 1) {
			buf.WriteString(c)
		}
	}

	newRef.ChapVerse, err = parseChapterVerse(buf.String())
	if err != nil {
		return &newRef, err
	}

	// log.Printf("Parsed %s => %+v\n", in, newRef)
	return &newRef, nil
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

func parseChapterVerse(in string) ([]chapVerse, error) {
	out := make([]chapVerse, 0)
	cv := chapVerse{}
	workbuf := bytes.Buffer{}
	inChapter := true // in chapter or verse
	startRef := true  // in the first part of the reference
	runes := []rune(in)

	// save some redundancy
	var wb = func() error {
		if inChapter {
			if startRef {
				sci, err := strconv.Atoi(workbuf.String())
				if err != nil {
					return err
				}
				cv.StartChapter = uint8(sci)
				cv.EndChapter = uint8(sci)
			} else {
				eci, err := strconv.Atoi(workbuf.String())
				if err != nil {
					return err
				}
				cv.EndChapter = uint8(eci)
			}
		} else {
			if startRef {
				svi, err := strconv.Atoi(workbuf.String())
				if err != nil {
					return err
				}
				cv.StartVerse = uint8(svi)
				cv.EndVerse = uint8(svi)
			} else {
				evi, err := strconv.Atoi(workbuf.String())
				if err != nil {
					return err
				}
				cv.EndVerse = uint8(evi)
			}
		}
		workbuf.Truncate(0)
		return nil
	}

	for _, r := range runes {
		switch r {
		case ',':
			// , ends the current reference and starts a new one
			if err := wb(); err != nil {
				return out, err
			}
			out = append(out, cv)
			cchap := cv.StartChapter // for Gen 2:2,4,9 style references, this is ambiguous
			cv = chapVerse{}
			if !inChapter {
				cv.StartChapter = cchap
				cv.EndChapter = cchap
			}
			startRef = true
			// no inChapter = true due to ambiguitiy above?
		case '-', '—', '–':
			// - moves from the first part of a reference to the end of one (either chapter or verse)
			if err := wb(); err != nil {
				return out, err
			}
			startRef = false
		case ':':
			// : moves from chapter to verse
			if err := wb(); err != nil {
				return out, err
			}
			inChapter = false
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			workbuf.WriteRune(r)
		case 'a', 'b', 'f', 'A', 'B', 'F': // use f as shorthand for ff
			if err := wb(); err != nil {
				return out, err
			}
			// do we need to handle Genesis 1ff or Genesis 1-3ff? or is that noise
			if startRef && inChapter {
				cv.StartVerseSuffix = r
			} else {
				cv.EndVerseSuffix = r
			}
		default:
			log.Println("ignoring noise %r", r)
		}
	}
	if err := wb(); err != nil {
		return out, err
	}

	// validate that start verse is < end verse, & chapters

	out = append(out, cv)
	return out, nil
}
