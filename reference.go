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
	Prefix    rune
	Book      string
	ChapVerse []ChapVerse
}

// ChapVerse is the chapter/verse portion of a reference "1:14-2:7a"
type ChapVerse struct {
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
	"Genesis":       {"genesis", "gen"},
	"Exodus":        {"exodus", "ex"},
	"Leviticus":     {"leviticus", "lev"},
	"Numbers":       {"numbers", "num"},
	"Deuteronomy":   {"deuteronomy", "deut"},
	"Joshua":        {"joshua", "josh"},
	"Judges":        {"judges", "judg", "jud"},
	"Ruth":          {"ruth"},
	"Samuel":        {"samuel", "sam"},
	"Kings":         {"kings"},
	"Chronicles":    {"chronicles", "chron"},
	"Ezra":          {"ezra"},
	"Nehemiah":      {"nehemiah", "neh"},
	"Esther":        {"esther", "esth"},
	"Job":           {"job"},
	"Psalm":         {"psalm"}, // sometimes it's OK to have the singular title
	"Psalms":        {"psalms", "ps"},
	"Proverbs":      {"proverbs", "prov"},
	"Ecclesiastes":  {"ecclesiastes", "eccl"},
	"Song of Songs": {"song of songs", "song"},
	"Isaiah":        {"isaiah", "is"},
	"Jeremiah":      {"jeremiah", "jer"},
	"Lamentations":  {"lamentations", "lam"},
	"Ezekiel":       {"ezekiel", "ezek"},
	"Daniel":        {"daniel", "dan"},
	"Hosea":         {"hosea", "hos"},
	"Joel":          {"joel"},
	"Amos":          {"amos"},
	"Obadiah":       {"obadiah", "obad"},
	"Jonah":         {"jonah"},
	"Micah":         {"micah"},
	"Nahum":         {"nahum", "nah"},
	"Habakkuk":      {"habakkuk", "hab"},
	"Zephaniah":     {"zephaniah", "zep"},
	"Haggai":        {"haggai", "hag"},
	"Zechariah":     {"zechariah", "zech"},
	"Malachi":       {"malachi", "mal"},
	"Matthew":       {"matthew", "matt"},
	"Mark":          {"mark"},
	"Luke":          {"luke"},
	"John":          {"john"},
	"Acts":          {"acts"},
	"Romans":        {"romans", "rom"},
	"Corinthians":   {"corinthians", "cor"},
	"Galatians":     {"galatians", "gal"},
	"Ephesians":     {"ephesians", "eph"},
	"Philippians":   {"philippians", "phil"},
	"Colossians":    {"colossians", "col"},
	"Thessalonians": {"thessalonians", "thess"},
	"Timothy":       {"timothy", "tim"},
	"Titus":         {"titus"},
	"Philemon":      {"phileomon", "philemon"},
	"Hebrews":       {"hebrews", "heb"},
	"James":         {"james"},
	"Peter":         {"peter"},
	"Jude":          {"jude"},
	"Revelation":    {"revelation", "rev"},
	"Wisdom":        {"wisdom", "wis"},
}

// books which have prefixes (John makes this complicated)
var bookswithprefix = []string{
	"samuel", "kings", "chronicles", "corinthians", "thessalonians", "timothy", "peter", "john",
}

// String returns a normalized reference to a scripture passage
func (r *Reference) String() string {
	var unset rune
	first := true
	var prevChap uint8

	buf := bytes.Buffer{}
	if r.Prefix != unset {
		buf.WriteRune(r.Prefix)
		buf.WriteRune(' ')
	}
	buf.WriteString(r.Book)
	if len(r.ChapVerse) > 0 && r.ChapVerse[0].StartChapter != 0 {
		buf.WriteRune(' ')
	}

	for _, v := range r.ChapVerse {
		if !first {
			buf.WriteString(",")
		} else {
			first = false
		}

		if v.StartChapter != prevChap {
			buf.WriteString(fmt.Sprintf("%d", v.StartChapter))
		}
		if v.StartVerse != 0 {
			if v.StartChapter != prevChap {
				buf.WriteRune(':')
			}
			buf.WriteString(fmt.Sprintf("%d", v.StartVerse))
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
			buf.WriteString(fmt.Sprintf("%d", v.EndChapter))
			if v.EndVerse != 0 {
				buf.WriteRune(':')
			}
		}
		if v.EndVerse != v.StartVerse {
			buf.WriteString(fmt.Sprintf("%d", v.EndVerse))
			if v.EndVerseSuffix != unset {
				if v.EndVerseSuffix != 'f' {
					buf.WriteRune(v.EndVerseSuffix)
				} else {
					buf.WriteString("ff")
				}
			}
		}
		prevChap = v.EndChapter
	}

	return buf.String()
}

