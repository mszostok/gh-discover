package issues

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
		Issues []EngagementIssueDetails
	}
	EngagementIssueDetails struct {
		Number int
		Title  string
		// Key is a reason: Author/Commenter/ReactionGiver
		Users map[string][]EngagedUser
	}
	EngagedUser struct {
		Name string
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

	var out FetchEngagementOutput
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

		issue := EngagementIssueDetails{
			Number: number,
			Title:  strings.TrimSpace(resp.Repository.Issue.Title),
		}

		issue.Users = map[string][]EngagedUser{}
		collector.Collect(number, issue.Users, KindAuthor, resp.Repository.Issue.Author.Login)

		for _, item := range resp.Repository.Issue.Reactions.Edges {
			reaction := gh.ReactionType(item.Node.Content)
			if in.OnlyPositiveReaction && reaction.HasNegativeVibe() {
				continue
			}
			collector.Collect(number, issue.Users, KindReactionGiver, item.Node.User.Login)
		}

		for _, item := range resp.Repository.Issue.Comments.Nodes {
			collector.Collect(number, issue.Users, KindCommenter, item.Author.Login)
		}

		out.Issues = append(out.Issues, issue)
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

func (u *uniqueCollector) Collect(number int, in map[string][]EngagedUser, key, val string) {
	if _, found := u.ignoredVals[val]; found {
		return
	}
	if key != KindAuthor {
		if !u.isUnique(number, key, val) {
			return
		}
	}
	in[key] = append(in[key], EngagedUser{Name: val})
	u.report(number, key, val)

	return
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
	KindAuthor        = "You are the author"
	KindCommenter     = "You added a comment"
	KindReactionGiver = "You added a reaction"
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
