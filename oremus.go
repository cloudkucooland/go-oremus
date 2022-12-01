package oremus

import (
	"bytes"
	"context"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

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
		// log.Println(err.Error())
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// log.Printf("%s\n", err.Error())
		return "", err
	}

	parsed := parse(string(body[:]))
	return parsed, nil
}

func parse(in string) string {
	z := html.NewTokenizer(strings.NewReader(in))
	var out = bytes.Buffer{}
	var inLection = false
	var passageDepth = 0

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			// hit EOF -- cleanup double-spaces on the way out
			b := bytes.Buffer{}
			prevIsSpace := false

			for {
				i, _, err := out.ReadRune()
				if err != nil {
					return b.String()
				}
				if unicode.IsSpace(i) {
					if !prevIsSpace {
						b.WriteRune(' ')
					}
					prevIsSpace = true
				} else {
					b.WriteRune(i)
					prevIsSpace = false
				}
			}
			return b.String()
		case html.TextToken:
			if inLection {
				out.Write(z.Text())
			}
		case html.StartTagToken:
			tn, hasAttr := z.TagName()
			if inLection {
				switch string(tn) {
				case "p":
					out.WriteString("<p>")
				case "nn":
					out.WriteString("<i>")
				case "span":
					out.WriteString(" <span class='adonai'>")
				default:
					// log.Printf("%+v\n", string(tn))
				}
				passageDepth++
			}
			// oremus tags the lection <p class="bibletext">
			// we only pay attention to this block
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
				tn, _ := z.TagName()
				switch string(tn) {
				case "p":
					out.WriteString("</p>\n")
				case "nn":
					out.WriteString("</i>")
				case "span":
					out.WriteString("</span> ")
				default:
					// log.Printf("%+v\n", string(tn))
				}
				if passageDepth == 0 { // found the </p> to close class="bibletext" -- quit processing
					inLection = false
				}
				passageDepth--
			}
		case html.SelfClosingTagToken:
			if inLection {
				out.WriteString(" ") // can this be removed?
			}
		}
	}

	// not reached
	return out.String()
}
