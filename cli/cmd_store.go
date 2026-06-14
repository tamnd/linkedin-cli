package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

func dbCmd() kit.Command {
	return kit.Command{
		Use:   "db",
		Short: "Inspect the local record store",
		Sub: []kit.Command{
			{
				Use:   "path",
				Short: "Print the store path",
				Run: func(ctx context.Context, args []string) error {
					a := appFromCtx(ctx)
					_, err := fmt.Fprintln(os.Stdout, a.StorePath())
					return err
				},
			},
			{
				Use:   "count",
				Short: "Print record counts per kind",
				Run: func(ctx context.Context, args []string) error {
					a := appFromCtx(ctx)
					st, err := a.openStore()
					if err != nil {
						return err
					}
					counts, err := st.CountsByKind()
					if err != nil {
						return err
					}
					out, err := a.out()
					if err != nil {
						return err
					}
					for k, n := range counts {
						if err := out.Emit(Row{
							Cols:  []string{"kind", "count"},
							Vals:  []string{k, fmt.Sprintf("%d", n)},
							Value: map[string]any{"kind": k, "count": n},
						}); err != nil {
							return err
						}
					}
					return out.Flush()
				},
			},
			{
				Use:   "query <kind>",
				Short: "Stream stored records of a kind as JSONL",
				Args:  kit.ExactArgs(1),
				Run: func(ctx context.Context, args []string) error {
					a := appFromCtx(ctx)
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
						return errs.NoResults("no stored %s records", args[0])
					}
					return nil
				},
			},
		},
	}
}

func cacheCmd() kit.Command {
	return kit.Command{
		Use:   "cache",
		Short: "Manage the on-disk page cache",
		Sub: []kit.Command{
			{
				Use:   "path",
				Short: "Print the cache directory",
				Run: func(ctx context.Context, args []string) error {
					a := appFromCtx(ctx)
					_, err := fmt.Fprintln(os.Stdout, a.config().CacheDir())
					return err
				},
			},
			{
				Use:   "info",
				Short: "Print cache file count and size",
				Run: func(ctx context.Context, args []string) error {
					a := appFromCtx(ctx)
					files, bytes, err := a.cache.Stats()
					if err != nil {
						return err
					}
					_, err = fmt.Fprintf(os.Stdout, "%d files, %s\n", files, humanize.Bytes(uint64(bytes)))
					return err
				},
			},
			{
				Use:   "clear",
				Short: "Remove the entire cache",
				Write: true,
				Run: func(ctx context.Context, args []string) error {
					a := appFromCtx(ctx)
					if err := a.cache.Clear(); err != nil {
						return err
					}
					a.logf("cache cleared")
					return nil
				},
			},
		},
	}
}
