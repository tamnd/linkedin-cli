package cli

import (
	"strings"

	"github.com/tamnd/any-cli/kit/errs"
)

// joinArgs reassembles a multi-word query the user typed without quotes.
func joinArgs(args []string) string { return strings.TrimSpace(strings.Join(args, " ")) }

// emit renders a list of rows, applying --limit, and returns the no-results
// error (exit 3) when there is nothing to show.
func (a *App) emit(rows []Row) error {
	if len(rows) == 0 {
		return errs.NoResults("no results")
	}
	out, err := a.out()
	if err != nil {
		return err
	}
	for i, r := range rows {
		if a.limit > 0 && i >= a.limit {
			break
		}
		if err := out.Emit(r); err != nil {
			return err
		}
	}
	return out.Flush()
}

// finish renders the gathered rows for a multi-argument run. When nothing came
// back it reports the first fetch error (mapped to its kind) if any, else the
// no-results code. Per-argument failures are surfaced as stderr warnings while
// the run was gathering, so a partial run still prints what it got.
func (a *App) finish(rows []Row, firstErr error) error {
	if len(rows) == 0 {
		if firstErr != nil {
			return mapErr(firstErr)
		}
		return errs.NoResults("no results")
	}
	return a.emit(rows)
}
