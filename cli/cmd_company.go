package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/linkedin-cli/linkedin"
)

func (a *App) companyCmd() *cobra.Command {
	var save, posts bool
	cmd := &cobra.Command{
		Use:   "company <slug-or-url> [more...]",
		Short: "Fetch one or more company pages",
		Long: "Fetch public company pages, parsed from the Organization JSON-LD. With\n" +
			"--posts, emit the recent company posts carried in the page's JSON-LD as\n" +
			"Post records instead of the Company record.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if posts {
				var out []linkedin.Post
				var firstErr error
				fails := 0
				for _, in := range args {
					ps, err := a.client.FetchCompanyPosts(ctx, a.cache, a.cfg, in)
					if err != nil {
						a.progressf("company %s: %v", in, err)
						if firstErr == nil {
							firstErr = err
						}
						fails++
						continue
					}
					out = append(out, ps...)
				}
				return a.finishMulti(out, len(args), fails, firstErr)
			}

			var out []linkedin.Company
			var firstErr error
			fails := 0
			for _, in := range args {
				c, err := a.client.FetchCompany(ctx, a.cache, a.cfg, in)
				if err != nil {
					a.progressf("company %s: %v", in, err)
					if firstErr == nil {
						firstErr = err
					}
					fails++
					continue
				}
				out = append(out, *c)
				if save {
					if st, e := a.openStore(); e == nil {
						_ = st.Put(linkedin.KindCompany, c.Slug, c.URL, c)
					}
				}
			}
			return a.finishMulti(out, len(args), fails, firstErr)
		},
	}
	cmd.Flags().BoolVar(&save, "save", false, "upsert each record into the store")
	cmd.Flags().BoolVar(&posts, "posts", false, "emit recent company posts instead of the company record")
	return cmd
}
