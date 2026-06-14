package linkedin

import "time"

// Profile is a public member profile, parsed from the page's Person JSON-LD.
type Profile struct {
	Slug        string        `json:"slug"`
	URL         string        `json:"url"`
	Name        string        `json:"name"`
	Headline    string        `json:"headline"`
	Description string        `json:"description"`
	JobTitles   []string      `json:"job_titles"`
	Location    string        `json:"location"`
	Country     string        `json:"country"`
	Followers   int64         `json:"followers"`
	Connections int64         `json:"connections"`
	ImageURL    string        `json:"image_url"`
	Languages   []string      `json:"languages"`
	Awards      []string      `json:"awards"`
	WorksFor    []Affiliation `json:"works_for"`
	AlumniOf    []Affiliation `json:"alumni_of"`
	SameAs      []string      `json:"same_as"`
	FetchedAt   time.Time     `json:"fetched_at"`
}

// Affiliation is a company or school a person works for or studied at, with the
// role's year range when LinkedIn exposes it.
type Affiliation struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Slug      string `json:"slug"`
	StartDate int    `json:"start_date"`
	EndDate   int    `json:"end_date"`
}

// Company is a public company page, parsed from the Organization JSON-LD.
type Company struct {
	Slug        string    `json:"slug"`
	URL         string    `json:"url"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Slogan      string    `json:"slogan"`
	Website     string    `json:"website"`
	Employees   int64     `json:"employees"`
	Street      string    `json:"street"`
	Locality    string    `json:"locality"`
	Region      string    `json:"region"`
	PostalCode  string    `json:"postal_code"`
	Country     string    `json:"country"`
	LogoURL     string    `json:"logo_url"`
	FetchedAt   time.Time `json:"fetched_at"`
}

// Job is a single job posting, parsed from the guest job-detail fragment.
type Job struct {
	JobID          string    `json:"job_id"`
	URL            string    `json:"url"`
	Title          string    `json:"title"`
	Company        string    `json:"company"`
	CompanyURL     string    `json:"company_url"`
	CompanySlug    string    `json:"company_slug"`
	Location       string    `json:"location"`
	Posted         string    `json:"posted"`
	Applicants     int       `json:"applicants"`
	Seniority      string    `json:"seniority"`
	EmploymentType string    `json:"employment_type"`
	JobFunction    string    `json:"job_function"`
	Industries     string    `json:"industries"`
	Description    string    `json:"description"`
	ApplyURL       string    `json:"apply_url"`
	FetchedAt      time.Time `json:"fetched_at"`
}

// JobStub is one job card from the guest search endpoint: enough to list and to
// fetch the full Job.
type JobStub struct {
	JobID       string    `json:"job_id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	CompanyURL  string    `json:"company_url"`
	CompanySlug string    `json:"company_slug"`
	Location    string    `json:"location"`
	Posted      string    `json:"posted"`
	Benefits    string    `json:"benefits"`
	FetchedAt   time.Time `json:"fetched_at"`
}

// Post is a best-effort public post or article, parsed from Open Graph tags and
// any Article/DiscussionForumPosting JSON-LD that serves anonymously.
type Post struct {
	URL       string    `json:"url"`
	Author    string    `json:"author"`
	AuthorURL string    `json:"author_url"`
	Title     string    `json:"title"`
	Text      string    `json:"text"`
	Published string    `json:"published"`
	Likes     int64     `json:"likes"`
	Comments  int64     `json:"comments"`
	ImageURL  string    `json:"image_url"`
	FetchedAt time.Time `json:"fetched_at"`
}

// Ref is the classification of a user-supplied input (the `id` command output).
type Ref struct {
	Input string `json:"input"`
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	URL   string `json:"url"`
}
