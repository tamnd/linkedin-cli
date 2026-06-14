package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (a *App) infoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Print the resolved configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			store := a.storePtr
			if store == "" {
				store = a.cfg.StorePath()
			}
			cookies := "none"
			if a.cfg.CookiePath != "" {
				cookies = a.cfg.CookiePath
			}
			_, _ = fmt.Fprintf(os.Stdout, "data-dir:  %s\n", a.cfg.DataDir)
			_, _ = fmt.Fprintf(os.Stdout, "cache-dir: %s\n", a.cfg.CacheDir())
			_, _ = fmt.Fprintf(os.Stdout, "store:     %s\n", store)
			_, _ = fmt.Fprintf(os.Stdout, "cache-ttl: %s\n", a.cfg.CacheTTL)
			_, _ = fmt.Fprintf(os.Stdout, "delay:     %s\n", a.cfg.Delay)
			_, _ = fmt.Fprintf(os.Stdout, "cookies:   %s\n", cookies)
			return nil
		},
	}
	return cmd
}

func (a *App) versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version, commit, and build date",
		RunE: func(_ *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintf(os.Stdout, "linkedin %s (commit %s, built %s)\n", Version, Commit, Date)
			return nil
		},
	}
	return cmd
}
