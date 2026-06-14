---
title: "CLI"
description: "Every command and subcommand, with the flags that matter and one example each."
weight: 10
---

```
linkedin <command> [args] [flags]
```

Run `linkedin <command> --help` for the full flag list on any command. This page
is the map. `profile`, `company`, `job`, and `jobs` work for anonymous visitors,
and `post` is best effort but generally returns data for single public posts and
articles. When a page is behind the sign-in wall, linkedin exits with code 4. See
[troubleshooting](/reference/troubleshooting/).

## Commands

| Command | What it does |
|---|---|
| `profile` | Fetch public member profiles from the Person JSON-LD |
| `company` | Fetch company pages from the Organization JSON-LD |
| `job` | Fetch a job posting from the guest job-detail fragment |
| `jobs` | Search the jobs board through the anonymous guest endpoint |
| `post` | Fetch a single public post or article (best effort) |
| `id` | Classify and normalize an input into (kind, id) without fetching |
| `url` | Build a canonical LinkedIn URL from a kind and an id |
| `db` | Inspect the local SQLite record store |
| `cache` | Inspect and clear the on-disk page cache |
| `info` | Print the resolved configuration |
| `version` | Print version, commit, and build date |

## profile

```
linkedin profile <slug|url> [slug|url ...] [flags]
```

Fetches one or more public member profiles, parsed from the page's Person
JSON-LD. Accepts a slug (`williamhgates`), an `/in/<slug>` path, or a full URL.
Profile pages return 200 and work; an exit 4 here usually means IP-level
rate-limiting (slow down with `--rate` or lend `--cookies`).

Fields include the name, headline, location, country, followers, current and past
positions (`works_for`, `alumni_of`), `member_of` (boards and groups, an array of
affiliations with name, url, slug, start_date, end_date), and the canonical URL.
There is no connection count; LinkedIn does not expose one anonymously.

| Flag | Meaning |
|---|---|
| `--posts` | Also emit the member's recent posts (the `DiscussionForumPosting` nodes in the page's JSON-LD `@graph`) as Post records |
| `--articles` | Also emit the member's long-form articles (the `Article` nodes in the same graph) as Article records |
| `--save` | Upsert each profile into the store |

If both `--posts` and `--articles` are given, `--posts` wins.

```bash
linkedin profile williamhgates --output json
linkedin profile williamhgates --posts
```

## company

```
linkedin company <slug|url> [slug|url ...] [flags]
```

Fetches one or more company pages from the Organization JSON-LD plus the company
about panel. Fields from the JSON-LD: name, description, website, `employees` (a
point estimate from numberOfEmployees), and the canonical URL. Fields from the
about panel and og:description: `followers`, `industry`, `company_size` (the UI
band like "10,001+ employees", separate from `employees`), `company_type` (for
example "Public Company"), `founded` (the year, when the company lists it; some
like Microsoft omit it), `specialties` (a comma list), and `headquarters` (for
example "Sherman Oaks, CA"). When the company lists funding, the record also
carries `funding_rounds` (the round count) and `funding_url` (a Crunchbase link
to the latest round).

`--posts` instead collects the company's recent public posts (the
`DiscussionForumPosting` nodes in the main page's JSON-LD graph; the `/posts/`
subpage is login-walled). `--locations` emits the full office list, one record
per office, with `primary` marking the registered headquarters and `address`
holding the city, region, postal code, and country line. `--affiliated` emits the
related affiliated and showcase pages, each with its `slug`, `name`, `industry`,
and `location`. `--save` upserts each company into the store.

```bash
linkedin company microsoft --posts
linkedin company microsoft --locations
linkedin company microsoft --affiliated
linkedin company xsolla --fields name,founded,headquarters,followers
```

## job

```
linkedin job <id|url> [id|url ...]
```

Fetches one or more job postings from the guest job-detail fragment. Fields
include the title, company, `company_logo`, location, applicant count, posting
date, full description, and criteria (seniority, employment type, job function,
industries).

