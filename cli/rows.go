package cli

import (
	"strconv"
	"strings"

	"github.com/tamnd/linkedin-cli/linkedin"
)

// Entity → Row mappers: a curated default column set for the table/list/csv
// views, with the full typed object kept as Value for json, jsonl, and template.

func profileRow(p *linkedin.Profile) Row {
	return Row{
		Cols:  []string{"slug", "name", "headline", "location", "followers", "url"},
		Vals:  []string{p.Slug, oneline(p.Name), oneline(p.Headline), oneline(p.Location), itoa64(p.Followers), p.URL},
		Value: p,
	}
}

func companyRow(c *linkedin.Company) Row {
	return Row{
		Cols:  []string{"slug", "name", "industry", "company_size", "followers", "url"},
		Vals:  []string{c.Slug, oneline(c.Name), oneline(c.Industry), oneline(c.CompanySize), itoa64(c.Followers), c.URL},
		Value: c,
	}
}

func jobRow(j *linkedin.Job) Row {
	return Row{
		Cols:  []string{"job_id", "title", "company", "location", "posted", "url"},
		Vals:  []string{j.JobID, oneline(j.Title), oneline(j.Company), oneline(j.Location), j.Posted, j.URL},
		Value: j,
	}
}

func jobStubRow(s *linkedin.JobStub) Row {
	return Row{
		Cols:  []string{"job_id", "title", "company", "location", "posted", "url"},
		Vals:  []string{s.JobID, oneline(s.Title), oneline(s.Company), oneline(s.Location), s.Posted, s.URL},
		Value: s,
	}
}

func postRow(p *linkedin.Post) Row {
	return Row{
		Cols:  []string{"author", "title", "published", "likes", "comments", "url"},
		Vals:  []string{oneline(p.Author), oneline(p.Title), p.Published, itoa64(p.Likes), itoa64(p.Comments), p.URL},
		Value: p,
	}
}

func articleRow(a *linkedin.Article) Row {
	return Row{
		Cols:  []string{"title", "author", "published", "reactions", "comments", "url"},
		Vals:  []string{oneline(a.Title), oneline(a.Author), a.Published, itoa64(a.Reactions), itoa64(a.Comments), a.URL},
		Value: a,
	}
}

func locationRow(l *linkedin.Location) Row {
	return Row{
		Cols:  []string{"slug", "primary", "address", "url"},
		Vals:  []string{l.Slug, yn(l.Primary), oneline(l.Address), l.URL},
		Value: l,
	}
}

func orgRefRow(o *linkedin.OrgRef) Row {
	return Row{
		Cols:  []string{"slug", "name", "industry", "location", "url"},
		Vals:  []string{o.Slug, oneline(o.Name), oneline(o.Industry), oneline(o.Location), o.URL},
		Value: o,
	}
}

func refRow(r linkedin.Ref) Row {
	return Row{
		Cols:  []string{"input", "kind", "id", "url"},
		Vals:  []string{r.Input, r.Kind, r.ID, r.URL},
		Value: r,
	}
}

func itoa64(n int64) string { return strconv.FormatInt(n, 10) }

func yn(b bool) string {
	if b {
		return "yes"
	}
	return ""
}

func oneline(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 80 {
		return s[:79] + "…"
	}
	return s
}
