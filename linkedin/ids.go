package linkedin

import (
	"net/url"
	"regexp"
	"strings"
)

// Kinds of LinkedIn entities the tool understands.
const (
	KindProfile = "profile"
	KindCompany = "company"
	KindSchool  = "school"
	KindJob     = "job"
	KindPost    = "post"
	KindUnknown = "unknown"
)

var reJobURN = regexp.MustCompile(`urn:li:jobPosting:(\d+)`)
var reJobID = regexp.MustCompile(`/jobs/view/(?:[^/]*-)?(\d+)`)
var reDigits = regexp.MustCompile(`^\d+$`)

// segment returns the path segment after marker in a LinkedIn URL or path, with
// any trailing slash, query, or fragment trimmed.
func segment(s, marker string) string {
	i := strings.Index(s, marker)
	if i < 0 {
		return ""
	}
	rest := s[i+len(marker):]
	rest = strings.TrimSuffix(rest, "/")
	if j := strings.IndexAny(rest, "/?#"); j >= 0 {
		rest = rest[:j]
	}
	return rest
}

// slugFromCompanyURL pulls the slug out of a /company/<slug> URL.
func slugFromCompanyURL(u string) string { return segment(u, "/company/") }

// slugFromAffiliationURL pulls the slug out of either a /company/<slug> or a
// /school/<slug> URL, since a person's affiliations mix companies and schools.
func slugFromAffiliationURL(u string) string {
	switch {
	case strings.Contains(u, "/school/"):
		return segment(u, "/school/")
	case strings.Contains(u, "/showcase/"):
		return segment(u, "/showcase/")
	default:
		return slugFromCompanyURL(u)
	}
}

// slugFromProfileURL pulls the slug out of an /in/<slug> URL.
func slugFromProfileURL(u string) string { return segment(u, "/in/") }

// NormalizeProfileSlug accepts a slug, an /in/<slug> path, or a full URL and
// returns the bare slug.
func NormalizeProfileSlug(in string) string {
	in = strings.TrimSpace(in)
	if strings.Contains(in, "/in/") {
		return slugFromProfileURL(in)
	}
	return strings.Trim(in, "/")
}

// NormalizeCompanySlug accepts a slug, a /company/<slug> path, or a full URL and
// returns the bare slug.
func NormalizeCompanySlug(in string) string {
	in = strings.TrimSpace(in)
	if strings.Contains(in, "/company/") {
		return slugFromCompanyURL(in)
	}
	return strings.Trim(in, "/")
}

// NormalizeJobID accepts a numeric id, a urn:li:jobPosting URN, or a
// /jobs/view/... URL and returns the bare numeric id.
func NormalizeJobID(in string) string {
	in = strings.TrimSpace(in)
	if m := reJobURN.FindStringSubmatch(in); m != nil {
		return m[1]
	}
	if m := reJobID.FindStringSubmatch(in); m != nil {
		return m[1]
	}
	if reDigits.MatchString(in) {
		return in
	}
	// A trailing numeric token, as on "backend-engineer-go-4391940951".
	if i := strings.LastIndex(in, "-"); i >= 0 && reDigits.MatchString(in[i+1:]) {
		return in[i+1:]
	}
	return strings.Trim(in, "/")
}

// ProfileURL builds the canonical profile URL.
func ProfileURL(slug string) string { return BaseURL + "/in/" + NormalizeProfileSlug(slug) }

// CompanyURL builds the canonical company URL.
func CompanyURL(slug string) string { return BaseURL + "/company/" + NormalizeCompanySlug(slug) }

// SchoolURL builds the canonical school URL.
func SchoolURL(slug string) string { return BaseURL + "/school/" + strings.Trim(slug, "/") }

// JobURL builds the canonical job-view URL.
func JobURL(id string) string { return BaseURL + "/jobs/view/" + NormalizeJobID(id) }

// Classify inspects an input and returns its kind, id/slug, and canonical URL.
func Classify(input string) Ref {
	in := strings.TrimSpace(input)
	r := Ref{Input: input, Kind: KindUnknown}

	switch {
	case reJobURN.MatchString(in):
		id := NormalizeJobID(in)
		return Ref{Input: input, Kind: KindJob, ID: id, URL: JobURL(id)}
	case strings.Contains(in, "/jobs/view/"):
		id := NormalizeJobID(in)
		return Ref{Input: input, Kind: KindJob, ID: id, URL: JobURL(id)}
	case strings.Contains(in, "/in/"):
		s := slugFromProfileURL(in)
		return Ref{Input: input, Kind: KindProfile, ID: s, URL: ProfileURL(s)}
	case strings.Contains(in, "/company/"):
		s := slugFromCompanyURL(in)
		return Ref{Input: input, Kind: KindCompany, ID: s, URL: CompanyURL(s)}
	case strings.Contains(in, "/school/"):
		s := segment(in, "/school/")
		return Ref{Input: input, Kind: KindSchool, ID: s, URL: SchoolURL(s)}
	case strings.Contains(in, "/posts/") || strings.Contains(in, "/pulse/"):
		return Ref{Input: input, Kind: KindPost, ID: "", URL: canonicalURL(in)}
	case reDigits.MatchString(in):
		return Ref{Input: input, Kind: KindJob, ID: in, URL: JobURL(in)}
	}
	return r
}

// canonicalURL trims tracking query parameters from a LinkedIn URL.
func canonicalURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}