```bash
linkedin job 3801234567 --output json
```

## jobs

```
linkedin jobs <keywords...> [flags]
```

Searches the jobs board through the anonymous guest endpoint, paginating in pages
of 25 until `-n` results or the endpoint runs dry. Emits JobStub records by
default (title, company, `company_logo`, location, id, URL), or full Job records
with `--hydrate`.

| Flag | Meaning |
|---|---|
| `--location` | Free-text location (e.g. `Remote`, `"United States"`, a city) |
| `--geo-id` | LinkedIn geo id, when you know it |
| `--posted` | Posting age: `r86400` (24h), `r604800` (week), `r2592000` (month) |
| `--remote` | Workplace: `1` on-site, `2` remote, `3` hybrid |
| `--experience` | Experience level, `1`..`6` |
| `--job-type` | Type: `F`, `P`, `C`, `T`, `I`, `V`, `O` |
| `--sort` | `R` relevance, `DD` date |
| `--hydrate` | Follow each stub to a full job record |
| `--save` | With `--hydrate`, upsert each job into the store |

```bash
linkedin jobs "golang engineer" --remote 2 --posted r604800 -n 50
```

## post

```
linkedin post <url> [url ...]
```

Fetches a single public post or article, best effort: JSON-LD first (the
`DiscussionForumPosting` or `Article` node) with an Open Graph backstop. Single
public posts and articles generally return data; when one is walled, linkedin
exits 4.

```bash
linkedin post https://www.linkedin.com/posts/example-activity-123456789
```

## id

```
linkedin id <input> [input ...]
```

Classifies and normalizes each argument into a (kind, id) pair without fetching.
Kinds are `profile`, `company`, `school`, `job`, `post`, and `unknown`. Pure
local work, never blocked.

```bash
linkedin id https://www.linkedin.com/in/williamhgates
```

## url

```
linkedin url <kind> <id>
```

Builds a canonical LinkedIn URL from a kind and an id.

```bash
linkedin url profile williamhgates
```

## db

| Subcommand | Does |
|---|---|
| `db path` | Print the store file path |
| `db count` | Count stored records |
| `db query` | Read stored records back out |

```bash
linkedin db count
```

## cache

| Subcommand | Does |
|---|---|
| `cache path` | Print the cache directory path |
| `cache info` | Location, file count, and size |
| `cache clear` | Remove every cached page |

```bash
linkedin cache info
```

## Meta

| Command | Does |
|---|---|
| `info` | Print the resolved configuration |
| `version` | Print version, commit, and build date |

```bash
linkedin info
```

## Global flags

These apply to every command. See [configuration](/reference/configuration/) for
the full list and their defaults.

| Flag | Meaning |
|---|---|
| `-o, --output` | Output format, `auto`/`list`/`table`/`markdown`/`json`/`jsonl`/`csv`/`tsv`/`url`/`raw` (default `auto`: list on a TTY, jsonl piped) |
| `--fields` | Comma-separated columns to include |
| `--no-header` | Omit the header row in table/csv/tsv output |
| `--template` | Go text/template applied per record |
| `--color` | `auto`, `always`, or `never` |
| `-n, --limit` | Limit number of results (`0` means no limit) |
| `-q, --quiet` | Suppress progress on stderr |
| `-v, --verbose` | Increase verbosity (repeatable) |
| `--rate` | Minimum spacing between requests (default 2s) |
| `--timeout` | Per-request timeout (default 30s) |
| `--retries` | Retry attempts on 429/5xx (default 3) |
| `--cache-ttl` | On-disk cache freshness window (default 24h) |
| `--no-cache` | Bypass the on-disk page cache for this run |
| `--refresh` | Force a re-fetch and overwrite the cache |
| `--data-dir` | Root dir for cache and store (env `LINKEDIN_DATA_DIR`) |
| `--cookies` | Netscape cookie jar to lend a session |
| `--dry-run` | Print actions, do not perform them |
| `--profile` | Named profile to load |
