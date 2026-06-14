---
title: "Configuration"
description: "The data directory, the store, cookies, politeness knobs, environment, and exit codes."
weight: 20
---

linkedin needs almost no configuration. There is no config file; every option is
a flag or an environment variable, and the defaults are chosen so the common
case needs neither. See everything linkedin resolved with:

```bash
linkedin info
```

It prints the resolved configuration and the paths.

## The data directory

linkedin keeps its state under one tree: the on-disk page cache and the SQLite
store. It defaults to the XDG data directory (for example
`~/.local/share/linkedin` on Linux). Point it elsewhere with `--data-dir` or the
`LINKEDIN_DATA_DIR` environment variable.

## The store

Records you save with `--save` land in a SQLite file fixed at
`<data-dir>/linkedin.db`. To keep one corpus per project, point the whole data
dir at a per-project directory with `--data-dir`; the store follows it:

```bash
linkedin company microsoft --save --data-dir ~/projects/hiring
```

`db path`, `db count`, and `db query` all read this file.

## Cookies

The `--cookies` flag takes a Netscape `cookies.txt` jar exported from a
signed-in browser session. linkedin sends those cookies with each request, which
lends it a real session and often gets past the sign-in wall on a walled post or
when an anonymous request is rate-limited. linkedin never logs in for you and
never stores credentials; it only replays the jar you hand it.

```bash
linkedin profile williamhgates --cookies ~/cookies.txt
```

See [troubleshooting](/reference/troubleshooting/) for the cookie file format.

## Caching

Every fetch goes through a content-addressed gzip cache on disk so a repeat run
does not re-fetch unchanged pages. `--cache-ttl` sets how long an entry stays
fresh (default 24h). `--no-cache` bypasses it for one run, and `--refresh`
forces a re-fetch and rewrites the entry. Manage the cache with `cache info`,
`cache path`, and `cache clear`.

## Politeness

linkedin is gentle by default so a busy session stays a good citizen against a
public site:

| Flag | Default | Meaning |
|---|---|---|
| `--rate` | `2s` | Minimum spacing between requests |
| `--timeout` | `30s` | Per-request timeout |
| `--retries` | `3` | Retry attempts on 429/5xx |

Lower `--rate` only when you have a reason to, and keep it modest.

## Environment variables

| Variable | Used for |
|---|---|
| `LINKEDIN_DATA_DIR` | Root data directory (overrides the XDG default) |

## Global flags

| Flag | Default | Meaning |
|---|---|---|
| `-o, --output` | auto | `auto`, `list`, `table`, `markdown`, `json`, `jsonl`, `csv`, `tsv`, `url`, `raw` (auto: list on a TTY, jsonl piped) |
| `--fields` | all | Comma-separated columns to include |
| `--no-header` | off | Omit the header row in table/csv/tsv output |
| `--template` | none | Go text/template applied per record |
| `--color` | auto | `auto`, `always`, or `never` |
| `-n, --limit` | `0` | Limit number of results; `0` is no limit |
| `-q, --quiet` | off | Suppress progress on stderr |
| `-v, --verbose` | off | Increase verbosity (repeatable) |
| `--rate` | `2s` | Minimum spacing between requests |
| `--timeout` | `30s` | Per-request timeout |
| `--retries` | `3` | Retry attempts on 429/5xx |
| `--cache-ttl` | `24h` | On-disk cache freshness window |
| `--no-cache` | off | Bypass the on-disk page cache |
| `--refresh` | off | Force a re-fetch and overwrite the cache |
| `--data-dir` | XDG | Root dir for cache and store (env `LINKEDIN_DATA_DIR`) |
| `--cookies` | none | Netscape cookie jar |
| `--dry-run` | off | Print actions, do not perform them |
| `--profile` | none | Named profile to load |

## Output auto-detection

The default output format adapts to where it is going: the readable list view
when the output is a terminal, JSONL when it is piped. That keeps interactive use
readable and scripted use parseable without you setting `--output` either time.
Reach for `-o table` for the bordered grid or `-o markdown` for a paste-ready
table. See [output formats](/reference/output/) for the full set.

## Exit codes

linkedin returns a stable exit code so scripts can branch on the outcome:

| Code | Meaning |
|---|---|
| `0` | OK |
| `1` | Error (generic failure) |
| `2` | Usage error (bad flags or arguments) |
| `3` | No results (nothing matched) |
| `4` | Auth required (behind the sign-in wall; pass `--cookies`) |
| `5` | Rate limited (HTTP 429 after retries) |
| `6` | Not found (a 404 or an unknown id) |
