package linkedin

import "testing"

func TestClassify(t *testing.T) {
	tests := []struct {
		in   string
		kind string
		id   string
		url  string
	}{
		{"https://www.linkedin.com/in/williamhgates", KindProfile, "williamhgates", "https://www.linkedin.com/in/williamhgates"},
		{"https://www.linkedin.com/company/microsoft/about/", KindCompany, "microsoft", "https://www.linkedin.com/company/microsoft"},
		{"https://www.linkedin.com/school/harvard-university/", KindSchool, "harvard-university", "https://www.linkedin.com/school/harvard-university"},
		{"urn:li:jobPosting:4391940951", KindJob, "4391940951", "https://www.linkedin.com/jobs/view/4391940951"},
		{"https://ca.linkedin.com/jobs/view/backend-engineer-go-at-xsolla-4391940951?x=1", KindJob, "4391940951", "https://www.linkedin.com/jobs/view/4391940951"},
		{"4391940951", KindJob, "4391940951", "https://www.linkedin.com/jobs/view/4391940951"},
		{"https://www.linkedin.com/posts/williamhgates_activity-123", KindPost, "", "https://www.linkedin.com/posts/williamhgates_activity-123"},
		{"some random text", KindUnknown, "", ""},
	}
	for _, tc := range tests {
		got := Classify(tc.in)
		if got.Kind != tc.kind || got.ID != tc.id || got.URL != tc.url {
			t.Errorf("Classify(%q) = {%s %s %s}, want {%s %s %s}", tc.in, got.Kind, got.ID, got.URL, tc.kind, tc.id, tc.url)
		}
	}
}

func TestNormalizeSlugs(t *testing.T) {
	if s := NormalizeProfileSlug("https://www.linkedin.com/in/williamhgates/"); s != "williamhgates" {
		t.Errorf("profile slug = %q", s)
	}
	if s := NormalizeCompanySlug("/company/microsoft"); s != "microsoft" {
		t.Errorf("company slug = %q", s)
	}
	if id := NormalizeJobID("backend-engineer-go-at-xsolla-4391940951"); id != "4391940951" {
		t.Errorf("job id = %q", id)
	}
}
