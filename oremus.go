package oremus

import (
	"bytes"
	"context"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Get fetches a passage from bible.oremus.org, parses the result and returns the text formatted as simple HTML
func Get(ctx context.Context, ref string) (string, error) {
	c := http.Client{}
	data := url.Values{}
	cleanref := strings.Trim(ref, "	")
	data.Set("passage", cleanref)
	data.Set("vnum", "no")
	data.Set("fnote", "no")
	data.Set("heading", "no")
	data.Set("show_ref", "no")
	data.Set("show_adj", "no")
	data.Set("omithidden", "yes")

	resp, err := c.PostForm("https://bible.oremus.org/", data)
	if err != nil {
		log.Printf("%v\n", err)
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%v\n", err)
		return "", err
	}

	parsed := parse(string(body[:]))
	return parsed, nil
}

// parse is a special purpose parser for bible.oremus.org's results
func parse(in string) string {
	z := html.NewTokenizer(strings.NewReader(in))
	var out = bytes.Buffer{}
	var inLection = false
	var passageDepth = 0
	var prevIsTextToken = false

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			// hit EOF, quit parsing
			return out.String()
		case html.TextToken:
			if inLection {
				if prevIsTextToken {
					out.WriteRune(' ')
				}
				txt := z.Text()
				out.WriteString(strings.TrimSpace(string(txt)))
				prevIsTextToken = true
			}
		case html.StartTagToken:
			tn, hasAttr := z.TagName()
			if inLection {
				prevIsTextToken = false

				switch string(tn) {
				case "p":
					out.WriteString("<p>")
				case "nn":
					out.WriteString("\n<i>")
				case "span":
					out.WriteString("\n<span class='adonai'>")
				default:
					log.Printf("unprocessed open tag %+v\n", string(tn))
				}
				passageDepth++
			}
			// oremus tags the lection <p class="bibletext">
			// we pay attention to only this block
			if hasAttr {
				for hasAttr {
					ta, val, attr := z.TagAttr()
					hasAttr = attr
					if string(ta) == "class" && string(val) == "bibletext" {
						inLection = true
					}
				}
			}
		case html.EndTagToken:
			if inLection {
				prevIsTextToken = false

				tn, _ := z.TagName()
				switch string(tn) {
				case "p":
					out.WriteString("</p>\n")
				case "nn":
					out.WriteString("</i>\n")
				case "span":
					out.WriteString("</span>\n")
				default:
					log.Printf("unprocessed close tag %+v\n", string(tn))
				}
				if passageDepth == 0 { // found the </p> to close class="bibletext" -- quit processing
					inLection = false
				}
				passageDepth--
			}
		case html.SelfClosingTagToken:
			if inLection {
				prevIsTextToken = false

				tn, _ := z.TagName()
				switch string(tn) {
				case "br":
					out.WriteString("<br />")
				default:
					log.Printf("unprocessed self-close tag <%s />\n", tn)
				}
			}
		}
	}

	// not reached
	return out.String()
}
