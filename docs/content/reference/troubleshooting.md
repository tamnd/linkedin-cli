---
title: "Troubleshooting"
description: "The handful of things that trip people up, and how to fix each one."
weight: 40
---

Most of these come down to network reality, not a bug. LinkedIn is a public
website with a sign-in wall, and linkedin is honest about what it can and cannot
read.

## "blocked" and exit code 5

LinkedIn serves some surfaces to anonymous visitors and walls the rest behind a
sign-in wall. When the page linkedin asks for is walled, it exits with code 5
("blocked") rather than returning the wall as if it were data. The wall shows up
two ways: HTTP 999 is LinkedIn's bot block, and an authwall is a redirect to
`/authwall`, `/uas/login`, `/login`, `/checkpoint`, or `/signup`.

What is walled, plainly:

- **School pages** return the bot block (HTTP 999).
- **People search** is walled.
- **`post`** is best effort; most posts are walled.
- **`profile`** works for many members, but some are walled.

What works anonymously: `company`, `job`, `jobs` search, and many `profile`
lookups.

What to do, in order:

1. **Use the surfaces that work anonymously.** `jobs` and `job` use the guest
   endpoints, `company` reads the Organization JSON-LD, and most public profiles
   read fine. Prefer them for the fields they carry.
2. **Slow down and retry.** The default `--delay` is already two seconds. A
   block is sometimes transient; the same page can succeed a moment later.
3. **Lend a session with `--cookies`.** Export a Netscape `cookies.txt` jar from
   a signed-in browser and pass it:

   ```bash
   linkedin profile some-walled-member --cookies ~/cookies.txt
   ```

   A real session usually clears the wall.

## The cookies.txt format

`--cookies` expects a Netscape cookie jar: the plain-text format most browser
extensions export and `curl` reads. Each line is tab-separated:

```
www.linkedin.com	FALSE	/	TRUE	0	li_at	abc123...
```

Lines starting with `#` are comments. Export it from a browser where you are
signed in to LinkedIn, save it somewhere private, and pass its path to
`--cookies`. linkedin only replays the jar; it never logs in for you and never
stores credentials.

## "no data" and exit code 3

Exit code 3 means linkedin reached the page but found nothing to return: a 404,
a job search with no matches, an empty result. Check the slug, id, or URL is
right (use `linkedin id <input>` to see how linkedin classifies it), broaden a
job search, or loosen the filters.

## Rate limiting (429)

If LinkedIn returns 429 (too many requests), linkedin backs off and retries up
to `--retries` times. If you see this often, you are going too fast: raise
`--delay`, lower `--workers`, and let the cache absorb repeat fetches. The
defaults (two second delay, two workers) are set to avoid this.

## A multi-fetch reports failures (exit code 4)

When you pass several inputs at once (`profile a b c`, `job 1 2 3`, or `jobs
--hydrate`), linkedin exits 4 if it returned some records but others failed
(often a walled profile or post in the batch). The records that did parse are
emitted; re-run the failed ones later, or pass `--cookies` for the walled ones.
Exit 3 means nothing came back at all.

## Where state lives

The on-disk cache and the SQLite store both live under the data dir (the XDG
data directory by default, or `LINKEDIN_DATA_DIR` / `--data-dir`). The store file
alone can be moved with `--store`. To see the resolved paths:

```bash
linkedin info
```

To clear the cache and start fresh:

```bash
linkedin cache clear
```
