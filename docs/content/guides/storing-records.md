---
title: "Storing records"
description: "Save records into a local SQLite store across calls, read them back with db query, and manage the page cache."
weight: 40
---

linkedin has no bulk crawler: no sitemap seed, no queue, no drain. The way you
build a dataset is to pass `--save` on the commands that support it, run them as
many times as you like, and read the accumulated records back out with `db`.
Everything lands in one SQLite file under your data dir.

## Saving records

`profile`, `company`, and `jobs --hydrate` all take `--save`, which upserts each
record into the store. Run them across a session and the store fills up:

```bash
linkedin company microsoft --save
linkedin company github --save
linkedin profile williamhgates --save
linkedin jobs "golang engineer" --hydrate --save -n 50
```

Because it is an upsert, re-running a command updates the existing row rather than
duplicating it, so you can refresh a record later by saving it again (add
`--refresh` to force a fresh fetch first).

## Reading the store back

`db` works with the local SQLite store:

```bash
linkedin db path     # print the store file path
linkedin db count    # how many records are stored
linkedin db query    # read stored records back out
```

`db query` emits the stored records through the same formatter as everything
else, so you can shape and pipe them:

```bash
linkedin db query --output jsonl | jq -r .name
linkedin db query --fields name,industry,employees
```

## A small dataset, end to end

Building a slice of companies and their open roles looks like this:

```bash
linkedin company microsoft --posts --save
linkedin jobs "golang engineer" --location "United States" --hydrate --save -n 100
linkedin db query --output jsonl > dataset.jsonl
```

## The page cache

Separate from the store, every fetch goes through an on-disk cache
(content-addressed, gzip), so a repeat run does not re-fetch pages that have not
changed. `cache` manages it:

```bash
linkedin cache path                                       # the cache directory path
linkedin cache info                                       # location, file count, size
linkedin cache clear                                      # remove every cached page
```

The cache TTL defaults to 24 hours. Bypass it for one run with `--no-cache`, or
force a re-fetch with `--refresh`.

## Where state lives

The store and cache both live under the data dir; point that elsewhere with
`--data-dir` or `LINKEDIN_DATA_DIR`. The store file is fixed at
`<data-dir>/linkedin.db`, so to keep one corpus per project, point `--data-dir`
at a per-project directory. See [configuration](/reference/configuration/).