// CleanReference takes a string, returns a normalized string for a semi-colon separated list of scripture references
func CleanReference(in string) (string, error) {
	r, err := ParseReferences(in)
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
	return buf.String(), nil
}

// ParseReference parses a free-form reference to a scripture passage and returns a []*Reference
// use the stringer method to get a normalized string format back
func ParseReferences(in string) ([]*Reference, error) {
	var out []*Reference

	refs := strings.Split(in, ";")
	for _, r := range refs {
		parsed, err := ParseReference(r)
		if err != nil {
			return out, err
		}
		out = append(out, parsed)
	}
	return out, nil
}

// ParseReference parses a single free-form reference to a scripture passage and returns a *Reference
// use the stringer method to get a normalized string format back
// XXX This doesn't work for "Acts of the Apostles"
func ParseReference(in string) (*Reference, error) {
	newRef := Reference{}
	hasPrefix := false
	var err error
	buf := bytes.Buffer{}

	chunks := strings.Fields(in)
	for pos, c := range chunks {
		if pos == 0 {
			if b, id := isPrefix(c); b {
				newRef.Prefix = id
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

func parseChapterVerse(in string) ([]ChapVerse, error) {
	runes := []rune(in)         // input string as runes
	out := make([]ChapVerse, 0) // slice for multiple references
	cv := ChapVerse{}           // the reference we are working on
	workbuf := bytes.Buffer{}   // a buffer is overkill since we need, at most, 3 bytes, but this is efficient enough. Psalm 119:129
	inChapter := true           // in chapter or verse?
	startRef := true            // in the first part of the reference?

	var unbuffer uint8

	// a closure to save some labor below
	var wb = func() {
		s := workbuf.String()
		if s == "" {
			return
		}
		si, err := strconv.Atoi(s)
		if err != nil {
			si = 0
		}
		unbuffer = uint8(si)
		if inChapter {
			if startRef {
				cv.StartChapter = uint8(si)
			}
			// for both
			cv.EndChapter = uint8(si)
		} else {
			if startRef {
				cv.StartVerse = uint8(si)
			}
			// for both
			cv.EndVerse = uint8(si)
		}
		workbuf.Reset()
	}

	for _, r := range runes {
		switch r {
		case ',':
			// , ends the current reference and starts a new one
			wb()
			out = append(out, cv)
			cchap := cv.StartChapter
			cv = ChapVerse{}
			if !inChapter {
				cv.StartChapter = cchap
				cv.EndChapter = cchap
			}
			startRef = true
			// no inChapter = true
		case '-', '—', '–': // hyphen, en dash, and em dashes all found in the wild
			// - moves from the first part of a reference to the end of one (either chapter or verse)
			wb()
			startRef = false
		case ':':
			// : moves from chapter to verse
			wb()
			if startRef && !inChapter {
				// we should have been in the chapter, but weren't, must be on the "w" of a w:x-y:z
				cv.StartChapter = unbuffer
				cv.EndChapter = unbuffer
			}
			if !startRef && !inChapter {
				// we should have been in the chapter, but weren't, must be on the "y" of a w:x-y:z
				cv.EndChapter = unbuffer
			}
			inChapter = false
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			workbuf.WriteRune(r)
		case 'a', 'b', 'f': // f is shorthand for ff
			wb()
			if startRef && !inChapter {
				cv.StartVerseSuffix = r
			}
			if !startRef && !inChapter {
				cv.EndVerseSuffix = r
			}
		default:
			log.Println("ignoring noise:", r)
		}
	}
	wb()

	out = append(out, cv)
	return out, nil
}
