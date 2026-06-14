package cli

import (
	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/linkedin-cli/linkedin"
)

// Build metadata, stamped via -ldflags by goreleaser. goreleaser targets
// github.com/tamnd/linkedin-cli/cli.{Version,Commit,Date}.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// New builds the kit App: the identity, the linkedin-specific global flags, the
// per-run config defaults, and every command as a kit escape hatch. cli only
// touches kit here; kit wraps cobra/fang internally.
func New() *kit.App {
	app := kit.New(kit.Identity{
		Binary: "linkedin",
		Short:  "Read public LinkedIn data from the command line",
		Long: "linkedin reads the public profile, company, and job pages and the guest " +
			"job-search endpoints LinkedIn serves to anonymous visitors, and returns " +
			"rich, structured records as a list, table, JSON, JSONL, CSV, TSV, or URLs. " +
			"It uses no API key and no account. Some member surfaces are gated behind a " +
			"sign-in wall, so commands report clearly when a page is walled. linkedin is " +
			"an independent tool and is not affiliated with, endorsed by, or sponsored by " +
			"LinkedIn.",
		Version: Version,
		Site:    "https://www.linkedin.com",
		Repo:    "https://github.com/tamnd/linkedin-cli",
	}, kit.WithDefaults(withLinkedinDefaults))

	app.GlobalFlags(bindLinkedinFlags)

	app.AddCommand(profileCmd())
	app.AddCommand(companyCmd())
	app.AddCommand(jobCmd())
	app.AddCommand(jobsCmd())
	app.AddCommand(postCmd())
	app.AddCommand(idCmd())
	app.AddCommand(urlCmd())
	app.AddCommand(dbCmd())
	app.AddCommand(cacheCmd())
	app.AddCommand(infoCmd())
	app.AddCommand(versionCmd())
	return app
}

// withLinkedinDefaults overlays linkedin's politer request defaults onto the kit
// baseline so help and the resolved config read the same whether or not the user
// passes flags.
func withLinkedinDefaults(c *kit.Config) {
	c.Rate = linkedin.DefaultDelay
	c.Retries = linkedin.DefaultRetries
	c.Timeout = linkedin.DefaultTimeout
}

// bindLinkedinFlags registers the linkedin-only persistent flags on the root.
// The framework already provides -o/--output, --fields, --template, --no-header,
// -n/--limit, --rate, --retries, --timeout, --data-dir, --no-cache, -q/--quiet,
// -v, --color, and --dry-run, so linkedin adds only these three.
func bindLinkedinFlags(f *kit.FlagSet) {
	f.StringVar(&flagCookies, "cookies", "", "Netscape cookie jar for a lent session")
	f.DurationVar(&flagCacheTTL, "cache-ttl", linkedin.DefaultCacheTTL, "on-disk cache freshness window")
	f.BoolVar(&flagRefresh, "refresh", false, "force re-fetch and overwrite the cache")
}
