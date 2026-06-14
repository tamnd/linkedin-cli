package cli

import (
	"context"
	"errors"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
	"github.com/tamnd/linkedin-cli/linkedin"
)

// errNeedQuery is returned when `jobs` is invoked with neither keywords nor a
// location.
var errNeedQuery = errors.New("provide search keywords or a --location")

func jobCmd() kit.Command {
	var save bool
	return kit.Command{
		Use:   "job <id-or-url> [more...]",
		Short: "Fetch one or more job postings",
		Long: "Fetch job postings through the anonymous guest job-detail endpoint. Accepts " +
			"a numeric job id, a urn:li:jobPosting URN, or a /jobs/view/ URL.",
		Args: kit.MinimumNArgs(1),
		Flags: func(f *kit.FlagSet) {
			f.BoolVar(&save, "save", false, "upsert each record into the store")
		},
		Run: func(ctx context.Context, args []string) error {
			a := appFromCtx(ctx)
			client, err := a.clientOf()
			if err != nil {
				return err
			}
			sp := a.progress("fetching jobs")
			defer sp.stop()

			var rows []Row
			var firstErr error
			for _, in := range args {
				j, err := client.FetchJob(a.ctx(), a.cache, a.cfg, in)
				if err != nil {
					a.logf("job %s: %v", in, err)
					if firstErr == nil {
						firstErr = err
					}
					continue
				}
				rows = append(rows, jobRow(j))
				if save {
					if st, e := a.openStore(); e == nil {
						_ = st.Put(linkedin.KindJob, j.JobID, j.URL, j)
					}
				}
			}
			sp.stop()
			return a.finish(rows, firstErr)
		},
	}
}

func jobsCmd() kit.Command {
	var opts linkedin.JobSearchOptions
	var hydrate, save bool
	return kit.Command{
		Use:   "jobs <keywords...>",
		Short: "Search jobs through the guest endpoint",
		Long: "Search jobs through the anonymous guest endpoint, paginating until -n results " +
			"are gathered or the endpoint runs dry. Emits JobStub records, or full Job " +
			"records with --hydrate.",
		Flags: func(f *kit.FlagSet) {
			f.StringVar(&opts.Location, "location", "", "location filter (e.g. Remote, \"United States\", a city)")
			f.StringVar(&opts.GeoID, "geo-id", "", "LinkedIn geo id when known")
			f.StringVar(&opts.Posted, "posted", "", "date posted window: r86400 (24h), r604800 (week), r2592000 (month)")
			f.StringVar(&opts.Remote, "remote", "", "workplace type: 1 on-site, 2 remote, 3 hybrid")
			f.StringVar(&opts.Experience, "experience", "", "experience level: 1..6")
			f.StringVar(&opts.JobType, "job-type", "", "job type: F,P,C,T,I,V,O")
			f.StringVar(&opts.Sort, "sort", "", "sort: R relevance, DD date")
			f.BoolVar(&hydrate, "hydrate", false, "follow each stub to a full job record")
			f.BoolVar(&save, "save", false, "with --hydrate, upsert each job into the store")
		},
		Run: func(ctx context.Context, args []string) error {
			a := appFromCtx(ctx)
			client, err := a.clientOf()
			if err != nil {
				return err
			}
			opts.Keywords = joinArgs(args)
			if opts.Keywords == "" && opts.Location == "" {
				return errs.Usage("%v", errNeedQuery)
			}
			sp := a.progress("searching jobs")
			defer sp.stop()

			stubs, err := client.SearchJobs(a.ctx(), a.cache, a.cfg, opts, a.limit)
			if err != nil {
				return mapErr(err)
			}
			if !hydrate {
				sp.stop()
				rows := make([]Row, 0, len(stubs))
				for i := range stubs {
					rows = append(rows, jobStubRow(&stubs[i]))
				}
				return a.finish(rows, nil)
			}

			var rows []Row
			for i := range stubs {
				s := &stubs[i]
				a.logf("hydrating %d/%d job %s", i+1, len(stubs), s.JobID)
				j, e := client.FetchJob(a.ctx(), a.cache, a.cfg, s.JobID)
				if e != nil {
					a.logf("job %s: %v", s.JobID, e)
					continue
				}
				rows = append(rows, jobRow(j))
				if save {
					if st, se := a.openStore(); se == nil {
						_ = st.Put(linkedin.KindJob, j.JobID, j.URL, j)
					}
				}
			}
			sp.stop()
			return a.finish(rows, nil)
		},
	}
}
