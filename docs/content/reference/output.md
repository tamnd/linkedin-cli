---
title: "Output formats"
description: "Every output format, how to narrow fields, and how to template records."
weight: 30
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

## Narrowing fields

Keep only the fields you want:

```bash
linkedin job 3801234567 --fields title,company,location
```

`--no-header` drops the header row in `table`, `csv`, and `tsv` output, which is
handy when a downstream tool expects bare rows.

## Templating records

For full control over each line, apply a Go text/template. The fields are the
record's keys:

```bash
linkedin job 3801234567 --template '{{.title}} at {{.company}}'
linkedin jobs "golang engineer" --template '{{.title}}	{{.location}}'
```

## Why auto-detection helps

Because the default adapts to the destination, the same command reads well by
hand and parses cleanly in a pipe:

```bash
linkedin jobs "golang engineer"                  # a table, because this is a terminal
linkedin jobs "golang engineer" | jq -r .url     # JSONL, because this is a pipe
```

You only reach for `--format` when you want something other than that default.

## Color

`--color` is `auto` by default: linkedin colors table output on a terminal and
drops color when piped. Force it with `--color always` or turn it off with
`--color never`.
