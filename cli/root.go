package cli

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/tamnd/linkedin-cli/linkedin"
)

// Build metadata, injected via -ldflags by the Makefile/goreleaser.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// App holds shared state threaded through every command.
type App struct {
	cfg      linkedin.Config
	client   *linkedin.Client
	cache    *linkedin.Cache
	store    *linkedin.Store
	storePtr string

	// global flags
	format   string
	jsonl    bool
	fields   []string
	noHeader bool
	template string
	color    string
	limit    int
	quiet    bool
}

// exit codes (see spec §6).
const (
	exitError   = 1
	exitUsage   = 2
	exitNoData  = 3
	exitPartial = 4
	exitBlocked = 5
)

// ExitError carries a process exit code up to main.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit %d", e.Code)
}

func (e *ExitError) Unwrap() error { return e.Err }

func codeError(code int, err error) error { return &ExitError{Code: code, Err: err} }

// NewRootCmd builds the full command tree.
func NewRootCmd() *cobra.Command {
	app := &App{cfg: linkedin.DefaultConfig()}

	root := &cobra.Command{
		Use:   "linkedin",
		Short: "Read public LinkedIn data from the command line",
		Long: "linkedin reads the public profile, company, and job pages and the guest\n" +
			"job-search endpoints LinkedIn serves to anonymous visitors, and returns\n" +
			"rich, structured records as table, JSON, JSONL, CSV, TSV, or URLs.\n\n" +
			"It uses no API key and no account. Most member surfaces are gated behind a\n" +
			"sign-in wall, so commands report clearly when a page is walled.\n\n" +
			"linkedin is an independent tool and is not affiliated with, endorsed by, or\n" +
			"sponsored by LinkedIn.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return app.setup(cmd)
		},
		PersistentPostRun: func(_ *cobra.Command, _ []string) {
			if app.store != nil {
				_ = app.store.Close()
			}
		},
	}

	pf := root.PersistentFlags()
	pf.StringVarP(&app.format, "format", "f", "", "output: table|json|jsonl|csv|tsv|url|raw (default: table on a TTY, jsonl when piped)")
	pf.BoolVarP(&app.jsonl, "jsonl", "j", false, "shorthand for --format jsonl")
	pf.StringSliceVar(&app.fields, "fields", nil, "comma-separated columns to include")
	pf.BoolVar(&app.noHeader, "no-header", false, "omit the header row in table/csv/tsv")
	pf.StringVar(&app.template, "template", "", "Go text/template applied per record")
	pf.StringVar(&app.color, "color", "auto", "color: auto|always|never")
	pf.IntVarP(&app.limit, "limit", "n", 0, "limit number of results (0 = no limit)")
	pf.BoolVarP(&app.quiet, "quiet", "q", false, "suppress progress on stderr")

	pf.IntVar(&app.cfg.Workers, "workers", linkedin.DefaultWorkers, "concurrent workers for multi-fetch")
	pf.DurationVar(&app.cfg.Delay, "delay", linkedin.DefaultDelay, "minimum spacing between requests")
	pf.DurationVar(&app.cfg.Timeout, "timeout", linkedin.DefaultTimeout, "per-request timeout")
	pf.IntVar(&app.cfg.Retries, "retries", linkedin.DefaultRetries, "retry attempts on 429/5xx")
	pf.DurationVar(&app.cfg.CacheTTL, "cache-ttl", linkedin.DefaultCacheTTL, "on-disk cache freshness window")
	pf.BoolVar(&app.cfg.NoCache, "no-cache", false, "bypass the on-disk page cache")
	pf.BoolVar(&app.cfg.Refresh, "refresh", false, "force re-fetch and overwrite the cache")
	pf.StringVar(&app.cfg.DataDir, "data-dir", app.cfg.DataDir, "root directory for cache and store")
	pf.StringVar(&app.storePtr, "store", "", "SQLite store path (default: <data-dir>/linkedin.db)")
	pf.StringVar(&app.cfg.CookiePath, "cookies", "", "Netscape cookie jar for a lent session")

	root.AddCommand(
		app.profileCmd(),
		app.companyCmd(),
		app.jobCmd(),
		app.jobsCmd(),
		app.postCmd(),
		app.idCmd(),
		app.urlCmd(),
		app.dbCmd(),
		app.cacheCmd(),
		app.infoCmd(),
		app.versionCmd(),
	)
	return root
}

// setup resolves output defaults and constructs the shared client/cache.
func (a *App) setup(_ *cobra.Command) error {
	if a.jsonl {
		a.format = string(FormatJSONL)
	}
	if a.format == "" {
		if isatty.IsTerminal(os.Stdout.Fd()) {
			a.format = string(FormatTable)
		} else {
			a.format = string(FormatJSONL)
		}
	}
	if a.cfg.CookiePath != "" {
		cookies, err := linkedin.LoadCookies(a.cfg.CookiePath)
		if err != nil {
			return codeError(exitUsage, fmt.Errorf("load cookies: %w", err))
		}
		c, err := linkedin.NewClientWithCookies(a.cfg, cookies)
		if err != nil {
			return err
		}
		a.client = c
	} else {
		a.client = linkedin.NewClient(a.cfg)
	}
	a.cache = linkedin.NewCache(a.cfg)
	return nil
}

// openStore lazily opens the SQLite store.
func (a *App) openStore() (*linkedin.Store, error) {
	if a.store != nil {
		return a.store, nil
	}
	path := a.storePtr
	if path == "" {
		path = a.cfg.StorePath()
	}
	st, err := linkedin.OpenStore(path)
	if err != nil {
		return nil, err
	}
	a.store = st
	return st, nil
}

// render writes records using the resolved global flags.
func (a *App) render(records any) error {
	r := NewRenderer(os.Stdout, Format(a.format), a.fields, a.noHeader, a.template)
	return r.Render(records)
}

// renderOrEmpty renders records, mapping an empty result to exit code 3.
func (a *App) renderOrEmpty(records any, n int) error {
	if err := a.render(records); err != nil {
		return err
	}
	if n == 0 {
		return codeError(exitNoData, nil)
	}
	return nil
}

// progressf prints a progress line to stderr unless --quiet.
func (a *App) progressf(format string, args ...any) {
	if a.quiet {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// mapFetchErr converts a library error into the right exit code.
func mapFetchErr(err error) error {
	switch {
	case err == nil:
		return nil
	case isBlocked(err):
		return codeError(exitBlocked, fmt.Errorf("%w\nhint: this surface is often gated; pass --cookies to lend a signed-in session, or slow down with --delay", err))
	case isNotFound(err):
		return codeError(exitNoData, err)
	default:
		return codeError(exitError, err)
	}
}
