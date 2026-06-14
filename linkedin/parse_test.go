package linkedin

import (
	"strings"
	"testing"
	"time"

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
 "memberOf":[{"@type":"Organization","name":"Breakthrough Energy","url":"https://www.linkedin.com/company/breakthrough-energy"}],
 "interactionStatistic":{"@type":"InteractionCounter","interactionType":"https://schema.org/FollowAction","userInteractionCount":40419217}},
{"@type":"DiscussionForumPosting","url":"https://www.linkedin.com/posts/williamhgates_post-1","headline":"On clean energy","text":"A note on energy.","datePublished":"2026-02-01","author":{"@type":"Person","name":"Bill Gates","url":"https://www.linkedin.com/in/williamhgates"},"interactionStatistic":[{"@type":"InteractionCounter","interactionType":"https://schema.org/LikeAction","userInteractionCount":120},{"@type":"InteractionCounter","interactionType":"https://schema.org/CommentAction","userInteractionCount":7}]},
{"@type":"Article","url":"https://www.linkedin.com/pulse/the-year-ahead-williamhgates","headline":"The Year Ahead","datePublished":"2026-01-15","author":{"@type":"Person","name":"Bill Gates","url":"https://www.linkedin.com/in/williamhgates"},"image":"https://media.licdn.com/article.jpg","interactionStatistic":[{"@type":"InteractionCounter","interactionType":"https://schema.org/LikeAction","userInteractionCount":3400}]}
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
	if len(p.AlumniOf) != 1 || p.AlumniOf[0].EndDate != 1975 || p.AlumniOf[0].Slug != "harvard-university" {
		t.Errorf("alumni = %+v", p.AlumniOf)
	}
	if p.ImageURL != "https://media.licdn.com/photo.jpg" {
		t.Errorf("image = %q", p.ImageURL)
	}
	if len(p.MemberOf) != 1 || p.MemberOf[0].Slug != "breakthrough-energy" {
		t.Errorf("member of = %+v", p.MemberOf)
	}
}

func TestPostsFromGraph(t *testing.T) {
	doc := docOf(t, profileHTML)
	posts := postsFromGraph(doc, "https://www.linkedin.com/in/williamhgates")
	if len(posts) != 1 {
		t.Fatalf("posts = %d, want 1", len(posts))
	}
	got := posts[0]
	if got.Title != "On clean energy" {
		t.Errorf("title = %q", got.Title)
	}
	if got.Likes != 120 || got.Comments != 7 {
		t.Errorf("likes = %d comments = %d", got.Likes, got.Comments)
	}
	if got.Author != "Bill Gates" {
		t.Errorf("author = %q", got.Author)
	}
}

func TestArticlesFromGraph(t *testing.T) {
	doc := docOf(t, profileHTML)
	var now time.Time
	var arts []Article
	for _, raw := range nodesOfType(ldNodes(doc), "Article") {
		if a, ok := articleFromNode(raw, now); ok {
			arts = append(arts, a)
		}
	}
	if len(arts) != 1 {
		t.Fatalf("articles = %d, want 1", len(arts))
	}
	a := arts[0]
	if a.Title != "The Year Ahead" {
		t.Errorf("title = %q", a.Title)
	}
	if a.Reactions != 3400 {
		t.Errorf("reactions = %d", a.Reactions)
	}
	if a.ImageURL != "https://media.licdn.com/article.jpg" {
		t.Errorf("image = %q", a.ImageURL)
	}
}

const companyHTML = `<html><head>
<meta property="og:description" content="Xsolla | 1,234,567 followers on LinkedIn. Sell more games.">
<script type="application/ld+json">
{"@context":"http://schema.org","@graph":[
{"@type":"DiscussionForumPosting","url":"https://www.linkedin.com/posts/x","headline":"Update","text":"hi","datePublished":"2026-01-01"},
{"@type":"Organization","name":"Xsolla","url":"https://www.linkedin.com/company/xsolla",
 "description":"Video game business engine.","slogan":"Sell more games.",
 "sameAs":"xsolla.com",
 "address":{"streetAddress":"15260 Ventura Blvd","addressLocality":"Sherman Oaks","addressRegion":"CA","postalCode":"91403","addressCountry":"US"},
 "numberOfEmployees":{"value":1279,"@type":"QuantitativeValue"},
 "logo":{"contentUrl":"https://media.licdn.com/logo.jpg","@type":"ImageObject"}}
]}
</script></head><body>
<dl>
<dt>Industry</dt><dd>Computer Games</dd>
<dt>Company size</dt><dd>1,001-5,000 employees</dd>
<dt>Type</dt><dd>Privately Held</dd>
<dt>Headquarters</dt><dd>Sherman Oaks, CA</dd>
<dt>Founded</dt><dd>2005</dd>
<dt>Specialties</dt><dd>game monetization, payments</dd>
<dt>Website</dt><dd>https://xsolla.com External link for Xsolla</dd>
</dl>
</body></html>`

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
	if c.Followers != 1234567 {
		t.Errorf("followers = %d", c.Followers)
	}
	if c.Industry != "Computer Games" {
		t.Errorf("industry = %q", c.Industry)
	}
	if c.CompanySize != "1,001-5,000 employees" {
		t.Errorf("company size = %q", c.CompanySize)
	}
	if c.CompanyType != "Privately Held" {
		t.Errorf("company type = %q", c.CompanyType)
	}
	if c.Headquarters != "Sherman Oaks, CA" {
		t.Errorf("headquarters = %q", c.Headquarters)
	}
	if c.Founded != "2005" {
		t.Errorf("founded = %q", c.Founded)
	}
	if c.Specialties != "game monetization, payments" {
		t.Errorf("specialties = %q", c.Specialties)
	}
}

func TestFirstField(t *testing.T) {
	got := firstField("https://xsolla.com External link for Xsolla")
	if got != "https://xsolla.com" {
		t.Errorf("firstField = %q", got)
	}
}

func TestParseCompanyMissing(t *testing.T) {
	doc := docOf(t, `<html><body>no json-ld here</body></html>`)
	if _, err := ParseCompany(doc, "x", "https://www.linkedin.com/company/x"); err != ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}
