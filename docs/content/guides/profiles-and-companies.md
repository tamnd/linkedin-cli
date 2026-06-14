---
title: "Profiles and companies"
description: "Fetch public member profiles and company pages from their JSON-LD, with the wall caveat."
weight: 10
---

The two record lookups: a member profile and a company page. Each reads the
page's JSON-LD first (schema.org `Person` on a profile, `Organization` on a
company), with HTML selectors as a fallback. Both accept a slug, a path, or a
full URL.

## A profile

```bash
linkedin profile williamhgates
```

`profile` accepts a bare slug, an `/in/<slug>` path, or a full URL, and takes
several at once:

```bash
linkedin profile williamhgates --format json
linkedin profile williamhgates /in/satyanadella --format csv
linkedin profile https://www.linkedin.com/in/williamhgates
```

A profile record carries the fields the Person JSON-LD exposes: the name,
headline, location, current and past positions, education, `member_of` (boards
and groups, an array of affiliations with name, url, slug, start_date, and
end_date), and the canonical URL. There is no connection count: LinkedIn never
exposes one anonymously, so the record has no such field.

Add `--posts` to also emit the member's recent posts (the `DiscussionForumPosting`
nodes in the page's JSON-LD `@graph`) as Post records, or `--articles` for the
member's long-form articles (the `Article` nodes in the same graph) as Article
records. If you pass both, `--posts` wins:

```bash
linkedin profile williamhgates --posts
linkedin profile williamhgates --articles --format json
```

Add `--save` to upsert each profile into the local store:

```bash
linkedin profile williamhgates --save
```

## A company

```bash
linkedin company microsoft
```

Like `profile`, it accepts a slug or a URL and takes several at once:

```bash
linkedin company microsoft github --format csv
```

A company record carries the Organization JSON-LD fields plus the company about
panel, which the JSON-LD leaves out. From the JSON-LD: the name, description,
website, `employees` (a point estimate from numberOfEmployees), and the canonical
URL. From the about panel (a list of label and value pairs) and the og:description:
`followers`, `industry`, `company_size` (the UI band like "10,001+ employees", a
separate number from `employees`), `company_type` (for example "Public Company"),
`founded` (the year, when the company lists it; some like Microsoft omit it),
`specialties` (a comma list), and `headquarters` (for example "Sherman Oaks, CA").
When the company lists funding, the record also carries `funding_rounds` and a
`funding_url` Crunchbase link to the latest round.

For example `xsolla` lists `founded` 2005 and `headquarters` "Sherman Oaks, CA",
while `microsoft` omits the founded year:

```bash
linkedin company xsolla --fields name,founded,headquarters,followers
```

Add `--posts` to instead collect the company's recent public posts (the
`DiscussionForumPosting` nodes in the page's JSON-LD graph; the dedicated
`/posts/` subpage is login-walled, so they come from the main page):

```bash
linkedin company microsoft --posts
```

`--locations` emits the full office list, one record per office, with `primary`
marking the registered headquarters and `address` holding the city, region,
postal code, and country line. `--affiliated` emits the related affiliated and
showcase pages, each with its slug, name, industry, and location:

```bash
linkedin company microsoft --locations
linkedin company microsoft --affiliated
```

`--save` upserts each company into the store:

```bash
linkedin company microsoft --save
```

## The wall caveat

`profile` and `company` both return 200 and work reliably. linkedin sends no
`Referer` header, which is what avoids LinkedIn's HTTP 999 bot block on these
pages. So you should not normally see a wall here. If you do see exit code 5
("blocked") on one of these, it usually means IP-level rate-limiting rather than
the page itself being walled.

Slowing down is the first fix (the default `--delay` is already two seconds; raise
it). Lending the request a real session helps too: export a Netscape `cookies.txt`
jar from a signed-in browser and pass it:

```bash
linkedin profile williamhgates --delay 5s
linkedin profile williamhgates --cookies ~/cookies.txt
```

See [troubleshooting](/reference/troubleshooting/) for the cookie file format and
more.
