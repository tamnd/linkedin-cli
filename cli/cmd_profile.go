package cli

import (
	"context"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/linkedin-cli/linkedin"
)

func profileCmd() kit.Command {
	var save, posts, articles bool
	return kit.Command{
		Use:   "profile <slug-or-url> [more...]",
		Short: "Fetch one or more public member profiles",
		Long: "Fetch public member profiles. Accepts a slug (williamhgates), an /in/<slug> " +
			"path, or a full URL. Profiles are parsed from the page's Person JSON-LD. With " +
			"--posts, emit the recent posts carried in the page's JSON-LD; with --articles, " +
			"emit the member's long-form articles. Many profiles are gated behind a sign-in " +
			"wall, in which case the command exits with the need-auth code.",
		Args: kit.MinimumNArgs(1),
		Flags: func(f *kit.FlagSet) {
			f.BoolVar(&save, "save", false, "upsert each record into the store")
			f.BoolVar(&posts, "posts", false, "emit the profile's recent posts instead of the profile record")
			f.BoolVar(&articles, "articles", false, "emit the profile's long-form articles instead of the profile record")
		},
		Run: func(ctx context.Context, args []string) error {
			a := appFromCtx(ctx)
			client, err := a.clientOf()
			if err != nil {
				return err
			}
			sp := a.progress("fetching profiles")
			defer sp.stop()

			var rows []Row
			var firstErr error
			switch {
			case posts:
				for _, in := range args {
					ps, err := client.FetchProfilePosts(a.ctx(), a.cache, a.cfg, in)
					if err != nil {
						a.logf("profile %s: %v", in, err)
						if firstErr == nil {
							firstErr = err
						}
						continue
					}
					for i := range ps {
						rows = append(rows, postRow(&ps[i]))
					}
				}
			case articles:
				for _, in := range args {
					as, err := client.FetchProfileArticles(a.ctx(), a.cache, a.cfg, in)
					if err != nil {
						a.logf("profile %s: %v", in, err)
						if firstErr == nil {
							firstErr = err
						}
						continue
					}
					for i := range as {
						rows = append(rows, articleRow(&as[i]))
					}
				}
			default:
				for _, in := range args {
					p, err := client.FetchProfile(a.ctx(), a.cache, a.cfg, in)
					if err != nil {
						a.logf("profile %s: %v", in, err)
						if firstErr == nil {
							firstErr = err
						}
						continue
					}
					rows = append(rows, profileRow(p))
					if save {
						if st, e := a.openStore(); e == nil {
							_ = st.Put(linkedin.KindProfile, p.Slug, p.URL, p)
						}
					}
				}
			}
			sp.stop()
			return a.finish(rows, firstErr)
		},
	}
}
