package linkedin

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func docOf(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}
	return doc
}

// A trimmed copy of the @graph LinkedIn ships on a public profile page.
const profileHTML = `<html><head>
<script type="application/ld+json">
{"@context":"http://schema.org","@graph":[
{"@type":"WebPage","url":"https://www.linkedin.com/in/williamhgates"},
{"@type":"Person","name":"Bill Gates","disambiguatingDescription":"Creator, Top Voice",
 "description":"Chair of Gates Foundation.",
 "jobTitle":["Co-chair","Founder"],
 "address":{"@type":"PostalAddress","addressCountry":"US","addressLocality":"Seattle, Washington, United States"},
 "image":{"@type":"ImageObject","contentUrl":"https://media.licdn.com/photo.jpg"},
 "sameAs":"https://www.linkedin.com/in/williamhgates",
 "url":"https://www.linkedin.com/in/williamhgates",
 "knowsLanguage":["English"],
 "worksFor":[{"@type":"Organization","name":"Gates Foundation","url":"https://www.linkedin.com/company/gates-foundation","member":{"@type":"OrganizationRole","startDate":2000}},
             {"@type":"Organization","name":"Microsoft","url":"https://www.linkedin.com/company/microsoft","member":{"@type":"OrganizationRole","startDate":1975}}],
 "alumniOf":[{"@type":"EducationalOrganization","name":"Harvard University","url":"https://www.linkedin.com/school/harvard-university/","member":{"@type":"OrganizationRole","startDate":1973,"endDate":1975}}],
 "interactionStatistic":{"@type":"InteractionCounter","interactionType":"https://schema.org/FollowAction","userInteractionCount":40419217}}
]}
</script></head><body></body></html>`

func TestParseProfile(t *testing.T) {
	doc := docOf(t, profileHTML)
	p, err := ParseProfile(doc, "williamhgates", "https://www.linkedin.com/in/williamhgates")
	if err != nil {
		t.Fatalf("ParseProfile: %v", err)
	}
	if p.Name != "Bill Gates" {
		t.Errorf("name = %q", p.Name)
	}
	if p.Headline != "Creator, Top Voice" {
		t.Errorf("headline = %q", p.Headline)
	}
	if p.Location != "Seattle, Washington, United States" || p.Country != "US" {
		t.Errorf("location = %q country = %q", p.Location, p.Country)
	}
	if p.Followers != 40419217 {
		t.Errorf("followers = %d", p.Followers)
	}
	if len(p.JobTitles) != 2 || p.JobTitles[0] != "Co-chair" {
		t.Errorf("job titles = %v", p.JobTitles)
	}
	if len(p.WorksFor) != 2 || p.WorksFor[0].Slug != "gates-foundation" || p.WorksFor[0].StartDate != 2000 {
		t.Errorf("works for = %+v", p.WorksFor)
	}
	if len(p.AlumniOf) != 1 || p.AlumniOf[0].EndDate != 1975 {
		t.Errorf("alumni = %+v", p.AlumniOf)
	}
	if p.ImageURL != "https://media.licdn.com/photo.jpg" {
		t.Errorf("image = %q", p.ImageURL)
	}
}

const companyHTML = `<html><head>
<script type="application/ld+json">
{"@context":"http://schema.org","@graph":[
{"@type":"DiscussionForumPosting","url":"https://www.linkedin.com/posts/x","text":"hi","datePublished":"2026-01-01"},
{"@type":"Organization","name":"Xsolla","url":"https://www.linkedin.com/company/xsolla",
 "description":"Video game business engine.","slogan":"Sell more games.",
 "sameAs":"xsolla.com",
 "address":{"streetAddress":"15260 Ventura Blvd","addressLocality":"Sherman Oaks","addressRegion":"CA","postalCode":"91403","addressCountry":"US"},
 "numberOfEmployees":{"value":1279,"@type":"QuantitativeValue"},
 "logo":{"contentUrl":"https://media.licdn.com/logo.jpg","@type":"ImageObject"}}
]}
</script></head><body></body></html>`

func TestParseCompany(t *testing.T) {
	doc := docOf(t, companyHTML)
	c, err := ParseCompany(doc, "xsolla", "https://www.linkedin.com/company/xsolla")
	if err != nil {
		t.Fatalf("ParseCompany: %v", err)
	}
	if c.Name != "Xsolla" {
		t.Errorf("name = %q", c.Name)
	}
	if c.Employees != 1279 {
		t.Errorf("employees = %d", c.Employees)
	}
	if c.Website != "xsolla.com" {
		t.Errorf("website = %q", c.Website)
	}
	if c.Locality != "Sherman Oaks" || c.Region != "CA" || c.Country != "US" {
		t.Errorf("address = %q %q %q", c.Locality, c.Region, c.Country)
	}
	if c.LogoURL != "https://media.licdn.com/logo.jpg" {
		t.Errorf("logo = %q", c.LogoURL)
	}
}

func TestParseCompanyMissing(t *testing.T) {
	doc := docOf(t, `<html><body>no json-ld here</body></html>`)
	if _, err := ParseCompany(doc, "x", "https://www.linkedin.com/company/x"); err != ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}
