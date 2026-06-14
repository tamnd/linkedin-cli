---
title: "Output formats"
description: "Render records as a table, JSON, CSV, or your own template, and script against the exit codes."
weight: 50
---

Every command that emits records renders through the same formatter. Pick a
format with `--output` (or `-o`), or let linkedin choose: the readable list view
when writing to a terminal, JSONL when piped.

## Formats

```bash
linkedin jobs "golang engineer" -o list     # a readable per-record section view
linkedin jobs "golang engineer" -o table    # bordered, aligned columns
linkedin jobs "golang engineer" -o markdown # a GitHub-flavored pipe table
linkedin jobs "golang engineer" -o jsonl    # one JSON object per line, for piping
linkedin jobs "golang engineer" -o json     # a single JSON array
linkedin jobs "golang engineer" -o csv      # spreadsheet friendly
linkedin jobs "golang engineer" -o tsv      # tab-separated
linkedin jobs "golang engineer" -o url      # just the LinkedIn URL of each row
linkedin jobs "golang engineer" -o raw      # the underlying bytes, unformatted
```

| Format | Best for |
|---|---|
| `list` | Reading on a terminal, the default there |
| `table` | A bordered, aligned grid |
| `markdown` | Pasting into a doc or an issue |
| `jsonl` | Piping into another tool, one object at a time |
| `json` | Loading a whole result as an array |
| `csv` / `tsv` | Spreadsheets and quick column math |
| `url` | Feeding URLs into other commands |
| `raw` | The unformatted bytes |

## Narrowing columns

Keep only the fields you want:

```bash
linkedin job 3801234567 --fields title,company,location
linkedin company microsoft --fields name,industry,employees
```

`--no-header` drops the header row in `table`, `csv`, and `tsv` output, which is
handy when a downstream tool expects bare rows.

## Templating rows

For full control over each line, apply a Go text/template. The fields are the
record's keys:

```bash
linkedin job 3801234567 --template '{{.title}} at {{.company}} ({{.location}})'
linkedin jobs "golang engineer" --template '{{.title}}	{{.location}}'
```

## Piping

Because the default adapts to the destination, the same command reads well by
hand and parses cleanly in a pipe:

```bash
linkedin jobs "golang engineer"                  # a list, because this is a terminal
linkedin jobs "golang engineer" | jq -r .url     # JSONL, because this is a pipe
```

`--limit` (or `-n`) caps the number of results; `0` means no limit.

## Exit codes for scripting

linkedin returns a stable exit code so a script can branch on the outcome:

| Code | Meaning |
|---|---|
| `0` | OK |
| `1` | Error (generic failure) |
| `2` | Usage error (bad flags or arguments) |
| `3` | No results (nothing matched) |
| `4` | Auth required (behind the sign-in wall; pass `--cookies`) |
| `5` | Rate limited (HTTP 429 after retries) |
| `6` | Not found (a 404 or an unknown id) |

For example, treat a walled page differently from a real failure:

```bash
linkedin profile some-member --output json > profile.json
case $? in
  0) echo "got it" ;;
  4) echo "walled, retry with --cookies" ;;
  *) echo "failed" ;;
esac
```

See [troubleshooting](/reference/troubleshooting/) for what to do about exit 4
and exit 3.
