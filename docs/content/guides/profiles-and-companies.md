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
headline, location, current and past positions, education, and the canonical
URL. Add `--save` to upsert each profile into the local store:

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

A company record carries the Organization JSON-LD fields: the name, description,
industry, headquarters, website, employee count, and the canonical URL. Add
`--posts` to also collect the company's recent public posts (the
DiscussionForumPosting nodes on the page):

```bash
linkedin company microsoft --posts
```

`--save` upserts each company into the store:

```bash
linkedin company microsoft --save
```

## The wall caveat

`profile` works for many public members, but some are walled behind LinkedIn's
sign-in wall, and `company` works for public company pages. When a page is
walled, linkedin exits with code 5 ("blocked") rather than returning the wall as
if it were data. HTTP 999 is LinkedIn's bot block; an authwall is a redirect to
`/authwall`, `/uas/login`, `/login`, `/checkpoint`, or `/signup`.

The fix is to lend the request a real session. Export a Netscape `cookies.txt`
jar from a signed-in browser and pass it:

```bash
linkedin profile some-walled-member --cookies ~/cookies.txt
```

Slowing down helps too (the default `--delay` is already two seconds). See
[troubleshooting](/reference/troubleshooting/) for the cookie file format and
more.
