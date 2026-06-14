---
title: "Quick start"
description: "From an empty terminal to a real job search, a real profile, and a real company record, in a handful of commands."
weight: 30
---

This walks the core loop: search the jobs board, fetch a job in full, look up a
profile, and read a company page. The jobs commands hit the guest endpoint, so
they are quick and reliable, and profile and company pages return 200 and work.

## 1. Search the jobs board

```bash
linkedin jobs "golang engineer" -n 25
```

Each row is a job stub from the anonymous guest endpoint. Narrow it with the
filters, for example remote jobs posted in the last week:

```bash
linkedin jobs "golang engineer" --remote 2 --posted r604800 -n 25
```

Want full job records instead of thin stubs? Add `--hydrate` to follow each stub
to its detail page:

```bash
linkedin jobs "golang engineer" --hydrate -n 10 --format json
```

## 2. Fetch one job in full

```bash
linkedin job 3801234567
```

`job` takes a job id or a full URL and returns the title, company, location,
applicant count, posting date, full description, and criteria (seniority,
employment type, job function, industries). The same record as JSON:

```bash
linkedin job 3801234567 --format json
```

## 3. Look up a profile

`profile` takes a slug, an `/in/<slug>` path, or a full URL:

```bash
linkedin profile williamhgates
```

It reads the page's Person JSON-LD. Add `--posts` to also emit the member's
recent posts, or `--articles` for their long-form articles (if both are given,
`--posts` wins):

```bash
linkedin profile williamhgates --posts
```

Profile and company pages return 200 and work. If you do see exit code 5
("blocked") on a normally-working surface, it usually means IP-level
rate-limiting; slow down with `--delay` or lend a signed-in session with
`--cookies` (see [troubleshooting](/reference/troubleshooting/)).

## 4. Read a company page

```bash
linkedin company microsoft
```

`company` reads the Organization JSON-LD. Add `--posts` to also collect the
company's recent public posts:

```bash
linkedin company microsoft --posts
```

## 5. Classify without fetching

`id` turns a slug, path, or URL into a (kind, id) pair without touching the
network, which is handy in scripts:

```bash
linkedin id https://www.linkedin.com/in/williamhgates
```

```
profile	williamhgates
```

## 6. Compose

Output that pipes is the point. Pull the apply URLs off a job search:

```bash
linkedin jobs "golang engineer" --format jsonl | jq -r .url
```

Keep just a couple of fields off a company:

```bash
linkedin company microsoft --fields name,industry,employees
```

## Where to next

You have the core loop. From here:

- [Profiles and companies](/guides/profiles-and-companies/) covers `profile` and
  `company`, the JSON-LD fields, and the wall caveat.
- [Finding jobs](/guides/finding-jobs/) goes deep on `jobs` search, its filters,
  `--hydrate`, and `job` detail.
- [Posts and lookups](/guides/posts-and-lookups/) covers `post`, `id`, and `url`.
- [Storing records](/guides/storing-records/) covers `--save`, `db`, and `cache`.
- The [CLI reference](/reference/cli/) lists every command and flag.
