package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

func (a *App) dbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Inspect the SQLite record store",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "path",
			Short: "Print the store path",
			RunE: func(_ *cobra.Command, _ []string) error {
				path := a.storePtr
				if path == "" {
					path = a.cfg.StorePath()
				}
				_, _ = fmt.Fprintln(os.Stdout, path)
				return nil
			},
		},
		&cobra.Command{
			Use:   "count",
			Short: "Print record counts per kind",
			RunE: func(_ *cobra.Command, _ []string) error {
				st, err := a.openStore()
				if err != nil {
					return err
				}
				counts, err := st.CountsByKind()
				if err != nil {
					return err
				}
				for k, n := range counts {
					_, _ = fmt.Fprintf(os.Stdout, "%s\t%d\n", k, n)
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "query <kind>",
			Short: "Stream stored records of a kind as JSONL",
			Args:  cobra.ExactArgs(1),
			RunE: func(_ *cobra.Command, args []string) error {
				st, err := a.openStore()
				if err != nil {
					return err
				}
				enc := json.NewEncoder(os.Stdout)
				n := 0
				err = st.Each(args[0], func(_ string, data []byte) error {
					n++
					var v any
					if json.Unmarshal(data, &v) != nil {
						return nil
					}
					return enc.Encode(v)
				})
				if err != nil {
					return err
				}
				if n == 0 {
					return codeError(exitNoData, nil)
				}
				return nil
			},
		},
	)
	return cmd
}

func (a *App) cacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the on-disk page cache",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "path",
			Short: "Print the cache directory",
			RunE: func(_ *cobra.Command, _ []string) error {
				_, _ = fmt.Fprintln(os.Stdout, a.cfg.CacheDir())
				return nil
			},
		},
		&cobra.Command{
			Use:   "info",
			Short: "Print cache file count and size",
			RunE: func(_ *cobra.Command, _ []string) error {
				files, bytes, err := a.cache.Stats()
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(os.Stdout, "%d files, %s\n", files, humanize.Bytes(uint64(bytes)))
				return nil
			},
		},
		&cobra.Command{
			Use:   "clear",
			Short: "Remove the entire cache",
			RunE: func(_ *cobra.Command, _ []string) error {
				if err := a.cache.Clear(); err != nil {
					return err
				}
				_, _ = fmt.Fprintln(os.Stderr, "cache cleared")
				return nil
			},
		},
	)
	return cmd
}
