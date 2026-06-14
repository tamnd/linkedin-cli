# linkedin

[![CI](https://github.com/tamnd/linkedin-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/tamnd/linkedin-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/tamnd/linkedin-cli)](https://github.com/tamnd/linkedin-cli/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/tamnd/linkedin-cli.svg)](https://pkg.go.dev/github.com/tamnd/linkedin-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/tamnd/linkedin-cli)](https://goreportcard.com/report/github.com/tamnd/linkedin-cli)
[![License](https://img.shields.io/github/license/tamnd/linkedin-cli)](./LICENSE)

A command line for public LinkedIn data. `linkedin` reads member profiles,
company pages, job postings, and the guest job search and returns rich,
structured records. One pure-Go binary, no API key, no login.

[Install](#install) • [Commands](#commands) • [Usage](#usage) • [What works anonymously](#what-works-anonymously)

![linkedin searching jobs and reading a member profile from the command line](docs/static/demo.gif)

It reads the same pages a logged-out browser sees: the JSON-LD LinkedIn marks up
for machines, then precise HTML selectors for anything the structured data misses.
Responses are cached on disk so a repeat call is instant. When a page is behind
the sign-in wall, `linkedin` exits with a distinct code and you can lend a
signed-in session with `--cookies`.

`linkedin` is an independent tool. It is not affiliated with, endorsed by, or
sponsored by LinkedIn or Microsoft.

## Install

```bash
go install github.com/tamnd/linkedin-cli/cmd/linkedin@latest
```

Or grab a prebuilt binary, a Linux package (`deb`/`rpm`/`apk`), or a container
image from the [releases](https://github.com/tamnd/linkedin-cli/releases):

```bash
brew install tamnd/tap/linkedin
docker run --rm ghcr.io/tamnd/linkedin:latest jobs "golang" -n 5
```

Shell completion is built in: `linkedin completion bash|zsh|fish|powershell`.

## Commands

| Command | Reads |
| --- | --- |
| `linkedin profile <slug\|url>...` | a public member profile; `--posts`, `--articles` |
| `linkedin company <slug\|url>...` | a company page; `--posts`, `--locations`, `--affiliated` |
| `linkedin job <id\|url>...` | a full job posting |
| `linkedin jobs <keywords...>` | guest job search; `--location`, `--remote`, `--posted`, `--experience`, `--job-type`, `--sort`, `--hydrate` |
| `linkedin post <url>...` | a public post or article (best effort via JSON-LD) |
| `linkedin id <input>...` | classify and normalize a URL or id without fetching |
| `linkedin url <kind> <id>` | build a canonical LinkedIn URL |
| `linkedin db path\|count\|query` | inspect the local record store |
| `linkedin serve` | serve all operations over HTTP (NDJSON) |
| `linkedin mcp` | run as an MCP server over stdio |
| `linkedin info` | show the resolved configuration and data paths |
| `linkedin cache path\|info\|clear` | inspect or clear the on-disk page cache |
| `linkedin version` | print version, commit, and build date |

Full reference and guides live at [linkedin-cli.tamnd.com](https://linkedin-cli.tamnd.com).

## Usage

```bash
linkedin profile williamhgates                    # a member profile
linkedin profile williamhgates --posts            # the member's recent public posts
linkedin company microsoft                        # a company page
linkedin company microsoft --locations            # the company's offices
linkedin jobs "golang backend" --location Remote  # job stubs from the guest search
linkedin job 4391940951                           # a full job posting
```

Records come out as a table (the default on a terminal), list, markdown, JSON,
JSONL, CSV, TSV, url, or raw. The output is colored on a terminal:

```bash
linkedin jobs "golang" --fields job_id,title,company,location,posted -o table
linkedin profile williamhgates -o json
linkedin jobs "data engineer" -n 50 -o jsonl | jq 'select(.location | contains("Remote"))'
linkedin jobs "golang" -o url
linkedin company microsoft -o markdown
```

Fetch full job records by hydrating each stub from the guest search:

```bash
linkedin jobs "golang backend" --hydrate -n 20 -o jsonl > jobs.jsonl
```

### Global flags

```
-o, --output      auto|list|table|markdown|json|jsonl|csv|tsv|url|raw   (auto: table on a TTY, jsonl when piped)
    --fields      comma-separated columns to include
    --no-header   omit the header row
    --template    Go text/template applied per record
-n, --limit       max records (0 = no limit)
    --cookies     Netscape cookie jar to lend a signed-in session
-q, --quiet       suppress progress output
    --color       auto|always|never
    --rate        min spacing between requests
    --timeout     per-request timeout
    --retries     retry attempts on rate limit or 5xx
    --no-cache    bypass the on-disk cache
    --refresh     force re-fetch and overwrite the cache
    --cache-ttl   cache freshness window (default 24h)
    --data-dir    override the data directory
    --db          tee every record into a store (sqlite path or postgres:// URL)
    --dry-run     print actions, do not perform them
```

## What works anonymously

LinkedIn serves some surfaces to anonymous visitors and gates the rest behind a
sign-in wall. `linkedin` reads what is public and exits with code 4 when a page
is walled. Pass `--cookies` (a Netscape `cookies.txt` jar from your browser) to
lend a signed-in session.

| Surface | Command | Anonymous |
| --- | --- | --- |
| Member profile | `profile` | Yes — Person JSON-LD |
| Profile posts | `profile --posts` | Yes — JSON-LD graph |
| Profile articles | `profile --articles` | Yes — JSON-LD graph |
| Company page | `company` | Yes — Organization JSON-LD + about panel |
| Company posts | `company --posts` | Yes — JSON-LD graph |
| Company offices | `company --locations` | Yes — location cards |
| Company related pages | `company --affiliated` | Yes — affiliated links |
| Job posting | `job` | Yes — guest job-detail fragment |
| Job search | `jobs` | Yes — guest job-search endpoint |
| Post or article | `post` | Best effort via JSON-LD and Open Graph |

## Storing records

Records can be upserted into a local SQLite store with `--save`, keyed by kind
and id, so you can build a dataset across many calls and query it back:

```bash
linkedin jobs "data engineer" --hydrate -n 50 --save   # fetch and store
linkedin company microsoft --save                       # store a company
linkedin db count                                       # records by kind
linkedin db query --kind job -n 10                      # read them back
```

Pass `--db path/to/file.db` to tee records into any SQLite file while they
stream, without `--save`.

## Exit codes

```
0  success
1  error
2  usage error
3  no results
4  auth required (page is behind the sign-in wall; pass --cookies)
5  rate limited (HTTP 429 after retries)
6  not found
```

## Development

```
cmd/linkedin/   thin main entry point
cli/            cobra commands and output rendering
linkedin/       HTTP client, parsers, and models
docs/           documentation site (Hugo, tago-doks theme)
```

```bash
make build   # ./bin/linkedin
make test    # go test ./...
make vet     # go vet ./...
make fmt     # gofmt -s -w .
```

Requires Go 1.23+.

## Releasing

Push a version tag and GitHub Actions runs GoReleaser:

```bash
git tag -a v0.2.0 -m "v0.2.0"
git push --tags
```

The image tag carries no `v` prefix (`ghcr.io/tamnd/linkedin:0.2.0`).

## License

Apache-2.0. See [LICENSE](LICENSE).
