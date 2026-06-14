package cli

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/tamnd/any-cli/kit"
)

func infoCmd() kit.Command {
	return kit.Command{
		Use:   "info",
		Short: "Print the resolved configuration",
		Run: func(ctx context.Context, args []string) error {
			a := appFromCtx(ctx)
			cfg := a.config()
			cookies := "none"
			if cfg.CookiePath != "" {
				cookies = cfg.CookiePath
			}
			return a.emitKV(
				"data-dir", cfg.DataDir,
				"cache-dir", cfg.CacheDir(),
				"store", a.StorePath(),
				"cache-ttl", cfg.CacheTTL.String(),
				"delay", cfg.Delay.String(),
				"timeout", cfg.Timeout.String(),
				"retries", fmt.Sprintf("%d", cfg.Retries),
				"cookies", cookies,
			)
		},
	}
}

func versionCmd() kit.Command {
	return kit.Command{
		Use:   "version",
		Short: "Print version, commit, and build date",
		Run: func(ctx context.Context, args []string) error {
			_, err := fmt.Fprintf(os.Stdout, "linkedin %s (commit %s, built %s, %s/%s)\n",
				Version, Commit, Date, runtime.GOOS, runtime.GOARCH)
			return err
		},
	}
}

// emitKV renders an ordered sequence of key/value pairs as records, so the
// resolved configuration formats the same way (list, json, csv) as any other
// command. Pairs are passed flat: key, value, key, value, ...
func (a *App) emitKV(pairs ...string) error {
	out, err := a.out()
	if err != nil {
		return err
	}
	for i := 0; i+1 < len(pairs); i += 2 {
		k, v := pairs[i], pairs[i+1]
		if err := out.Emit(Row{
			Cols:  []string{"key", "value"},
			Vals:  []string{k, v},
			Value: map[string]any{"key": k, "value": v},
		}); err != nil {
			return err
		}
	}
	return out.Flush()
}
