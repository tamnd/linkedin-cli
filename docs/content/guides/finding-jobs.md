---
title: "Finding jobs"
description: "Search the jobs board through the anonymous guest endpoint, filter it down, and read a posting in full."
weight: 20
---

The jobs board reads reliably for anonymous visitors, because search and detail
both run through LinkedIn's guest endpoints. Search the board with `jobs`, narrow
it with the filters, then read a posting in full with `job`.

## Search the board

```bash
linkedin jobs "golang engineer" -n 25
```

`jobs` searches through the anonymous guest endpoint
(`/jobs-guest/jobs/api/seeMoreJobPostings/search`), paginating in pages of 25
until `-n` results or the endpoint runs dry. Each row is a JobStub (the title,
company, company logo, location, id, and URL).

## Filters

The filters map onto the board's own parameters:

| Flag | Meaning |
|---|---|
| `--location` | Free-text location, e.g. `Remote`, `"United States"`, a city |
| `--geo-id` | LinkedIn geo id, when you know it |
| `--posted` | Posting age: `r86400` (24h), `r604800` (week), `r2592000` (month) |
| `--remote` | Workplace: `1` on-site, `2` remote, `3` hybrid |
| `--experience` | Experience level, `1`..`6` |
| `--job-type` | Type: `F`, `P`, `C`, `T`, `I`, `V`, `O` |
| `--sort` | `R` relevance, `DD` date |

Remote roles posted in the last week, sorted by date:

```bash
linkedin jobs "golang engineer" --remote 2 --posted r604800 --sort DD -n 50
```

In a city, full-time, mid-senior:

```bash
linkedin jobs "backend engineer" --location "Berlin" --job-type F --experience 4
```

## Full records with --hydrate

A stub is thin. Add `--hydrate` to follow each one to its detail page and emit a
full Job record instead:

```bash
linkedin jobs "golang engineer" --hydrate -n 10 --output json
```

With `--hydrate`, `--save` upserts each job into the local store as it goes:

```bash
linkedin jobs "golang engineer" --hydrate --save -n 50
```

## Read one posting in full

```bash
linkedin job 3801234567
```

`job` takes a job id or a full URL and reads the guest job-detail fragment
(`/jobs-guest/jobs/api/jobPosting/<id>`). A job record carries the title,
company, company logo, location, applicant count, posting date, the full
description, and the criteria (seniority, employment type, job function,
industries). It takes several ids at once:

```bash
linkedin job 3801234567 3809876543 --output csv
```

## From a search to a record

A stub carries the job id, so a search composes straight into a detail lookup:

```bash
linkedin jobs "golang engineer" --output jsonl | jq -r .job_id | head -1
# then: linkedin job <that id>
```

Or just use `--hydrate` to get the full records in one pass. For saving a search
into a local dataset, see [storing records](/guides/storing-records/).
