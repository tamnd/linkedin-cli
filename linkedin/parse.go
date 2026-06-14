package linkedin

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ── Profile ─────────────────────────────────────────────────────────────────

type personLD struct {
	Name                      string          `json:"name"`
	DisambiguatingDescription string          `json:"disambiguatingDescription"`
	Description               string          `json:"description"`
	JobTitle                  json.RawMessage `json:"jobTitle"`
	URL                       string          `json:"url"`
	SameAs                    json.RawMessage `json:"sameAs"`
	KnowsLanguage             json.RawMessage `json:"knowsLanguage"`
	Awards                    json.RawMessage `json:"awards"`
	Image                     struct {
		ContentURL string `json:"contentUrl"`
	} `json:"image"`
	Address struct {
		Locality string `json:"addressLocality"`
		Country  string `json:"addressCountry"`
	} `json:"address"`
	WorksFor             []affiliationLD `json:"worksFor"`
	AlumniOf             []affiliationLD `json:"alumniOf"`
	InteractionStatistic json.RawMessage `json:"interactionStatistic"`
}

type affiliationLD struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Member struct {
		StartDate int `json:"startDate"`
		EndDate   int `json:"endDate"`
	} `json:"member"`
}

type interactionLD struct {
	InteractionType json.RawMessage `json:"interactionType"`
	Count           int64           `json:"userInteractionCount"`
}

// interactionCount sums the userInteractionCount of the entries whose
// interactionType ends with the given schema action (e.g. "FollowAction").
func interactionCount(raw json.RawMessage, action string) int64 {
	if len(raw) == 0 {
		return 0
	}
	var arr []interactionLD
	if json.Unmarshal(raw, &arr) != nil {
		var one interactionLD
		if json.Unmarshal(raw, &one) != nil {
			return 0
		}
		arr = []interactionLD{one}
	}
	for _, it := range arr {
		t := ldString(it.InteractionType)
		if strings.HasSuffix(t, action) {
			return it.Count
		}
	}
	return 0
}

func toAffiliations(in []affiliationLD) []Affiliation {
	out := make([]Affiliation, 0, len(in))
	for _, a := range in {
		out = append(out, Affiliation{
			Name:      strings.TrimSpace(a.Name),
			URL:       a.URL,
			Slug:      slugFromCompanyURL(a.URL),
			StartDate: a.Member.StartDate,
			EndDate:   a.Member.EndDate,
		})
	}
	return out
}

// ParseProfile parses a member page into a Profile (JSON-LD Person first).
func ParseProfile(doc *goquery.Document, slug, pageURL string) (*Profile, error) {
	nodes := ldNodes(doc)
	raw, ok := firstNodeOfType(nodes, "Person")
	if !ok {
		return nil, ErrNotFound
	}
	var p personLD
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	prof := &Profile{
		Slug:        slug,
		URL:         pageURL,
		Name:        strings.TrimSpace(p.Name),
		Headline:    strings.TrimSpace(p.DisambiguatingDescription),
		Description: cleanText(p.Description),
		JobTitles:   ldStrings(p.JobTitle),
		Location:    strings.TrimSpace(p.Address.Locality),
		Country:     strings.TrimSpace(p.Address.Country),
		Followers:   interactionCount(p.InteractionStatistic, "FollowAction"),
		ImageURL:    p.Image.ContentURL,
		Languages:   ldStrings(p.KnowsLanguage),
		Awards:      ldStrings(p.Awards),
		WorksFor:    toAffiliations(p.WorksFor),
		AlumniOf:    toAffiliations(p.AlumniOf),
		SameAs:      ldStrings(p.SameAs),
		FetchedAt:   time.Now(),
	}
	if prof.Name == "" {
		prof.Name = metaContent(doc, "og:title")
	}
	if prof.ImageURL == "" {
		prof.ImageURL = metaContent(doc, "og:image")
	}
	return prof, nil
}

// ── Company ─────────────────────────────────────────────────────────────────

