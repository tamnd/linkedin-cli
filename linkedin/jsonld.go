package linkedin

import (
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ldNodes returns every JSON-LD node on the page. LinkedIn ships a single
// <script type="application/ld+json"> whose body is {"@context":..., "@graph":[...]},
// but a node can also appear bare, so both shapes are flattened into one slice.
func ldNodes(doc *goquery.Document) []json.RawMessage {
	var nodes []json.RawMessage
	doc.Find(`script[type="application/ld+json"]`).Each(func(_ int, s *goquery.Selection) {
		raw := strings.TrimSpace(s.Text())
		if raw == "" {
			return
		}
		var wrap struct {
			Graph []json.RawMessage `json:"@graph"`
		}
		if err := json.Unmarshal([]byte(raw), &wrap); err == nil && len(wrap.Graph) > 0 {
			nodes = append(nodes, wrap.Graph...)
			return
		}
		// A bare object, or an array of objects.
		var arr []json.RawMessage
		if err := json.Unmarshal([]byte(raw), &arr); err == nil && len(arr) > 0 {
			nodes = append(nodes, arr...)
			return
		}
		nodes = append(nodes, json.RawMessage(raw))
	})
	return nodes
}

// nodeType reports the @type of a JSON-LD node ("" when absent).
func nodeType(raw json.RawMessage) string {
	var n struct {
		Type string `json:"@type"`
	}
	_ = json.Unmarshal(raw, &n)
	return n.Type
}

// firstNodeOfType returns the first node whose @type matches one of types.
func firstNodeOfType(nodes []json.RawMessage, types ...string) (json.RawMessage, bool) {
	for _, raw := range nodes {
		t := nodeType(raw)
		for _, want := range types {
			if t == want {
				return raw, true
			}
		}
	}
	return nil, false
}

// nodesOfType returns every node whose @type matches one of types.
func nodesOfType(nodes []json.RawMessage, types ...string) []json.RawMessage {
	var out []json.RawMessage
	for _, raw := range nodes {
		t := nodeType(raw)
		for _, want := range types {
			if t == want {
				out = append(out, raw)
			}
		}
	}
	return out
}

// ldString unmarshals a value that may be a plain string or an object carrying a
// "name"/"contentUrl"/"url" field, returning the most useful scalar.
func ldString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return strings.TrimSpace(s)
	}
	var obj struct {
		Name       string `json:"name"`
		ContentURL string `json:"contentUrl"`
		URL        string `json:"url"`
	}
	if json.Unmarshal(raw, &obj) == nil {
		switch {
		case obj.ContentURL != "":
			return obj.ContentURL
		case obj.Name != "":
			return obj.Name
		case obj.URL != "":
			return obj.URL
		}
	}
	return ""
}

// ldStrings unmarshals a value that may be a single string or an array of
// strings into a slice.
func ldStrings(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var arr []string
	if json.Unmarshal(raw, &arr) == nil {
		return arr
	}
	var one string
	if json.Unmarshal(raw, &one) == nil && one != "" {
		return []string{one}
	}
	return nil
}
