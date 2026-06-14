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
	MemberOf             []affiliationLD `json:"memberOf"`
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
			Slug:      slugFromAffiliationURL(a.URL),
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
		MemberOf:    toAffiliations(p.MemberOf),
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
	parseCompanyAbout(doc, c)
	parseCompanyFunding(doc, c)
	return c, nil
}

// parseCompanyFunding reads the funding card a company page carries: the total
// number of rounds and the Crunchbase link to the most recent round.
func parseCompanyFunding(doc *goquery.Document, c *Company) {
	doc.Find(`a[href*="crunchbase.com/funding_round/"]`).EachWithBreak(func(_ int, a *goquery.Selection) bool {
		if href, ok := a.Attr("href"); ok {
			c.FundingURL = stripQuery(href)
			return false
		}
		return true
	})
	doc.Find("body").Each(func(_ int, b *goquery.Selection) {
		if m := reFundingRounds.FindStringSubmatch(b.Text()); len(m) == 2 && c.FundingRounds == 0 {
			c.FundingRounds = atoiClean(m[1])
		}
	})
}

// ParseCompanyLocations reads the office cards a company page lists, one per
// "Get directions" entry, marking the primary office (the headquarters).
func ParseCompanyLocations(doc *goquery.Document, slug, pageURL string, now time.Time) []Location {
	var out []Location
	doc.Find(`a[href*="bing.com/maps"], a[href*="google.com/maps"]`).Each(func(_ int, a *goquery.Selection) {
		li := a.Closest("li")
		if li.Length() == 0 {
			return
		}
		var lines []string
		li.Find(`div[id^="address"] p`).Each(func(_ int, p *goquery.Selection) {
			if s := cleanText(p.Text()); s != "" {
				lines = append(lines, s)
			}
		})
		if len(lines) == 0 {
			return
		}
		primary := strings.Contains(li.Find("span").Text(), "Primary")
		loc := Location{Slug: slug, Primary: primary, URL: pageURL, FetchedAt: now}
		loc.Street = lines[0]
		if len(lines) > 1 {
			loc.Address = strings.Join(lines[1:], ", ")
		}
		out = append(out, loc)
	})
	return out
}

// ParseCompanyAffiliated reads the related pages a company links to (affiliated
// pages and showcase pages), each with its name, industry, and location.
func ParseCompanyAffiliated(doc *goquery.Document, now time.Time) []OrgRef {
	var out []OrgRef
	seen := map[string]bool{}
	doc.Find(`a[href*="trk=affiliated-pages"]`).Each(func(_ int, a *goquery.Selection) {
		href, ok := a.Attr("href")
		if !ok {
			return
		}
		url := stripQuery(href)
		if seen[url] {
			return
		}
		seen[url] = true
		ref := OrgRef{
			Slug:      slugFromAffiliationURL(url),
			URL:       url,
			FetchedAt: now,
		}
		card := a.Closest("li")
		if card.Length() == 0 {
			card = a
		}
		if h := strings.TrimSpace(card.Find("h3").First().Text()); h != "" {
			ref.Name = cleanText(h)
		} else {
			ref.Name = cleanText(a.Text())
		}
		ps := card.Find("p")
		if ps.Length() > 0 {
			ref.Industry = cleanText(ps.Eq(0).Text())
		}
		if ps.Length() > 1 {
			ref.Location = cleanText(ps.Eq(1).Text())
		}
		out = append(out, ref)
	})
	return out
}

// parseCompanyAbout fills the fields the Organization JSON-LD leaves out by
// reading the company page's about panel: a list of <dt> label / <dd> value
// pairs (Industry, Company size, Type, Headquarters, Founded, Specialties,
// Website) plus the follower count carried in the Open Graph description.
func parseCompanyAbout(doc *goquery.Document, c *Company) {
	doc.Find("dt").Each(func(_ int, dt *goquery.Selection) {
		label := strings.TrimSpace(dt.Text())
		dd := dt.NextFiltered("dd")
		if dd.Length() == 0 {
			return
		}
		val := cleanText(dd.Text())
		switch label {
		case "Industry":
			if c.Industry == "" {
				c.Industry = val
			}
		case "Company size":
			if c.CompanySize == "" {
				c.CompanySize = val
			}
		case "Type":
			if c.CompanyType == "" {
				c.CompanyType = val
			}
		case "Headquarters":
			if c.Headquarters == "" {
				c.Headquarters = val
			}
		case "Founded":
			if c.Founded == "" {
				c.Founded = val
			}
		case "Specialties":
			if c.Specialties == "" {
				c.Specialties = val
			}
		case "Website":
			if c.Website == "" {
				c.Website = firstField(val)
			}
		}
	})
	if c.Followers == 0 {
		c.Followers = int64(atoiClean(followersFromText(metaContent(doc, "og:description"))))
	}
}

