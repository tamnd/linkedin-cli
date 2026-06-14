---
title: "Introduction"
description: "What linkedin reads, how it turns a page into a record, and the sign-in wall it works around."
weight: 10
---

[LinkedIn](https://www.linkedin.com) is a large network of public member
profiles, company pages, job postings, and posts. There is no open public API to
read this, so the only way to get it programmatically is to fetch a page and
parse it.

linkedin does that part. It is a single binary that fetches a public LinkedIn
page and turns it into a structured record. You ask for a profile, a company, or
a job, and it hands you fields, not HTML.

## From a page to a record

Most LinkedIn pages carry a [JSON-LD](https://json-ld.org) block: a chunk of
structured data the page ships for search engines. linkedin reads that first,
because it is the cleanest source on the page. Profiles carry a schema.org
`Person`, company pages carry an `Organization`, and a profile's recent posts and
articles ride along in the same `@graph` as `DiscussionForumPosting` and `Article`
nodes. When a page does not carry the field it needs, linkedin falls back to
reading the HTML with CSS selectors, for example the company about panel that the
Organization JSON-LD leaves out. The result either way is a record with real
fields.

Jobs are different. The jobs board and job detail come from LinkedIn's guest
endpoints (`/jobs-guest/jobs/api/seeMoreJobPostings/search` and
`/jobs-guest/jobs/api/jobPosting/<id>`), which serve anonymous visitors. That is
the reliable path for job data.

## The sign-in wall, and what works around it

Here is the honest part. LinkedIn serves some surfaces to anonymous visitors and
walls the rest behind a sign-in wall. The key detail: linkedin sends no `Referer`
header. A same-site referer is one of the signals LinkedIn reads as scraping and
answers with HTTP 999, so leaving it off is what makes profile and company pages
return 200. What works anonymously:

- **`profile`** reads public member profiles from the Person JSON-LD, and with
  `--posts` or `--articles` emits the recent posts and articles from the same
  graph.
- **`company`** reads company pages from the Organization JSON-LD plus the about
  panel in the HTML.
- **`job`** reads a single posting from the guest job-detail fragment.
- **`jobs`** searches the board through the anonymous guest endpoint.
- **`post`** reads single public posts and articles best effort, JSON-LD first
  with an Open Graph backstop, and generally returns data.

What is still walled:

- **School pages** (`/school/<slug>`) return LinkedIn's bot block (HTTP 999).
  There is no `school` fetch command; `id` only classifies a school URL.
- **The activity and `/posts/` subpages** of profiles and companies (a profile's
  `/recent-activity/`, a company's `/posts/`) redirect to `/uas/login`. That is
  why posts come from the JSON-LD graph on the main page, not from those subpages.
- **People and company search**, typeahead, and `/search/results/*` require
  sign-in.

When a page is walled, linkedin exits with code 5 ("blocked") rather than
pretending it got data. HTTP 999 is LinkedIn's bot block; an authwall shows up as
a redirect to `/authwall`, `/uas/login`, `/login`, `/checkpoint`, or `/signup`.
The hint suggests passing `--cookies`: a Netscape `cookies.txt` jar exported from
a signed-in browser session, which lends the request a real session and often
gets through.

## Polite by default

linkedin waits two seconds between requests and runs two workers by default, so
a busy session stays a good citizen against a public site. You can tune
`--delay` and `--workers`, but the defaults are deliberately gentle.

## Independent and public-data only

linkedin is an independent, open-source tool. It is not affiliated with,
endorsed by, or sponsored by LinkedIn or Microsoft. It reads only public pages,
at a polite default rate. It does not log in for you, store your credentials, or
touch anything behind an account.

Next: [install it](/getting-started/installation/), then take the
[quick start](/getting-started/quick-start/).
