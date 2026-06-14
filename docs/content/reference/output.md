---
title: "Output formats"
description: "Every output format, how to narrow fields, and how to template records."
weight: 30
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
linkedin jobs "golang engineer"                  # a list, because this is a terminal
linkedin jobs "golang engineer" | jq -r .url     # JSONL, because this is a pipe
```

You only reach for `--output` when you want something other than that default,
for example `-o table` for the bordered grid or `-o markdown` for a paste-ready
table.

## Color

`--color` is `auto` by default: linkedin colors table output on a terminal and
drops color when piped. Force it with `--color always` or turn it off with
`--color never`.
