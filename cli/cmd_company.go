package cli

import (
	"context"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/linkedin-cli/linkedin"
)

func companyCmd() kit.Command {
	var save, posts, locations, affiliated bool
	return kit.Command{
		Use:   "company <slug-or-url> [more...]",
		Short: "Fetch one or more company pages",
		Long: "Fetch public company pages, parsed from the Organization JSON-LD. With " +
			"--posts, --locations, or --affiliated, emit the recent company posts, the " +
			"office list, or the related pages the page carries instead of the Company " +
			"record.",
		Args: kit.MinimumNArgs(1),
		Flags: func(f *kit.FlagSet) {
			f.BoolVar(&save, "save", false, "upsert each record into the store")
			f.BoolVar(&posts, "posts", false, "emit recent company posts instead of the company record")
			f.BoolVar(&locations, "locations", false, "emit the company's office locations instead of the company record")
			f.BoolVar(&affiliated, "affiliated", false, "emit the company's affiliated and showcase pages instead of the company record")
		},
		Run: func(ctx context.Context, args []string) error {
			a := appFromCtx(ctx)
			client, err := a.clientOf()
			if err != nil {
				return err
			}
			sp := a.progress("fetching companies")
			defer sp.stop()

			var rows []Row
			var firstErr error
			note := func(in string, err error) {
				a.logf("company %s: %v", in, err)
				if firstErr == nil {
					firstErr = err
				}
			}
			switch {
			case posts:
				for _, in := range args {
					ps, err := client.FetchCompanyPosts(a.ctx(), a.cache, a.cfg, in)
					if err != nil {
						note(in, err)
						continue
					}
					for i := range ps {
						rows = append(rows, postRow(&ps[i]))
					}
				}
			case locations:
				for _, in := range args {
					ls, err := client.FetchCompanyLocations(a.ctx(), a.cache, a.cfg, in)
					if err != nil {
						note(in, err)
						continue
					}
					for i := range ls {
						rows = append(rows, locationRow(&ls[i]))
					}
				}
			case affiliated:
				for _, in := range args {
					rs, err := client.FetchCompanyAffiliated(a.ctx(), a.cache, a.cfg, in)
					if err != nil {
						note(in, err)
						continue
					}
					for i := range rs {
						rows = append(rows, orgRefRow(&rs[i]))
					}
				}
			default:
				for _, in := range args {
					c, err := client.FetchCompany(a.ctx(), a.cache, a.cfg, in)
					if err != nil {
						note(in, err)
						continue
					}
					rows = append(rows, companyRow(c))
					if save {
						if st, e := a.openStore(); e == nil {
							_ = st.Put(linkedin.KindCompany, c.Slug, c.URL, c)
						}
					}
				}
			}
			sp.stop()
			return a.finish(rows, firstErr)
		},
	}
}
