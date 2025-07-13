package cmd

import (
	"log"
	"slices"
	"strings"

	gh "github.com/cli/go-gh"
	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/shurcooL/githubv4"
)

type Repository struct {
	ID               string `graphql:"id"`
	Name             string `graphql:"name"`
	DefaultBranchRef struct {
		Name string `graphql:"name"`
	} `graphql:"defaultBranchRef"`
	Owner struct {
		ID    string `graphql:"id"`
		Login string `graphql:"login"`
	}
}

func parseRepository(repository string) (string, string) {
	if repository == "" {
		return "", ""
	}

	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}

func GetRepository(client *api.GraphQLClient, repository string) Repository {
	owner, name := parseRepository(repository)

	if owner == "" || name == "" {
		repo, err := gh.CurrentRepository()
		if err != nil {
			log.Fatal(err)
		}
		owner = repo.Owner()
		name = repo.Name()
	}

	var query struct {
		Repository Repository `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]any{
		"owner": graphql.String(owner),
		"name":  graphql.String(name),
	}

	err := client.Query("Repository", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	return query.Repository
}

type PullRequest struct {
	ID          string `json:"id" graphql:"id"`
	Title       string `json:"title" graphql:"title"`
	URL         string `json:"url" graphql:"url"`
	IsDraft     bool   `json:"isDraft" graphql:"isDraft"`
	BaseRefName string `json:"baseRefName" graphql:"baseRefName"`
	HeadRefName string `json:"headRefName" graphql:"headRefName"`
	Author      struct {
		Login string `json:"login" graphql:"login"`
	} `json:"author" graphql:"author"`
}

type Options struct {
	Repository    string
	States        []githubv4.PullRequestState
	Authors       []string
	ExcludeDrafts bool
	Format        Format
}

func GetPullRequests(client *api.GraphQLClient, owner string, name string, states []githubv4.PullRequestState, authors []string, excludeDrafts bool, format Format) []PullRequest {
	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes    []PullRequest
				PageInfo struct {
					HasNextPage bool   `graphql:"hasNextPage"`
					EndCursor   string `graphql:"endCursor"`
				} `graphql:"pageInfo"`
			} `graphql:"pullRequests(states: $states, first: $first, after: $after)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]any{
		"owner":  graphql.String(owner),
		"name":   graphql.String(name),
		"states": states,
		"first":  graphql.Int(50),
		"after":  (*graphql.String)(nil),
	}

	pullRequests := []PullRequest{}
	for {
		err := client.Query("PullRequests", &query, variables)
		if err != nil {
			log.Fatal(err)
		}

		pullRequests = append(pullRequests, query.Repository.PullRequests.Nodes...)

		if query.Repository.PullRequests.PageInfo.HasNextPage {
			variables["after"] = graphql.String(query.Repository.PullRequests.PageInfo.EndCursor)
		} else {
			break
		}
	}

	if excludeDrafts {
		pullRequests = filterDrafts(pullRequests)
	}

	if len(authors) > 0 {
		pullRequests = filterByAuthors(pullRequests, authors)
	}

	return pullRequests
}

func filterDrafts(pullRequests []PullRequest) []PullRequest {
	var res []PullRequest
	for _, pr := range pullRequests {
		if !pr.IsDraft {
			res = append(res, pr)
		}
	}
	return res
}

func filterByAuthors(pullRequests []PullRequest, authors []string) []PullRequest {
	var res []PullRequest
	for _, pr := range pullRequests {
		if slices.Contains(authors, pr.Author.Login) {
			res = append(res, pr)
		}
	}
	return res
}