// followersFromText pulls the "<n> followers" count out of a blob of text, such
// as a company's Open Graph description ("Acme | 12,345 followers on LinkedIn").
func followersFromText(s string) string {
	m := reFollowers.FindStringSubmatch(s)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

// firstField returns the first whitespace-delimited token, used to drop the
// "External link for X" suffix LinkedIn appends to the about-panel website.
func firstField(s string) string {
	if i := strings.IndexAny(s, " \t\n"); i >= 0 {
		return s[:i]
	}
	return s
}

// ── Posts (best effort) ─────────────────────────────────────────────────────

type postLD struct {
	Name                 string          `json:"name"`
	Headline             string          `json:"headline"`
	Text                 string          `json:"text"`
	DatePublished        string          `json:"datePublished"`
	DateModified         string          `json:"dateModified"`
	URL                  string          `json:"url"`
	Image                json.RawMessage `json:"image"`
	Author               json.RawMessage `json:"author"`
	InteractionStatistic json.RawMessage `json:"interactionStatistic"`
}

// postTitle prefers the JSON-LD name (the real headline on a pulse article)
// and falls back to the headline field (which a feed post carries instead, and
// which a pulse page reuses for the body's opening hook).
func postTitle(name, headline string) string {
	if t := cleanText(name); t != "" {
		return t
	}
	return cleanText(headline)
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
			post.Title = postTitle(p.Name, p.Headline)
			post.Text = cleanText(p.Text)
			post.Published = p.DatePublished
			post.Modified = p.DateModified
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

// postFromNode builds a Post from a DiscussionForumPosting/Article JSON-LD node.
func postFromNode(raw json.RawMessage, now time.Time) (Post, bool) {
	var p postLD
	if json.Unmarshal(raw, &p) != nil {
		return Post{}, false
	}
	a := parseAuthor(p.Author)
	post := Post{
		URL:       canonicalURL(p.URL),
		Author:    strings.TrimSpace(a.Name),
		AuthorURL: a.URL,
		Title:     postTitle(p.Name, p.Headline),
		Text:      cleanText(p.Text),
		Published: p.DatePublished,
		Modified:  p.DateModified,
		ImageURL:  ldString(p.Image),
		Likes:     interactionCount(p.InteractionStatistic, "LikeAction"),
		Comments:  interactionCount(p.InteractionStatistic, "CommentAction"),
		FetchedAt: now,
	}
	return post, true
}

// ── Articles ────────────────────────────────────────────────────────────────

type articleLD struct {
	Name                 string          `json:"name"`
	Headline             string          `json:"headline"`
	URL                  string          `json:"url"`
	DatePublished        string          `json:"datePublished"`
	DateModified         string          `json:"dateModified"`
	Image                json.RawMessage `json:"image"`
	Author               json.RawMessage `json:"author"`
	InteractionStatistic json.RawMessage `json:"interactionStatistic"`
}

// articleFromNode builds an Article from an Article JSON-LD node.
func articleFromNode(raw json.RawMessage, now time.Time) (Article, bool) {
	var a articleLD
	if json.Unmarshal(raw, &a) != nil {
		return Article{}, false
	}
	au := parseAuthor(a.Author)
	return Article{
		URL:       canonicalURL(a.URL),
		Title:     postTitle(a.Name, a.Headline),
		Author:    strings.TrimSpace(au.Name),
		AuthorURL: au.URL,
		Published: a.DatePublished,
		Modified:  a.DateModified,
		Reactions: interactionCount(a.InteractionStatistic, "LikeAction"),
		Comments:  interactionCount(a.InteractionStatistic, "CommentAction"),
		ImageURL:  ldString(a.Image),
		FetchedAt: now,
	}, true
}
