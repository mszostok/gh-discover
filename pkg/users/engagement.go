package users

import (
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"

	"github.com/mszostok/gh-discover/pkg/gh"
)

type (
	FetchEngagementInput struct {
		Repository           repository.Repository
		Issues               []int
		UniqueInIssue        bool
		UniqueAcrossIssues   bool
		OnlyPositiveReaction bool
		IgnoredUsers         []string
	}
	FetchEngagementOutput struct {
		// Key is a username
		Users map[string]EngagementUserDetails
	}
	EngagementUserDetails struct {
		ReactionGiver map[PR]ReactionDetails
		Author        map[PR]struct{}
		Commenter     map[PR]int
	}
	PR struct {
		Title  string
		Number int
	}
	ReactionDetails struct {
		Emoji string
	}
)

type Engagement struct {
	client api.GQLClient
}

func NewEngagement(client api.GQLClient) *Engagement {
	return &Engagement{client: client}
}

func (e *Engagement) Fetch(in FetchEngagementInput) (FetchEngagementOutput, error) {
	query := heredoc.Doc(`
		query FetchIssue(
					$owner: String!,
					$name: String!,
					$issueNumber: Int!,

				) {
			repository(owner:$owner, name:$name) {
				issue(number: $issueNumber)  {
					title
					author {
						login
					}
					comments(first: 100) {
						nodes {
							author {
								login
							}
						}
					}
					reactions(first: 100) {
						edges {
							node {
  							content
								user {
									login
								}
							}
						}
					}
				}
			}
		}
		`)

	collector := newUniqueCollector(in.UniqueInIssue, in.UniqueAcrossIssues, in.IgnoredUsers)

	out := FetchEngagementOutput{
		Users: map[string]EngagementUserDetails{},
	}
	for _, number := range in.Issues {
		variables := map[string]interface{}{
			"owner":       in.Repository.Owner(),
			"name":        in.Repository.Name(),
			"issueNumber": number,
		}
		var resp fetchEngagementResponse

		// TODO: paginate
		err := e.client.Do(query, variables, &resp)
		if err != nil {
			return FetchEngagementOutput{}, err
		}

		pr := PR{
			Number: number,
			Title:  strings.TrimSpace(resp.Repository.Issue.Title),
		}
		collector.Collect(pr, out.Users, KindAuthor, resp.Repository.Issue.Author.Login, "")

		for _, item := range resp.Repository.Issue.Reactions.Edges {
			reaction := gh.ReactionType(item.Node.Content)
			if in.OnlyPositiveReaction && reaction.HasNegativeVibe() {
				continue
			}
			name := item.Node.User.Login
			// TODO: fix setting reaction...
			collector.Collect(pr, out.Users, KindReactionGiver, name, reaction.Emoji())
		}

		for _, item := range resp.Repository.Issue.Comments.Nodes {
			collector.Collect(pr, out.Users, KindCommenter, item.Author.Login, "")

		}
	}

	return out, nil
}

type uniqueCollector struct {
	collected             map[int]map[string]map[string]struct{}
	ignoredVals           map[string]struct{}
	uniqueInAllCategories bool
	uniqueAcrossIssues    bool
}

func newUniqueCollector(uniqueInIssue bool, uniqueAcrossIssues bool, ignoreVals []string) *uniqueCollector {
	ignoredVals := map[string]struct{}{}
	for _, name := range ignoreVals {
		ignoredVals[name] = struct{}{}
	}
	return &uniqueCollector{
		collected:             map[int]map[string]map[string]struct{}{},
		uniqueInAllCategories: uniqueInIssue,
		uniqueAcrossIssues:    uniqueAcrossIssues,
		ignoredVals:           ignoredVals,
	}
}

func (u *uniqueCollector) Collect(pr PR, users map[string]EngagementUserDetails, kind, name string, args ...string) bool {
	if _, found := u.ignoredVals[name]; found {
		return false
	}
	if !u.isUnique(pr.Number, kind, name) {
		return false
	}

	if _, found := users[name]; !found {
		users[name] = EngagementUserDetails{
			Author:        map[PR]struct{}{},
			ReactionGiver: map[PR]ReactionDetails{},
			Commenter:     map[PR]int{},
		}
	}

	switch kind {
	case KindCommenter:
		users[name].Commenter[pr] += 1
	case KindAuthor:
		users[name].Author[pr] = struct{}{}
	case KindReactionGiver:
		if len(args) == 0 {
			args = []string{""}
		}
		users[name].ReactionGiver[pr] = ReactionDetails{Emoji: args[0]}
	}

	u.report(pr.Number, kind, name)
	return true
}

func (u *uniqueCollector) isUnique(issue int, category, val string) bool {
	if u.uniqueInAllCategories {
		category = "all"
	}
	if u.uniqueAcrossIssues {
		issue = 0
	}
	_, found := u.collected[issue][category][val]
	return !found
}
func (u *uniqueCollector) report(issue int, category, val string) {
	if u.uniqueInAllCategories {
		category = "all"
	}
	if u.uniqueAcrossIssues {
		issue = 0
	}
	if u.collected[issue] == nil {
		u.collected[issue] = map[string]map[string]struct{}{}
	}
	if u.collected[issue][category] == nil {
		u.collected[issue][category] = map[string]struct{}{}
	}
	u.collected[issue][category][val] = struct{}{}
}

const (
	KindAuthor        = "You are the issue author"
	KindCommenter     = "You commented under the issue"
	KindReactionGiver = "You added reaction under the issue"
)

type (
	fetchEngagementResponse struct {
		Repository struct {
			Issue struct {
				Title     string    `json:"title"`
				Author    Author    `json:"author"`
				Comments  Comments  `json:"comments"`
				Reactions Reactions `json:"reactions"`
			} `json:"issue"`
		} `json:"repository"`
	}
	Author struct {
		Login string `json:"login"`
	}
	Comments struct {
		Nodes []struct {
			Author Author `json:"author"`
		} `json:"nodes"`
	}
	Reactions struct {
		Edges []struct {
			Node struct {
				User    Author `json:"user"`
				Content string `json:"content"`
			} `json:"node"`
		} `json:"edges"`
	}
)
