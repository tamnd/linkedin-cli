package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/linkedin-cli/linkedin"
)

func (a *App) jobCmd() *cobra.Command {
	var save bool
	cmd := &cobra.Command{
		Use:   "job <id-or-url> [more...]",
		Short: "Fetch one or more job postings",
		Long: "Fetch job postings through the anonymous guest job-detail endpoint.\n" +
			"Accepts a numeric job id, a urn:li:jobPosting URN, or a /jobs/view/ URL.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var out []linkedin.Job
			var firstErr error
			fails := 0
			for _, in := range args {
				j, err := a.client.FetchJob(ctx, a.cache, a.cfg, in)
				if err != nil {
					a.progressf("job %s: %v", in, err)
					if firstErr == nil {
						firstErr = err
					}
					fails++
					continue
				}
				out = append(out, *j)
				if save {
					if st, e := a.openStore(); e == nil {
						_ = st.Put(linkedin.KindJob, j.JobID, j.URL, j)
					}
				}
			}
			return a.finishMulti(out, len(args), fails, firstErr)
		},
	}
	cmd.Flags().BoolVar(&save, "save", false, "upsert each record into the store")
	return cmd
}

func (a *App) jobsCmd() *cobra.Command {
	var opts linkedin.JobSearchOptions
	var hydrate, save bool
	cmd := &cobra.Command{
		Use:   "jobs <keywords...>",
		Short: "Search jobs through the guest endpoint",
		Long: "Search jobs through the anonymous guest endpoint, paginating until -n\n" +
			"results are gathered or the endpoint runs dry. Emits JobStub records, or\n" +
			"full Job records with --hydrate.",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			opts.Keywords = joinArgs(args)
			if opts.Keywords == "" && opts.Location == "" {
				return codeError(exitUsage, errNeedQuery)
			}
			stubs, err := a.client.SearchJobs(ctx, a.cache, a.cfg, opts, a.limit)
			if err != nil {
				return mapFetchErr(err)
			}
			if !hydrate {
				return a.renderOrEmpty(stubs, len(stubs))
			}

			var out []linkedin.Job
			fails := 0
			for i, s := range stubs {
				a.progressf("hydrating %d/%d job %s", i+1, len(stubs), s.JobID)
				j, e := a.client.FetchJob(ctx, a.cache, a.cfg, s.JobID)
				if e != nil {
					a.progressf("job %s: %v", s.JobID, e)
					fails++
					continue
				}
				out = append(out, *j)
				if save {
					if st, se := a.openStore(); se == nil {
						_ = st.Put(linkedin.KindJob, j.JobID, j.URL, j)
					}
				}
			}
			return a.finishMulti(out, len(stubs), fails, linkedin.ErrBlocked)
		},
	}
	f := cmd.Flags()
	f.StringVar(&opts.Location, "location", "", "location filter (e.g. Remote, \"United States\", a city)")
	f.StringVar(&opts.GeoID, "geo-id", "", "LinkedIn geo id when known")
	f.StringVar(&opts.Posted, "posted", "", "date posted window: r86400 (24h), r604800 (week), r2592000 (month)")
	f.StringVar(&opts.Remote, "remote", "", "workplace type: 1 on-site, 2 remote, 3 hybrid")
	f.StringVar(&opts.Experience, "experience", "", "experience level: 1..6")
	f.StringVar(&opts.JobType, "job-type", "", "job type: F,P,C,T,I,V,O")
	f.StringVar(&opts.Sort, "sort", "", "sort: R relevance, DD date")
	f.BoolVar(&hydrate, "hydrate", false, "follow each stub to a full job record")
	f.BoolVar(&save, "save", false, "with --hydrate, upsert each job into the store")
	return cmd
}