type orgLD struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Slogan      string `json:"slogan"`
	SameAs      string `json:"sameAs"`
	Address     struct {
		Street     string `json:"streetAddress"`
		Locality   string `json:"addressLocality"`
		Region     string `json:"addressRegion"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"addressCountry"`
	} `json:"address"`
	NumberOfEmployees struct {
		Value int64 `json:"value"`
	} `json:"numberOfEmployees"`
	Logo struct {
		ContentURL string `json:"contentUrl"`
	} `json:"logo"`
}

// ParseCompany parses a company page into a Company (JSON-LD Organization).
func ParseCompany(doc *goquery.Document, slug, pageURL string) (*Company, error) {
	nodes := ldNodes(doc)
	raw, ok := firstNodeOfType(nodes, "Organization")
	if !ok {
		return nil, ErrNotFound
	}
	var o orgLD
	if err := json.Unmarshal(raw, &o); err != nil {
		return nil, err
	}
	c := &Company{
		Slug:        slug,
		URL:         pageURL,
		Name:        strings.TrimSpace(o.Name),
		Description: cleanText(o.Description),
		Slogan:      cleanText(o.Slogan),
		Website:     strings.TrimSpace(o.SameAs),
		Employees:   o.NumberOfEmployees.Value,
		Street:      strings.TrimSpace(o.Address.Street),
		Locality:    strings.TrimSpace(o.Address.Locality),
		Region:      strings.TrimSpace(o.Address.Region),
		PostalCode:  strings.TrimSpace(o.Address.PostalCode),
		Country:     strings.TrimSpace(o.Address.Country),
		LogoURL:     o.Logo.ContentURL,
		FetchedAt:   time.Now(),
	}
	if c.Name == "" {
		c.Name = metaContent(doc, "og:title")
	}
	return c, nil
}

// ── Posts (best effort) ─────────────────────────────────────────────────────

type postLD struct {
	Headline             string          `json:"headline"`
	Text                 string          `json:"text"`
	DatePublished        string          `json:"datePublished"`
	URL                  string          `json:"url"`
	Image                json.RawMessage `json:"image"`
	Author               json.RawMessage `json:"author"`
	InteractionStatistic json.RawMessage `json:"interactionStatistic"`
}

type authorLD struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func parseAuthor(raw json.RawMessage) authorLD {
	if len(raw) == 0 {
		return authorLD{}
	}
	var arr []authorLD
	if json.Unmarshal(raw, &arr) == nil && len(arr) > 0 {
		return arr[0]
	}
	var one authorLD
	_ = json.Unmarshal(raw, &one)
	return one
}

// ParsePost parses a public post or article into a Post, JSON-LD first and Open
// Graph as a backstop.
func ParsePost(doc *goquery.Document, pageURL string) (*Post, error) {
	post := &Post{URL: pageURL, FetchedAt: time.Now()}
	nodes := ldNodes(doc)
	if raw, ok := firstNodeOfType(nodes, "Article", "DiscussionForumPosting", "SocialMediaPosting", "BlogPosting"); ok {
		var p postLD
		if err := json.Unmarshal(raw, &p); err == nil {
			a := parseAuthor(p.Author)
			post.Author = strings.TrimSpace(a.Name)
			post.AuthorURL = a.URL
			post.Title = cleanText(p.Headline)
			post.Text = cleanText(p.Text)
			post.Published = p.DatePublished
			post.ImageURL = ldString(p.Image)
			post.Likes = interactionCount(p.InteractionStatistic, "LikeAction")
			post.Comments = interactionCount(p.InteractionStatistic, "CommentAction")
		}
	}
	if post.Title == "" {
		post.Title = metaContent(doc, "og:title")
	}
	if post.Text == "" {
		post.Text = metaContent(doc, "og:description")
	}
	if post.ImageURL == "" {
		post.ImageURL = metaContent(doc, "og:image")
	}
	if post.Title == "" && post.Text == "" {
		return nil, ErrNotFound
	}
	return post, nil
}
