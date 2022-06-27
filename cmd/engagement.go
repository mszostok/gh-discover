package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/mszostok/gh-discover/pkg/issues"
	"github.com/mszostok/gh-discover/pkg/users"
)

type EngagementOptions struct {
	Issues               []int
	Raw                  bool
	CacheTTL             time.Duration
	UniqueInIssue        bool
	OnlyPositiveReaction bool
	IgnoredUsers         []string
	UniqueAcrossIssues   bool
	GroupBy              string
}

// NewEngagement returns a root cobra.Command .
func NewEngagement() *cobra.Command {
	var opts EngagementOptions
	cmd := &cobra.Command{
		Use:          "engagement",
		Short:        "Check the user engagement",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gh.GQLClient(&api.ClientOptions{
				CacheTTL:    opts.CacheTTL,
				EnableCache: opts.CacheTTL > 0,
			})
			if err != nil {
				return err
			}

			repo, err := gh.CurrentRepository()
			if err != nil {
				return err
			}

			var output string
			switch opts.GroupBy {
			case "users":
				output, err = byUser(opts, client, repo)
				if err != nil {
					return err
				}
			case "issues":
				output, err = byIssue(opts, client, repo)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown %s group by, allowed values: users, issues", opts.GroupBy)
			}

			return printStdout(opts.Raw, output)
		},
	}

	flags := cmd.Flags()

	flags.IntSliceVar(&opts.Issues, "issues", nil, "Comma separated list of issues numbers.")
	flags.BoolVar(&opts.Raw, "raw", false, "Set to print raw Markdown format.")
	flags.DurationVar(&opts.CacheTTL, "cache-ttl", 0, `Cache the response, e.g. "3600s", "60m", "1h". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	flags.BoolVar(&opts.UniqueInIssue, "unique-in-issue", true, `Specify the username only once in the whole issue, e.g. don't mention twice if it's an author and also commented under own issue.`)
	flags.BoolVar(&opts.UniqueAcrossIssues, "unique-across-issues", false, `Specify the username only once in the whole issue, e.g. don't mention twice if it's an author and also commented under own issue.`)
	flags.BoolVar(&opts.OnlyPositiveReaction, "positive-reaction", false, "Count user only with a positive reaction")
	flags.StringSliceVar(&opts.IgnoredUsers, "ignore-users", nil, "Comma separated list of usernames that comments and reactions are ignored.")
	flags.StringVar(&opts.GroupBy, "group-by", "", "Define if output should be grouped by users or issues. Allowe values: users, issues")

	_ = cmd.MarkFlagRequired("group-by")

	return cmd
}

func byUser(opts EngagementOptions, client api.GQLClient, repo repository.Repository) (string, error) {
	engagement := users.NewEngagement(client)
	out, err := engagement.Fetch(users.FetchEngagementInput{
		Repository:           repo,
		Issues:               opts.Issues,
		UniqueInIssue:        opts.UniqueInIssue,
		OnlyPositiveReaction: opts.OnlyPositiveReaction,
		IgnoredUsers:         opts.IgnoredUsers,
		UniqueAcrossIssues:   opts.UniqueAcrossIssues,
	})
	if err != nil {
		return "", err
	}

	var buff strings.Builder

	buff.WriteString("\n## List of interested users to update on progress\n")
	for name, item := range out.Users {
		buff.WriteString(fmt.Sprintf("\n### User [%s](https://github.com/%s)\n", name, name))

		if len(item.Author) > 0 {
			buff.WriteString(fmt.Sprintf("\n\n- created:\n\n"))
			for pr := range item.Author {
				buff.WriteString(fmt.Sprintf("\n  - %s.\n", prURL(repo, pr)))
			}
		}
		if len(item.Commenter) > 0 {
			buff.WriteString(fmt.Sprintf("\n\n- commented on:\n\n"))
			for pr, number := range item.Commenter {
				buff.WriteString(fmt.Sprintf("\n  - %s - %d time(s).\n", prURL(repo, pr), number))
			}
		}
		if len(item.ReactionGiver) > 0 {
			buff.WriteString(fmt.Sprintf("\n\n- reacted to:\n\n"))
			for pr, details := range item.ReactionGiver {
				buff.WriteString(fmt.Sprintf("\n  - %s - added %s reaction.\n", prURL(repo, pr), details.Emoji))
			}
		}
	}

	return buff.String(), nil
}

func byIssue(opts EngagementOptions, client api.GQLClient, repo repository.Repository) (string, error) {
	engagement := issues.NewEngagement(client)
	out, err := engagement.Fetch(issues.FetchEngagementInput{
		Repository:           repo,
		Issues:               opts.Issues,
		UniqueInIssue:        opts.UniqueInIssue,
		OnlyPositiveReaction: opts.OnlyPositiveReaction,
		IgnoredUsers:         opts.IgnoredUsers,
		UniqueAcrossIssues:   opts.UniqueAcrossIssues,
	})
	if err != nil {
		return "", err
	}

	var buff strings.Builder

	buff.WriteString("\n## List of interested users to update on progress\n")
	for _, item := range out.Issues {
		// List of interested parties to be informed
		// can be informed about progress
		buff.WriteString(fmt.Sprintf("\n### Issue %s\n", prURL(repo, users.PR{Title: item.Title, Number: item.Number})))
		for _, category := range []string{issues.KindAuthor, issues.KindCommenter, issues.KindReactionGiver} {
			users := item.Users[category]
			if len(users) == 0 {
				continue
			}
			buff.WriteString(fmt.Sprintf("\n\n- %s:\n\n", category))
			for _, user := range users {
				buff.WriteString(fmt.Sprintf("  - @%s\n", user.Name))
			}
		}
	}

	return buff.String(), nil
}

func prURL(repo repository.Repository, pr users.PR) string {
	return fmt.Sprintf("%q ([#%d](https://github.com/%s/%s/issues/%d))", pr.Title, pr.Number, repo.Owner(), repo.Name(), pr.Number)
}

func printStdout(raw bool, data string) error {
	if raw {
		fmt.Println(data)
		return nil
	}

	width := 10000 // long by default, e.g. when save to file
	if term.IsTerminal(int(os.Stdout.Fd())) {
		w, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return err
		}
		width = w
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return err
	}

	rendered, err := r.Render(data)
	if err != nil {
		return err
	}
	fmt.Print(rendered)
	return nil
}
