package oremus

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const unset = rune(0)

// Reference is the datastructure used for parsing/validating references
type Reference struct {
	Book              string
	ChapterVerseRange []ChapterVerseRange
	Prefix            rune
}

// ChapterVerseRange is the chapter/verse portion of a reference "1:14-2:7a"
type ChapterVerseRange struct {
	StartVerseSuffix rune // a, b, f
	EndVerseSuffix   rune // a, b, f
	StartChapter     int
	StartVerse       int
	EndChapter       int
	EndVerse         int
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
	"Acts":          {"acts", "acts of the apostles"},
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
var booksWithPrefix = map[string][]rune{
	"samuel": {'1', '2'}, "kings": {'1', '2'}, "chronicles": {'1', '2'}, "corinthians": {'1', '2'}, "thessalonians": {'1', '2'}, "timothy": {'1', '2'}, "peter": {'1', '2'}, "john": {'1', '2', '3'},
}

func allowedPrefix(book string, prefix rune) bool {
	list, ok := booksWithPrefix[strings.ToLower(book)]
	if !ok {
		return false
	}
	return slices.Contains(list, prefix)
}

var bookLookup map[string]string
var prefixLookup map[string]rune

func init() {
	bookLookup = make(map[string]string)
	for canonical, variants := range books {
		for _, v := range variants {
			bookLookup[v] = canonical
		}
	}

	prefixLookup = make(map[string]rune)
	for k, variants := range prefixes {
		for _, v := range variants {
			prefixLookup[v] = k
		}
	}
}

// String returns a normalized reference to a scripture passage
func (r *Reference) String() string {
	first := true
	var prevChap int

	var buf strings.Builder
	if r.Prefix != unset {
		buf.WriteRune(r.Prefix)
		buf.WriteRune(' ')
	}

	if r.Book != "" {
		buf.WriteString(r.Book)
	}

	if len(r.ChapterVerseRange) > 0 && r.ChapterVerseRange[0].StartChapter != 0 {
		buf.WriteRune(' ')
	}

	for _, v := range r.ChapterVerseRange {
		if !first {
			buf.WriteString(",")
		} else {
			first = false
		}

		if v.StartChapter != prevChap {
			buf.WriteString(strconv.Itoa(v.StartChapter))
		}
		if v.StartVerse != 0 {
			if v.StartChapter != prevChap {
				buf.WriteRune(':')
			}
			buf.WriteString(strconv.Itoa(v.StartVerse))
			if v.StartVerseSuffix != unset {
				if v.StartVerseSuffix != 'f' {
					buf.WriteRune(v.StartVerseSuffix)
				} else {
					buf.WriteString("ff")
				}
			}

			// catch edge-cases
			if v.EndVerse == 0 {
				v.EndVerse = v.StartVerse
			}
		}

		if v.EndChapter != v.StartChapter || v.EndVerse != v.StartVerse {
			buf.WriteRune('-')
		}
		if v.EndChapter != v.StartChapter {
			buf.WriteString(strconv.Itoa(v.EndChapter))
			if v.EndVerse != 0 {
				buf.WriteRune(':')
			}
		}
		if v.EndVerse != v.StartVerse {
			buf.WriteString(strconv.Itoa(v.EndVerse))
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

func parseBook(chunks []string) (book string, prefix rune, rest []string, err error) {
	if len(chunks) == 0 {
		return "", 0, nil, errors.New("empty reference")
	}

	lcChunks := make([]string, len(chunks))
	for i, c := range chunks {
		lcChunks[i] = strings.ToLower(strings.TrimSuffix(c, "."))
	}

	// try prefix first
	if p, ok := prefixLookup[lcChunks[0]]; ok {
		for i := len(chunks); i > 1; i-- {
			candidate := strings.Join(lcChunks[1:i], " ")
			if b, ok := bookLookup[candidate]; ok && allowedPrefix(b, p) {
				return b, p, chunks[i:], nil
			}
		}
		return "", 0, nil, errors.New("invalid prefixed book")
	}

	// no prefix
	for i := len(chunks); i > 0; i-- {
		candidate := strings.Join(lcChunks[:i], " ")
		if b, ok := bookLookup[candidate]; ok {
			return b, 0, chunks[i:], nil
		}
	}

	return "", 0, nil, errors.New("invalid book")
}

// CleanReference takes a string, returns a normalized string for a semi-colon separated list of scripture references
func CleanReference(in string) (string, error) {
	r, err := ParseReferences(in)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
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
		if r == "" {
			continue
		}
		parsed, err := ParseReference(r)
		if err != nil {
			return nil, err
		}
		out = append(out, parsed)
	}
	return out, nil
}

// ParseReference parses a single free-form reference to a scripture passage and returns a *Reference
// use the stringer method to get a normalized string format back
func ParseReference(in string) (*Reference, error) {
	newRef := Reference{}
	var rest []string
	var err error

	newRef.Book, newRef.Prefix, rest, err = parseBook(strings.Fields(in))
	if err != nil {
		return nil, err
	}
	if newRef.Book == "" {
		return nil, errors.New("invalid book")
	}
	newRef.ChapterVerseRange, err = parseChapterVerse(strings.Join(rest, " "))
	if err != nil {
		return nil, err
	}

	return &newRef, nil
}

/*
func isBook(in string) (bool, string) {
	lc := strings.TrimSuffix(strings.ToLower(in), ".")
	book, ok := bookLookup[lc]
	return ok, book
}

func isPrefix(in string) (bool, rune) {
	lc := strings.ToLower(in)
	p, ok := prefixLookup[lc]
	return ok, p
} */

/*
func isSuffix(r rune) bool {
	return r == 'a' || r == 'b' || r == 'f'
} */

type parseState int

const (
	stateStartChapter parseState = iota // before first :
	stateStartVerse                     // after first :
	stateAmbiguous                      // working on something after a -, could be endchapter or endverse
	stateEndVerse                       // working on the end verse
	stateAfterSuffix                    // after suffix, before -,;
)

func parseChapterVerse(in string) ([]ChapterVerseRange, error) {
	out := make([]ChapterVerseRange, 0) // slice for multiple references
	current := ChapterVerseRange{}      // the reference we are working on
	var workbuf strings.Builder
	var state parseState

	// var toResolve int

	// a helper to save some labor below
	var flushBuffer = func() (int, error) {
		s := workbuf.String()
		if s == "" {
			return 0, nil
		}
		asInt, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("invalid numeric token %q", s)
		}
		workbuf.Reset()
		return asInt, nil
	}

	for _, r := range in {
		// toResolove = 0

		switch r {
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			workbuf.WriteRune(r)
		case 'a', 'b', 'f': // isSuffix() -- f is shorthand for ff
			i, err := flushBuffer()
			if err != nil {
				return nil, err
			}

			switch state {
			case stateStartVerse:
				current.StartVerse = i
				current.EndVerse = i
				current.StartVerseSuffix = r
				state = stateAfterSuffix
			case stateEndVerse, stateAmbiguous:
				current.EndVerse = i
				current.EndVerseSuffix = r
				state = stateAfterSuffix
			case stateAfterSuffix:
				// ignore (ff)
			default:
				return nil, errors.New("suffix but not on verse")
			}
		case ',', unset:
			// , ends the current reference and starts a new one
			i, err := flushBuffer()
			if err != nil {
				return nil, err
			}
			switch state {
			case stateStartChapter:
				// whole chapter reference (7 in Gen 1,7)
				current.StartChapter = i
				current.EndChapter = i
			case stateStartVerse:
				// must be a single-verse reference (Gen 1:1)
				current.StartVerse = i
				current.EndVerse = i
			case stateAmbiguous:
				// if we have a verse, assume verse, otherwise assume chapter
				if current.StartVerse != 0 {
					// Gen 1:1-5
					current.EndVerse = i
				} else {
					// Gen 1-2
					current.EndChapter = i
				}
			case stateEndVerse:
				// Gen 1:2-3:4
				current.EndVerse = i
			default:
				return nil, errors.New("comma in invalid state")
			}

			// catch double ,,
			if current.StartChapter == 0 {
				continue
			}

			if err := current.Validate(); err != nil {
				return nil, err
			}
			out = append(out, current)
			cchap := current.StartChapter
			current = ChapterVerseRange{}
			// prime the new reference in case of the following
			if cchap != 0 { // Gen 7:1,4,9 -- chapter 7 verses 1, 4 and 9; that is 3 ChapterVerseRanges
				current.StartChapter = cchap
				current.EndChapter = cchap
			}
			state = stateStartChapter
		case '-', '—', '–': // hyphen, en dash, and em dashes all found in the wild
			// - moves from the first part of a reference to the end of one (either chapter or verse)
			i, err := flushBuffer()
			if err != nil {
				return nil, err
			}

			switch state {
			case stateStartChapter:
				current.StartChapter = i
				current.EndChapter = i
			case stateStartVerse:
				current.StartVerse = i
				current.EndVerse = i
			case stateAfterSuffix:
				// nothing
			default:
				return nil, fmt.Errorf("unexpected '-' in state %v", state)
			}
			state = stateAmbiguous
		case ':':
			i, err := flushBuffer()
			if err != nil {
				return nil, err
			}
			switch state {
			case stateStartChapter:
				// the "w" of a w:x-y:z
				current.StartChapter = i
				current.EndChapter = i
				state = stateStartVerse
			case stateAmbiguous: // resolves the ambiguity
				current.EndChapter = i
				state = stateEndVerse
			default:
				return nil, fmt.Errorf("unexpected ':' in state %v", state)
			}
		case ' ':
			// spaces can be safely ignored
		default:
			return nil, fmt.Errorf("invalid character in reference: %d", r)
		}
	}

	i, err := flushBuffer()
	if err != nil {
		return nil, err
	}
	switch state {
	case stateStartChapter:
		// whole chapter reference (7 in Gen 1,7)
		current.StartChapter = i
		current.EndChapter = i
	case stateStartVerse:
		// must be a single-verse reference (Gen 1:1)
		current.StartVerse = i
		current.EndVerse = i
	case stateAmbiguous:
		// if we have a verse, assume verse, otherwise assume chapter
		if current.StartVerse != 0 {
			// Gen 1:1-5
			current.EndVerse = i
		} else {
			// Gen 1-2
			current.EndChapter = i
		}
	case stateEndVerse:
		// Gen 1:2-3:4
		current.EndVerse = i
	case stateAfterSuffix:
		// nothing
	default:
		return nil, fmt.Errorf("parser ended in state %v", state)
	}
	if err := current.Validate(); err != nil {
		return nil, err
	}

	out = append(out, current)
	return out, nil
}

func (cv ChapterVerseRange) Validate() error {
	// fmt.Printf("%+v\n", cv)
	if cv.StartChapter > cv.EndChapter {
		return errors.New("invalid chapter range")
	}
	if cv.StartChapter == cv.EndChapter && cv.StartVerse > cv.EndVerse {
		return errors.New("invalid verse range")
	}
	// an entire book is a valid reference (e.g. James)
	/* if cv.StartChapter == 0 {
		    return errors.New("missing start chapter")
	    } */
	if cv.StartChapter == 0 && cv.StartVerse != 0 {
		return errors.New("verse without chapter")
	}
	if cv.StartVerse < 0 || cv.EndVerse < 0 {
		return errors.New("invalid verse")
	}

	return nil
}
