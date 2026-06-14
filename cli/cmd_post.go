package cli

import (
	"context"

	"github.com/tamnd/any-cli/kit"
)

func postCmd() kit.Command {
	return kit.Command{
		Use:   "post <url> [more...]",
		Short: "Fetch one or more public posts or articles (best effort)",
		Long: "Fetch public posts and articles, parsed from Open Graph tags and any Article " +
			"JSON-LD that serves anonymously. Posts are frequently gated, so this command " +
			"often exits with the need-auth code.",
		Args: kit.MinimumNArgs(1),
		Run: func(ctx context.Context, args []string) error {
			a := appFromCtx(ctx)
			client, err := a.clientOf()
			if err != nil {
				return err
			}
			sp := a.progress("fetching posts")
			defer sp.stop()

			var rows []Row
			var firstErr error
			for _, in := range args {
				p, err := client.FetchPost(a.ctx(), a.cache, a.cfg, in)
				if err != nil {
					a.logf("post %s: %v", in, err)
					if firstErr == nil {
						firstErr = err
					}
					continue
				}
				rows = append(rows, postRow(p))
			}
			sp.stop()
			return a.finish(rows, firstErr)
		},
	}
}
