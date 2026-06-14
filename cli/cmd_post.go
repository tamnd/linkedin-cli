package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/linkedin-cli/linkedin"
)

func (a *App) postCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post <url> [more...]",
		Short: "Fetch one or more public posts or articles (best effort)",
		Long: "Fetch public posts and articles, parsed from Open Graph tags and any\n" +
			"Article JSON-LD that serves anonymously. Posts are frequently gated, so\n" +
			"this command often exits with the blocked code.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var out []linkedin.Post
			var firstErr error
			fails := 0
			for _, in := range args {
				p, err := a.client.FetchPost(ctx, a.cache, a.cfg, in)
				if err != nil {
					a.progressf("post %s: %v", in, err)
					if firstErr == nil {
						firstErr = err
					}
					fails++
					continue
				}
				out = append(out, *p)
			}
			return a.finishMulti(out, len(args), fails, firstErr)
		},
	}
	return cmd
}
