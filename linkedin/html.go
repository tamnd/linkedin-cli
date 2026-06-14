package linkedin

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var reTags = regexp.MustCompile(`<[^>]+>`)
var reWS = regexp.MustCompile(`\s+`)

// cleanHTML strips tags and unescapes the common HTML entities.
func cleanHTML(s string) string {
	s = reTags.ReplaceAllString(s, "")
	r := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
		"&nbsp;", " ",
	)
	return strings.TrimSpace(r.Replace(s))
}

// cleanText strips tags, unescapes entities, and collapses interior whitespace.
func cleanText(s string) string {
	return strings.TrimSpace(reWS.ReplaceAllString(cleanHTML(s), " "))
}

// atoiClean parses an integer from text that may carry commas and stray runes.
func atoiClean(s string) int {
	s = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
	n, _ := strconv.Atoi(s)
	return n
}

// metaContent returns the content of the first <meta property="..."> (or name)
// matching key.
func metaContent(doc *goquery.Document, key string) string {
	var out string
	doc.Find(`meta`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		prop, _ := s.Attr("property")
		name, _ := s.Attr("name")
		if prop == key || name == key {
			out, _ = s.Attr("content")
			return false
		}
		return true
	})
	return strings.TrimSpace(out)
}

// firstText returns the trimmed, whitespace-collapsed text of the first node
// matching sel within root.
func firstText(root *goquery.Selection, sel string) string {
	return cleanText(root.Find(sel).First().Text())
}
