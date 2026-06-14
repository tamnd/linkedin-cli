package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
	"github.com/tamnd/any-cli/kit/render"
	"github.com/tamnd/linkedin-cli/linkedin"
)

// Row is one output record: an ordered, curated column set (Cols/Vals) for the
// table, csv, tsv, and url views, plus the full typed object as Value for json,
// jsonl, raw, and template. It is kit's render.Record, so every row builder
// feeds straight into the shared renderer with no per-format code of our own.
type Row = render.Record

// linkedin-only global flags, bound on the root by the GlobalFlags hook in
// root.go. They are not part of the kit baseline (which already carries --rate,
// --retries, --timeout, --data-dir, --no-cache), so linkedin owns just these.
var (
	flagCookies  string
	flagCacheTTL time.Duration
	flagRefresh  bool
)

// App is the per-run state every command works through. Each command is a
// kit.Command escape hatch that rebuilds this state from the run context with
// appFromCtx, so they share the resolved config, the client/cache, and the
// output settings kit resolved once for the run.
type App struct {
	actx   context.Context
	st     *kit.State // the resolved run state; nil only on the defensive fallback
	cfg    linkedin.Config
	client *linkedin.Client
	cache  *linkedin.Cache
	store  *linkedin.Store
	limit  int
	quiet  bool
	dryRun bool
}

// appFromCtx assembles the run's App from the resolved kit State. It folds the
// framework globals (rate, retries, timeout, data dir, output settings) and the
// linkedin-specific globals (cookies, cache-ttl, refresh) into the
// linkedin.Config the standalone binary has always built.
func appFromCtx(ctx context.Context) *App {
	st := kit.FromContext(ctx)
	a := &App{actx: ctx, st: st}
	if st == nil {
		// No run state resolved (should not happen in normal use); fall back to
		// plain defaults so the command still does something sane.
		a.cfg = linkedinConfig(kit.Config{})
		a.cache = linkedin.NewCache(a.cfg)
		return a
	}
	kc := st.Config
	a.cfg = linkedinConfig(kc)
	a.limit = st.Globals.Limit
	a.quiet = kc.Quiet
	a.dryRun = kc.DryRun
	a.cache = linkedin.NewCache(a.cfg)
	return a
}

// linkedinConfig builds the resolved linkedin.Config from the library defaults
// and the kit globals folded on top. kit's "unset" sentinels (rate 0, retries
// -1, timeout 0) leave the politer linkedin defaults (2s, 3, 30s) standing; the
// library maps kit's --rate onto its request spacing.
func linkedinConfig(kc kit.Config) linkedin.Config {
	cfg := linkedin.DefaultConfig()
	if kc.Rate > 0 {
		cfg.Delay = kc.Rate
	}
	if kc.Retries >= 0 {
		cfg.Retries = kc.Retries
	}
	if kc.Timeout > 0 {
		cfg.Timeout = kc.Timeout
	}
	if kc.NoCache {
		cfg.NoCache = true
	}
	if kc.DataDir != "" {
		cfg.DataDir = kc.DataDir
	}
	if flagCacheTTL > 0 {
		cfg.CacheTTL = flagCacheTTL
	}
	if flagRefresh {
		cfg.Refresh = true
	}
	if flagCookies != "" {
		cfg.CookiePath = flagCookies
	}
	return cfg
}

// ctx returns the run context (carries cancellation from the signal handler).
func (a *App) ctx() context.Context { return a.actx }

// config returns the resolved linkedin configuration.
func (a *App) config() linkedin.Config { return a.cfg }

// clientOf builds (once) the HTTP client, loading the lent cookie jar when one
// is configured.
func (a *App) clientOf() (*linkedin.Client, error) {
	if a.client != nil {
		return a.client, nil
	}
	if a.cfg.CookiePath != "" {
		cookies, err := linkedin.LoadCookies(a.cfg.CookiePath)
		if err != nil {
			return nil, errs.Usage("load cookies: %v", err)
		}
		c, err := linkedin.NewClientWithCookies(a.cfg, cookies)
		if err != nil {
			return nil, err
		}
		a.client = c
	} else {
		a.client = linkedin.NewClient(a.cfg)
	}
	return a.client, nil
}

// out builds the shared kit renderer over stdout from the run's resolved output
// settings. It differs from kit's default in one place: with no -o, linkedin
// prints the readable list view on a terminal and jsonl when piped, so scripts
// stay machine-readable. An explicit -o or --template always wins.
func (a *App) out() (*render.Renderer, error) {
	if a.st == nil {
		return render.New(render.Options{Format: render.List, Writer: os.Stdout})
	}
	o := a.st.Output
	format := render.Format(o.Format)
	if o.Template == "" && (format == "" || format == render.Auto) {
		if o.IsTTY {
			format = render.List
		} else {
			format = render.JSONL
		}
	}
	return render.New(render.Options{
		Format:   format,
		IsTTY:    o.IsTTY,
		Color:    o.Color,
		Fields:   o.Fields,
		NoHeader: o.NoHeader,
		Template: o.Template,
		Width:    o.Width,
		Writer:   os.Stdout,
	})
}

// StorePath is the fixed location of the typed local store, under the data dir.
func (a *App) StorePath() string {
	dir := a.cfg.DataDir
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "linkedin.db")
}

// openStore opens (creating the data dir) the local record store at the fixed
// path.
func (a *App) openStore() (*linkedin.Store, error) {
	if a.store != nil {
		return a.store, nil
	}
	st, err := linkedin.OpenStore(a.StorePath())
	if err != nil {
		return nil, err
	}
	a.store = st
	return st, nil
}

// logf prints progress to stderr unless --quiet.
func (a *App) logf(format string, args ...any) {
	if !a.quiet {
		_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// mapErr converts a library error into the kit error kind that carries the
// matching exit code (need-auth 4, rate-limited 5, not-found 6); anything else
// stays a generic failure (exit 1). Every escape-hatch Run wraps its error in
// this.
func mapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, linkedin.ErrBlocked):
		return errs.NeedAuth("%v\nhint: this surface is often gated; pass --cookies to lend a signed-in session, or slow down with --rate", err)
	case errors.Is(err, linkedin.ErrRateLimited):
		return errs.RateLimited("%v", err)
	case errors.Is(err, linkedin.ErrNotFound):
		return errs.NotFound("%v", err)
	default:
		return err
	}
}
