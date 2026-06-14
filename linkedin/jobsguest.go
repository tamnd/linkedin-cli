package linkedin

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// JobSearchOptions carries the query for the guest job-search endpoint. Every
// field maps to a parameter the public jobs page puts on the wire (spec §4.1).
type JobSearchOptions struct {
	Keywords   string
	Location   string
	GeoID      string
	Posted     string // f_TPR, e.g. r86400 (24h), r604800 (week), r2592000 (month)
	Remote     string // f_WT: 1 on-site, 2 remote, 3 hybrid
	Experience string // f_E: 1..6
	JobType    string // f_JT: F,P,C,T,I,V,O
	Sort       string // sortBy: R relevance, DD date
}

// searchURL builds the guest search URL for a page offset.
func (o JobSearchOptions) searchURL(start int) string {
	q := url.Values{}
	if o.Keywords != "" {
		q.Set("keywords", o.Keywords)
	}
	if o.Location != "" {
		q.Set("location", o.Location)
	}
	if o.GeoID != "" {
		q.Set("geoId", o.GeoID)
	}
	if o.Posted != "" {
		q.Set("f_TPR", o.Posted)
	}
	if o.Remote != "" {
		q.Set("f_WT", o.Remote)
	}
	if o.Experience != "" {
		q.Set("f_E", o.Experience)
	}
	if o.JobType != "" {
		q.Set("f_JT", o.JobType)
	}
	if o.Sort != "" {
		q.Set("sortBy", o.Sort)
	}
	q.Set("start", strconv.Itoa(start))
	return GuestJobsSearch + "?" + q.Encode()
}

// SearchJobs walks the guest search endpoint, page by page, until it has limit
// stubs or the endpoint returns an empty page. A limit of 0 means "one page".
func (c *Client) SearchJobs(ctx context.Context, cache *Cache, cfg Config, opts JobSearchOptions, limit int) ([]JobStub, error) {
	var out []JobStub
	seen := map[string]bool{}
	for start := 0; ; start += 25 {
		body, err := c.CachingFetch(ctx, cache, cfg, opts.searchURL(start))
		if err != nil {
			if len(out) > 0 {
				return out, nil
			}
			return nil, err
		}
		doc, err := DocFromBytes(body)
		if err != nil {
			return out, err
		}
		stubs := parseJobCards(doc)
		if len(stubs) == 0 {
			break
		}
		fresh := 0
		for _, s := range stubs {
			if seen[s.JobID] {
				continue
			}
			seen[s.JobID] = true
			fresh++
			out = append(out, s)
			if limit > 0 && len(out) >= limit {
				return out, nil
			}
		}
		// A page that adds nothing new means we have reached the tail.
		if fresh == 0 {
			break
		}
		if limit == 0 {
			break
		}
	}
	return out, nil
}

// parseJobCards parses the <li> job cards returned by the guest search endpoint.
func parseJobCards(doc *goquery.Document) []JobStub {
	var out []JobStub
	now := time.Now()
	doc.Find("li").Each(func(_ int, li *goquery.Selection) {
		card := li.Find("div.base-card").First()
		if card.Length() == 0 {
			card = li
		}
		urn, _ := card.Attr("data-entity-urn")
		id := NormalizeJobID(urn)
		if id == "" || !reDigits.MatchString(id) {
			return
		}
		s := JobStub{
			JobID:     id,
			URL:       JobURL(id),
			Title:     firstText(li, ".base-search-card__title"),
			Company:   firstText(li, ".base-search-card__subtitle"),
			Location:  firstText(li, ".job-search-card__location"),
			Benefits:  firstText(li, ".job-posting-benefits"),
			FetchedAt: now,
		}
		if a := li.Find(".base-search-card__subtitle a").First(); a.Length() > 0 {
			if href, ok := a.Attr("href"); ok {
				s.CompanyURL = canonicalURL(href)
				s.CompanySlug = slugFromCompanyURL(href)
			}
		}
		if t := li.Find("time").First(); t.Length() > 0 {
			if dt, ok := t.Attr("datetime"); ok && dt != "" {
				s.Posted = dt
			} else {
				s.Posted = cleanText(t.Text())
			}
		}
		out = append(out, s)
	})
	return out
}

// FetchJob fetches one job posting through the guest detail fragment.
func (c *Client) FetchJob(ctx context.Context, cache *Cache, cfg Config, id string) (*Job, error) {
	id = NormalizeJobID(id)
	if id == "" {
		return nil, ErrNotFound
	}
	body, err := c.CachingFetch(ctx, cache, cfg, GuestJobDetail+id)
	if err != nil {
		return nil, err
	}
	doc, err := DocFromBytes(body)
	if err != nil {
		return nil, err
	}
	return parseJobDetail(doc, id)
}

// parseJobDetail parses the guest job-detail fragment into a Job.
func parseJobDetail(doc *goquery.Document, id string) (*Job, error) {
	j := &Job{JobID: id, URL: JobURL(id), FetchedAt: time.Now()}
	j.Title = firstText(doc.Selection, ".topcard__title")
	j.Posted = firstText(doc.Selection, ".posted-time-ago__text")
	j.Applicants = atoiClean(firstText(doc.Selection, ".num-applicants__caption"))

	if org := doc.Find(".topcard__org-name-link").First(); org.Length() > 0 {
		j.Company = cleanText(org.Text())
		if href, ok := org.Attr("href"); ok {
			j.CompanyURL = canonicalURL(href)
			j.CompanySlug = slugFromCompanyURL(href)
		}
	}
	// The bulleted flavor holds the location; the first flavor is the org name.
	if loc := doc.Find(".topcard__flavor--bullet").First(); loc.Length() > 0 {
		j.Location = cleanText(loc.Text())
	}
	if desc := doc.Find(".show-more-less-html__markup").First(); desc.Length() > 0 {
		j.Description = cleanText(desc.Text())
	} else {
		j.Description = firstText(doc.Selection, ".description__text")
	}

	doc.Find(".description__job-criteria-item").Each(func(_ int, item *goquery.Selection) {
		head := strings.ToLower(firstText(item, ".description__job-criteria-subheader"))
		val := firstText(item, ".description__job-criteria-text")
		switch {
		case strings.Contains(head, "seniority"):
			j.Seniority = val
		case strings.Contains(head, "employment"):
			j.EmploymentType = val
		case strings.Contains(head, "job function"):
			j.JobFunction = val
		case strings.Contains(head, "industries"):
			j.Industries = val
		}
	})

	if j.Title == "" {
		return nil, ErrNotFound
	}
	return j, nil
}
