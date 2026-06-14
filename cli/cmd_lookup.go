package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tamnd/linkedin-cli/linkedin"
)

func (a *App) idCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id <input> [more...]",
		Short: "Classify and normalize LinkedIn inputs",
		Long: "Classify a slug, URL, or urn:li:... URN into its kind (profile, company,\n" +
			"school, job, post) plus a canonical id and URL. Offline, no network.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			out := make([]linkedin.Ref, 0, len(args))
			for _, in := range args {
				out = append(out, linkedin.Classify(in))
			}
			return a.renderOrEmpty(out, len(out))
		},
	}
	return cmd
}

func (a *App) urlCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "url <kind> <id>",
		Short: "Build a canonical LinkedIn URL",
		Long: "Build a canonical URL from a kind and an id or slug.\n" +
			"Kinds: profile, company, school, job.",
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
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
				return codeError(exitUsage, fmt.Errorf("unknown kind %q (want profile|company|school|job)", kind))
			}
			_, _ = fmt.Fprintln(os.Stdout, u)
			return nil
		},
	}
	return cmd
}
