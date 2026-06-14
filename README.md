# linkedin

A command line for public [LinkedIn](https://www.linkedin.com) data. One binary
that turns member profiles, company pages, job postings, and the guest job
search into rich, structured records as a table, JSON, JSONL, CSV, TSV, or plain
URLs.

```
linkedin jobs "golang backend" --location Remote -n 5
```

```
JOB_ID      TITLE                       COMPANY    LOCATION                   POSTED
4391940951  Backend Engineer (Go)       Xsolla     Montreal, Quebec, Canada   2025-12-01
4388210773  Senior Go Engineer          Datadog    Remote                     2025-11-29
...
```

Full documentation: [linkedin-cli.tamnd.com](https://linkedin-cli.tamnd.com).

> linkedin is an independent, open-source tool. It is not affiliated with,
> endorsed by, or sponsored by LinkedIn or Microsoft. It reads only public
> pages, with no account and no API key, at a polite default rate.

## Why

LinkedIn's public data sits behind server-rendered pages and a handful of guest
endpoints, with no open API to read it. Pulling anything structured out means
hand-rolling a scraper, guessing at selectors, and redoing the work whenever a
page changes. linkedin puts the public surface behind one tool with sensible
defaults, real output formats, and pipelines that compose. It reads each page
JSON-LD first and falls back to HTML selectors, so a profile or a company comes
back as a complete record, not a bag of strings.

It speaks to `www.linkedin.com` over plain HTTPS. The binary is pure Go with no
runtime dependencies.

## What works anonymously, and what is walled

LinkedIn serves some surfaces to anonymous visitors and gates the rest behind a
sign-in wall. linkedin reads what is public and reports clearly when a page is
walled.

| Surface | Command | Anonymous access |
| --- | --- | --- |
| Member profile | `profile` | Works for many profiles via the Person JSON-LD |
| Company page | `company` | Works via the Organization JSON-LD |
| Job posting | `job` | Works via the guest job-detail fragment |
| Job search | `jobs` | Works via the guest job-search endpoint |
| Public post or article | `post` | Best effort; most are walled |
| School page | (via `company`/`url`) | Usually returns LinkedIn's bot block (999) |

When a page is gated, linkedin exits with code 5 and you can lend a signed-in
session with `--cookies` (a Netscape `cookies.txt` jar exported from your
browser).

## Install

```sh
go install github.com/tamnd/linkedin-cli/cmd/linkedin@latest
```

Or grab a prebuilt binary from the [releases page](https://github.com/tamnd/linkedin-cli/releases),
install a Linux package (`deb`, `rpm`, `apk`), or pull the container image:

```sh
docker run --rm ghcr.io/tamnd/linkedin jobs "golang" -n 5
```

Homebrew and Scoop:

```sh
brew install --cask tamnd/tap/linkedin
scoop install linkedin
```

Build from source:

```sh
git clone https://github.com/tamnd/linkedin-cli
cd linkedin-cli
make build      # produces ./bin/linkedin
```

## Quick start

```sh
linkedin profile williamhgates              # a member profile as a record
linkedin company microsoft                  # a company page as a record
linkedin company microsoft --posts          # the company's recent public posts
linkedin jobs "golang backend" --location Remote   # job stubs from the guest search
linkedin job 4391940951                     # a full job posting
linkedin jobs "data engineer" --hydrate -n 20 --save  # full jobs, into the store
```

## How it works

linkedin reads the same pages a logged-out browser sees and normalizes each one
into a struct, with explicit empty, zero, or `[]` for fields that are genuinely
absent. Most public pages carry a JSON-LD block; linkedin reads that first and
uses CSS selectors to fill in the rest. Responses are cached on disk
(content-addressed and gzipped) so a repeat call is instant and does not hit the
network.

- `profile` reads the **Person JSON-LD** a public profile ships: name, headline,
  location, follower count, current roles, and schools.
- `company` reads the **Organization JSON-LD**: name, description, website,
  address, employee count, and logo. With `--posts` it also collects the
  `DiscussionForumPosting` nodes the page carries.
- `jobs` reads the anonymous **guest job-search endpoint**, paginating in pages
  of 25 until `-n` results are gathered or the endpoint runs dry. With
  `--hydrate` it follows each stub to the full job record.
- `job` reads the guest **job-detail fragment**: title, company, location,
  applicant count, posting date, the full description, and the criteria block
  (seniority, employment type, function, industries).

When a page is behind the sign-in wall or returns LinkedIn's bot block (HTTP
999), linkedin exits cleanly with code 5.

## Commands

| Command | What it does |
| --- | --- |
| `profile <slug\|url>...` | Fetch one or more public member profiles (`--save`) |
| `company <slug\|url>...` | Fetch one or more company pages (`--posts`, `--save`) |
| `job <id\|url>...` | Fetch one or more job postings |
| `jobs <keywords...>` | Search jobs through the guest endpoint (`--location`, `--posted`, `--remote`, `--experience`, `--job-type`, `--sort`, `--hydrate`, `--save`) |
| `post <url>...` | Fetch one or more public posts or articles (best effort) |
| `id <input>...` | Classify and normalize a URL or id into (kind, id) without fetching |
| `url <kind> <id>` | Build a canonical LinkedIn URL |
| `db` | Inspect the local store (`path`, `count`, `query`) |
| `cache` | Inspect and clear the page cache (`path`, `info`, `clear`) |
| `info` | Show the resolved configuration and paths |
| `version` | Print version, commit, and build date |

## Output

Output is a table on a terminal and JSONL when piped, so it drops straight into
a pipeline. Pick any format explicitly with `-f`:

```sh
linkedin company microsoft -f json           # pretty JSON array
linkedin jobs "golang" -f jsonl              # one JSON object per line
linkedin profile williamhgates -f csv        # CSV with a header row
linkedin jobs "golang" -f url                 # just the job URLs
linkedin job 4391940951 --fields title,company,location -f tsv
linkedin profile williamhgates --template '{{.Name}} has {{.Followers}} followers'
```

Choose columns with `--fields`, drop the header with `--no-header`, and apply a
Go `text/template` per record with `--template`.

## Storing records

Records can be upserted into a local SQLite store, keyed by kind and id, so you
can build a dataset over many calls and query it back:

```sh
linkedin jobs "data engineer" --hydrate -n 50 --save   # fetch and store
linkedin company microsoft --save                       # store one company
linkedin db count                                       # how many records, by kind
linkedin db query --kind job -n 10                      # read them back
```

The fetcher is polite by default (2 workers, a 2s spacing) and you can tune it
with `--workers` and `--delay`.

## Exit codes

| Code | Meaning |
| --- | --- |
| 0 | success |
| 1 | error |
| 2 | usage error |
| 3 | no data (not found, empty result) |
| 4 | partial (some items in a batch failed) |
| 5 | blocked (a sign-in wall or bot block; try `--cookies`) |

## Configuration

State lives under `$XDG_DATA_HOME/linkedin` (or `~/.local/share/linkedin`),
overridable with `--data-dir` or `LINKEDIN_DATA_DIR`. The page cache and the
SQLite store both sit there. Politeness and networking knobs (`--delay`,
`--workers`, `--timeout`, `--retries`, `--cache-ttl`, `--no-cache`, `--refresh`,
`--cookies`) are global flags on every command. Run `linkedin info` to see the
resolved paths and `linkedin <command> --help` for the full surface.

## Development

```sh
make build      # build ./bin/linkedin
make test       # go test ./...
make vet        # go vet ./...
make fmt        # gofmt -s -w .
```

## License

[Apache-2.0](LICENSE).
