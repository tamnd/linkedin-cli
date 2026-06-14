package linkedin

import "testing"

// A trimmed copy of a guest search result card.
const jobCardHTML = `<ul>
<li>
<div class="base-card relative job-search-card" data-entity-urn="urn:li:jobPosting:4391940951">
<a class="base-card__full-link" href="https://ca.linkedin.com/jobs/view/backend-engineer-go-at-xsolla-4391940951?position=1"><span class="sr-only">Backend Engineer (Go)</span></a>
<div class="base-search-card__info">
<h3 class="base-search-card__title">Backend Engineer (Go)</h3>
<h4 class="base-search-card__subtitle"><a class="hidden-nested-link" href="https://www.linkedin.com/company/xsolla?trk=public_jobs">Xsolla</a></h4>
<div class="base-search-card__metadata">
<span class="job-search-card__location">Montreal, Quebec, Canada</span>
<time class="job-search-card__listdate" datetime="2025-12-01">2 weeks ago</time>
</div></div></div>
</li>
</ul>`

func TestParseJobCards(t *testing.T) {
	doc := docOf(t, jobCardHTML)
	stubs := parseJobCards(doc)
	if len(stubs) != 1 {
		t.Fatalf("got %d stubs", len(stubs))
	}
	s := stubs[0]
	if s.JobID != "4391940951" {
		t.Errorf("job id = %q", s.JobID)
	}
	if s.URL != "https://www.linkedin.com/jobs/view/4391940951" {
		t.Errorf("url = %q", s.URL)
	}
	if s.Title != "Backend Engineer (Go)" {
		t.Errorf("title = %q", s.Title)
	}
	if s.Company != "Xsolla" || s.CompanySlug != "xsolla" {
		t.Errorf("company = %q slug = %q", s.Company, s.CompanySlug)
	}
	if s.Location != "Montreal, Quebec, Canada" {
		t.Errorf("location = %q", s.Location)
	}
	if s.Posted != "2025-12-01" {
		t.Errorf("posted = %q", s.Posted)
	}
}

// A trimmed copy of the guest job-detail fragment.
const jobDetailHTML = `<section>
<h2 class="topcard__title">Backend Engineer (Go)</h2>
<a class="topcard__org-name-link" href="https://www.linkedin.com/company/xsolla?trk=x">Xsolla</a>
<span class="topcard__flavor topcard__flavor--bullet">Montreal, Quebec, Canada</span>
<span class="num-applicants__caption">91 applicants</span>
<span class="posted-time-ago__text">2 weeks ago</span>
<div class="show-more-less-html__markup">We are hiring a Go engineer.</div>
<ul class="description__job-criteria-list">
<li class="description__job-criteria-item"><h3 class="description__job-criteria-subheader">Seniority level</h3><span class="description__job-criteria-text">Mid-Senior level</span></li>
<li class="description__job-criteria-item"><h3 class="description__job-criteria-subheader">Employment type</h3><span class="description__job-criteria-text">Full-time</span></li>
<li class="description__job-criteria-item"><h3 class="description__job-criteria-subheader">Job function</h3><span class="description__job-criteria-text">Engineering</span></li>
<li class="description__job-criteria-item"><h3 class="description__job-criteria-subheader">Industries</h3><span class="description__job-criteria-text">Computer Games</span></li>
</ul></section>`

func TestParseJobDetail(t *testing.T) {
	doc := docOf(t, jobDetailHTML)
	j, err := parseJobDetail(doc, "4391940951")
	if err != nil {
		t.Fatalf("parseJobDetail: %v", err)
	}
	if j.Title != "Backend Engineer (Go)" {
		t.Errorf("title = %q", j.Title)
	}
	if j.Company != "Xsolla" || j.CompanySlug != "xsolla" {
		t.Errorf("company = %q slug = %q", j.Company, j.CompanySlug)
	}
	if j.Location != "Montreal, Quebec, Canada" {
		t.Errorf("location = %q", j.Location)
	}
	if j.Applicants != 91 {
		t.Errorf("applicants = %d", j.Applicants)
	}
	if j.Seniority != "Mid-Senior level" || j.EmploymentType != "Full-time" {
		t.Errorf("criteria seniority=%q employment=%q", j.Seniority, j.EmploymentType)
	}
	if j.JobFunction != "Engineering" || j.Industries != "Computer Games" {
		t.Errorf("criteria function=%q industries=%q", j.JobFunction, j.Industries)
	}
	if j.Description != "We are hiring a Go engineer." {
		t.Errorf("description = %q", j.Description)
	}
}

func TestSearchURL(t *testing.T) {
	o := JobSearchOptions{Keywords: "golang", Location: "Remote", Posted: "r604800", Remote: "2"}
	got := o.searchURL(25)
	for _, want := range []string{"keywords=golang", "location=Remote", "f_TPR=r604800", "f_WT=2", "start=25"} {
		if !contains(got, want) {
			t.Errorf("search url %q missing %q", got, want)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
