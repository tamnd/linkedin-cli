---
title: "Posts and lookups"
description: "Read public posts when you can, classify any LinkedIn input, and build canonical URLs."
weight: 30
---

The best-effort and local-only side of linkedin: reading public posts, turning
any input into a (kind, id) pair without fetching, and building a canonical URL
from a kind and an id.

## Posts

```bash
linkedin post https://www.linkedin.com/posts/example-activity-123456789
```

`post` reads public posts and articles from the page's DiscussionForumPosting
JSON-LD when it can. Be honest with yourself about this one: most posts are
walled behind the sign-in wall, so `post` is best effort. When a post is walled,
linkedin exits with code 5 ("blocked"). Lending a session with `--cookies` gets
more of them through:

```bash
linkedin post https://www.linkedin.com/posts/example-activity-123456789 --cookies ~/cookies.txt
```

See [troubleshooting](/reference/troubleshooting/) for the cookie file format.

## Classify an input

`id` turns a slug, path, or URL into a (kind, id) pair without fetching anything.
It is pure local work, never blocked, and made for scripts:

```bash
linkedin id https://www.linkedin.com/in/williamhgates
```

```
profile	williamhgates
```

The kinds are `profile`, `company`, `school`, `job`, `post`, and `unknown`. It
takes several at once:

```bash
linkedin id williamhgates https://www.linkedin.com/company/microsoft 3801234567
```

Use it to route inputs to the right command, or to validate input before you
spend a request on it. (Note that `school` is a recognized kind, but school
pages are walled, so there is no `school` fetch command.)

## Build a URL

`url` is the inverse of `id`: give it a kind and an id, and it builds the
canonical LinkedIn URL:

```bash
linkedin url profile williamhgates
linkedin url company microsoft
linkedin url job 3801234567
```

Together, `id` and `url` let a script normalize input and re-emit clean links
without touching the network.
