---
title: "Output formats"
description: "Render records as a table, JSON, CSV, or your own template, and script against the exit codes."
weight: 50
---

Every command that emits records renders through the same formatter. Pick a
format with `--format` (or `-f`), or let linkedin choose: a table when writing to
a terminal, JSONL when piped.

## Formats

```bash
linkedin jobs "golang engineer" -f table   # aligned columns for reading
linkedin jobs "golang engineer" -f jsonl   # one JSON object per line, for piping
linkedin jobs "golang engineer" -f json    # a single JSON array
linkedin jobs "golang engineer" -f csv     # spreadsheet friendly
linkedin jobs "golang engineer" -f tsv     # tab-separated
linkedin jobs "golang engineer" -f url     # just the LinkedIn URL of each row
linkedin jobs "golang engineer" -f raw     # the underlying bytes, unformatted
```

`-j` is shorthand for `-f jsonl`.

| Format | Best for |
|---|---|
| `table` | Reading on a terminal |
| `jsonl` | Piping into another tool, one object at a time |
| `json` | Loading a whole result as an array |
| `csv` / `tsv` | Spreadsheets and quick column math |
| `url` | Feeding URLs into other commands |
| `raw` | The unformatted bytes |

## Narrowing columns

Keep only the fields you want:

```bash
linkedin job 3801234567 --fields title,company,location
linkedin company microsoft --fields name,industry,employee_count
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
linkedin jobs "golang engineer"                  # a table, because this is a terminal
linkedin jobs "golang engineer" | jq -r .url     # JSONL, because this is a pipe
```

`--limit` (or `-n`) caps the number of results; `0` means no limit.

## Exit codes for scripting

linkedin returns a stable exit code so a script can branch on the outcome:

| Code | Meaning |
|---|---|
| `0` | OK |
| `1` | Error |
| `2` | Usage error |
| `3` | No data (nothing matched) |
| `4` | Partial (some items failed) |
| `5` | Blocked (behind the sign-in wall) |

For example, treat a walled page differently from a real failure:

```bash
linkedin profile some-member --format json > profile.json
case $? in
  0) echo "got it" ;;
  5) echo "walled, retry with --cookies" ;;
  *) echo "failed" ;;
esac
```

See [troubleshooting](/reference/troubleshooting/) for what to do about exit 5
and exit 3.
