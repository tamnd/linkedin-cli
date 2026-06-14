package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/linkedin-cli/linkedin"
)

func (a *App) profileCmd() *cobra.Command {
	var save bool
	cmd := &cobra.Command{
		Use:   "profile <slug-or-url> [more...]",
		Short: "Fetch one or more public member profiles",
		Long: "Fetch public member profiles. Accepts a slug (williamhgates), an\n" +
			"/in/<slug> path, or a full URL. Profiles are parsed from the page's\n" +
			"Person JSON-LD. Many profiles are gated behind a sign-in wall, in which\n" +
			"case the command exits with the blocked code.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var out []linkedin.Profile
			var firstErr error
			fails := 0
			for _, in := range args {
				p, err := a.client.FetchProfile(ctx, a.cache, a.cfg, in)
				if err != nil {
					a.progressf("profile %s: %v", in, err)
					if firstErr == nil {
						firstErr = err
					}
					fails++
					continue
				}
				out = append(out, *p)
				if save {
					if st, e := a.openStore(); e == nil {
						_ = st.Put(linkedin.KindProfile, p.Slug, p.URL, p)
					}
				}
			}
			return a.finishMulti(out, len(args), fails, firstErr)
		},
	}
	cmd.Flags().BoolVar(&save, "save", false, "upsert each record into the store")
	return cmd
}

// finishMulti renders records and picks the right exit code for a multi-arg run.
func (a *App) finishMulti(records any, total, fails int, firstErr error) error {
	n := sliceLen(records)
	if n == 0 {
		if fails > 0 {
			return mapFetchErr(firstErr)
		}
		return codeError(exitNoData, nil)
	}
	if err := a.render(records); err != nil {
		return err
	}
	if fails > 0 && fails < total {
		return codeError(exitPartial, nil)
	}
	return nil
}
