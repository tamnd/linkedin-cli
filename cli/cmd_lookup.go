package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
	"github.com/tamnd/linkedin-cli/linkedin"
)

func idCmd() kit.Command {
	return kit.Command{
		Use:   "id <input> [more...]",
		Short: "Classify and normalize LinkedIn inputs",
		Long: "Classify a slug, URL, or urn:li:... URN into its kind (profile, company, " +
			"school, job, post) plus a canonical id and URL. Offline, no network.",
		Args: kit.MinimumNArgs(1),
		Run: func(ctx context.Context, args []string) error {
			a := appFromCtx(ctx)
			rows := make([]Row, 0, len(args))
			for _, in := range args {
				rows = append(rows, refRow(linkedin.Classify(in)))
			}
			return a.finish(rows, nil)
		},
	}
}

func urlCmd() kit.Command {
	return kit.Command{
		Use:   "url <kind> <id>",
		Short: "Build a canonical LinkedIn URL",
		Long: "Build a canonical URL from a kind and an id or slug.\n" +
			"Kinds: profile, company, school, job.",
		Args: kit.ExactArgs(2),
		Run: func(ctx context.Context, args []string) error {
			kind, id := args[0], args[1]
			var u string
			switch kind {
			case linkedin.KindProfile:
				u = linkedin.ProfileURL(id)
			case linkedin.KindCompany:
				u = linkedin.CompanyURL(id)
			case linkedin.KindSchool:
				u = linkedin.SchoolURL(id)
			case linkedin.KindJob:
				u = linkedin.JobURL(id)
			default:
				return errs.Usage("unknown kind %q (want profile|company|school|job)", kind)
			}
			_, err := fmt.Fprintln(os.Stdout, u)
			return err
		},
	}
}
