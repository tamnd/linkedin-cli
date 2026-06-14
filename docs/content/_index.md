---
title: "linkedin"
description: "A command line for public LinkedIn data. Fetch a member profile, a company page, or a job posting, search jobs, and build datasets, all from one binary."
heroTitle: "LinkedIn, from the command line"
heroLead: "linkedin is a single pure-Go binary that turns public LinkedIn pages into clean, structured records. Fetch profiles and company pages, read a job posting, search the jobs board, and save records to a local store, with no API key and nothing to sign up for."
heroPrimaryURL: "/getting-started/quick-start/"
heroPrimaryText: "Get started"
---

LinkedIn has no open public API for this, so getting structured data out of it
usually means writing a scraper. linkedin does that part for you: it reads a
public page and turns it into a record with real fields, JSON-LD first and an
HTML fallback when a page does not carry it.

```bash
linkedin profile williamhgates              # a public member profile as a record
linkedin company microsoft                  # a company page from its Organization JSON-LD
linkedin jobs "golang engineer" -n 25       # the jobs board via the guest endpoint
linkedin job 3801234567                     # one job posting in full
```

It talks to `www.linkedin.com` over plain HTTPS with no API key. The binary is
pure Go with no runtime dependencies. Output is a table on a terminal and JSONL
when piped, with json, csv, tsv, url, raw, `--fields`, and `--template` when you
want something else.

## What you can do with it

- **Fetch profiles and companies.** `linkedin profile` reads a public member
  profile from the page's Person JSON-LD, with `--posts` and `--articles` to also
  emit the member's recent posts and long-form articles. `linkedin company` reads
  a company page from its Organization JSON-LD and the about panel, with `--posts`
  to also collect recent public company posts.
- **Read and search jobs.** `linkedin job` fetches a single posting in full, and
  `linkedin jobs` searches the board through the anonymous guest endpoint, with
  filters for location, posting age, remote mode, experience, and job type.
- **Classify input.** `linkedin id` turns a slug, path, or URL into a (kind, id)
  pair without fetching, and `linkedin url` builds a canonical URL from a kind
  and an id.
- **Build a dataset.** `--save` upserts records into a local SQLite store across
  calls, and `db query` reads them back. `cache` manages the on-disk page cache.
- **Script it.** Stable exit codes, a `--template` for any line shape, and the
  local `id` command make linkedin safe to drop into a pipeline.

## Honest about what is walled

LinkedIn serves some surfaces to anonymous visitors and walls the rest behind a
sign-in wall. Profiles, company pages, and job detail all return 200 and work
reliably, and jobs search works through the guest endpoint. Single public posts
and articles generally return data, best effort, through JSON-LD with an Open
Graph backstop. What is still walled: school pages return LinkedIn's bot block
(HTTP 999), the dedicated activity and `/posts/` subpages of profiles and
companies redirect to a login (which is why posts come from the JSON-LD graph on
the main page instead), and people and company search require sign-in. When a
page is walled, linkedin exits with code 5 and you can lend a session with
`--cookies` (a Netscape cookies.txt jar).

## Independent and public-data only

linkedin is an independent, open-source tool. It is not affiliated with,
endorsed by, or sponsored by LinkedIn or Microsoft. It reads only public pages,
at a polite default rate (a two second gap between requests, two workers).

## Where to go next

- New here? Start with the [introduction](/getting-started/introduction/) for
  the mental model, then the [quick start](/getting-started/quick-start/).
- Want to install it? See [installation](/getting-started/installation/).
- Looking for a specific task? The [guides](/guides/) cover profiles and
  companies, finding jobs, posts and lookups, storing records, and output
  formats.
- Need every flag? The [CLI reference](/reference/cli/) is the full surface.
